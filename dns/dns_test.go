//go:build windows

package dns

import (
	"strings"
	"testing"
)

func TestCache(t *testing.T) {
	entries, err := Cache()
	if err != nil {
		if strings.Contains(err.Error(), "Incorrect function") {
			t.Skip("DnsGetCacheDataTable not supported on this Windows version")
		}
		t.Fatalf("Cache() failed: %v", err)
	}
	t.Logf("Found %d DNS cache entries", len(entries))
	for _, e := range entries {
		if e.Name != "" {
			t.Logf("  %s -> %s (type=%d TTL=%d)", e.Name, e.Data, e.Type, e.TTL)
			break
		}
	}
}
