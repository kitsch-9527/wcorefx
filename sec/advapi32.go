//go:build windows

package sec

import (
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/kitsch-9527/wcorefx/internal/winapi"
)

var (
	procLookupPrivilegeNameW = winapi.NewProc("Advapi32.dll", "LookupPrivilegeNameW")
	procCheckTokenMembership = winapi.NewProc("Advapi32.dll", "CheckTokenMembership")
)

// NETSETUP_JOIN_STATUS 镜像 NetSetup.h 中的 C 枚举类型。
type NETSETUP_JOIN_STATUS uint32

const (
	NetSetupUnknownStatus NETSETUP_JOIN_STATUS = 0
	NetSetupUnjoined      NETSETUP_JOIN_STATUS = 1
	NetSetupWorkgroupName NETSETUP_JOIN_STATUS = 2
	NetSetupDomainName    NETSETUP_JOIN_STATUS = 3
)

// TokenGroupsAndPrivileges 镜像 Windows TOKEN_GROUPS_AND_PRIVILEGES 结构体。
type TokenGroupsAndPrivileges struct {
	SidCount            uint32
	SidLength           uint32
	Sids                *windows.SIDAndAttributes
	RestrictedSidCount  uint32
	RestrictedSidLength uint32
	RestrictedSids      *windows.SIDAndAttributes
	PrivilegeCount      uint32
	PrivilegeLength     uint32
	Privileges          *windows.LUIDAndAttributes
	AuthenticationId    windows.LUID
}

// ElevationInfo 镜像 Windows TOKEN_ELEVATION 结构体。
type ElevationInfo struct {
	TokenIsElevated uint32
}

// CheckTokenMembership 检查指定令牌是否属于指定的 SID 组。
func CheckTokenMembership(TokenHandle windows.Handle, SidToCheck *windows.SID) (bool, error) {
	var isMember uint32
	err := procCheckTokenMembership.Call(
		uintptr(TokenHandle),
		uintptr(unsafe.Pointer(SidToCheck)),
		uintptr(unsafe.Pointer(&isMember)),
	)
	if err != nil {
		return false, err
	}
	return isMember != 0, nil
}

// LookupPrivilegeName 封装 Advapi32 的 LookupPrivilegeNameW 函数。
func LookupPrivilegeName(lpSystemName *uint16, lpLuid *windows.LUID, lpName *uint16, cchName *uint32) error {
	return procLookupPrivilegeNameW.Call(
		uintptr(unsafe.Pointer(lpSystemName)),
		uintptr(unsafe.Pointer(lpLuid)),
		uintptr(unsafe.Pointer(lpName)),
		uintptr(unsafe.Pointer(cchName)),
	)
}
