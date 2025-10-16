package os

import (
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	WTS_CURRENT_SERVER_HANDLE = 0
	WTS_CURRENT_SESSION       = 0xFFFFFFFF
	WTSUSERNAME               = 5
)

var (
	modwtsapi32 = syscall.NewLazyDLL("wtsapi32.dll")
	modkernel32 = syscall.NewLazyDLL("kernel32.dll")
)
var (
	procWTSQuerySessionInformationW = modwtsapi32.NewProc("WTSQuerySessionInformationW")
	procWTSFreeMemory               = modwtsapi32.NewProc("WTSFreeMemory")
	procProcessIdToSessionId        = modkernel32.NewProc("ProcessIdToSessionId")
	procGetNativeSystemInfo         = modkernel32.NewProc("GetNativeSystemInfo")
)

// WTSQuerySessionInformation 封装系统调用，获取指定会话的信息并返回字符串
func procWTSQuerySessionInformation(hServer windows.Handle, sessionId uint32, infoClass int) (string, error) {
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

type systemInfo struct {
	WProcessorArchitecture      uint16
	WReserved                   uint16
	DwpageSize                  uint32
	LpMinimumApplicationAddress uintptr
	LpMaximumApplicationAddress uintptr
	DwActiveProcessorMask       uintptr
	DwNumberOfProcessors        uint32
	DwProcessorType             uint32
	DwAllocationGranularity     uint32
	WProcessorLevel             uint16
	WProcessorRevision          uint16
}

func GetSystemInfo() systemInfo {
	var si systemInfo
	syscall.Syscall(
		procGetNativeSystemInfo.Addr(),
		1,
		uintptr(unsafe.Pointer(&si)),
		0,
		0,
	)
	return si
}
