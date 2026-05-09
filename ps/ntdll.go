//go:build windows

package ps

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// OBJECT_INFORMATION_CLASS for NtQueryObject
type objectInformationClass int32

const (
	objectBasicInformation objectInformationClass = 0
	objectNameInformation  objectInformationClass = 1
	objectTypeInformation  objectInformationClass = 2
)

type publicObjectTypeInformation struct {
	TypeName struct {
		Length    uint16
		MaxLength uint16
		Buffer    *uint16
	}
	_ [22]byte // remaining fields not needed
}

type unicodeString struct {
	Length    uint16
	MaxLength uint16
	Buffer    *uint16
}

var (
	modntdll = windows.NewLazySystemDLL("ntdll.dll")

	procNtQueryObject      = modntdll.NewProc("NtQueryObject")
	procNtDuplicateObject  = modntdll.NewProc("NtDuplicateObject")
)

func ntQueryObject(handle windows.Handle, infoClass objectInformationClass, buf []byte) ([]byte, error) {
	var returnLength uint32
	bufSize := uint32(len(buf))
	if bufSize == 0 {
		bufSize = 256
		buf = make([]byte, bufSize)
	}
	status, _, _ := procNtQueryObject.Call(
		uintptr(handle),
		uintptr(infoClass),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(bufSize),
		uintptr(unsafe.Pointer(&returnLength)),
	)
	if status == 0xC0000004 || status == 0xC0000023 { // STATUS_INFO_LENGTH_MISMATCH / STATUS_BUFFER_TOO_SMALL
		buf = make([]byte, returnLength)
		status, _, _ = procNtQueryObject.Call(
			uintptr(handle),
			uintptr(infoClass),
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(returnLength),
			uintptr(unsafe.Pointer(&returnLength)),
		)
	}
	if status != 0 {
		return nil, syscall.Errno(status)
	}
	return buf, nil
}

func ntDuplicateObject(srcProc windows.Handle, srcHandle windows.Handle, dstProc windows.Handle, options uint32) (windows.Handle, error) {
	var targetHandle windows.Handle
	status, _, _ := procNtDuplicateObject.Call(
		uintptr(srcProc),
		uintptr(srcHandle),
		uintptr(dstProc),
		uintptr(unsafe.Pointer(&targetHandle)),
		0, // PROCESS_DUP_HANDLE
		0,
		uintptr(options),
	)
	if status != 0 {
		return 0, syscall.Errno(status)
	}
	return targetHandle, nil
}
