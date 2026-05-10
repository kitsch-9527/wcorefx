//go:build windows

package reg

import (
	"testing"

	"golang.org/x/sys/windows/registry"
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

func TestSetString_RoundTrip(t *testing.T) {
	// Use a test key under HKCU\Software to avoid system impact
	testPath := `HKCU\Software\wcorefx_test`
	defer registry.DeleteKey(registry.CURRENT_USER, `Software\wcorefx_test`)

	// Create the key first before setting values
	if err := CreateKey(testPath); err != nil {
		t.Fatalf("CreateKey() error = %v", err)
	}

	err := SetString(testPath, "TestValue", "hello")
	if err != nil {
		t.Fatalf("SetString() error = %v", err)
	}

	got, err := GetValue(testPath, "TestValue")
	if err != nil {
		t.Fatalf("GetValue() error = %v", err)
	}
	if got != "hello" {
		t.Errorf("GetValue() = %q, want %q", got, "hello")
	}
}

func TestSetDWORD_RoundTrip(t *testing.T) {
	testPath := `HKCU\Software\wcorefx_test`
	defer registry.DeleteKey(registry.CURRENT_USER, `Software\wcorefx_test`)

	// Create the key first before setting values
	if err := CreateKey(testPath); err != nil {
		t.Fatalf("CreateKey() error = %v", err)
	}

	err := SetDWORD(testPath, "TestDWORD", 42)
	if err != nil {
		t.Fatalf("SetDWORD() error = %v", err)
	}
	// Read back using raw registry API to verify
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\wcorefx_test`, registry.QUERY_VALUE)
	if err != nil {
		t.Fatalf("OpenKey error = %v", err)
	}
	defer k.Close()
	val, _, err := k.GetIntegerValue("TestDWORD")
	if err != nil {
		t.Fatalf("GetIntegerValue error = %v", err)
	}
	if val != 42 {
		t.Errorf("value = %d, want 42", val)
	}
}

func TestCreateKey_ThenCheckPath(t *testing.T) {
	testPath := `HKCU\Software\wcorefx_test_create`
	defer registry.DeleteKey(registry.CURRENT_USER, `Software\wcorefx_test_create`)

	err := CreateKey(testPath)
	if err != nil {
		t.Fatalf("CreateKey() error = %v", err)
	}
	if !CheckPath(testPath) {
		t.Error("CheckPath() = false after CreateKey")
	}
}

func TestDeleteValue_RoundTrip(t *testing.T) {
	testPath := `HKCU\Software\wcorefx_test`
	defer registry.DeleteKey(registry.CURRENT_USER, `Software\wcorefx_test`)

	// Create the key first before setting values
	if err := CreateKey(testPath); err != nil {
		t.Fatalf("CreateKey() error = %v", err)
	}

	if err := SetString(testPath, "ToDelete", "value"); err != nil {
		t.Fatalf("SetString() error = %v", err)
	}
	err := DeleteValue(testPath, "ToDelete")
	if err != nil {
		t.Fatalf("DeleteValue() error = %v", err)
	}
	// The key still exists, but the value should be gone
	vals, err := EnumValues(testPath)
	if err != nil {
		t.Fatalf("EnumValues() error = %v", err)
	}
	for _, v := range vals {
		if v.Name == "ToDelete" {
			t.Error("DeleteValue() did not remove the value")
		}
	}
}
