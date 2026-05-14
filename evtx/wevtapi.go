//go:build windows

package evtx

import (
	"fmt"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/kitsch-9527/wcorefx/internal/winapi"
)

// EvtHandle 是事件日志API的句柄。
type EvtHandle uintptr

// NilHandle 表示一个空的EvtHandle值。
const NilHandle EvtHandle = 0

// Close 关闭EvtHandle句柄。
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
	EvtSubscribeToFutureEvents          EvtSubscribeFlag = 1
	EvtSubscribeStartAtOldestRecord     EvtSubscribeFlag = 2
	EvtSubscribeStartAfterBookmark      EvtSubscribeFlag = 3
	EvtSubscribeOriginMask              EvtSubscribeFlag = 0x3
	EvtSubscribeTolerateQueryErrors     EvtSubscribeFlag = 0x1000
	EvtSubscribeStrict                  EvtSubscribeFlag = 0x10000
)

// EvtRenderFlag 定义渲染内容类型。
type EvtRenderFlag uint32

const (
	EvtRenderEventValues EvtRenderFlag = iota
	EvtRenderEventXml
	EvtRenderBookmark
)

// EvtRenderContextFlag 定义渲染上下文访问类型。
type EvtRenderContextFlag uint32

const (
	EvtRenderContextValues EvtRenderContextFlag = iota
	EvtRenderContextSystem
	EvtRenderContextUser
)

// EvtFormatMessageFlag 定义消息格式类型。
type EvtFormatMessageFlag uint32

const (
	EvtFormatMessageEvent EvtFormatMessageFlag = iota + 1
	EvtFormatMessageLevel
	EvtFormatMessageTask
	EvtFormatMessageOpcode
	EvtFormatMessageKeyword
	EvtFormatMessageChannel
	EvtFormatMessageProvider
	EvtFormatMessageId
	EvtFormatMessageXml
)

// EvtQueryFlag 定义查询行为选项。
type EvtQueryFlag uint32

const (
	EvtQueryChannelPath          EvtQueryFlag = 0x1
	EvtQueryFilePath             EvtQueryFlag = 0x2
	EvtQueryForwardDirection     EvtQueryFlag = 0x100
	EvtQueryReverseDirection     EvtQueryFlag = 0x200
	EvtQueryTolerateQueryErrors  EvtQueryFlag = 0x1000
)

// EvtOpenLogFlag 定义打开日志的类型。
type EvtOpenLogFlag uint32

const (
	EvtOpenChannelPath EvtOpenLogFlag = 0x1
	EvtOpenFilePath    EvtOpenLogFlag = 0x2
)

// EvtSeekFlag 定义搜索方向。
type EvtSeekFlag uint32

const (
	EvtSeekRelativeToFirst    EvtSeekFlag = 1
	EvtSeekRelativeToLast     EvtSeekFlag = 2
	EvtSeekRelativeToCurrent  EvtSeekFlag = 3
	EvtSeekRelativeToBookmark EvtSeekFlag = 4
	EvtSeekOriginMask         EvtSeekFlag = 7
	EvtSeekStrict             EvtSeekFlag = 0x10000
)

// EvtSystemPropertyID 标识系统属性的类型。
type EvtSystemPropertyID uint32

const (
	EvtSystemProviderName         EvtSystemPropertyID = iota
	EvtSystemProviderGuid
	EvtSystemEventID
	EvtSystemQualifiers
	EvtSystemLevel
	EvtSystemTask
	EvtSystemOpcode
	EvtSystemKeywords
	EvtSystemTimeCreated
	EvtSystemEventRecordId
	EvtSystemActivityID
	EvtSystemRelatedActivityID
	EvtSystemProcessID
	EvtSystemThreadID
	EvtSystemChannel
	EvtSystemComputer
	EvtSystemUserID
	EvtSystemVersion
	EvtSystemPropertyIdEND
)

// EvtVariantType 定义EVT_VARIANT类型。
type EvtVariantType uint32

