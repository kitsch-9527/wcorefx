//go:build windows
// +build windows

package net

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"reflect"
	"unsafe"

	win "golang.org/x/sys/windows"
)

// MIB_TCPROW_OWNER_PID 对应Windows API中的结构体
type MIB_TCPROW_OWNER_PID struct {
	DwState      uint32 // 连接状态
	DwLocalAddr  uint32 // 本地IP地址(网络字节序)
	DwLocalPort  uint32 // 本地端口(网络字节序)
	DwRemoteAddr uint32 // 远程IP地址(网络字节序)
	DwRemotePort uint32 // 远程端口(网络字节序)
	DwOwningPid  uint32 // 所属进程ID
}

// MIB_TCP6ROW_OWNER_PID 对应Windows API中的结构体
type MIB_TCP6ROW_OWNER_PID struct {
	LocalAddr     [16]byte // 本地IPv6地址(网络字节序)
	LocalScopeId  uint32   // 本地作用域ID
	LocalPort     uint32   // 本地端口(网络字节序)
	RemoteAddr    [16]byte // 远程IPv6地址(网络字节序)
	RemoteScopeId uint32   // 远程作用域ID
	RemotePort    uint32   // 远程端口(网络字节序)
	DwState       uint32   // 连接状态
	DwOwningPid   uint32   // 所属进程ID
}

// MIB_UDPROW_OWNER_PID 对应Windows API中的IPv4 UDP结构体
type MIB_UDPROW_OWNER_PID struct {
	DwLocalAddr uint32 // 本地IP地址(网络字节序)
	DwLocalPort uint32 // 本地端口(网络字节序)
	DwOwningPid uint32 // 所属进程ID
}

// MIB_UDP6ROW_OWNER_PID 对应Windows API中的IPv6 UDP结构体
type MIB_UDP6ROW_OWNER_PID struct {
	LocalAddr    [16]byte // 本地IPv6地址(网络字节序)
	LocalScopeId uint32   // 本地作用域ID
	LocalPort    uint32   // 本地端口(网络字节序)
	DwOwningPid  uint32   // 所属进程ID
}

// 常量定义
const (
	AF_INET  = 2  // IPv4地址族
	AF_INET6 = 23 // IPv6地址族

	TCP_TABLE_OWNER_PID_ALL = 5   // 包含所有拥有者PID的TCP表
	UDP_TABLE_OWNER_PID     = 1   // 包含所有拥有者PID的UDP表
	RPC_C_AUTHN_WINNT       = 10  // Windows NT身份验证
	errorInsufficientBuffer = 122 // 缓冲区不足错误码
	noError                 = 0   // 无错误
)

// TCP连接状态常量
const (
	MibTCPStateClosed      = iota + 1 // 关闭
	MibTCPStateListen                 // 监听
	MibTCPStateSynSent                // 同步已发送
	MibTCPStateSynRcvd                // 同步已接收
	MibTCPStateEstablished            // 已建立
	MibTCPStateFinWait1               // 终止等待1
	MibTCPStateFinWait2               // 终止等待2
	MibTCPStateCloseWait              // 关闭等待
	MibTCPStateClosing                // 关闭中
	MibTCPStateLastAck                // 最后确认
	MibTCPStateTimeWait               // 时间等待
	MibTCPStateDeleteTCB              // 删除传输控制块
)

// GetReadableState 将TCP状态数值转换为可读字符串
func GetReadableState(state uint32) string {
	switch state {
	case MibTCPStateClosed:
		return "CLOSED"
	case MibTCPStateListen:
		return "LISTENING"
	case MibTCPStateSynSent:
		return "SYN_SENT"
	case MibTCPStateSynRcvd:
		return "SYN_RCVD"
	case MibTCPStateEstablished:
		return "ESTABLISHED"
	case MibTCPStateFinWait1:
		return "FIN_WAIT1"
	case MibTCPStateFinWait2:
		return "FIN_WAIT2"
	case MibTCPStateCloseWait:
		return "CLOSE_WAIT"
	case MibTCPStateClosing:
		return "CLOSING"
	case MibTCPStateLastAck:
		return "LAST_ACK"
	case MibTCPStateTimeWait:
		return "TIME_WAIT"
	case MibTCPStateDeleteTCB:
		return "DELETE_TCB"
	default:
		return fmt.Sprintf("UNKNOWN(%d)", state)
	}
}

