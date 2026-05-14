//go:build windows

package ps

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/kitsch-9527/wcorefx/internal/winapi"
)

const (
	errStatusBufferTooSmall       = syscall.Errno(0xC0000023)
	errStatusInfoLengthMismatch   = syscall.Errno(0xC0000004)
)

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
	_ [22]byte
}

type unicodeString struct {
	Length    uint16
	MaxLength uint16
	Buffer    *uint16
}

var (
	procNtQueryObject     = winapi.NewProc("ntdll.dll", "NtQueryObject", winapi.ConvNTSTATUS)
	procNtDuplicateObject = winapi.NewProc("ntdll.dll", "NtDuplicateObject", winapi.ConvNTSTATUS)
)

func ntQueryObject(handle windows.Handle, infoClass objectInformationClass, buf []byte) ([]byte, error) {
	var returnLength uint32
	bufSize := uint32(len(buf))
	if bufSize == 0 {
		bufSize = 256
		buf = make([]byte, bufSize)
	}
	_, err := procNtQueryObject.CallRet(
		uintptr(handle),
		uintptr(infoClass),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(bufSize),
		uintptr(unsafe.Pointer(&returnLength)),
	)
	if err != nil && err != errStatusBufferTooSmall && err != errStatusInfoLengthMismatch {
		return nil, err
	}
	if err != nil {
		buf = make([]byte, returnLength)
		_, err = procNtQueryObject.CallRet(
			uintptr(handle),
			uintptr(infoClass),
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(returnLength),
			uintptr(unsafe.Pointer(&returnLength)),
		)
		if err != nil {
			return nil, err
		}
	}
	return buf, nil
}

func ntDuplicateObject(srcProc windows.Handle, srcHandle windows.Handle, dstProc windows.Handle, options uint32) (windows.Handle, error) {
	var targetHandle windows.Handle
	_, err := procNtDuplicateObject.CallRet(
		uintptr(srcProc),
		uintptr(srcHandle),
		uintptr(dstProc),
		uintptr(unsafe.Pointer(&targetHandle)),
		0,
		0,
		uintptr(options),
	)
	if err != nil {
		return 0, err
	}
	return targetHandle, nil
}
