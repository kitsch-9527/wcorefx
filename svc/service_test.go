//go:build windows

package svc

import (
	"testing"
)

func TestList(t *testing.T) {
	services, err := List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}
	if len(services) == 0 {
		t.Fatal("List() returned empty slice")
	}
	found := false
	for _, s := range services {
		if s.Name == "" {
			t.Error("List() returned service with empty Name")
		}
		if s.Status == "" {
			t.Error("List() returned service with empty Status")
		}
		if s.Name == "winmgmt" || s.Name == "WmiApSrv" {
			found = true
			t.Logf("Found service: %s (%s) PID=%d", s.Name, s.Status, s.PID)
		}
	}
	if !found {
		t.Log("Note: WMI service not found (may not be installed)")
	}
}

func TestStatus_ExistingService(t *testing.T) {
	si, err := Status("winmgmt")
	if err != nil {
		t.Fatalf("Status() failed: %v", err)
	}
	t.Logf("winmgmt status: %s, PID=%d", si.Status, si.PID)
}

func TestStatus_NonExistentService(t *testing.T) {
	_, err := Status("NonexistentService_XYZ123")
	if err == nil {
		t.Error("Status() expected error for non-existent service")
	}
}

func TestConfig_ExistingService(t *testing.T) {
	si, err := Config("winmgmt")
	if err != nil {
		t.Fatalf("Config() failed: %v", err)
	}
	t.Logf("winmgmt config: start=%s account=%s", si.StartType, si.Account)
}
