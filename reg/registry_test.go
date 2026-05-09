//go:build windows

package reg

import (
	"testing"
)

func TestCheckPath_ValidPath(t *testing.T) {
	got := CheckPath(`HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion`)
	if !got {
		t.Errorf("CheckPath() = %v, want true", got)
	}
}

func TestCheckPath_InvalidPath(t *testing.T) {
	got := CheckPath(`HKLM\NONEXISTENT_KEY_12345`)
	if got {
		t.Errorf("CheckPath() = %v, want false", got)
	}
}

func TestCheckPath_MalformedRootKey(t *testing.T) {
	got := CheckPath(`INVALID_ROOT\Some\Path`)
	if got {
		t.Errorf("CheckPath() = %v, want false", got)
	}
}

func TestCheckPath_EmptyPath(t *testing.T) {
	got := CheckPath(``)
	if got {
		t.Errorf("CheckPath() = %v, want false", got)
	}
}

func TestGetValue_KnownKey(t *testing.T) {
	got, err := GetValue(`HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion`, "SystemRoot")
	if err != nil {
		t.Fatalf("GetValue() error = %v", err)
	}
	if got == "" {
		t.Errorf("GetValue() = %q, want non-empty string", got)
	}
}

func TestGetValue_NonExistentKey(t *testing.T) {
	_, err := GetValue(`HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion`, "NonExistentKey_XYZ")
	if err == nil {
		t.Errorf("GetValue() expected error for non-existent key")
	}
}

func TestGetValue_InvalidPath(t *testing.T) {
	_, err := GetValue(``, "Test")
	if err == nil {
		t.Errorf("GetValue() expected error for invalid path")
	}
}

func TestEnumValues_KnownKey(t *testing.T) {
	values, err := EnumValues(`HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion`)
	if err != nil {
		t.Fatalf("EnumValues() error = %v", err)
	}
	if len(values) == 0 {
		t.Fatal("EnumValues() returned empty slice")
	}
	// Should contain at least one known value
	found := false
	for _, v := range values {
		if v.Name == "SystemRoot" || v.Name == "CurrentBuild" {
			found = true
			if v.Type == "" {
				t.Errorf("EnumValues() value %q has empty Type", v.Name)
			}
			break
		}
	}
	if !found {
		t.Error("EnumValues() did not find expected value (SystemRoot or CurrentBuild)")
	}
}

func TestEnumValues_NonExistentKey(t *testing.T) {
	_, err := EnumValues(`HKLM\NONEXISTENT_KEY_XYZ123`)
	if err == nil {
		t.Errorf("EnumValues() expected error for non-existent key")
	}
}

func TestEnumValues_InvalidPath(t *testing.T) {
	_, err := EnumValues(``)
	if err == nil {
		t.Errorf("EnumValues() expected error for invalid path")
	}
}

func TestEnumSubKeys_KnownKey(t *testing.T) {
	keys, err := EnumSubKeys(`HKLM\SOFTWARE\Microsoft\Windows NT\CurrentVersion`)
	if err != nil {
		t.Fatalf("EnumSubKeys() error = %v", err)
	}
	if len(keys) == 0 {
		t.Fatal("EnumSubKeys() returned empty slice")
	}
}

func TestEnumSubKeys_NonExistentKey(t *testing.T) {
	_, err := EnumSubKeys(`HKLM\NONEXISTENT_KEY_XYZ123`)
	if err == nil {
		t.Errorf("EnumSubKeys() expected error for non-existent key")
	}
}

func TestEnumSubKeys_InvalidPath(t *testing.T) {
	_, err := EnumSubKeys(``)
	if err == nil {
		t.Errorf("EnumSubKeys() expected error for invalid path")
	}
}
