//go:build windows

package os

import (
	"testing"
)

func TestNativePathToDosPath(t *testing.T) {
	path, err := NativePathToDosPath(`\SystemRoot\System32\ntdll.dll`)
	if err != nil {
		t.Fatalf("NativePathToDosPath() failed: %v", err)
	}
	if len(path) < 3 || path[1] != ':' {
		t.Fatalf("expected DOS path with drive letter, got: %s", path)
	}
}

func TestNativePathToDosPathInvalid(t *testing.T) {
	_, err := NativePathToDosPath(`\Some\Bogus\Path\that\does\not\exist`)
	if err == nil {
		t.Fatal("expected error for non-existent native path, got nil")
	}
}
