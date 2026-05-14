//go:build windows

package obj

import (
	"fmt"
	"strings"
	"syscall"
	"unsafe"

	"github.com/kitsch-9527/wcorefx/os"
	"github.com/kitsch-9527/wcorefx/internal/winapi"
	"golang.org/x/sys/windows"
)

const maxPath = 260
const ptrSize = uint32(unsafe.Sizeof(uintptr(0)))

// DriverList returns all loaded device driver base addresses.
func DriverList() ([]uintptr, error) {
	var lpcNeeded uint32
	err := procEnumDeviceDrivers.Call(0, 0, uintptr(unsafe.Pointer(&lpcNeeded)))
	if err != nil && !winapi.IsErrInsufficientBuffer(err) {
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
	err = procEnumDeviceDrivers.Call(
		uintptr(unsafe.Pointer(&drivers[0])),
		uintptr(bufferSize),
		uintptr(unsafe.Pointer(&lpcNeeded)),
	)
	if err != nil {
		return nil, fmt.Errorf("EnumDeviceDrivers second call failed: %w", err)
	}

	actualCount := int(lpcNeeded / ptrSize)
	if actualCount < len(drivers) {
		drivers = drivers[:actualCount]
	}
	return drivers, nil
}

// DriverPath returns the file path for the given driver base address.
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

// DriverName returns the file name for the given driver base address.
func DriverName(driver uintptr) (string, error) {
	var lpName [maxPath + 1]uint16
	err := getDeviceDriverBaseName(driver, &lpName[0], uint32(len(lpName)))
	if err != nil {
		return "", fmt.Errorf("GetDeviceDriverBaseName failed: %w", err)
	}
	return windows.UTF16ToString(lpName[:]), nil
}

// KernelModuleInfo represents a loaded kernel module.
type KernelModuleInfo struct {
	Name      string
	Path      string
	ImageBase uint64
	ImageSize uint32
}

// KernelModules returns all loaded kernel modules.
func KernelModules() ([]KernelModuleInfo, error) {
	var returnLen uint32
	err := ntQuerySystemInformation(systemModuleInformation, nil, 0, &returnLen)
	if err != nil && err != windows.ERROR_INSUFFICIENT_BUFFER && err != syscall.Errno(0xC0000004) {
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

// ObjectEntry represents an entry in an NT object directory.
type ObjectEntry struct {
	Name     string
	TypeName string
}

// ObjectDirectory enumerates entries in the specified NT object directory.
func ObjectDirectory(objectName string) ([]ObjectEntry, error) {
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
	err = procNtOpenDirectoryObject.Call(
		uintptr(unsafe.Pointer(&handle)),
		uintptr(0x0003), // DIRECTORY_QUERY | DIRECTORY_TRAVERSE
		uintptr(unsafe.Pointer(&oa)),
	)
	if err != nil {
		return nil, fmt.Errorf("NtOpenDirectoryObject failed: %w", err)
	}
	defer windows.CloseHandle(handle)

	var entries []ObjectEntry
	var context uint32
	var buf [8192]byte

	for {
		var returnLen uint32
		restartScan := uint32(0)
		if context == 0 {
			restartScan = 1
		}
		err := ntQueryDirectoryObject(uintptr(handle), unsafe.Pointer(&buf[0]), uint32(len(buf)), &returnLen, &context, 1, restartScan)
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

// Devices enumerates the \Device\ directory.
func Devices() ([]ObjectEntry, error) {
	return ObjectDirectory(`\Device`)
}
