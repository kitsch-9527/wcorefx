package iphlpapi

import "golang.org/x/sys/windows"

var modiphlpapi = windows.NewLazySystemDLL("iphlpapi.dll")

var (
	procGetExtendedTcpTable = modiphlpapi.NewProc("GetExtendedTcpTable")
	procGetExtendedUdpTable = modiphlpapi.NewProc("GetExtendedUdpTable")
)
