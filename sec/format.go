package sec

import (
	"fmt"

	"github.com/kitsch-9527/wcorefx/winapi/dll/advapi32"
	"golang.org/x/sys/windows"
)

func FormatSIDAttributes(label string, sidsAttr uint32) string {
	// 检查每个标志位并确定对应的字符
	c1 := ' '
	if sidsAttr&windows.SE_GROUP_MANDATORY != 0 {
		c1 = 'M'
	}

	c2 := ' '
	if sidsAttr&windows.SE_GROUP_ENABLED_BY_DEFAULT != 0 {
		c2 = 'D'
	}

	c3 := ' '
	if sidsAttr&windows.SE_GROUP_ENABLED != 0 {
		c3 = 'E'
	}

	c4 := ' '
	if sidsAttr&windows.SE_GROUP_OWNER != 0 {
		c4 = 'O'
	}

	c5 := ' '
	if sidsAttr&windows.SE_GROUP_USE_FOR_DENY_ONLY != 0 {
		c5 = 'U'
	}

	c6 := ' '
	if sidsAttr&windows.SE_GROUP_LOGON_ID != 0 {
		c6 = 'L'
	}

	c7 := ' '
	if sidsAttr&windows.SE_GROUP_RESOURCE != 0 {
		c7 = 'R'
	}
	return fmt.Sprintf("g:[%c%c%c%c%c%c%c]", c1, c2, c3, c4, c5, c6, c7)
}

// 权限状态格式化
func FormatPrivilegeStatus(privAttr uint32) string {
	// 检查每个标志位并确定对应的字符
	c1 := ' '
	if privAttr&windows.SE_PRIVILEGE_ENABLED_BY_DEFAULT != 0 {
		c1 = 'D'
	}

	c2 := ' '
	if privAttr&windows.SE_PRIVILEGE_ENABLED != 0 {
		c2 = 'E'
	}

	c3 := ' '
	if privAttr&windows.SE_PRIVILEGE_REMOVED != 0 {
		c3 = 'R'
	}

	c4 := ' '
	if privAttr&windows.SE_PRIVILEGE_USED_FOR_ACCESS != 0 {
		c4 = 'A'
	}
	return fmt.Sprintf("P:[%c%c%c%c]", c1, c2, c3, c4)
}

func FormatJoinStatus(s advapi32.NETSETUP_JOIN_STATUS) string {
	switch s {
	case advapi32.NetSetupUnknownStatus:
		return "NetSetupUnknownStatus"
		//return "未知状态"
	case advapi32.NetSetupUnjoined:
		return "NetSetupUnjoined"
		//return "未加入域或工作组"
	case advapi32.NetSetupWorkgroupName:
		return "NetSetupWorkgroupName"
		//return "已加入工作组"
	case advapi32.NetSetupDomainName:
		return "NetSetupDomainName"
		//return "已加入域"
	default:
		return fmt.Sprintf("NETSETUP_JOIN_STATUS(%d)", s) // 未知值时返回原始数字
	}
}
