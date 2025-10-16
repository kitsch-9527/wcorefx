//go:build windows
// +build windows

package os

import (
	"fmt"

	"golang.org/x/sys/windows"
	win "golang.org/x/sys/windows"
)

func WinDir() (string, error) {
	n, err := win.GetWindowsDirectory()
	if err != nil {
		return "", fmt.Errorf("GetWindowsDirectory failed: %w", err)
	}
	return n, nil
}

// 根据会话ID获取用户名
func SessionUserName(sessionId uint32) (string, error) {
	name, err := procWTSQuerySessionInformation(WTS_CURRENT_SERVER_HANDLE, sessionId, WTSUSERNAME)
	if err != nil {
		return "", fmt.Errorf("WTSQuerySessionInformation failed: %w", err)
	}
	return name, nil
}

func GetGroupsBySid(sid *windows.SID) {

}

// GetProcessorNumber 获取CPU数量
func GetProcessorNumber() uint32 {
	systemInfo := GetSystemInfo()
	return systemInfo.DwNumberOfProcessors
}
