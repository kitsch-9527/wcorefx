//go:build windows
// +build windows

package se

import (
	"fmt"
	"unsafe"

	"github.com/kitsch-9527/wcorefx/internal/dll/advapi32"
	"github.com/kitsch-9527/wcorefx/internal/dll/ntdll"
	"golang.org/x/sys/windows"
)

// Windows特权名称常量
const (
	debugName = "SeDebugPrivilege" // 调试权限名称
)

// 权限操作枚举类型（包内私有，仅内部使用）
type privilegeOperate int

// 枚举常量（包内私有）
const (
	privilegeUnknown privilegeOperate = iota // 0
	privilegeEnable                          // 1 - 启用权限
	privilegeDisable                         // 2 - 禁用权限
)

// VerifyPrivileges 验证文件签名
func VerifyPrivileges(path string) error {
	fileInfo := windows.WinTrustFileInfo{
		Size:     uint32(unsafe.Sizeof(windows.WinTrustFileInfo{})),
		FilePath: windows.StringToUTF16Ptr(path),
	}

	winT := windows.WinTrustData{
		Size:                            uint32(unsafe.Sizeof(windows.WinTrustData{})),
		UIChoice:                        windows.WTD_UI_NONE,
		RevocationChecks:                windows.WTD_REVOKE_NONE,
		UnionChoice:                     windows.WTD_CHOICE_FILE,
		StateAction:                     windows.WTD_STATEACTION_VERIFY,
		FileOrCatalogOrBlobOrSgnrOrCert: unsafe.Pointer(&fileInfo),
	}
	err := windows.WinVerifyTrustEx(0, &windows.WINTRUST_ACTION_GENERIC_VERIFY_V2, &winT)
	if err != nil {
		return fmt.Errorf("WinVerifyTrustEx failed: %v", err)
	}
	fmt.Println("文件签名验证通过")
	return nil
}

// SeEnableDebugPrivilege 启用调试权限
// 参数b: true使用匿名api方法，false使用标准API方法
// 返回值: 操作是否成功
func SeEnableDebugPrivilege(b bool) bool {
	if b {
		return SeEnablePrivilege("", windows.SE_GROUP_INTEGRITY)
	} else {
		return SeEnablePrivilege(debugName, 0)
	}
}

// 实现 String() 方法，方便打印枚举值对应的名称（类似 C 枚举的字符串化）
func NetJoinStatfmt(s advapi32.NETSETUP_JOIN_STATUS) string {
	switch s {
	case advapi32.NetSetupUnknownStatus:
		//return "NetSetupUnknownStatus"
		return "未知状态"
	case advapi32.NetSetupUnjoined:
		//return "NetSetupUnjoined"
		return "未加入域或工作组"
	case advapi32.NetSetupWorkgroupName:
		//return "NetSetupWorkgroupName"
		return "已加入工作组"
	case advapi32.NetSetupDomainName:
		//return "NetSetupDomainName"
		return "已加入域"
	default:
		return fmt.Sprintf("NETSETUP_JOIN_STATUS(%d)", s) // 未知值时返回原始数字
	}
}

// SeDisableDebugPrivilege 禁用调试权限
// 参数b: true使用匿名api方法，false使用标准API方法
// 返回值: 操作是否成功
func SeDisableDebugPrivilege(b bool) bool {
	if b {
		return SeDisablePrivilege("", windows.SE_GROUP_INTEGRITY)
	} else {
		return SeDisablePrivilege(debugName, 0)
	}
}

// SeEnablePrivilege 启用指定权限
// 参数privName: 权限名称(API方式)，为空则使用匿名api
// 参数privNumber: 权限编号(原生方式)
// 返回值: 操作是否成功
func SeEnablePrivilege(privName string, privNumber int) bool {
	if privName != "" {
		err := processPrivilegeByApi(privName, privilegeEnable)
		if err != nil {
			fmt.Println(err)
			return false
		}
	} else {
		processPrivilegeByNative(uint32(privNumber), privilegeEnable)
	}

	return true
}

// SeDisablePrivilege 禁用指定权限
// 参数privName: 权限名称(API方式)，为空则使用匿名api
// 参数privNumber: 权限编号(原生方式)
// 返回值: 操作是否成功
func SeDisablePrivilege(privName string, privNumber int) bool {
	if privName != "" {
		err := processPrivilegeByApi(privName, privilegeDisable)
		if err != nil {
			fmt.Println(err)
			return false
		}
	} else {
		processPrivilegeByNative(uint32(privNumber), privilegeDisable)
	}

	return true
}

