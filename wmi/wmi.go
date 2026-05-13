//go:build windows

// Package wmi 提供 Windows Management Instrumentation 查询功能。
package wmi

import (
	"fmt"
	"math"
	"reflect"
	"strings"
	"sync"
	"time"

	ole "github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

// Session 表示一个 WMI 连接会话
type Session struct {
	mu      sync.Mutex
	service *ole.IDispatch
	closed  bool
}

// Connect 建立 WMI 连接会话
//   namespace - 可选参数，WMI 命名空间（默认 root\cimv2）
//   返回 - WMI 会话对象
//   返回 - 错误信息
func Connect(namespace ...string) (*Session, error) {
	ns := `root\cimv2`
	if len(namespace) > 0 && namespace[0] != "" {
		ns = namespace[0]
	}

	if err := ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED); err != nil {
		return nil, fmt.Errorf("CoInitializeEx 失败: %w", err)
	}

	unknown, err := oleutil.CreateObject("WbemScripting.SWbemLocator")
	if err != nil {
		ole.CoUninitialize()
		return nil, fmt.Errorf("创建 SWbemLocator 失败: %w", err)
	}
	defer unknown.Release()

	locator, err := unknown.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		ole.CoUninitialize()
		return nil, fmt.Errorf("QueryInterface SWbemLocator 失败: %w", err)
	}
	defer locator.Release()

	serviceVar, err := oleutil.CallMethod(locator, "ConnectServer", ns)
	if err != nil {
		ole.CoUninitialize()
		return nil, fmt.Errorf("ConnectServer 失败: %w", err)
	}

	service := serviceVar.ToIDispatch()
	serviceVar.Clear()

	if service == nil {
		ole.CoUninitialize()
		return nil, fmt.Errorf("ConnectServer 返回空的 service 对象")
	}

	return &Session{service: service}, nil
}

// Close 关闭 WMI 连接并释放 COM 资源
func (s *Session) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	s.closed = true
	if s.service != nil {
		s.service.Release()
	}
	ole.CoUninitialize()
}

// QueryResult 存储 WMI 查询结果
type QueryResult struct {
	// Columns 列名列表
	Columns []string
	// Rows 查询结果行数据，每行为列名到值的映射
	Rows []map[string]interface{}
}

// Query 执行 WQL 查询并返回动态结果
//   wql - WQL 查询语句
//   返回 - 查询结果
//   返回 - 错误信息
func (s *Session) Query(wql string) (*QueryResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil, fmt.Errorf("session 已关闭")
	}

	resultVar, err := oleutil.CallMethod(s.service, "ExecQuery", wql)
	if err != nil {
		return nil, fmt.Errorf("ExecQuery 失败: %w", err)
	}
	defer resultVar.Clear()

	objectSet := resultVar.ToIDispatch()
	if objectSet == nil {
		return nil, fmt.Errorf("ExecQuery 返回空结果集")
	}
	defer objectSet.Release()

	newEnum, err := objectSet.GetProperty("_NewEnum")
	if err != nil {
		return nil, fmt.Errorf("获取 _NewEnum 失败: %w", err)
	}
	defer newEnum.Clear()

	enum, err := newEnum.ToIUnknown().IEnumVARIANT(ole.IID_IEnumVariant)
	if err != nil {
		return nil, fmt.Errorf("获取 IEnumVARIANT 失败: %w", err)
	}
	defer enum.Release()

	var columns []string
	var rows []map[string]interface{}

	for {
		item, length, err := enum.Next(1)
		if length == 0 {
			break
		}
		if err != nil {
			break
		}

		obj := item.ToIDispatch()
		if obj == nil {
			item.Clear()
			continue
		}

		if columns == nil {
			if err := discoverColumns(obj, &columns); err != nil {
				obj.Release()
				item.Clear()
				return nil, fmt.Errorf("发现列名失败: %w", err)
			}
		}

		row := make(map[string]interface{}, len(columns))
		for _, col := range columns {
			valVar, err := oleutil.GetProperty(obj, col)
			if err != nil {
				continue
			}
			row[col] = variantToValue(valVar)
			valVar.Clear()
		}
		rows = append(rows, row)

		obj.Release()
		item.Clear()
	}

	return &QueryResult{Columns: columns, Rows: rows}, nil
}

// discoverColumns 从 WMI 对象的 Properties_ 集合中提取所有属性名称
func discoverColumns(obj *ole.IDispatch, columns *[]string) error {
	propsVar, err := oleutil.GetProperty(obj, "Properties_")
	if err != nil {
		return fmt.Errorf("获取 Properties_ 失败: %w", err)
	}
	defer propsVar.Clear()

	props := propsVar.ToIDispatch()
	if props == nil {
		return fmt.Errorf("Properties_ 返回空集合")
	}
	defer props.Release()

	newEnum, err := props.GetProperty("_NewEnum")
	if err != nil {
		return fmt.Errorf("获取 Properties_._NewEnum 失败: %w", err)
	}
	defer newEnum.Clear()

	enum, err := newEnum.ToIUnknown().IEnumVARIANT(ole.IID_IEnumVariant)
	if err != nil {
		return fmt.Errorf("获取 Properties_ IEnumVARIANT 失败: %w", err)
	}
	defer enum.Release()

	for {
		pItem, pLen, pErr := enum.Next(1)
		if pLen == 0 {
			break
		}
		if pErr != nil {
			break
		}

		pObj := pItem.ToIDispatch()
		if pObj == nil {
			pItem.Clear()
			continue
		}

		nameVar, err := oleutil.GetProperty(pObj, "Name")
		if err != nil {
			pObj.Release()
			pItem.Clear()
			continue
		}
		*columns = append(*columns, nameVar.ToString())
		nameVar.Clear()

		pObj.Release()
		pItem.Clear()
	}

	return nil
}

