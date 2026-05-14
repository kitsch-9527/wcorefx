//go:build windows

package netapi

import (
	"encoding/binary"
	"fmt"
	"net"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/kitsch-9527/wcorefx/internal/winapi"
)

// MIB_TCPROW_OWNER_PID 表示 MIB_TCPROW_OWNER_PID 结构，包含 TCP 连接信息及所属进程 PID。
type MIB_TCPROW_OWNER_PID struct {
	DwState      uint32
	DwLocalAddr  uint32
	DwLocalPort  uint32
	DwRemoteAddr uint32
	DwRemotePort uint32
	DwOwningPid  uint32
}

// MIB_TCP6ROW_OWNER_PID 表示 MIB_TCP6ROW_OWNER_PID 结构，包含 IPv6 TCP 连接信息及所属进程 PID。
type MIB_TCP6ROW_OWNER_PID struct {
	LocalAddr     [16]byte
	LocalScopeId  uint32
	LocalPort     uint32
	RemoteAddr    [16]byte
	RemoteScopeId uint32
	RemotePort    uint32
	DwState       uint32
	DwOwningPid   uint32
}

// MIB_UDPROW_OWNER_PID 表示 MIB_UDPROW_OWNER_PID 结构，包含 UDP 端点信息及所属进程 PID。
type MIB_UDPROW_OWNER_PID struct {
	DwLocalAddr uint32
	DwLocalPort uint32
	DwOwningPid uint32
}

// MIB_UDP6ROW_OWNER_PID 表示 MIB_UDP6ROW_OWNER_PID 结构，包含 IPv6 UDP 端点信息及所属进程 PID。
type MIB_UDP6ROW_OWNER_PID struct {
	LocalAddr    [16]byte
	LocalScopeId uint32
	LocalPort    uint32
	DwOwningPid  uint32
}

// InterfaceInfo 表示网络接口信息。
type InterfaceInfo struct {
	Name        string
	Description string
	IP          string
	MAC         []byte
	IsUp        bool
	Speed       uint64
}

// ARPEntry 表示 ARP 表条目。
type ARPEntry struct {
	IP      string
	MAC     []byte
	IfIndex uint32
}

// RouteEntry 表示路由表条目。
type RouteEntry struct {
	Destination string
	NextHop     string
	IfIndex     uint32
	Metric      uint32
}

const (
	AF_INET                  = 2
	AF_INET6                 = 23
	TCP_TABLE_OWNER_PID_ALL  = 5
	UDP_TABLE_OWNER_PID      = 1
)

// MibTCPState 常量定义 TCP 连接状态枚举值。
const (
	MibTCPStateClosed      = iota + 1
	MibTCPStateListen
	MibTCPStateSynSent
	MibTCPStateSynRcvd
	MibTCPStateEstablished
	MibTCPStateFinWait1
	MibTCPStateFinWait2
	MibTCPStateCloseWait
	MibTCPStateClosing
	MibTCPStateLastAck
	MibTCPStateTimeWait
	MibTCPStateDeleteTCB
)

// TCPState 将 TCP 状态数值转换为可读的字符串描述。
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
func InetNtoa(addr uint32) string {
	return net.IPv4(
		byte(addr>>24),
		byte(addr>>16),
		byte(addr>>8),
		byte(addr),
	).String()
}

// InetNtoa6 将 128 位网络字节序的 IPv6 地址转换为字符串。
func InetNtoa6(addr [16]byte) string {
	return net.IP(addr[:]).String()
}

// Ntohs 将网络字节序（大端）的 16 位端口值转换为主机字节序。
func Ntohs(port uint32) uint16 {
	return uint16((port>>8)|(port<<8)) & 0xffff
}

// CalloutInfo 存储 WFP 标注信息。
type CalloutInfo struct {
	CalloutId   uint32
	CalloutKey  windows.GUID
	Name        string
	Description string
}

// FwpmCallout 存储完整的 WFP 标注数据。
type FwpmCallout struct {
	CalloutKey   windows.GUID
	Name         string
	Description  string
	Flags        FwpmFilterFlags
	ProviderKey  *windows.GUID
	ProviderData FwpByteBlob
	LayerKey     LayerID
	CalloutId    uint32
}

