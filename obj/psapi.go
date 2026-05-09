//go:build windows

package obj

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var modpsapi = windows.NewLazySystemDLL("psapi.dll")

var (
	procEnumDeviceDrivers        = modpsapi.NewProc("EnumDeviceDrivers")
	procGetDeviceDriverFileNameW = modpsapi.NewProc("GetDeviceDriverFileNameW")
	procGetDeviceDriverBaseNameW = modpsapi.NewProc("GetDeviceDriverBaseNameW")
	procGetProcessMemoryInfo     = modpsapi.NewProc("GetProcessMemoryInfo")
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
	r1, _, e1 := syscall.SyscallN(
		procEnumDeviceDrivers.Addr(),
		uintptr(unsafe.Pointer(drivers)),
		uintptr(cb),
		uintptr(unsafe.Pointer(lpcNeeded)),
	)
	if r1 == 0 {
		if e1 != 0 {
			return e1
		}
		return syscall.EINVAL
	}
	return nil
}

func getDeviceDriverFileName(driver uintptr, lpFilename *uint16, nSize uint32) error {
	r1, _, e1 := syscall.SyscallN(
		procGetDeviceDriverFileNameW.Addr(),
		driver,
		uintptr(unsafe.Pointer(lpFilename)),
		uintptr(nSize),
	)
	if r1 == 0 {
		if e1 != 0 {
			return e1
		}
		return syscall.EINVAL
	}
	return nil
}

func getDeviceDriverBaseName(driver uintptr, lpBaseName *uint16, nSize uint32) error {
	r1, _, e1 := syscall.SyscallN(
		procGetDeviceDriverBaseNameW.Addr(),
		driver,
		uintptr(unsafe.Pointer(lpBaseName)),
		uintptr(nSize),
	)
	if r1 == 0 {
		if e1 != 0 {
			return e1
		}
		return syscall.EINVAL
	}
	return nil
}

func getProcessMemoryInfo(h windows.Handle, mem *processMemoryCounters) error {
	r1, _, e1 := syscall.SyscallN(
		procGetProcessMemoryInfo.Addr(),
		uintptr(h),
		uintptr(unsafe.Pointer(mem)),
		uintptr(unsafe.Sizeof(*mem)),
	)
	if r1 == 0 {
		if e1 != 0 {
			return e1
		}
		return syscall.EINVAL
	}
	return nil
}
