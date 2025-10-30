//go:build windows
// +build windows

package obj

import (
	"fmt"
	"strings"
	"syscall"

	comm "github.com/kitsch-9527/wcorefx/common"
	"github.com/kitsch-9527/wcorefx/os"
	"github.com/kitsch-9527/wcorefx/winapi/dll/psapi"
	"golang.org/x/sys/windows"
)

type ImageBase = uintptr

// GetDriverList 获得系统中所有驱动程序
func GetDriverList() ([]ImageBase, error) {
	var lpcNeeded uint32
	// 第一次调用获取所需缓冲区大小
	err := psapi.EnumDeviceDrivers(nil, 0, &lpcNeeded)
	if err != nil && err != syscall.ERROR_INSUFFICIENT_BUFFER {
		return nil, fmt.Errorf("EnumDeviceDrivers first call failed: %w", err)
	}
	// 检查是否需要分配缓冲区
	if lpcNeeded == 0 {
		return nil, fmt.Errorf("no drivers found")
	}
	// 验证指针大小的有效性
	if comm.PtrSize == 0 {
		return nil, fmt.Errorf("invalid pointer size: %d", comm.PtrSize)
	}
	// 计算驱动程序数量并分配缓冲区
	driverCount := int(lpcNeeded / uint32(comm.PtrSize))
	drivers := make([]ImageBase, driverCount)

	// 第二次调用获取实际的驱动程序列表
	bufferSize := uint32(len(drivers)) * uint32(comm.PtrSize)
	err = psapi.EnumDeviceDrivers(&drivers[0], bufferSize, &lpcNeeded)
	if err != nil {
		return nil, fmt.Errorf("EnumDeviceDrivers second call failed: %w", err)
	}

	// 可能返回的驱动数量比请求的少，进行裁剪
	actualCount := int(lpcNeeded / uint32(comm.PtrSize))
	if actualCount < len(drivers) {
		drivers = drivers[:actualCount]
	}
	return drivers, nil
}

// GetDriverPath 获得指定驱动程序的路径
func GetDriverPath(driver ImageBase) (string, error) {
	var lpFilename [comm.MAXPATH + 1]uint16
	err := psapi.GetDeviceDriverFileName(driver, &lpFilename[0], uint32(len(lpFilename)))
	if err != nil {
		return "", fmt.Errorf("GetDeviceDriverFileName failed: %w", err)
	}
	path := windows.UTF16ToString(lpFilename[:])
	// 替换系统根目录
	sysroot := `\SystemRoot`
	if strings.HasPrefix(path, sysroot) {
		// 获取系统Windows目录替换SystemRoot
		winDir, err := os.WinDir()
		if err == nil {
			path = strings.Replace(path, sysroot, winDir, 1)
		}
	}
	// 移除路径中的"\??\"前缀
	path = strings.ReplaceAll(path, `\??\`, "")
	return path, nil
}

func GetDriverName(driver ImageBase) (string, error) {
	var lpName [comm.MAXPATH + 1]uint16
	err := psapi.GetDeviceDriverBaseName(driver, &lpName[0], uint32(len(lpName)))
	if err != nil {
		return "", fmt.Errorf("GetDeviceDriverBaseName failed: %w", err)
	}
	return windows.UTF16ToString(lpName[:]), nil
}
