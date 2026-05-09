//go:build windows

package sec

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var modadvapi32 = windows.NewLazySystemDLL("Advapi32.dll")

var (
	procLookupPrivilegeNameW = modadvapi32.NewProc("LookupPrivilegeNameW")
	procCheckTokenMembership = modadvapi32.NewProc("CheckTokenMembership")
)

// NETSETUP_JOIN_STATUS 镜像 NetSetup.h 中的 C 枚举类型。
type NETSETUP_JOIN_STATUS uint32

const (
	// NetSetupUnknownStatus 域状态未知。
	NetSetupUnknownStatus NETSETUP_JOIN_STATUS = 0
	// NetSetupUnjoined 计算机未加入域或工作组。
	NetSetupUnjoined      NETSETUP_JOIN_STATUS = 1
	// NetSetupWorkgroupName 计算机已加入工作组。
	NetSetupWorkgroupName NETSETUP_JOIN_STATUS = 2
	// NetSetupDomainName 计算机已加入域。
	NetSetupDomainName    NETSETUP_JOIN_STATUS = 3
)

// TokenGroupsAndPrivileges 镜像 Windows TOKEN_GROUPS_AND_PRIVILEGES 结构体。
type TokenGroupsAndPrivileges struct {
	// SidCount SID 条目数量。
	SidCount            uint32
	// SidLength SID 属性数组的长度（字节）。
	SidLength           uint32
	// Sids SID 属性数组的指针。
	Sids                *windows.SIDAndAttributes
	// RestrictedSidCount 受限 SID 数量。
	RestrictedSidCount  uint32
	// RestrictedSidLength 受限 SID 属性数组长度（字节）。
	RestrictedSidLength uint32
	// RestrictedSids 受限 SID 属性数组的指针。
	RestrictedSids      *windows.SIDAndAttributes
	// PrivilegeCount 权限条目数量。
	PrivilegeCount      uint32
	// PrivilegeLength 权限属性数组长度（字节）。
	PrivilegeLength     uint32
	// Privileges 权限属性数组的指针。
	Privileges          *windows.LUIDAndAttributes
	// AuthenticationId 认证标识符 LUID。
	AuthenticationId    windows.LUID
}

// ElevationInfo 镜像 Windows TOKEN_ELEVATION 结构体。
type ElevationInfo struct {
	// TokenIsElevated 令牌是否已提升（非零表示已提升）。
	TokenIsElevated uint32
}

// CheckTokenMembership 检查指定令牌是否属于指定的 SID 组。
//   TokenHandle - 要检查的令牌句柄（0 表示当前进程令牌）。
//   SidToCheck - 要检查的 SID 指针。
//   返回1 - 令牌是否属于该 SID 组。
//   返回2 - 操作失败时返回错误信息。
func CheckTokenMembership(TokenHandle windows.Handle, SidToCheck *windows.SID) (bool, error) {
	var isMember uint32
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

// LookupPrivilegeName 封装 Advapi32 的 LookupPrivilegeNameW 函数。
//   lpSystemName - 系统名称指针（nil 表示本地系统）。
//   lpLuid     - 要查询的 LUID 指针。
//   lpName     - 接收名称的缓冲区指针。
//   cchName    - 接收名称缓冲区的大小（字符数），返回后为实际字符数。
//   返回 - 操作成功返回 nil，否则返回错误码。
func LookupPrivilegeName(lpSystemName *uint16, lpLuid *windows.LUID, lpName *uint16, cchName *uint32) error {
	r1, _, e1 := syscall.SyscallN(procLookupPrivilegeNameW.Addr(),
		uintptr(unsafe.Pointer(lpSystemName)),
		uintptr(unsafe.Pointer(lpLuid)),
		uintptr(unsafe.Pointer(lpName)),
		uintptr(unsafe.Pointer(cchName)))
	if r1 == 0 {
		return windows.Errno(e1)
	}
	return nil
}
