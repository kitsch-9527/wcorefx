//go:build windows

// Package os provides system information functions for Windows.
package os

import (
	"fmt"
	"os"
	"strings"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// MemoryInfo 保存系统内存信息
type MemoryInfo struct {
	// TotalPhysical 物理内存总量（字节）
	TotalPhysical uint64
	// AvailablePhysical 可用物理内存（字节）
	AvailablePhysical uint64
	// UsedPhysical 已用物理内存（字节）
	UsedPhysical uint64
	// TotalVirtual 虚拟内存总量（字节，含 pagefile）
	TotalVirtual uint64
	// AvailableVirtual 可用虚拟内存（字节）
	AvailableVirtual uint64
	// UsedVirtual 已用虚拟内存（字节）
	UsedVirtual uint64
	// MemoryLoad 内存使用率（0-100）
	MemoryLoad uint32
}

// DriveInfo 保存逻辑驱动器信息
type DriveInfo struct {
	// Drive 盘符（如 "C:"）
	Drive string
	// Type 驱动器类型（如 "Fixed"、"CD-ROM"、"Removable"）
	Type string
	// TotalBytes 驱动器总空间（字节），不可用时为 0
	TotalBytes uint64
	// FreeBytes 驱动器可用空间（字节），不可用时为 0
	FreeBytes uint64
}

// Is64 返回操作系统是否为64位
//   返回 - 64位返回true，否则返回false
func Is64() bool {
	info := getNativeSystemInfo()
	return info.wProcessorArchitecture == 9 // PROCESSOR_ARCHITECTURE_AMD64
}

// IsVistaUpper 返回操作系统版本是否为Vista或更高版本（主版本号>=6）
//   返回 - Vista或更高版本返回true，否则返回false
func IsVistaUpper() bool {
	return MajorVersion() >= 6
}

// MajorVersion 返回操作系统内核主版本号
//   返回 - 主版本号
func MajorVersion() uint32 {
	major, _, _ := rtlGetNtVersionNumbers()
	return major
}

// MinorVersion 返回操作系统内核副版本号
//   返回 - 副版本号
func MinorVersion() uint32 {
	_, minor, _ := rtlGetNtVersionNumbers()
	return minor
}

// BuildNumber 返回操作系统内核构建号
//   返回 - 内核构建号
func BuildNumber() uint32 {
	_, _, build := rtlGetNtVersionNumbers()
	return build
}

// ReleaseID 返回Windows发行版本标识符（如"22H2"）
//   返回 - 发行版本标识符，未知版本返回空字符串
func ReleaseID() string {
	pairs := map[uint32]string{
		10240: "1507",
		10586: "1511",
		14393: "1607",
		15063: "1703",
		16299: "1709",
		17134: "1803",
		17763: "1809",
		18362: "1903",
		18363: "1909",
		19041: "2004",
		19042: "20H2",
		19043: "21H1",
		19044: "21H2",
		19045: "22H2",
		22000: "21H2",
		22621: "22H2",
		22631: "23H2",
		26100: "24H2",
	}
	if s, ok := pairs[BuildNumber()]; ok {
		return s
	}
	return ""
}

// VersionInfo 返回人类可读的Windows版本字符串
//   返回 - 人类可读的版本字符串（如"Windows 10"、"Windows 11"）
func VersionInfo() string {
	major := MajorVersion()
	minor := MinorVersion()
	build := BuildNumber()

	switch {
	case major == 10 && build >= 22000:
		return "Windows 11"
	case major == 10:
		return "Windows 10"
	case major == 6 && minor == 3:
		return "Windows 8.1"
	case major == 6 && minor == 2:
		return "Windows 8"
	case major == 6 && minor == 1:
		return "Windows 7"
	case major == 6 && minor == 0:
		return "Windows Vista"
	case major == 5 && minor == 1:
		return "Windows XP"
	case major == 5 && minor == 0:
		return "Windows 2000"
	default:
		return fmt.Sprintf("Windows %d.%d (build %d)", major, minor, build)
	}
}

// CPUCount 返回CPU处理器数量
//   返回 - 逻辑处理器数量
func CPUCount() uint32 {
	info := getNativeSystemInfo()
	return info.dwNumberOfProcessors
}

// TickCount 返回系统运行时间（毫秒）
//   返回 - 自系统启动以来的毫秒数
func TickCount() uint64 {
	return getTickCount64()
}

// StartupTime 返回系统启动时间
//   返回 - 系统启动的时间点
func StartupTime() time.Time {
	now := time.Now()
	uptime := time.Duration(TickCount()) * time.Millisecond
	return now.Add(-uptime)
}

// NetBiosName 返回NetBIOS计算机名
//   返回 - NetBIOS计算机名
//   返回 - 错误信息
func NetBiosName() (string, error) {
	var buf [windows.MAX_COMPUTERNAME_LENGTH + 1]uint16
	size := uint32(len(buf))
	err := windows.GetComputerName(&buf[0], &size)
	if err != nil {
		return "", fmt.Errorf("GetComputerName failed: %w", err)
	}
	return windows.UTF16ToString(buf[:size]), nil
}

// HostName 返回DNS主机名
//   返回 - DNS主机名
//   返回 - 错误信息
func HostName() (string, error) {
	n, err := os.Hostname()
	if err != nil {
		return "", fmt.Errorf("os.Hostname failed: %w", err)
	}
	return n, nil
}

// UserName 返回当前用户名
//   返回 - 当前用户名（格式：域名\用户名）
//   返回 - 错误信息
func UserName() (string, error) {
	var size uint32
	// First call to get buffer size
	windows.GetUserNameEx(3, nil, &size) // NameSamCompatible
	if size == 0 {
		// Fallback: use environment variable
		user := os.Getenv("USERNAME")
		if user != "" {
			return user, nil
		}
		return "", fmt.Errorf("GetUserNameEx failed to return size")
	}
	buf := make([]uint16, size)
	err := windows.GetUserNameEx(3, &buf[0], &size)
	if err != nil {
		return "", fmt.Errorf("GetUserNameEx failed: %w", err)
	}
	return windows.UTF16ToString(buf[:size]), nil
}

// SessionUserName 返回指定会话ID的用户名
// Use ^uint32(0) (WTS_CURRENT_SESSION) for the current session.
//   sessionID - 会话ID，使用^uint32(0)表示当前会话
//   返回 - 用户名
//   返回 - 错误信息
func SessionUserName(sessionID uint32) (string, error) {
	return wtsQuerySessionInformation(sessionID)
}

// WinDir 返回Windows目录（如C:\Windows）
//   返回 - Windows目录路径
//   返回 - 错误信息
func WinDir() (string, error) {
	n, err := windows.GetWindowsDirectory()
	if err != nil {
		return "", fmt.Errorf("GetWindowsDirectory failed: %w", err)
	}
	return n, nil
}

// SystemDir 返回系统目录（32位）
// On 64-bit systems, this returns Syswow64Dir; on 32-bit, System32Dir.
//   返回 - 系统目录路径
//   返回 - 错误信息
func SystemDir() (string, error) {
	if Is64() {
		return Syswow64Dir()
	}
	return System32Dir()
}

// System32Dir 返回System32目录
//   返回 - System32目录路径
//   返回 - 错误信息
func System32Dir() (string, error) {
	winDir, err := WinDir()
	if err != nil {
		return "", err
	}
	return winDir + "\\System32", nil
}

// Syswow64Dir 返回SysWOW64目录
//   返回 - SysWOW64目录路径
//   返回 - 错误信息
func Syswow64Dir() (string, error) {
	winDir, err := WinDir()
	if err != nil {
		return "", err
	}
	return winDir + "\\SysWOW64", nil
}

// Getenv 获取环境变量值（支持环境变量扩展）
//   name - 环境变量名
//   返回 - 扩展后的环境变量值
func Getenv(name string) string {
	return os.ExpandEnv(os.Getenv(name))
}

// Environ 返回所有环境变量映射
//   返回 - 环境变量名到值的映射
func Environ() map[string]string {
	env := os.Environ()
	m := make(map[string]string, len(env))
	for _, e := range env {
		if k, v, ok := strings.Cut(e, "="); ok {
			m[k] = v
		}
	}
	return m
}

// DosErrorMsg 返回Windows错误码对应的错误信息
//   errCode - Windows错误码
//   返回 - 对应的错误描述信息
func DosErrorMsg(errCode uint32) string {
	flags := uint32(windows.FORMAT_MESSAGE_FROM_SYSTEM | windows.FORMAT_MESSAGE_IGNORE_INSERTS)
	var buf [512]uint16
	n, err := windows.FormatMessage(flags, 0, errCode, 0, buf[:], nil)
	if err != nil {
		return fmt.Sprintf("unknown error %d", errCode)
	}
	msg := windows.UTF16ToString(buf[:n])
	msg = strings.TrimRight(msg, "\r\n ")
	return msg
}

// Reboot 重启系统
//   返回 - 错误信息
func Reboot() error {
	return exitWindowsEx(0x00000006, 0) // EWX_REBOOT
}

// Poweroff 关闭系统
//   返回 - 错误信息
func Poweroff() error {
	return exitWindowsEx(0x00000008, 0) // EWX_POWEROFF
}

// Memory 返回系统物理内存和虚拟内存信息
//   返回 - 系统内存信息
//   返回 - 错误信息
func Memory() (MemoryInfo, error) {
	ms, err := globalMemoryStatusEx()
	if err != nil {
		return MemoryInfo{}, err
	}
	return MemoryInfo{
		TotalPhysical:     ms.ullTotalPhys,
		AvailablePhysical: ms.ullAvailPhys,
		UsedPhysical:      ms.ullTotalPhys - ms.ullAvailPhys,
		TotalVirtual:      ms.ullTotalPageFile,
		AvailableVirtual:  ms.ullAvailPageFile,
		UsedVirtual:       ms.ullTotalPageFile - ms.ullAvailPageFile,
		MemoryLoad:        ms.dwMemoryLoad,
	}, nil
}

// CPUModel 返回 CPU 型号名称
//   返回 - CPU 型号名称（如 "Intel(R) Core(TM) i7-10700K CPU @ 3.80GHz"）
//   返回 - 错误信息
func CPUModel() (string, error) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`HARDWARE\DESCRIPTION\System\CentralProcessor\0`,
		registry.QUERY_VALUE)
	if err != nil {
		return "", fmt.Errorf("open CPU registry key failed: %w", err)
	}
	defer k.Close()

	name, _, err := k.GetStringValue("ProcessorNameString")
	if err != nil {
		return "", fmt.Errorf("read ProcessorNameString failed: %w", err)
	}
	return strings.TrimSpace(name), nil
}

