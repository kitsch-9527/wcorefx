//go:build windows

package task

import (
	"testing"
)

const sampleTaskXML = `<?xml version="1.0" encoding="UTF-16"?>
<Task xmlns="http://schemas.microsoft.com/windows/2004/02/mit/task">
  <RegistrationInfo>
    <Author>Microsoft Corporation</Author>
  </RegistrationInfo>
  <Settings>
    <Enabled>true</Enabled>
  </Settings>
  <Actions Context="Author">
    <Exec>
      <Command>C:\Windows\system32\cmd.exe</Command>
      <Arguments>/c echo test</Arguments>
    </Exec>
  </Actions>
  <Principals>
    <Principal id="Author">
      <UserId>SYSTEM</UserId>
    </Principal>
  </Principals>
</Task>`

const sampleTaskXMLDisabled = `<?xml version="1.0" encoding="UTF-16"?>
<Task xmlns="http://schemas.microsoft.com/windows/2004/02/mit/task">
  <Settings>
    <Enabled>false</Enabled>
  </Settings>
  <Actions>
    <Exec>
      <Command>notepad.exe</Command>
    </Exec>
  </Actions>
</Task>`

const sampleTaskXMLComHandler = `<?xml version="1.0" encoding="UTF-16"?>
<Task xmlns="http://schemas.microsoft.com/windows/2004/02/mit/task">
  <Settings>
    <Enabled>true</Enabled>
  </Settings>
  <Actions>
    <ComHandler>
      <Clsid>{00000000-0000-0000-0000-000000000000}</Clsid>
    </ComHandler>
  </Actions>
</Task>`

const sampleTaskXMLUTF8 = `<?xml version="1.0" encoding="UTF-8"?>
<Task>
  <RegistrationInfo>
    <Author>Test User</Author>
  </RegistrationInfo>
  <Settings>
    <Enabled>true</Enabled>
  </Settings>
  <Actions>
    <Exec>
      <Command>C:\test.exe</Command>
      <Arguments>-flag</Arguments>
    </Exec>
  </Actions>
  <Principals>
    <Principal>
      <UserId>LOCAL SERVICE</UserId>
    </Principal>
  </Principals>
</Task>`

func TestParseXML_Exec(t *testing.T) {
	task, err := ParseXML([]byte(sampleTaskXML))
	if err != nil {
		t.Fatalf("ParseXML() error = %v", err)
	}
	if task.Command != `C:\Windows\system32\cmd.exe` {
		t.Errorf("Command = %q, want %q", task.Command, `C:\Windows\system32\cmd.exe`)
	}
	if task.Arguments != "/c echo test" {
		t.Errorf("Arguments = %q, want %q", task.Arguments, "/c echo test")
	}
	if !task.Enabled {
		t.Error("Enabled = false, want true")
	}
	if task.Author != "Microsoft Corporation" {
		t.Errorf("Author = %q, want %q", task.Author, "Microsoft Corporation")
	}
	if task.UserId != "SYSTEM" {
		t.Errorf("UserId = %q, want %q", task.UserId, "SYSTEM")
	}
	if task.ComHandler {
		t.Error("ComHandler = true, want false")
	}
}

func TestParseXML_Disabled(t *testing.T) {
	task, err := ParseXML([]byte(sampleTaskXMLDisabled))
	if err != nil {
		t.Fatalf("ParseXML() error = %v", err)
	}
	if task.Command != "notepad.exe" {
		t.Errorf("Command = %q, want %q", task.Command, "notepad.exe")
	}
	if task.Enabled {
		t.Error("Enabled = true, want false")
	}
}

func TestParseXML_ComHandler(t *testing.T) {
	task, err := ParseXML([]byte(sampleTaskXMLComHandler))
	if err != nil {
		t.Fatalf("ParseXML() error = %v", err)
	}
	if !task.ComHandler {
		t.Error("ComHandler = false, want true")
	}
	if task.Clsid != "{00000000-0000-0000-0000-000000000000}" {
		t.Errorf("Clsid = %q", task.Clsid)
	}
	if task.Command != "" {
		t.Errorf("Command = %q, want empty for ComHandler", task.Command)
	}
}

func TestParseXML_UTF8(t *testing.T) {
	task, err := ParseXML([]byte(sampleTaskXMLUTF8))
	if err != nil {
		t.Fatalf("ParseXML() error = %v", err)
	}
	if task.Command != `C:\test.exe` {
		t.Errorf("Command = %q", task.Command)
	}
	if task.Arguments != "-flag" {
		t.Errorf("Arguments = %q", task.Arguments)
	}
	if !task.Enabled {
		t.Error("Enabled = false, want true")
	}
	if task.Author != "Test User" {
		t.Errorf("Author = %q", task.Author)
	}
	if task.UserId != "LOCAL SERVICE" {
		t.Errorf("UserId = %q", task.UserId)
	}
}

func TestParseXML_InvalidXML(t *testing.T) {
	_, err := ParseXML([]byte("not xml"))
	if err == nil {
		t.Error("ParseXML() expected error for invalid XML")
	}
}

func TestParseXML_Empty(t *testing.T) {
	_, err := ParseXML([]byte{})
	if err == nil {
		t.Error("ParseXML() expected error for empty data")
	}
}

func TestTaskDir(t *testing.T) {
	dir, err := taskDir()
	if err != nil {
		t.Fatalf("taskDir() error = %v", err)
	}
	if dir == "" {
		t.Fatal("taskDir() returned empty")
	}
	t.Logf("Task directory: %s", dir)
}

func TestList(t *testing.T) {
	tasks, err := List()
	if err != nil {
		// Tasks directory might be inaccessible in some environments
		t.Skipf("List() failed (may require elevation): %v", err)
	}
	if len(tasks) == 0 {
		t.Log("List() returned empty (no tasks found)")
		return
	}
	t.Logf("Found %d tasks", len(tasks))
	for _, task := range tasks {
		t.Logf("  %s: enabled=%v cmd=%s", task.Name, task.Enabled, task.Command)
	}
}

func TestDecodeUTF16(t *testing.T) {
	// Create a UTF-16LE encoded sample
	utf16Bytes := []byte{
		0xFF, 0xFE, // BOM
		'T', 0, 'e', 0, 's', 0, 't', 0, // "Test"
	}
	result := decodeUTF16(utf16Bytes)
	if string(result) != "Test" {
		t.Errorf("decodeUTF16 = %q, want %q", string(result), "Test")
	}
}

func TestDecodeUTF16_NoBOM(t *testing.T) {
	data := []byte("plain UTF-8")
	result := decodeUTF16(data)
	if string(result) != "plain UTF-8" {
		t.Errorf("decodeUTF16 = %q, want %q", string(result), "plain UTF-8")
	}
}

func TestNormalizeName(t *testing.T) {
	root := `C:\Windows\System32\Tasks`
	path := `C:\Windows\System32\Tasks\Microsoft\Windows\Example`
	name := normalizeName(root, path)
	if name != `\Microsoft\Windows\Example` {
		t.Errorf("normalizeName = %q, want %q", name, `\Microsoft\Windows\Example`)
	}
}

func TestListFrom_InvalidDir(t *testing.T) {
	_, err := ListFrom(`Z:\NONEXISTENT_DIR_XYZ123`)
	if err == nil {
		t.Errorf("ListFrom() expected error for invalid directory")
	}
}

// Benchmark XML parsing
func BenchmarkParseXML(b *testing.B) {
	data := []byte(sampleTaskXML)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseXML(data)
	}
}
