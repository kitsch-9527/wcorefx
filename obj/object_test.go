//go:build windows

package obj

import (
	"strings"
	"testing"
)

func TestDriverList(t *testing.T) {
	drivers, err := DriverList()
	if err != nil {
		t.Fatalf("DriverList() failed: %v", err)
	}
	if len(drivers) == 0 {
		t.Fatal("DriverList() returned empty list")
	}
}

func TestDriverName(t *testing.T) {
	drivers, err := DriverList()
	if err != nil {
		t.Fatalf("DriverList() failed: %v", err)
	}
	if len(drivers) == 0 {
		t.Fatal("no drivers to test DriverName()")
	}

	name, err := DriverName(drivers[0])
	if err != nil {
		t.Fatalf("DriverName() failed: %v", err)
	}
	if name == "" {
		t.Fatal("DriverName() returned empty string")
	}
}

func TestDriverPath(t *testing.T) {
	drivers, err := DriverList()
	if err != nil {
		t.Fatalf("DriverList() failed: %v", err)
	}
	if len(drivers) == 0 {
		t.Fatal("no drivers to test DriverPath()")
	}

	path, err := DriverPath(drivers[0])
	if err != nil {
		t.Fatalf("DriverPath() failed: %v", err)
	}
	if path == "" {
		t.Fatal("DriverPath() returned empty string")
	}
	// The path should be a DOS path (drive letter followed by colon)
	if len(path) < 2 || path[1] != ':' {
		t.Fatalf("DriverPath() expected DOS path, got: %s", path)
	}
}

func TestDriverNameKnownDriver(t *testing.T) {
	drivers, err := DriverList()
	if err != nil {
		t.Fatalf("DriverList() failed: %v", err)
	}
	if len(drivers) == 0 {
		t.Fatal("no drivers to test")
	}

	// Find ntoskrnl.exe in the driver list
	found := false
	for _, d := range drivers {
		name, err := DriverName(d)
		if err != nil {
			continue
		}
		if strings.EqualFold(name, "ntoskrnl.exe") {
			found = true
			path, err := DriverPath(d)
			if err != nil {
				t.Fatalf("DriverPath(ntoskrnl) failed: %v", err)
			}
			if !strings.HasSuffix(path, ".exe") {
				t.Fatalf("DriverPath(ntoskrnl) expected .exe suffix, got: %s", path)
			}
			break
		}
	}
	if !found {
		t.Log("ntoskrnl.exe not found in driver list (this is unusual)")
	}
}
