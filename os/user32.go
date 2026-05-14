//go:build windows

package os

import (
	"fmt"

	"golang.org/x/sys/windows"

	"github.com/kitsch-9527/wcorefx/internal/winapi"
)

var procExitWindowsEx = winapi.NewProc("user32.dll", "ExitWindowsEx")

func exitWindowsEx(flags uint32, reason uint32) error {
	_ = enableShutdownPrivilege()
	err := procExitWindowsEx.Call(uintptr(flags), uintptr(reason))
	if err != nil {
		return fmt.Errorf("ExitWindowsEx failed: %w", err)
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
