package advapi32

import (
	"golang.org/x/sys/windows"
)

var (
	modadvapi32 = windows.NewLazySystemDLL("Advapi32.dll")
)

var (
	procLookupPrivilegeNameW = modadvapi32.NewProc("LookupPrivilegeNameW")
	procCheckTokenMembership = modadvapi32.NewProc("CheckTokenMembership")
)
