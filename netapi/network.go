//go:build windows

package netapi

import (
	"encoding/binary"
	"fmt"
	"net"
	"reflect"
	"syscall"
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

// InterfaceInfo 表示网络接口信息。
type InterfaceInfo struct {
	// Name 接口名称。
	Name        string
	// Description 接口描述。
	Description string
	// IP 接口 IP 地址。
	IP          string
	// MAC 接口 MAC 地址。
	MAC         []byte
	// IsUp 接口是否启用。
	IsUp        bool
	// Speed 接口速度（bps）。
	Speed       uint64
}

// ARPEntry 表示 ARP 表条目。
type ARPEntry struct {
	// IP 目标 IP 地址。
	IP      string
	// MAC MAC 地址。
	MAC     []byte
	// IfIndex 接口索引。
	IfIndex uint32
}

// RouteEntry 表示路由表条目。
type RouteEntry struct {
	// Destination 目标网络。
	Destination string
	// NextHop 下一跳。
	NextHop     string
	// IfIndex 接口索引。
	IfIndex     uint32
	// Metric 路由度量。
	Metric      uint32
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

// Interfaces 返回所有网络接口信息。
//   返回1 - 网络接口信息列表
//   返回2 - 错误信息
func Interfaces() ([]InterfaceInfo, error) {
	var bufSize uint32
	// First call to get required buffer size (expect ERROR_BUFFER_OVERFLOW)
	r1, _, _ := syscall.SyscallN(procGetAdaptersAddresses.Addr(),
		uintptr(windows.AF_UNSPEC),
		uintptr(0x0010), // GAA_FLAG_INCLUDE_PREFIX
		0,
		0,
		uintptr(unsafe.Pointer(&bufSize)),
	)
	if r1 != 0 && syscall.Errno(r1) != windows.ERROR_BUFFER_OVERFLOW {
		return nil, fmt.Errorf("GetAdaptersAddresses size query: %w", syscall.Errno(r1))
	}
	if bufSize == 0 {
		return nil, fmt.Errorf("GetAdaptersAddresses returned buffer size 0")
	}

	buf := make([]byte, bufSize)
	r1, _, _ = syscall.SyscallN(procGetAdaptersAddresses.Addr(),
		uintptr(windows.AF_UNSPEC),
		uintptr(0x0010),
		0,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&bufSize)),
	)
	if r1 != 0 {
		return nil, fmt.Errorf("GetAdaptersAddresses failed: %w", syscall.Errno(r1))
	}

	var ifaces []InterfaceInfo
	p := (*windows.IpAdapterAddresses)(unsafe.Pointer(&buf[0]))
	for p != nil {
		info := InterfaceInfo{
			Name:  windows.UTF16PtrToString(p.FriendlyName),
			IsUp:  p.OperStatus == 1, // IfOperStatusUp
			Speed: p.TransmitLinkSpeed,
		}
		if p.Description != nil {
			info.Description = windows.UTF16PtrToString(p.Description)
		}
		// Extract IP from first unicast address
		if p.FirstUnicastAddress != nil {
			sa := p.FirstUnicastAddress.Address.Sockaddr
			if sa != nil {
				family := *(*uint16)(unsafe.Pointer(sa))
				if family == windows.AF_INET {
					sa4 := (*windows.RawSockaddrInet4)(unsafe.Pointer(sa))
					info.IP = net.IP(sa4.Addr[:]).String()
				} else if family == windows.AF_INET6 {
					sa6 := (*windows.RawSockaddrInet6)(unsafe.Pointer(sa))
					info.IP = net.IP(sa6.Addr[:]).String()
				}
			}
		}
		// Copy MAC address
		if p.PhysicalAddressLength > 0 {
			info.MAC = make([]byte, p.PhysicalAddressLength)
			copy(info.MAC, p.PhysicalAddress[:p.PhysicalAddressLength])
		}
		ifaces = append(ifaces, info)
		p = p.Next
	}
	return ifaces, nil
}