// InetNtoa 将32位网络字节序IP地址转换为点分十进制字符串
func InetNtoa(addr uint32) string {
	return net.IPv4(
		byte(addr>>24),
		byte(addr>>16),
		byte(addr>>8),
		byte(addr),
	).String()
}

// InetNtoa6 将128位网络字节序IPv6地址转换为标准格式字符串
func InetNtoa6(addr [16]byte) string {
	return net.IP(addr[:]).String()
}

// Ntohs 将网络字节序(大端)的16位端口号转换为主机字节序
func Ntohs(port uint32) uint16 {
	return uint16((port>>8)|(port<<8)) & 0xffff
}

// 安全读取MIB_TCPROW_OWNER_PID结构体
func ReadTCPRow(buffer []byte, offset uintptr) (MIB_TCPROW_OWNER_PID, error) {
	rowSize := unsafe.Sizeof(MIB_TCPROW_OWNER_PID{})
	if offset+rowSize > uintptr(len(buffer)) {
		return MIB_TCPROW_OWNER_PID{}, fmt.Errorf("缓冲区越界: 偏移量=%d, 所需大小=%d, 缓冲区大小=%d",
			offset, rowSize, len(buffer))
	}
	return *(*MIB_TCPROW_OWNER_PID)(unsafe.Pointer(&buffer[offset])), nil
}

// 安全读取MIB_TCP6ROW_OWNER_PID结构体
func ReadTCP6Row(buffer []byte, offset uintptr) (MIB_TCP6ROW_OWNER_PID, error) {
	rowSize := unsafe.Sizeof(MIB_TCP6ROW_OWNER_PID{})
	if offset+rowSize > uintptr(len(buffer)) {
		return MIB_TCP6ROW_OWNER_PID{}, fmt.Errorf("缓冲区越界: 偏移量=%d, 所需大小=%d, 缓冲区大小=%d",
			offset, rowSize, len(buffer))
	}
	return *(*MIB_TCP6ROW_OWNER_PID)(unsafe.Pointer(&buffer[offset])), nil
}

// ReadUDPRow 安全读取MIB_UDPROW_OWNER_PID结构体
func ReadUDPRow(buffer []byte, offset uintptr) (MIB_UDPROW_OWNER_PID, error) {
	rowSize := unsafe.Sizeof(MIB_UDPROW_OWNER_PID{})
	if offset+rowSize > uintptr(len(buffer)) {
		return MIB_UDPROW_OWNER_PID{}, fmt.Errorf("缓冲区越界: 偏移量=%d, 所需大小=%d, 缓冲区大小=%d",
			offset, rowSize, len(buffer))
	}
	return *(*MIB_UDPROW_OWNER_PID)(unsafe.Pointer(&buffer[offset])), nil
}

// ReadUDP6Row 安全读取MIB_UDP6ROW_OWNER_PID结构体
func ReadUDP6Row(buffer []byte, offset uintptr) (MIB_UDP6ROW_OWNER_PID, error) {
	rowSize := unsafe.Sizeof(MIB_UDP6ROW_OWNER_PID{})
	if offset+rowSize > uintptr(len(buffer)) {
		return MIB_UDP6ROW_OWNER_PID{}, fmt.Errorf("缓冲区越界: 偏移量=%d, 所需大小=%d, 缓冲区大小=%d",
			offset, rowSize, len(buffer))
	}
	return *(*MIB_UDP6ROW_OWNER_PID)(unsafe.Pointer(&buffer[offset])), nil
}

// getUdpTableBuffer 获取UDP表缓冲区（修正笔误并完善）
func getUdpTableBuffer(sort bool, af int, tableClass, reserved uint32) ([]byte, uint32, error) {
	var bufferSize uint32 = 0

	// 第一次调用获取所需缓冲区大小
	err := GetExtendedUdpTable(nil, &bufferSize, sort, uint32(af), tableClass, reserved)
	if err != nil && bufferSize == 0 {
		return nil, 0, fmt.Errorf("获取缓冲区大小失败: %w", err)
	}

	// 分配缓冲区
	buffer := make([]byte, bufferSize)
	if uint32(len(buffer)) != bufferSize {
		return nil, 0, errors.New("缓冲区分配大小不匹配")
	}

	// 第二次调用获取实际数据
	if err = GetExtendedUdpTable(&buffer[0], &bufferSize, sort, uint32(af), tableClass, reserved); err != nil {
		return nil, 0, fmt.Errorf("获取UDP表数据失败: %w", err)
	}

	// 解析条目数量（头部4字节存储条目数）
	if len(buffer) < 4 {
		return nil, 0, errors.New("缓冲区太小，无法读取条目数量")
	}
	numEntries := binary.LittleEndian.Uint32(buffer[:4])

	return buffer, numEntries, nil
}

