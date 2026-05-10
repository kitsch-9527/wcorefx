//go:build windows

package obj

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var modntdll = windows.NewLazySystemDLL("ntdll.dll")

var (
	procNtQuerySystemInformation = modntdll.NewProc("NtQuerySystemInformation")
	procNtQueryDirectoryObject   = modntdll.NewProc("NtQueryDirectoryObject")
)

const (
	systemModuleInformation = 11
)

// systemModuleInfoEntry represents a single kernel module entry returned
// by NtQuerySystemInformation with SystemModuleInformation class.
// The FullPathName field contains the full path as UTF-16LE characters.
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

// unicodeString matches the native NT UNICODE_STRING structure.
type unicodeString struct {
	Length        uint16
	MaximumLength uint16
	Buffer        *uint16
}

// objectAttributes matches the native NT OBJECT_ATTRIBUTES structure.
type objectAttributes struct {
	Length                   uint32
	RootDirectory            windows.Handle
	ObjectName               *unicodeString
	Attributes               uint32
	SecurityDescriptor       unsafe.Pointer
	SecurityQualityOfService unsafe.Pointer
}

// objectDirectoryInformation represents a single OBJECT_DIRECTORY_INFORMATION
// returned by NtQueryDirectoryObject. Name and TypeName are inline
// UNICODE_STRING structs whose Buffer pointers reference string data
// within the same output buffer.
type objectDirectoryInformation struct {
	Name     unicodeString
	TypeName unicodeString
}

func ntQuerySystemInformation(infoClass uint32, buf unsafe.Pointer, bufSize uint32, returnLen *uint32) error {
	r1, _, _ := syscall.SyscallN(procNtQuerySystemInformation.Addr(),
		uintptr(infoClass),
		uintptr(buf),
		uintptr(bufSize),
		uintptr(unsafe.Pointer(returnLen)),
	)
	if r1 != 0 {
		return syscall.Errno(r1)
	}
	return nil
}

func ntQueryDirectoryObject(handle windows.Handle, buf unsafe.Pointer, bufSize uint32, returnLen *uint32, context *uint32, returnSingleEntry, restartScan uint32) error {
	r1, _, _ := syscall.SyscallN(procNtQueryDirectoryObject.Addr(),
		uintptr(handle),
		uintptr(buf),
		uintptr(bufSize),
		uintptr(returnSingleEntry),
		uintptr(restartScan),
		uintptr(unsafe.Pointer(context)),
		uintptr(unsafe.Pointer(returnLen)),
	)
	// STATUS_MORE_ENTRIES (0x105) means the buffer was filled and more
	// entries exist -- it is not an error condition.
	if r1 != 0 && r1 != 0x105 {
		return syscall.Errno(r1)
	}
	return nil
}