// variantToValue 将 VARIANT 转换为对应的 Go 类型值
func variantToValue(v *ole.VARIANT) interface{} {
	if v == nil {
		return nil
	}
	switch v.VT {
	case ole.VT_I4:
		return int32(v.Val)
	case ole.VT_UI4:
		return uint32(v.Val)
	case ole.VT_I8:
		return v.Val
	case ole.VT_UI8:
		return uint64(v.Val)
	case ole.VT_BSTR:
		return v.ToString()
	case ole.VT_BOOL:
		return (v.Val & 0xffff) != 0
	case ole.VT_R4:
		return math.Float32frombits(uint32(v.Val))
	case ole.VT_R8:
		return math.Float64frombits(uint64(v.Val))
	case ole.VT_DATE:
		d := uint64(v.Val)
		date, err := ole.GetVariantDate(d)
		if err != nil {
			return time.Time{}
		}
		return date
	case ole.VT_NULL, ole.VT_EMPTY:
		return nil
	default:
		// 不支持的 VARTYPE 返回零值
		return nil
	}
}

// QueryStruct 执行 WQL 查询并将结果映射到结构体切片
//   wql - WQL 查询语句
//   dest - 目标结构体切片指针，如 &[]Process{}
//   返回 - 错误信息
//
// 映射规则：优先使用 wmi 结构体标签，否则按字段名不区分大小写匹配
func (s *Session) QueryStruct(wql string, dest interface{}) error {
	destVal := reflect.ValueOf(dest)
	if destVal.Kind() != reflect.Ptr || destVal.IsNil() {
		return fmt.Errorf("dest 必须是非空指针")
	}

	sliceVal := destVal.Elem()
	if sliceVal.Kind() != reflect.Slice {
		return fmt.Errorf("dest 必须指向切片")
	}

	elemType := sliceVal.Type().Elem()
	if elemType.Kind() != reflect.Struct {
		return fmt.Errorf("切片元素必须是结构体")
	}

	result, err := s.Query(wql)
	if err != nil {
		return err
	}

	if len(result.Columns) == 0 || len(result.Rows) == 0 {
		sliceVal.Set(reflect.MakeSlice(sliceVal.Type(), 0, 0))
		return nil
	}

	// 构建列名到结构体字段索引的映射
	fieldMap := make(map[string]int)
	for i := 0; i < elemType.NumField(); i++ {
		f := elemType.Field(i)
		tag := f.Tag.Get("wmi")
		if tag != "" {
			fieldMap[tag] = i
		} else {
			fieldMap[strings.ToUpper(f.Name)] = i
		}
	}

	slice := reflect.MakeSlice(sliceVal.Type(), len(result.Rows), len(result.Rows))
	for i, row := range result.Rows {
		elem := slice.Index(i)
		for colName, colVal := range row {
			idx, ok := fieldMap[colName]
			if !ok {
				idx, ok = fieldMap[strings.ToUpper(colName)]
				if !ok {
					continue
				}
			}
			field := elem.Field(idx)
			if field.CanSet() {
				setFieldValue(field, colVal)
			}
		}
	}

	sliceVal.Set(slice)
	return nil
}

// setFieldValue 尝试将值设置到结构体字段中，自动处理类型转换
func setFieldValue(field reflect.Value, val interface{}) {
	if val == nil {
		return
	}

	v := reflect.ValueOf(val)
	vt := v.Type()

	// 直接可赋值
	if vt.AssignableTo(field.Type()) {
		field.Set(v)
		return
	}

	// 尝试类型转换
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		switch x := val.(type) {
		case int32:
			field.SetInt(int64(x))
		case int64:
			field.SetInt(x)
		case uint32:
			field.SetInt(int64(x))
		case uint64:
			field.SetInt(int64(x))
		case float32:
			field.SetInt(int64(x))
		case float64:
			field.SetInt(int64(x))
		}

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		switch x := val.(type) {
		case int32:
			field.SetUint(uint64(x))
		case int64:
			field.SetUint(uint64(x))
		case uint32:
			field.SetUint(uint64(x))
		case uint64:
			field.SetUint(x)
		case float32:
			field.SetUint(uint64(x))
		case float64:
			field.SetUint(uint64(x))
		}

	case reflect.Float32, reflect.Float64:
		switch x := val.(type) {
		case int32:
			field.SetFloat(float64(x))
		case int64:
			field.SetFloat(float64(x))
		case uint32:
			field.SetFloat(float64(x))
		case uint64:
			field.SetFloat(float64(x))
		case float32:
			field.SetFloat(float64(x))
		case float64:
			field.SetFloat(x)
		}

	case reflect.String:
		if s, ok := val.(string); ok {
			field.SetString(s)
		}

	case reflect.Bool:
		if b, ok := val.(bool); ok {
			field.SetBool(b)
		}
	}
}
