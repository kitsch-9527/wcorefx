//go:build windows
// +build windows

package fs

import (
	"fmt"
	"unsafe"

	comm "github.com/kitsch-9527/wcorefx/common"
	"golang.org/x/sys/windows"
)

// InfoType 表示文件资源信息的类型
type InfoType string

// 文件信息类型常量定义，对应Windows资源信息中的标准字段
const (
	FileDescription  InfoType = "FileDescription"  // 文件描述信息
	CompanyName      InfoType = "CompanyName"      // 公司名称
	InternalName     InfoType = "InternalName"     // 内部名称
	OriginalFileName InfoType = "OriginalFileName" // 原始文件名
	LegalCopyright   InfoType = "LegalCopyright"   // 版权信息
	ProductName      InfoType = "ProductName"      // 产品名称
	ProductVersion   InfoType = "ProductVersion"   // 产品版本
)

// GetFileCreateTime 文件创建时间 返回时间戳
func GetFileCreateTime(path string) (int64, error) {
	var creationTime windows.Filetime
	err := GetFileTime(path, &creationTime, nil, nil)
	if err != nil {
		return 0, fmt.Errorf("GetFileTime failed: %w", err)
	}
	return creationTime.Nanoseconds() / 1e9, nil
}

// GetFileAccessTime 文件访问时间
func GetFileAccessTime(path string) (int64, error) {
	var lastAccessTime windows.Filetime
	err := GetFileTime(path, nil, &lastAccessTime, nil)
	if err != nil {
		return 0, fmt.Errorf("GetFileTime failed: %w", err)
	}
	return lastAccessTime.Nanoseconds() / 1e9, nil
}

// GetFileModifyTime 文件修改时间
func GetFileModifyTime(path string) (int64, error) {
	var lastWriteTime windows.Filetime
	err := GetFileTime(path, nil, nil, &lastWriteTime)
	if err != nil {
		fmt.Errorf("GetFileTime failed: %w", err)
		return 0, err
	}
	return lastWriteTime.Nanoseconds() / 1e9, nil
}

// GetFileTime 获取文件的创建、访问、修改时间
func GetFileTime(path string, creationTime, lastAccessTime, lastWriteTime *windows.Filetime) error {
	handle, err := windows.CreateFile(
		windows.StringToUTF16Ptr(path),
		windows.GENERIC_READ,
		windows.FILE_SHARE_READ,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_ATTRIBUTE_NORMAL,
		0,
	)
	if err != nil {
		return fmt.Errorf("CreateFile failed: %w", err)
	}
	defer windows.CloseHandle(handle)
	err = windows.GetFileTime(
		handle,
		creationTime,
		lastAccessTime,
		lastWriteTime,
	)
	if err != nil {
		return fmt.Errorf("GetFileTime failed: %w", err)
	}
	return nil
}

// GetFileInfo 获取文件指定类型的资源信息
// path: 文件路径
// infoType: 要获取的信息类型（如FileDescription、CompanyName等）
// subTranslation: 可选的语言/代码页组合（格式为"HHHHHHHH"，如"040904B0"），为nil则自动获取
// 返回值: 对应的信息字符串和可能的错误
func GetFileInfo(path string, infoType InfoType, subTranslation *string) (string, error) {
	versionInfo, err := getResourceVersionInfo(path)
	if err != nil {
		return "", fmt.Errorf("getResourceVersionInfo failed: %w", err)
	}
	return getFileInfoByBlock(versionInfo, infoType, subTranslation)
}

