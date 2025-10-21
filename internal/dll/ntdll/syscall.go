package ntdll

import (
	"syscall"
	"unsafe"

	"github.com/kitsch-9527/wcorefx/internal/common"
	"golang.org/x/sys/windows"
)

func NtDuplicateObject(
	sourceProcessHandle windows.Handle,
	sourceHandle windows.Handle,
	targetProcessHandle windows.Handle,
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

// RtlAdjustPrivilege
func RtlAdjustPrivilege(priv uint32, enable bool, current bool, enabled *bool) (err error) {
	r1, _, e1 := syscall.SyscallN(procRtlAdjustPrivilege.Addr(), uintptr(priv), common.Boo2Ptr(enable), common.Boo2Ptr(current), uintptr(unsafe.Pointer(enabled)))
	if r1 == 0 {
		err = common.ErrnoErr(e1)
	}
	return
}
