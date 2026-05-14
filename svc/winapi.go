//go:build windows

package svc

import (
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/kitsch-9527/wcorefx/internal/winapi"
)

var (
	procOpenSCManager        = winapi.NewProc("Advapi32.dll", "OpenSCManagerW")
	procCloseServiceHandle   = winapi.NewProc("Advapi32.dll", "CloseServiceHandle")
	procEnumServicesStatusEx = winapi.NewProc("Advapi32.dll", "EnumServicesStatusExW")
	procQueryServiceStatusEx = winapi.NewProc("Advapi32.dll", "QueryServiceStatusEx")
	procQueryServiceConfigW  = winapi.NewProc("Advapi32.dll", "QueryServiceConfigW")
	procOpenService          = winapi.NewProc("Advapi32.dll", "OpenServiceW")
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
func openSCManager(machineName, databaseName *uint16, desiredAccess uint32) (windows.Handle, error) {
	h, err := procOpenSCManager.CallRet(
		uintptr(unsafe.Pointer(machineName)),
		uintptr(unsafe.Pointer(databaseName)),
		uintptr(desiredAccess),
	)
	if h == 0 {
		return 0, err
	}
	return windows.Handle(h), nil
}

// closeServiceHandle 关闭服务句柄。
func closeServiceHandle(h windows.Handle) error {
	return procCloseServiceHandle.Call(uintptr(h))
}

// openService 打开指定服务。
func openService(scm windows.Handle, serviceName *uint16, desiredAccess uint32) (windows.Handle, error) {
	h, err := procOpenService.CallRet(
		uintptr(scm),
		uintptr(unsafe.Pointer(serviceName)),
		uintptr(desiredAccess),
	)
	if h == 0 {
		return 0, err
	}
	return windows.Handle(h), nil
}

// enumServicesStatusEx 枚举服务状态。
func enumServicesStatusEx(scm windows.Handle, infoLevel, serviceType, serviceState uint32,
	services *byte, bufSize uint32, bytesNeeded, servicesReturned, resumeHandle *uint32,
	groupName string) error {

	var groupNamePtr *uint16
	if groupName != "" {
		groupNamePtr, _ = windows.UTF16PtrFromString(groupName)
	}

	return procEnumServicesStatusEx.Call(
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
}

// queryServiceStatusEx 查询服务状态。
func queryServiceStatusEx(h windows.Handle, status *serviceStatusProcess) error {
	var bytesNeeded uint32
	return procQueryServiceStatusEx.Call(
		uintptr(h),
		uintptr(0), // SC_STATUS_PROCESS_INFO
		uintptr(unsafe.Pointer(status)),
		uintptr(unsafe.Sizeof(*status)),
		uintptr(unsafe.Pointer(&bytesNeeded)),
	)
}
