//go:build windows

// Package task provides Windows Scheduled Task enumeration.
package task

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf16"

	"golang.org/x/sys/windows"
)

// TaskInfo 表示计划任务信息
type TaskInfo struct {
	// Name 任务名称（含路径，如 \Microsoft\Windows\ExampleTask）
	Name string
	// Path 任务文件完整路径
	Path string
	// Enabled 是否启用
	Enabled bool
	// Command 执行的命令
	Command string
	// Arguments 命令行参数
	Arguments string
	// Author 创建者
	Author string
	// UserId 运行账户
	UserId string
	// ComHandler 是否为COM Handler
	ComHandler bool
	// Clsid COM Handler 的 CLSID（如适用）
	Clsid string
}

// taskXML 无命名空间的XML结构（用于解析）
type taskXML struct {
	RegistrationInfo *struct {
		Author string `xml:"Author"`
	} `xml:"RegistrationInfo"`
	Settings *struct {
		Enabled string `xml:"Enabled"`
	} `xml:"Settings"`
	Actions *struct {
		Exec *struct {
			Command   string `xml:"Command"`
			Arguments string `xml:"Arguments"`
		} `xml:"Exec"`
		ComHandler *struct {
			Clsid string `xml:"Clsid"`
		} `xml:"ComHandler"`
	} `xml:"Actions"`
	Principals *struct {
		Principal []struct {
			UserId string `xml:"UserId"`
		} `xml:"Principal"`
	} `xml:"Principals"`
}

// taskDir 返回系统任务目录
func taskDir() (string, error) {
	winDir, err := windows.GetWindowsDirectory()
	if err != nil {
		return "", fmt.Errorf("GetWindowsDirectory failed: %w", err)
	}
	return winDir + "\\System32\\Tasks", nil
}

// List 枚举所有计划任务
//   返回 - 计划任务信息列表
//   返回 - 错误信息
func List() ([]TaskInfo, error) {
	dir, err := taskDir()
	if err != nil {
		return nil, err
	}
	return ListFrom(dir)
}

// ListFrom 枚举指定目录下的所有计划任务
//   dir - 任务文件目录路径
//   返回 - 计划任务信息列表
//   返回 - 错误信息
func ListFrom(dir string) ([]TaskInfo, error) {
	fi, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("stat task dir failed: %w", err)
	}
	if !fi.IsDir() {
		return nil, fmt.Errorf("not a directory: %s", dir)
	}

	var tasks []TaskInfo
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip inaccessible entries
		}
		if info.IsDir() {
			return nil
		}
		task, err := parseTaskXML(path)
		if err != nil {
			return nil // skip unparseable files
		}
		task.Path = path
		task.Name = normalizeName(dir, path)
		tasks = append(tasks, *task)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk task dir failed: %w", err)
	}
	return tasks, nil
}

// normalizeName 将文件路径转换为任务名称（如 \Microsoft\Windows\Task）
func normalizeName(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return "\\" + rel
}

// ParseXML 解析任务XML内容并返回任务信息
//   data - 任务XML数据（UTF-8或UTF-16均可）
//   返回 - 解析后的任务信息
//   返回 - 错误信息
func ParseXML(data []byte) (*TaskInfo, error) {
	return parseTaskXMLBytes(data)
}

// parseTaskXML 读取并解析任务XML文件
func parseTaskXML(path string) (*TaskInfo, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parseTaskXMLBytes(data)
}

// parseTaskXMLBytes 从字节数据解析任务信息
func parseTaskXMLBytes(data []byte) (*TaskInfo, error) {
	// Convert UTF-16 to UTF-8 if needed
	data = decodeUTF16(data)

	// Strip namespace and encoding declarations to simplify parsing
	data = bytes.ReplaceAll(data, []byte(`xmlns="http://schemas.microsoft.com/windows/2004/02/mit/task"`), nil)
	data = bytes.ReplaceAll(data, []byte(` encoding="UTF-16"`), nil)
	data = bytes.ReplaceAll(data, []byte(` encoding="UTF-8"`), nil)

	var raw taskXML
	if err := xml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("xml unmarshal failed: %w", err)
	}

	t := &TaskInfo{}

	if raw.RegistrationInfo != nil {
		t.Author = raw.RegistrationInfo.Author
	}

	// Enabled: default true, override if <Enabled> exists
	t.Enabled = true
	if raw.Settings != nil {
		// Trim whitespace since some task XMLs have spaces around the value
		enabled := strings.TrimSpace(raw.Settings.Enabled)
		if enabled != "" {
			t.Enabled = strings.EqualFold(enabled, "true")
		}
	}

	if raw.Actions != nil {
		if raw.Actions.Exec != nil {
			t.Command = raw.Actions.Exec.Command
			t.Arguments = raw.Actions.Exec.Arguments
		}
		if raw.Actions.ComHandler != nil {
			t.ComHandler = true
			t.Clsid = raw.Actions.ComHandler.Clsid
		}
	}

	if raw.Principals != nil && len(raw.Principals.Principal) > 0 {
		t.UserId = raw.Principals.Principal[0].UserId
	}

	return t, nil
}

// decodeUTF16 将UTF-16带BOM的数据转换为UTF-8
func decodeUTF16(data []byte) []byte {
	if len(data) < 2 {
		return data
	}
	var u16 []uint16
	if data[0] == 0xFF && data[1] == 0xFE {
		// UTF-16LE BOM
		u16 = make([]uint16, (len(data)-2)/2)
		for i := range u16 {
			u16[i] = uint16(data[2*i+2]) | uint16(data[2*i+3])<<8
		}
	} else if data[0] == 0xFE && data[1] == 0xFF {
		// UTF-16BE BOM
		u16 = make([]uint16, (len(data)-2)/2)
		for i := range u16 {
			u16[i] = uint16(data[2*i+3]) | uint16(data[2*i+2])<<8
		}
	} else {
		return data // already UTF-8 or no BOM
	}
	runes := utf16.Decode(u16)
	return []byte(string(runes))
}
