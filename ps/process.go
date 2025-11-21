//go:build windows
// +build windows

package ps

import (
	"fmt"
	"syscall"
	"unsafe"

	"github.com/emirpasic/gods/maps/treemap"
	"github.com/emirpasic/gods/utils"
	comm "github.com/kitsch-9527/wcorefx/common"
	"github.com/kitsch-9527/wcorefx/fs"
	se "github.com/kitsch-9527/wcorefx/sec"
	"github.com/kitsch-9527/wcorefx/winapi/dll/ntdll"
	"github.com/kitsch-9527/wcorefx/winapi/dll/psapi"
	"golang.org/x/sys/windows"
)

// todo 后续增加进程map 缓存
// EnumProcessMap 枚举所有进程返回treeset map
// key为进程id，值为进程信息结构体 windows.ProcessEntry32
func GetMap() (*treemap.Map, error) {
	tree := treemap.NewWith(utils.UInt32Comparator)
	h, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return tree, fmt.Errorf("CreateToolhelp32Snapshot failed: %w", err)
	}
	defer windows.CloseHandle(h)
	var pe windows.ProcessEntry32
	pe.Size = uint32(unsafe.Sizeof(pe))
	err = windows.Process32First(h, &pe)
	if err != nil {
		return tree, fmt.Errorf("Process32First failed: %w", err)
	}
	for {
		tree.Put(pe.ProcessID, pe)
		err = windows.Process32Next(h, &pe)
		if err != nil {
			break
		}
	}
	return tree, nil
}

