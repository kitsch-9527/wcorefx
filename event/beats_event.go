package event

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/elastic/beats/v7/winlogbeat/sys"
	"github.com/elastic/beats/v7/winlogbeat/sys/winevent"
	winevelog "github.com/elastic/beats/v7/winlogbeat/sys/wineventlog"
	"golang.org/x/sys/windows"
)

// 常量定义（补充原代码缺失的常量）
const (
	defaultMaxRead   = 100     // 单次读取最大事件数
	renderBufferSize = 1 << 14 // 16KB 事件渲染缓冲区大小
)

// Record 封装事件日志记录及其元数据（补全字段赋值）
type Record struct {
	winevent.Event
	API string // 用于读取记录的事件日志API类型
	XML string // 事件的XML表示形式
}

// EventLog 定义Windows事件日志操作接口（保持原接口设计）
type EventLog interface {
	Open(recordNumber uint64) error // 打开日志，从指定记录后读取
	Read() ([]Record, error)        // 读取事件记录
	Close() error                   // 关闭日志资源
}

// winEventLog 实现EventLog接口（优化字段命名与状态管理）
type winEventLog struct {
	query        string              // 查询条件字符串
	target       string              // 目标（通道名或.evtx文件路径）
	isFileLog    bool                // 是否为文件日志（区分订阅策略）
	isFirstQuery bool                // 首次查询标记（替代全局变量）
	subscription winevelog.EvtHandle // 订阅/查询句柄
	maxRead      int                 // 单次最大读取数
	lastRead     uint64              // 最后读取的记录号
	renderBuf    []byte              // 渲染缓冲区
	outputBuf    *sys.ByteBuffer     // XML输出缓冲区
}

// 确保winEventLog实现EventLog接口（编译期校验）
var _ EventLog = &winEventLog{}

// CreateBookmarkFromRecord 从记录编号创建书签（保留核心逻辑，优化错误提示）
func CreateBookmarkFromRecord(channelName string, recordNumber uint64) (winevelog.Bookmark, error) {
	switch {
	case channelName == "":
		return 0, fmt.Errorf("bookmark: 通道名称不能为空")
	case recordNumber == 0:
		return 0, fmt.Errorf("bookmark: 记录编号必须大于0")
	}

	bookmarkXML := fmt.Sprintf(`<BookmarkList>
  <Bookmark Channel='%s' RecordId='%d' IsCurrent='true'/>
</BookmarkList>`, channelName, recordNumber)
	return winevelog.NewBookmarkFromXML(bookmarkXML)
}

// isFileLog 判断目标是否为本地.evtx文件（新增：日志类型判断）
func isFileLog(target string) bool {
	if target == "" {
		return false
	}
	info, err := os.Stat(target)
	return err == nil && !info.IsDir() && len(target) >= 5 && target[len(target)-5:] == ".evtx"
}

// newWinEventLog 创建事件日志读取器（优化参数与初始化逻辑）
func newWinEventLog(target, eventID string) (EventLog, error) {
	// 首次查询逻辑：本地文件默认全量读取，通道日志默认读取2年历史
	isFile := isFileLog(target)
	var ignoreOlder time.Duration
	if isFile {
		ignoreOlder = 0 // 文件日志忽略时间限制（全量读取）
	} else {
		ignoreOlder = 732 * 24 * time.Hour //天数 * 24 * time.Hour // 通道日志首次读取2年历史
	}

	// 构建查询条件（精简参数传递）
	query, err := winevelog.Query{
		Log:         target,
		IgnoreOlder: ignoreOlder,
		EventID:     eventID,
		Provider:    []string{},
	}.Build()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	// 初始化读取器（合并缓冲区初始化逻辑）
	return &winEventLog{
		query:        query,
		target:       target,
		isFileLog:    isFile,
		isFirstQuery: !isFile, // 仅通道日志启用首次查询逻辑
		maxRead:      defaultMaxRead,
		renderBuf:    make([]byte, renderBufferSize),
		outputBuf:    sys.NewByteBuffer(renderBufferSize),
	}, nil
}

// Open 打开日志（核心优化：根据日志类型选择订阅策略）
func (l *winEventLog) Open(recordNumber uint64) error {
	// 1. 文件日志：使用EvtQuery读取本地文件
	if l.isFileLog {
		handle, err := winevelog.EvtQuery(
			0,                          // 本地主机
			l.target,                   // .evtx文件路径
			l.query,                    // 查询条件
			winevelog.EvtQueryFilePath, // 文件查询标志
		)
		if err != nil {
			return fmt.Errorf("evtquery file: %w", err)
		}
		l.subscription = handle
		return nil
	}

	// 2. 通道日志：使用EvtSubscribe订阅实时+历史
	// 创建书签（标记读取起始位置）
	bookmark, err := CreateBookmarkFromRecord(l.target, recordNumber)
	if err != nil {
		return fmt.Errorf("create bookmark: %w", err)
	}
	defer bookmark.Close()

	// 创建信号事件（订阅通知）
	signalEvent, err := windows.CreateEvent(nil, 0, 0, nil)
	if err != nil {
		return fmt.Errorf("create signal event: %w", err)
	}

	// 创建订阅句柄
	handle, err := winevelog.Subscribe(
		0,                                        // 本地主机
		signalEvent,                              // 通知信号
		l.target,                                 // 通道名
		l.query,                                  // 查询条件
		winevelog.EvtHandle(bookmark),            // 起始书签
		winevelog.EvtSubscribeStartAfterBookmark, // 从书签后读取
	)
	if err != nil {
		return fmt.Errorf("evtsubscribe channel: %w", err)
	}
	l.subscription = handle
	return nil
}

