package eve

import (
	"syscall"
	"unsafe"

	win "golang.org/x/sys/windows"
)

var (
	modwevtapi = win.NewLazySystemDLL("wevtapi.dll")
)

var (
	evtExportLog = modwevtapi.NewProc("EvtExportLog")
)

// EvtExportLog exports the events that match the specified query from the specified
func EvtExportLog(session win.Handle, path *uint16, TrageFilePath *uint16, query *uint16, flags EVT_EXPORTLOG_FLAGS) (ret error) {
	r0, _, err := syscall.Syscall6(evtExportLog.Addr(), 6, uintptr(session), uintptr(unsafe.Pointer(path)), uintptr(unsafe.Pointer(TrageFilePath)), uintptr(unsafe.Pointer(query)), uintptr(uint32(flags)), 0)
	if r0 == 0 {
		ret = syscall.Errno(err)
	}
	return
}