// getTcpTableBuffer 获取TCP表缓冲区
func getTcpTableBuffer(sort bool, af int, tableClass, reserved uint32) ([]byte, uint32, error) {
	var bufferSize uint32 = 0

	// 第一次调用获取所需缓冲区大小
	err := GetExtendedTcpTable(nil, &bufferSize, sort, uint32(af), tableClass, reserved)
	if err != nil && bufferSize == 0 {
		return nil, 0, fmt.Errorf("获取缓冲区大小失败: %w", err)
	}

	// 分配缓冲区
	buffer := make([]byte, bufferSize)
	if uint32(len(buffer)) != bufferSize {
		return nil, 0, errors.New("缓冲区分配大小不匹配")
	}
	// 第二次调用获取实际数据
	if err = GetExtendedTcpTable(&buffer[0], &bufferSize, sort, uint32(af), tableClass, reserved); err != nil {
		return nil, 0, fmt.Errorf("获取TCP表数据失败: %w", err)
	}

	// 解析条目数量
	if len(buffer) < 4 {
		return nil, 0, errors.New("缓冲区太小，无法读取条目数量")
	}
	numEntries := binary.LittleEndian.Uint32(buffer[:4])

	return buffer, numEntries, nil
}

// GetTcp4Endpoints 获取IPv4 TCP连接端点信息（使用unsafe优化）
func GetTcp4Endpoints(sort bool, tableClass uint32, reserved uint32) ([]MIB_TCPROW_OWNER_PID, error) {
	buffer, numEntries, err := getTcpTableBuffer(sort, AF_INET, tableClass, reserved)
	if err != nil {
		return nil, fmt.Errorf("获取IPv4 TCP表失败: %w", err)
	}

	if numEntries == 0 {
		return []MIB_TCPROW_OWNER_PID{}, nil
	}

	rowSize := unsafe.Sizeof(MIB_TCPROW_OWNER_PID{})
	headerSize := uintptr(4) // 头部4字节存储条目数量
	totalSize := headerSize + rowSize*uintptr(numEntries)

	if uintptr(len(buffer)) < totalSize {
		return nil, fmt.Errorf("缓冲区太小，需要%d字节，实际%d字节", totalSize, len(buffer))
	}

	endpoints := make([]MIB_TCPROW_OWNER_PID, 0, numEntries)
	for i := uint32(0); i < numEntries; i++ {
		offset := headerSize + rowSize*uintptr(i)
		row, err := ReadTCPRow(buffer, offset)
		if err != nil {
			return nil, fmt.Errorf("读取条目%d失败: %w", i, err)
		}
		endpoints = append(endpoints, row)
	}

	return endpoints, nil
}

// GetTcp6Endpoints 获取IPv6 TCP连接端点信息（使用unsafe优化）
func GetTcp6Endpoints(sort bool, tableClass uint32, reserved uint32) ([]MIB_TCP6ROW_OWNER_PID, error) {
	buffer, numEntries, err := getTcpTableBuffer(sort, AF_INET6, tableClass, reserved)
	if err != nil {
		return nil, fmt.Errorf("获取IPv6 TCP表失败: %w", err)
	}

	if numEntries == 0 {
		return []MIB_TCP6ROW_OWNER_PID{}, nil
	}

	rowSize := unsafe.Sizeof(MIB_TCP6ROW_OWNER_PID{})
	headerSize := uintptr(4) // 头部4字节存储条目数量
	totalSize := headerSize + rowSize*uintptr(numEntries)

	if uintptr(len(buffer)) < totalSize {
		return nil, fmt.Errorf("缓冲区太小，需要%d字节，实际%d字节", totalSize, len(buffer))
	}

	endpoints := make([]MIB_TCP6ROW_OWNER_PID, 0, numEntries)
	for i := uint32(0); i < numEntries; i++ {
		offset := headerSize + rowSize*uintptr(i)
		row, err := ReadTCP6Row(buffer, offset)
		if err != nil {
			return nil, fmt.Errorf("读取条目%d失败: %w", i, err)
		}
		endpoints = append(endpoints, row)
	}

	return endpoints, nil
}

