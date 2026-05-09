//go:build windows

package sec

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	debugName = "SeDebugPrivilege"
)

type privilegeOperate int

const (
	privilegeUnknown privilegeOperate = iota
	privilegeEnable
	privilegeDisable
)

// PrivilegeDetail 表示权限名称及其格式化状态的结构体。
type PrivilegeDetail struct {
	// Status 权限状态（如启用/禁用等）。
	Status string
	// Name 权限名称。
	Name   string
}

// Domain 表示域加入信息的结构体。
type Domain struct {
	// Server 域服务器名称。
	Server string
	// Name 域名。
	Name   string
	// Status 域加入状态。
	Status string
}

// VerifyFileSignature 使用 WinVerifyTrustEx 验证文件的数字签名。
//   path - 要验证的文件路径。
//   返回 - 验证成功返回 nil，否则返回错误信息。
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

// EnableDebugPrivilege 启用 SeDebugPrivilege，useNative 为 true 时使用 ntdll RtlAdjustPrivilege，否则使用标准 AdjustTokenPrivileges。
//   useNative - true 时使用 ntdll 原生 API，false 时使用标准 Windows API。
//   返回 - 操作成功返回 true，否则返回 false。
func EnableDebugPrivilege(useNative bool) bool {
	if useNative {
		return EnablePrivilege("", windows.SE_GROUP_INTEGRITY)
	}
	return EnablePrivilege(debugName, 0)
}

// DisableDebugPrivilege 禁用 SeDebugPrivilege，useNative 为 true 时使用 ntdll RtlAdjustPrivilege，否则使用标准 AdjustTokenPrivileges。
//   useNative - true 时使用 ntdll 原生 API，false 时使用标准 Windows API。
//   返回 - 操作成功返回 true，否则返回 false。
func DisableDebugPrivilege(useNative bool) bool {
	if useNative {
		return DisablePrivilege("", windows.SE_GROUP_INTEGRITY)
	}
	return DisablePrivilege(debugName, 0)
}

// EnablePrivilege 启用指定权限，name 非空时使用 AdjustTokenPrivileges，否则使用 ntdll RtlAdjustPrivilege。
//   name   - 权限名称（如 "SeDebugPrivilege"），非空时使用标准 API。
//   number - 权限编号，当 name 为空时通过 ntdll 原生 API 使用此编号。
//   返回 - 操作成功返回 true，否则返回 false。
func EnablePrivilege(name string, number int) bool {
	if name != "" {
		err := adjustPrivilegeByAPI(name, privilegeEnable)
		if err != nil {
			return false
		}
	} else {
		err := adjustPrivilegeByNative(uint32(number), privilegeEnable)
		if err != nil {
			return false
		}
	}
	return true
}

// DisablePrivilege 禁用指定权限，name 非空时使用 AdjustTokenPrivileges，否则使用 ntdll RtlAdjustPrivilege。
//   name   - 权限名称（如 "SeDebugPrivilege"），非空时使用标准 API。
//   number - 权限编号，当 name 为空时通过 ntdll 原生 API 使用此编号。
//   返回 - 操作成功返回 true，否则返回 false。
func DisablePrivilege(name string, number int) bool {
	if name != "" {
		err := adjustPrivilegeByAPI(name, privilegeDisable)
		if err != nil {
			return false
		}
	} else {
		err := adjustPrivilegeByNative(uint32(number), privilegeDisable)
		if err != nil {
			return false
		}
	}
	return true
}

// adjustPrivilegeByAPI adjusts a privilege via the standard Windows API
// (OpenProcessToken / LookupPrivilegeValue / AdjustTokenPrivileges).
func adjustPrivilegeByAPI(privName string, op privilegeOperate) error {
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
	return nil
}

// adjustPrivilegeByNative adjusts a privilege via the ntdll RtlAdjustPrivilege
// native API.
func adjustPrivilegeByNative(privNumber uint32, op privilegeOperate) error {
	enable := (op == privilegeEnable)
	var wasEnabled bool
	err := RtlAdjustPrivilege(privNumber, enable, false, &wasEnabled)
	if err != nil {
		return fmt.Errorf("RtlAdjustPrivilege failed: %v", err)
	}
	return nil
}

