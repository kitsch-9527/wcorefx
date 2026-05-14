//go:build windows

package netapi

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/kitsch-9527/wcorefx/internal/winapi"
)

var (
	procFwpmEngineOpen               = winapi.NewProc("fwpuclnt.dll", "FwpmEngineOpen0", winapi.ConvErrnoReturn)
	procFwpmEngineClose              = winapi.NewProc("fwpuclnt.dll", "FwpmEngineClose0", winapi.ConvErrnoReturn)
	procFwpmFreeMemory               = winapi.NewProc("fwpuclnt.dll", "FwpmFreeMemory0", winapi.ConvErrnoReturn)
	procFwpmCalloutCreateEnumHandle  = winapi.NewProc("fwpuclnt.dll", "FwpmCalloutCreateEnumHandle0", winapi.ConvErrnoReturn)
	procFwpmCalloutDestroyEnumHandle = winapi.NewProc("fwpuclnt.dll", "FwpmCalloutDestroyEnumHandle0", winapi.ConvErrnoReturn)
	procFwpmCalloutEnum              = winapi.NewProc("fwpuclnt.dll", "FwpmCalloutEnum0", winapi.ConvErrnoReturn)
	procFwpmCalloutGetByKey          = winapi.NewProc("fwpuclnt.dll", "FwpmCalloutGetByKey0", winapi.ConvErrnoReturn)
	procFwpmFilterCreateEnumHandle   = winapi.NewProc("fwpuclnt.dll", "FwpmFilterCreateEnumHandle0", winapi.ConvErrnoReturn)
	procFwpmFilterDestroyEnumHandle  = winapi.NewProc("fwpuclnt.dll", "FwpmFilterDestroyEnumHandle0", winapi.ConvErrnoReturn)
	procFwpmFilterEnum               = winapi.NewProc("fwpuclnt.dll", "FwpmFilterEnum0", winapi.ConvErrnoReturn)
	procFwpmFilterGetByKey           = winapi.NewProc("fwpuclnt.dll", "FwpmFilterGetByKey0", winapi.ConvErrnoReturn)
)

// FwpmDisplayData0 对应 FWPM_DISPLAY_DATA0 结构。
type FwpmDisplayData0 struct {
	Name        *uint16
	Description *uint16
}

// FwpmSession0 对应 FWPM_SESSION0 结构。
type FwpmSession0 struct {
	SessionKey           windows.GUID
	DisplayData          FwpmDisplayData0
	Flags                uint32
	TxnWaitTimeoutMillis uint32
	ProcessID            uint32
	SID                  *windows.SID
	Username             *uint16
	KernelMode           uint8
}

// FwpByteBlob 对应 FWP_BYTE_BLOB 结构。
type FwpByteBlob struct {
	Size uint32
	Data *uint8
}

// LayerID 标识 WFP 层。
type LayerID windows.GUID

// SublayerID 标识 WFP 子层。
type SublayerID windows.GUID

// FieldID 标识 WFP 层字段。
type FieldID windows.GUID

// RuleID 标识 WFP 规则。
type RuleID windows.GUID

// ProviderID 标识 WFP 提供商。
type ProviderID windows.GUID

// FwpmFilterFlags 定义 WFP 过滤器标志位。
type FwpmFilterFlags uint32

const (
	FwpmFilterFlagsPersistent              FwpmFilterFlags = 1 << iota
	FwpmFilterFlagsBootTime
	FwpmFilterFlagsHasProviderContext
	FwpmFilterFlagsClearActionRight
	FwpmFilterFlagsPermitIfCalloutUnregistered
	FwpmFilterFlagsDisabled
	FwpmFilterFlagsIndexed
)

// Action 是过滤引擎可执行的动作类型。
type Action uint32

const (
	ActionBlock              Action = 0x1001
	ActionPermit             Action = 0x1002
	ActionCalloutTerminating Action = 0x5003
	ActionCalloutInspection  Action = 0x6004
	ActionCalloutUnknown     Action = 0x4005
)

// FwpmAction0 对应 FWPM_ACTION0 结构。
type FwpmAction0 struct {
	Type Action
	GUID windows.GUID
}

// DataType 描述 FwpValue0 中的数据类型。
type DataType uint32

