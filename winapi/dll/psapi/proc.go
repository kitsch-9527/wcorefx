package psapi

import "golang.org/x/sys/windows"

var modpsapi = windows.NewLazySystemDLL("psapi.dll")

var (
	procQueryWorkingSet          = modpsapi.NewProc("QueryWorkingSet")
	procGetProcessMemoryInfo     = modpsapi.NewProc("GetProcessMemoryInfo")
	procEnumDeviceDrivers        = modpsapi.NewProc("EnumDeviceDrivers")
	procGetDeviceDriverBaseNameW = modpsapi.NewProc("GetDeviceDriverBaseNameW")
	procGetDeviceDriverFileNameW = modpsapi.NewProc("GetDeviceDriverFileNameW")
)
