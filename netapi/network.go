//go:build windows

package netapi

import (
	"encoding/binary"
	"fmt"
	"net"
	"reflect"
	"unsafe"

	"golang.org/x/sys/windows"
)

// MIB_TCPROW_OWNER_PID 表示 MIB_TCPROW_OWNER_PID 结构，包含 TCP 连接信息及所属进程 PID。
type MIB_TCPROW_OWNER_PID struct {
	// DwState TCP 状态。
	DwState      uint32
	// DwLocalAddr 本地 IPv4 地址（网络字节序）。
	DwLocalAddr  uint32
	// DwLocalPort 本地端口（网络字节序）。
	DwLocalPort  uint32
	// DwRemoteAddr 远端 IPv4 地址（网络字节序）。
	DwRemoteAddr uint32
	// DwRemotePort 远端端口（网络字节序）。
	DwRemotePort uint32
	// DwOwningPid 拥有进程的 PID。
	DwOwningPid  uint32
}

// MIB_TCP6ROW_OWNER_PID 表示 MIB_TCP6ROW_OWNER_PID 结构，包含 IPv6 TCP 连接信息及所属进程 PID。
type MIB_TCP6ROW_OWNER_PID struct {
	// LocalAddr 本地 IPv6 地址。
	LocalAddr     [16]byte
	// LocalScopeId 本地作用域 ID。
	LocalScopeId  uint32
	// LocalPort 本地端口。
	LocalPort     uint32
	// RemoteAddr 远端 IPv6 地址。
	RemoteAddr    [16]byte
	// RemoteScopeId 远端作用域 ID。
	RemoteScopeId uint32
	// RemotePort 远端端口。
	RemotePort    uint32
	// DwState TCP 状态。
	DwState       uint32
	// DwOwningPid 拥有进程的 PID。
	DwOwningPid   uint32
}

// MIB_UDPROW_OWNER_PID 表示 MIB_UDPROW_OWNER_PID 结构，包含 UDP 端点信息及所属进程 PID。
type MIB_UDPROW_OWNER_PID struct {
	// DwLocalAddr 本地 IPv4 地址（网络字节序）。
	DwLocalAddr uint32
	// DwLocalPort 本地端口（网络字节序）。
	DwLocalPort uint32
	// DwOwningPid 拥有进程的 PID。
	DwOwningPid uint32
}

// MIB_UDP6ROW_OWNER_PID 表示 MIB_UDP6ROW_OWNER_PID 结构，包含 IPv6 UDP 端点信息及所属进程 PID。
type MIB_UDP6ROW_OWNER_PID struct {
	// LocalAddr 本地 IPv6 地址。
	LocalAddr    [16]byte
	// LocalScopeId 本地作用域 ID。
	LocalScopeId uint32
	// LocalPort 本地端口。
	LocalPort    uint32
	// DwOwningPid 拥有进程的 PID。
	DwOwningPid  uint32
}

const (
	// AF_INET IPv4 地址族常量。
	AF_INET                  = 2
	// AF_INET6 IPv6 地址族常量。
	AF_INET6                 = 23
	// TCP_TABLE_OWNER_PID_ALL TCP 所有连接表标识。
	TCP_TABLE_OWNER_PID_ALL  = 5
	// UDP_TABLE_OWNER_PID UDP 监听表标识。
	UDP_TABLE_OWNER_PID      = 1
	_ = 122 // errorInsufficientBuffer
)