// EnumProcess 枚举所有进程返回进程列表
func GetList() ([]windows.ProcessEntry32, error) {
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

type ProcessTimes struct {
	CreationTime windows.Filetime
	ExitTime     windows.Filetime
	KernelTime   windows.Filetime
	UserTime     windows.Filetime
}

// IsDeleted 判断指定进程是否已被删除
// false表示进程存在，true表示进程已被删除
func IsTerminated(pid uint32) bool {
	result := false
	h, err := OpenWithMinimalAccess(pid)
	defer windows.CloseHandle(h)
	if err != nil {
		result = true
		return result

	} else {
		event, err := windows.WaitForSingleObject(h, 0)
		if err != nil {
			if event == windows.WAIT_OBJECT_0 {
				result = true
				return result
			}
		}
	}
	return result
}

// 根据进程ID获取会话ID
func GetSessionID(pid uint32) (uint32, error) {
	var sessionid uint32
	err := windows.ProcessIdToSessionId(pid, &sessionid)
	if err != nil {
		return 0, fmt.Errorf("ProcessIdToSessionId failed: %w", err)
	}
	return sessionid, nil
}

// GetProcMemoryInfo 获取指定进程的内存信息
func GetMemoryInfo(pid uint32) (psapi.PROCESS_MEMORY_COUNTERS, error) {
	var mem psapi.PROCESS_MEMORY_COUNTERS
	c, err := OpenWithMinimalAccess(pid)
	if err != nil {
		return mem, err
	}
	defer windows.CloseHandle(c)
	if err := psapi.GetProcessMemoryInfo(c, &mem); err != nil {
		return mem, fmt.Errorf("getProcessMemoryInfo failed: %w", err)
	}

	return mem, err
}

// GetCreateTime 获取指定进程的创建时间
func GetTimes(pid uint32) (ProcessTimes, error) {
	h, err := OpenWithMinimalAccess(pid)
	if err != nil {
		return ProcessTimes{}, err
	}
	defer windows.CloseHandle(h)
	var creationTime windows.Filetime
	var exitTime windows.Filetime
	var kernelTime windows.Filetime
	var userTime windows.Filetime
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

// GetParentPid 获取指定进程的父进程id
func GetParentID(pid uint32) uint32 {
	h, err := OpenWithMinimalAccess(pid)
	if err != nil {
		return 0
	}
	defer windows.CloseHandle(h)
	pbi, err := GetBasicInfo(h)
	if err != nil {
		return 0
	}
	return uint32(pbi.InheritedFromUniqueProcessId)
}

// OpenWithMinimalAccess 打开指定进程的句柄，最小权限
func OpenWithMinimalAccess(pid uint32) (windows.Handle, error) {
	handle, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, pid)
	if err != nil {
		return 0, fmt.Errorf("failed to open process limitied handle: %w", err)
	}
	return handle, nil
}

// OpenProcessWithFullAccess 打开指定进程的句柄
func OpenWithFullAccess(processId uint32) (windows.Handle, error) {
	handle, err := windows.OpenProcess(windows.PROCESS_ALL_ACCESS, true, processId)
	if err != nil {
		return 0, fmt.Errorf("OpenProcess failed: %w", err)
	}
	return handle, nil
}

// GetProcessBasicInfo 获取指定进程的基本信息
func GetBasicInfo(h windows.Handle) (windows.PROCESS_BASIC_INFORMATION, error) {
	var retlen uint32
	pe := windows.PROCESS_BASIC_INFORMATION{}
	err := windows.NtQueryInformationProcess(h, windows.ProcessBasicInformation, unsafe.Pointer(&pe), uint32(unsafe.Sizeof(pe)), &retlen)
	if err != nil {
		return windows.PROCESS_BASIC_INFORMATION{}, fmt.Errorf("NtQueryInformationProcess failed : %w", err)
	}
	return pe, nil
}

func GetCommandLine(pid uint32) (string, error) {
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_INFORMATION|windows.PROCESS_VM_READ, false, pid)
	if err != nil {
		return "", fmt.Errorf("OpenProcess failed: %w", err)
	}
	defer windows.CloseHandle(h)
	basicInfo, err := GetBasicInfo(h)
	if err != nil {
		return "", err
	}
	pebAddr := basicInfo.PebBaseAddress
	if pebAddr == nil {
		return "", fmt.Errorf("PEB地址无效")
	}

	// 2. 读取目标进程中的PEB结构（PEB位于目标进程地址空间，需通过ReadProcessMemory读取）
	// 2. 读取PEB结构
	var peb windows.PEB
	var bytesRead uintptr
	err = windows.ReadProcessMemory(
		h,
		uintptr(unsafe.Pointer(pebAddr)),
		(*byte)(unsafe.Pointer(&peb)), // 转换为*byte类型的缓冲区指针
		uintptr(unsafe.Sizeof(peb)),   // 要读取的大小
		&bytesRead,                    // 接收实际读取的字节数
	)
	if err != nil || bytesRead != uintptr(unsafe.Sizeof(peb)) {
		return "", fmt.Errorf("读取PEB失败: %w (预期读取 %d, 实际读取 %d)",
			err, unsafe.Sizeof(peb), bytesRead)
	}

	// 3. 获取ProcessParameters地址
	procParamsAddr := peb.ProcessParameters
	if procParamsAddr == nil {
		return "", fmt.Errorf("ProcessParameters地址无效")
	}

	// 4. 读取RTL_USER_PROCESS_PARAMETERS结构
	var procParams windows.RTL_USER_PROCESS_PARAMETERS
	err = windows.ReadProcessMemory(
		h,
		uintptr(unsafe.Pointer(procParamsAddr)),
		(*byte)(unsafe.Pointer(&procParams)),
		uintptr(unsafe.Sizeof(procParams)),
		&bytesRead,
	)
	if err != nil || bytesRead != uintptr(unsafe.Sizeof(procParams)) {
		return "", fmt.Errorf("读取ProcessParameters失败: %w (预期读取 %d, 实际读取 %d)",
			err, unsafe.Sizeof(procParams), bytesRead)
	}

	// 5. 处理命令行信息
	cmdLine := procParams.CommandLine
	if cmdLine.Buffer == nil || cmdLine.Length == 0 {
		return "", nil // 命令行为空
	}

	// 6. 读取命令行内容（UTF-16编码）
	bufSize := uintptr(cmdLine.Length) // 字节数
	buf := make([]uint16, bufSize/2)   // 转换为uint16数组
	err = windows.ReadProcessMemory(
		h,
		uintptr(unsafe.Pointer(cmdLine.Buffer)),
		(*byte)(unsafe.Pointer(&buf[0])), // 缓冲区首地址转换为*byte
		bufSize,
		&bytesRead,
	)
	if err != nil || bytesRead != bufSize {
		return "", fmt.Errorf("读取命令行内容失败: %w (预期读取 %d, 实际读取 %d)",
			err, bufSize, bytesRead)
	}
	// 7. 转换为Go字符串
	return windows.UTF16ToString(buf), nil
}

