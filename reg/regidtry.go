//go:build windows
// +build windows

package reg

import (
	"fmt"
	"strings"
	"testing"

	"golang.org/x/sys/windows/registry"
)

var rootKeyMap = map[string]registry.Key{
	"HKEY_CLASSES_ROOT":   registry.CLASSES_ROOT,
	"HKCR":                registry.CLASSES_ROOT,
	"HKEY_CURRENT_USER":   registry.CURRENT_USER,
	"HKCU":                registry.CURRENT_USER,
	"HKEY_LOCAL_MACHINE":  registry.LOCAL_MACHINE,
	"HKLM":                registry.LOCAL_MACHINE,
	"HKEY_USERS":          registry.USERS,
	"HKU":                 registry.USERS,
	"HKEY_CURRENT_CONFIG": registry.CURRENT_CONFIG,
	"HKCC":                registry.CURRENT_CONFIG,
}

// ParsePath 解析注册表路径，返回根键和子键
func parsePath(path string) (registry.Key, string, error) {
	// 分割路径为根键和子键
	parts := strings.SplitN(path, "\\", 2)
	if len(parts) < 2 {
		return 0, "", fmt.Errorf("invalid registry path: %s", path)
	}
	// 查找根键
	rootKeyStr := strings.ToUpper(parts[0])
	rootKey, exists := rootKeyMap[rootKeyStr]
	if !exists {
		return 0, "", fmt.Errorf("unknown root key: %s", parts[0])
	}

	return rootKey, parts[1], nil
}

func CheckPath(p string) bool {
	rootKey, patch, err := parsePath(p)
	if err != nil {
		return false
	}
	k, err := registry.OpenKey(rootKey, patch, registry.QUERY_VALUE)
	if err != nil {
		k.Close()
		return false
	}
	k.Close()
	return true
}

func GetSValue(p string, key string) (string, error) {
	rootKey, patch, err := parsePath(p)
	if err != nil {
		return "", err
	}
	k, err := registry.OpenKey(rootKey, patch, registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer k.Close()
	s, _, err := k.GetStringValue(key)
	if err != nil {
		return "", err
	}
	return s, nil
}

func TestRegPath(t *testing.T) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
	if err != nil {
		fmt.Print(err)
	}
	defer k.Close()
	s, _, err := k.GetStringValue("SystemRoot")
	if err != nil {
		fmt.Print(err)
	}
	fmt.Printf("Windows system root is %q\n", s)
}

// 添加注册表根据路径导出为reg 文件 使用Windows api
func ExportRegPathToFile(regPath string, filePath string) error {
	rootKey, patch, err := parsePath(regPath)
	if err != nil {
		return err
	}

	k, err := registry.OpenKey(rootKey, patch, registry.QUERY_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()

	// 使用 Windows API 导出注册表路径
	return nil
}