// MibTCPState 常量定义 TCP 连接状态枚举值。
const (
	// MibTCPStateClosed TCP 连接已关闭。
	MibTCPStateClosed      = iota + 1
	// MibTCPStateListen TCP 正在监听。
	MibTCPStateListen
	// MibTCPStateSynSent TCP 已发送 SYN。
	MibTCPStateSynSent
	// MibTCPStateSynRcvd TCP 已收到 SYN。
	MibTCPStateSynRcvd
	// MibTCPStateEstablished TCP 连接已建立。
	MibTCPStateEstablished
	// MibTCPStateFinWait1 TCP 正在 FIN-WAIT-1 状态。
	MibTCPStateFinWait1
	// MibTCPStateFinWait2 TCP 正在 FIN-WAIT-2 状态。
	MibTCPStateFinWait2
	// MibTCPStateCloseWait TCP 正在 CLOSE-WAIT 状态。
	MibTCPStateCloseWait
	// MibTCPStateClosing TCP 正在 CLOSING 状态。
	MibTCPStateClosing
	// MibTCPStateLastAck TCP 正在 LAST-ACK 状态。
	MibTCPStateLastAck
	// MibTCPStateTimeWait TCP 正在 TIME-WAIT 状态。
	MibTCPStateTimeWait
	// MibTCPStateDeleteTCB TCP 正在删除 TCB。
	MibTCPStateDeleteTCB
)

