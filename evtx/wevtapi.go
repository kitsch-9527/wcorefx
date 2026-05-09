//go:build windows

package evtx

import (
	"fmt"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// EvtHandle 是事件日志API的句柄。
type EvtHandle uintptr

// NilHandle 表示一个空的EvtHandle值。
const NilHandle EvtHandle = 0

// Close 关闭EvtHandle句柄。
//   返回 - 关闭过程中的错误，成功时为nil
func (h EvtHandle) Close() error {
	return evtClose(h)
}

// Error codes.
const (
	ERROR_INVALID_HANDLE        syscall.Errno = 6
	ERROR_INSUFFICIENT_BUFFER   syscall.Errno = 122
	ERROR_NO_MORE_ITEMS         syscall.Errno = 259
	RPC_S_SERVER_UNAVAILABLE    syscall.Errno = 1722
	RPC_S_INVALID_BOUND         syscall.Errno = 1734
	ERROR_INVALID_OPERATION     syscall.Errno = 4317
	ERROR_EVT_CHANNEL_NOT_FOUND syscall.Errno = 15007
)

// EvtSubscribeFlag 定义订阅开始的时间点。
type EvtSubscribeFlag uint32

const (
	// EvtSubscribeToFutureEvents 仅订阅未来事件。
	EvtSubscribeToFutureEvents EvtSubscribeFlag = 1
	// EvtSubscribeStartAtOldestRecord 从最旧记录开始订阅。
	EvtSubscribeStartAtOldestRecord EvtSubscribeFlag = 2
	// EvtSubscribeStartAfterBookmark 从书签之后开始订阅。
	EvtSubscribeStartAfterBookmark EvtSubscribeFlag = 3
	// EvtSubscribeOriginMask 订阅原点掩码。
	EvtSubscribeOriginMask EvtSubscribeFlag = 0x3
	// EvtSubscribeTolerateQueryErrors 容忍查询错误。
	EvtSubscribeTolerateQueryErrors EvtSubscribeFlag = 0x1000
	// EvtSubscribeStrict 严格模式订阅。
	EvtSubscribeStrict EvtSubscribeFlag = 0x10000
)

// EvtRenderFlag 定义渲染内容类型。
type EvtRenderFlag uint32

const (
	// EvtRenderEventValues 渲染事件值。
	EvtRenderEventValues EvtRenderFlag = iota
	// EvtRenderEventXml 渲染事件XML。
	EvtRenderEventXml
	// EvtRenderBookmark 渲染书签。
	EvtRenderBookmark
)

// EvtRenderContextFlag 定义渲染上下文访问类型。
type EvtRenderContextFlag uint32

const (
	// EvtRenderContextValues 访问事件值的渲染上下文。
	EvtRenderContextValues EvtRenderContextFlag = iota
	// EvtRenderContextSystem 访问系统属性的渲染上下文。
	EvtRenderContextSystem
	// EvtRenderContextUser 访问用户数据的渲染上下文。
	EvtRenderContextUser
)

// EvtFormatMessageFlag 定义消息格式类型。
type EvtFormatMessageFlag uint32

const (
	// EvtFormatMessageEvent 格式化事件消息。
	EvtFormatMessageEvent EvtFormatMessageFlag = iota + 1
	// EvtFormatMessageLevel 格式化级别消息。
	EvtFormatMessageLevel
	// EvtFormatMessageTask 格式化任务消息。
	EvtFormatMessageTask
	// EvtFormatMessageOpcode 格式化操作码消息。
	EvtFormatMessageOpcode
	// EvtFormatMessageKeyword 格式化关键字消息。
	EvtFormatMessageKeyword
	// EvtFormatMessageChannel 格式化通道消息。
	EvtFormatMessageChannel
	// EvtFormatMessageProvider 格式化提供程序消息。
	EvtFormatMessageProvider
	// EvtFormatMessageId 格式化消息ID。
	EvtFormatMessageId
	// EvtFormatMessageXml 格式化XML消息。
	EvtFormatMessageXml
)

// EvtQueryFlag 定义查询行为选项。
type EvtQueryFlag uint32

const (
	// EvtQueryChannelPath 查询通道路径。
	EvtQueryChannelPath EvtQueryFlag = 0x1
	// EvtQueryFilePath 查询文件路径。
	EvtQueryFilePath EvtQueryFlag = 0x2
	// EvtQueryForwardDirection 正向查询方向。
	EvtQueryForwardDirection EvtQueryFlag = 0x100
	// EvtQueryReverseDirection 反向查询方向。
	EvtQueryReverseDirection EvtQueryFlag = 0x200
	// EvtQueryTolerateQueryErrors 容忍查询错误。
	EvtQueryTolerateQueryErrors EvtQueryFlag = 0x1000
)

// EvtOpenLogFlag 定义打开日志的类型。
type EvtOpenLogFlag uint32

const (
	// EvtOpenChannelPath 打开通道路径。
	EvtOpenChannelPath EvtOpenLogFlag = 0x1
	// EvtOpenFilePath 打开文件路径。
	EvtOpenFilePath EvtOpenLogFlag = 0x2
)

// EvtSeekFlag 定义搜索方向。
type EvtSeekFlag uint32

const (
	// EvtSeekRelativeToFirst 从第一个事件开始搜索。
	EvtSeekRelativeToFirst EvtSeekFlag = 1
	// EvtSeekRelativeToLast 从最后一个事件开始搜索。
	EvtSeekRelativeToLast EvtSeekFlag = 2
	// EvtSeekRelativeToCurrent 从当前位置开始搜索。
	EvtSeekRelativeToCurrent EvtSeekFlag = 3
	// EvtSeekRelativeToBookmark 从书签位置开始搜索。
	EvtSeekRelativeToBookmark EvtSeekFlag = 4
	// EvtSeekOriginMask 搜索原点掩码。
	EvtSeekOriginMask EvtSeekFlag = 7
	// EvtSeekStrict 严格模式搜索。
	EvtSeekStrict EvtSeekFlag = 0x10000
)

// EvtSystemPropertyID 标识系统属性的类型。
type EvtSystemPropertyID uint32

const (
	// EvtSystemProviderName 提供程序名称属性。
	EvtSystemProviderName EvtSystemPropertyID = iota
	// EvtSystemProviderGuid 提供程序GUID属性。
	EvtSystemProviderGuid
	// EvtSystemEventID 事件ID属性。
	EvtSystemEventID
	// EvtSystemQualifiers 限定符属性。
	EvtSystemQualifiers
	// EvtSystemLevel 级别属性。
	EvtSystemLevel
	// EvtSystemTask 任务属性。
	EvtSystemTask
	// EvtSystemOpcode 操作码属性。
	EvtSystemOpcode
	// EvtSystemKeywords 关键字属性。
	EvtSystemKeywords
	// EvtSystemTimeCreated 创建时间属性。
	EvtSystemTimeCreated
	// EvtSystemEventRecordId 事件记录ID属性。
	EvtSystemEventRecordId
	// EvtSystemActivityID 活动ID属性。
	EvtSystemActivityID
	// EvtSystemRelatedActivityID 相关活动ID属性。
	EvtSystemRelatedActivityID
	// EvtSystemProcessID 进程ID属性。
	EvtSystemProcessID
	// EvtSystemThreadID 线程ID属性。
	EvtSystemThreadID
	// EvtSystemChannel 通道属性。
	EvtSystemChannel
	// EvtSystemComputer 计算机名属性。
	EvtSystemComputer
	// EvtSystemUserID 用户ID属性。
	EvtSystemUserID
	// EvtSystemVersion 版本属性。
	EvtSystemVersion
	// EvtSystemPropertyIdEND 属性ID结束标记。
	EvtSystemPropertyIdEND
)

// EvtVariantType 定义EVT_VARIANT类型。
type EvtVariantType uint32

const (
	// EvtVarTypeNull 空类型。
	EvtVarTypeNull EvtVariantType = 0
	// EvtVarTypeString 字符串类型。
	EvtVarTypeString EvtVariantType = 1
	// EvtVarTypeSByte 有符号字节类型。
	EvtVarTypeSByte EvtVariantType = 2
	// EvtVarTypeByte 无符号字节类型。
	EvtVarTypeByte EvtVariantType = 3
	// EvtVarTypeInt16 16位有符号整数类型。
	EvtVarTypeInt16 EvtVariantType = 4
	// EvtVarTypeUInt16 16位无符号整数类型。
	EvtVarTypeUInt16 EvtVariantType = 5
	// EvtVarTypeInt32 32位有符号整数类型。
	EvtVarTypeInt32 EvtVariantType = 6
	// EvtVarTypeUInt32 32位无符号整数类型。
	EvtVarTypeUInt32 EvtVariantType = 7
	// EvtVarTypeInt64 64位有符号整数类型。
	EvtVarTypeInt64 EvtVariantType = 8
	// EvtVarTypeUInt64 64位无符号整数类型。
	EvtVarTypeUInt64 EvtVariantType = 9
	// EvtVarTypeSingle 单精度浮点类型。
	EvtVarTypeSingle EvtVariantType = 10
	// EvtVarTypeDouble 双精度浮点类型。
	EvtVarTypeDouble EvtVariantType = 11
	// EvtVarTypeBoolean 布尔类型。
	EvtVarTypeBoolean EvtVariantType = 12
	// EvtVarTypeBinary 二进制数据类型。
	EvtVarTypeBinary EvtVariantType = 13
	// EvtVarTypeGuid GUID类型。
	EvtVarTypeGuid EvtVariantType = 14
	// EvtVarTypeSizeT size_t类型。
	EvtVarTypeSizeT EvtVariantType = 15
	// EvtVarTypeFileTime 文件时间类型。
	EvtVarTypeFileTime EvtVariantType = 16
	// EvtVarTypeSysTime 系统时间类型。
	EvtVarTypeSysTime EvtVariantType = 17
	// EvtVarTypeSid SID类型。
	EvtVarTypeSid EvtVariantType = 18
	// EvtVarTypeHexInt32 十六进制32位整数类型。
	EvtVarTypeHexInt32 EvtVariantType = 19
	// EvtVarTypeHexInt64 十六进制64位整数类型。
	EvtVarTypeHexInt64 EvtVariantType = 20
	// EvtVarTypeEvtHandle EvtHandle类型。
	EvtVarTypeEvtHandle EvtVariantType = 32
	// EvtVarTypeEvtXml XML类型。
	EvtVarTypeEvtXml EvtVariantType = 35
	// EvtVariantTypeMask 类型掩码。
	EvtVariantTypeMask EvtVariantType = 0x7f
	// EvtVariantTypeArray 数组标志。
	EvtVariantTypeArray EvtVariantType = 128
)

// EvtVariant 对应Windows的EVT_VARIANT结构。
type EvtVariant struct {
	// Value 变体的原始值数据（8字节）。
	Value [8]byte
	// Count 数组中元素的数量。
	Count uint32
	// Type 变体的数据类型。
	Type  EvtVariantType
}

// ValueAsUint64 将Value解析为uint64类型。
//   返回 - 转换为uint64的值
func (v EvtVariant) ValueAsUint64() uint64 { return *(*uint64)(unsafe.Pointer(&v.Value)) }

// ValueAsUint32 将Value解析为uint32类型。
//   返回 - 转换为uint32的值
func (v EvtVariant) ValueAsUint32() uint32 { return *(*uint32)(unsafe.Pointer(&v.Value)) }

// ValueAsUint16 将Value解析为uint16类型。
//   返回 - 转换为uint16的值
func (v EvtVariant) ValueAsUint16() uint16 { return *(*uint16)(unsafe.Pointer(&v.Value)) }

// ValueAsUint8 将Value解析为uint8类型。
//   返回 - 转换为uint8的值
func (v EvtVariant) ValueAsUint8() uint8 { return *(*uint8)(unsafe.Pointer(&v.Value)) }

// ValueAsUintPtr 将Value解析为uintptr类型。
//   返回 - 转换为uintptr的值
func (v EvtVariant) ValueAsUintPtr() uintptr { return *(*uintptr)(unsafe.Pointer(&v.Value)) }

// ValueAsFloat32 将Value解析为float32类型。
//   返回 - 转换为float32的值
func (v EvtVariant) ValueAsFloat32() float32 { return *(*float32)(unsafe.Pointer(&v.Value)) }

// ValueAsFloat64 将Value解析为float64类型。
//   返回 - 转换为float64的值
func (v EvtVariant) ValueAsFloat64() float64 { return *(*float64)(unsafe.Pointer(&v.Value)) }

// EvtPublisherMetadataPropertyID 定义发布者元数据属性。
type EvtPublisherMetadataPropertyID uint32

const (
	// EvtPublisherMetadataPublisherGuid 发布者GUID。
	EvtPublisherMetadataPublisherGuid EvtPublisherMetadataPropertyID = iota
	// EvtPublisherMetadataResourceFilePath 资源文件路径。
	EvtPublisherMetadataResourceFilePath
	// EvtPublisherMetadataParameterFilePath 参数文件路径。
	EvtPublisherMetadataParameterFilePath
	// EvtPublisherMetadataMessageFilePath 消息文件路径。
	EvtPublisherMetadataMessageFilePath
	// EvtPublisherMetadataHelpLink 帮助链接。
	EvtPublisherMetadataHelpLink
	// EvtPublisherMetadataPublisherMessageID 发布者消息ID。
	EvtPublisherMetadataPublisherMessageID
	// EvtPublisherMetadataChannelReferences 通道引用。
	EvtPublisherMetadataChannelReferences
	// EvtPublisherMetadataChannelReferencePath 通道引用路径。
	EvtPublisherMetadataChannelReferencePath
	// EvtPublisherMetadataChannelReferenceIndex 通道引用索引。
	EvtPublisherMetadataChannelReferenceIndex
	// EvtPublisherMetadataChannelReferenceID 通道引用ID。
	EvtPublisherMetadataChannelReferenceID
	// EvtPublisherMetadataChannelReferenceFlags 通道引用标志。
	EvtPublisherMetadataChannelReferenceFlags
	// EvtPublisherMetadataChannelReferenceMessageID 通道引用消息ID。
	EvtPublisherMetadataChannelReferenceMessageID
	// EvtPublisherMetadataLevels 级别列表。
	EvtPublisherMetadataLevels
	// EvtPublisherMetadataLevelName 级别名称。
	EvtPublisherMetadataLevelName
	// EvtPublisherMetadataLevelValue 级别值。
	EvtPublisherMetadataLevelValue
	// EvtPublisherMetadataLevelMessageID 级别消息ID。
	EvtPublisherMetadataLevelMessageID
	// EvtPublisherMetadataTasks 任务列表。
	EvtPublisherMetadataTasks
	// EvtPublisherMetadataTaskName 任务名称。
	EvtPublisherMetadataTaskName
	// EvtPublisherMetadataTaskEventGuid 任务事件GUID。
	EvtPublisherMetadataTaskEventGuid
	// EvtPublisherMetadataTaskValue 任务值。
	EvtPublisherMetadataTaskValue
	// EvtPublisherMetadataTaskMessageID 任务消息ID。
	EvtPublisherMetadataTaskMessageID
	// EvtPublisherMetadataOpcodes 操作码列表。
	EvtPublisherMetadataOpcodes
	// EvtPublisherMetadataOpcodeName 操作码名称。
	EvtPublisherMetadataOpcodeName
	// EvtPublisherMetadataOpcodeValue 操作码值。
	EvtPublisherMetadataOpcodeValue
	// EvtPublisherMetadataOpcodeMessageID 操作码消息ID。
	EvtPublisherMetadataOpcodeMessageID
	// EvtPublisherMetadataKeywords 关键字列表。
	EvtPublisherMetadataKeywords
	// EvtPublisherMetadataKeywordName 关键字名称。
	EvtPublisherMetadataKeywordName
	// EvtPublisherMetadataKeywordValue 关键字值。
	EvtPublisherMetadataKeywordValue
	// EvtPublisherMetadataKeywordMessageID 关键字消息ID。
	EvtPublisherMetadataKeywordMessageID
)

// EvtEventMetadataPropertyID 定义事件元数据属性。
type EvtEventMetadataPropertyID uint32

const (
	// EventMetadataEventID 事件ID元数据。
	EventMetadataEventID EvtEventMetadataPropertyID = iota
	// EventMetadataEventVersion 事件版本元数据。
	EventMetadataEventVersion
	// EventMetadataEventChannel 事件通道元数据。
	EventMetadataEventChannel
	// EventMetadataEventLevel 事件级别元数据。
	EventMetadataEventLevel
	// EventMetadataEventOpcode 事件操作码元数据。
	EventMetadataEventOpcode
	// EventMetadataEventTask 事件任务元数据。
	EventMetadataEventTask
	// EventMetadataEventKeyword 事件关键字元数据。
	EventMetadataEventKeyword
	// EventMetadataEventMessageID 事件消息ID元数据。
	EventMetadataEventMessageID
	// EventMetadataEventTemplate 事件模板元数据。
	EventMetadataEventTemplate
)

// EvtObjectArrayPropertyHandle 数组属性句柄。
type EvtObjectArrayPropertyHandle uint32

// Close 关闭EvtObjectArrayPropertyHandle句柄。
//   返回 - 关闭过程中的错误，成功时为nil
func (h EvtObjectArrayPropertyHandle) Close() error {
	return evtClose(EvtHandle(h))
}

// EventLevel 标识事件的六个级别。
type EventLevel uint16

const (
	// EVENTLOG_LOGALWAYS_LEVEL 始终记录级别。
	EVENTLOG_LOGALWAYS_LEVEL EventLevel = iota
	// EVENTLOG_CRITICAL_LEVEL 严重错误级别。
	EVENTLOG_CRITICAL_LEVEL
	// EVENTLOG_ERROR_LEVEL 错误级别。
	EVENTLOG_ERROR_LEVEL
	// EVENTLOG_WARNING_LEVEL 警告级别。
	EVENTLOG_WARNING_LEVEL
	// EVENTLOG_INFORMATION_LEVEL 信息级别。
	EVENTLOG_INFORMATION_LEVEL
	// EVENTLOG_VERBOSE_LEVEL 详细级别。
	EVENTLOG_VERBOSE_LEVEL
)

var eventLevelToString = map[EventLevel]string{
	EVENTLOG_LOGALWAYS_LEVEL:   "Information",
	EVENTLOG_INFORMATION_LEVEL: "Information",
	EVENTLOG_CRITICAL_LEVEL:    "Critical",
	EVENTLOG_ERROR_LEVEL:       "Error",
	EVENTLOG_WARNING_LEVEL:     "Warning",
	EVENTLOG_VERBOSE_LEVEL:     "Verbose",
}

// String 返回EventLevel的可读名称。
//   返回 - 级别的文本描述，未知级别返回"Level(N)"格式
func (et EventLevel) String() string {
	s, ok := eventLevelToString[et]
	if ok {
		return s
	}
	return fmt.Sprintf("Level(%d)", et)
}

// EVT_EXPORTLOG_FLAGS 用于EvtExportLog的标志。
type EVT_EXPORTLOG_FLAGS uint32

const (
	// EvtExportLogChannelPath 导出日志的通道路径标志。
	EvtExportLogChannelPath EVT_EXPORTLOG_FLAGS = 0x1
	// EvtExportLogFilePath 导出日志的文件路径标志。
	EvtExportLogFilePath EVT_EXPORTLOG_FLAGS = 0x2
	// EvtExportLogTolerateQueryErrors 容忍查询错误标志。
	EvtExportLogTolerateQueryErrors EVT_EXPORTLOG_FLAGS = 0x1000
	// EvtExportLogOverwrite 覆盖现有文件标志。
	EvtExportLogOverwrite EVT_EXPORTLOG_FLAGS = 0x2000
)

// DLL and proc declarations.
var modwevtapi = windows.NewLazySystemDLL("wevtapi.dll")

var (
	procEvtClearLog                     = modwevtapi.NewProc("EvtClearLog")
	procEvtClose                        = modwevtapi.NewProc("EvtClose")
	procEvtCreateBookmark               = modwevtapi.NewProc("EvtCreateBookmark")
	procEvtCreateRenderContext          = modwevtapi.NewProc("EvtCreateRenderContext")
	procEvtExportLog                    = modwevtapi.NewProc("EvtExportLog")
	procEvtFormatMessage                = modwevtapi.NewProc("EvtFormatMessage")
	procEvtGetEventMetadataProperty     = modwevtapi.NewProc("EvtGetEventMetadataProperty")
	procEvtGetObjectArrayProperty       = modwevtapi.NewProc("EvtGetObjectArrayProperty")
	procEvtGetObjectArraySize           = modwevtapi.NewProc("EvtGetObjectArraySize")
	procEvtGetPublisherMetadataProperty = modwevtapi.NewProc("EvtGetPublisherMetadataProperty")
	procEvtNext                         = modwevtapi.NewProc("EvtNext")
	procEvtNextChannelPath              = modwevtapi.NewProc("EvtNextChannelPath")
	procEvtNextEventMetadata            = modwevtapi.NewProc("EvtNextEventMetadata")
	procEvtNextPublisherId              = modwevtapi.NewProc("EvtNextPublisherId")
	procEvtOpenChannelEnum              = modwevtapi.NewProc("EvtOpenChannelEnum")
	procEvtOpenEventMetadataEnum        = modwevtapi.NewProc("EvtOpenEventMetadataEnum")
	procEvtOpenLog                      = modwevtapi.NewProc("EvtOpenLog")
	procEvtOpenPublisherEnum            = modwevtapi.NewProc("EvtOpenPublisherEnum")
	procEvtOpenPublisherMetadata        = modwevtapi.NewProc("EvtOpenPublisherMetadata")
	procEvtQuery                        = modwevtapi.NewProc("EvtQuery")
	procEvtRender                       = modwevtapi.NewProc("EvtRender")
	procEvtSeek                         = modwevtapi.NewProc("EvtSeek")
	procEvtSubscribe                    = modwevtapi.NewProc("EvtSubscribe")
	procEvtUpdateBookmark               = modwevtapi.NewProc("EvtUpdateBookmark")
)

func evtClearLog(session EvtHandle, channelPath, targetFilePath *uint16, flags uint32) error {
	r1, _, e1 := syscall.SyscallN(procEvtClearLog.Addr(),
		uintptr(session),
		uintptr(unsafe.Pointer(channelPath)),
		uintptr(unsafe.Pointer(targetFilePath)),
		uintptr(flags))
	if r1 == 0 {
		return e1
	}
	return nil
}

func evtClose(object EvtHandle) error {
	r1, _, e1 := syscall.SyscallN(procEvtClose.Addr(), uintptr(object))
	if r1 == 0 {
		return e1
	}
	return nil
}

func evtCreateBookmark(bookmarkXML *uint16) (EvtHandle, error) {
	r0, _, e1 := syscall.SyscallN(procEvtCreateBookmark.Addr(), uintptr(unsafe.Pointer(bookmarkXML)))
	h := EvtHandle(r0)
	if h == 0 {
		return 0, e1
	}
	return h, nil
}

func evtCreateRenderContext(valuePathsCount uint32, valuePaths **uint16, flags EvtRenderContextFlag) (EvtHandle, error) {
	r0, _, e1 := syscall.SyscallN(procEvtCreateRenderContext.Addr(),
		uintptr(valuePathsCount),
		uintptr(unsafe.Pointer(valuePaths)),
		uintptr(flags))
	h := EvtHandle(r0)
	if h == 0 {
		return 0, e1
	}
	return h, nil
}

func evtExportLog(session EvtHandle, path, targetFilePath, query *uint16, flags EVT_EXPORTLOG_FLAGS) error {
	r1, _, e1 := syscall.SyscallN(procEvtExportLog.Addr(),
		uintptr(session),
		uintptr(unsafe.Pointer(path)),
		uintptr(unsafe.Pointer(targetFilePath)),
		uintptr(unsafe.Pointer(query)),
		uintptr(flags))
	if r1 == 0 {
		return e1
	}
	return nil
}

func evtFormatMessage(publisherMetadata, event EvtHandle, messageID uint32, valueCount uint32, values *EvtVariant, flags EvtFormatMessageFlag, bufferSize uint32, buffer *byte, bufferUsed *uint32) error {
	r1, _, e1 := syscall.SyscallN(procEvtFormatMessage.Addr(),
		uintptr(publisherMetadata),
		uintptr(event),
		uintptr(messageID),
		uintptr(valueCount),
		uintptr(unsafe.Pointer(values)),
		uintptr(flags),
		uintptr(bufferSize),
		uintptr(unsafe.Pointer(buffer)),
		uintptr(unsafe.Pointer(bufferUsed)))
	if r1 == 0 {
		return e1
	}
	return nil
}

func evtGetEventMetadataProperty(eventMetadata EvtHandle, propertyID EvtEventMetadataPropertyID, flags uint32, bufferSize uint32, variant *EvtVariant, bufferUsed *uint32) error {
	r1, _, e1 := syscall.SyscallN(procEvtGetEventMetadataProperty.Addr(),
		uintptr(eventMetadata),
		uintptr(propertyID),
		uintptr(flags),
		uintptr(bufferSize),
		uintptr(unsafe.Pointer(variant)),
		uintptr(unsafe.Pointer(bufferUsed)))
	if r1 == 0 {
		return e1
	}
	return nil
}

func evtGetObjectArrayProperty(objectArray EvtObjectArrayPropertyHandle, propertyID EvtPublisherMetadataPropertyID, arrayIndex uint32, flags uint32, bufferSize uint32, evtVariant *EvtVariant, bufferUsed *uint32) error {
	r1, _, e1 := syscall.SyscallN(procEvtGetObjectArrayProperty.Addr(),
		uintptr(objectArray),
		uintptr(propertyID),
		uintptr(arrayIndex),
		uintptr(flags),
		uintptr(bufferSize),
		uintptr(unsafe.Pointer(evtVariant)),
		uintptr(unsafe.Pointer(bufferUsed)))
	if r1 == 0 {
		return e1
	}
	return nil
}

func evtGetObjectArraySize(objectArray EvtObjectArrayPropertyHandle, arraySize *uint32) error {
	r1, _, e1 := syscall.SyscallN(procEvtGetObjectArraySize.Addr(),
		uintptr(objectArray),
		uintptr(unsafe.Pointer(arraySize)))
	if r1 == 0 {
		return e1
	}
	return nil
}

func evtGetPublisherMetadataProperty(publisherMetadata EvtHandle, propertyID EvtPublisherMetadataPropertyID, flags uint32, bufferSize uint32, variant *EvtVariant, bufferUsed *uint32) error {
	r1, _, e1 := syscall.SyscallN(procEvtGetPublisherMetadataProperty.Addr(),
		uintptr(publisherMetadata),
		uintptr(propertyID),
		uintptr(flags),
		uintptr(bufferSize),
		uintptr(unsafe.Pointer(variant)),
		uintptr(unsafe.Pointer(bufferUsed)))
	if r1 == 0 {
		return e1
	}
	return nil
}

func evtNext(resultSet EvtHandle, eventArraySize uint32, eventArray *EvtHandle, timeout uint32, flags uint32, numReturned *uint32) error {
	r1, _, e1 := syscall.SyscallN(procEvtNext.Addr(),
		uintptr(resultSet),
		uintptr(eventArraySize),
		uintptr(unsafe.Pointer(eventArray)),
		uintptr(timeout),
		uintptr(flags),
		uintptr(unsafe.Pointer(numReturned)))
	if r1 == 0 {
		return e1
	}
	return nil
}

func evtNextChannelPath(channelEnum EvtHandle, bufferSize uint32, buffer *uint16, bufferUsed *uint32) error {
	r1, _, e1 := syscall.SyscallN(procEvtNextChannelPath.Addr(),
		uintptr(channelEnum),
		uintptr(bufferSize),
		uintptr(unsafe.Pointer(buffer)),
		uintptr(unsafe.Pointer(bufferUsed)))
	if r1 == 0 {
		return e1
	}
	return nil
}

func evtNextEventMetadata(enumerator EvtHandle, flags uint32) (EvtHandle, error) {
	r0, _, e1 := syscall.SyscallN(procEvtNextEventMetadata.Addr(),
		uintptr(enumerator), uintptr(flags))
	h := EvtHandle(r0)
	if h == 0 {
		return 0, e1
	}
	return h, nil
}

func evtNextPublisherId(enumerator EvtHandle, bufferSize uint32, buffer *uint16, bufferUsed *uint32) error {
	r1, _, e1 := syscall.SyscallN(procEvtNextPublisherId.Addr(),
		uintptr(enumerator),
		uintptr(bufferSize),
		uintptr(unsafe.Pointer(buffer)),
		uintptr(unsafe.Pointer(bufferUsed)))
	if r1 == 0 {
		return e1
	}
	return nil
}

func evtOpenChannelEnum(session EvtHandle, flags uint32) (EvtHandle, error) {
	r0, _, e1 := syscall.SyscallN(procEvtOpenChannelEnum.Addr(),
		uintptr(session), uintptr(flags))
	h := EvtHandle(r0)
	if h == 0 {
		return 0, e1
	}
	return h, nil
}

func evtOpenEventMetadataEnum(publisherMetadata EvtHandle, flags uint32) (EvtHandle, error) {
	r0, _, e1 := syscall.SyscallN(procEvtOpenEventMetadataEnum.Addr(),
		uintptr(publisherMetadata), uintptr(flags))
	h := EvtHandle(r0)
	if h == 0 {
		return 0, e1
	}
	return h, nil
}

func evtOpenLog(session EvtHandle, path *uint16, flags uint32) (EvtHandle, error) {
	r0, _, e1 := syscall.SyscallN(procEvtOpenLog.Addr(),
		uintptr(session), uintptr(unsafe.Pointer(path)), uintptr(flags))
	h := EvtHandle(r0)
	if h == 0 {
		return 0, e1
	}
	return h, nil
}

func evtOpenPublisherEnum(session EvtHandle, flags uint32) (EvtHandle, error) {
	r0, _, e1 := syscall.SyscallN(procEvtOpenPublisherEnum.Addr(),
		uintptr(session), uintptr(flags))
	h := EvtHandle(r0)
	if h == 0 {
		return 0, e1
	}
	return h, nil
}

func evtOpenPublisherMetadata(session EvtHandle, publisherIdentity, logFilePath *uint16, locale uint32, flags uint32) (EvtHandle, error) {
	r0, _, e1 := syscall.SyscallN(procEvtOpenPublisherMetadata.Addr(),
		uintptr(session),
		uintptr(unsafe.Pointer(publisherIdentity)),
		uintptr(unsafe.Pointer(logFilePath)),
		uintptr(locale),
		uintptr(flags))
	h := EvtHandle(r0)
	if h == 0 {
		return 0, e1
	}
	return h, nil
}

func evtQuery(session EvtHandle, path, query *uint16, flags uint32) (EvtHandle, error) {
	r0, _, e1 := syscall.SyscallN(procEvtQuery.Addr(),
		uintptr(session),
		uintptr(unsafe.Pointer(path)),
		uintptr(unsafe.Pointer(query)),
		uintptr(flags))
	h := EvtHandle(r0)
	if h == 0 {
		return 0, e1
	}
	return h, nil
}

func evtRender(context, fragment EvtHandle, flags EvtRenderFlag, bufferSize uint32, buffer *byte, bufferUsed, propertyCount *uint32) error {
	r1, _, e1 := syscall.SyscallN(procEvtRender.Addr(),
		uintptr(context),
		uintptr(fragment),
		uintptr(flags),
		uintptr(bufferSize),
		uintptr(unsafe.Pointer(buffer)),
		uintptr(unsafe.Pointer(bufferUsed)),
		uintptr(unsafe.Pointer(propertyCount)))
	if r1 == 0 {
		return e1
	}
	return nil
}

func evtSeek(resultSet EvtHandle, position int64, bookmark EvtHandle, timeout uint32, flags uint32) error {
	r1, _, e1 := syscall.SyscallN(procEvtSeek.Addr(),
		uintptr(resultSet),
		uintptr(position),
		uintptr(bookmark),
		uintptr(timeout),
		uintptr(flags))
	if r1 == 0 {
		return e1
	}
	return nil
}

func evtSubscribe(session EvtHandle, signalEvent uintptr, channelPath, query *uint16, bookmark EvtHandle, context uintptr, callback syscall.Handle, flags EvtSubscribeFlag) (EvtHandle, error) {
	r0, _, e1 := syscall.SyscallN(procEvtSubscribe.Addr(),
		uintptr(session),
		uintptr(signalEvent),
		uintptr(unsafe.Pointer(channelPath)),
		uintptr(unsafe.Pointer(query)),
		uintptr(bookmark),
		uintptr(context),
		uintptr(callback),
		uintptr(flags))
	h := EvtHandle(r0)
	if h == 0 {
		return 0, e1
	}
	return h, nil
}

func evtUpdateBookmark(bookmark, event EvtHandle) error {
	r1, _, e1 := syscall.SyscallN(procEvtUpdateBookmark.Addr(),
		uintptr(bookmark), uintptr(event))
	if r1 == 0 {
		return e1
	}
	return nil
}

// Higher-level wrappers.

// IsAvailable 检查wevtapi DLL是否可加载。
//   返回1 - DLL是否可用
//   返回2 - 加载DLL时的错误，成功时为nil
func IsAvailable() (bool, error) {
	err := modwevtapi.Load()
	if err != nil {
		return false, err
	}
	return true, nil
}

// EvtQuery 对通道或日志文件执行查询。
//   session - 会话句柄，0表示本地会话
//   path - 通道名称或.evtx文件路径
//   query - 结构化XML查询字符串
//   flags - 查询行为标志，指定path是通道还是文件
//   返回1 - 查询结果集的句柄
//   返回2 - 查询过程中的错误，成功时为nil
func EvtQuery(session EvtHandle, path, query string, flags EvtQueryFlag) (EvtHandle, error) {
	var pathPtr, queryPtr *uint16
	var err error
	if path != "" {
		pathPtr, err = syscall.UTF16PtrFromString(path)
		if err != nil {
			return 0, err
		}
	}
	if query != "" {
		queryPtr, err = syscall.UTF16PtrFromString(query)
		if err != nil {
			return 0, err
		}
	}
	return evtQuery(session, pathPtr, queryPtr, uint32(flags))
}

// Subscribe 创建事件订阅。
//   session - 会话句柄，0表示本地会话
//   signalEvent - 用于通知的事件句柄
//   channelPath - 通道名称
//   query - 结构化XML查询字符串
//   bookmark - 书签句柄，指定订阅起始位置
//   flags - 订阅标志，指定订阅起始点和行为
//   返回1 - 订阅句柄
//   返回2 - 订阅过程中的错误，成功时为nil
func Subscribe(session EvtHandle, signalEvent windows.Handle, channelPath, query string, bookmark EvtHandle, flags EvtSubscribeFlag) (EvtHandle, error) {
	var cp, q *uint16
	var err error
	if channelPath != "" {
		cp, err = syscall.UTF16PtrFromString(channelPath)
		if err != nil {
			return 0, err
		}
	}
	if query != "" {
		q, err = syscall.UTF16PtrFromString(query)
		if err != nil {
			return 0, err
		}
	}
	return evtSubscribe(session, uintptr(signalEvent), cp, q, bookmark, 0, 0, flags)
}

// EventHandles 从订阅中读取事件句柄。
//   subscription - 订阅或查询结果集的句柄
//   maxHandles - 单次读取的最大事件句柄数
//   返回1 - 事件句柄切片
//   返回2 - 读取过程中的错误，成功时为nil；无更多事件时返回ERROR_NO_MORE_ITEMS
func EventHandles(subscription EvtHandle, maxHandles int) ([]EvtHandle, error) {
	if maxHandles < 1 {
		return nil, fmt.Errorf("maxHandles must be > 0")
	}
	handles := make([]EvtHandle, maxHandles)
	var numRead uint32
	err := evtNext(subscription, uint32(len(handles)), &handles[0], 0, 0, &numRead)
	if err != nil {
		if err == ERROR_INVALID_OPERATION && numRead == 0 {
			return nil, ERROR_NO_MORE_ITEMS
		}
		return nil, err
	}
	return handles[:numRead], nil
}

// RenderEventXML 将事件句柄渲染为XML。
//   eventHandle - 待渲染的事件句柄
//   renderBuf - 渲染用的字节缓冲区
//   out - 输出缓冲区，渲染结果写入此处
//   返回 - 渲染过程中的错误，成功时为nil；缓冲区不足时返回InsufficientBufferError
func RenderEventXML(eventHandle EvtHandle, renderBuf []byte, out *ByteBuffer) error {
	var bufferUsed, propertyCount uint32
	err := evtRender(0, eventHandle, EvtRenderEventXml,
		uint32(len(renderBuf)), &renderBuf[0], &bufferUsed, &propertyCount)
	if err == ERROR_INSUFFICIENT_BUFFER {
		return InsufficientBufferError{RequiredSize: int(bufferUsed)}
	}
	if err != nil {
		return err
	}
	if int(bufferUsed) > len(renderBuf) {
		return fmt.Errorf("EvtRender wrote %d bytes but buffer has %d", bufferUsed, len(renderBuf))
	}
	if bufferUsed > 0 {
		_, err = out.Write(renderBuf[:bufferUsed])
	}
	return err
}

// InsufficientBufferError 表示缓冲区大小不足的错误。
type InsufficientBufferError struct {
	// RequiredSize 所需的缓冲区大小。
	RequiredSize int
}

// Error 返回缓冲区不足的错误描述。
//   返回 - 包含所需缓冲区大小的错误文本
func (e InsufficientBufferError) Error() string {
	return fmt.Sprintf("insufficient buffer, need %d bytes", e.RequiredSize)
}

// CreateBookmarkFromXML 从XML创建书签。
//   bookmarkXML - 书签的XML字符串
//   返回1 - 书签句柄
//   返回2 - 创建过程中的错误，成功时为nil
func CreateBookmarkFromXML(bookmarkXML string) (EvtHandle, error) {
	utf16, err := syscall.UTF16PtrFromString(bookmarkXML)
	if err != nil {
		return 0, err
	}
	return evtCreateBookmark(utf16)
}

// CreateRenderContext 创建渲染上下文。
//   valuePaths - 要渲染的属性路径列表
//   flag - 渲染上下文标志，指定访问类型
//   返回1 - 渲染上下文句柄
//   返回2 - 创建过程中的错误，成功时为nil
func CreateRenderContext(valuePaths []string, flag EvtRenderContextFlag) (EvtHandle, error) {
	paths := make([]*uint16, len(valuePaths))
	for i, p := range valuePaths {
		var err error
		paths[i], err = syscall.UTF16PtrFromString(p)
		if err != nil {
			return 0, err
		}
	}
	var pathsAddr **uint16
	if len(paths) > 0 {
		pathsAddr = &paths[0]
	}
	return evtCreateRenderContext(uint32(len(paths)), pathsAddr, flag)
}

// OpenPublisherMetadata 打开发布者的元数据。
//   session - 会话句柄，0表示本地会话
//   publisherName - 发布者名称
//   lang - 区域设置标识符(LCID)，用于消息格式化
//   返回1 - 发布者元数据句柄
//   返回2 - 打开过程中的错误，成功时为nil
func OpenPublisherMetadata(session EvtHandle, publisherName string, lang uint32) (EvtHandle, error) {
	p, err := syscall.UTF16PtrFromString(publisherName)
	if err != nil {
		return 0, err
	}
	return evtOpenPublisherMetadata(session, p, nil, lang, 0)
}

// Channels 列出所有已注册的通道。
//   返回1 - 通道名称字符串切片
//   返回2 - 遍历过程中的错误，成功时为nil
func Channels() ([]string, error) {
	handle, err := evtOpenChannelEnum(0, 0)
	if err != nil {
		return nil, err
	}
	defer handle.Close()

	var channels []string
	buf := make([]uint16, 512)
	for {
		var used uint32
		err := evtNextChannelPath(handle, uint32(len(buf)), &buf[0], &used)
		if err != nil {
			if err == ERROR_INSUFFICIENT_BUFFER {
				buf = make([]uint16, 2*len(buf))
				continue
			}
			if err == ERROR_NO_MORE_ITEMS {
				break
			}
			return nil, err
		}
		channels = append(channels, syscall.UTF16ToString(buf[:used]))
	}
	return channels, nil
}

// EvtVariantData 从EvtVariant中读取类型化数据。
//   v - 待读取的EvtVariant变体
//   buf - 用于解析引用类型（字符串、GUID、SID）的原始缓冲区
//   返回1 - 根据变体类型解析后的值
//   返回2 - 解析过程中的错误，成功时为nil
func EvtVariantData(v EvtVariant, buf []byte) (any, error) {
	typ := v.Type & EvtVariantTypeMask
	switch typ {
	case EvtVarTypeNull:
		return nil, nil
	case EvtVarTypeString:
		return readVariantString(v, buf)
	case EvtVarTypeSByte:
		return int8(v.ValueAsUint8()), nil
	case EvtVarTypeByte:
		return v.ValueAsUint8(), nil
	case EvtVarTypeInt16:
		return int16(v.ValueAsUint16()), nil
	case EvtVarTypeInt32, EvtVarTypeHexInt32:
		return int32(v.ValueAsUint32()), nil
	case EvtVarTypeInt64, EvtVarTypeHexInt64:
		return int64(v.ValueAsUint64()), nil
	case EvtVarTypeUInt16:
		return v.ValueAsUint16(), nil
	case EvtVarTypeUInt32:
		return v.ValueAsUint32(), nil
	case EvtVarTypeUInt64:
		return v.ValueAsUint64(), nil
	case EvtVarTypeSingle:
		return v.ValueAsFloat32(), nil
	case EvtVarTypeDouble:
		return v.ValueAsFloat64(), nil
	case EvtVarTypeBoolean:
		return v.ValueAsUint8() != 0, nil
	case EvtVarTypeGuid:
		return readVariantGUID(v, buf)
	case EvtVarTypeFileTime:
		ft := (*windows.Filetime)(unsafe.Pointer(&v.Value))
		return time.Unix(0, ft.Nanoseconds()).UTC(), nil
	case EvtVarTypeSid:
		return readVariantSID(v, buf)
	case EvtVarTypeEvtHandle:
		return EvtHandle(v.ValueAsUintPtr()), nil
	default:
		return nil, fmt.Errorf("unhandled EvtVariant type: %d", typ)
	}
}

func readVariantString(v EvtVariant, buf []byte) (string, error) {
	addr := unsafe.Pointer(&buf[0])
	offset := v.ValueAsUintPtr() - uintptr(addr)
	if int(offset) >= len(buf) {
		return "", fmt.Errorf("string offset out of bounds")
	}
	aligned := offset & ^uintptr(1) // align to 2-byte boundary
	u16 := unsafe.Slice((*uint16)(unsafe.Pointer(&buf[aligned])), (len(buf)-int(aligned))/2)
	return windows.UTF16ToString(u16), nil
}

func readVariantGUID(v EvtVariant, buf []byte) (windows.GUID, error) {
	addr := unsafe.Pointer(&buf[0])
	offset := v.ValueAsUintPtr() - uintptr(addr)
	if int(offset+unsafe.Sizeof(windows.GUID{})) > len(buf) {
		return windows.GUID{}, fmt.Errorf("guid offset out of bounds")
	}
	return *(*windows.GUID)(unsafe.Pointer(&buf[offset])), nil
}

func readVariantSID(v EvtVariant, buf []byte) (*windows.SID, error) {
	addr := unsafe.Pointer(&buf[0])
	offset := v.ValueAsUintPtr() - uintptr(addr)
	if int(offset) >= len(buf) {
		return nil, fmt.Errorf("sid offset out of bounds")
	}
	sidPtr := (*windows.SID)(unsafe.Pointer(&buf[offset]))
	return sidPtr.Copy()
}

// Publishers 返回排序后的事件发布者列表。
//   返回1 - 发布者名称字符串切片
//   返回2 - 遍历过程中的错误，成功时为nil
func Publishers() ([]string, error) {
	enumerator, err := evtOpenPublisherEnum(0, 0)
	if err != nil {
		return nil, fmt.Errorf("EvtOpenPublisherEnum: %w", err)
	}
	defer enumerator.Close()

	var publishers []string
	buf := make([]uint16, 1024)
	for {
		var used uint32
		err := evtNextPublisherId(enumerator, uint32(len(buf)), &buf[0], &used)
		if err != nil {
			if err == ERROR_NO_MORE_ITEMS {
				break
			}
			if err == ERROR_INSUFFICIENT_BUFFER {
				buf = make([]uint16, used)
				continue
			}
			return nil, fmt.Errorf("EvtNextPublisherId: %w", err)
		}
		publishers = append(publishers, windows.UTF16ToString(buf))
	}
	return publishers, nil
}
