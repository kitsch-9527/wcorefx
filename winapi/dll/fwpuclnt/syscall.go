package fwpuclnt

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// FwpmEngineOpen 打开Windows防火墙策略引擎并返回句柄
// serverName: 远程服务器名称，本地使用nil
// authnService: 认证服务类型，如RPC_C_AUTHN_WINNT
// authIdentity: 认证身份信息，不需要时使用nil
// session: 会话参数，使用默认会话时为nil
// 返回值: 引擎句柄和可能的错误
func FwpmEngineOpen(
	serverName *uint16,
	authnService uint32,
	authIdentity *secWinntAuthIdentity,
	session *FwpmSession0,
) (windows.Handle, error) {
	var engineHandle windows.Handle
	r0, _, _ := syscall.Syscall6(
		procFwpmEngineOpen.Addr(),
		5,
		uintptr(unsafe.Pointer(serverName)),
		uintptr(authnService),
		uintptr(unsafe.Pointer(authIdentity)),
		uintptr(unsafe.Pointer(session)),
		uintptr(unsafe.Pointer(&engineHandle)),
		0, // 保留参数
	)

	if r0 != 0 {
		return engineHandle, syscall.Errno(r0)
	}

	return engineHandle, nil
}

func FwpmCalloutCreateEnumHandle(
	engineHandle windows.Handle,
	enumTemplate *fwpmCalloutEnumTemplate0,
) (windows.Handle, error) {
	var enumHandle windows.Handle
	r0, _, _ := syscall.Syscall(
		procFwpmCalloutCreateEnumHandle.Addr(),
		3,
		uintptr(engineHandle),
		uintptr(unsafe.Pointer(enumTemplate)),
		uintptr(unsafe.Pointer(&enumHandle)),
	)
	if r0 != 0 {
		return enumHandle, syscall.Errno(r0)
	}

	return enumHandle, nil
}

// FwpmCalloutEnum 枚举 WFP 标注
func FwpmCalloutEnum(
	engineHandle windows.Handle,
	enumHandle windows.Handle,
	numEntries uint32,
	entries ***FwpmCallout0,
	numEntriesReturned *uint32,
) (err error) {
	r0, _, _ := syscall.Syscall6(
		procFwpmCalloutEnum.Addr(),
		5,
		uintptr(engineHandle),
		uintptr(enumHandle),
		uintptr(numEntries),
		uintptr(unsafe.Pointer(entries)),
		uintptr(unsafe.Pointer(numEntriesReturned)),
		0, // 保留参数
	)
	if r0 != 0 {
		err = syscall.Errno(r0)
	}
	return
}

func FwpmFilterEnum(
	engineHandle windows.Handle,
	enumHandle windows.Handle,
	numEntries uint32,
	entries ***FwpmFilter0,
	numEntriesReturned *uint32,
) (err error) {
	r0, _, _ := syscall.Syscall6(
		procFwpmFilterEnum.Addr(),
		5,
		uintptr(engineHandle),
		uintptr(enumHandle),
		uintptr(numEntries),
		uintptr(unsafe.Pointer(entries)),
		uintptr(unsafe.Pointer(numEntriesReturned)),
		0, // 保留参数
	)
	if r0 != 0 {
		err = syscall.Errno(r0)
	}
	return
}

func FwpmFilterCreateEnumHandle(
	engineHandle windows.Handle,
	enumTemplate *fwpmFilterEnumTemplate0,
) (windows.Handle, error) {
	var enumHandle windows.Handle
	r0, _, _ := syscall.Syscall(
		procFwpmFilterCreateEnumHandle.Addr(),
		3,
		uintptr(engineHandle),
		uintptr(unsafe.Pointer(enumTemplate)),
		uintptr(unsafe.Pointer(&enumHandle)),
	)
	if r0 != 0 {
		return enumHandle, syscall.Errno(r0)
	}

	return enumHandle, nil
}
func FwpmFilterGetByKey(engineHandle windows.Handle,
	key *windows.GUID,
	filter **FwpmFilter0) error {

	r0, _, _ := syscall.Syscall(
		procFwpmFilterGetByKey.Addr(),
		3,
		uintptr(engineHandle),
		uintptr(unsafe.Pointer(key)),
		uintptr(unsafe.Pointer(filter)),
	)
	if r0 != 0 {
		return syscall.Errno(r0)
	}
	return nil
}
func FwpmCalloutGetByKey(engineHandle windows.Handle,
	key *windows.GUID,
	filter **FwpmCallout0) error {

	r0, _, _ := syscall.Syscall(
		procFwpmCalloutGetByKey.Addr(),
		3,
		uintptr(engineHandle),
		uintptr(unsafe.Pointer(key)),
		uintptr(unsafe.Pointer(filter)),
	)
	if r0 != 0 {
		return syscall.Errno(r0)
	}
	return nil
}

func FwpmCalloutDestroyEnumHandle(engineHandle, enumHandle windows.Handle) error {
	r0, _, _ := syscall.Syscall(
		procFwpmCalloutDestroyEnumHandle.Addr(),
		2,
		uintptr(engineHandle),
		uintptr(enumHandle),
		0, // 保留参数
	)
	if r0 != 0 {
		return syscall.Errno(r0)
	}
	return nil
}

func FwpmFilterDestroyEnumHandle(engineHandle, enumHandle windows.Handle) error {
	r0, _, _ := syscall.Syscall(
		procFwpmFilterDestroyEnumHandle.Addr(),
		2,
		uintptr(engineHandle),
		uintptr(enumHandle),
		0, // 保留参数
	)
	if r0 != 0 {
		return syscall.Errno(r0)
	}
	return nil
}

func FwpmEngineClose(engineHandle windows.Handle) error {
	r0, _, _ := syscall.Syscall(
		procFwpmEngineClose.Addr(),
		1,
		uintptr(engineHandle),
		0, // 保留参数
		0, // 保留参数
	)
	if r0 != 0 {
		return syscall.Errno(r0)
	}
	return nil
}

func FwpmFreeMemory(p *struct{}) {
	syscall.Syscall(procFwpmFreeMemory.Addr(), 1, uintptr(unsafe.Pointer(p)), 0, 0)
	return
}