// processPrivilegeByApi 通过API方式处理权限（内部函数）
func processPrivilegeByApi(privName string, op privilegeOperate) error {
	//fmt.Println("processPrivilegeByApi:", privName, op)
	var token windows.Token
	hProcess := windows.CurrentProcess()
	err := windows.OpenProcessToken(hProcess, windows.TOKEN_ADJUST_PRIVILEGES|windows.TOKEN_QUERY, &token)
	if err != nil {
		return fmt.Errorf("OpenProcessToken failed: %v", err)
	}
	defer token.Close()
	var luid windows.LUID
	err = windows.LookupPrivilegeValue(nil, windows.StringToUTF16Ptr(privName), &luid)
	if err != nil {
		return fmt.Errorf("LookupPrivilegeValue failed: %v", err)

	}
	// 设置权限属性
	attr := uint32(0)
	if op == privilegeEnable {
		attr = windows.SE_PRIVILEGE_ENABLED
	} else {
		attr = windows.SE_PRIVILEGE_REMOVED
	}
	tp := windows.Tokenprivileges{
		PrivilegeCount: 1,
		Privileges: [1]windows.LUIDAndAttributes{
			{
				Luid:       luid,
				Attributes: attr,
			},
		},
	}
	err = windows.AdjustTokenPrivileges(token, false, &tp, 0, nil, nil)
	if err != nil {
		return fmt.Errorf("AdjustTokenPrivileges failed: %v", err)

	}
	// 执行权限调整
	return nil
}

// todo 此方法有问题无法关闭bug权限后续修复
// processPrivilegeByNative 通过未公布的ntdll.dllapi处理权限（内部函数）
func processPrivilegeByNative(privNumber uint32, op privilegeOperate) error {
	enable := (op == privilegeEnable)
	fmt.Println("processPrivilegeByNative:", privNumber, enable)
	var wasEnabled bool
	err := ntdll.RtlAdjustPrivilege(privNumber, enable, false, &wasEnabled)
	if err != nil {
		return fmt.Errorf("RtlAdjustPrivilege failed: %v", err)
	}
	fmt.Printf("权限操作结果 - 权限: %d, 操作: %v, 之前状态: %v \n",
		privNumber, op, wasEnabled)
	// 执行权限调整
	return nil
}

// CheckAdmin 检查当前进程是否具有管理员权限
// 返回值: 是否具有管理员权限以及可能的错误
// ture: 具有管理员权限 false: 没有管理员权限
func CheckAdmin() (bool, error) {
	ntAuthority := windows.SECURITY_NT_AUTHORITY
	var amdinGroup *windows.SID
	err := windows.AllocateAndInitializeSid(&ntAuthority, 2, windows.SECURITY_BUILTIN_DOMAIN_RID, windows.DOMAIN_ALIAS_RID_ADMINS, 0, 0, 0, 0, 0, 0, &amdinGroup)
	if err != nil {
		return false, fmt.Errorf("AllocateAndInitializeSid failed: %v", err)
	}
	defer windows.FreeSid(amdinGroup)
	isElevated, err := advapi32.CheckTokenMembership(0, amdinGroup)
	if err != nil {
		return false, fmt.Errorf("CheckTokenMembership failed: %v", err)
	}
	return isElevated, nil
}
func TokenDisplayAccount(token windows.Token) {
	// windows.GetTokenInformation(token, windows.TokenStatistics, nil, 0, nil)
	// windows.GetTokenInformation(token, windows.TokenSessionId, nil, 0, nil)
	// windows.GetTokenInformation(token, windows.TokenElevationType, nil, 0, nil)
	// windows.GetTokenInformation(token, windows.TokenGroupsAndPrivileges, nil, 0, nil)

	windows.GetTokenInformation(token, windows.TokenOrigin, nil, 0, nil)
}
func PrintSIDAttributes(l string, sidsAttr uint32) string {
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
	// 格式化输出，与原C++代码保持一致的格式
	return fmt.Sprintf("g:[%c%c%c%c%c%c%c] ", c1, c2, c3, c4, c5, c6, c7)
}

type PrivilegeDetail struct {
	Status string
	Name   string
}

func TokenDisplayAccountSid(u string, cont uint32, sids *windows.SIDAndAttributes) ([]PrivilegeDetail, error) {
	pList := make([]PrivilegeDetail, cont)
	for i := 0; i < int(cont); i++ {
		sidAttrAddr := uintptr(unsafe.Pointer(sids)) + uintptr(i)*unsafe.Sizeof(windows.SIDAndAttributes{})
		sid := (*windows.SIDAndAttributes)(unsafe.Pointer(sidAttrAddr))
		d, n, err := GetDonameInSid(sid.Sid)
		if err != nil {
			return nil, err
		}
		pList[i] = PrivilegeDetail{Status: PrintSIDAttributes(u, sid.Attributes), Name: fmt.Sprint(d, "/", n)}
	}
	return pList, nil
}

