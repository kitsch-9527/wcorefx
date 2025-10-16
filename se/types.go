package se

import "fmt"

// 定义枚举类型（基于 int，模拟 C 中的枚举）
type NETSETUP_JOIN_STATUS uint32

// 枚举常量（与 C 中枚举值一一对应）
const (
	NetSetupUnknownStatus NETSETUP_JOIN_STATUS = 0 // 未知状态
	NetSetupUnjoined      NETSETUP_JOIN_STATUS = 1 // 未加入域或工作组
	NetSetupWorkgroupName NETSETUP_JOIN_STATUS = 2 // 已加入工作组
	NetSetupDomainName    NETSETUP_JOIN_STATUS = 3 // 已加入域
)

// 实现 String() 方法，方便打印枚举值对应的名称（类似 C 枚举的字符串化）
func (s NETSETUP_JOIN_STATUS) String() string {
	switch s {
	case NetSetupUnknownStatus:
		//return "NetSetupUnknownStatus"
		return "未知状态"
	case NetSetupUnjoined:
		//return "NetSetupUnjoined"
		return "未加入域或工作组"
	case NetSetupWorkgroupName:
		//return "NetSetupWorkgroupName"
		return "已加入工作组"
	case NetSetupDomainName:
		//return "NetSetupDomainName"
		return "已加入域"
	default:
		return fmt.Sprintf("NETSETUP_JOIN_STATUS(%d)", s) // 未知值时返回原始数字
	}
}
