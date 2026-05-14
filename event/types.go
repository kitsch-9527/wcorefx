//go:build windows

package event

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sys/windows"
)

// Event 表示Windows事件日志中的一条事件记录。
type Event struct {
	// Provider 生成该事件的提供程序信息。
	Provider        Provider        `xml:"System>Provider"`
	// EventIdentifier 标识具体的事件类型和ID。
	EventIdentifier EventIdentifier `xml:"System>EventID"`
	// Version 事件的版本号。
	Version         Version         `xml:"System>Version"`
	// LevelRaw 事件的原始严重级别数值。
	LevelRaw        uint8           `xml:"System>Level"`
	// TaskRaw 事件的原始任务类别数值。
	TaskRaw         uint16          `xml:"System>Task"`
	// OpcodeRaw 事件的原始操作码数值。
	OpcodeRaw       *uint8          `xml:"System>Opcode,omitempty"`
	// KeywordsRaw 事件的原始关键字掩码（十六进制）。
	KeywordsRaw     HexInt64        `xml:"System>Keywords"`
	// TimeCreated 事件创建的系统时间。
	TimeCreated     TimeCreated     `xml:"System>TimeCreated"`
	// RecordID 事件的唯一记录编号。
	RecordID        uint64          `xml:"System>EventRecordID"`
	// Correlation 关联的活动标识符。
	Correlation     Correlation     `xml:"System>Correlation"`
	// Execution 执行事件的进程和线程信息。
	Execution       Execution       `xml:"System>Execution"`
	// Channel 事件来源的通道名称。
	Channel         string          `xml:"System>Channel"`
	// Computer 生成事件的计算机名称。
	Computer        string          `xml:"System>Computer"`
	// User 与事件关联的安全标识符。
	User            SID             `xml:"System>Security"`
	// EventData 事件数据中的键值对。
	EventData       EventData       `xml:"EventData"`
	// UserData 用户自定义的事件数据。
	UserData        UserData        `xml:"UserData"`
	// Message 事件的格式化消息文本。
	Message         string          `xml:"RenderingInfo>Message"`
	// Level 事件级别的可读名称。
	Level           string          `xml:"RenderingInfo>Level"`
	// Task 事件任务类别的可读名称。
	Task            string          `xml:"RenderingInfo>Task"`
	// Opcode 事件操作码的可读名称。
	Opcode          string          `xml:"RenderingInfo>Opcode"`
	// Keywords 事件关键字的可读名称列表。
	Keywords        []string        `xml:"RenderingInfo>Keywords>Keyword"`
	// RenderErrorCode 渲染事件时的错误码。
	RenderErrorCode uint32          `xml:"ProcessingErrorData>ErrorCode"`
	// RenderErr 渲染事件过程中的错误信息列表。
	RenderErr       []string
}

// Record 包装Event及其读取元数据。
type Record struct {
	Event
	// API 读取该事件所使用的API名称。
	API string
	// XML 事件的原始XML数据。
	XML string
}

// Provider 标识生成事件的提供程序。
type Provider struct {
	// Name 提供程序的名称。
	Name            string `xml:"Name,attr"`
	// GUID 提供程序的全局唯一标识符。
	GUID            string `xml:"Guid,attr"`
	// EventSourceName 提供程序的事件源名称。
	EventSourceName string `xml:"EventSourceName,attr"`
}

// Correlation 包含活动标识符信息。
type Correlation struct {
	// ActivityID 当前活动的唯一标识符。
	ActivityID        string `xml:"ActivityID,attr"`
	// RelatedActivityID 相关活动的标识符。
	RelatedActivityID string `xml:"RelatedActivityID,attr"`
}

// Execution 包含执行事件的进程和线程信息。
type Execution struct {
	// ProcessID 生成事件的进程ID。
	ProcessID     uint32 `xml:"ProcessID,attr"`
	// ThreadID 生成事件的线程ID。
	ThreadID      uint32 `xml:"ThreadID,attr"`
	// ProcessorID 处理事件的处理器ID。
	ProcessorID   uint32 `xml:"ProcessorID,attr"`
	// SessionID 生成事件的会话ID。
	SessionID     uint32 `xml:"SessionID,attr"`
	// KernelTime 事件在内核模式下消耗的时间。
	KernelTime    uint32 `xml:"KernelTime,attr"`
	// UserTime 事件在用户模式下消耗的时间。
	UserTime      uint32 `xml:"UserTime,attr"`
	// ProcessorTime 事件消耗的总处理器时间。
	ProcessorTime uint32 `xml:"ProcessorTime,attr"`
}

// EventIdentifier 标识具体的事件类型。
type EventIdentifier struct {
	// Qualifiers 事件限定符。
	Qualifiers uint16 `xml:"Qualifiers,attr"`
	// ID 事件标识符编号。
	ID         uint32 `xml:",chardata"`
}

// TimeCreated 包含事件的系统创建时间。
type TimeCreated struct {
	// SystemTime 事件的系统时间。
	SystemTime time.Time
}

// UnmarshalXML 从XML中解析SystemTime属性。
//   d - XML解码器
//   start - XML起始元素
//   返回 - 解析过程中的错误，成功时为nil
func (t *TimeCreated) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	attrs := struct {
		SystemTime string `xml:"SystemTime,attr"`
		RawTime    uint64 `xml:"RawTime,attr"`
	}{}
	err := d.DecodeElement(&attrs, &start)
	if err != nil {
		return err
	}
	if attrs.SystemTime != "" {
		t.SystemTime, err = time.Parse(time.RFC3339Nano, attrs.SystemTime)
	} else if attrs.RawTime != 0 {
		err = fmt.Errorf("RawTime=%d not supported", attrs.RawTime)
	}
	return err
}

