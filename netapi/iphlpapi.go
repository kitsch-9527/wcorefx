//go:build windows

package netapi

import "github.com/kitsch-9527/wcorefx/internal/winapi"

var (
	procGetExtendedTcpTable  = winapi.NewProc("iphlpapi.dll", "GetExtendedTcpTable", winapi.ConvErrnoReturn)
	procGetExtendedUdpTable  = winapi.NewProc("iphlpapi.dll", "GetExtendedUdpTable", winapi.ConvErrnoReturn)
	procGetAdaptersAddresses = winapi.NewProc("iphlpapi.dll", "GetAdaptersAddresses", winapi.ConvErrnoReturn)
	procGetIpNetTable2       = winapi.NewProc("iphlpapi.dll", "GetIpNetTable2", winapi.ConvErrnoReturn)
	procGetIpForwardTable2   = winapi.NewProc("iphlpapi.dll", "GetIpForwardTable2", winapi.ConvErrnoReturn)
	procFreeMibTable         = winapi.NewProc("iphlpapi.dll", "FreeMibTable", winapi.ConvErrnoReturn)
)
