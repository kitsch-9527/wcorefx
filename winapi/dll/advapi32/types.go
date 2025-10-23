package advapi32

import "golang.org/x/sys/windows"

// 定义枚举类型（基于 int，模拟 C 中的枚举）
type NETSETUP_JOIN_STATUS uint32

// 枚举常量（与 C 中枚举值一一对应）
const (
	NetSetupUnknownStatus NETSETUP_JOIN_STATUS = 0 // 未知状态
	NetSetupUnjoined      NETSETUP_JOIN_STATUS = 1 // 未加入域或工作组
	NetSetupWorkgroupName NETSETUP_JOIN_STATUS = 2 // 已加入工作组
	NetSetupDomainName    NETSETUP_JOIN_STATUS = 3 // 已加入域
)

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
