//go:build windows

package os

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	modwtsapi32 = windows.NewLazySystemDLL("wtsapi32.dll")

	procWTSQuerySessionInformationW = modwtsapi32.NewProc("WTSQuerySessionInformationW")
	procWTSFreeMemory              = modwtsapi32.NewProc("WTSFreeMemory")
)

// WTS_CURRENT_SESSION 表示当前会话
const (
	WTS_CURRENT_SESSION       = ^uint32(0)
	// WTS_CURRENT_SERVER_HANDLE 表示当前服务器句柄
	WTS_CURRENT_SERVER_HANDLE = 0
	// WTSUserName 表示WTS用户名信息类
	WTSUserName               = 5
)

func wtsQuerySessionInformation(sessionID uint32) (string, error) {
	var name *uint16
	var size uint32
	ret, _, _ := procWTSQuerySessionInformationW.Call(
		WTS_CURRENT_SERVER_HANDLE,
		uintptr(sessionID),
		WTSUserName,
		uintptr(unsafe.Pointer(&name)),
		uintptr(unsafe.Pointer(&size)),
	)
	if ret == 0 {
		return "", syscall.GetLastError()
	}
	defer procWTSFreeMemory.Call(uintptr(unsafe.Pointer(name)))
	return windows.UTF16PtrToString(name), nil
}
