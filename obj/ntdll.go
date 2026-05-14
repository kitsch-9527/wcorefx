//go:build windows

package obj

import (
	"unsafe"

	"github.com/kitsch-9527/wcorefx/internal/winapi"
)

var (
	procNtQuerySystemInformation = winapi.NewProc("ntdll.dll", "NtQuerySystemInformation", winapi.ConvNTSTATUS)
	procNtQueryDirectoryObject   = winapi.NewProc("ntdll.dll", "NtQueryDirectoryObject", winapi.ConvNTSTATUS)
	procNtOpenDirectoryObject   = winapi.NewProc("ntdll.dll", "NtOpenDirectoryObject", winapi.ConvNTSTATUS)
)

const (
	systemModuleInformation = 11
)

type systemModuleInfoEntry struct {
	Reserved1       [2]uint64
	ImageBase      uint64
	ImageSize      uint32
	Flags          uint32
	LoadCount      uint16
	LoadOrderIndex uint16
	InitOrderIndex uint16
	InitOrderOffset uint16
	FullPathName   [256]byte
}

type systemModuleInfo struct {
	ModulesCount uint32
	_            [4]byte
	Modules      [1]systemModuleInfoEntry
}

type unicodeString struct {
	Length        uint16
	MaximumLength uint16
	Buffer        *uint16
}

type objectAttributes struct {
	Length                   uint32
	RootDirectory            uintptr
	ObjectName               *unicodeString
	Attributes               uint32
	SecurityDescriptor       unsafe.Pointer
	SecurityQualityOfService unsafe.Pointer
}

type objectDirectoryInformation struct {
	Name     unicodeString
	TypeName unicodeString
}

func ntQuerySystemInformation(infoClass uint32, buf unsafe.Pointer, bufSize uint32, returnLen *uint32) error {
	return procNtQuerySystemInformation.Call(
		uintptr(infoClass),
		uintptr(buf),
		uintptr(bufSize),
		uintptr(unsafe.Pointer(returnLen)),
	)
}

func ntQueryDirectoryObject(handle uintptr, buf unsafe.Pointer, bufSize uint32, retLength *uint32, context *uint32, returnSingleEntry, restartScan uint32) error {
	r1, err := procNtQueryDirectoryObject.CallRet(
		handle,
		uintptr(buf),
		uintptr(bufSize),
		uintptr(returnSingleEntry),
		uintptr(restartScan),
		uintptr(unsafe.Pointer(context)),
		uintptr(unsafe.Pointer(retLength)),
	)
	// STATUS_MORE_ENTRIES (0x105) means buffer filled and more exist — not an error.
	if err != nil && r1 != 0x105 {
		return err
	}
	return nil
}
