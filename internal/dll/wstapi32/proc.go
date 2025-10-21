package psapi

var modpsapi = windows.NewLazySystemDLL("psapi.dll")

var (
	QueryWorkingSet      = modpsapi.NewProc("QueryWorkingSet")
	GetProcessMemoryInfo = modpsapi.NewProc("GetProcessMemoryInfo")
)
