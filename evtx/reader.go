//go:build windows

package evtx

import (
	"fmt"
	"log"
	"os"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"

	"github.com/kitsch-9527/wcorefx/event"
)

const (
	defaultMaxRead   = 100
	renderBufferSize = 1 << 14 // 16KB
)

// Reader 读取Windows事件日志事件。
type Reader struct {
	query        string
	target       string
	isFile       bool
	isFirstQuery bool
	subscription EvtHandle
	maxRead      int
	lastRead     uint64
	renderBuf    []byte
	outputBuf    *ByteBuffer
}

// NewReader 创建一个新的事件日志读取器。
//   target - 通道名称或.evtx文件路径
//   eventID - 要过滤的事件ID字符串，为空时不过滤
//   返回1 - 新创建的Reader指针
//   返回2 - 创建过程中的错误，成功时为nil
func NewReader(target, eventID string) (*Reader, error) {
	isFile := isFileLog(target)
	var ignoreOlder time.Duration
	if isFile {
		ignoreOlder = 0
	} else {
		ignoreOlder = 732 * 24 * time.Hour
	}

	q, err := buildQuery(target, ignoreOlder, eventID)
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	return &Reader{
		query:        q,
		target:       target,
		isFile:       isFile,
		isFirstQuery: !isFile,
		maxRead:      defaultMaxRead,
		renderBuf:    make([]byte, renderBufferSize),
		outputBuf:    NewByteBuffer(renderBufferSize),
	}, nil
}

// Open 从指定记录编号开始打开日志。
//   recordNumber - 要开始读取的记录编号，用于创建书签
//   返回 - 打开过程中的错误，成功时为nil
func (r *Reader) Open(recordNumber uint64) error {
	if r.isFile {
		handle, err := EvtQuery(0, r.target, r.query, EvtQueryFilePath)
		if err != nil {
			return fmt.Errorf("evtquery file: %w", err)
		}
		r.subscription = handle
		return nil
	}

	bookmarkXML := fmt.Sprintf(`<BookmarkList>
  <Bookmark Channel='%s' RecordId='%d' IsCurrent='true'/>
</BookmarkList>`, r.target, recordNumber)

	bookmark, err := CreateBookmarkFromXML(bookmarkXML)
	if err != nil {
		return fmt.Errorf("create bookmark: %w", err)
	}
	defer bookmark.Close()

	signalEvent, err := windows.CreateEvent(nil, 0, 0, nil)
	if err != nil {
		return fmt.Errorf("create signal event: %w", err)
	}

	handle, err := Subscribe(0, signalEvent, r.target, r.query, bookmark, EvtSubscribeStartAfterBookmark)
	if err != nil {
		return fmt.Errorf("evtsubscribe: %w", err)
	}
	r.subscription = handle
	return nil
}

// Read 从日志中读取事件。
//   返回1 - 事件记录切片
//   返回2 - 读取过程中的错误，成功时为nil
func (r *Reader) Read() ([]event.Record, error) {
	handles, err := r.getEventHandles()
	if err != nil {
		return nil, fmt.Errorf("get handles: %w", err)
	}
	if len(handles) == 0 {
		return nil, nil
	}

	defer func() {
		for _, h := range handles {
			h.Close()
		}
	}()

	var records []event.Record
	for _, h := range handles {
		r.outputBuf.Reset()
		err := RenderEventXML(h, r.renderBuf, r.outputBuf)
		if bufErr, ok := err.(InsufficientBufferError); ok {
			r.renderBuf = make([]byte, bufErr.RequiredSize)
			err = RenderEventXML(h, r.renderBuf, r.outputBuf)
		}
		if err != nil {
			log.Printf("warn: render event: %v", err)
			continue
		}

		xmlData := r.outputBuf.Bytes()
		evt, err := event.UnmarshalXML(xmlData)
		if err != nil {
			log.Printf("warn: unmarshal xml: %v", err)
			continue
		}

		event.PopulateAccount(&evt.User)

		records = append(records, event.Record{
			Event: evt,
			API:   "evtx",
			XML:   string(xmlData),
		})
		r.lastRead = evt.RecordID
	}

	if r.isFirstQuery {
		r.isFirstQuery = false
		q, err := buildQuery(r.target, 60*time.Second, r.query)
		if err == nil {
			r.query = q
		}
	}
	return records, nil
}

// Close 关闭读取器。
//   返回 - 关闭过程中的错误，成功时为nil
func (r *Reader) Close() error {
	if r.subscription != 0 {
		return r.subscription.Close()
	}
	return nil
}

func (r *Reader) getEventHandles() ([]EvtHandle, error) {
	var allHandles []EvtHandle
	for {
		handles, err := EventHandles(r.subscription, r.maxRead)
		switch {
		case err == nil:
			allHandles = append(allHandles, handles...)
			if len(handles) < r.maxRead {
				continue
			}
		case err == ERROR_NO_MORE_ITEMS:
			return allHandles, nil
		case err == RPC_S_INVALID_BOUND:
			if closeErr := r.Close(); closeErr != nil {
				return nil, fmt.Errorf("reconnect close: %w", closeErr)
			}
			if openErr := r.Open(r.lastRead); openErr != nil {
				return nil, fmt.Errorf("reconnect open: %w", openErr)
			}
			r.maxRead /= 2
			if r.maxRead < 1 {
				r.maxRead = 1
			}
		default:
			return allHandles, fmt.Errorf("evt handles: %w", err)
		}
	}
}

