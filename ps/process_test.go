//go:build windows

package ps

import (
	"os"
	"strings"
	"testing"
)

func TestList(t *testing.T) {
	procs, err := List()
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}
	if len(procs) == 0 {
		t.Fatal("List() returned empty slice")
	}
}

func TestFind(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	exeName := exe
	if idx := strings.LastIndex(exe, "\\"); idx >= 0 {
		exeName = exe[idx+1:]
	}

	matches, err := Find(exeName)
	if err != nil {
		t.Fatalf("Find(%q) failed: %v", exeName, err)
	}
	if len(matches) == 0 {
		t.Fatalf("Find(%q) returned no matches", exeName)
	}
}

func TestFind_NotFound(t *testing.T) {
	matches, err := Find("thisprocessshouldnotexist_xyz_1234567890.exe")
	if err != nil {
		t.Fatalf("Find() failed: %v", err)
	}
	if len(matches) != 0 {
		t.Logf("Found unexpected matches: %d", len(matches))
	}
}

func TestCommandLine(t *testing.T) {
	cmdline, err := CommandLine(uint32(os.Getpid()))
	if err != nil {
		t.Fatalf("CommandLine() failed: %v", err)
	}
	if cmdline == "" {
		t.Log("CommandLine returned empty (may be expected in some test runners)")
	}
}

func TestMemoryInfo(t *testing.T) {
	mem, err := MemoryInfo(uint32(os.Getpid()))
	if err != nil {
		t.Fatalf("MemoryInfo() failed: %v", err)
	}
	if mem.WorkingSetSize == 0 {
		t.Log("MemoryInfo: WorkingSetSize is 0")
	}
}

func TestTimes(t *testing.T) {
	times, err := Times(uint32(os.Getpid()))
	if err != nil {
		t.Fatalf("Times() failed: %v", err)
	}
	if times.CreationTime.Nanoseconds() == 0 {
		t.Fatal("Times: CreationTime is zero")
	}
}

func TestPath(t *testing.T) {
	path, err := Path(uint32(os.Getpid()))
	if err != nil {
		t.Fatalf("Path() failed: %v", err)
	}
	if path == "" {
		t.Fatal("Path() returned empty")
	}
}

func TestUser(t *testing.T) {
	domain, name, err := User(uint32(os.Getpid()))
	if err != nil {
		t.Fatalf("User() failed: %v", err)
	}
	if name == "" {
		t.Fatal("User() returned empty name")
	}
	t.Logf("Current user: %s\\%s", domain, name)
}

func TestIsTerminated(t *testing.T) {
	if IsTerminated(uint32(os.Getpid())) {
		t.Fatal("IsTerminated should be false for current process")
	}
}

func TestParentID(t *testing.T) {
	ppid := ParentID(uint32(os.Getpid()))
	if ppid == 0 {
		t.Log("ParentID returned 0 (may be expected)")
	}
}

func TestSessionID(t *testing.T) {
	sid, err := SessionID(uint32(os.Getpid()))
	if err != nil {
		t.Fatalf("SessionID() failed: %v", err)
	}
	if sid == 0 {
		t.Log("SessionID is 0 (may be expected for system services)")
	}
}

func TestModules(t *testing.T) {
	mods, err := Modules(uint32(os.Getpid()))
	if err != nil {
		t.Fatalf("Modules() failed: %v", err)
	}
	if len(mods) == 0 {
		t.Fatal("Modules() returned empty slice")
	}
}

func TestFindByPath(t *testing.T) {
	selfPath, err := Path(uint32(os.Getpid()))
	if err != nil {
		t.Fatalf("Path() failed: %v", err)
	}
	if selfPath == "" {
		t.Skip("Path() returned empty")
	}
	pids, err := FindByPath(selfPath)
	if err != nil {
		t.Fatalf("FindByPath() error = %v", err)
	}
	if len(pids) == 0 {
		t.Fatalf("FindByPath(%q) returned no matches for current process", selfPath)
	}
}

func TestFindByPath_NotFound(t *testing.T) {
	pids, err := FindByPath("C:\\NONEXISTENT_PATH_XYZ123.exe")
	if err != nil {
		t.Fatalf("FindByPath() error = %v", err)
	}
	if len(pids) != 0 {
		t.Errorf("FindByPath() returned %d matches for nonexistent path", len(pids))
	}
}

func TestOpenToken(t *testing.T) {
	token, err := openToken(uint32(os.Getpid()))
	if err != nil {
		t.Fatalf("openToken() failed: %v", err)
	}
	token.Close()
}

func TestThreads(t *testing.T) {
	threads, err := Threads(uint32(os.Getpid()))
	if err != nil {
		t.Fatalf("Threads() failed: %v", err)
	}
	if len(threads) == 0 {
		t.Fatal("Threads() returned empty slice for current process")
	}
	t.Logf("Current process has %d threads", len(threads))
	for _, th := range threads {
		if th.ID == 0 {
			t.Error("Threads() returned thread with ID=0")
		}
	}
}

func TestWindows(t *testing.T) {
	wins, err := Windows()
	if err != nil {
		t.Fatalf("Windows() failed: %v", err)
	}
	t.Logf("Found %d top-level windows", len(wins))
	for _, w := range wins {
		if w.Title != "" {
			t.Logf("  Window: %q class=%q", w.Title, w.ClassName)
			break
		}
	}
}
