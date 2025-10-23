package ntdll

import "golang.org/x/sys/windows"

var modntdll = windows.NewLazySystemDLL("ntdll.dll")

var (
	procNtDuplicateObject  = modntdll.NewProc("NtDuplicateObject")
	procNtQueryObject      = modntdll.NewProc("NtQueryObject")
	procRtlAdjustPrivilege = modntdll.NewProc("RtlAdjustPrivilege")
)
