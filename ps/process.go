//go:build windows

// Package ps provides Windows process information functions.
package ps

import (
	"fmt"
	"strings"
	"unsafe"

	"github.com/kitsch-9527/wcorefx/obj"
	"golang.org/x/sys/windows"
)

// ProcessTimes 保存进程时间信息
type ProcessTimes struct {
	// CreationTime 进程创建时间
	CreationTime windows.Filetime
	// ExitTime 进程退出时间
	ExitTime     windows.Filetime
	// KernelTime 进程内核态时间
	KernelTime   windows.Filetime
	// UserTime 进程用户态时间
	UserTime     windows.Filetime
}

const (
	maxPath = 260
)

// List 返回所有正在运行的进程列表
//   返回 - 进程条目列表
//   返回 - 错误信息
func List() ([]windows.ProcessEntry32, error) {
	h, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, fmt.Errorf("CreateToolhelp32Snapshot failed: %w", err)
	}
	defer windows.CloseHandle(h)

	var pe windows.ProcessEntry32
	pe.Size = uint32(unsafe.Sizeof(pe))
	err = windows.Process32First(h, &pe)
	if err != nil {
		return nil, fmt.Errorf("Process32First failed: %w", err)
	}

	var procs []windows.ProcessEntry32
	for {
		procs = append(procs, pe)
		err = windows.Process32Next(h, &pe)
		if err != nil {
			break
		}
	}
	return procs, nil
}

// Find 根据可执行文件名（模糊匹配）查找进程
//   name - 可执行文件名（支持模糊匹配）
//   返回 - 匹配的进程条目列表
//   返回 - 错误信息
func Find(name string) ([]windows.ProcessEntry32, error) {
	procs, err := List()
	if err != nil {
		return nil, err
	}
	nameUpper := strings.ToUpper(name)
	var matches []windows.ProcessEntry32
	for _, p := range procs {
		exeName := strings.ToUpper(windows.UTF16ToString(p.ExeFile[:]))
		if strings.Contains(exeName, nameUpper) {
			matches = append(matches, p)
		}
	}
	return matches, nil
}

// openWithMinimalAccess opens a process handle with PROCESS_QUERY_LIMITED_INFORMATION.
func openWithMinimalAccess(pid uint32) (windows.Handle, error) {
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return 0, fmt.Errorf("OpenProcess failed: %w", err)
	}
	return h, nil
}

// openWithFullAccess opens a process handle with PROCESS_ALL_ACCESS.
func openWithFullAccess(pid uint32) (windows.Handle, error) {
	h, err := windows.OpenProcess(windows.PROCESS_ALL_ACCESS, true, pid)
	if err != nil {
		return 0, fmt.Errorf("OpenProcess failed: %w", err)
	}
	return h, nil
}

// getBasicInfo returns PROCESS_BASIC_INFORMATION for the given handle.
func getBasicInfo(h windows.Handle) (windows.PROCESS_BASIC_INFORMATION, error) {
	var retlen uint32
	var pbi windows.PROCESS_BASIC_INFORMATION
	err := windows.NtQueryInformationProcess(h, windows.ProcessBasicInformation, unsafe.Pointer(&pbi), uint32(unsafe.Sizeof(pbi)), &retlen)
	if err != nil {
		return pbi, fmt.Errorf("NtQueryInformationProcess failed: %w", err)
	}
	return pbi, nil
}

// CommandLine 返回指定PID进程的命令行参数
//   pid - 进程ID
//   返回 - 命令行参数字符串
//   返回 - 错误信息
func CommandLine(pid uint32) (string, error) {
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_INFORMATION|windows.PROCESS_VM_READ, false, pid)
	if err != nil {
		return "", fmt.Errorf("OpenProcess failed: %w", err)
	}
	defer windows.CloseHandle(h)

	basicInfo, err := getBasicInfo(h)
	if err != nil {
		return "", err
	}
	pebAddr := basicInfo.PebBaseAddress
	if pebAddr == nil {
		return "", fmt.Errorf("PEB address is nil")
	}

	var peb windows.PEB
	var bytesRead uintptr
	err = windows.ReadProcessMemory(h, uintptr(unsafe.Pointer(pebAddr)), (*byte)(unsafe.Pointer(&peb)), uintptr(unsafe.Sizeof(peb)), &bytesRead)
	if err != nil {
		return "", fmt.Errorf("ReadProcessMemory PEB failed: %w", err)
	}

	procParamsAddr := peb.ProcessParameters
	if procParamsAddr == nil {
		return "", fmt.Errorf("ProcessParameters is nil")
	}

	var procParams windows.RTL_USER_PROCESS_PARAMETERS
	err = windows.ReadProcessMemory(h, uintptr(unsafe.Pointer(procParamsAddr)), (*byte)(unsafe.Pointer(&procParams)), uintptr(unsafe.Sizeof(procParams)), &bytesRead)
	if err != nil {
		return "", fmt.Errorf("ReadProcessMemory ProcessParameters failed: %w", err)
	}

	cmdLine := procParams.CommandLine
	if cmdLine.Buffer == nil || cmdLine.Length == 0 {
		return "", nil
	}

	bufSize := uintptr(cmdLine.Length)
	buf := make([]uint16, bufSize/2)
	err = windows.ReadProcessMemory(h, uintptr(unsafe.Pointer(cmdLine.Buffer)), (*byte)(unsafe.Pointer(&buf[0])), bufSize, &bytesRead)
	if err != nil {
		return "", fmt.Errorf("ReadProcessMemory command line failed: %w", err)
	}
	return windows.UTF16ToString(buf), nil
}

