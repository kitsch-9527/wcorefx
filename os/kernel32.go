//go:build windows

package os

import (
	"fmt"
	"unsafe"

	"github.com/kitsch-9527/wcorefx/internal/winapi"
)

var (
	procGetSystemInfo        = winapi.NewProc("kernel32.dll", "GetNativeSystemInfo")
	procGetTickCount64       = winapi.NewProc("kernel32.dll", "GetTickCount64")
	procGlobalMemoryStatusEx = winapi.NewProc("kernel32.dll", "GlobalMemoryStatusEx")
	procRtlGetNtVersionNumbers = winapi.NewProc("ntdll.dll", "RtlGetNtVersionNumbers")
)

type SYSTEM_INFO struct {
	wProcessorArchitecture      uint16
	wReserved                   uint16
	dwPageSize                  uint32
	lpMinimumApplicationAddress uintptr
	lpMaximumApplicationAddress uintptr
	dwActiveProcessorMask       uintptr
	dwNumberOfProcessors        uint32
	dwProcessorType             uint32
	dwAllocationGranularity     uint32
	wProcessorLevel             uint16
	wProcessorRevision          uint16
}

type MEMORYSTATUSEX struct {
	dwLength                uint32
	dwMemoryLoad            uint32
	ullTotalPhys            uint64
	ullAvailPhys            uint64
	ullTotalPageFile        uint64
	ullAvailPageFile        uint64
	ullTotalVirtual         uint64
	ullAvailVirtual         uint64
	ullAvailExtendedVirtual uint64
}

func getNativeSystemInfo() SYSTEM_INFO {
	var info SYSTEM_INFO
	procGetSystemInfo.Call(uintptr(unsafe.Pointer(&info)))
	return info
}

func getTickCount64() uint64 {
	ret, _ := procGetTickCount64.CallRet()
	return uint64(ret)
}

func globalMemoryStatusEx() (MEMORYSTATUSEX, error) {
	var ms MEMORYSTATUSEX
	ms.dwLength = uint32(unsafe.Sizeof(ms))
	err := procGlobalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&ms)))
	if err != nil {
		return ms, fmt.Errorf("GlobalMemoryStatusEx failed: %w", err)
	}
	return ms, nil
}