// Read 读取事件记录（精简错误处理与渲染逻辑）
func (l *winEventLog) Read() ([]Record, error) {
	// 获取事件句柄（合并重试逻辑）
	handles, err := l.getEventHandles()
	if err != nil {
		return nil, fmt.Errorf("get handles: %w", err)
	}
	if len(handles) == 0 {
		return nil, nil
	}

	// 延迟关闭所有句柄（确保资源释放）
	defer func() {
		for _, h := range handles {
			if err := winevelog.Close(h); err != nil {
				log.Printf("warn: close event handle: %v", err)
			}
		}
	}()

	// 解析事件为Record（精简循环逻辑）
	var records []Record
	for _, h := range handles {
		l.outputBuf.Reset()
		renderErr := winevelog.RenderEvent(h, 0, l.renderBuf, nil, l.outputBuf)

		// 处理缓冲区不足（精简重试逻辑）
		if bufErr, ok := renderErr.(sys.InsufficientBufferError); ok {
			l.renderBuf = make([]byte, bufErr.RequiredSize)
			renderErr = winevelog.RenderEvent(h, 0, l.renderBuf, nil, l.outputBuf)
		}

		// 构建记录（补全API/XML字段）
		xmlData := l.outputBuf.Bytes()
		event, err := winevent.UnmarshalXML(xmlData)
		if err != nil {
			log.Printf("warn: unmarshal xml: %v (xml: %s)", err, xmlData)
			continue
		}

		// 补充事件元数据
		winevent.PopulateAccount(&event.User)
		if event.Level == "" {
			event.Level = winevelog.EventLevel(event.LevelRaw).String()
		}

		// 记录最后读取位置
		records = append(records, Record{
			Event: event,
			API:   "wineventlog",
			XML:   string(xmlData),
		})
		l.lastRead = event.RecordID
	}

	// 首次查询后切换为增量读取（更新状态）
	if l.isFirstQuery {
		l.isFirstQuery = false
		l.query, _ = winevelog.Query{ // 重新构建1分钟增量查询
			Log:         l.target,
			IgnoreOlder: 60 * time.Second,
			EventID:     l.query, // 复用原事件ID过滤
			Provider:    []string{},
		}.Build()
	}
	return records, nil
}

// todo 需要添加分页设计协程改造调用Format 函数 格式化保留按照时间排序
func (l *winEventLog) getEventHandles() ([]winevelog.EvtHandle, error) {
	var allHandles []winevelog.EvtHandle // 用于累积所有句柄

	for {
		handles, err := winevelog.EventHandles(l.subscription, l.maxRead)
		switch {
		case err == nil:
			// 成功获取一批句柄，追加到结果集
			allHandles = append(allHandles, handles...)
			// 如果本次获取数量小于maxRead，可能已接近末尾，继续循环确认
			if len(handles) < l.maxRead {
				continue
			}

		case err == winevelog.ERROR_NO_MORE_ITEMS:
			// 所有句柄已获取完成，退出循环
			return allHandles, nil

		case err == winevelog.RPC_S_INVALID_BOUND:
			// 处理边界错误：关闭重连并减半读取量，然后继续获取剩余句柄
			if err := l.Close(); err != nil {
				return nil, fmt.Errorf("reconnect close: %w", err)
			}
			if err := l.Open(l.lastRead); err != nil {
				return nil, fmt.Errorf("reconnect open: %w", err)
			}
			l.maxRead /= 2
			// 确保maxRead不会小于1（避免无效读取）
			if l.maxRead < 1 {
				l.maxRead = 1
			}
			// 继续循环获取剩余句柄

		default:
			// 其他错误，返回已获取的句柄和错误信息
			return allHandles, fmt.Errorf("evt handles err: %w", err)
		}
	}
}

// Close 关闭日志资源（保持简洁）
func (l *winEventLog) Close() error {
	if l.subscription != 0 {
		return winevelog.Close(l.subscription)
	}
	return nil
}

// getEvents 读取指定事件ID的日志（精简冗余代码）
func GetEvents(target, eventID string) []Record {
	// 创建读取器
	reader, err := newWinEventLog(target, eventID)
	if err != nil {
		log.Printf("warn: create reader (id: %s): %v", eventID, err)
		return nil
	}
	defer reader.Close()
	// 打开并读取
	if err := reader.Open(1); err != nil {
		log.Printf("warn: open log (id: %s): %v", eventID, err)
		return nil
	}
	records, err := reader.Read()
	if err != nil {
		log.Printf("warn: read log (id: %s): %v", eventID, err)
	}
	return records
}

type Msg struct {
	Msg              string `json:"msg"`
	Time             string `json:"time"`
	EventID          uint32 `json:"event_id"`
	ProviderName     string `json:"provider_name"`
	LevelDisplayName string `json:"level_display_name"`
}

func (s Msg) Format(records []Record, f func(r Record) bool) []Msg {
	var results []Msg
	for _, rec := range records {
		if f != nil && f(rec) {
			continue
		}
		results = append(results, Msg{
			Msg:              rec.Message,
			Time:             rec.TimeCreated.SystemTime.Local().Format("2006-01-02 15:04:05"),
			EventID:          rec.EventIdentifier.ID,
			ProviderName:     rec.Provider.Name,
			LevelDisplayName: rec.Level,
		})
	}
	return results
}
