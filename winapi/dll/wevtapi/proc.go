package wevtapi

import "golang.org/x/sys/windows"

var modwevtapi = windows.NewLazySystemDLL("wevtapi.dll")

var (
	procEvtExportLog = modwevtapi.NewProc("EvtExportLog")
)
