//go:build windows
// +build windows

package se

import (
	"syscall"
	"unsafe"

	"github.com/kitsch-9527/wcorefx/comm"
	"golang.org/x/sys/windows"
	win "golang.org/x/sys/windows"
)

var (
	modNtdll    = win.NewLazySystemDLL("ntdll.dll")
	modsamlib   = syscall.NewLazyDLL("samlib.dll") // 用于获取用户组信息
	modadvapi32 = windows.NewLazySystemDLL("Advapi32.dll")
)

var (
	rtlAdjustPrivilege       = modNtdll.NewProc("RtlAdjustPrivilege")
	procLookupPrivilegeNameW = modadvapi32.NewProc("LookupPrivilegeNameW")
	procCheckTokenMembership = modadvapi32.NewProc("CheckTokenMembership")
)

// RtlAdjustPrivilege
func RtlAdjustPrivilege(priv uint32, enable bool, current bool, enabled *bool) (err error) {
	r1, _, e1 := syscall.Syscall6(rtlAdjustPrivilege.Addr(), 4, uintptr(priv), comm.Boo2Ptr(enable), comm.Boo2Ptr(current), uintptr(unsafe.Pointer(enabled)), 0, 0)
	if r1 == 0 {
		err = comm.ErrnoErr(e1)
	}
	return
}

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

type TokenGroupsAndPrivileges struct {
	SidCount            uint32                     // SID的数量
	SidLength           uint32                     // SID结构的总长度
	Sids                *windows.SIDAndAttributes  // 指向SID_AND_ATTRIBUTES数组的指针
	RestrictedSidCount  uint32                     // 受限制的SID的数量
	RestrictedSidLength uint32                     // 受限制的SID结构的总长度
	RestrictedSids      *windows.SIDAndAttributes  // 指向受限制的SID_AND_ATTRIBUTES数组的指针
	PrivilegeCount      uint32                     // 特权的数量
	PrivilegeLength     uint32                     // 特权结构的总长度
	Privileges          *windows.LUIDAndAttributes // 指向LUID_AND_ATTRIBUTES数组的指针
	AuthenticationId    windows.LUID
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
