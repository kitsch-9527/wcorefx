//go:build windows

package os

import (
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/kitsch-9527/wcorefx/internal/winapi"
)

var (
	procWTSQuerySessionInformationW = winapi.NewProc("wtsapi32.dll", "WTSQuerySessionInformationW")
	procWTSFreeMemory              = winapi.NewProc("wtsapi32.dll", "WTSFreeMemory")
)

const (
	WTS_CURRENT_SESSION       = ^uint32(0)
	WTS_CURRENT_SERVER_HANDLE = 0
	WTSUserName               = 5
)

func wtsQuerySessionInformation(sessionID uint32) (string, error) {
	var name *uint16
	var size uint32
	err := procWTSQuerySessionInformationW.Call(
		WTS_CURRENT_SERVER_HANDLE,
		uintptr(sessionID),
		WTSUserName,
		uintptr(unsafe.Pointer(&name)),
		uintptr(unsafe.Pointer(&size)),
	)
	if err != nil {
		return "", err
	}
	defer procWTSFreeMemory.Call(uintptr(unsafe.Pointer(name)))
	return windows.UTF16PtrToString(name), nil
}
