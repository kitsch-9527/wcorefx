//go:build windows

package evtx

import (
	"bytes"
	"encoding/xml"
	"fmt"
)

// UnmarshalXML 从原始XML解析事件数据。
//   rawXML - 事件的原始XML字节数据
//   返回 - 解析后的事件对象；解析过程中的错误，成功时为nil
func UnmarshalXML(rawXML []byte) (Event, error) {
	var event Event
	decoder := xml.NewDecoder(bytes.NewReader(rawXML))
	err := decoder.Decode(&event)
	return event, err
}

// EnrichRawValuesWithNames 为原始系统属性值添加人类可读的名称。
//   publisherMeta - 事件发布者的元数据，用于查找自定义名称；可为nil
//   event - 待丰富的事件对象，其Level、Task、Opcode、Keywords字段将被填充
func EnrichRawValuesWithNames(publisherMeta *WinMeta, event *Event) {
	rawKeyword := int64(event.KeywordsRaw)

	if len(event.Keywords) == 0 {
		for mask, keyword := range defaultWinMeta.Keywords {
			if rawKeyword&mask != 0 {
				event.Keywords = append(event.Keywords, keyword)
				rawKeyword &^= mask
			}
		}
		if publisherMeta != nil {
			for mask, keyword := range publisherMeta.Keywords {
				if rawKeyword&mask != 0 {
					event.Keywords = append(event.Keywords, keyword)
					rawKeyword &^= mask
				}
			}
		}
	}

	if event.Opcode == "" && event.OpcodeRaw != nil {
		var found bool
		event.Opcode, found = defaultWinMeta.Opcodes[*event.OpcodeRaw]
		if !found && publisherMeta != nil {
			event.Opcode = publisherMeta.Opcodes[*event.OpcodeRaw]
		}
	}

	if event.Level == "" {
		var found bool
		event.Level, found = defaultWinMeta.Levels[event.LevelRaw]
		if !found && publisherMeta != nil {
			event.Level = publisherMeta.Levels[event.LevelRaw]
		}
	}

	if event.Task == "" {
		var found bool
		if publisherMeta != nil {
			event.Task, found = publisherMeta.Tasks[event.TaskRaw]
			if !found {
				event.Task = defaultWinMeta.Tasks[event.TaskRaw]
			}
		} else {
			event.Task = defaultWinMeta.Tasks[event.TaskRaw]
		}
	}
}

// Format 为Record切片添加日志格式化前缀。
// Deprecated: use direct iteration instead.
//   records - 事件记录切片
//   fn - 可选的过滤函数，返回true时跳过该记录；可为nil
//   返回 - 格式化后的结果切片，每项包含消息、时间、事件ID、提供程序名称和级别名称
func Format(records []Record, fn func(r Record) bool) []struct {
	Msg              string
	Time             string
	EventID          uint32
	ProviderName     string
	LevelDisplayName string
} {
	var results []struct {
		Msg              string
		Time             string
		EventID          uint32
		ProviderName     string
		LevelDisplayName string
	}
	for _, rec := range records {
		if fn != nil && fn(rec) {
			continue
		}
		results = append(results, struct {
			Msg              string
			Time             string
			EventID          uint32
			ProviderName     string
			LevelDisplayName string
		}{
			Msg:              rec.Message,
			Time:             rec.TimeCreated.SystemTime.Local().Format("2006-01-02 15:04:05"),
			EventID:          rec.EventIdentifier.ID,
			ProviderName:     rec.Provider.Name,
			LevelDisplayName: rec.Level,
		})
	}
	return results
}

// String 将Event转换为人类可读的字符串。
//   返回 - 包含事件ID、提供程序名称、级别和时间的格式化字符串
func (e Event) String() string {
	return fmt.Sprintf("Event ID=%d Provider=%s Level=%s Time=%s",
		e.EventIdentifier.ID, e.Provider.Name, e.Level, e.TimeCreated.SystemTime)
}