const (
	EvtVarTypeNull    EvtVariantType = 0
	EvtVarTypeString  EvtVariantType = 1
	EvtVarTypeSByte   EvtVariantType = 2
	EvtVarTypeByte    EvtVariantType = 3
	EvtVarTypeInt16   EvtVariantType = 4
	EvtVarTypeUInt16  EvtVariantType = 5
	EvtVarTypeInt32   EvtVariantType = 6
	EvtVarTypeUInt32  EvtVariantType = 7
	EvtVarTypeInt64   EvtVariantType = 8
	EvtVarTypeUInt64  EvtVariantType = 9
	EvtVarTypeSingle  EvtVariantType = 10
	EvtVarTypeDouble  EvtVariantType = 11
	EvtVarTypeBoolean EvtVariantType = 12
	EvtVarTypeBinary  EvtVariantType = 13
	EvtVarTypeGuid    EvtVariantType = 14
	EvtVarTypeSizeT   EvtVariantType = 15
	EvtVarTypeFileTime EvtVariantType = 16
	EvtVarTypeSysTime EvtVariantType = 17
	EvtVarTypeSid     EvtVariantType = 18
	EvtVarTypeHexInt32 EvtVariantType = 19
	EvtVarTypeHexInt64 EvtVariantType = 20
	EvtVarTypeEvtHandle EvtVariantType = 32
	EvtVarTypeEvtXml  EvtVariantType = 35
	EvtVariantTypeMask EvtVariantType = 0x7f
	EvtVariantTypeArray EvtVariantType = 128
)

// EvtVariant 对应Windows的EVT_VARIANT结构。
type EvtVariant struct {
	Value [8]byte
	Count uint32
	Type  EvtVariantType
}

func (v EvtVariant) ValueAsUint64() uint64 { return *(*uint64)(unsafe.Pointer(&v.Value)) }
func (v EvtVariant) ValueAsUint32() uint32 { return *(*uint32)(unsafe.Pointer(&v.Value)) }
func (v EvtVariant) ValueAsUint16() uint16 { return *(*uint16)(unsafe.Pointer(&v.Value)) }
func (v EvtVariant) ValueAsUint8() uint8   { return *(*uint8)(unsafe.Pointer(&v.Value)) }
func (v EvtVariant) ValueAsUintPtr() uintptr { return *(*uintptr)(unsafe.Pointer(&v.Value)) }
func (v EvtVariant) ValueAsFloat32() float32 { return *(*float32)(unsafe.Pointer(&v.Value)) }
func (v EvtVariant) ValueAsFloat64() float64 { return *(*float64)(unsafe.Pointer(&v.Value)) }

// EvtPublisherMetadataPropertyID 定义发布者元数据属性。
type EvtPublisherMetadataPropertyID uint32

const (
	EvtPublisherMetadataPublisherGuid           EvtPublisherMetadataPropertyID = iota
	EvtPublisherMetadataResourceFilePath
	EvtPublisherMetadataParameterFilePath
	EvtPublisherMetadataMessageFilePath
	EvtPublisherMetadataHelpLink
	EvtPublisherMetadataPublisherMessageID
	EvtPublisherMetadataChannelReferences
	EvtPublisherMetadataChannelReferencePath
	EvtPublisherMetadataChannelReferenceIndex
	EvtPublisherMetadataChannelReferenceID
	EvtPublisherMetadataChannelReferenceFlags
	EvtPublisherMetadataChannelReferenceMessageID
	EvtPublisherMetadataLevels
	EvtPublisherMetadataLevelName
	EvtPublisherMetadataLevelValue
	EvtPublisherMetadataLevelMessageID
	EvtPublisherMetadataTasks
	EvtPublisherMetadataTaskName
	EvtPublisherMetadataTaskEventGuid
	EvtPublisherMetadataTaskValue
	EvtPublisherMetadataTaskMessageID
	EvtPublisherMetadataOpcodes
	EvtPublisherMetadataOpcodeName
	EvtPublisherMetadataOpcodeValue
	EvtPublisherMetadataOpcodeMessageID
	EvtPublisherMetadataKeywords
	EvtPublisherMetadataKeywordName
	EvtPublisherMetadataKeywordValue
	EvtPublisherMetadataKeywordMessageID
)

// EvtEventMetadataPropertyID 定义事件元数据属性。
type EvtEventMetadataPropertyID uint32

const (
	EventMetadataEventID        EvtEventMetadataPropertyID = iota
	EventMetadataEventVersion
	EventMetadataEventChannel
	EventMetadataEventLevel
	EventMetadataEventOpcode
	EventMetadataEventTask
	EventMetadataEventKeyword
	EventMetadataEventMessageID
	EventMetadataEventTemplate
)

// EvtObjectArrayPropertyHandle 数组属性句柄。
type EvtObjectArrayPropertyHandle uint32

func (h EvtObjectArrayPropertyHandle) Close() error {
	return evtClose(EvtHandle(h))
}

// EventLevel 标识事件的六个级别。
type EventLevel uint16