// OpenProcessToken 打开指定进程的令牌
func OpenToken(pid uint32) (windows.Token, error) {
	var token windows.Token
	h, err := OpenWithMinimalAccess(pid)
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

// GetProcUserName 获取指定进程的用户名
func GetUser(pid uint32) (string, string, error) {
	token, err := OpenToken(pid)
	if err != nil {
		return "", "", err
	}
	var size uint32
	err = windows.GetTokenInformation(token, windows.TokenUser, nil, 0, &size)
	if err != nil {
		if err != windows.ERROR_INSUFFICIENT_BUFFER {
			return "", "", fmt.Errorf("GetTokenInformation failed: %w", err)
		}
	}
	// 分配缓冲区
	buffer := make([]byte, size)
	err = windows.GetTokenInformation(token, windows.TokenUser, &buffer[0], size, &size)
	if err != nil {
		return "", "", fmt.Errorf("GetTokenInformation failed: %w", err)
	}
	// 解析TOKEN_USER结构
	tokenUser := (*windows.Tokenuser)(unsafe.Pointer(&buffer[0]))
	return se.LookupSIDAccount(tokenUser.User.Sid)
}

// GetModuleFileName 获得指定模块的路径
func GetModuleFileName(hProc windows.Handle, hMod windows.Handle) (string, error) {
	var path [comm.MAXPATH + 1]uint16
	err := windows.GetModuleFileNameEx(hProc, hMod, &path[0], comm.MAXPATH)
	if err != nil {
		return "", fmt.Errorf("GetModuleFileNameEx failed: %w", err)
	}
	return windows.UTF16ToString(path[:]), nil
}

// EnumProcessModules 枚举指定进程加载的模块
func EnumProcessModules(handle windows.Handle) ([]windows.Handle, error) {
	initialSize := 10
	modules := make([]windows.Handle, initialSize)
	var bytesNeeded uint32
	err := windows.EnumProcessModulesEx(
		windows.Handle(handle),
		&modules[0],
		uint32(len(modules))*comm.PtrSize,
		&bytesNeeded,
		windows.LIST_MODULES_ALL, // 枚举所有32/64位模块
	)
	if err != nil {
		// 如果是缓冲区不足错误，且bytesNeeded有效，继续处理
		if err != windows.ERROR_INSUFFICIENT_BUFFER {
			return nil, fmt.Errorf("first EnumProcessModulesEx failed: %w", err)
		}
	}
	neededCount := int(bytesNeeded / comm.PtrSize)
	if len(modules) < neededCount {
		modules := make([]windows.Handle, neededCount)
		err := windows.EnumProcessModulesEx(
			windows.Handle(handle),
			&modules[0],
			uint32(len(modules))*comm.PtrSize,
			&bytesNeeded,
			windows.LIST_MODULES_ALL, // 枚举所有32/64位模块
		)
		if err != nil {
			return nil, fmt.Errorf("second EnumProcessModulesEx failed: %w", err)
		}
		return modules, nil
	} else {
		modules = modules[:neededCount]
		return modules, nil
	}
}

// PsGetProcessPath 获取指定进程的路径
func GetPath(pid uint32) (string, error) {
	if pid == 0 {
		return "System Idle Process", nil
	}
	if pid == 4 {
		return "System", nil
	}
	var handle windows.Handle
	handle, err := OpenWithMinimalAccess(pid)
	if err != nil {
		return "", err
	}
	defer windows.CloseHandle(handle)
	path, err := GetModuleFileName(handle, 0)
	if err != nil {
		size := uint32(comm.MAXPATH + 1)
		var path [comm.MAXPATH + 1]uint16
		windows.QueryFullProcessImageName(handle, windows.PROCESS_NAME_NATIVE, &path[0], &size)
		nativePath := windows.UTF16ToString(path[:size])
		// 映射 native 路径到盘符路径
		mappedPath, mapErr := fs.NativePathToDosPath(nativePath)
		if mapErr != nil {
			return "", mapErr
		}
		return mappedPath, nil
	}
	return path, nil
}

// func GetOpenFiles(pid uint32) error {
// 	var (
// 		buffer       []byte
// 		returnLength uint32
// 	)
// 	// 初始缓冲区（1MB），第一次调用获取所需大小
// 	buffer = make([]byte, 1024*1024)
// 	err := windows.NtQuerySystemInformation(windows.SystemHandleInformation,
// 		unsafe.Pointer(&buffer[0]),
// 		uint32(len(buffer)),
// 		&returnLength)
// 	if err == windows.STATUS_INFO_LENGTH_MISMATCH {
// 		returnLength += returnLength * 4
// 		buffer = make([]byte, returnLength)
// 		err = windows.NtQuerySystemInformation(windows.SystemHandleInformation,
// 			unsafe.Pointer(&buffer[0]),
// 			returnLength,
// 			&returnLength)
// 	}
// 	if err != nil {
// 		return fmt.Errorf("NtQuerySystemInformation failed: %w", err)
// 	}
// 	handleTable := (*ntdll.PSystemHandleInformation)(unsafe.Pointer(&buffer[0]))
// 	for i := uint32(0); i < handleTable.NumberOfHandles; i++ {
// 		// 计算当前句柄条目的地址
// 		entryAddr := uintptr(unsafe.Pointer(&handleTable.Handles)) + uintptr(i)*unsafe.Sizeof(ntdll.SystemHandleTableEntryInfo{})
// 		handleInfo := *(*ntdll.SystemHandleTableEntryInfo)(unsafe.Pointer(entryAddr))
// 		if handleInfo.UniqueProcessId == uint16(pid) {
// 			//fmt.Println(handleInfo)
// 			h, err := duplicateProcessHandle(uint32(handleInfo.UniqueProcessId), windows.Handle(handleInfo.HandleValue))
// 			if err != nil {
// 				//	fmt.Println("duplicateAnotherProcessHandle error:", err)
// 			} else {
// 				t, err := GetHandleType(h)
// 				if err != nil {
// 					fmt.Println("GetHandleType error:", err)
// 				}
// 				if t == "File" {
// 					n, err := GetHandleName(h)
// 					if err != nil {
// 						fmt.Println("GetHandleName error:", err)
// 					}
// 					dosPath, err := fs.NativePathToDosPath(n)
// 					if err == nil {
// 						n = dosPath
// 					}
// 					fmt.Println(t, ":", n)
// 				}
// 			}
// 		}
// 	}
// 	return nil
// }
func duplicateProcessHandle(pid uint32, hSource windows.Handle) (windows.Handle, error) {
	hPreces, err := OpenWithFullAccess(pid)
	if err != nil {
		return 0, err
	}
	var b bool
	windows.IsWow64Process(hPreces, &b)
	if !b {

	}
	hCurrent, err := windows.GetCurrentProcess()
	if err != nil {
		return 0, fmt.Errorf("GetCurrentProcess failed: %w", err)
	}
	targetHandle, err := ntdll.NtDuplicateObject(hPreces, hSource, hCurrent, 0, 0, 0)
	if err != nil {
		return 0, fmt.Errorf("NtDuplicateObject failed: %w", err)
	}
	return targetHandle, nil
}

func GetHandleType(h windows.Handle) (string, error) {
	// 分配缓冲区
	bufferSize := uint32(0x1000)
	var returnLength uint32
	objectTypeInfo := make([]byte, bufferSize)
	err := ntdll.NtQueryObject(h, ntdll.ObjectTypeInformation, uintptr(unsafe.Pointer(&objectTypeInfo[0])), bufferSize, &returnLength)
	if err != nil {
		fmt.Println("NtQueryObject failed:", err)
		return "", err
	}
	// 转换缓冲区为PUBLIC_OBJECT_TYPE_INFORMATION结构体
	objTypeInfo := (*ntdll.PUBLIC_OBJECT_TYPE_INFORMATION)(unsafe.Pointer(&objectTypeInfo[0]))
	return syscall.UTF16ToString(
		unsafe.Slice(
			objTypeInfo.TypeName.Buffer,
			objTypeInfo.TypeName.Length/2, // Length是字节数，需转换为字符数
		),
	), nil
}

func GetHandleName(h windows.Handle) (string, error) {
	// 分配缓冲区
	bufferSize := uint32(0x1000)
	var returnLength uint32
	objectNameInfo := make([]byte, bufferSize)
	err := ntdll.NtQueryObject(h, ntdll.ObjectNameInformation, uintptr(unsafe.Pointer(&objectNameInfo[0])), bufferSize, &returnLength)
	if err == windows.STATUS_INFO_LENGTH_MISMATCH {
		return "", fmt.Errorf("NtQueryObject failed: %w", err)
	}
	objectNameInfo = make([]byte, returnLength)
	err = ntdll.NtQueryObject(h, ntdll.ObjectNameInformation, uintptr(unsafe.Pointer(&objectNameInfo[0])), bufferSize, &returnLength)
	if err != nil {
		return "", fmt.Errorf("NtQueryObject failed: %w", err)
	}
	// 转换缓冲区为OBJECT_NAME_INFORMATION结构体
	objNameInfo := (*ntdll.UNICODE_STRING)(unsafe.Pointer(&objectNameInfo[0]))
	return syscall.UTF16ToString(
		unsafe.Slice(
			objNameInfo.Buffer,
			objNameInfo.Length/2, // Length是字节数，需转换为字符数
		),
	), nil
}
