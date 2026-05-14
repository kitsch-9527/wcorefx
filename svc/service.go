//go:build windows

// Package svc 提供 Windows 服务枚举和信息查询功能。
package svc

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	serviceStopped         = 1
	serviceStartPending    = 2
	serviceStopPending     = 3
	serviceRunning         = 4
	serviceContinuePending = 5
	servicePausePending    = 6
	servicePaused          = 7

	serviceTypeWin32 = 0x00000030

	startBoot     = 0
	startSystem   = 1
	startAuto     = 2
	startManual   = 3
	startDisabled = 4

	serviceStateAll    = 3
	scEnumProcessInfo  = 0
	scManagerAllAccess = 0xF003F

	serviceQueryConfig = 0x00001
	serviceQueryStatus = 0x00004
)

// ServiceInfo 表示 Windows 服务信息。
type ServiceInfo struct {
	Name        string
	DisplayName string
	Status      string
	PID         uint32
	StartType   string
	Account     string
}

// statusString 将服务状态常量转换为可读字符串。
func statusString(s uint32) string {
	switch s {
	case serviceStopped:
		return "Stopped"
	case serviceStartPending:
		return "StartPending"
	case serviceStopPending:
		return "StopPending"
	case serviceRunning:
		return "Running"
	case serviceContinuePending:
		return "ContinuePending"
	case servicePausePending:
		return "PausePending"
	case servicePaused:
		return "Paused"
	default:
		return fmt.Sprintf("Unknown(%d)", s)
	}
}

// startTypeString 将启动类型常量转换为可读字符串。
func startTypeString(s uint32) string {
	switch s {
	case startBoot:
		return "Boot"
	case startSystem:
		return "System"
	case startAuto:
		return "Auto"
	case startManual:
		return "Manual"
	case startDisabled:
		return "Disabled"
	default:
		return fmt.Sprintf("Unknown(%d)", s)
	}
}

// openSCM 打开服务控制管理器。
func openSCM() (windows.Handle, error) {
	h, err := openSCManager(nil, nil, scManagerAllAccess)
	if err != nil {
		return 0, err
	}
	return h, nil
}

// List 返回所有正在运行的服务列表。
func List() ([]ServiceInfo, error) {
	scm, err := openSCM()
	if err != nil {
		return nil, fmt.Errorf("OpenSCManager failed: %w", err)
	}
	defer closeServiceHandle(scm)

	var bytesNeeded, servicesReturned, resumeHandle uint32

	err = enumServicesStatusEx(scm, scEnumProcessInfo, serviceTypeWin32, serviceStateAll,
		nil, 0, &bytesNeeded, &servicesReturned, &resumeHandle, "")
	if err != windows.ERROR_MORE_DATA {
		return nil, fmt.Errorf("EnumServicesStatusEx size query failed: %w", err)
	}

	buf := make([]byte, bytesNeeded)
	err = enumServicesStatusEx(scm, scEnumProcessInfo, serviceTypeWin32, serviceStateAll,
		&buf[0], uint32(len(buf)), &bytesNeeded, &servicesReturned, &resumeHandle, "")
	if err != nil {
		return nil, fmt.Errorf("EnumServicesStatusEx failed: %w", err)
	}

	entries := unsafe.Slice((*enumServiceStatusProcess)(unsafe.Pointer(&buf[0])), servicesReturned)
	services := make([]ServiceInfo, 0, servicesReturned)

	for _, e := range entries {
		si := ServiceInfo{
			Name:        windows.UTF16PtrToString(e.ServiceName),
			DisplayName: windows.UTF16PtrToString(e.DisplayName),
			Status:      statusString(e.CurrentState),
			PID:         e.ProcessID,
		}

		h, err := openService(scm, e.ServiceName, serviceQueryConfig)
		if err == nil {
			config, cerr := queryConfig(h)
			if cerr == nil {
				si.StartType = startTypeString(config.StartType)
				if config.ServiceStartName != nil {
					si.Account = windows.UTF16PtrToString(config.ServiceStartName)
				}
			}
			closeServiceHandle(h)
		}

		services = append(services, si)
	}

	return services, nil
}

// Status 返回指定服务的当前状态。
func Status(name string) (ServiceInfo, error) {
	scm, err := openSCM()
	if err != nil {
		return ServiceInfo{}, fmt.Errorf("OpenSCManager failed: %w", err)
	}
	defer closeServiceHandle(scm)

	svcHandle, err := openService(scm, windows.StringToUTF16Ptr(name), serviceQueryStatus)
	if err != nil {
		return ServiceInfo{}, fmt.Errorf("OpenService failed: %w", err)
	}
	defer closeServiceHandle(svcHandle)

	var status serviceStatusProcess
	err = queryServiceStatusEx(svcHandle, &status)
	if err != nil {
		return ServiceInfo{}, fmt.Errorf("QueryServiceStatusEx failed: %w", err)
	}

	si := ServiceInfo{
		Name:   name,
		Status: statusString(status.CurrentState),
		PID:    status.ProcessID,
	}

	return si, nil
}

// Config 返回指定服务的配置信息。
func Config(name string) (ServiceInfo, error) {
	scm, err := openSCM()
	if err != nil {
		return ServiceInfo{}, fmt.Errorf("OpenSCManager failed: %w", err)
	}
	defer closeServiceHandle(scm)

	svcHandle, err := openService(scm, windows.StringToUTF16Ptr(name), serviceQueryConfig)
	if err != nil {
		return ServiceInfo{}, fmt.Errorf("OpenService failed: %w", err)
	}
	defer closeServiceHandle(svcHandle)

	config, err := queryConfig(svcHandle)
	if err != nil {
		return ServiceInfo{}, fmt.Errorf("QueryServiceConfig failed: %w", err)
	}

	si := ServiceInfo{
		Name:      name,
		StartType: startTypeString(config.StartType),
	}
	if config.ServiceStartName != nil {
		si.Account = windows.UTF16PtrToString(config.ServiceStartName)
	}

	return si, nil
}

func queryConfig(h windows.Handle) (*queryServiceConfigW, error) {
	var bytesNeeded uint32
	procQueryServiceConfigW.Call(
		uintptr(h),
		0,
		0,
		uintptr(unsafe.Pointer(&bytesNeeded)),
	)

	buf := make([]byte, bytesNeeded)
	err := procQueryServiceConfigW.Call(
		uintptr(h),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(bytesNeeded),
		uintptr(unsafe.Pointer(&bytesNeeded)),
	)
	if err != nil {
		return nil, fmt.Errorf("QueryServiceConfigW failed: %w", err)
	}
	return (*queryServiceConfigW)(unsafe.Pointer(&buf[0])), nil
}
