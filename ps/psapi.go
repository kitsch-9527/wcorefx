//go:build windows

package ps

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

type processMemoryCounters struct {
	CB                         uint32
	PageFaultCount             uint32
	PeakWorkingSetSize         uintptr
	WorkingSetSize             uintptr
	QuotaPeakPagedPoolUsage    uintptr
	QuotaPagedPoolUsage        uintptr
	QuotaPeakNonPagedPoolUsage uintptr
	QuotaNonPagedPoolUsage     uintptr
	PagefileUsage              uintptr
	PeakPagefileUsage          uintptr
}

var (
	modpsapi = windows.NewLazySystemDLL("psapi.dll")

	procGetProcessMemoryInfo = modpsapi.NewProc("GetProcessMemoryInfo")
)

func getProcessMemoryInfo(handle windows.Handle) (processMemoryCounters, error) {
	var mem processMemoryCounters
	mem.CB = uint32(unsafe.Sizeof(mem))
	r1, _, _ := procGetProcessMemoryInfo.Call(
		uintptr(handle),
		uintptr(unsafe.Pointer(&mem)),
		uintptr(mem.CB),
	)
	if r1 == 0 {
		return mem, syscall.GetLastError()
	}
	return mem, nil
}
