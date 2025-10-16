package ps

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modpsapi = windows.NewLazySystemDLL("psapi.dll")
	modntdll = windows.NewLazySystemDLL("ntdll.dll")
)

var (
	procQueryWorkingSet      = modpsapi.NewProc("QueryWorkingSet")
	procGetProcessMemoryInfo = modpsapi.NewProc("GetProcessMemoryInfo")
	procNtDuplicateObject    = modntdll.NewProc("NtDuplicateObject")
	procNtQueryObject        = modntdll.NewProc("NtQueryObject")
)

func QueryWorkingSet(h windows.Handle, pv uintptr, cb uint32) (err error) {
	r1, _, e1 := syscall.Syscall(procQueryWorkingSet.Addr(), 3, uintptr(h), uintptr(pv), uintptr(cb))
	if r1 == 0 {
		err = windows.Errno(e1)
	}
	return
}

func getProcessMemoryInfo(h windows.Handle, mem *PROCESS_MEMORY_COUNTERS) (err error) {
	r1, _, e1 := syscall.Syscall(procGetProcessMemoryInfo.Addr(), 3, uintptr(h), uintptr(unsafe.Pointer(mem)), uintptr(unsafe.Sizeof(*mem)))
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}

// NtDuplicateObject
func NtDuplicateObject(
	sourceProcessHandle windows.Handle,
	sourceHandle windows.Handle,
	targetProcessHandle windows.Handle,
	// targetHandle *windows.Handle,
	desiredAccess windows.ACCESS_MASK,
	handleAttributes uint32,
	options uint32,
) (windows.Handle, error) {
	var targetHandle windows.Handle
	r0, _, _ := syscall.SyscallN(procNtDuplicateObject.Addr(),
		uintptr(sourceProcessHandle),
		uintptr(sourceHandle),
		uintptr(targetProcessHandle),
		uintptr(unsafe.Pointer(&targetHandle)),
		uintptr(desiredAccess),
		uintptr(handleAttributes),
		uintptr(options))
	if r0 != 0 {
		return 0, syscall.Errno(r0)
	}
	return targetHandle, nil
}
func NtQueryObject(
	handle windows.Handle,
	objectInformationClass OBJECT_INFORMATION_CLASS,
	objectInformation uintptr,
	objectInformationLength uint32,
	returnLength *uint32,
) error {
	r0, _, _ := syscall.SyscallN(procNtQueryObject.Addr(),
		uintptr(handle),
		uintptr(objectInformationClass),
		uintptr(objectInformation),
		uintptr(objectInformationLength),
		uintptr(unsafe.Pointer(returnLength)))
	if r0 != 0 {
		return syscall.Errno(r0)
	}
	return nil
}
