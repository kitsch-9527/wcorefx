package ntdll

import "golang.org/x/sys/windows"

// SYSTEM_HANDLE_INFORMATION
type PSystemHandleInformation struct {
	NumberOfHandles uint32 // 32位偏移0x00，64位偏移0x00
	//Reserved        uintptr                        // 32位偏移0x04，64位偏移0x08
	Handles [1]SystemHandleTableEntryInfo // 32位偏移0x08，64位偏移0x10（变长数组，这里以[1]表示）
}

// SYSTEM_HANDLE_TABLE_ENTRY_INFO
type SystemHandleTableEntryInfo struct {
	UniqueProcessId       uint16  // USHORT 对应16位无符号整数（进程ID）
	CreatorBackTraceIndex uint16  // USHORT 对应16位无符号整数（创建者回溯索引）
	ObjectTypeIndex       byte    // UCHAR 对应8位无符号整数（对象类型索引）
	HandleAttributes      byte    // UCHAR 对应8位无符号整数（句柄属性）
	HandleValue           uint16  // USHORT 对应16位无符号整数（句柄值）
	Object                uintptr // PVOID 对应64位指针（对象地址）
	GrantedAccess         windows.ACCESS_MASK
}

// OBJECT_INFORMATION_CLASS 对应 C 中的 _OBJECT_INFORMATION_CLASS 枚举
// 用于指定获取或设置对象信息的类型
type OBJECT_INFORMATION_CLASS uint32

const (
	// ObjectBasicInformation 获取对象的基本信息
	// 对应查询类型: OBJECT_BASIC_INFORMATION
	ObjectBasicInformation OBJECT_INFORMATION_CLASS = iota

	// ObjectNameInformation 获取对象的名称信息
	// 对应查询类型: OBJECT_NAME_INFORMATION
	ObjectNameInformation

	// ObjectTypeInformation 获取对象的类型信息
	// 对应查询类型: OBJECT_TYPE_INFORMATION
	ObjectTypeInformation

	// ObjectTypesInformation 获取对象类型的集合信息
	// 对应查询类型: OBJECT_TYPES_INFORMATION
	ObjectTypesInformation

	// ObjectHandleFlagInformation 获取或设置对象句柄的标志信息
	// 对应查询/设置类型: OBJECT_HANDLE_FLAG_INFORMATION
	ObjectHandleFlagInformation

	// ObjectSessionInformation 更改对象的会话信息
	// 需要 SeTcbPrivilege 权限
	ObjectSessionInformation

	// ObjectSessionObjectInformation 更改对象的会话对象信息
	// 需要 SeTcbPrivilege 权限
	ObjectSessionObjectInformation

	// ObjectSetRefTraceInformation 自 25H2 版本开始支持
	ObjectSetRefTraceInformation

	// MaxObjectInfoClass 枚举的最大值，用于范围检查
	MaxObjectInfoClass
)

// PUBLIC_OBJECT_TYPE_INFORMATION结构体定义
type PUBLIC_OBJECT_TYPE_INFORMATION struct {
	TypeName UNICODE_STRING
	Reserved [22]uint8 // 预留字段，根据系统架构可能有所不同
}

// UNICODE_STRING结构体定义
type UNICODE_STRING struct {
	Length        uint16
	MaximumLength uint16
	Buffer        *uint16
}
