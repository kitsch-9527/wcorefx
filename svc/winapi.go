//go:build windows

package svc

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var modadvapi32 = windows.NewLazySystemDLL("Advapi32.dll")

var (
	procOpenSCManager        = modadvapi32.NewProc("OpenSCManagerW")
	procCloseServiceHandle   = modadvapi32.NewProc("CloseServiceHandle")
	procEnumServicesStatusEx = modadvapi32.NewProc("EnumServicesStatusExW")
	procQueryServiceStatusEx = modadvapi32.NewProc("QueryServiceStatusEx")
	procQueryServiceConfigW  = modadvapi32.NewProc("QueryServiceConfigW")
	procOpenService          = modadvapi32.NewProc("OpenServiceW")
)

// enumServiceStatusProcess 对应 Windows ENUM_SERVICE_STATUS_PROCESSW 结构体。
type enumServiceStatusProcess struct {
	ServiceName      *uint16
	DisplayName      *uint16
	ServiceType      uint32
	CurrentState     uint32
	ControlsAccepted uint32
	Win32ExitCode    uint32
	ServiceSpecific  uint32
	CheckPoint       uint32
	WaitHint         uint32
	ProcessID        uint32
	ServiceFlags     uint32
}

// serviceStatusProcess 对应 Windows SERVICE_STATUS_PROCESS 结构体。
type serviceStatusProcess struct {
	ServiceType      uint32
	CurrentState     uint32
	ControlsAccepted uint32
	Win32ExitCode    uint32
	ServiceSpecific  uint32
	CheckPoint       uint32
	WaitHint         uint32
	ProcessID        uint32
	ServiceFlags     uint32
}

// queryServiceConfigW 对应 Windows QUERY_SERVICE_CONFIGW 结构体。
type queryServiceConfigW struct {
	ServiceType      uint32
	StartType        uint32
	ErrorControl     uint32
	TagID            uint32
	BinaryPathName   *uint16
	LoadOrderGroup   *uint16
	Dependencies     *uint16
	ServiceStartName *uint16
	DisplayName      *uint16
}

// openSCManager 打开服务控制管理器。
//   machineName  - 目标机器名称（nil 表示本地机器）
//   databaseName - 数据库名称（nil 表示默认数据库）
//   desiredAccess - 访问权限
//   返回 - 服务控制管理器句柄
//   返回 - 错误信息
func openSCManager(machineName, databaseName *uint16, desiredAccess uint32) (windows.Handle, error) {
	r1, _, err := syscall.SyscallN(procOpenSCManager.Addr(),
		uintptr(unsafe.Pointer(machineName)),
		uintptr(unsafe.Pointer(databaseName)),
		uintptr(desiredAccess),
	)
	if r1 == 0 {
		return 0, err
	}
	return windows.Handle(r1), nil
}

// closeServiceHandle 关闭服务句柄。
func closeServiceHandle(h windows.Handle) error {
	r1, _, err := syscall.SyscallN(procCloseServiceHandle.Addr(),
		uintptr(h),
	)
	if r1 == 0 {
		return err
	}
	return nil
}

// openService 打开指定服务。
func openService(scm windows.Handle, serviceName *uint16, desiredAccess uint32) (windows.Handle, error) {
	r1, _, err := syscall.SyscallN(procOpenService.Addr(),
		uintptr(scm),
		uintptr(unsafe.Pointer(serviceName)),
		uintptr(desiredAccess),
	)
	if r1 == 0 {
		return 0, err
	}
	return windows.Handle(r1), nil
}

// enumServicesStatusEx 枚举服务状态。
func enumServicesStatusEx(scm windows.Handle, infoLevel, serviceType, serviceState uint32,
	services *byte, bufSize uint32, bytesNeeded, servicesReturned, resumeHandle *uint32,
	groupName string) error {

	var groupNamePtr *uint16
	if groupName != "" {
		groupNamePtr, _ = windows.UTF16PtrFromString(groupName)
	}

	r1, _, err := syscall.SyscallN(procEnumServicesStatusEx.Addr(),
		uintptr(scm),
		uintptr(infoLevel),
		uintptr(serviceType),
		uintptr(serviceState),
		uintptr(unsafe.Pointer(services)),
		uintptr(bufSize),
		uintptr(unsafe.Pointer(bytesNeeded)),
		uintptr(unsafe.Pointer(servicesReturned)),
		uintptr(unsafe.Pointer(resumeHandle)),
		uintptr(unsafe.Pointer(groupNamePtr)),
	)
	if r1 == 0 {
		return err
	}
	return nil
}

// queryServiceStatusEx 查询服务状态。
func queryServiceStatusEx(h windows.Handle, status *serviceStatusProcess) error {
	var bytesNeeded uint32
	r1, _, err := syscall.SyscallN(procQueryServiceStatusEx.Addr(),
		uintptr(h),
		uintptr(0), // SC_STATUS_PROCESS_INFO
		uintptr(unsafe.Pointer(status)),
		uintptr(unsafe.Sizeof(*status)),
		uintptr(unsafe.Pointer(&bytesNeeded)),
	)
	if r1 == 0 {
		return err
	}
	return nil
}
