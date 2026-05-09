package sec

import (
	"fmt"

	"golang.org/x/sys/windows"
)

// FormatSIDAttributes 格式化 SID 属性标志为可读字符串。
//   label    - SID 类型标签（如 "group"、"alias" 等），用于上下文标识。
//   sidsAttr - SID 属性标志位组合。
//   返回 - 格式化后的 SID 属性字符串，如 "g:[M DE   ]"。
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

// FormatPrivilegeStatus 格式化权限属性标志为可读字符串。
//   privAttr - 权限属性标志位组合。
//   返回 - 格式化后的权限状态字符串，如 "P:[ DE  ]"。
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

// FormatJoinStatus 格式化域加入状态为可读字符串。
//   s - 要格式化的域加入状态枚举值。
//   返回 - 对应的可读状态描述字符串。
func FormatJoinStatus(s NETSETUP_JOIN_STATUS) string {
	switch s {
	case NetSetupUnknownStatus:
		return "NetSetupUnknownStatus"
	case NetSetupUnjoined:
		return "NetSetupUnjoined"
	case NetSetupWorkgroupName:
		return "NetSetupWorkgroupName"
	case NetSetupDomainName:
		return "NetSetupDomainName"
	default:
		return fmt.Sprintf("NETSETUP_JOIN_STATUS(%d)", s)
	}
}
