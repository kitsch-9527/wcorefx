//go:build windows
// +build windows

package sec

import (
	"fmt"
	"unsafe"

	"github.com/kitsch-9527/wcorefx/winapi/dll/advapi32"
	"github.com/kitsch-9527/wcorefx/winapi/dll/ntdll"
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

// 文件签名验证相关
func VerifyFileSignature(path string) error {
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
	return nil
}

// 调试权限相关
func EnableDebugPrivilege(useNative bool) bool {
	if useNative {
		return EnablePrivilege("", windows.SE_GROUP_INTEGRITY)
	} else {
		return EnablePrivilege(debugName, 0)
	}
}

func DisableDebugPrivilege(useNative bool) bool {
	if useNative {
		return DisablePrivilege("", windows.SE_GROUP_INTEGRITY)
	} else {
		return DisablePrivilege(debugName, 0)
	}
}

// 通用权限管理
func EnablePrivilege(name string, number int) bool {
	if name != "" {
		err := adjustPrivilegeByAPI(name, privilegeEnable)
		if err != nil {
			fmt.Println(err)
			return false
		}
	} else {
		adjustPrivilegeByNative(uint32(number), privilegeEnable)
	}

	return true
}

func DisablePrivilege(name string, number int) bool {
	if name != "" {
		err := adjustPrivilegeByAPI(name, privilegeDisable)
		if err != nil {
			fmt.Println(err)
			return false
		}
	} else {
		adjustPrivilegeByNative(uint32(number), privilegeDisable)
	}

	return true
}

// 内部辅助方法
func adjustPrivilegeByAPI(privName string, op privilegeOperate) error {
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
func adjustPrivilegeByNative(privNumber uint32, op privilegeOperate) error {
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

// 管理员权限检查
func IsAdmin() (bool, error) {
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

func DisplayTokenAccount(token windows.Token) {
	// windows.GetTokenInformation(token, windows.TokenStatistics, nil, 0, nil)
	// windows.GetTokenInformation(token, windows.TokenSessionId, nil, 0, nil)
	// windows.GetTokenInformation(token, windows.TokenElevationType, nil, 0, nil)
	// windows.GetTokenInformation(token, windows.TokenGroupsAndPrivileges, nil, 0, nil)

	windows.GetTokenInformation(token, windows.TokenOrigin, nil, 0, nil)
}

type PrivilegeDetail struct {
	Status string
	Name   string
}

func GetTokenAccountSIDs(userType string, count uint32, sids *windows.SIDAndAttributes) ([]PrivilegeDetail, error) {
	pList := make([]PrivilegeDetail, count)
	for i := 0; i < int(count); i++ {
		sidAttrAddr := uintptr(unsafe.Pointer(sids)) + uintptr(i)*unsafe.Sizeof(windows.SIDAndAttributes{})
		sid := (*windows.SIDAndAttributes)(unsafe.Pointer(sidAttrAddr))
		d, n, err := LookupSIDAccount(sid.Sid)
		if err != nil {
			return nil, err
		}
		pList[i] = PrivilegeDetail{Status: FormatSIDAttributes(userType, sid.Attributes), Name: fmt.Sprint(d, "/", n)}
	}
	return pList, nil
}

func LookupSIDAccount(sid *windows.SID) (domain, name string, err error) {

	// 查找账户名
	var (
		userNameSize uint32 = 0
		domainSize   uint32 = 0
		sidNameUse   uint32
	)
	err = windows.LookupAccountSid(nil, sid, nil, &userNameSize, nil, &domainSize, &sidNameUse)
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

// Token 信息获取
func GetTokenGroupsAndPrivileges(token windows.Token) (*advapi32.TokenGroupsAndPrivileges, error) {
	buffer, err := GetTokenInformation(token, windows.TokenGroupsAndPrivileges)
	if err != nil {
		return nil, err
	}
	return (*advapi32.TokenGroupsAndPrivileges)(unsafe.Pointer(&buffer[0])), nil
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
		name, err := LookupPrivilegeNameByLUID(luid.Luid)
		if err != nil {
			return nil, err
		}
		groups[i] = PrivilegeDetail{
			Status: FormatPrivilegeStatus(attributes),
			Name:   name,
		}
	}
	return groups, nil
}

func LookupPrivilegeNameByLUID(luid windows.LUID) (string, error) {
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

// 域信息相关
func GetDomainJoinInfo() {
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
	fmt.Println("NetGetJoinInformation succeeded:", server, windows.UTF16PtrToString(name), bufferByte, FormatJoinStatus(status))
}
