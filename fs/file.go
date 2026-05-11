//go:build windows

package fs

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

const maxPath = 260

// InfoType 表示文件版本信息查询类型的字符串类型。
type InfoType string

const (
	// FileDescription 文件描述信息。
	FileDescription  InfoType = "FileDescription"
	// CompanyName 公司名称信息。
	CompanyName      InfoType = "CompanyName"
	// OriginalFileName 原始文件名信息。
	OriginalFileName InfoType = "OriginalFileName"
	// LegalCopyright 法律版权信息。
	LegalCopyright   InfoType = "LegalCopyright"
	// ProductName 产品名称信息。
	ProductName      InfoType = "ProductName"
	// ProductVersion 产品版本信息。
	ProductVersion   InfoType = "ProductVersion"
)

// DirEntry 表示目录条目信息
type DirEntry struct {
	// Name 文件名（不含路径）
	Name string
	// Path 完整路径
	Path string
	// Size 文件大小（字节），目录为0
	Size int64
	// IsDir 是否为目录
	IsDir bool
	// ModTime 最后修改时间
	ModTime time.Time
}

// ListDir 列出指定目录下的所有条目
//   path - 目录路径
//   返回 - 目录条目列表
//   返回 - 错误信息
func ListDir(path string) ([]DirEntry, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("read dir failed: %w", err)
	}
	result := make([]DirEntry, 0, len(entries))
	for _, e := range entries {
		info, err := e.Info()
		if err != nil {
			continue
		}
		result = append(result, DirEntry{
			Name:    e.Name(),
			Path:    filepath.Join(path, e.Name()),
			Size:    info.Size(),
			IsDir:   e.IsDir(),
			ModTime: info.ModTime(),
		})
	}
	return result, nil
}

// CreateTime 返回指定文件的创建时间戳（Unix 时间戳）。
//   path - 目标文件路径。
//   返回 - Unix 时间戳（秒），失败时返回错误。
func CreateTime(path string) (int64, error) {
	var creationTime windows.Filetime
	err := getFileTime(path, &creationTime, nil, nil)
	if err != nil {
		return 0, fmt.Errorf("getFileTime failed: %w", err)
	}
	return creationTime.Nanoseconds() / 1e9, nil
}

// AccessTime 返回指定文件的最后访问时间戳（Unix 时间戳）。
//   path - 目标文件路径。
//   返回 - Unix 时间戳（秒），失败时返回错误。
func AccessTime(path string) (int64, error) {
	var lastAccessTime windows.Filetime
	err := getFileTime(path, nil, &lastAccessTime, nil)
	if err != nil {
		return 0, fmt.Errorf("getFileTime failed: %w", err)
	}
	return lastAccessTime.Nanoseconds() / 1e9, nil
}

// ModifyTime 返回指定文件的最后修改时间戳（Unix 时间戳）。
//   path - 目标文件路径。
//   返回 - Unix 时间戳（秒），失败时返回错误。
func ModifyTime(path string) (int64, error) {
	var lastWriteTime windows.Filetime
	err := getFileTime(path, nil, nil, &lastWriteTime)
	if err != nil {
		return 0, fmt.Errorf("getFileTime failed: %w", err)
	}
	return lastWriteTime.Nanoseconds() / 1e9, nil
}

// getFileTime opens the file and retrieves its timestamps.
func getFileTime(path string, creationTime, lastAccessTime, lastWriteTime *windows.Filetime) error {
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

	err = windows.GetFileTime(handle, creationTime, lastAccessTime, lastWriteTime)
	if err != nil {
		return fmt.Errorf("GetFileTime failed: %w", err)
	}
	return nil
}

// Info 从文件中检索指定的版本信息字段。
//   path           - 目标文件路径。
//   infoType       - 要检索的版本信息类型（如 FileDescription、ProductName 等）。
//   subTranslation - 可选的语言/代码页翻译字符串指针（nil 表示自动检测）。
//   返回 - 查询到的版本信息字符串，失败时返回错误。
func Info(path string, infoType InfoType, subTranslation *string) (string, error) {
	versionInfo, err := getResourceVersionInfo(path)
	if err != nil {
		return "", fmt.Errorf("getResourceVersionInfo failed: %w", err)
	}
	return getFileInfoByBlock(versionInfo, infoType, subTranslation)
}

