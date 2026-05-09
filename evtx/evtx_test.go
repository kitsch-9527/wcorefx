//go:build windows

package evtx

import (
	"testing"
	"unsafe"
)

func TestNewByteBuffer(t *testing.T) {
	b := NewByteBuffer(1024)
	if b == nil {
		t.Fatal("NewByteBuffer returned nil")
	}
	if b.Len() != 0 {
		t.Errorf("Len = %d, want 0", b.Len())
	}
}

func TestByteBufferWrite(t *testing.T) {
	b := NewByteBuffer(16)
	n, err := b.Write([]byte("hello"))
	if err != nil {
		t.Fatal(err)
	}
	if n != 5 {
		t.Errorf("Write returned %d, want 5", n)
	}
	if b.Len() != 5 {
		t.Errorf("Len = %d, want 5", b.Len())
	}
	if string(b.Bytes()) != "hello" {
		t.Errorf("Bytes = %q, want hello", string(b.Bytes()))
	}
}

func TestByteBufferReset(t *testing.T) {
	b := NewByteBuffer(16)
	b.Write([]byte("hello"))
	b.Reset()
	if b.Len() != 0 {
		t.Errorf("Len after reset = %d, want 0", b.Len())
	}
	// Underlying buffer should be retained
	b.Write([]byte("world"))
	if string(b.Bytes()) != "world" {
		t.Errorf("Bytes after reset+write = %q, want world", string(b.Bytes()))
	}
}

func TestByteBufferGrow(t *testing.T) {
	b := NewByteBuffer(4)
	data := []byte("this is a long string that exceeds initial capacity")
	n, err := b.Write(data)
	if err != nil {
		t.Fatal(err)
	}
	if n != len(data) {
		t.Errorf("Write returned %d, want %d", n, len(data))
	}
	if b.Len() != len(data) {
		t.Errorf("Len = %d, want %d", b.Len(), len(data))
	}
}

func TestByteBufferReserve(t *testing.T) {
	b := NewByteBuffer(10)
	b.Reserve(100)
	if b.Len() != 100 {
		t.Errorf("Len after Reserve = %d, want 100", b.Len())
	}
}

func TestByteBufferPtrAt(t *testing.T) {
	b := NewByteBuffer(10)
	b.Write([]byte("hello"))
	p := b.PtrAt(0)
	if p == nil {
		t.Error("PtrAt(0) returned nil")
	}
	p = b.PtrAt(100)
	if p != nil {
		t.Error("PtrAt(100) should return nil for out of bounds")
	}
}

func TestEventLevelString(t *testing.T) {
	tests := []struct {
		level EventLevel
		want  string
	}{
		{EVENTLOG_LOGALWAYS_LEVEL, "Information"},
		{EVENTLOG_CRITICAL_LEVEL, "Critical"},
		{EVENTLOG_ERROR_LEVEL, "Error"},
		{EVENTLOG_WARNING_LEVEL, "Warning"},
		{EVENTLOG_INFORMATION_LEVEL, "Information"},
		{EVENTLOG_VERBOSE_LEVEL, "Verbose"},
		{EventLevel(99), "Level(99)"},
	}
	for _, tt := range tests {
		got := tt.level.String()
		if got != tt.want {
			t.Errorf("EventLevel(%d).String() = %q, want %q", tt.level, got, tt.want)
		}
	}
}

func TestIsAvailable(t *testing.T) {
	avail, err := IsAvailable()
	if err != nil {
		t.Fatal(err)
	}
	if !avail {
		t.Skip("wevtapi not available")
	}
}

func TestEvtVariantValueAccessors(t *testing.T) {
	var v EvtVariant

	// Test ValueAsUint64
	*(*uint64)(unsafe.Pointer(&v.Value)) = 42
	if v.ValueAsUint64() != 42 {
		t.Errorf("ValueAsUint64 = %d, want 42", v.ValueAsUint64())
	}
}

func TestBuildQuery(t *testing.T) {
	q, err := buildQuery("Application", 0, "")
	if err != nil {
		t.Fatal(err)
	}
	if q == "" {
		t.Error("buildQuery returned empty")
	}
	t.Logf("Query: %s", q)

	q, err = buildQuery("System", 3600000000000, "100")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Query with filters: %s", q)
}

func TestIsFileLog(t *testing.T) {
	if isFileLog("") {
		t.Error("isFileLog('') should be false")
	}
	if isFileLog("Application") {
		t.Error("isFileLog('Application') should be false")
	}
}
