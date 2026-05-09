//go:build windows

package sec

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var modntdll = windows.NewLazySystemDLL("ntdll.dll")

var procRtlAdjustPrivilege = modntdll.NewProc("RtlAdjustPrivilege")

// RtlAdjustPrivilege 封装 ntdll 的 RtlAdjustPrivilege 原生 API，用于启用或禁用指定权限。
//   privilege - 要调整的权限编号。
//   enable    - true 表示启用权限，false 表示禁用。
//   current   - true 表示仅影响当前线程，false 表示影响整个进程。
//   enabled   - 输出参数，接收操作执行前权限的启用状态。
//   返回 - 操作成功返回 nil，否则返回错误码。
func RtlAdjustPrivilege(privilege uint32, enable, current bool, enabled *bool) error {
	r1, _, _ := syscall.SyscallN(procRtlAdjustPrivilege.Addr(),
		uintptr(privilege),
		boolToUintptr(enable),
		boolToUintptr(current),
		uintptr(unsafe.Pointer(enabled)))
	if r1 != 0 {
		return syscall.Errno(r1)
	}
	return nil
}

// boolToUintptr converts a bool to a uintptr (1 for true, 0 for false).
func boolToUintptr(b bool) uintptr {
	if b {
		return 1
	}
	return 0
}
