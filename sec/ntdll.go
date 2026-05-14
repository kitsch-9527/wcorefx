//go:build windows

package sec

import (
	"unsafe"

	"github.com/kitsch-9527/wcorefx/internal/winapi"
)

var procRtlAdjustPrivilege = winapi.NewProc("ntdll.dll", "RtlAdjustPrivilege", winapi.ConvNTSTATUS)

// RtlAdjustPrivilege 封装 ntdll 的 RtlAdjustPrivilege 原生 API。
func RtlAdjustPrivilege(privilege uint32, enable, current bool, enabled *bool) error {
	return procRtlAdjustPrivilege.Call(
		uintptr(privilege),
		boolToUintptr(enable),
		boolToUintptr(current),
		uintptr(unsafe.Pointer(enabled)),
	)
}

func boolToUintptr(b bool) uintptr {
	if b {
		return 1
	}
	return 0
}
