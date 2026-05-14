//go:build windows

package os

import (
	"fmt"
	"strings"

	"golang.org/x/sys/windows"
)

// NativePathToDosPath 将 NT 设备本地路径转换为 DOS 驱动器路径。
func NativePathToDosPath(nativePath string) (string, error) {
	const maxPath = 260

	sysroot := `\SystemRoot`
	if strings.HasPrefix(nativePath, sysroot) {
		winDir, err := WinDir()
		if err == nil {
			nativePath = strings.Replace(nativePath, sysroot, winDir, 1)
		}
	}

	if len(nativePath) >= 2 && nativePath[1] == ':' {
		return nativePath, nil
	}

	for c := 'A'; c <= 'Z'; c++ {
		dosDevice := fmt.Sprintf("%c:", c)
		var target [maxPath + 1]uint16

		n, err := windows.QueryDosDevice(
			windows.StringToUTF16Ptr(dosDevice),
			&target[0],
			maxPath,
		)
		if err != nil || n == 0 {
			continue
		}
		devicePath := windows.UTF16ToString(target[:n])
		if len(devicePath) > 0 &&
			len(nativePath) > len(devicePath) &&
			nativePath[:len(devicePath)] == devicePath {
			return fmt.Sprintf("%s%s", dosDevice, nativePath[len(devicePath):]), nil
		}
	}
	return "", fmt.Errorf("no matching DOS device for native path: %s", nativePath)
}