// getFileInfoByBlock 从版本信息块中提取指定类型的信息
func getFileInfoByBlock(versionInfo []byte, infoType InfoType, subTranslation *string) (string, error) {
	// LANGANDCODEPAGE 对应Windows API中的语言和代码页结构
	type LANGANDCODEPAGE struct {
		Language uint16
		CodePage uint16
	}

	var (
		translation *LANGANDCODEPAGE
		bufferSize  uint32
		subBlock    string
	)

	// 当未指定子翻译时，自动获取系统默认的语言和代码页
	if subTranslation == nil {
		// 查询语言和代码页信息（对应Windows的Translation资源）
		err := windows.VerQueryValue(
			unsafe.Pointer(&versionInfo[0]),
			"\\VarFileInfo\\Translation",
			unsafe.Pointer(&translation),
			&bufferSize,
		)
		if err != nil || translation == nil || bufferSize == 0 {
			return "", fmt.Errorf("获取语言/代码页信息失败: %w", err)
		}

		// 构建查询路径，格式为\StringFileInfo\语言代码页\信息类型
		subBlock = fmt.Sprintf(
			"\\StringFileInfo\\%04X%04X\\%s",
			translation.Language,
			translation.CodePage,
			infoType,
		)
	} else {
		// 使用指定的语言/代码页组合
		subBlock = fmt.Sprintf("\\StringFileInfo\\%s\\%s", *subTranslation, infoType)
	}

	// 检查路径长度是否超过系统限制
	if len(subBlock) > comm.MAXPATH {
		return "", fmt.Errorf("查询路径过长（最大%d字节）", comm.MAXPATH)
	}

	// 提取指定类型的具体信息
	var infoBuf unsafe.Pointer
	err := windows.VerQueryValue(
		unsafe.Pointer(&versionInfo[0]),
		subBlock,
		unsafe.Pointer(&infoBuf),
		&bufferSize,
	)
	if err != nil || infoBuf == nil || bufferSize == 0 {
		return "", fmt.Errorf("获取%s信息失败: %w", infoType, err)
	}

	// 将UTF-16编码的结果转换为Go字符串
	return windows.UTF16PtrToString((*uint16)(infoBuf)), nil
}

// getResourceVersionInfo 获取文件的完整版本资源信息
// path: 文件路径
// 返回值: 版本信息字节数组和可能的错误
func getResourceVersionInfo(path string) ([]byte, error) {
	var zeroHandle = windows.Handle(0)

	// 获取版本信息缓冲区大小
	bufferSize, err := windows.GetFileVersionInfoSize(path, &zeroHandle)
	if err != nil {
		return nil, fmt.Errorf("GetFileVersionInfoSize failed: %w", err)
	}
	if bufferSize == 0 {
		return nil, fmt.Errorf("file has no version information")
	}

	// 申请缓冲区并读取版本信息
	buffer := make([]byte, bufferSize)
	err = windows.GetFileVersionInfo(
		path,
		0,
		bufferSize,
		unsafe.Pointer(&buffer[0]),
	)
	if err != nil {
		return nil, fmt.Errorf("GetFileVersionInfo failed: %w", err)
	}

	return buffer, nil
}

// GetFileVersionInfo 获取文件的版本号（格式为major.minor.build.revision）
// path: 文件路径
// 返回值: 版本号字符串和可能的错误
func GetFileVersionInfo(path string) (string, error) {
	versionInfo, err := getResourceVersionInfo(path)
	if err != nil {
		return "", fmt.Errorf("getResourceVersionInfo failed: %w", err)
	}

	// 解析固定版本信息结构
	var fixedInfo *windows.VS_FIXEDFILEINFO
	bufferSize := uint32(unsafe.Sizeof(*fixedInfo))
	err = windows.VerQueryValue(
		unsafe.Pointer(&versionInfo[0]),
		"\\",
		unsafe.Pointer(&fixedInfo),
		&bufferSize,
	)
	if err != nil || fixedInfo == nil {
		return "", fmt.Errorf("VerQueryValue failed: %w", err)
	}

	// 计算版本号各部分（高16位和低16位拆分）
	major := fixedInfo.FileVersionMS >> 16
	minor := fixedInfo.FileVersionMS & 0xFFFF
	build := fixedInfo.FileVersionLS >> 16
	revision := fixedInfo.FileVersionLS & 0xFFFF

	return fmt.Sprintf("%d.%d.%d.%d", major, minor, build, revision), nil
}
