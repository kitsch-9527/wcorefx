//go:build windows

package ps

import (
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/kitsch-9527/wcorefx/internal/winapi"
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

var procGetProcessMemoryInfo = winapi.NewProc("psapi.dll", "GetProcessMemoryInfo")

func getProcessMemoryInfo(handle windows.Handle) (processMemoryCounters, error) {
	var mem processMemoryCounters
	mem.CB = uint32(unsafe.Sizeof(mem))
	err := procGetProcessMemoryInfo.Call(
		uintptr(handle),
		uintptr(unsafe.Pointer(&mem)),
		uintptr(mem.CB),
	)
	if err != nil {
		return mem, err
	}
	return mem, nil
}
