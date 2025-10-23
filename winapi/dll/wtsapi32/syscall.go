package wtsapi32

import (
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// WTSQuerySessionInformation 封装系统调用，获取指定会话的信息并返回字符串
func WTSQuerySessionInformation(hServer windows.Handle, sessionId uint32, infoClass int) (string, error) {
	var buffer *uint16 // 接收UTF-16字符串的缓冲区
	var bytesReturned uint32
	// 调用系统API
	r1, _, e1 := syscall.Syscall6(
		procWTSQuerySessionInformationW.Addr(),
		5,
		uintptr(hServer),
		uintptr(sessionId),
		uintptr(infoClass),
		uintptr(unsafe.Pointer(&buffer)),
		uintptr(unsafe.Pointer(&bytesReturned)),
		0,
	)

	// 检查调用结果
	if r1 == 0 {
		return "", fmt.Errorf("WTSQuerySessionInformation failed: %w", windows.Errno(e1))
	}
	// 确保内存被释放
	defer windows.WTSFreeMemory(uintptr(unsafe.Pointer(buffer)))
	// 将UTF-16字符串转换为Go字符串
	return windows.UTF16PtrToString(buffer), nil
}
