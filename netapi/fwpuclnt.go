//go:build windows

package netapi

import (
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var modfwpmu = windows.NewLazySystemDLL("fwpuclnt.dll")

var (
	procFwpmEngineOpen               = modfwpmu.NewProc("FwpmEngineOpen0")
	procFwpmEngineClose              = modfwpmu.NewProc("FwpmEngineClose0")
	procFwpmFreeMemory               = modfwpmu.NewProc("FwpmFreeMemory0")
	procFwpmCalloutCreateEnumHandle  = modfwpmu.NewProc("FwpmCalloutCreateEnumHandle0")
	procFwpmCalloutDestroyEnumHandle = modfwpmu.NewProc("FwpmCalloutDestroyEnumHandle0")
	procFwpmCalloutEnum              = modfwpmu.NewProc("FwpmCalloutEnum0")
	procFwpmCalloutGetByKey          = modfwpmu.NewProc("FwpmCalloutGetByKey0")
	procFwpmFilterCreateEnumHandle   = modfwpmu.NewProc("FwpmFilterCreateEnumHandle0")
	procFwpmFilterDestroyEnumHandle  = modfwpmu.NewProc("FwpmFilterDestroyEnumHandle0")
	procFwpmFilterEnum               = modfwpmu.NewProc("FwpmFilterEnum0")
	procFwpmFilterGetByKey           = modfwpmu.NewProc("FwpmFilterGetByKey0")
)

// FwpmDisplayData0 对应 FWPM_DISPLAY_DATA0 结构，包含名称和描述信息。
type FwpmDisplayData0 struct {
	// Name 显示名称。
	Name        *uint16
	// Description 显示描述。
	Description *uint16
}

// FwpmSession0 对应 FWPM_SESSION0 结构，表示 WFP 引擎会话。
type FwpmSession0 struct {
	// SessionKey 会话键。
	SessionKey           windows.GUID
	// DisplayData 会话显示数据。
	DisplayData          FwpmDisplayData0
	// Flags 会话标志。
	Flags                uint32
	// TxnWaitTimeoutMillis 事务等待超时时间（毫秒）。
	TxnWaitTimeoutMillis uint32
	// ProcessID 会话进程 ID。
	ProcessID            uint32
	// SID 会话用户 SID。
	SID                  *windows.SID
	// Username 会话用户名。
	Username             *uint16
	// KernelMode 是否为内核模式。
	KernelMode           uint8
}

// FwpByteBlob 对应 FWP_BYTE_BLOB 结构，包含字节数据及其大小。
type FwpByteBlob struct {
	// Size 数据大小（字节）。
	Size uint32
	// Data 数据指针。
	Data *uint8
}

// LayerID 标识 WFP 层。
type LayerID windows.GUID

// SublayerID 标识 WFP 子层。
type SublayerID windows.GUID

// FieldID 标识 WFP 层字段。
type FieldID windows.GUID

// RuleID 标识 WFP 规则。
type RuleID windows.GUID

// ProviderID 标识 WFP 提供商。
type ProviderID windows.GUID

// FwpmFilterFlags 定义 WFP 过滤器标志位。
type FwpmFilterFlags uint32

const (
	// FwpmFilterFlagsPersistent 过滤器持久化，重启后保留。
	FwpmFilterFlagsPersistent              FwpmFilterFlags = 1 << iota
	// FwpmFilterFlagsBootTime 过滤器在启动时加载。
	FwpmFilterFlagsBootTime
	// FwpmFilterFlagsHasProviderContext 过滤器关联提供商上下文。
	FwpmFilterFlagsHasProviderContext
	// FwpmFilterFlagsClearActionRight 清除操作权限位。
	FwpmFilterFlagsClearActionRight
	// FwpmFilterFlagsPermitIfCalloutUnregistered 标注未注册时放行。
	FwpmFilterFlagsPermitIfCalloutUnregistered
	// FwpmFilterFlagsDisabled 过滤器已禁用。
	FwpmFilterFlagsDisabled
	// FwpmFilterFlagsIndexed 过滤器已建立索引。
	FwpmFilterFlagsIndexed
)

// Action 是过滤引擎可执行的动作类型。
type Action uint32

const (
	// ActionBlock 阻止流量。
	ActionBlock              Action = 0x1001
	// ActionPermit 允许流量。
	ActionPermit             Action = 0x1002
	// ActionCalloutTerminating 终止型标注处理。
	ActionCalloutTerminating Action = 0x5003
	// ActionCalloutInspection 检查型标注处理。
	ActionCalloutInspection  Action = 0x6004
	// ActionCalloutUnknown 未知类型的标注处理。
	ActionCalloutUnknown     Action = 0x4005
)

// FwpmAction0 对应 FWPM_ACTION0 结构，描述过滤动作。
type FwpmAction0 struct {
	// Type 动作类型。
	Type Action
	// GUID 动作 GUID。
	GUID windows.GUID
}

// DataType 描述 FwpValue0 中的数据类型。
type DataType uint32

const (
	// DataTypeEmpty 空数据类型。
	DataTypeEmpty                  DataType = 0
	// DataTypeUint8 8 位无符号整数。
	DataTypeUint8                  DataType = 1
	// DataTypeUint16 16 位无符号整数。
	DataTypeUint16                 DataType = 2
	// DataTypeUint32 32 位无符号整数。
	DataTypeUint32                 DataType = 3
	// DataTypeUint64 64 位无符号整数。
	DataTypeUint64                 DataType = 4
	// DataTypeByteArray16 16 字节数组。
	DataTypeByteArray16            DataType = 11
	// DataTypeByteBlob 字节块。
	DataTypeByteBlob               DataType = 12
	// DataTypeSID 安全标识符（SID）。
	DataTypeSID                    DataType = 13
	// DataTypeSecurityDescriptor 安全描述符。
	DataTypeSecurityDescriptor     DataType = 14
	// DataTypeTokenInformation 令牌信息。
	DataTypeTokenInformation       DataType = 15
	// DataTypeTokenAccessInformation 令牌访问信息。
	DataTypeTokenAccessInformation DataType = 16
	// DataTypeV4AddrMask IPv4 地址掩码。
	DataTypeV4AddrMask             DataType = 256
	// DataTypeV6AddrMask IPv6 地址掩码。
	DataTypeV6AddrMask             DataType = 257
	// DataTypeRange 范围数据类型。
	DataTypeRange                  DataType = 258
)

// FwpValue0 对应 FWP_VALUE0 结构，包含数据和类型。
type FwpValue0 struct {
	// Type 数据类型。
	Type  DataType
	// Value 数据值（通用指针）。
	Value uintptr
}

// MatchType 是用于字段测试的匹配运算符。
type MatchType uint32

const (
	// MatchTypeEqual 相等匹配（==）。
	MatchTypeEqual MatchType = iota
	// MatchTypeGreater 大于匹配（>）。
	MatchTypeGreater
	// MatchTypeLess 小于匹配（<）。
	MatchTypeLess
	// MatchTypeGreaterOrEqual 大于等于匹配（>=）。
	MatchTypeGreaterOrEqual
	// MatchTypeLessOrEqual 小于等于匹配（<=）。
	MatchTypeLessOrEqual
	// MatchTypeRange 范围匹配（in）。
	MatchTypeRange
	// MatchTypeFlagsAllSet 所有标志位已设置。
	MatchTypeFlagsAllSet
	// MatchTypeFlagsAnySet 任意标志位已设置。
	MatchTypeFlagsAnySet
	// MatchTypeFlagsNoneSet 无标志位设置。
	MatchTypeFlagsNoneSet
	// MatchTypeEqualCaseInsensitive 不区分大小写的相等匹配。
	MatchTypeEqualCaseInsensitive
	// MatchTypeNotEqual 不等匹配（!=）。
	MatchTypeNotEqual
	// MatchTypePrefix 前缀匹配。
	MatchTypePrefix
	// MatchTypeNotPrefix 非前缀匹配。
	MatchTypeNotPrefix
)

// String 将 MatchType 运算符转换为字符串表示。
//   返回 - 匹配运算符的字符串表示
func (m MatchType) String() string {
	switch m {
	case MatchTypeEqual:
		return "=="
	case MatchTypeGreater:
		return ">"
	case MatchTypeLess:
		return "<"
	case MatchTypeGreaterOrEqual:
		return ">="
	case MatchTypeLessOrEqual:
		return "<="
	case MatchTypeRange:
		return "in"
	case MatchTypeFlagsAllSet:
		return "F[all]"
	case MatchTypeFlagsAnySet:
		return "F[any]"
	case MatchTypeFlagsNoneSet:
		return "F[none]"
	case MatchTypeEqualCaseInsensitive:
		return "i=="
	case MatchTypeNotEqual:
		return "!="
	case MatchTypePrefix:
		return "pfx"
	case MatchTypeNotPrefix:
		return "!pfx"
	default:
		return fmt.Sprintf("MatchType(%d)", m)
	}
}

// Match 是针对 WFP 层字段的匹配测试。
type Match struct {
	// Field 待匹配的字段标识。
	Field FieldID
	// Op 匹配运算符。
	Op    MatchType
	// Value 匹配值。
	Value any
}

// String 将 Match 格式化为可读的字符串描述。
//   返回 - 匹配条件的可读字符串描述
func (m Match) String() string {
	return fmt.Sprintf("%v %s %v (%T)", m.Field, m.Op, m.Value, m.Value)
}

// FwpmFilterCondition0 对应 FWPM_FILTER_CONDITION0 结构，描述过滤条件。
type FwpmFilterCondition0 struct {
	// FieldKey 字段标识。
	FieldKey  FieldID
	// MatchType 匹配运算符。
	MatchType MatchType
	// Value 匹配值。
	Value     struct {
		Type  DataType
		Value uintptr
	}
}

// FwpmCallout0 对应 FWPM_CALLOUT0 结构，描述 WFP 标注对象。
type FwpmCallout0 struct {
	// CalloutKey 标注 GUID 键。
	CalloutKey   windows.GUID
	// DisplayData 标注显示数据。
	DisplayData  FwpmDisplayData0
	// Flags 标注标志。
	Flags        FwpmFilterFlags
	// ProviderKey 提供商键（可选）。
	ProviderKey  *windows.GUID
	// ProviderData 提供商数据。
	ProviderData FwpByteBlob
	// LayerKey 层标识。
	LayerKey     LayerID
	// CalloutId 标注 ID。
	CalloutId    uint32
}

// FwpmFilter0 对应 FWPM_FILTER0 结构，描述 WFP 过滤器对象。
type FwpmFilter0 struct {
	// FilterKey 过滤器 GUID 键。
	FilterKey           windows.GUID
	// DisplayData 过滤器显示数据。
	DisplayData         FwpmDisplayData0
	// Flags 过滤器标志。
	Flags               FwpmFilterFlags
	// ProviderKey 提供商键（可选）。
	ProviderKey         *windows.GUID
	// ProviderData 提供商数据。
	ProviderData        FwpByteBlob
	// LayerKey 层标识。
	LayerKey            LayerID
	// SublayerKey 子层标识。
	SublayerKey         SublayerID
	// Weight 过滤器权重。
	Weight              FwpValue0
	// NumFilterConditions 过滤条件数量。
	NumFilterConditions uint32
	// FilterConditions 过滤条件数组指针。
	FilterConditions    *FwpmFilterCondition0
	// Action 过滤动作。
	Action              FwpmAction0
	// RawContext 原始上下文。
	RawContext          uint64
	// ProviderContextKey 提供商上下文键。
	ProviderContextKey  windows.GUID
	// Reserved 保留字段。
	Reserved            *windows.GUID
	// FilterID 过滤器 ID。
	FilterID            uint64
	// EffectiveWeight 有效权重。
	EffectiveWeight     FwpValue0
}

// FwpmEngineOpen 打开与 WFP 引擎的连接。
//   serverName - 服务器名称（nil 表示本地引擎）
//   authnService - 身份验证服务类型（10 表示 RPC_C_AUTHN_WINNT）
//   authIdentity - 身份验证身份信息（可选）
//   session - WFP 会话信息
//   返回1 - WFP 引擎句柄
//   返回2 - 错误信息
func FwpmEngineOpen(serverName *uint16, authnService uint32, authIdentity *byte, session *FwpmSession0) (windows.Handle, error) {
	var engineHandle windows.Handle
	r0, _, _ := syscall.SyscallN(procFwpmEngineOpen.Addr(),
		uintptr(unsafe.Pointer(serverName)),
		uintptr(authnService),
		uintptr(unsafe.Pointer(authIdentity)),
		uintptr(unsafe.Pointer(session)),
		uintptr(unsafe.Pointer(&engineHandle)),
	)
	if r0 != 0 {
		return engineHandle, syscall.Errno(r0)
	}
	return engineHandle, nil
}

// FwpmEngineClose 关闭与 WFP 引擎的连接。
//   engineHandle - WFP 引擎句柄
//   返回 - 错误信息
func FwpmEngineClose(engineHandle windows.Handle) error {
	r1, _, _ := syscall.SyscallN(procFwpmEngineClose.Addr(), uintptr(engineHandle))
	if r1 != 0 {
		return syscall.Errno(r1)
	}
	return nil
}

// FwpmFreeMemory 释放 WFP 分配的内存。
//   p - 待释放的内存指针
func FwpmFreeMemory(p unsafe.Pointer) {
	syscall.SyscallN(procFwpmFreeMemory.Addr(), uintptr(p))
}

// FwpmCalloutCreateEnumHandle 创建 WFP 标注枚举句柄。
//   engineHandle - WFP 引擎句柄
//   enumTemplate - 枚举模板（nil 表示枚举所有标注）
//   返回1 - 枚举句柄
//   返回2 - 错误信息
func FwpmCalloutCreateEnumHandle(engineHandle windows.Handle, enumTemplate *byte) (windows.Handle, error) {
	var enumHandle windows.Handle
	r0, _, _ := syscall.SyscallN(procFwpmCalloutCreateEnumHandle.Addr(),
		uintptr(engineHandle),
		uintptr(unsafe.Pointer(enumTemplate)),
		uintptr(unsafe.Pointer(&enumHandle)),
	)
	if r0 != 0 {
		return enumHandle, syscall.Errno(r0)
	}
	return enumHandle, nil
}

// FwpmCalloutDestroyEnumHandle 销毁 WFP 标注枚举句柄。
//   engineHandle - WFP 引擎句柄
//   enumHandle - 待销毁的枚举句柄
//   返回 - 错误信息
func FwpmCalloutDestroyEnumHandle(engineHandle, enumHandle windows.Handle) error {
	r1, _, _ := syscall.SyscallN(procFwpmCalloutDestroyEnumHandle.Addr(),
		uintptr(engineHandle), uintptr(enumHandle))
	if r1 != 0 {
		return syscall.Errno(r1)
	}
	return nil
}

// FwpmCalloutEnum 枚举 WFP 标注。
//   engineHandle - WFP 引擎句柄
//   enumHandle - 枚举句柄
//   numEntries - 请求的最大返回条目数
//   entries - 接收返回的标注条目数组指针
//   numReturned - 接收实际返回的条目数
//   返回 - 错误信息
func FwpmCalloutEnum(engineHandle, enumHandle windows.Handle, numEntries uint32, entries ***FwpmCallout0, numReturned *uint32) error {
	r0, _, _ := syscall.SyscallN(procFwpmCalloutEnum.Addr(),
		uintptr(engineHandle),
		uintptr(enumHandle),
		uintptr(numEntries),
		uintptr(unsafe.Pointer(entries)),
		uintptr(unsafe.Pointer(numReturned)),
	)
	if r0 != 0 {
		return syscall.Errno(r0)
	}
	return nil
}

// FwpmCalloutGetByKey 通过 GUID 键获取 WFP 标注。
//   engineHandle - WFP 引擎句柄
//   key - 标注 GUID 键
//   callout - 接收返回的标注对象指针
//   返回 - 错误信息
func FwpmCalloutGetByKey(engineHandle windows.Handle, key *windows.GUID, callout ***FwpmCallout0) error {
	r0, _, _ := syscall.SyscallN(procFwpmCalloutGetByKey.Addr(),
		uintptr(engineHandle),
		uintptr(unsafe.Pointer(key)),
		uintptr(unsafe.Pointer(callout)),
	)
	if r0 != 0 {
		return syscall.Errno(r0)
	}
	return nil
}

// FwpmFilterCreateEnumHandle 创建 WFP 过滤器枚举句柄。
//   engineHandle - WFP 引擎句柄
//   enumTemplate - 枚举模板（nil 表示枚举所有过滤器）
//   返回1 - 枚举句柄
//   返回2 - 错误信息
func FwpmFilterCreateEnumHandle(engineHandle windows.Handle, enumTemplate *byte) (windows.Handle, error) {
	var enumHandle windows.Handle
	r0, _, _ := syscall.SyscallN(procFwpmFilterCreateEnumHandle.Addr(),
		uintptr(engineHandle),
		uintptr(unsafe.Pointer(enumTemplate)),
		uintptr(unsafe.Pointer(&enumHandle)),
	)
	if r0 != 0 {
		return enumHandle, syscall.Errno(r0)
	}
	return enumHandle, nil
}

// FwpmFilterDestroyEnumHandle 销毁 WFP 过滤器枚举句柄。
//   engineHandle - WFP 引擎句柄
//   enumHandle - 待销毁的枚举句柄
//   返回 - 错误信息
func FwpmFilterDestroyEnumHandle(engineHandle, enumHandle windows.Handle) error {
	r1, _, _ := syscall.SyscallN(procFwpmFilterDestroyEnumHandle.Addr(),
		uintptr(engineHandle), uintptr(enumHandle))
	if r1 != 0 {
		return syscall.Errno(r1)
	}
	return nil
}

// FwpmFilterEnum 枚举 WFP 过滤器。
//   engineHandle - WFP 引擎句柄
//   enumHandle - 枚举句柄
//   numEntries - 请求的最大返回条目数
//   entries - 接收返回的过滤器条目数组指针
//   numReturned - 接收实际返回的条目数
//   返回 - 错误信息
func FwpmFilterEnum(engineHandle, enumHandle windows.Handle, numEntries uint32, entries ***FwpmFilter0, numReturned *uint32) error {
	r0, _, _ := syscall.SyscallN(procFwpmFilterEnum.Addr(),
		uintptr(engineHandle),
		uintptr(enumHandle),
		uintptr(numEntries),
		uintptr(unsafe.Pointer(entries)),
		uintptr(unsafe.Pointer(numReturned)),
	)
	if r0 != 0 {
		return syscall.Errno(r0)
	}
	return nil
}

// FwpmFilterGetByKey 通过 GUID 键获取 WFP 过滤器。
//   engineHandle - WFP 引擎句柄
//   key - 过滤器 GUID 键
//   filter - 接收返回的过滤器对象指针
//   返回 - 错误信息
func FwpmFilterGetByKey(engineHandle windows.Handle, key *windows.GUID, filter ***FwpmFilter0) error {
	r0, _, _ := syscall.SyscallN(procFwpmFilterGetByKey.Addr(),
		uintptr(engineHandle),
		uintptr(unsafe.Pointer(key)),
		uintptr(unsafe.Pointer(filter)),
	)
	if r0 != 0 {
		return syscall.Errno(r0)
	}
	return nil
}