// getFileInfoByBlock extracts version information from a version resource block.
func getFileInfoByBlock(versionInfo []byte, infoType InfoType, subTranslation *string) (string, error) {
	type langAndCodePage struct {
		Language uint16
		CodePage uint16
	}

	var (
		translation *langAndCodePage
		bufferSize  uint32
		subBlock    string
	)

	if subTranslation == nil {
		err := windows.VerQueryValue(
			unsafe.Pointer(&versionInfo[0]),
			"\\VarFileInfo\\Translation",
			unsafe.Pointer(&translation),
			&bufferSize,
		)
		if err != nil || translation == nil || bufferSize == 0 {
			return "", fmt.Errorf("failed to get language/codepage: %w", err)
		}
		subBlock = fmt.Sprintf(
			"\\StringFileInfo\\%04X%04X\\%s",
			translation.Language,
			translation.CodePage,
			infoType,
		)
	} else {
		subBlock = fmt.Sprintf("\\StringFileInfo\\%s\\%s", *subTranslation, infoType)
	}

	if len(subBlock) > maxPath {
		return "", fmt.Errorf("query path too long (max %d bytes)", maxPath)
	}

	var infoBuf unsafe.Pointer
	err := windows.VerQueryValue(
		unsafe.Pointer(&versionInfo[0]),
		subBlock,
		unsafe.Pointer(&infoBuf),
		&bufferSize,
	)
	if err != nil || infoBuf == nil || bufferSize == 0 {
		return "", fmt.Errorf("failed to get %s info: %w", infoType, err)
	}

	return windows.UTF16PtrToString((*uint16)(infoBuf)), nil
}

// getResourceVersionInfo retrieves the raw version resource data for a file.
func getResourceVersionInfo(path string) ([]byte, error) {
	var zeroHandle windows.Handle
	bufferSize, err := windows.GetFileVersionInfoSize(path, &zeroHandle)
	if err != nil {
		return nil, fmt.Errorf("GetFileVersionInfoSize failed: %w", err)
	}
	if bufferSize == 0 {
		return nil, fmt.Errorf("file has no version information")
	}

	buffer := make([]byte, bufferSize)
	err = windows.GetFileVersionInfo(path, 0, bufferSize, unsafe.Pointer(&buffer[0]))
	if err != nil {
		return nil, fmt.Errorf("GetFileVersionInfo failed: %w", err)
	}
	return buffer, nil
}

// VolumeInfo 表示逻辑卷信息
type VolumeInfo struct {
	// Name 卷名（如 C:\）
	Name string
	// Label 卷标
	Label string
	// Type 文件系统类型（NTFS, FAT32 等）
	Type string
	// TotalSize 总大小（字节）
	TotalSize uint64
	// FreeSize 可用大小（字节）
	FreeSize uint64
}

// Volumes 返回所有逻辑卷信息
//	 返回 - 逻辑卷信息列表
//	 返回 - 错误信息
func Volumes() ([]VolumeInfo, error) {
	drives, err := windows.GetLogicalDrives()
	if err != nil {
		return nil, fmt.Errorf("GetLogicalDrives failed: %w", err)
	}

	var vols []VolumeInfo
	for i := 0; i < 26; i++ {
		if drives&(1<<i) == 0 {
			continue
		}
		root := string(rune('A'+i)) + ":\\"
		info, err := queryVolume(root)
		if err != nil {
			continue
		}
		vols = append(vols, info)
	}
	return vols, nil
}

// queryVolume queries information for a specific volume root path.
func queryVolume(root string) (VolumeInfo, error) {
	rootPtr := windows.StringToUTF16Ptr(root)

	var nameBuf [256]uint16
	var fsBuf [128]uint16
	var serial, maxComp uint32
	var flags uint32

	err := windows.GetVolumeInformation(
		rootPtr,
		&nameBuf[0], uint32(len(nameBuf)),
		&serial, &maxComp, &flags,
		&fsBuf[0], uint32(len(fsBuf)),
	)
	if err != nil {
		return VolumeInfo{}, err
	}

	var freeBytesAvailable, totalBytes, totalFreeBytes uint64
	err = windows.GetDiskFreeSpaceEx(rootPtr, &freeBytesAvailable, &totalBytes, &totalFreeBytes)
	if err != nil {
		return VolumeInfo{}, err
	}

	return VolumeInfo{
		Name:      root,
		Label:     windows.UTF16ToString(nameBuf[:]),
		Type:      windows.UTF16ToString(fsBuf[:]),
		TotalSize: totalBytes,
		FreeSize:  totalFreeBytes,
	}, nil
}

// VersionInfo 返回文件的版本字符串（主版本.次版本.构建号.修订号）。
//   path - 目标文件路径。
//   返回 - 格式为 "major.minor.build.revision" 的版本字符串，失败时返回错误。
func VersionInfo(path string) (string, error) {
	versionInfo, err := getResourceVersionInfo(path)
	if err != nil {
		return "", fmt.Errorf("getResourceVersionInfo failed: %w", err)
	}

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

	major := fixedInfo.FileVersionMS >> 16
	minor := fixedInfo.FileVersionMS & 0xFFFF
	build := fixedInfo.FileVersionLS >> 16
	revision := fixedInfo.FileVersionLS & 0xFFFF

	return fmt.Sprintf("%d.%d.%d.%d", major, minor, build, revision), nil
}

// Sha1 返回指定文件的 SHA1 哈希值（十六进制小写字符串）。
//   path - 目标文件路径。
//   返回 - 40 字符的 SHA1 十六进制字符串，失败时返回错误。
func Sha1(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("open file failed: %w", err)
	}
	defer f.Close()

	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("sha1 compute failed: %w", err)
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