// GetEvents 读取匹配指定eventID的事件列表。
//   target - 通道名称或.evtx文件路径
//   eventID - 要过滤的事件ID字符串
//   返回 - 事件记录切片，出错或未找到时返回nil
func GetEvents(target, eventID string) []event.Record {
	reader, err := NewReader(target, eventID)
	if err != nil {
		log.Printf("warn: create reader: %v", err)
		return nil
	}
	defer reader.Close()

	if err := reader.Open(1); err != nil {
		log.Printf("warn: open log: %v", err)
		return nil
	}
	records, err := reader.Read()
	if err != nil {
		log.Printf("warn: read log: %v", err)
	}
	return records
}

// isFileLog checks if target is a .evtx file path.
func isFileLog(target string) bool {
	if target == "" {
		return false
	}
	info, err := os.Stat(target)
	return err == nil && !info.IsDir() && len(target) >= 5 && target[len(target)-5:] == ".evtx"
}

// buildQuery constructs a WEL query XML string.
func buildQuery(logName string, ignoreOlder time.Duration, eventID string) (string, error) {
	var selects []string
	if ignoreOlder > 0 {
		ms := ignoreOlder.Nanoseconds() / int64(time.Millisecond)
		selects = append(selects,
			fmt.Sprintf("TimeCreated[timediff(@SystemTime) &lt;= %d]", ms))
	}
	if eventID != "" {
		selects = append(selects, fmt.Sprintf("EventID=%s", eventID))
	}

	selectXML := ""
	if len(selects) > 0 {
		selectXML = fmt.Sprintf("[System[%s]]", joinStrings(selects, " and "))
	}

	return fmt.Sprintf(`<QueryList>
  <Query Id="0">
    <Select Path="%s">*%s</Select>
  </Query>
</QueryList>`, logName, selectXML), nil
}

func joinStrings(elems []string, sep string) string {
	if len(elems) == 0 {
		return ""
	}
	n := len(sep) * (len(elems) - 1)
	for _, e := range elems {
		n += len(e)
	}
	b := make([]byte, n)
	bp := copy(b, elems[0])
	for _, s := range elems[1:] {
		bp += copy(b[bp:], sep)
		bp += copy(b[bp:], s)
	}
	return string(b)
}

// Bookmark 表示事件日志书签句柄。
type Bookmark EvtHandle

// Close 关闭书签句柄。
//   返回 - 关闭过程中的错误，成功时为nil
func (b Bookmark) Close() error {
	return evtClose(EvtHandle(b))
}

// NewBookmarkFromXML 从XML创建书签。
//   xml - 书签的XML字符串
//   返回1 - 书签对象
//   返回2 - 创建过程中的错误，成功时为nil
func NewBookmarkFromXML(xml string) (Bookmark, error) {
	utf16, err := windows.UTF16PtrFromString(xml)
	if err != nil {
		return 0, err
	}
	h, err := evtCreateBookmark(utf16)
	return Bookmark(h), err
}

// RenderEvent 将事件渲染为XML并包含消息字符串。
//   eventHandle - 待渲染的事件句柄
//   lang - 区域设置标识符(LCID)
//   renderBuf - 渲染用的字节缓冲区
//   out - 输出缓冲区，为nil时自动创建
//   返回 - 渲染过程中的错误，成功时为nil
func RenderEvent(eventHandle EvtHandle, lang uint32, renderBuf []byte, out *ByteBuffer) error {
	if out == nil {
		out = NewByteBuffer(renderBufferSize)
	}
	return RenderEventXML(eventHandle, renderBuf, out)
}

// FormatEventString 格式化事件消息字符串（兼容旧接口）。
//   messageFlag - 消息格式标志，指定要格式化的消息类型
//   eventHandle - 事件句柄
//   publisher - 发布者名称，当publisherHandle为0时用于打开元数据
//   publisherHandle - 发布者元数据句柄，为0时自动打开
//   lang - 区域设置标识符(LCID)
//   buffer - 消息缓冲区，为nil时自动计算所需大小
//   out - 输出缓冲区
//   返回 - 格式化过程中的错误，成功时为nil
func FormatEventString(messageFlag EvtFormatMessageFlag, eventHandle EvtHandle,
	publisher string, publisherHandle EvtHandle, lang uint32, buffer []byte, out *ByteBuffer) error {
	ph := publisherHandle
	if ph == 0 {
		var err error
		ph, err = OpenPublisherMetadata(0, publisher, lang)
		if err != nil {
			return err
		}
		defer ph.Close()
	}

	var bufferUsed uint32
	if buffer == nil {
		err := evtFormatMessage(ph, eventHandle, 0, 0, nil, messageFlag, 0, nil, &bufferUsed)
		if err != nil && err != ERROR_INSUFFICIENT_BUFFER {
			return err
		}
		bufferUsed *= 2
		buffer = make([]byte, bufferUsed)
		bufferUsed = 0
	}

	return evtFormatMessage(ph, eventHandle, 0, 0, nil, messageFlag,
		uint32(len(buffer)/2), &buffer[0], &bufferUsed)
}

// EvtOpenLog 获取通道或日志文件的句柄。
//   session - 会话句柄，0表示本地会话
//   path - 通道名称或.evtx文件路径
//   flags - 打开日志标志，指定path类型
//   返回1 - 日志句柄
//   返回2 - 打开过程中的错误，成功时为nil
func EvtOpenLog(session EvtHandle, path string, flags EvtOpenLogFlag) (EvtHandle, error) {
	var pathPtr *uint16
	var err error
	if path != "" {
		pathPtr, err = windows.UTF16PtrFromString(path)
		if err != nil {
			return 0, err
		}
	}
	return evtOpenLog(session, pathPtr, uint32(flags))
}

// Ensure unused import suppression.
var _ = unsafe.Pointer(nil)
