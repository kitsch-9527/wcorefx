//go:build windows
// +build windows

package os

import (
	"fmt"

	"github.com/kitsch-9527/wcorefx/winapi/dll/kernel32"
	"github.com/kitsch-9527/wcorefx/winapi/dll/wtsapi32"
	"golang.org/x/sys/windows"
)

func WinDir() (string, error) {
	n, err := windows.GetWindowsDirectory()
	if err != nil {
		return "", fmt.Errorf("GetWindowsDirectory failed: %w", err)
	}
	return n, nil
}

// 根据会话ID获取用户名
func SessionUserName(sessionId uint32) (string, error) {
	name, err := wtsapi32.WTSQuerySessionInformation(wtsapi32.WTS_CURRENT_SERVER_HANDLE, sessionId, wtsapi32.WTSUSERNAME)
	if err != nil {
		return "", fmt.Errorf("WTSQuerySessionInformation failed: %w", err)
	}
	return name, nil
}

func GetGroupsBySid(sid *windows.SID) {

}

// GetCPUCount 获取CPU数量
func GetCPUCount() uint32 {
	systemInfo := kernel32.GetSystemInfo()
	return systemInfo.DwNumberOfProcessors
}