const (
	EVENTLOG_LOGALWAYS_LEVEL   EventLevel = iota
	EVENTLOG_CRITICAL_LEVEL
	EVENTLOG_ERROR_LEVEL
	EVENTLOG_WARNING_LEVEL
	EVENTLOG_INFORMATION_LEVEL
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
	EvtExportLogChannelPath          EVT_EXPORTLOG_FLAGS = 0x1
	EvtExportLogFilePath             EVT_EXPORTLOG_FLAGS = 0x2
	EvtExportLogTolerateQueryErrors  EVT_EXPORTLOG_FLAGS = 0x1000
	EvtExportLogOverwrite            EVT_EXPORTLOG_FLAGS = 0x2000
)

// Proc declarations.
var (
	procEvtClearLog                     = winapi.NewProc("wevtapi.dll", "EvtClearLog")
	procEvtClose                        = winapi.NewProc("wevtapi.dll", "EvtClose")
	procEvtCreateBookmark               = winapi.NewProc("wevtapi.dll", "EvtCreateBookmark")
	procEvtCreateRenderContext          = winapi.NewProc("wevtapi.dll", "EvtCreateRenderContext")
	procEvtExportLog                    = winapi.NewProc("wevtapi.dll", "EvtExportLog")
	procEvtFormatMessage                = winapi.NewProc("wevtapi.dll", "EvtFormatMessage")
	procEvtGetEventMetadataProperty     = winapi.NewProc("wevtapi.dll", "EvtGetEventMetadataProperty")
	procEvtGetObjectArrayProperty       = winapi.NewProc("wevtapi.dll", "EvtGetObjectArrayProperty")
	procEvtGetObjectArraySize           = winapi.NewProc("wevtapi.dll", "EvtGetObjectArraySize")
	procEvtGetPublisherMetadataProperty = winapi.NewProc("wevtapi.dll", "EvtGetPublisherMetadataProperty")
	procEvtNext                         = winapi.NewProc("wevtapi.dll", "EvtNext")
	procEvtNextChannelPath              = winapi.NewProc("wevtapi.dll", "EvtNextChannelPath")
	procEvtNextEventMetadata            = winapi.NewProc("wevtapi.dll", "EvtNextEventMetadata")
	procEvtNextPublisherId              = winapi.NewProc("wevtapi.dll", "EvtNextPublisherId")
	procEvtOpenChannelEnum              = winapi.NewProc("wevtapi.dll", "EvtOpenChannelEnum")
	procEvtOpenEventMetadataEnum        = winapi.NewProc("wevtapi.dll", "EvtOpenEventMetadataEnum")
	procEvtOpenLog                      = winapi.NewProc("wevtapi.dll", "EvtOpenLog")
	procEvtOpenPublisherEnum            = winapi.NewProc("wevtapi.dll", "EvtOpenPublisherEnum")
	procEvtOpenPublisherMetadata        = winapi.NewProc("wevtapi.dll", "EvtOpenPublisherMetadata")
	procEvtQuery                        = winapi.NewProc("wevtapi.dll", "EvtQuery")
	procEvtRender                       = winapi.NewProc("wevtapi.dll", "EvtRender")
	procEvtSeek                         = winapi.NewProc("wevtapi.dll", "EvtSeek")
	procEvtSubscribe                    = winapi.NewProc("wevtapi.dll", "EvtSubscribe")
	procEvtUpdateBookmark               = winapi.NewProc("wevtapi.dll", "EvtUpdateBookmark")
)

func evtClearLog(session EvtHandle, channelPath, targetFilePath *uint16, flags uint32) error {
	return procEvtClearLog.Call(
		uintptr(session),
		uintptr(unsafe.Pointer(channelPath)),
		uintptr(unsafe.Pointer(targetFilePath)),
		uintptr(flags),
	)
}

func evtClose(object EvtHandle) error {
	return procEvtClose.Call(uintptr(object))
}

func evtCreateBookmark(bookmarkXML *uint16) (EvtHandle, error) {
	h, err := procEvtCreateBookmark.CallRet(uintptr(unsafe.Pointer(bookmarkXML)))
	if err != nil {
		return 0, err
	}
	return EvtHandle(h), nil
}

func evtCreateRenderContext(valuePathsCount uint32, valuePaths **uint16, flags EvtRenderContextFlag) (EvtHandle, error) {
	h, err := procEvtCreateRenderContext.CallRet(
		uintptr(valuePathsCount),
		uintptr(unsafe.Pointer(valuePaths)),
		uintptr(flags),
	)
	if err != nil {
		return 0, err
	}
	return EvtHandle(h), nil
}

func evtExportLog(session EvtHandle, path, targetFilePath, query *uint16, flags EVT_EXPORTLOG_FLAGS) error {
	return procEvtExportLog.Call(
		uintptr(session),
		uintptr(unsafe.Pointer(path)),
		uintptr(unsafe.Pointer(targetFilePath)),
		uintptr(unsafe.Pointer(query)),
		uintptr(flags),
	)
}

