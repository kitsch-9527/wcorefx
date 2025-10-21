package common

import (
	"unsafe"
)

const (
	MAXPATH = 260
	PtrSize = uint32(unsafe.Sizeof(uintptr(0)))
)