// GetUdp4Endpoints 获取IPv4 UDP连接端点信息
func GetUdp4Endpoints(sort bool, tableClass uint32, reserved uint32) ([]MIB_UDPROW_OWNER_PID, error) {
	// 调用缓冲区函数获取IPv4 UDP表数据
	buffer, numEntries, err := getUdpTableBuffer(sort, AF_INET, tableClass, reserved)
	if err != nil {
		return nil, fmt.Errorf("获取IPv4 UDP表失败: %w", err)
	}

	if numEntries == 0 {
		return []MIB_UDPROW_OWNER_PID{}, nil
	}

	// 计算单个条目大小及总所需大小
	rowSize := unsafe.Sizeof(MIB_UDPROW_OWNER_PID{})
	headerSize := uintptr(4) // 头部4字节为条目数
	totalSize := headerSize + rowSize*uintptr(numEntries)

	if uintptr(len(buffer)) < totalSize {
		return nil, fmt.Errorf("缓冲区太小，需要%d字节，实际%d字节", totalSize, len(buffer))
	}

	// 循环读取所有条目
	endpoints := make([]MIB_UDPROW_OWNER_PID, 0, numEntries)
	for i := uint32(0); i < numEntries; i++ {
		offset := headerSize + rowSize*uintptr(i)
		row, err := ReadUDPRow(buffer, offset)
		if err != nil {
			return nil, fmt.Errorf("读取条目%d失败: %w", i, err)
		}
		endpoints = append(endpoints, row)
	}

	return endpoints, nil
}

// GetUdp6Endpoints 获取IPv6 UDP连接端点信息
func GetUdp6Endpoints(sort bool, tableClass uint32, reserved uint32) ([]MIB_UDP6ROW_OWNER_PID, error) {
	// 调用缓冲区函数获取IPv6 UDP表数据
	buffer, numEntries, err := getUdpTableBuffer(sort, AF_INET6, tableClass, reserved)
	if err != nil {
		return nil, fmt.Errorf("获取IPv6 UDP表失败: %w", err)
	}

	if numEntries == 0 {
		return []MIB_UDP6ROW_OWNER_PID{}, nil
	}

	// 计算单个条目大小及总所需大小
	rowSize := unsafe.Sizeof(MIB_UDP6ROW_OWNER_PID{})
	headerSize := uintptr(4) // 头部4字节为条目数
	totalSize := headerSize + rowSize*uintptr(numEntries)

	if uintptr(len(buffer)) < totalSize {
		return nil, fmt.Errorf("缓冲区太小，需要%d字节，实际%d字节", totalSize, len(buffer))
	}

	// 循环读取所有条目
	endpoints := make([]MIB_UDP6ROW_OWNER_PID, 0, numEntries)
	for i := uint32(0); i < numEntries; i++ {
		offset := headerSize + rowSize*uintptr(i)
		row, err := ReadUDP6Row(buffer, offset)
		if err != nil {
			return nil, fmt.Errorf("读取条目%d失败: %w", i, err)
		}
		endpoints = append(endpoints, row)
	}

	return endpoints, nil
}

// CalloutInfo 存储标注的关键信息
type CalloutInfo struct {
	CalloutId   uint32
	CalloutKey  win.GUID
	Name        string
	Description string
}
type FwpmCallout = struct {
	CalloutKey   win.GUID
	Name         string
	Description  string
	Flags        fwpmFilterFlags
	ProviderKey  *win.GUID
	ProviderData fwpByteBlob
	LayerKey     LayerID
	CalloutId    uint32
}

