//go:build windows

// Package dns 提供 Windows DNS 缓存信息查询功能。
package dns

import (
	"fmt"
	"net"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// DNSEntry 表示 DNS 缓存条目。
type DNSEntry struct {
	// Name 域名
	Name string
	// Data DNS 记录数据（如 IP 地址）
	Data string
	// Type DNS 记录类型（1=A, 28=AAAA, 5=CNAME 等）
	Type uint16
	// TTL 缓存 TTL（秒）
	TTL uint32
}

// Cache 返回当前 DNS 缓存内容。
//
//	 返回1 - DNS 缓存条目列表
//	 返回2 - 错误信息
func Cache() ([]DNSEntry, error) {
	var table unsafe.Pointer
	r1, _, _ := syscall.SyscallN(procDnsGetCacheDataTable.Addr(),
		uintptr(unsafe.Pointer(&table)),
	)
	if r1 != 0 {
		return nil, fmt.Errorf("DnsGetCacheDataTable failed: %d", r1)
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
			case rec.DataType == 1 && rec.DataLength >= 4: // A 记录
				ipData := (*[4]byte)(dataPtr)
				data = fmt.Sprintf("%d.%d.%d.%d", ipData[0], ipData[1], ipData[2], ipData[3])
			case rec.DataType == 28 && rec.DataLength >= 16: // AAAA 记录
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