// Drives 返回所有逻辑驱动器信息（类型、总空间、可用空间）
//   返回 - 逻辑驱动器信息列表
//   返回 - 错误信息
func Drives() ([]DriveInfo, error) {
	buf := make([]uint16, 256)
	n, err := windows.GetLogicalDriveStrings(uint32(len(buf)), &buf[0])
	if err != nil {
		return nil, fmt.Errorf("GetLogicalDriveStrings failed: %w", err)
	}

	var drives []DriveInfo
	for _, s := range strings.Split(windows.UTF16ToString(buf[:n]), "\x00") {
		if s == "" {
			continue
		}
		root := windows.StringToUTF16Ptr(s)
		driveType := windows.GetDriveType(root)

		di := DriveInfo{
			Drive: s,
			Type:  driveTypeString(driveType),
		}

		// GetDiskFreeSpaceEx fails for some drive types (e.g. empty CD-ROM)
		var total, free uint64
		if err := windows.GetDiskFreeSpaceEx(root, nil, &total, &free); err == nil {
			di.TotalBytes = total
			di.FreeBytes = free
		}

		drives = append(drives, di)
	}
	return drives, nil
}

// --- private helpers ---

func driveTypeString(dt uint32) string {
	switch dt {
	case 2:
		return "Removable"
	case 3:
		return "Fixed"
	case 4:
		return "Remote"
	case 5:
		return "CD-ROM"
	case 6:
		return "RAM Disk"
	case 1:
		return "No Root"
	default:
		return "Unknown"
	}
}

func rtlGetNtVersionNumbers() (major, minor, build uint32) {
	mod := windows.NewLazySystemDLL("ntdll.dll")
	proc := mod.NewProc("RtlGetNtVersionNumbers")
	proc.Call(
		uintptr(unsafe.Pointer(&major)),
		uintptr(unsafe.Pointer(&minor)),
		uintptr(unsafe.Pointer(&build)),
	)
	build &^= 0xF0000000
	return
}