// IsAdmin 检查当前进程是否以管理员权限运行。
//   返回1 - 当前进程是否以管理员权限运行。
//   返回2 - 操作失败时返回错误信息。
func IsAdmin() (bool, error) {
	ntAuthority := windows.SECURITY_NT_AUTHORITY
	var adminGroup *windows.SID
	err := windows.AllocateAndInitializeSid(&ntAuthority, 2,
		windows.SECURITY_BUILTIN_DOMAIN_RID,
		windows.DOMAIN_ALIAS_RID_ADMINS,
		0, 0, 0, 0, 0, 0, &adminGroup)
	if err != nil {
		return false, fmt.Errorf("AllocateAndInitializeSid failed: %v", err)
	}
	defer windows.FreeSid(adminGroup)

	isElevated, err := CheckTokenMembership(0, adminGroup)
	if err != nil {
		return false, fmt.Errorf("CheckTokenMembership failed: %v", err)
	}
	return isElevated, nil
}

// TokenElevation 返回指定进程句柄的令牌提升状态。
//   procH - 目标进程句柄。
//   返回1 - ElevationInfo 包含令牌提升状态信息。
//   返回2 - 操作失败时返回错误信息。
func TokenElevation(procH windows.Handle) (ElevationInfo, error) {
	var (
		tokenH    windows.Token
		elevation ElevationInfo
		returnLen uint32
	)

	err := windows.OpenProcessToken(procH, windows.TOKEN_QUERY, &tokenH)
	if err != nil {
		return elevation, err
	}
	defer tokenH.Close()

	err = windows.GetTokenInformation(
		tokenH,
		windows.TokenElevation,
		(*byte)(unsafe.Pointer(&elevation)),
		uint32(unsafe.Sizeof(elevation)),
		&returnLen,
	)
	if err != nil {
		return elevation, fmt.Errorf("GetTokenInformation failed: %w", err)
	}
	return elevation, nil
}

// GetTokenAccountSIDs 将 SID 列表解析为人类可读的账户名称。
//   userType - SID 类型标签，用于格式化 SID 属性。
//   count    - SID 结构的数量。
//   sids     - SID 属性数组的指针。
//   返回 - 解析后的 PrivilegeDetail 切片，失败时返回错误。
func GetTokenAccountSIDs(userType string, count uint32, sids *windows.SIDAndAttributes) ([]PrivilegeDetail, error) {
	pList := make([]PrivilegeDetail, count)
	for i := 0; i < int(count); i++ {
		sid := &unsafe.Slice(sids, count)[i]
		d, n, err := LookupSIDAccount(sid.Sid)
		if err != nil {
			return nil, err
		}
		pList[i] = PrivilegeDetail{
			Status: FormatSIDAttributes(userType, sid.Attributes),
			Name:   fmt.Sprint(d, "/", n),
		}
	}
	return pList, nil
}

// LookupSIDAccount 查找指定 SID 对应的域名和账户名。
//   sid - 要查找的安全标识符（SID）。
//   返回1 - 域名。
//   返回2 - 账户名。
//   返回3 - 操作失败时返回错误信息。
func LookupSIDAccount(sid *windows.SID) (domain, name string, err error) {
	var (
		userNameSize uint32
		domainSize   uint32
		sidNameUse   uint32
	)

	err = windows.LookupAccountSid(nil, sid, nil, &userNameSize, nil, &domainSize, &sidNameUse)
	if err != nil {
		if err != windows.ERROR_INSUFFICIENT_BUFFER {
			return "", "", fmt.Errorf("LookupAccountSid failed: %w", err)
		}
	}

	userName := make([]uint16, userNameSize)
	domainName := make([]uint16, domainSize)

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

	return windows.UTF16ToString(domainName[:domainSize]), windows.UTF16ToString(userName[:userNameSize]), nil
}

