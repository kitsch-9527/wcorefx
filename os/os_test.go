//go:build windows

package os

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestIs64(t *testing.T) {
	got := Is64()
	switch runtime.GOARCH {
	case "amd64", "arm64":
		if !got {
			t.Errorf("Is64() = false on %s, want true", runtime.GOARCH)
		}
	default: // 386 etc.
		if got {
			t.Errorf("Is64() = true on %s, want false", runtime.GOARCH)
		}
	}
}

func TestIsVistaUpper(t *testing.T) {
	if !IsVistaUpper() {
		t.Error("IsVistaUpper() = false, expected true (all modern Windows is Vista+)")
	} else {
		t.Log("System is Windows Vista or later")
	}
}

func TestMajorVersion(t *testing.T) {
	v := MajorVersion()
	if v == 0 {
		t.Error("MajorVersion() = 0, expected > 0")
	}
	t.Logf("MajorVersion() = %d", v)
}

func TestMinorVersion(t *testing.T) {
	v := MinorVersion()
	t.Logf("MinorVersion() = %d", v)
}

func TestBuildNumber(t *testing.T) {
	b := BuildNumber()
	if b == 0 {
		t.Error("BuildNumber() = 0, expected > 0")
	}
	if b < 10000 {
		t.Errorf("BuildNumber() = %d, seems too low for modern Windows", b)
	}
	t.Logf("BuildNumber() = %d", b)
}

func TestReleaseID(t *testing.T) {
	rid := ReleaseID()
	t.Logf("ReleaseID() = %q", rid)
}

func TestVersionInfo(t *testing.T) {
	vi := VersionInfo()
	if !strings.HasPrefix(vi, "Windows") {
		t.Errorf("VersionInfo() = %q, want prefix 'Windows'", vi)
	}
	t.Logf("VersionInfo() = %s", vi)
}

func TestCPUCount(t *testing.T) {
	c := CPUCount()
	if c == 0 {
		t.Error("CPUCount() = 0, expected >= 1")
	}
	t.Logf("CPUCount() = %d", c)
}

func TestTickCount(t *testing.T) {
	tc := TickCount()
	if tc == 0 {
		t.Error("TickCount() = 0, expected > 0 (system has been up for some time)")
	}
	t.Logf("TickCount() = %d ms", tc)
}

func TestStartupTime(t *testing.T) {
	now := time.Now()
	st := StartupTime()
	if st.After(now) {
		t.Errorf("StartupTime() = %v, is after now (%v)", st, now)
	}
	ref := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	if st.Before(ref) {
		t.Errorf("StartupTime() = %v, seems too old (before year 2000)", st)
	}
	t.Logf("StartupTime() = %v", st)
	t.Logf("Uptime ~= %v", now.Sub(st).Round(time.Second))
}

func TestNetBiosName(t *testing.T) {
	name, err := NetBiosName()
	if err != nil {
		t.Fatalf("NetBiosName() failed: %v", err)
	}
	if name == "" {
		t.Error("NetBiosName() returned empty string")
	}
	t.Logf("NetBiosName() = %s", name)
}

func TestHostName(t *testing.T) {
	name, err := HostName()
	if err != nil {
		t.Fatalf("HostName() failed: %v", err)
	}
	if name == "" {
		t.Error("HostName() returned empty string")
	}
	t.Logf("HostName() = %s", name)
}

func TestUserName(t *testing.T) {
	name, err := UserName()
	if err != nil {
		t.Fatalf("UserName() failed: %v", err)
	}
	if name == "" {
		t.Error("UserName() returned empty string")
	}
	t.Logf("UserName() = %s", name)
}

func TestSessionUserName(t *testing.T) {
	name, err := SessionUserName(WTS_CURRENT_SESSION)
	if err != nil {
		t.Fatalf("SessionUserName(WTS_CURRENT_SESSION) failed: %v", err)
	}
	if name == "" {
		t.Error("SessionUserName() returned empty string")
	}
	t.Logf("SessionUserName(current) = %s", name)
}

func TestWinDir(t *testing.T) {
	dir, err := WinDir()
	if err != nil {
		t.Fatalf("WinDir() failed: %v", err)
	}
	if dir == "" {
		t.Error("WinDir() returned empty string")
	}
	if !strings.Contains(dir, "Windows") {
		t.Errorf("WinDir() = %q, expected path containing 'Windows'", dir)
	}
	t.Logf("WinDir() = %s", dir)
}

func TestSystemDir(t *testing.T) {
	dir, err := SystemDir()
	if err != nil {
		t.Fatalf("SystemDir() failed: %v", err)
	}
	if dir == "" {
		t.Error("SystemDir() returned empty string")
	}
	t.Logf("SystemDir() = %s", dir)
}

func TestSystem32Dir(t *testing.T) {
	dir, err := System32Dir()
	if err != nil {
		t.Fatalf("System32Dir() failed: %v", err)
	}
	if dir == "" {
		t.Error("System32Dir() returned empty string")
	}
	if !strings.HasSuffix(dir, "System32") {
		t.Errorf("System32Dir() = %q, want suffix '\\System32'", dir)
	}
	t.Logf("System32Dir() = %s", dir)
}

