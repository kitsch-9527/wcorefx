package common

import (
	"unsafe"
)

const (
	MAXPATH = 260                               // stdlib.h 中定义的最大路径长度
	PtrSize = uint32(unsafe.Sizeof(uintptr(0))) // 指针大小（字长），用于处理不同架构的指针运算
)
