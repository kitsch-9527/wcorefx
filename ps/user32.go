//go:build windows

package ps

import (
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/kitsch-9527/wcorefx/internal/winapi"
)

var (
	procEnumWindows     = winapi.NewProc("user32.dll", "EnumWindows")
	procGetWindowTextW  = winapi.NewProc("user32.dll", "GetWindowTextW")
	procGetClassNameW   = winapi.NewProc("user32.dll", "GetClassNameW")
	procIsWindowVisible = winapi.NewProc("user32.dll", "IsWindowVisible")
)

// WindowInfo 表示窗口信息
type WindowInfo struct {
	HWND      uintptr
	Title     string
	ClassName string
	Visible   bool
}

// Windows 返回所有顶层窗口列表
func Windows() ([]WindowInfo, error) {
	var wins []WindowInfo
	cb := syscall.NewCallback(func(hwnd uintptr, lparam uintptr) uintptr {
		var buf [512]uint16
		var classBuf [256]uint16

		procGetWindowTextW.CallRet(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
		title := windows.UTF16ToString(buf[:])

		procGetClassNameW.CallRet(hwnd, uintptr(unsafe.Pointer(&classBuf[0])), uintptr(len(classBuf)))
		className := windows.UTF16ToString(classBuf[:])

		visible, _ := procIsWindowVisible.CallRet(hwnd)

		wins = append(wins, WindowInfo{
			HWND:      hwnd,
			Title:     title,
			ClassName: className,
			Visible:   visible != 0,
		})
		return 1
	})

	r1, _ := procEnumWindows.CallRet(cb, 0)
	if r1 == 0 {
		return nil, fmt.Errorf("EnumWindows failed")
	}

	return wins, nil
}