// GetTokenGroupsAndPrivileges 获取指定令牌的 TOKEN_GROUPS_AND_PRIVILEGES 结构体。
//   token - 目标令牌对象。
//   返回 - 解析后的 TokenGroupsAndPrivileges 指针，失败时返回错误。
func GetTokenGroupsAndPrivileges(token windows.Token) (*TokenGroupsAndPrivileges, error) {
	buffer, err := GetTokenInformation(token, windows.TokenGroupsAndPrivileges)
	if err != nil {
		return nil, err
	}
	return (*TokenGroupsAndPrivileges)(unsafe.Pointer(&buffer[0])), nil
}

// GetTokenInformation 以原始字节缓冲区形式检索令牌信息。
//   token     - 目标令牌对象。
//   infoClass - 要检索的令牌信息类别（如 TokenGroupsAndPrivileges）。
//   返回 - 包含令牌信息的字节切片，失败时返回错误。
func GetTokenInformation(token windows.Token, infoClass uint32) ([]byte, error) {
	var size uint32
	err := windows.GetTokenInformation(token, infoClass, nil, 0, &size)
	if err != nil {
		if err != windows.ERROR_INSUFFICIENT_BUFFER {
			return nil, fmt.Errorf("GetTokenInformation failed: %w", err)
		}
	}

	buffer := make([]byte, size)
	err = windows.GetTokenInformation(token, infoClass, &buffer[0], size, &size)
	if err != nil {
		return nil, fmt.Errorf("GetTokenInformation failed: %w", err)
	}
	return buffer, nil
}

// GetTokenPrivilegeNames 将 TokenGroupsAndPrivileges 中的 LUID 解析为可读的权限名称和状态。
//   tokenGroups - 包含权限 LUID 和属性的令牌组信息。
//   返回 - 解析后的 PrivilegeDetail 切片，失败时返回错误。
func GetTokenPrivilegeNames(tokenGroups TokenGroupsAndPrivileges) ([]PrivilegeDetail, error) {
	tokenGroupsCount := tokenGroups.PrivilegeCount
	groups := make([]PrivilegeDetail, tokenGroupsCount)
	for i := uint32(0); i < tokenGroupsCount; i++ {
		luid := &unsafe.Slice(tokenGroups.Privileges, tokenGroupsCount)[i]
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

// LookupPrivilegeNameByLUID 将 LUID 解析为对应的显示名称。
//   luid - 要查询的权限 LUID。
//   返回 - 权限显示名称，失败时返回错误。
func LookupPrivilegeNameByLUID(luid windows.LUID) (string, error) {
	var (
		Name            = make([]uint16, 256)
		NameSize uint32
	)

	err := LookupPrivilegeName(nil, &luid, nil, &NameSize)
	if err != windows.ERROR_INSUFFICIENT_BUFFER {
		return "", fmt.Errorf("LookupPrivilegeName size query failed: %w", err)
	}

	err = LookupPrivilegeName(nil, &luid, &Name[0], &NameSize)
	if err != nil {
		return "", fmt.Errorf("LookupPrivilegeName failed: %w", err)
	}

	return windows.UTF16ToString(Name[:NameSize]), nil
}

// GetDomainJoinInfo 检索本地计算机的域加入状态。
//   返回 - Domain 包含服务器、域名和加入状态信息，失败时返回错误。
func GetDomainJoinInfo() (Domain, error) {
	var (
		server     uint16
		name       *uint16
		bufferByte uint32
		status     NETSETUP_JOIN_STATUS
	)

	err := windows.NetGetJoinInformation(&server, &name, &bufferByte)
	if err != nil {
		return Domain{}, fmt.Errorf("NetGetJoinInformation failed: %v", err)
	}

	status = NETSETUP_JOIN_STATUS(bufferByte)
	statusType := FormatJoinStatus(status)
	return Domain{
		Server: windows.UTF16PtrToString(&server),
		Name:   windows.UTF16PtrToString(name),
		Status: statusType,
	}, nil
}
