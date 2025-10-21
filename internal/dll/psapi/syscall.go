package psapi

import (
	"syscall"
	"unsafe"

	"github.com/kitsch-9527/wcorefx/internal/common"
	"golang.org/x/sys/windows"
)

func QueryWorkingSet(h windows.Handle, pv uintptr, cb uint32) (err error) {
	r1, _, e1 := syscall.SyscallN(procQueryWorkingSet.Addr(), uintptr(h), uintptr(pv), uintptr(cb))
	if r1 == 0 {
		err = windows.Errno(e1)
	}
	return
}

func GetProcessMemoryInfo(h windows.Handle, mem *PROCESS_MEMORY_COUNTERS) (err error) {
	r1, _, e1 := syscall.SyscallN(procGetProcessMemoryInfo.Addr(), uintptr(h), uintptr(unsafe.Pointer(mem)), uintptr(unsafe.Sizeof(*mem)))
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

// EnumDeviceDrivers 枚举系统中所有驱动程序
func EnumDeviceDrivers(drivers *uintptr, cb uint32, lpcNeeded *uint32) (err error) {
	r1, _, e1 := syscall.SyscallN(procEnumDeviceDrivers.Addr(), uintptr(unsafe.Pointer(drivers)), uintptr(cb), uintptr(unsafe.Pointer(lpcNeeded)))
	if r1 == 0 {
		err = common.ErrnoErr(e1)
	}
	return
}

// GetDeviceDriverBaseName 获得指定驱动程序的名称
func GetDeviceDriverBaseName(driver uintptr, lpBaseName *uint16, nSize uint32) (err error) {
	r1, _, e1 := syscall.SyscallN(procGetDeviceDriverBaseNameW.Addr(), uintptr(driver), uintptr(unsafe.Pointer(lpBaseName)), uintptr(nSize))
	if r1 == 0 {
		err = common.ErrnoErr(e1)
	}
	return
}

// GetDeviceDriverFileName 获得指定驱动程序的路径
func GetDeviceDriverFileName(driver uintptr, lpFilename *uint16, nSize uint32) (err error) {
	r1, _, e1 := syscall.SyscallN(procGetDeviceDriverFileNameW.Addr(), uintptr(driver), uintptr(unsafe.Pointer(lpFilename)), uintptr(nSize))
	if r1 == 0 {
		err = common.ErrnoErr(e1)
	}
	return
}
