//go:build windows

package reg

import (
	"fmt"
	"strings"

	"golang.org/x/sys/windows/registry"
)

// ValueInfo 表示注册表值的信息
type ValueInfo struct {
	Name  string // 值名称
	Type  string // 值类型（如 REG_SZ, REG_DWORD, REG_EXPAND_SZ 等）
	Value string // 值的字符串表示
}

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

// parsePath 解析注册表路径，返回根键和子键
func parsePath(path string) (registry.Key, string, error) {
	parts := strings.SplitN(path, "\\", 2)
	if len(parts) < 2 {
		return 0, "", fmt.Errorf("invalid registry path: %s", path)
	}
	rootKeyStr := strings.ToUpper(parts[0])
	rootKey, exists := rootKeyMap[rootKeyStr]
	if !exists {
		return 0, "", fmt.Errorf("unknown root key: %s", parts[0])
	}
	return rootKey, parts[1], nil
}

// CheckPath 检查注册表路径是否存在
//   p - 完整的注册表路径（如HKLM\Software\Microsoft）
//   返回 - 路径存在返回true，否则返回false
func CheckPath(p string) bool {
	rootKey, subPath, err := parsePath(p)
	if err != nil {
		return false
	}
	k, err := registry.OpenKey(rootKey, subPath, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	k.Close()
	return true
}

// GetValue 获取注册表字符串值
//   p - 完整的注册表路径
//   key - 注册表项名称
//   返回 - 字符串值
//   返回 - 错误信息
func GetValue(p string, key string) (string, error) {
	rootKey, subPath, err := parsePath(p)
	if err != nil {
		return "", err
	}
	k, err := registry.OpenKey(rootKey, subPath, registry.QUERY_VALUE)
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

// EnumValues 枚举指定注册表路径下的所有值
//   p - 完整的注册表路径（如 HKLM\Software\Microsoft\Windows\CurrentVersion\Run）
//   返回 - 值信息列表
//   返回 - 错误信息
func EnumValues(p string) ([]ValueInfo, error) {
	rootKey, subPath, err := parsePath(p)
	if err != nil {
		return nil, err
	}
	k, err := registry.OpenKey(rootKey, subPath, registry.QUERY_VALUE)
	if err != nil {
		return nil, fmt.Errorf("open key failed: %w", err)
	}
	defer k.Close()

	names, err := k.ReadValueNames(0)
	if err != nil {
		return nil, fmt.Errorf("read value names failed: %w", err)
	}

	values := make([]ValueInfo, 0, len(names))
	for _, name := range names {
		vi := ValueInfo{Name: name}

		bufSize, valType, err := k.GetValue(name, nil)
		if err != nil {
			continue
		}
		vi.Type = typeName(valType)

		switch valType {
		case registry.SZ, registry.EXPAND_SZ:
			s, _, err := k.GetStringValue(name)
			if err == nil {
				vi.Value = s
			}
		case registry.MULTI_SZ:
			strs, _, err := k.GetStringsValue(name)
			if err == nil {
				vi.Value = strings.Join(strs, ", ")
			}
		case registry.DWORD, registry.QWORD:
			n, _, err := k.GetIntegerValue(name)
			if err == nil {
				vi.Value = fmt.Sprintf("%d (0x%X)", n, n)
			}
		case registry.BINARY:
			buf := make([]byte, bufSize)
			n, _, err := k.GetValue(name, buf)
			if err == nil {
				vi.Value = fmt.Sprintf("%X", buf[:n])
			}
		default:
			buf := make([]byte, bufSize)
			n, _, err := k.GetValue(name, buf)
			if err == nil {
				vi.Value = string(buf[:n])
			}
		}
		values = append(values, vi)
	}
	return values, nil
}

// EnumSubKeys 枚举指定注册表路径下的所有子键
//   p - 完整的注册表路径（如 HKLM\Software\Microsoft\Windows\CurrentVersion\Run）
//   返回 - 子键名称列表
//   返回 - 错误信息
func EnumSubKeys(p string) ([]string, error) {
	rootKey, subPath, err := parsePath(p)
	if err != nil {
		return nil, err
	}
	k, err := registry.OpenKey(rootKey, subPath, registry.ENUMERATE_SUB_KEYS)
	if err != nil {
		return nil, fmt.Errorf("open key failed: %w", err)
	}
	defer k.Close()

	names, err := k.ReadSubKeyNames(0)
	if err != nil {
		return nil, fmt.Errorf("read subkey names failed: %w", err)
	}
	return names, nil
}

// SetString 设置注册表字符串值 (REG_SZ)。
//   path  - 完整的注册表路径（如 HKLM\Software\MyKey）
//   key   - 值名称
//   value - 要设置的字符串值
//   返回 - 错误信息
func SetString(path, key, value string) error {
	rootKey, subPath, err := parsePath(path)
	if err != nil {
		return err
	}
	k, err := registry.OpenKey(rootKey, subPath, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("open key failed: %w", err)
	}
	defer k.Close()
	return k.SetStringValue(key, value)
}

// SetDWORD 设置注册表 DWORD 值 (REG_DWORD)。
//   path  - 完整的注册表路径
//   key   - 值名称
//   value - 要设置的 DWORD 值
//   返回 - 错误信息
func SetDWORD(path, key string, value uint32) error {
	rootKey, subPath, err := parsePath(path)
	if err != nil {
		return err
	}
	k, err := registry.OpenKey(rootKey, subPath, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("open key failed: %w", err)
	}
	defer k.Close()
	return k.SetDWordValue(key, value)
}

// CreateKey 创建注册表键。
//   path - 完整的注册表路径
//   返回 - 错误信息
func CreateKey(path string) error {
	rootKey, subPath, err := parsePath(path)
	if err != nil {
		return err
	}
	k, _, err := registry.CreateKey(rootKey, subPath, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("create key failed: %w", err)
	}
	k.Close()
	return nil
}

// DeleteKey 删除注册表键及其所有子键。
//   path - 完整的注册表路径
//   返回 - 错误信息
func DeleteKey(path string) error {
	rootKey, subPath, err := parsePath(path)
	if err != nil {
		return err
	}
	err = registry.DeleteKey(rootKey, subPath)
	if err != nil {
		return fmt.Errorf("delete key failed: %w", err)
	}
	return nil
}

// DeleteValue 删除注册表值。
//   path - 完整的注册表路径
//   key  - 要删除的值名称
//   返回 - 错误信息
func DeleteValue(path, key string) error {
	rootKey, subPath, err := parsePath(path)
	if err != nil {
		return err
	}
	k, err := registry.OpenKey(rootKey, subPath, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("open key failed: %w", err)
	}
	defer k.Close()
	return k.DeleteValue(key)
}

// typeName 将注册表值类型转换为字符串表示
func typeName(t uint32) string {
	switch t {
	case registry.NONE:
		return "REG_NONE"
	case registry.SZ:
		return "REG_SZ"
	case registry.EXPAND_SZ:
		return "REG_EXPAND_SZ"
	case registry.BINARY:
		return "REG_BINARY"
	case registry.DWORD:
		return "REG_DWORD"
	case registry.DWORD_BIG_ENDIAN:
		return "REG_DWORD_BIG_ENDIAN"
	case registry.LINK:
		return "REG_LINK"
	case registry.MULTI_SZ:
		return "REG_MULTI_SZ"
	case registry.QWORD:
		return "REG_QWORD"
	default:
		return fmt.Sprintf("REG_0x%X", t)
	}
}