func evtFormatMessage(publisherMetadata, event EvtHandle, messageID uint32, valueCount uint32, values *EvtVariant, flags EvtFormatMessageFlag, bufferSize uint32, buffer *byte, bufferUsed *uint32) error {
	return procEvtFormatMessage.Call(
		uintptr(publisherMetadata),
		uintptr(event),
		uintptr(messageID),
		uintptr(valueCount),
		uintptr(unsafe.Pointer(values)),
		uintptr(flags),
		uintptr(bufferSize),
		uintptr(unsafe.Pointer(buffer)),
		uintptr(unsafe.Pointer(bufferUsed)),
	)
}

func evtGetEventMetadataProperty(eventMetadata EvtHandle, propertyID EvtEventMetadataPropertyID, flags uint32, bufferSize uint32, variant *EvtVariant, bufferUsed *uint32) error {
	return procEvtGetEventMetadataProperty.Call(
		uintptr(eventMetadata),
		uintptr(propertyID),
		uintptr(flags),
		uintptr(bufferSize),
		uintptr(unsafe.Pointer(variant)),
		uintptr(unsafe.Pointer(bufferUsed)),
	)
}

func evtGetObjectArrayProperty(objectArray EvtObjectArrayPropertyHandle, propertyID EvtPublisherMetadataPropertyID, arrayIndex uint32, flags uint32, bufferSize uint32, evtVariant *EvtVariant, bufferUsed *uint32) error {
	return procEvtGetObjectArrayProperty.Call(
		uintptr(objectArray),
		uintptr(propertyID),
		uintptr(arrayIndex),
		uintptr(flags),
		uintptr(bufferSize),
		uintptr(unsafe.Pointer(evtVariant)),
		uintptr(unsafe.Pointer(bufferUsed)),
	)
}

func evtGetObjectArraySize(objectArray EvtObjectArrayPropertyHandle, arraySize *uint32) error {
	return procEvtGetObjectArraySize.Call(
		uintptr(objectArray),
		uintptr(unsafe.Pointer(arraySize)),
	)
}

func evtGetPublisherMetadataProperty(publisherMetadata EvtHandle, propertyID EvtPublisherMetadataPropertyID, flags uint32, bufferSize uint32, variant *EvtVariant, bufferUsed *uint32) error {
	return procEvtGetPublisherMetadataProperty.Call(
		uintptr(publisherMetadata),
		uintptr(propertyID),
		uintptr(flags),
		uintptr(bufferSize),
		uintptr(unsafe.Pointer(variant)),
		uintptr(unsafe.Pointer(bufferUsed)),
	)
}

func evtNext(resultSet EvtHandle, eventArraySize uint32, eventArray *EvtHandle, timeout uint32, flags uint32, numReturned *uint32) error {
	return procEvtNext.Call(
		uintptr(resultSet),
		uintptr(eventArraySize),
		uintptr(unsafe.Pointer(eventArray)),
		uintptr(timeout),
		uintptr(flags),
		uintptr(unsafe.Pointer(numReturned)),
	)
}

func evtNextChannelPath(channelEnum EvtHandle, bufferSize uint32, buffer *uint16, bufferUsed *uint32) error {
	return procEvtNextChannelPath.Call(
		uintptr(channelEnum),
		uintptr(bufferSize),
		uintptr(unsafe.Pointer(buffer)),
		uintptr(unsafe.Pointer(bufferUsed)),
	)
}

func evtNextEventMetadata(enumerator EvtHandle, flags uint32) (EvtHandle, error) {
	h, err := procEvtNextEventMetadata.CallRet(uintptr(enumerator), uintptr(flags))
	if err != nil {
		return 0, err
	}
	return EvtHandle(h), nil
}

func evtNextPublisherId(enumerator EvtHandle, bufferSize uint32, buffer *uint16, bufferUsed *uint32) error {
	return procEvtNextPublisherId.Call(
		uintptr(enumerator),
		uintptr(bufferSize),
		uintptr(unsafe.Pointer(buffer)),
		uintptr(unsafe.Pointer(bufferUsed)),
	)
}

func evtOpenChannelEnum(session EvtHandle, flags uint32) (EvtHandle, error) {
	h, err := procEvtOpenChannelEnum.CallRet(uintptr(session), uintptr(flags))
	if err != nil {
		return 0, err
	}
	return EvtHandle(h), nil
}

