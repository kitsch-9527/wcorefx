package wtsapi32

import "golang.org/x/sys/windows"

var modwtsapi32 = windows.NewLazySystemDLL("wtsapi32.dll")

var (
	procWTSQuerySessionInformationW = modwtsapi32.NewProc("WTSQuerySessionInformationW")
)
