//go:build windows

package netapi

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var modiphlpapi = windows.NewLazySystemDLL("iphlpapi.dll")

var (
	procGetExtendedTcpTable   = modiphlpapi.NewProc("GetExtendedTcpTable")
	procGetExtendedUdpTable   = modiphlpapi.NewProc("GetExtendedUdpTable")
	procGetAdaptersAddresses  = modiphlpapi.NewProc("GetAdaptersAddresses")
	procGetIpNetTable2        = modiphlpapi.NewProc("GetIpNetTable2")
	procGetIpForwardTable2    = modiphlpapi.NewProc("GetIpForwardTable2")
	procFreeMibTable          = modiphlpapi.NewProc("FreeMibTable")
)

func getExtendedTcpTable(table *byte, size *uint32, sort bool, af, tableClass, reserved uint32) error {
	in := uint32(0)
	if sort {
		in = 1
	}
	r1, _, _ := syscall.SyscallN(procGetExtendedTcpTable.Addr(),
		uintptr(unsafe.Pointer(table)),
		uintptr(unsafe.Pointer(size)),
		uintptr(in),
		uintptr(af),
		uintptr(tableClass),
		uintptr(reserved),
	)
	if r1 != 0 {
		return syscall.Errno(r1)
	}
	return nil
}

func getExtendedUdpTable(table *byte, size *uint32, sort bool, af, tableClass, reserved uint32) error {
	in := uint32(0)
	if sort {
		in = 1
	}
	r1, _, _ := syscall.SyscallN(procGetExtendedUdpTable.Addr(),
		uintptr(unsafe.Pointer(table)),
		uintptr(unsafe.Pointer(size)),
		uintptr(in),
		uintptr(af),
		uintptr(tableClass),
		uintptr(reserved),
	)
	if r1 != 0 {
		return syscall.Errno(r1)
	}
	return nil
}

func freeMibTable(table unsafe.Pointer) {
	syscall.SyscallN(procFreeMibTable.Addr(), uintptr(table))
}
