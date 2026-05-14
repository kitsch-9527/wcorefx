//go:build windows

package dns

import "github.com/kitsch-9527/wcorefx/internal/winapi"

var procDnsGetCacheDataTable = winapi.NewProc("dnsapi.dll", "DnsGetCacheDataTable", winapi.ConvErrnoReturn)

// dnsRecord 对应 Windows DNS_RECORDW 结构体（64 位）。
type dnsRecord struct {
	pNext      *dnsRecord
	pName      *uint16
	DataType   uint16
	DataLength uint16
	_          [4]byte
	dwTtl      uint32
	_          [4]byte
}

// dnsCacheEntry 对应 DNS 缓存表条目（未公开结构）。
type dnsCacheEntry struct {
	_       [8]byte
	Name    *uint16
	_       [8]byte
	Entries *dnsRecord
}

// dnsCacheTable 对应 DNS 缓存表头（未公开结构）。
type dnsCacheTable struct {
	Version uint32
	Count   uint32
	Entries *dnsCacheEntry
}
