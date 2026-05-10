//go:build windows

package obj

import (
	"fmt"
	"strings"
	"syscall"
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

// KernelModuleInfo 表示内核模块信息
type KernelModuleInfo struct {
	// Name 模块文件名（如 ntoskrnl.exe）
	Name string
	// Path 模块完整路径
	Path string
	// ImageBase 模块基址
	ImageBase uint64
	// ImageSize 模块镜像大小
	ImageSize uint32
}

// KernelModules 返回所有已加载的内核模块
//   返回 - 内核模块信息列表
//   返回 - 错误信息
func KernelModules() ([]KernelModuleInfo, error) {
	var returnLen uint32
	err := ntQuerySystemInformation(systemModuleInformation, nil, 0, &returnLen)
	if err != nil && err != windows.ERROR_INSUFFICIENT_BUFFER && err != syscall.Errno(0xC0000004) {
		// STATUS_INFO_LENGTH_MISMATCH is expected when querying buffer size
		_ = err
	}

	buf := make([]byte, returnLen)
	err = ntQuerySystemInformation(systemModuleInformation, unsafe.Pointer(&buf[0]), returnLen, &returnLen)
	if err != nil {
		return nil, fmt.Errorf("NtQuerySystemInformation failed: %w", err)
	}

	info := (*systemModuleInfo)(unsafe.Pointer(&buf[0]))
	count := info.ModulesCount
	entries := unsafe.Slice(&info.Modules[0], count)

	modules := make([]KernelModuleInfo, 0, count)
	for _, e := range entries {
		path := windows.UTF16ToString((*[128]uint16)(unsafe.Pointer(&e.FullPathName[0]))[:])

		name := path
		if lastSlash := strings.LastIndex(path, `\`); lastSlash >= 0 {
			name = path[lastSlash+1:]
		}

		modules = append(modules, KernelModuleInfo{
			Name:      name,
			Path:      path,
			ImageBase: e.ImageBase,
			ImageSize: e.ImageSize,
		})
	}
	return modules, nil
}

// ObjectEntry 表示对象目录中的一个条目
type ObjectEntry struct {
	// Name 对象名称
	Name string
	// TypeName 对象类型名（如 Device, File, Key 等）
	TypeName string
}

// ObjectDirectory 枚举指定对象目录下的条目
//   objectName - 对象目录路径（如 \GLOBAL??, \Device, \ObjectTypes）
//   返回 - 对象条目列表
//   返回 - 错误信息
func ObjectDirectory(objectName string) ([]ObjectEntry, error) {
	// 1. Open the object directory
	oa := objectAttributes{}
	nameUTF16, err := syscall.UTF16PtrFromString(objectName)
	if err != nil {
		return nil, fmt.Errorf("invalid object directory name: %w", err)
	}
	oa.Length = uint32(unsafe.Sizeof(oa))
	oa.ObjectName = &unicodeString{
		Buffer:        nameUTF16,
		Length:        uint16(len(objectName) * 2),
		MaximumLength: uint16((len(objectName) + 1) * 2),
	}
	oa.Attributes = 0x40 // OBJ_CASE_INSENSITIVE

	var handle windows.Handle
	r1, _, _ := syscall.SyscallN(modntdll.NewProc("NtOpenDirectoryObject").Addr(),
		uintptr(unsafe.Pointer(&handle)),
		uintptr(0x0003), // DIRECTORY_QUERY | DIRECTORY_TRAVERSE
		uintptr(unsafe.Pointer(&oa)),
	)
	if r1 != 0 {
		return nil, fmt.Errorf("NtOpenDirectoryObject failed: %w", syscall.Errno(r1))
	}
	defer windows.CloseHandle(handle)

	// 2. Query directory entries
	var entries []ObjectEntry
	var context uint32
	var buf [8192]byte

	for {
		var returnLen uint32
		restartScan := uint32(0)
		if context == 0 {
			restartScan = 1
		}
		err := ntQueryDirectoryObject(handle, unsafe.Pointer(&buf[0]), uint32(len(buf)), &returnLen, &context, 1, restartScan)
		if err != nil {
			break
		}
		if returnLen == 0 {
			break
		}

		de := (*objectDirectoryInformation)(unsafe.Pointer(&buf[0]))
		name := ""
		typeName := ""
		if de.Name.Buffer != nil {
			name = windows.UTF16ToString(unsafe.Slice(de.Name.Buffer, de.Name.Length/2))
		}
		if de.TypeName.Buffer != nil {
			typeName = windows.UTF16ToString(unsafe.Slice(de.TypeName.Buffer, de.TypeName.Length/2))
		}
		entries = append(entries, ObjectEntry{Name: name, TypeName: typeName})
	}

	return entries, nil
}

// Devices 枚举 \Device\ 目录下的所有设备对象
//   返回 - 设备对象条目列表
//   返回 - 错误信息
func Devices() ([]ObjectEntry, error) {
	return ObjectDirectory(`\Device`)
}
