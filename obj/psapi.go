//go:build windows

package obj

import (
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/kitsch-9527/wcorefx/internal/winapi"
)

var (
	procEnumDeviceDrivers        = winapi.NewProc("psapi.dll", "EnumDeviceDrivers")
	procGetDeviceDriverFileNameW = winapi.NewProc("psapi.dll", "GetDeviceDriverFileNameW")
	procGetDeviceDriverBaseNameW = winapi.NewProc("psapi.dll", "GetDeviceDriverBaseNameW")
	procGetProcessMemoryInfo     = winapi.NewProc("psapi.dll", "GetProcessMemoryInfo")
)

type processMemoryCounters struct {
	CB                         uint32
	PageFaultCount             uint32
	PeakWorkingSetSize         uint64
	WorkingSetSize             uint64
	QuotaPeakPagedPoolUsage    uint64
	QuotaPagedPoolUsage        uint64
	QuotaPeakNonPagedPoolUsage uint64
	QuotaNonPagedPoolUsage     uint64
	PagefileUsage              uint64
	PeakPagefileUsage          uint64
}

func enumDeviceDrivers(drivers *uintptr, cb uint32, lpcNeeded *uint32) error {
	return procEnumDeviceDrivers.Call(
		uintptr(unsafe.Pointer(drivers)),
		uintptr(cb),
		uintptr(unsafe.Pointer(lpcNeeded)),
	)
}

func getDeviceDriverFileName(driver uintptr, lpFilename *uint16, nSize uint32) error {
	return procGetDeviceDriverFileNameW.Call(
		driver,
		uintptr(unsafe.Pointer(lpFilename)),
		uintptr(nSize),
	)
}

func getDeviceDriverBaseName(driver uintptr, lpBaseName *uint16, nSize uint32) error {
	return procGetDeviceDriverBaseNameW.Call(
		driver,
		uintptr(unsafe.Pointer(lpBaseName)),
		uintptr(nSize),
	)
}

func getProcessMemoryInfo(h windows.Handle, mem *processMemoryCounters) error {
	return procGetProcessMemoryInfo.Call(
		uintptr(h),
		uintptr(unsafe.Pointer(mem)),
		uintptr(unsafe.Sizeof(*mem)),
	)
}