// callouts 全局变量，存储标注信息
func EnumWfpCallouts() ([]FwpmCallout, error) {
	var (
		callouts   []*fwpmCallout0
		numEntries uint32
		array      **fwpmCallout0
	)
	ssion := fwpmSession0{
		DisplayData: fwpmDisplayData0{Name: win.StringToUTF16Ptr("WFPSampler's User Mode Session")},
		Flags:       0,
	}
	engineH, err := FwpmEngineOpen(nil, uint32(RPC_C_AUTHN_WINNT), nil, &ssion)
	if err != nil {
		return nil, fmt.Errorf("FwpmEngineOpen failed: %w", err)
	}
	defer FwpmEngineClose(engineH)
	enumH, err := FwpmCalloutCreateEnumHandle(engineH, nil)
	if err != nil {
		return nil, fmt.Errorf("FwpmCalloutCreateEnumHandle failed: %w", err)
	}
	defer FwpmCalloutDestroyEnumHandle(engineH, enumH)
	err = FwpmCalloutEnum(engineH, enumH, 0xFFFFFFFF, &array, &numEntries)
	if err != nil {
		return nil, fmt.Errorf("FwpmCalloutEnum failed: %w", err)
	}
	if numEntries == 0 {
		return nil, nil
	}
	sh := (*reflect.SliceHeader)(unsafe.Pointer(&callouts))
	sh.Cap = int(numEntries)
	sh.Len = int(numEntries)
	sh.Data = uintptr(unsafe.Pointer(array))
	defer FwpmFreeMemory((*struct{})(unsafe.Pointer(&array)))
	var infos []FwpmCallout
	for _, callout := range callouts {
		info := FwpmCallout{
			CalloutKey:   callout.CalloutKey,
			CalloutId:    callout.CalloutId,
			Name:         win.UTF16PtrToString(callout.DisplayData.Name),
			Description:  win.UTF16PtrToString(callout.DisplayData.Description),
			Flags:        callout.Flags,
			ProviderKey:  callout.ProviderKey,
			ProviderData: callout.ProviderData,
			LayerKey:     callout.LayerKey,
		}
		infos = append(infos, info)

	}
	return infos, nil
}

type FwpmFilter struct {
	FilterKey           win.GUID
	Name                string
	Description         string
	Flags               fwpmFilterFlags
	ProviderKey         *win.GUID
	ProviderData        fwpByteBlob
	LayerKey            LayerID
	SublayerKey         SublayerID
	Weight              fwpValue0
	NumFilterConditions uint32
	FilterConditions    *fwpmFilterCondition0
	Action              fwpmAction0
	// Only one of RawContext/ProviderContextKey must be set.
	RawContext         uint64
	ProviderContextKey win.GUID
	Reserved           *win.GUID
	FilterID           uint64
	EffectiveWeight    fwpValue0
}

// callouts 全局变量，存储标注信息
func EnumWfpFilters() ([]FwpmFilter, error) {

	var (
		filters    []*fwpmFilter0
		numEntries uint32
		array      **fwpmFilter0
	)
	ssion := fwpmSession0{
		DisplayData: fwpmDisplayData0{Name: win.StringToUTF16Ptr("WFPSampler's User Mode Session")},
		Flags:       0,
	}
	//var ssion FWPM_SESSION0
	engineH, err := FwpmEngineOpen(nil, uint32(RPC_C_AUTHN_WINNT), nil, &ssion)
	if err != nil {
		return nil, fmt.Errorf("FwpmEngineOpen failed: %w", err)
	}
	defer FwpmEngineClose(engineH)
	enumH, err := FwpmFilterCreateEnumHandle(engineH, nil)
	if err != nil {
		return nil, fmt.Errorf("FwpmFilterCreateEnumHandle failed: %w", err)
	}
	defer FwpmFilterDestroyEnumHandle(engineH, enumH)

	err = FwpmFilterEnum(engineH, enumH, 0xFFFFFFFF, &array, &numEntries)
	if err != nil {
		return nil, fmt.Errorf("FwpmFilterEnum failed: %w", err)
	}

	if numEntries == 0 {
		return nil, nil
	}

	sh := (*reflect.SliceHeader)(unsafe.Pointer(&filters))
	sh.Cap = int(numEntries)
	sh.Len = int(numEntries)
	sh.Data = uintptr(unsafe.Pointer(array))
	defer FwpmFreeMemory((*struct{})(unsafe.Pointer(&array)))
	var infos []FwpmFilter
	for _, filter := range filters {
		info := FwpmFilter{
			FilterKey:   filter.FilterKey,
			Name:        win.UTF16PtrToString(filter.DisplayData.Name),
			Description: win.UTF16PtrToString(filter.DisplayData.Description),
			//DisplayData:         filter.DisplayData,
			Flags:               filter.Flags,
			ProviderKey:         filter.ProviderKey,
			ProviderData:        filter.ProviderData,
			LayerKey:            filter.LayerKey,
			SublayerKey:         filter.SublayerKey,
			Weight:              filter.Weight,
			NumFilterConditions: filter.NumFilterConditions,
			FilterConditions:    filter.FilterConditions,
			Action:              filter.Action,
			RawContext:          filter.RawContext,
			ProviderContextKey:  filter.ProviderContextKey,
			Reserved:            filter.Reserved,
			FilterID:            filter.FilterID,
		}
		infos = append(infos, info)

	}
	return infos, nil

}
