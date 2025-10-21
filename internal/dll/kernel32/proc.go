package kernel32

import "golang.org/x/sys/windows"

var modkernel32 = windows.NewLazySystemDLL("kernel32.dll")

var (
	procGetNativeSystemInfo = modkernel32.NewProc("GetNativeSystemInfo")
)
