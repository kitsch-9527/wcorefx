//go:build windows

package obj

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/kitsch-9527/wcorefx/os"
	"golang.org/x/sys/windows"
)

const maxPath = 260
const ptrSize = uint32(unsafe.Sizeof(uintptr(0)))

// DriverList 返回所有已加载设备驱动程序的基地址。
//   返回1 - 已加载设备驱动程序的基地址列表
//   返回2 - 错误信息
func DriverList() ([]uintptr, error) {
	var lpcNeeded uint32
	err := enumDeviceDrivers(nil, 0, &lpcNeeded)
	if err != nil && err != windows.ERROR_INSUFFICIENT_BUFFER {
		return nil, fmt.Errorf("EnumDeviceDrivers first call failed: %w", err)
	}
	if lpcNeeded == 0 {
		return nil, fmt.Errorf("no drivers found")
	}
	if ptrSize == 0 {
		return nil, fmt.Errorf("invalid pointer size: %d", ptrSize)
	}

	driverCount := int(lpcNeeded / ptrSize)
	drivers := make([]uintptr, driverCount)

	bufferSize := uint32(len(drivers)) * ptrSize
	err = enumDeviceDrivers(&drivers[0], bufferSize, &lpcNeeded)
	if err != nil {
		return nil, fmt.Errorf("EnumDeviceDrivers second call failed: %w", err)
	}

	actualCount := int(lpcNeeded / ptrSize)
	if actualCount < len(drivers) {
		drivers = drivers[:actualCount]
	}
	return drivers, nil
}

// DriverPath 返回指定驱动程序基地址对应的文件路径。
//   driver - 驱动程序基地址
//   返回1 - 驱动程序文件路径
//   返回2 - 错误信息
func DriverPath(driver uintptr) (string, error) {
	var lpFilename [maxPath + 1]uint16
	err := getDeviceDriverFileName(driver, &lpFilename[0], uint32(len(lpFilename)))
	if err != nil {
		return "", fmt.Errorf("GetDeviceDriverFileName failed: %w", err)
	}
	path := windows.UTF16ToString(lpFilename[:])

	sysroot := `\SystemRoot`
	if strings.HasPrefix(path, sysroot) {
		winDir, err := os.WinDir()
		if err == nil {
			path = strings.Replace(path, sysroot, winDir, 1)
		}
	}
	path = strings.ReplaceAll(path, `\??\`, "")
	return path, nil
}

// DriverName 返回指定驱动程序基地址对应的文件名。
//   driver - 驱动程序基地址
//   返回1 - 驱动程序文件名
//   返回2 - 错误信息
func DriverName(driver uintptr) (string, error) {
	var lpName [maxPath + 1]uint16
	err := getDeviceDriverBaseName(driver, &lpName[0], uint32(len(lpName)))
	if err != nil {
		return "", fmt.Errorf("GetDeviceDriverBaseName failed: %w", err)
	}
	return windows.UTF16ToString(lpName[:]), nil
}
