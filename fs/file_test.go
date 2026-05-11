//go:build windows

package fs

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"unsafe"

	"golang.org/x/sys/windows"
)

const versionFile = "C:\\Windows\\System32\\kernel32.dll"

func TestCreateTime(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable() failed: %v", err)
	}

	ts, err := CreateTime(exe)
	if err != nil {
		t.Fatalf("CreateTime(%q) failed: %v", exe, err)
	}
	if ts <= 0 {
		t.Errorf("CreateTime(%q) = %d, want > 0", exe, ts)
	}
}

func TestAccessTime(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable() failed: %v", err)
	}

	ts, err := AccessTime(exe)
	if err != nil {
		t.Fatalf("AccessTime(%q) failed: %v", exe, err)
	}
	if ts <= 0 {
		t.Errorf("AccessTime(%q) = %d, want > 0", exe, ts)
	}
}

func TestModifyTime(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable() failed: %v", err)
	}

	ts, err := ModifyTime(exe)
	if err != nil {
		t.Fatalf("ModifyTime(%q) failed: %v", exe, err)
	}
	if ts <= 0 {
		t.Errorf("ModifyTime(%q) = %d, want > 0", exe, ts)
	}
}

func TestVersionInfo(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{"kernel32", versionFile},
		{"ntdll", "C:\\Windows\\System32\\ntdll.dll"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			version, err := VersionInfo(tt.path)
			if err != nil {
				t.Fatalf("VersionInfo(%q) failed: %v", tt.path, err)
			}
			if version == "" {
				t.Errorf("VersionInfo(%q) = empty string, want non-empty", tt.path)
			}
		})
	}
}

func TestInfo(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		infoType InfoType
		want     string // empty means just check non-empty
	}{
		{
			name:     "kernel32/FileDescription",
			path:     versionFile,
			infoType: FileDescription,
		},
		{
			name:     "kernel32/CompanyName",
			path:     versionFile,
			infoType: CompanyName,
		},
		{
			name:     "kernel32/ProductVersion",
			path:     versionFile,
			infoType: ProductVersion,
		},
		{
			name:     "kernel32/ProductName",
			path:     versionFile,
			infoType: ProductName,
		},
		{
			name:     "kernel32/LegalCopyright",
			path:     versionFile,
			infoType: LegalCopyright,
		},
		{
			name:     "kernel32/OriginalFileName",
			path:     versionFile,
			infoType: OriginalFileName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Info(tt.path, tt.infoType, nil)
			if err != nil {
				t.Fatalf("Info(%q, %q, nil) failed: %v", tt.path, tt.infoType, err)
			}
			if got == "" {
				t.Errorf("Info(%q, %q, nil) = empty string, want non-empty", tt.path, tt.infoType)
			}
			t.Logf("%s = %q", tt.name, got)
		})
	}
}

func TestInfoCustomTranslation(t *testing.T) {
	// Discover available translations from the file instead of hardcoding.
	versionInfo, err := getResourceVersionInfo(versionFile)
	if err != nil {
		t.Fatalf("getResourceVersionInfo failed: %v", err)
	}

	type langAndCodePage struct {
		Language uint16
		CodePage uint16
	}
	var translation *langAndCodePage
	var bufferSize uint32
	err = windows.VerQueryValue(
		unsafe.Pointer(&versionInfo[0]),
		"\\VarFileInfo\\Translation",
		unsafe.Pointer(&translation),
		&bufferSize,
	)
	if err != nil || translation == nil || bufferSize == 0 {
		t.Fatalf("no translations found in version info")
	}

	translationStr := fmt.Sprintf("%04X%04X", translation.Language, translation.CodePage)
	got, err := Info(versionFile, FileDescription, &translationStr)
	if err != nil {
		t.Fatalf("Info(%q, FileDescription, %q) failed: %v", versionFile, translationStr, err)
	}
	if got == "" {
		t.Errorf("Info with custom translation returned empty string")
	}
	t.Logf("Translation=%q FileDescription=%q", translationStr, got)
}

