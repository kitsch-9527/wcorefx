//go:build windows

package obj

import (
	"fmt"
	"strings"

	"github.com/kitsch-9527/wcorefx/os"
	"golang.org/x/sys/windows"
)

// NativePathToDosPath 将 NT 设备本地路径转换为 DOS 驱动器路径（如 \SystemRoot\System32\ntdll.dll 转换为 C:\Windows\System32\ntdll.dll）。
//   nativePath - NT 设备本地路径
//   返回1 - 转换后的 DOS 驱动器路径
//   返回2 - 错误信息
func NativePathToDosPath(nativePath string) (string, error) {
	const maxPath = 260

	// Expand \SystemRoot symbolic link to the actual Windows directory.
	sysroot := `\SystemRoot`
	if strings.HasPrefix(nativePath, sysroot) {
		winDir, err := os.WinDir()
		if err == nil {
			nativePath = strings.Replace(nativePath, sysroot, winDir, 1)
		}
	}

	// If the path is already a DOS path (drive letter), return it directly.
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