// EventData 包含事件数据的键值对。
type EventData struct {
	// Pairs 事件数据的键值对列表。
	Pairs []KeyValue `xml:",any"`
}

// UserData 包含用户自定义的事件数据。
type UserData struct {
	// Name 用户数据的XML元素名称。
	Name  xmlName
	// Pairs 用户数据的键值对列表。
	Pairs []KeyValue
}

type xmlName struct {
	Local string
}

// KeyValue 表示事件数据或用户数据中的键值对。
type KeyValue struct {
	// Key 键值对的键名。
	Key   string
	// Value 键值对的值。
	Value string
}

// Version 表示事件版本号，支持自定义XML反序列化。
type Version uint8

// HexInt64 表示十六进制64位整数，支持自定义XML反序列化。
type HexInt64 uint64

// SID 表示Windows安全标识符（Security Identifier）。
type SID struct {
	// Identifier SID的字符串标识符。
	Identifier string `xml:"UserID,attr"`
	// Name SID对应的账户名称。
	Name       string
	// Domain SID所属的域名。
	Domain     string
	// Type SID的类型。
	Type       SIDType
}

// String 返回SID的字符串表示形式。
//   返回 - SID的格式化字符串，包含标识符、名称、域和类型
func (a SID) String() string {
	return fmt.Sprintf("SID Identifier[%s] Name[%s] Domain[%s] Type[%s]",
		a.Identifier, a.Name, a.Domain, a.Type)
}

// SIDType 标识SID的类型。
type SIDType uint32

const (
	// SidTypeUser 用户类型的SID。
	SidTypeUser SIDType = 1 + iota
	// SidTypeGroup 组类型的SID。
	SidTypeGroup
	// SidTypeDomain 域类型的SID。
	SidTypeDomain
	// SidTypeAlias 别名类型的SID。
	SidTypeAlias
	// SidTypeWellKnownGroup 已知组类型的SID。
	SidTypeWellKnownGroup
	// SidTypeDeletedAccount 已删除账户类型的SID。
	SidTypeDeletedAccount
	// SidTypeInvalid 无效类型的SID。
	SidTypeInvalid
	// SidTypeUnknown 未知类型的SID。
	SidTypeUnknown
	// SidTypeComputer 计算机类型的SID。
	SidTypeComputer
	// SidTypeLabel 标签类型的SID。
	SidTypeLabel
	// SidTypeLogonSession 登录会话类型的SID。
	SidTypeLogonSession
)

var sidTypeToString = map[SIDType]string{
	SidTypeUser:           "User",
	SidTypeGroup:          "Group",
	SidTypeDomain:         "Domain",
	SidTypeAlias:          "Alias",
	SidTypeWellKnownGroup: "Well Known Group",
	SidTypeDeletedAccount: "Deleted Account",
	SidTypeInvalid:        "Invalid",
	SidTypeUnknown:        "Unknown",
	SidTypeComputer:       "Computer",
	SidTypeLabel:          "Label",
	SidTypeLogonSession:   "Logon Session",
}

// String 返回SID类型的可读名称。
//   返回 - SID类型的文本描述，未知类型返回数字字符串
func (st SIDType) String() string {
	if typ, found := sidTypeToString[st]; found {
		return typ
	} else if st > 0 {
		return strconv.FormatUint(uint64(st), 10)
	}
	return ""
}

// WinMeta 存储来自事件发布者的元数据。
type WinMeta struct {
	// Keywords 关键字掩码到名称的映射。
	Keywords map[int64]string
	// Opcodes 操作码数值到名称的映射。
	Opcodes  map[uint8]string
	// Levels 级别数值到名称的映射。
	Levels   map[uint8]string
	// Tasks 任务数值到名称的映射。
	Tasks    map[uint16]string
}

// defaultWinMeta contains common Windows event metadata values.
var defaultWinMeta = &WinMeta{
	Keywords: map[int64]string{
		0:                "AnyKeyword",
		0x1000000000000:  "Response Time",
		0x4000000000000:  "WDI Diag",
		0x8000000000000:  "SQM",
		0x10000000000000: "Audit Failure",
		0x20000000000000: "Audit Success",
		0x40000000000000: "Correlation Hint",
		0x80000000000000: "Classic",
	},
	Opcodes: map[uint8]string{
		0: "Info",
		1: "Start",
		2: "Stop",
		3: "DCStart",
		4: "DCStop",
		5: "Extension",
		6: "Reply",
		7: "Resume",
		8: "Suspend",
		9: "Send",
	},
	Levels: map[uint8]string{
		0: "Information",
		1: "Critical",
		2: "Error",
		3: "Warning",
		4: "Information",
		5: "Verbose",
	},
	Tasks: map[uint16]string{
		0: "None",
	},
}

// PopulateAccount 通过SID查找并填充账户名称和类型。
//   sid - 待填充的SID对象，包含有效的Identifier
//   返回 - 查找过程中的错误，成功时为nil；若sid为nil或Identifier为空则直接返回nil
func PopulateAccount(sid *SID) error {
	if sid == nil || sid.Identifier == "" {
		return nil
	}
	s, err := windows.StringToSid(sid.Identifier)
	if err != nil {
		return err
	}
	account, domain, accType, err := s.LookupAccount("")
	if err != nil {
		return err
	}
	sid.Name = account
	sid.Domain = domain
	sid.Type = SIDType(accType)
	return nil
}

// RemoveWindowsLineEndings 将CRLF替换为LF并去除尾部换行符。
//   s - 输入的原始字符串
//   返回 - 处理后的字符串，所有CRLF已替换为LF，尾部换行符已去除
func RemoveWindowsLineEndings(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.TrimRight(s, "\n")
}
