//go:build windows

package dns

import (
	"golang.org/x/sys/windows"
)

var moddnsapi = windows.NewLazySystemDLL("dnsapi.dll")

var procDnsGetCacheDataTable = moddnsapi.NewProc("DnsGetCacheDataTable")

// dnsRecord 对应 Windows DNS_RECORDW 结构体（64 位）。
type dnsRecord struct {
	pNext      *dnsRecord // 0: 链表下一个记录
	pName      *uint16    // 8: 域名指针
	DataType   uint16     // 16: 记录类型（1=A, 28=AAAA）
	DataLength uint16     // 18: 数据长度
	_          [4]byte    // 20: Flags
	dwTtl      uint32     // 24: TTL（秒）
	_          [4]byte    // 28: dwReserved
	// Data at offset 32
}

// dnsCacheEntry 对应 DNS 缓存表条目（未公开结构）。
type dnsCacheEntry struct {
	_       [8]byte    // 0-7
	Name    *uint16    // 8-15: 域名指针
	_       [8]byte    // 16-23
	Entries *dnsRecord // 24-31: DNS_RECORD 链表指针
}

// dnsCacheTable 对应 DNS 缓存表头（未公开结构）。
type dnsCacheTable struct {
	Version uint32         // 0-3: 版本
	Count   uint32         // 4-7: 条目数
	Entries *dnsCacheEntry // 8-15: 条目数组指针
}
