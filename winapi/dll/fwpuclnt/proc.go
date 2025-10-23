package fwpuclnt

import "golang.org/x/sys/windows"

var modfwpmu = windows.NewLazySystemDLL("fwpuclnt.dll")

var (
	procFwpmEngineOpen               = modfwpmu.NewProc("FwpmEngineOpen0")
	procFwpmCalloutCreateEnumHandle  = modfwpmu.NewProc("FwpmCalloutCreateEnumHandle0")
	procFwpmCalloutEnum              = modfwpmu.NewProc("FwpmCalloutEnum0")
	procFwpmCalloutDestroyEnumHandle = modfwpmu.NewProc("FwpmCalloutDestroyEnumHandle0")
	procFwpmEngineClose              = modfwpmu.NewProc("FwpmEngineClose0")
	procFwpmFreeMemory               = modfwpmu.NewProc("FwpmFreeMemory0")

	procFwpmFilterEnum              = modfwpmu.NewProc("FwpmFilterEnum0")
	procFwpmFilterCreateEnumHandle  = modfwpmu.NewProc("FwpmFilterCreateEnumHandle0")
	procFwpmFilterGetByKey          = modfwpmu.NewProc("FwpmFilterGetByKey0")
	procFwpmFilterDestroyEnumHandle = modfwpmu.NewProc("FwpmFilterDestroyEnumHandle0")
	procFwpmCalloutGetByKey         = modfwpmu.NewProc("FwpmCalloutGetByKey0")

	procFwpmFilterGetById  = modfwpmu.NewProc("FwpmFilterGetById0")
	procFwpmCalloutGetById = modfwpmu.NewProc("FwpmCalloutGetById0")
)
