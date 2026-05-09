//go:build windows

package sec

import (
	"os"
	"testing"

	"golang.org/x/sys/windows"
)

func TestIsAdmin(t *testing.T) {
	admin, err := IsAdmin()
	if err != nil {
		t.Fatalf("IsAdmin failed: %v", err)
	}
	t.Logf("IsAdmin = %v", admin)
}

func TestEnableDisableDebugPrivilege(t *testing.T) {
	// API-based path (AdjustTokenPrivileges)
	result := EnableDebugPrivilege(false)
	t.Logf("EnableDebugPrivilege(api) = %v", result)

	result = DisableDebugPrivilege(false)
	t.Logf("DisableDebugPrivilege(api) = %v", result)

	// Native path (RtlAdjustPrivilege)
	result = EnableDebugPrivilege(true)
	t.Logf("EnableDebugPrivilege(native) = %v", result)

	result = DisableDebugPrivilege(true)
	t.Logf("DisableDebugPrivilege(native) = %v", result)
}

func TestLookupSIDAccount(t *testing.T) {
	ntAuthority := windows.SECURITY_NT_AUTHORITY
	var adminGroup *windows.SID
	err := windows.AllocateAndInitializeSid(&ntAuthority, 2,
		windows.SECURITY_BUILTIN_DOMAIN_RID,
		windows.DOMAIN_ALIAS_RID_ADMINS,
		0, 0, 0, 0, 0, 0, &adminGroup)
	if err != nil {
		t.Fatalf("AllocateAndInitializeSid failed: %v", err)
	}
	defer windows.FreeSid(adminGroup)

	domain, name, err := LookupSIDAccount(adminGroup)
	if err != nil {
		t.Fatalf("LookupSIDAccount failed: %v", err)
	}
	t.Logf("Admin SID -> domain=%q name=%q", domain, name)
	if domain == "" && name == "" {
		t.Error("LookupSIDAccount returned empty domain and name")
	}
}

func TestVerifyFileSignature(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Skipf("Cannot determine executable path: %v", err)
	}

	err = VerifyFileSignature(exe)
	if err != nil {
		// The test binary may or may not be signed; log the result but
		// don't fail.
		t.Logf("VerifyFileSignature(%q) = %v (expected if unsigned)", exe, err)
	} else {
		t.Logf("VerifyFileSignature(%q) passed", exe)
	}
}

func TestTokenElevation(t *testing.T) {
	_, err := TokenElevation(windows.CurrentProcess())
	if err != nil {
		t.Fatalf("TokenElevation failed: %v", err)
	}
}

func TestGetDomainJoinInfo(t *testing.T) {
	info, err := GetDomainJoinInfo()
	if err != nil {
		t.Fatalf("GetDomainJoinInfo failed: %v", err)
	}
	t.Logf("Domain join info: %+v", info)
}