// ARP 返回 ARP 表条目。
//   返回1 - ARP 表条目列表
//   返回2 - 错误信息
func ARP() ([]ARPEntry, error) {
	var table unsafe.Pointer
	r1, _, _ := syscall.SyscallN(procGetIpNetTable2.Addr(),
		uintptr(windows.AF_UNSPEC),
		uintptr(unsafe.Pointer(&table)),
		0,
	)
	if r1 != 0 {
		return nil, fmt.Errorf("GetIpNetTable2 failed: %w", syscall.Errno(r1))
	}
	defer freeMibTable(table)

	// MIB_IPNET_ROW2 layout from netioapi.h:
	// Address(SOCKADDR_INET=28) + InterfaceIndex(4) + InterfaceLuid(8) +
	// PhysicalAddress[IF_MAX_PHYS_ADDRESS_LENGTH=32] + PhysicalAddressLen(4) +
	// State(4) + Flags(1+pad3=4) + ReachabilityTime(4) = 88 bytes
	type mibIpNetRow2 struct {
		Address            [28]byte
		InterfaceIndex     uint32
		InterfaceLUID      uint64
		PhysicalAddress    [32]byte  // IF_MAX_PHYS_ADDRESS_LENGTH = 32
		PhysicalAddressLen uint32
		State              uint32
		_                  [8]byte   // trailing union fields
	}

	numEntries := *(*uint32)(table)
	rowSize := unsafe.Sizeof(mibIpNetRow2{})
	firstRow := unsafe.Pointer(uintptr(table) + 8)

	entries := make([]ARPEntry, 0, numEntries)
	for i := uint32(0); i < numEntries; i++ {
		r := *(*mibIpNetRow2)(unsafe.Pointer(uintptr(firstRow) + rowSize*uintptr(i)))

		macLen := r.PhysicalAddressLen
		if macLen > 32 {
			macLen = 32
		}
		mac := make([]byte, macLen)
		copy(mac, r.PhysicalAddress[:macLen])

		var ip string
		family := *(*uint16)(unsafe.Pointer(&r.Address[0]))
		if family == windows.AF_INET {
			ip = net.IP(r.Address[4:8]).String()
		} else if family == windows.AF_INET6 {
			ip = net.IP(r.Address[8:24]).String()
		}

		entries = append(entries, ARPEntry{
			IP:      ip,
			MAC:     mac,
			IfIndex: r.InterfaceIndex,
		})
	}
	return entries, nil
}

// Route 返回路由表条目。
//   返回1 - 路由表条目列表
//   返回2 - 错误信息
func Route() ([]RouteEntry, error) {
	var table unsafe.Pointer
	r1, _, _ := syscall.SyscallN(procGetIpForwardTable2.Addr(),
		uintptr(windows.AF_UNSPEC),
		uintptr(unsafe.Pointer(&table)),
	)
	if r1 != 0 {
		return nil, fmt.Errorf("GetIpForwardTable2 failed: %w", syscall.Errno(r1))
	}
	defer freeMibTable(table)

	// MIB_IPFORWARD_ROW2 layout from netioapi.h:
	// InterfaceLuid(8) + InterfaceIndex(4) + DestinationPrefix(IP_ADDRESS_PREFIX=32) +
	// NextHop(SOCKADDR_INET=28) + SitePrefixLength(1+pad3=4) +
	// ValidLifetime(4) + PreferredLifetime(4) + Metric(4) +
	// remaining fields(16) = 104 bytes
	type mibIpForwardRow2 struct {
		InterfaceLUID     uint64    // offset 0
		InterfaceIndex    uint32    // offset 8
		DestinationPrefix [32]byte  // offset 12: IP_ADDRESS_PREFIX = SOCKADDR_INET(28) + PrefixLength(1) + pad(3)
		NextHop           [28]byte  // offset 44: SOCKADDR_INET
		SitePrefixLength  uint8     // offset 72
		_                 [3]byte   // offset 73: padding
		ValidLifetime     uint32    // offset 76
		PreferredLifetime uint32    // offset 80
		Metric            uint32    // offset 84
		_                 [16]byte  // offset 88: remaining fields (Protocol, Loopback, AutoconfigureAddress, Publish, Immortal, Age, Origin)
	}

	numEntries := *(*uint32)(table)
	rowSize := unsafe.Sizeof(mibIpForwardRow2{})
	firstRow := unsafe.Pointer(uintptr(table) + 8)

	entries := make([]RouteEntry, 0, numEntries)
	for i := uint32(0); i < numEntries; i++ {
		r := *(*mibIpForwardRow2)(unsafe.Pointer(uintptr(firstRow) + rowSize*uintptr(i)))

		// DestinationPrefix is IP_ADDRESS_PREFIX: SOCKADDR_INET(28) + PrefixLength(1) + pad(3)
		destFamily := *(*uint16)(unsafe.Pointer(&r.DestinationPrefix[0]))
		var destIP string
		if destFamily == windows.AF_INET {
			destIP = net.IP(r.DestinationPrefix[4:8]).String()
		} else if destFamily == windows.AF_INET6 {
			destIP = net.IP(r.DestinationPrefix[8:24]).String()
		}
		prefixLen := r.DestinationPrefix[28]

		// NextHop is SOCKADDR_INET
		nextHopFamily := *(*uint16)(unsafe.Pointer(&r.NextHop[0]))
		var nextHopIP string
		if nextHopFamily == windows.AF_INET {
			nextHopIP = net.IP(r.NextHop[4:8]).String()
		} else if nextHopFamily == windows.AF_INET6 {
			nextHopIP = net.IP(r.NextHop[8:24]).String()
		}

		dest := destIP
		if prefixLen > 0 {
			dest = fmt.Sprintf("%s/%d", destIP, prefixLen)
		}

		entries = append(entries, RouteEntry{
			Destination: dest,
			NextHop:     nextHopIP,
			IfIndex:     r.InterfaceIndex,
			Metric:      r.Metric,
		})
	}
	return entries, nil
}