// MemoryInfo 返回指定PID进程的内存计数器信息
//   pid - 进程ID
//   返回 - 进程内存计数器信息
//   返回 - 错误信息
func MemoryInfo(pid uint32) (processMemoryCounters, error) {
	h, err := openWithMinimalAccess(pid)
	if err != nil {
		return processMemoryCounters{}, err
	}
	defer windows.CloseHandle(h)
	return getProcessMemoryInfo(h)
}

// Times 返回指定PID进程的时间信息
//   pid - 进程ID
//   返回 - 进程时间信息（创建时间、退出时间、内核态时间、用户态时间）
//   返回 - 错误信息
func Times(pid uint32) (ProcessTimes, error) {
	h, err := openWithMinimalAccess(pid)
	if err != nil {
		return ProcessTimes{}, err
	}
	defer windows.CloseHandle(h)

	var creationTime, exitTime, kernelTime, userTime windows.Filetime
	err = windows.GetProcessTimes(h, &creationTime, &exitTime, &kernelTime, &userTime)
	if err != nil {
		return ProcessTimes{}, fmt.Errorf("GetProcessTimes failed: %w", err)
	}
	return ProcessTimes{
		CreationTime: creationTime,
		ExitTime:     exitTime,
		KernelTime:   kernelTime,
		UserTime:     userTime,
	}, nil
}

// Path 返回指定PID进程的可执行文件路径
//   pid - 进程ID
//   返回 - 可执行文件完整路径
//   返回 - 错误信息
func Path(pid uint32) (string, error) {
	if pid == 0 {
		return "System Idle Process", nil
	}
	if pid == 4 {
		return "System", nil
	}

	h, err := openWithMinimalAccess(pid)
	if err != nil {
		return "", err
	}
	defer windows.CloseHandle(h)

	var buf [maxPath + 1]uint16
	err = windows.GetModuleFileNameEx(h, 0, &buf[0], maxPath)
	if err == nil {
		return windows.UTF16ToString(buf[:]), nil
	}

	// Fallback: QueryFullProcessImageName + NativePathToDosPath
	var size uint32 = maxPath + 1
	var nativeBuf [maxPath + 1]uint16
	err = windows.QueryFullProcessImageName(h, windows.PROCESS_NAME_NATIVE, &nativeBuf[0], &size)
	if err != nil {
		return "", fmt.Errorf("QueryFullProcessImageName failed: %w", err)
	}
	nativePath := windows.UTF16ToString(nativeBuf[:size])
	return obj.NativePathToDosPath(nativePath)
}

// User 返回指定PID进程所属的域\用户名
//   pid - 进程ID
//   返回 - 域名
//   返回 - 用户名
//   返回 - 错误信息
func User(pid uint32) (domain, username string, err error) {
	token, err := openToken(pid)
	if err != nil {
		return "", "", err
	}
	defer token.Close()

	var size uint32
	windows.GetTokenInformation(token, windows.TokenUser, nil, 0, &size)
	if size == 0 {
		return "", "", fmt.Errorf("GetTokenInformation failed to return size")
	}

	buffer := make([]byte, size)
	err = windows.GetTokenInformation(token, windows.TokenUser, &buffer[0], size, &size)
	if err != nil {
		return "", "", fmt.Errorf("GetTokenInformation failed: %w", err)
	}

	tokenUser := (*windows.Tokenuser)(unsafe.Pointer(&buffer[0]))
	return lookupSIDAccount(tokenUser.User.Sid)
}

// lookupSIDAccount resolves a SID to domain\username.
func lookupSIDAccount(sid *windows.SID) (domain, name string, err error) {
	var userSize, domainSize uint32
	var sidNameUse uint32
	windows.LookupAccountSid(nil, sid, nil, &userSize, nil, &domainSize, &sidNameUse)
	if userSize == 0 && domainSize == 0 {
		return "", "", fmt.Errorf("LookupAccountSid returned zero sizes")
	}

	userName := make([]uint16, userSize)
	domainName := make([]uint16, domainSize)
	err = windows.LookupAccountSid(nil, sid, &userName[0], &userSize, &domainName[0], &domainSize, &sidNameUse)
	if err != nil {
		return "", "", fmt.Errorf("LookupAccountSid failed: %w", err)
	}
	return windows.UTF16ToString(domainName[:domainSize]), windows.UTF16ToString(userName[:userSize]), nil
}

