//go:build windows

// Package dns 提供 Windows DNS 缓存信息查询功能。
package dns

import (
	"fmt"
	"net"
	"unsafe"

	"golang.org/x/sys/windows"
)

// DNSEntry 表示 DNS 缓存条目。
type DNSEntry struct {
	Name string
	Data string
	Type uint16
	TTL  uint32
}

// Cache 返回当前 DNS 缓存内容。
func Cache() ([]DNSEntry, error) {
	var table unsafe.Pointer
	_, err := procDnsGetCacheDataTable.CallRet(uintptr(unsafe.Pointer(&table)))
	if err != nil {
		return nil, fmt.Errorf("DnsGetCacheDataTable failed: %w", err)
	}
	if table == nil {
		return nil, nil
	}

	cacheTable := (*dnsCacheTable)(table)

	entries := make([]DNSEntry, 0, cacheTable.Count)
	for i := uint32(0); i < cacheTable.Count; i++ {
		entry := unsafe.Slice(cacheTable.Entries, cacheTable.Count)[i]
		name := windows.UTF16PtrToString(entry.Name)

		rec := entry.Entries
		for rec != nil {
			data := ""
			dataPtr := unsafe.Pointer(uintptr(unsafe.Pointer(rec)) + unsafe.Sizeof(dnsRecord{}))

			switch {
			case rec.DataType == 1 && rec.DataLength >= 4:
				ipData := (*[4]byte)(dataPtr)
				data = fmt.Sprintf("%d.%d.%d.%d", ipData[0], ipData[1], ipData[2], ipData[3])
			case rec.DataType == 28 && rec.DataLength >= 16:
				ipData := (*[16]byte)(dataPtr)
				data = net.IP(ipData[:]).String()
			}

			entries = append(entries, DNSEntry{
				Name: name,
				Data: data,
				Type: rec.DataType,
				TTL:  rec.dwTtl,
			})

			rec = rec.pNext
		}
	}

	return entries, nil
}