func GetDonameInSid(sid *windows.SID) (string, string, error) {

	// 查找账户名
	var (
		userNameSize uint32 = 0
		domainSize   uint32 = 0
		sidNameUse   uint32
	)
	err := windows.LookupAccountSid(nil, sid, nil, &userNameSize, nil, &domainSize, &sidNameUse)
	if err != nil {
		if err != windows.ERROR_INSUFFICIENT_BUFFER {
			return "", "", fmt.Errorf("LookupAccountSid failed: %w", err)
		}
	}
	var (
		userName   = make([]uint16, userNameSize)
		domainName = make([]uint16, domainSize)
	)
	err = windows.LookupAccountSid(
		nil,
		sid,
		&userName[0],
		&userNameSize,
		&domainName[0],
		&domainSize,
		&sidNameUse,
	)
	if err != nil {
		return "", "", fmt.Errorf("LookupAccountSid failed: %w", err)
	}
	// 转换为字符串并返回 domain\username 格式
	return windows.UTF16ToString(domainName[:domainSize]), windows.UTF16ToString(userName[:userNameSize]), nil
}

// PrintPrivilegeStatus 格式化打印权限属性状态
// 参数：privAttr 是权限的Attributes值
func PrintPrivilegeStatus(privAttr uint32) string {
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

	// 格式化输出，与原C++代码保持一致的格式
	return fmt.Sprintf("P:[%c%c%c%c]    ", c1, c2, c3, c4)
}

func GetTokenInformation(token windows.Token, infoClass uint32) ([]byte, error) {
	var size uint32
	err := windows.GetTokenInformation(token, infoClass, nil, 0, &size)
	if err != nil {
		if err != windows.ERROR_INSUFFICIENT_BUFFER {
			return nil, fmt.Errorf("GetTokenInformation failed: %w", err)
		}
	}
	// 分配缓冲区
	buffer := make([]byte, size)
	err = windows.GetTokenInformation(token, infoClass, &buffer[0], size, &size)
	if err != nil {
		return nil, fmt.Errorf("GetTokenInformation failed: %w", err)
	}
	return buffer, nil
}

func GetTokenPrivilegeNames(tokenGroups advapi32.TokenGroupsAndPrivileges) ([]PrivilegeDetail, error) {
	tokenGroupsCount := tokenGroups.PrivilegeCount
	groups := make([]PrivilegeDetail, tokenGroupsCount)
	for i := uint32(0); i < tokenGroupsCount; i++ {
		//TODO 后续更改为安全封装
		liuAttrAddr := uintptr(unsafe.Pointer(tokenGroups.Privileges)) + uintptr(i)*unsafe.Sizeof(windows.LUIDAndAttributes{})
		//内存指针转换
		luid := (*windows.LUIDAndAttributes)(unsafe.Pointer(liuAttrAddr))
		attributes := luid.Attributes
		name, err := LookupPrivilegeNameByLuid(luid.Luid)
		if err != nil {
			return nil, err
		}
		groups[i] = PrivilegeDetail{
			Status: PrintPrivilegeStatus(attributes),
			Name:   name,
		}
	}
	return groups, nil
}

func LookupPrivilegeNameByLuid(luid windows.LUID) (string, error) {
	var (
		Name            = make([]uint16, 256)
		NameSize uint32 = 0
	)
	err := advapi32.LookupPrivilegeName(
		nil,
		&luid,
		nil,
		&NameSize,
	)
	if err != windows.ERROR_INSUFFICIENT_BUFFER {
		return "", fmt.Errorf("LookupPrivilegeName failed: %w", err)
	}
	err = advapi32.LookupPrivilegeName(
		nil,
		&luid,
		&Name[0],
		&NameSize,
	)
	if err != nil {
		return "", fmt.Errorf("LookupPrivilegeName failed: %w", err)
	}
	return windows.UTF16ToString(Name[:NameSize]), nil
}

func GetJoinInformation() {
	var (
		server     uint16
		name       *uint16
		bufferByte uint32
		status     advapi32.NETSETUP_JOIN_STATUS
	)
	err := windows.NetGetJoinInformation(&server, &name, &bufferByte)
	if err != nil {
		fmt.Println("NetGetJoinInformation failed:", err)
		return
	}
	status = advapi32.NETSETUP_JOIN_STATUS(bufferByte)
	fmt.Println("NetGetJoinInformation succeeded:", server, windows.UTF16PtrToString(name), bufferByte, NetJoinStatfmt(status))
}