func TestSyswow64Dir(t *testing.T) {
	dir, err := Syswow64Dir()
	if err != nil {
		t.Fatalf("Syswow64Dir() failed: %v", err)
	}
	if dir == "" {
		t.Error("Syswow64Dir() returned empty string")
	}
	if !strings.HasSuffix(dir, "SysWOW64") {
		t.Errorf("Syswow64Dir() = %q, want suffix '\\SysWOW64'", dir)
	}
	t.Logf("Syswow64Dir() = %s", dir)
}

func TestGetenv(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantEmp bool
	}{
		{name: "SystemRoot", key: "SystemRoot"},
		{name: "PATH", key: "PATH"},
		{name: "ComSpec", key: "ComSpec"},
		{name: "NonExistent", key: "SOME_VAR_THAT_DOES_NOT_EXIST_XYZ", wantEmp: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Getenv(tt.key)
			if tt.wantEmp {
				if got != "" {
					t.Errorf("Getenv(%q) = %q, want empty", tt.key, got)
				}
				return
			}
			if got == "" {
				t.Errorf("Getenv(%q) returned empty, want non-empty value", tt.key)
			}
			t.Logf("Getenv(%q) = %q", tt.key, got)
		})
	}
}

func TestEnviron(t *testing.T) {
	env := Environ()
	if len(env) == 0 {
		t.Fatal("Environ() returned empty map")
	}
	// Environment variable names are case-insensitive on Windows.
	var hasSystemRoot bool
	for k := range env {
		if strings.EqualFold(k, "SystemRoot") {
			hasSystemRoot = true
			break
		}
	}
	if !hasSystemRoot {
		t.Error("Environ() missing expected key 'SystemRoot'")
	}
	var hasPath bool
	for k := range env {
		if strings.EqualFold(k, "PATH") {
			hasPath = true
			break
		}
	}
	if !hasPath {
		t.Error("Environ() missing expected key 'PATH'")
	}
	t.Logf("Environ() returned %d variables", len(env))
}

func TestMemory(t *testing.T) {
	mi, err := Memory()
	if err != nil {
		t.Fatalf("Memory() failed: %v", err)
	}
	if mi.TotalPhysical == 0 {
		t.Error("TotalPhysical = 0, expected > 0")
	}
	if mi.AvailablePhysical > mi.TotalPhysical {
		t.Errorf("AvailablePhysical (%d) > TotalPhysical (%d)", mi.AvailablePhysical, mi.TotalPhysical)
	}
	if mi.UsedPhysical > mi.TotalPhysical {
		t.Errorf("UsedPhysical (%d) > TotalPhysical (%d)", mi.UsedPhysical, mi.TotalPhysical)
	}
	t.Logf("Memory: total=%d MB, avail=%d MB, used=%d MB, load=%d%%",
		mi.TotalPhysical>>20, mi.AvailablePhysical>>20, mi.UsedPhysical>>20, mi.MemoryLoad)
}

func TestCPUModel(t *testing.T) {
	model, err := CPUModel()
	if err != nil {
		t.Fatalf("CPUModel() failed: %v", err)
	}
	if model == "" {
		t.Error("CPUModel() returned empty string")
	}
	t.Logf("CPUModel() = %s", model)
}

func TestDrives(t *testing.T) {
	drives, err := Drives()
	if err != nil {
		t.Fatalf("Drives() failed: %v", err)
	}
	if len(drives) == 0 {
		t.Error("Drives() returned empty list, expected at least one drive")
	}
	for _, d := range drives {
		t.Logf("Drive: %s  Type: %s  Total: %d MB  Free: %d MB",
			d.Drive, d.Type, d.TotalBytes>>20, d.FreeBytes>>20)
	}
}

func TestDosErrorMsg(t *testing.T) {
	tests := []struct {
		code  uint32
		known bool
	}{
		{code: 0, known: true},          // ERROR_SUCCESS
		{code: 2, known: true},          // ERROR_FILE_NOT_FOUND
		{code: 5, known: true},          // ERROR_ACCESS_DENIED
		{code: 87, known: true},         // ERROR_INVALID_PARAMETER
		{code: 999999, known: false},    // unknown code
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("code_%d", tt.code), func(t *testing.T) {
			got := DosErrorMsg(tt.code)
			if got == "" {
				t.Error("DosErrorMsg() returned empty string")
			}
			if tt.known && strings.HasPrefix(got, "unknown error") {
				t.Errorf("DosErrorMsg(%d) = %q, want known error message", tt.code, got)
			}
			if !tt.known && !strings.HasPrefix(got, "unknown error") {
				t.Errorf("DosErrorMsg(%d) = %q, want 'unknown error' prefix", tt.code, got)
			}
			t.Logf("DosErrorMsg(%d) = %q", tt.code, got)
		})
	}
}

/*
func TestReboot(t *testing.T) {
	err := Reboot()
	// Reboot requires SE_SHUTDOWN_NAME privilege. Test processes won't have
	// this unless running as Administrator, so we just verify the call doesn't
	// panic and that err is either nil or an expected access-denied error.
	if err != nil {
		t.Logf("Reboot() expectedly failed (not running as admin): %v", err)
	} else {
		t.Log("Reboot() returned nil (running with sufficient privileges)")
	}
}

func TestPoweroff(t *testing.T) {
	err := Poweroff()
	if err != nil {
		t.Logf("Poweroff() expectedly failed (not running as admin): %v", err)
	} else {
		t.Log("Poweroff() returned nil (running with sufficient privileges)")
	}
}
*/