// FwpmFilter 存储完整的 WFP 过滤器数据。
type FwpmFilter struct {
	FilterKey           windows.GUID
	Name                string
	Description         string
	Flags               FwpmFilterFlags
	ProviderKey         *windows.GUID
	ProviderData        FwpByteBlob
	LayerKey            LayerID
	SublayerKey         SublayerID
	Weight              FwpValue0
	NumFilterConditions uint32
	FilterConditions    *FwpmFilterCondition0
	Action              FwpmAction0
	RawContext          uint64
	ProviderContextKey  windows.GUID
	Reserved            *windows.GUID
	FilterID            uint64
	EffectiveWeight     FwpValue0
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
func Tcp4Endpoints() ([]MIB_TCPROW_OWNER_PID, error) {
	return tcpEndpoints(MIB_TCPROW_OWNER_PID{}, AF_INET, TCP_TABLE_OWNER_PID_ALL)
}

// Tcp6Endpoints 返回所有 IPv6 TCP 端点条目。
func Tcp6Endpoints() ([]MIB_TCP6ROW_OWNER_PID, error) {
	return tcpEndpoints(MIB_TCP6ROW_OWNER_PID{}, AF_INET6, TCP_TABLE_OWNER_PID_ALL)
}

func tcpEndpoints[T any](_ T, af, tableClass int) ([]T, error) {
	buf, err := winapi.BufferQuery(winapi.FuncStrategy(func(buf []byte) (int, error) {
		var size uint32 = uint32(len(buf))
		p := (*byte)(nil)
		if len(buf) > 0 {
			p = &buf[0]
		}
		err := procGetExtendedTcpTable.Call(
			uintptr(unsafe.Pointer(p)),
			uintptr(unsafe.Pointer(&size)),
			0, uintptr(af), uintptr(tableClass), 0,
		)
		if err != nil {
			if winapi.IsErrInsufficientBuffer(err) {
				return 0, &winapi.ErrInsufficientBuffer{Size: int(size)}
			}
			return 0, err
		}
		return int(size), nil
	}))
	if err != nil {
		return nil, fmt.Errorf("tcp: %w", err)
	}
	if len(buf) < 4 {
		return nil, fmt.Errorf("tcp: buffer too small for header")
	}
	return readRows[T](buf, binary.LittleEndian.Uint32(buf[:4]))
}

// Udp4Endpoints 返回所有 IPv4 UDP 端点条目。
func Udp4Endpoints() ([]MIB_UDPROW_OWNER_PID, error) {
	return udpEndpoints(MIB_UDPROW_OWNER_PID{}, AF_INET, UDP_TABLE_OWNER_PID)
}

// Udp6Endpoints 返回所有 IPv6 UDP 端点条目。
func Udp6Endpoints() ([]MIB_UDP6ROW_OWNER_PID, error) {
	return udpEndpoints(MIB_UDP6ROW_OWNER_PID{}, AF_INET6, UDP_TABLE_OWNER_PID)
}

func udpEndpoints[T any](_ T, af, tableClass int) ([]T, error) {
	buf, err := winapi.BufferQuery(winapi.FuncStrategy(func(buf []byte) (int, error) {
		var size uint32 = uint32(len(buf))
		p := (*byte)(nil)
		if len(buf) > 0 {
			p = &buf[0]
		}
		err := procGetExtendedUdpTable.Call(
			uintptr(unsafe.Pointer(p)),
			uintptr(unsafe.Pointer(&size)),
			0, uintptr(af), uintptr(tableClass), 0,
		)
		if err != nil {
			if winapi.IsErrInsufficientBuffer(err) {
				return 0, &winapi.ErrInsufficientBuffer{Size: int(size)}
			}
			return 0, err
		}
		return int(size), nil
	}))
	if err != nil {
		return nil, fmt.Errorf("udp: %w", err)
	}
	if len(buf) < 4 {
		return nil, fmt.Errorf("udp: buffer too small for header")
	}
	return readRows[T](buf, binary.LittleEndian.Uint32(buf[:4]))
}

// WfpCallouts 枚举所有 WFP 标注。
func WfpCallouts() ([]FwpmCallout, error) {
	s, err := NewWfpSession()
	if err != nil {
		return nil, err
	}
	defer s.Close()
	return s.Callouts()
}

// WfpFilters 枚举所有 WFP 过滤器。
func WfpFilters() ([]FwpmFilter, error) {
	s, err := NewWfpSession()
	if err != nil {
		return nil, err
	}
	defer s.Close()
	return s.Filters()
}

// Interfaces 返回所有网络接口信息。
func Interfaces() ([]InterfaceInfo, error) {
	buf, err := winapi.BufferQuery(winapi.FuncStrategy(func(buf []byte) (int, error) {
		var bufSize uint32 = uint32(len(buf))
		p := (*byte)(nil)
		if len(buf) > 0 {
			p = &buf[0]
		}
		err := procGetAdaptersAddresses.Call(
			uintptr(windows.AF_UNSPEC),
			uintptr(0x0010), // GAA_FLAG_INCLUDE_PREFIX
			0,
			uintptr(unsafe.Pointer(p)),
			uintptr(unsafe.Pointer(&bufSize)),
		)
		if err != nil {
			if err == windows.ERROR_BUFFER_OVERFLOW || winapi.IsErrInsufficientBuffer(err) {
				return int(bufSize), &winapi.ErrInsufficientBuffer{Size: int(bufSize)}
			}
			return 0, err
		}
		return int(bufSize), nil
	}))
	if err != nil {
		return nil, fmt.Errorf("GetAdaptersAddresses: %w", err)
	}
	if len(buf) == 0 {
		return nil, fmt.Errorf("GetAdaptersAddresses returned empty buffer")
	}

	var ifaces []InterfaceInfo
	p := (*windows.IpAdapterAddresses)(unsafe.Pointer(&buf[0]))
	for p != nil {
		info := InterfaceInfo{
			Name:  windows.UTF16PtrToString(p.FriendlyName),
			IsUp:  p.OperStatus == 1,
			Speed: p.TransmitLinkSpeed,
		}
		if p.Description != nil {
			info.Description = windows.UTF16PtrToString(p.Description)
		}
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
func ARP() ([]ARPEntry, error) {
	var table unsafe.Pointer
	err := procGetIpNetTable2.Call(
		uintptr(windows.AF_UNSPEC),
		uintptr(unsafe.Pointer(&table)),
		0,
	)
	if err != nil {
		return nil, fmt.Errorf("GetIpNetTable2: %w", err)
	}
	defer procFreeMibTable.Call(uintptr(table))

	type mibIpNetRow2 struct {
		Address            [28]byte
		InterfaceIndex     uint32
		InterfaceLUID      uint64
		PhysicalAddress    [32]byte
		PhysicalAddressLen uint32
		State              uint32
		_                  [8]byte
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
func Route() ([]RouteEntry, error) {
	var table unsafe.Pointer
	err := procGetIpForwardTable2.Call(
		uintptr(windows.AF_UNSPEC),
		uintptr(unsafe.Pointer(&table)),
	)
	if err != nil {
		return nil, fmt.Errorf("GetIpForwardTable2: %w", err)
	}
	defer procFreeMibTable.Call(uintptr(table))

	type mibIpForwardRow2 struct {
		InterfaceLUID     uint64
		InterfaceIndex    uint32
		DestinationPrefix [32]byte
		NextHop           [28]byte
		SitePrefixLength  uint8
		_                 [3]byte
		ValidLifetime     uint32
		PreferredLifetime uint32
		Metric            uint32
		_                 [16]byte
	}

	numEntries := *(*uint32)(table)
	rowSize := unsafe.Sizeof(mibIpForwardRow2{})
	firstRow := unsafe.Pointer(uintptr(table) + 8)

	entries := make([]RouteEntry, 0, numEntries)
	for i := uint32(0); i < numEntries; i++ {
		r := *(*mibIpForwardRow2)(unsafe.Pointer(uintptr(firstRow) + rowSize*uintptr(i)))

		destFamily := *(*uint16)(unsafe.Pointer(&r.DestinationPrefix[0]))
		var destIP string
		if destFamily == windows.AF_INET {
			destIP = net.IP(r.DestinationPrefix[4:8]).String()
		} else if destFamily == windows.AF_INET6 {
			destIP = net.IP(r.DestinationPrefix[8:24]).String()
		}
		prefixLen := r.DestinationPrefix[28]

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
