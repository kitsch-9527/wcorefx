package advapi32

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// CheckTokenMembership 检查指定令牌是否属于特定安全组
// TokenHandle: 要检查的令牌句柄，nil表示当前进程令牌
// SidToCheck: 要检查的安全标识符(SID)
// 返回值: 是否属于该组以及可能的错误
func CheckTokenMembership(TokenHandle windows.Handle, SidToCheck *windows.SID) (bool, error) {
	var isMember uint32

	// 调用Windows API
	r1, _, err := syscall.SyscallN(procCheckTokenMembership.Addr(),
		uintptr(TokenHandle),
		uintptr(unsafe.Pointer(SidToCheck)),
		uintptr(unsafe.Pointer(&isMember)),
	)

	if r1 == 0 {
		return false, err
	}

	return isMember != 0, nil
}

func LookupPrivilegeName(lpSystemName *uint16,
	lpLuid *windows.LUID,
	lpName *uint16,
	cchName *uint32) (err error) {
	r1, _, e1 := syscall.SyscallN(procLookupPrivilegeNameW.Addr(), uintptr(unsafe.Pointer(lpSystemName)), uintptr(unsafe.Pointer(lpLuid)), uintptr(unsafe.Pointer(lpName)), uintptr(unsafe.Pointer(cchName)))
	if r1 == 0 {
		err = windows.Errno(e1)
	}
	return
}