const (
	DataTypeEmpty                  DataType = 0
	DataTypeUint8                  DataType = 1
	DataTypeUint16                 DataType = 2
	DataTypeUint32                 DataType = 3
	DataTypeUint64                 DataType = 4
	DataTypeByteArray16            DataType = 11
	DataTypeByteBlob               DataType = 12
	DataTypeSID                    DataType = 13
	DataTypeSecurityDescriptor     DataType = 14
	DataTypeTokenInformation       DataType = 15
	DataTypeTokenAccessInformation DataType = 16
	DataTypeV4AddrMask             DataType = 256
	DataTypeV6AddrMask             DataType = 257
	DataTypeRange                  DataType = 258
)

// FwpValue0 对应 FWP_VALUE0 结构。
type FwpValue0 struct {
	Type  DataType
	Value uintptr
}

// MatchType 是用于字段测试的匹配运算符。
type MatchType uint32

const (
	MatchTypeEqual MatchType = iota
	MatchTypeGreater
	MatchTypeLess
	MatchTypeGreaterOrEqual
	MatchTypeLessOrEqual
	MatchTypeRange
	MatchTypeFlagsAllSet
	MatchTypeFlagsAnySet
	MatchTypeFlagsNoneSet
	MatchTypeEqualCaseInsensitive
	MatchTypeNotEqual
	MatchTypePrefix
	MatchTypeNotPrefix
)

// String 将 MatchType 运算符转换为字符串表示。
func (m MatchType) String() string {
	switch m {
	case MatchTypeEqual:
		return "=="
	case MatchTypeGreater:
		return ">"
	case MatchTypeLess:
		return "<"
	case MatchTypeGreaterOrEqual:
		return ">="
	case MatchTypeLessOrEqual:
		return "<="
	case MatchTypeRange:
		return "in"
	case MatchTypeFlagsAllSet:
		return "F[all]"
	case MatchTypeFlagsAnySet:
		return "F[any]"
	case MatchTypeFlagsNoneSet:
		return "F[none]"
	case MatchTypeEqualCaseInsensitive:
		return "i=="
	case MatchTypeNotEqual:
		return "!="
	case MatchTypePrefix:
		return "pfx"
	case MatchTypeNotPrefix:
		return "!pfx"
	default:
		return fmt.Sprintf("MatchType(%d)", m)
	}
}

// Match 是针对 WFP 层字段的匹配测试。
type Match struct {
	Field FieldID
	Op    MatchType
	Value any
}

// String 将 Match 格式化为可读的字符串描述。
func (m Match) String() string {
	return fmt.Sprintf("%v %s %v (%T)", m.Field, m.Op, m.Value, m.Value)
}

// FwpmFilterCondition0 对应 FWPM_FILTER_CONDITION0 结构。
type FwpmFilterCondition0 struct {
	FieldKey  FieldID
	MatchType MatchType
	Value     struct {
		Type  DataType
		Value uintptr
	}
}

// FwpmCallout0 对应 FWPM_CALLOUT0 结构。
type FwpmCallout0 struct {
	CalloutKey   windows.GUID
	DisplayData  FwpmDisplayData0
	Flags        FwpmFilterFlags
	ProviderKey  *windows.GUID
	ProviderData FwpByteBlob
	LayerKey     LayerID
	CalloutId    uint32
}

// FwpmFilter0 对应 FWPM_FILTER0 结构。
type FwpmFilter0 struct {
	FilterKey           windows.GUID
	DisplayData         FwpmDisplayData0
	Flags               FwpmFilterFlags
	ProviderKey         *windows.GUID
	ProviderData        FwpByteBlob
	LayerKey            LayerID
	SublayerKey         SublayerID
	Weight              FwpValue0
	NumFilterConditions uint32
	FilterConditions    *FwpmFilterCondition0
	Action              FwpmAction0
	RawContext          uint64
	ProviderContextKey  windows.GUID
	Reserved            *windows.GUID
	FilterID            uint64
	EffectiveWeight     FwpValue0
}

