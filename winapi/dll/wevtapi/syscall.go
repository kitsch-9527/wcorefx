package wevtapi

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

func EvtExportLog(session windows.Handle, path *uint16, TrageFilePath *uint16, query *uint16, flags EVT_EXPORTLOG_FLAGS) (ret error) {
	r0, _, err := syscall.Syscall6(procEvtExportLog.Addr(), 6, uintptr(session), uintptr(unsafe.Pointer(path)), uintptr(unsafe.Pointer(TrageFilePath)), uintptr(unsafe.Pointer(query)), uintptr(uint32(flags)), 0)
	if r0 == 0 {
		ret = syscall.Errno(err)
	}
	return
}