// IsTerminated 检查指定PID的进程是否已终止
//   pid - 进程ID
//   返回 - 已终止返回true，否则返回false
func IsTerminated(pid uint32) bool {
	h, err := openWithMinimalAccess(pid)
	if err != nil {
		return true
	}
	defer windows.CloseHandle(h)

	event, err := windows.WaitForSingleObject(h, 0)
	if err != nil {
		return false
	}
	return event == windows.WAIT_OBJECT_0
}

// ParentID 返回指定PID进程的父进程ID
//   pid - 进程ID
//   返回 - 父进程ID，获取失败返回0
func ParentID(pid uint32) uint32 {
	h, err := openWithMinimalAccess(pid)
	if err != nil {
		return 0
	}
	defer windows.CloseHandle(h)

	pbi, err := getBasicInfo(h)
	if err != nil {
		return 0
	}
	return uint32(pbi.InheritedFromUniqueProcessId)
}

// SessionID 返回指定PID进程的会话ID
//   pid - 进程ID
//   返回 - 会话ID
//   返回 - 错误信息
func SessionID(pid uint32) (uint32, error) {
	var sessionID uint32
	err := windows.ProcessIdToSessionId(pid, &sessionID)
	if err != nil {
		return 0, fmt.Errorf("ProcessIdToSessionId failed: %w", err)
	}
	return sessionID, nil
}

// ModuleInfo 保存已加载模块的信息
type ModuleInfo struct {
	// Base 模块基址句柄
	Base   windows.Handle
	// Size 模块镜像大小
	Size   uint32
	// Path 模块完整路径
	Path   string
	// Name 模块文件名
	Name   string
}

// Modules 返回指定PID进程加载的模块列表
//   pid - 进程ID
//   返回 - 已加载模块信息列表
//   返回 - 错误信息
func Modules(pid uint32) ([]ModuleInfo, error) {
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_INFORMATION|windows.PROCESS_VM_READ, false, pid)
	if err != nil {
		return nil, fmt.Errorf("OpenProcess failed: %w", err)
	}
	defer windows.CloseHandle(h)

	var needed uint32
	err = windows.EnumProcessModulesEx(h, nil, 0, &needed, windows.LIST_MODULES_ALL)
	if err != nil && err != windows.ERROR_INSUFFICIENT_BUFFER {
		return nil, fmt.Errorf("EnumProcessModulesEx failed: %w", err)
	}
	if needed == 0 {
		return nil, nil
	}

	count := int(needed) / 8 // 8 bytes per handle on 64-bit
	mods := make([]windows.Handle, count)
	err = windows.EnumProcessModulesEx(h, &mods[0], uint32(len(mods))*8, &needed, windows.LIST_MODULES_ALL)
	if err != nil {
		return nil, fmt.Errorf("EnumProcessModulesEx failed: %w", err)
	}

	modules := make([]ModuleInfo, 0, len(mods))
	for _, mod := range mods {
		var buf [maxPath + 1]uint16
		err = windows.GetModuleFileNameEx(h, mod, &buf[0], maxPath)
		if err != nil {
			continue
		}
		path := windows.UTF16ToString(buf[:])
		name := path[strings.LastIndex(path, "\\")+1:]

		var modInfo windows.ModuleInfo
		err = windows.GetModuleInformation(h, mod, &modInfo, uint32(unsafe.Sizeof(modInfo)))
		if err != nil {
			modules = append(modules, ModuleInfo{Base: mod, Path: path, Name: name})
			continue
		}
		modules = append(modules, ModuleInfo{
			Base: mod,
			Size: modInfo.SizeOfImage,
			Path: path,
			Name: name,
		})
	}
	return modules, nil
}

// FindByPath 根据可执行文件路径查找进程PID（不区分大小写）
//   path - 可执行文件完整路径
//   返回 - 匹配的PID列表
//   返回 - 错误信息
func FindByPath(path string) ([]uint32, error) {
	procs, err := List()
	if err != nil {
		return nil, err
	}
	target := strings.ToUpper(path)
	var matches []uint32
	for _, p := range procs {
		pid := p.ProcessID
		if pid == 0 || pid == 4 {
			continue
		}
		pPath, err := Path(pid)
		if err != nil {
			continue
		}
		if strings.ToUpper(pPath) == target {
			matches = append(matches, pid)
		}
	}
	return matches, nil
}

// OpenToken opens the process token for the given PID.
func openToken(pid uint32) (windows.Token, error) {
	var token windows.Token
	h, err := openWithMinimalAccess(pid)
	if err != nil {
		return 0, err
	}
	defer windows.CloseHandle(h)

	err = windows.OpenProcessToken(h, windows.TOKEN_QUERY, &token)
	if err != nil {
		return 0, fmt.Errorf("OpenProcessToken failed: %w", err)
	}
	return token, nil
}