// FwpmEngineOpen 打开与 WFP 引擎的连接。
func FwpmEngineOpen(serverName *uint16, authnService uint32, authIdentity *byte, session *FwpmSession0) (windows.Handle, error) {
	var engineHandle windows.Handle
	_, err := procFwpmEngineOpen.CallRet(
		uintptr(unsafe.Pointer(serverName)),
		uintptr(authnService),
		uintptr(unsafe.Pointer(authIdentity)),
		uintptr(unsafe.Pointer(session)),
		uintptr(unsafe.Pointer(&engineHandle)),
	)
	return engineHandle, err
}

// FwpmEngineClose 关闭与 WFP 引擎的连接。
func FwpmEngineClose(engineHandle windows.Handle) error {
	return procFwpmEngineClose.Call(uintptr(engineHandle))
}

// FwpmFreeMemory 释放 WFP 分配的内存。
func FwpmFreeMemory(p unsafe.Pointer) {
	procFwpmFreeMemory.Call(uintptr(p))
}

// FwpmCalloutCreateEnumHandle 创建 WFP 标注枚举句柄。
func FwpmCalloutCreateEnumHandle(engineHandle windows.Handle, enumTemplate *byte) (windows.Handle, error) {
	var enumHandle windows.Handle
	_, err := procFwpmCalloutCreateEnumHandle.CallRet(
		uintptr(engineHandle),
		uintptr(unsafe.Pointer(enumTemplate)),
		uintptr(unsafe.Pointer(&enumHandle)),
	)
	return enumHandle, err
}

// FwpmCalloutDestroyEnumHandle 销毁 WFP 标注枚举句柄。
func FwpmCalloutDestroyEnumHandle(engineHandle, enumHandle windows.Handle) error {
	return procFwpmCalloutDestroyEnumHandle.Call(uintptr(engineHandle), uintptr(enumHandle))
}

// FwpmCalloutEnum 枚举 WFP 标注。
func FwpmCalloutEnum(engineHandle, enumHandle windows.Handle, numEntries uint32, entries ***FwpmCallout0, numReturned *uint32) error {
	return procFwpmCalloutEnum.Call(
		uintptr(engineHandle),
		uintptr(enumHandle),
		uintptr(numEntries),
		uintptr(unsafe.Pointer(entries)),
		uintptr(unsafe.Pointer(numReturned)),
	)
}

// FwpmCalloutGetByKey 通过 GUID 键获取 WFP 标注。
func FwpmCalloutGetByKey(engineHandle windows.Handle, key *windows.GUID, callout ***FwpmCallout0) error {
	return procFwpmCalloutGetByKey.Call(
		uintptr(engineHandle),
		uintptr(unsafe.Pointer(key)),
		uintptr(unsafe.Pointer(callout)),
	)
}

// FwpmFilterCreateEnumHandle 创建 WFP 过滤器枚举句柄。
func FwpmFilterCreateEnumHandle(engineHandle windows.Handle, enumTemplate *byte) (windows.Handle, error) {
	var enumHandle windows.Handle
	_, err := procFwpmFilterCreateEnumHandle.CallRet(
		uintptr(engineHandle),
		uintptr(unsafe.Pointer(enumTemplate)),
		uintptr(unsafe.Pointer(&enumHandle)),
	)
	return enumHandle, err
}

// FwpmFilterDestroyEnumHandle 销毁 WFP 过滤器枚举句柄。
func FwpmFilterDestroyEnumHandle(engineHandle, enumHandle windows.Handle) error {
	return procFwpmFilterDestroyEnumHandle.Call(uintptr(engineHandle), uintptr(enumHandle))
}

// FwpmFilterEnum 枚举 WFP 过滤器。
func FwpmFilterEnum(engineHandle, enumHandle windows.Handle, numEntries uint32, entries ***FwpmFilter0, numReturned *uint32) error {
	return procFwpmFilterEnum.Call(
		uintptr(engineHandle),
		uintptr(enumHandle),
		uintptr(numEntries),
		uintptr(unsafe.Pointer(entries)),
		uintptr(unsafe.Pointer(numReturned)),
	)
}

// FwpmFilterGetByKey 通过 GUID 键获取 WFP 过滤器。
func FwpmFilterGetByKey(engineHandle windows.Handle, key *windows.GUID, filter ***FwpmFilter0) error {
	return procFwpmFilterGetByKey.Call(
		uintptr(engineHandle),
		uintptr(unsafe.Pointer(key)),
		uintptr(unsafe.Pointer(filter)),
	)
}