func evtOpenEventMetadataEnum(publisherMetadata EvtHandle, flags uint32) (EvtHandle, error) {
	h, err := procEvtOpenEventMetadataEnum.CallRet(uintptr(publisherMetadata), uintptr(flags))
	if err != nil {
		return 0, err
	}
	return EvtHandle(h), nil
}

func evtOpenLog(session EvtHandle, path *uint16, flags uint32) (EvtHandle, error) {
	h, err := procEvtOpenLog.CallRet(
		uintptr(session),
		uintptr(unsafe.Pointer(path)),
		uintptr(flags),
	)
	if err != nil {
		return 0, err
	}
	return EvtHandle(h), nil
}

func evtOpenPublisherEnum(session EvtHandle, flags uint32) (EvtHandle, error) {
	h, err := procEvtOpenPublisherEnum.CallRet(uintptr(session), uintptr(flags))
	if err != nil {
		return 0, err
	}
	return EvtHandle(h), nil
}

func evtOpenPublisherMetadata(session EvtHandle, publisherIdentity, logFilePath *uint16, locale uint32, flags uint32) (EvtHandle, error) {
	h, err := procEvtOpenPublisherMetadata.CallRet(
		uintptr(session),
		uintptr(unsafe.Pointer(publisherIdentity)),
		uintptr(unsafe.Pointer(logFilePath)),
		uintptr(locale),
		uintptr(flags),
	)
	if err != nil {
		return 0, err
	}
	return EvtHandle(h), nil
}

func evtQuery(session EvtHandle, path, query *uint16, flags uint32) (EvtHandle, error) {
	h, err := procEvtQuery.CallRet(
		uintptr(session),
		uintptr(unsafe.Pointer(path)),
		uintptr(unsafe.Pointer(query)),
		uintptr(flags),
	)
	if err != nil {
		return 0, err
	}
	return EvtHandle(h), nil
}

func evtRender(context, fragment EvtHandle, flags EvtRenderFlag, bufferSize uint32, buffer *byte, bufferUsed, propertyCount *uint32) error {
	return procEvtRender.Call(
		uintptr(context),
		uintptr(fragment),
		uintptr(flags),
		uintptr(bufferSize),
		uintptr(unsafe.Pointer(buffer)),
		uintptr(unsafe.Pointer(bufferUsed)),
		uintptr(unsafe.Pointer(propertyCount)),
	)
}

func evtSeek(resultSet EvtHandle, position int64, bookmark EvtHandle, timeout uint32, flags uint32) error {
	return procEvtSeek.Call(
		uintptr(resultSet),
		uintptr(position),
		uintptr(bookmark),
		uintptr(timeout),
		uintptr(flags),
	)
}

func evtSubscribe(session EvtHandle, signalEvent uintptr, channelPath, query *uint16, bookmark EvtHandle, context uintptr, callback syscall.Handle, flags EvtSubscribeFlag) (EvtHandle, error) {
	h, err := procEvtSubscribe.CallRet(
		uintptr(session),
		uintptr(signalEvent),
		uintptr(unsafe.Pointer(channelPath)),
		uintptr(unsafe.Pointer(query)),
		uintptr(bookmark),
		uintptr(context),
		uintptr(callback),
		uintptr(flags),
	)
	if err != nil {
		return 0, err
	}
	return EvtHandle(h), nil
}

func evtUpdateBookmark(bookmark, event EvtHandle) error {
	return procEvtUpdateBookmark.Call(uintptr(bookmark), uintptr(event))
}

// IsAvailable 检查wevtapi DLL是否可加载。
func IsAvailable() (bool, error) {
	err := windows.NewLazySystemDLL("wevtapi.dll").Load()
	if err != nil {
		return false, err
	}
	return true, nil
}

// EvtQuery 对通道或日志文件执行查询。
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
	RequiredSize int
}

func (e InsufficientBufferError) Error() string {
	return fmt.Sprintf("insufficient buffer, need %d bytes", e.RequiredSize)
}

// CreateBookmarkFromXML 从XML创建书签。
func CreateBookmarkFromXML(bookmarkXML string) (EvtHandle, error) {
	utf16, err := syscall.UTF16PtrFromString(bookmarkXML)
	if err != nil {
		return 0, err
	}
	return evtCreateBookmark(utf16)
}

// CreateRenderContext 创建渲染上下文。
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
func OpenPublisherMetadata(session EvtHandle, publisherName string, lang uint32) (EvtHandle, error) {
	p, err := syscall.UTF16PtrFromString(publisherName)
	if err != nil {
		return 0, err
	}
	return evtOpenPublisherMetadata(session, p, nil, lang, 0)
}

// Channels 列出所有已注册的通道。
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
	aligned := offset & ^uintptr(1)
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
