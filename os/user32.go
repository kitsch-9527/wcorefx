//go:build windows

package os

import (
	"golang.org/x/sys/windows"
)

var moduser32 = windows.NewLazySystemDLL("user32.dll")

var procExitWindowsEx = moduser32.NewProc("ExitWindowsEx")

func exitWindowsEx(flags uint32, reason uint32) error {
	_ = enableShutdownPrivilege()
	r1, _, _ := procExitWindowsEx.Call(uintptr(flags), uintptr(reason))
	if r1 == 0 {
		return windows.GetLastError()
	}
	return nil
}

func enableShutdownPrivilege() error {
	var token windows.Token
	hProcess := windows.CurrentProcess()
	err := windows.OpenProcessToken(hProcess, windows.TOKEN_ADJUST_PRIVILEGES|windows.TOKEN_QUERY, &token)
	if err != nil {
		return err
	}
	defer token.Close()

	var luid windows.LUID
	err = windows.LookupPrivilegeValue(nil, windows.StringToUTF16Ptr("SeShutdownPrivilege"), &luid)
	if err != nil {
		return err
	}

	tp := windows.Tokenprivileges{
		PrivilegeCount: 1,
		Privileges: [1]windows.LUIDAndAttributes{
			{Luid: luid, Attributes: windows.SE_PRIVILEGE_ENABLED},
		},
	}
	return windows.AdjustTokenPrivileges(token, false, &tp, 0, nil, nil)
}
