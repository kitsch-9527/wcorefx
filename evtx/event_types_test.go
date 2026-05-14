//go:build windows

package evtx

import (
	"testing"

	"github.com/kitsch-9527/wcorefx/internal/winapi"
)

func TestSIDString(t *testing.T) {
	s := SID{Identifier: "S-1-5-18", Name: "SYSTEM", Domain: "NT AUTHORITY", Type: SidTypeUser}
	got := s.String()
	if got == "" {
		t.Error("SID.String() returned empty")
	}
	t.Logf("SID: %s", got)
}

func TestSIDTypeString(t *testing.T) {
	tests := []struct {
		st   SIDType
		want string
	}{
		{SidTypeUser, "User"},
		{SidTypeGroup, "Group"},
		{SidTypeDomain, "Domain"},
		{SidTypeWellKnownGroup, "Well Known Group"},
		{SidTypeUnknown, "Unknown"},
		{SIDType(0), ""},
		{SIDType(99), "99"},
	}
	for _, tt := range tests {
		got := tt.st.String()
		if got != tt.want {
			t.Errorf("SIDType(%d).String() = %q, want %q", tt.st, got, tt.want)
		}
	}
}

func TestRemoveWindowsLineEndings(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello\r\nworld", "hello\nworld"},
		{"no change", "no change"},
		{"\r\n", ""},
		{"trailing\n", "trailing"},
	}
	for _, tt := range tests {
		got := RemoveWindowsLineEndings(tt.input)
		if got != tt.want {
			t.Errorf("RemoveWindowsLineEndings(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestEventString(t *testing.T) {
	e := Event{
		EventIdentifier: EventIdentifier{ID: 4624},
		Provider:        Provider{Name: "Microsoft-Windows-Security-Auditing"},
	}
	s := e.String()
	if s == "" {
		t.Error("Event.String() returned empty")
	}
	t.Logf("Event: %s", s)
}

func TestInsufficientBufferError(t *testing.T) {
	err := &winapi.ErrInsufficientBuffer{Size: 1024}
	if err.Error() == "" {
		t.Error("ErrInsufficientBuffer.Error() returned empty")
	}
}

func TestDefaultWinMeta(t *testing.T) {
	if defaultWinMeta == nil {
		t.Fatal("defaultWinMeta is nil")
	}
	if len(defaultWinMeta.Keywords) == 0 {
		t.Error("defaultWinMeta.Keywords is empty")
	}
	if len(defaultWinMeta.Levels) == 0 {
		t.Error("defaultWinMeta.Levels is empty")
	}
	if len(defaultWinMeta.Tasks) == 0 {
		t.Error("defaultWinMeta.Tasks is empty")
	}
	// Verify known values
	if defaultWinMeta.Levels[2] != "Error" {
		t.Errorf("Levels[2] = %q, want Error", defaultWinMeta.Levels[2])
	}
	if defaultWinMeta.Keywords[0x10000000000000] != "Audit Failure" {
		t.Errorf("Keywords[AuditFailure] = %q", defaultWinMeta.Keywords[0x10000000000000])
	}
}

