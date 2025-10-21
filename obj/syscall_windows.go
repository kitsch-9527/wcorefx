//go:build windows
// +build windows

package ob

import (
	"syscall"
	"unsafe"

	"github.com/kitsch-9527/wcorefx/comm"
	w "golang.org/x/sys/windows"
)

type ImageBase uintptr

var (
	//modshlwapi = w.NewLazySystemDLL("shlwapi.dll")
	modpsapi = w.NewLazySystemDLL("psapi.dll")
)

var (
	enumDeviceDrivers        = modpsapi.NewProc("EnumDeviceDrivers")
	getDeviceDriverBaseNameW = modpsapi.NewProc("GetDeviceDriverBaseNameW")
	getDeviceDriverFileNameW = modpsapi.NewProc("GetDeviceDriverFileNameW")
)

// EnumDeviceDrivers 枚举系统中所有驱动程序
func EnumDeviceDrivers(drivers *ImageBase, cb uint32, lpcNeeded *uint32) (err error) {
	r1, _, e1 := syscall.Syscall(enumDeviceDrivers.Addr(), 3, uintptr(unsafe.Pointer(drivers)), uintptr(cb), uintptr(unsafe.Pointer(lpcNeeded)))
	if r1 == 0 {
		err = comm.ErrnoErr(e1)
	}
	return
}

// GetDeviceDriverBaseName 获得指定驱动程序的名称
func GetDeviceDriverBaseName(driver ImageBase, lpBaseName *uint16, nSize uint32) (err error) {
	r1, _, e1 := syscall.Syscall(getDeviceDriverBaseNameW.Addr(), 3, uintptr(driver), uintptr(unsafe.Pointer(lpBaseName)), uintptr(nSize))
	if r1 == 0 {
		err = comm.ErrnoErr(e1)
	}
	return
}

// GetDeviceDriverFileName 获得指定驱动程序的路径
func GetDeviceDriverFileName(driver ImageBase, lpFilename *uint16, nSize uint32) (err error) {
	r1, _, e1 := syscall.Syscall(getDeviceDriverFileNameW.Addr(), 3, uintptr(driver), uintptr(unsafe.Pointer(lpFilename)), uintptr(nSize))
	if r1 == 0 {
		err = comm.ErrnoErr(e1)
	}
	return
}
