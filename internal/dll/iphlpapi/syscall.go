package iphlpapi

import (
	"syscall"
	"unsafe"

	"github.com/kitsch-9527/wcorefx/internal/common"
)

// GetExtendedTcpTable 封装Windows API调用
func GetExtendedTcpTable(table *byte, size *uint32, sort bool, af uint32, tableClass uint32, reserved uint32) (ret error) {
	r0, _, _ := syscall.SyscallN(
		procGetExtendedTcpTable.Addr(),
		uintptr(unsafe.Pointer(table)),
		uintptr(unsafe.Pointer(size)),
		common.Boo2Ptr(sort),
		uintptr(af),
		uintptr(tableClass),
		uintptr(reserved),
	)
	if r0 != 0 {
		ret = syscall.Errno(r0)
	}
	return
}

// GetExtendedUdpTable 封装Windows API调用
func GetExtendedUdpTable(table *byte, size *uint32, sort bool, af uint32, tableClass uint32, reserved uint32) (ret error) {
	r0, _, _ := syscall.SyscallN(
		procGetExtendedUdpTable.Addr(),
		uintptr(unsafe.Pointer(table)),
		uintptr(unsafe.Pointer(size)),
		common.Boo2Ptr(sort),
		uintptr(af),
		uintptr(tableClass),
		uintptr(reserved),
	)
	if r0 != 0 {
		ret = syscall.Errno(r0)
	}
	return
}