// TCPState 将 TCP 状态数值转换为可读的字符串描述。
//   state - TCP 状态数值
//   返回 - TCP 状态对应的字符串描述
func TCPState(state uint32) string {
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

// InetNtoa 将 32 位网络字节序的 IPv4 地址转换为字符串。
//   addr - 网络字节序的 32 位 IPv4 地址
//   返回 - IPv4 地址字符串
func InetNtoa(addr uint32) string {
	return net.IPv4(
		byte(addr>>24),
		byte(addr>>16),
		byte(addr>>8),
		byte(addr),
	).String()
}

// InetNtoa6 将 128 位网络字节序的 IPv6 地址转换为字符串。
//   addr - 网络字节序的 128 位 IPv6 地址
//   返回 - IPv6 地址字符串
func InetNtoa6(addr [16]byte) string {
	return net.IP(addr[:]).String()
}

// Ntohs 将网络字节序（大端）的 16 位端口值转换为主机字节序。
//   port - 网络字节序的 16 位端口值（存储在 uint32 低位）
//   返回 - 主机字节序的端口值
func Ntohs(port uint32) uint16 {
	return uint16((port>>8)|(port<<8)) & 0xffff
}

// CalloutInfo 存储 WFP 标注信息。
type CalloutInfo struct {
	// CalloutId 标注 ID。
	CalloutId   uint32
	// CalloutKey 标注 GUID 键。
	CalloutKey  windows.GUID
	// Name 标注名称。
	Name        string
	// Description 标注描述。
	Description string
}

// FwpmCallout 存储完整的 WFP 标注数据。
type FwpmCallout struct {
	// CalloutKey 标注 GUID 键。
	CalloutKey   windows.GUID
	// Name 标注名称。
	Name         string
	// Description 标注描述。
	Description  string
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

// FwpmFilter 存储完整的 WFP 过滤器数据。
type FwpmFilter struct {
	// FilterKey 过滤器 GUID 键。
	FilterKey           windows.GUID
	// Name 过滤器名称。
	Name                string
	// Description 过滤器描述。
	Description         string
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

func getTcpTableBuffer(sort bool, af int, tableClass, reserved uint32) ([]byte, uint32, error) {
	var bufSize uint32
	err := getExtendedTcpTable(nil, &bufSize, sort, uint32(af), tableClass, reserved)
	if err != nil && bufSize == 0 {
		return nil, 0, fmt.Errorf("get tcp table size: %w", err)
	}
	buf := make([]byte, bufSize)
	err = getExtendedTcpTable(&buf[0], &bufSize, sort, uint32(af), tableClass, reserved)
	if err != nil {
		return nil, 0, fmt.Errorf("get tcp table: %w", err)
	}
	if len(buf) < 4 {
		return nil, 0, fmt.Errorf("buffer too small for header")
	}
	return buf, binary.LittleEndian.Uint32(buf[:4]), nil
}

func getUdpTableBuffer(sort bool, af int, tableClass, reserved uint32) ([]byte, uint32, error) {
	var bufSize uint32
	err := getExtendedUdpTable(nil, &bufSize, sort, uint32(af), tableClass, reserved)
	if err != nil && bufSize == 0 {
		return nil, 0, fmt.Errorf("get udp table size: %w", err)
	}
	buf := make([]byte, bufSize)
	err = getExtendedUdpTable(&buf[0], &bufSize, sort, uint32(af), tableClass, reserved)
	if err != nil {
		return nil, 0, fmt.Errorf("get udp table: %w", err)
	}
	if len(buf) < 4 {
		return nil, 0, fmt.Errorf("buffer too small for header")
	}
	return buf, binary.LittleEndian.Uint32(buf[:4]), nil
}

func readRows[T any](buffer []byte, numEntries uint32) ([]T, error) {
	if numEntries == 0 {
		return nil, nil
	}
	var row T
	rowSize := unsafe.Sizeof(row)
	headerSize := uintptr(4)
	totalSize := headerSize + rowSize*uintptr(numEntries)
	if uintptr(len(buffer)) < totalSize {
		return nil, fmt.Errorf("buffer too small: need %d, have %d", totalSize, len(buffer))
	}
	rows := make([]T, numEntries)
	for i := uint32(0); i < numEntries; i++ {
		offset := headerSize + rowSize*uintptr(i)
		rows[i] = *(*T)(unsafe.Pointer(&buffer[offset]))
	}
	return rows, nil
}

// Tcp4Endpoints 返回所有 IPv4 TCP 端点条目。
//   返回1 - IPv4 TCP 端点条目列表
//   返回2 - 错误信息
func Tcp4Endpoints() ([]MIB_TCPROW_OWNER_PID, error) {
	buf, num, err := getTcpTableBuffer(false, AF_INET, TCP_TABLE_OWNER_PID_ALL, 0)
	if err != nil {
		return nil, fmt.Errorf("tcp4: %w", err)
	}
	return readRows[MIB_TCPROW_OWNER_PID](buf, num)
}

// Tcp6Endpoints 返回所有 IPv6 TCP 端点条目。
//   返回1 - IPv6 TCP 端点条目列表
//   返回2 - 错误信息
func Tcp6Endpoints() ([]MIB_TCP6ROW_OWNER_PID, error) {
	buf, num, err := getTcpTableBuffer(false, AF_INET6, TCP_TABLE_OWNER_PID_ALL, 0)
	if err != nil {
		return nil, fmt.Errorf("tcp6: %w", err)
	}
	return readRows[MIB_TCP6ROW_OWNER_PID](buf, num)
}

// Udp4Endpoints 返回所有 IPv4 UDP 端点条目。
//   返回1 - IPv4 UDP 端点条目列表
//   返回2 - 错误信息
func Udp4Endpoints() ([]MIB_UDPROW_OWNER_PID, error) {
	buf, num, err := getUdpTableBuffer(false, AF_INET, UDP_TABLE_OWNER_PID, 0)
	if err != nil {
		return nil, fmt.Errorf("udp4: %w", err)
	}
	return readRows[MIB_UDPROW_OWNER_PID](buf, num)
}

// Udp6Endpoints 返回所有 IPv6 UDP 端点条目。
//   返回1 - IPv6 UDP 端点条目列表
//   返回2 - 错误信息
func Udp6Endpoints() ([]MIB_UDP6ROW_OWNER_PID, error) {
	buf, num, err := getUdpTableBuffer(false, AF_INET6, UDP_TABLE_OWNER_PID, 0)
	if err != nil {
		return nil, fmt.Errorf("udp6: %w", err)
	}
	return readRows[MIB_UDP6ROW_OWNER_PID](buf, num)
}

// WfpCallouts 枚举所有 WFP 标注。
//   返回1 - WFP 标注列表
//   返回2 - 错误信息
func WfpCallouts() ([]FwpmCallout, error) {
	var (
		entries    []*FwpmCallout0
		numEntries uint32
		array      **FwpmCallout0
	)
	session := FwpmSession0{
		DisplayData: FwpmDisplayData0{
			Name: windows.StringToUTF16Ptr("wcorefx"),
		},
	}
	engineH, err := FwpmEngineOpen(nil, 10, nil, &session)
	if err != nil {
		return nil, fmt.Errorf("FwpmEngineOpen: %w", err)
	}
	defer FwpmEngineClose(engineH)

	enumH, err := FwpmCalloutCreateEnumHandle(engineH, nil)
	if err != nil {
		return nil, fmt.Errorf("FwpmCalloutCreateEnumHandle: %w", err)
	}
	defer FwpmCalloutDestroyEnumHandle(engineH, enumH)

	err = FwpmCalloutEnum(engineH, enumH, 0xFFFFFFFF, &array, &numEntries)
	if err != nil {
		return nil, fmt.Errorf("FwpmCalloutEnum: %w", err)
	}
	if numEntries == 0 {
		return nil, nil
	}

	sh := (*reflect.SliceHeader)(unsafe.Pointer(&entries))
	sh.Cap = int(numEntries)
	sh.Len = int(numEntries)
	sh.Data = uintptr(unsafe.Pointer(array))
	defer FwpmFreeMemory(unsafe.Pointer(array))

	result := make([]FwpmCallout, 0, numEntries)
	for _, c := range entries {
		result = append(result, FwpmCallout{
			CalloutKey:   c.CalloutKey,
			CalloutId:    c.CalloutId,
			Name:         windows.UTF16PtrToString(c.DisplayData.Name),
			Description:  windows.UTF16PtrToString(c.DisplayData.Description),
			Flags:        c.Flags,
			ProviderKey:  c.ProviderKey,
			ProviderData: c.ProviderData,
			LayerKey:     c.LayerKey,
		})
	}
	return result, nil
}

// WfpFilters 枚举所有 WFP 过滤器。
//   返回1 - WFP 过滤器列表
//   返回2 - 错误信息
func WfpFilters() ([]FwpmFilter, error) {
	var (
		entries    []*FwpmFilter0
		numEntries uint32
		array      **FwpmFilter0
	)
	session := FwpmSession0{
		DisplayData: FwpmDisplayData0{
			Name: windows.StringToUTF16Ptr("wcorefx"),
		},
	}
	engineH, err := FwpmEngineOpen(nil, 10, nil, &session)
	if err != nil {
		return nil, fmt.Errorf("FwpmEngineOpen: %w", err)
	}
	defer FwpmEngineClose(engineH)

	enumH, err := FwpmFilterCreateEnumHandle(engineH, nil)
	if err != nil {
		return nil, fmt.Errorf("FwpmFilterCreateEnumHandle: %w", err)
	}
	defer FwpmFilterDestroyEnumHandle(engineH, enumH)

	err = FwpmFilterEnum(engineH, enumH, 0xFFFFFFFF, &array, &numEntries)
	if err != nil {
		return nil, fmt.Errorf("FwpmFilterEnum: %w", err)
	}
	if numEntries == 0 {
		return nil, nil
	}

	sh := (*reflect.SliceHeader)(unsafe.Pointer(&entries))
	sh.Cap = int(numEntries)
	sh.Len = int(numEntries)
	sh.Data = uintptr(unsafe.Pointer(array))
	defer FwpmFreeMemory(unsafe.Pointer(array))

	result := make([]FwpmFilter, 0, numEntries)
	for _, f := range entries {
		result = append(result, FwpmFilter{
			FilterKey:           f.FilterKey,
			Name:                windows.UTF16PtrToString(f.DisplayData.Name),
			Description:         windows.UTF16PtrToString(f.DisplayData.Description),
			Flags:               f.Flags,
			ProviderKey:         f.ProviderKey,
			ProviderData:        f.ProviderData,
			LayerKey:            f.LayerKey,
			SublayerKey:         f.SublayerKey,
			Weight:              f.Weight,
			NumFilterConditions: f.NumFilterConditions,
			FilterConditions:    f.FilterConditions,
			Action:              f.Action,
			RawContext:          f.RawContext,
			ProviderContextKey:  f.ProviderContextKey,
			Reserved:            f.Reserved,
			FilterID:            f.FilterID,
			EffectiveWeight:     f.EffectiveWeight,
		})
	}
	return result, nil
}
