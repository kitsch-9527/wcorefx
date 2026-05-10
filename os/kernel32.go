//go:build windows

package os

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

var modkernel32 = windows.NewLazySystemDLL("kernel32.dll")

var (
	procGetSystemInfo         = modkernel32.NewProc("GetNativeSystemInfo")
	procGetTickCount64        = modkernel32.NewProc("GetTickCount64")
	procGlobalMemoryStatusEx  = modkernel32.NewProc("GlobalMemoryStatusEx")
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

// MEMORYSTATUSEX maps to Windows MEMORYSTATUSEX structure.
type MEMORYSTATUSEX struct {
	dwLength                uint32
	dwMemoryLoad            uint32
	ullTotalPhys            uint64
	ullAvailPhys            uint64
	ullTotalPageFile        uint64
	ullAvailPageFile        uint64
	ullTotalVirtual         uint64
	ullAvailVirtual         uint64
	ullAvailExtendedVirtual uint64
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

func globalMemoryStatusEx() (MEMORYSTATUSEX, error) {
	var ms MEMORYSTATUSEX
	ms.dwLength = uint32(unsafe.Sizeof(ms))
	ret, _, _ := procGlobalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&ms)))
	if ret == 0 {
		return ms, fmt.Errorf("GlobalMemoryStatusEx failed")
	}
	return ms, nil
}
