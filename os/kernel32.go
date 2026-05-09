//go:build windows

package os

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

var modkernel32 = windows.NewLazySystemDLL("kernel32.dll")

var (
	procGetSystemInfo  = modkernel32.NewProc("GetNativeSystemInfo")
	procGetTickCount64 = modkernel32.NewProc("GetTickCount64")
)

type SYSTEM_INFO struct {
	wProcessorArchitecture      uint16
	wReserved                   uint16
	dwPageSize                  uint32
	lpMinimumApplicationAddress uintptr
	lpMaximumApplicationAddress uintptr
	dwActiveProcessorMask       uintptr
	dwNumberOfProcessors        uint32
	dwProcessorType             uint32
	dwAllocationGranularity     uint32
	wProcessorLevel             uint16
	wProcessorRevision          uint16
}

func getNativeSystemInfo() SYSTEM_INFO {
	var info SYSTEM_INFO
	procGetSystemInfo.Call(uintptr(unsafe.Pointer(&info)))
	return info
}

func getTickCount64() uint64 {
	ret, _, _ := procGetTickCount64.Call()
	return uint64(ret)
}
