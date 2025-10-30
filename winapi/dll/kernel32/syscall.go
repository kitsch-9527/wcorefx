package kernel32

import (
	"syscall"
	"unsafe"
)

func GetSystemInfo() systemInfo {
	var si systemInfo
	syscall.Syscall(
		procGetNativeSystemInfo.Addr(),
		1,
		uintptr(unsafe.Pointer(&si)),
		0,
		0,
	)
	return si
}