func TestInfoInvalidPath(t *testing.T) {
	_, err := Info("Z:\\nonexistent.dll", FileDescription, nil)
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

func TestVersionInfoInvalidPath(t *testing.T) {
	_, err := VersionInfo("Z:\\nonexistent.dll")
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

func TestCreateTimeInvalidPath(t *testing.T) {
	_, err := CreateTime("Z:\\nonexistent.dll")
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

func TestFileTimesValidRange(t *testing.T) {
	// Verify timestamps for a well-known system file.
	path := versionFile

	create, err := CreateTime(path)
	if err != nil {
		t.Fatalf("CreateTime(%q) failed: %v", path, err)
	}
	mod, err := ModifyTime(path)
	if err != nil {
		t.Fatalf("ModifyTime(%q) failed: %v", path, err)
	}
	access, err := AccessTime(path)
	if err != nil {
		t.Fatalf("AccessTime(%q) failed: %v", path, err)
	}

	if create <= 0 {
		t.Errorf("CreateTime(%q) = %d, want > 0", path, create)
	}
	if mod <= 0 {
		t.Errorf("ModifyTime(%q) = %d, want > 0", path, mod)
	}
	if access <= 0 {
		t.Errorf("AccessTime(%q) = %d, want > 0", path, access)
	}
}

func BenchmarkCreateTime(b *testing.B) {
	exe, err := os.Executable()
	if err != nil {
		b.Fatalf("os.Executable() failed: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = CreateTime(exe)
	}
}

func BenchmarkVersionInfo(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = VersionInfo(versionFile)
	}
}

func TestListDir(t *testing.T) {
	entries, err := ListDir("C:\\Windows")
	if err != nil {
		t.Fatalf("ListDir() error = %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("ListDir() returned empty slice")
	}
	for _, e := range entries {
		if e.Name == "" {
			t.Error("ListDir() entry has empty Name")
		}
		if e.Path == "" {
			t.Error("ListDir() entry has empty Path")
		}
	}
}

func TestListDir_InvalidPath(t *testing.T) {
	_, err := ListDir("Z:\\NONEXISTENT_DIR_XYZ123")
	if err == nil {
		t.Errorf("ListDir() expected error for invalid path")
	}
}

func TestVolumes(t *testing.T) {
	vols, err := Volumes()
	if err != nil {
		t.Fatalf("Volumes() failed: %v", err)
	}
	if len(vols) == 0 {
		t.Fatal("Volumes() returned empty slice")
	}
	t.Logf("Found %d volumes", len(vols))
	for _, v := range vols {
		if v.Name != "" {
			t.Logf("  %s (%s) type=%s total=%d free=%d",
				v.Name, v.Label, v.Type, v.TotalSize, v.FreeSize)
			break
		}
	}
}

// TestSha1_KnownFile 测试已知文件的 SHA1 计算，验证返回 40 字符十六进制字符串。
func TestSha1_KnownFile(t *testing.T) {
	hash, err := Sha1(versionFile)
	if err != nil {
		t.Fatalf("Sha1(%q) failed: %v", versionFile, err)
	}
	if len(hash) != 40 {
		t.Errorf("Sha1(%q) len = %d, want 40", versionFile, len(hash))
	}
	t.Logf("Sha1(%q) = %s", versionFile, hash)
}

// TestSha1_KnownValue 测试已知内容的 SHA1 值。
func TestSha1_KnownValue(t *testing.T) {
	content := []byte("hello")
	expected := "aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d"

	dir := t.TempDir()
	path := filepath.Join(dir, "hello.txt")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	hash, err := Sha1(path)
	if err != nil {
		t.Fatalf("Sha1(%q) failed: %v", path, err)
	}
	if hash != expected {
		t.Errorf("Sha1(%q) = %s, want %s", path, hash, expected)
	}
}

// TestSha1_EmptyFile 测试空文件的 SHA1 值。
func TestSha1_EmptyFile(t *testing.T) {
	expected := "da39a3ee5e6b4b0d3255bfef95601890afd80709"

	dir := t.TempDir()
	path := filepath.Join(dir, "empty.txt")
	if err := os.WriteFile(path, []byte{}, 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	hash, err := Sha1(path)
	if err != nil {
		t.Fatalf("Sha1(%q) failed: %v", path, err)
	}
	if hash != expected {
		t.Errorf("Sha1(%q) = %s, want %s", path, hash, expected)
	}
}

// TestSha1_InvalidPath 测试无效路径返回错误。
func TestSha1_InvalidPath(t *testing.T) {
	_, err := Sha1("Z:\\nonexistent_file_012345")
	if err == nil {
		t.Error("expected error for nonexistent file, got nil")
	}
}

func TestExample(t *testing.T) {
	// Example: retrieving timestamps and version info for a system DLL.
	path := versionFile

	t.Run("timestamps", func(t *testing.T) {
		t.Logf("File: %s", filepath.Base(path))
		if ct, err := CreateTime(path); err == nil {
			t.Logf("  CreateTime : %d", ct)
		}
		if at, err := AccessTime(path); err == nil {
			t.Logf("  AccessTime : %d", at)
		}
		if mt, err := ModifyTime(path); err == nil {
			t.Logf("  ModifyTime : %d", mt)
		}
	})

	t.Run("version", func(t *testing.T) {
		if ver, err := VersionInfo(path); err == nil {
			t.Logf("VersionInfo: %s", ver)
		}
		if desc, err := Info(path, FileDescription, nil); err == nil {
			t.Logf("FileDescription: %s", desc)
		}
	})
}
