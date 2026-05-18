package claudedesktop_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/isaiahrafael/jira-go-mcp/internal/claudedesktop"
)

// TestPath_ReturnsPathOnSupportedOS verifies that Path() returns a non-empty
// path on Windows and macOS, and ErrUnsupported on Linux.
func TestPath_ReturnsPathOnSupportedOS(t *testing.T) {
	path, err := claudedesktop.Path()

	switch runtime.GOOS {
	case "windows", "darwin":
		if err != nil {
			t.Errorf("Path() error = %v, want nil on %s", err, runtime.GOOS)
		}
		if path == "" {
			t.Errorf("Path() = empty, want non-empty path on %s", runtime.GOOS)
		}
		if filepath.Base(path) != "claude_desktop_config.json" {
			t.Errorf("Path() base = %q, want %q", filepath.Base(path), "claude_desktop_config.json")
		}
	case "linux":
		if err == nil {
			t.Error("Path() error = nil, want ErrUnsupported on linux")
		}
		if err != claudedesktop.ErrUnsupported {
			t.Errorf("Path() error = %v, want ErrUnsupported", err)
		}
	}
}

// TestSave_CreatesFileWithEntry verifies that SaveToPath creates the config file
// with the jira-mcp mcpServers entry when the file does not exist.
func TestSave_CreatesFileWithEntry(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "claude_desktop_config.json")

	err := claudedesktop.SaveToPath(path, "/usr/local/bin/jira-mcp")
	if err != nil {
		t.Fatalf("SaveToPath() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	var result map[string]json.RawMessage
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	mcpRaw, ok := result["mcpServers"]
	if !ok {
		t.Fatal("expected 'mcpServers' key in result, not found")
	}

	var mcpMap map[string]json.RawMessage
	if err := json.Unmarshal(mcpRaw, &mcpMap); err != nil {
		t.Fatalf("unmarshal mcpServers = %v", err)
	}

	if _, ok := mcpMap["jira-mcp"]; !ok {
		t.Fatal("expected 'jira-mcp' key in mcpServers, not found")
	}
}

// TestSave_PreservesExistingKeys verifies that SaveToPath preserves all
// existing keys when merging the new entry.
func TestSave_PreservesExistingKeys(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "claude_desktop_config.json")

	initial := `{"mcpServers":{"other":{"command":"/bin/other","args":["mcp"]}}}`
	if err := os.WriteFile(path, []byte(initial), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	err := claudedesktop.SaveToPath(path, "/usr/local/bin/jira-mcp")
	if err != nil {
		t.Fatalf("SaveToPath() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	var result map[string]json.RawMessage
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal() = %v", err)
	}

	var mcpMap map[string]json.RawMessage
	if err := json.Unmarshal(result["mcpServers"], &mcpMap); err != nil {
		t.Fatalf("unmarshal mcpServers = %v", err)
	}

	if _, ok := mcpMap["other"]; !ok {
		t.Error("expected 'other' entry to be preserved, not found")
	}
	if _, ok := mcpMap["jira-mcp"]; !ok {
		t.Error("expected 'jira-mcp' entry to be added, not found")
	}
}

// TestCheck_ReturnsTrueWhenEntryPresent verifies Check() returns true when
// the jira-mcp key exists in mcpServers.
func TestCheck_ReturnsTrueWhenEntryPresent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "claude_desktop_config.json")

	content := `{"mcpServers":{"jira-mcp":{"command":"/bin/jira-mcp","args":["mcp"]}}}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	ok, err := claudedesktop.Check(path)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !ok {
		t.Error("Check() = false, want true when jira-mcp entry is present")
	}
}

// TestSave_UsesOSPath verifies Save() writes to the OS-resolved path
// when it is a supported platform (Windows/macOS).
func TestSave_UsesOSPath(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)
	t.Setenv("APPDATA", filepath.Join(dir, "AppData", "Roaming"))

	err := claudedesktop.Save("/usr/local/bin/jira-mcp")
	// On Linux: ErrUnsupported — expected
	// On Windows/macOS: should succeed
	if err != nil && err != claudedesktop.ErrUnsupported {
		t.Fatalf("Save() unexpected error = %v", err)
	}
}

// TestCheck_ReturnsFalseWhenFileAbsent verifies Check() returns false (not error)
// when the config file does not exist.
func TestCheck_ReturnsFalseWhenFileAbsent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.json")
	ok, err := claudedesktop.Check(path)
	if err != nil {
		t.Fatalf("Check() error = %v on absent file, want nil", err)
	}
	if ok {
		t.Error("Check() = true on absent file, want false")
	}
}

// TestCheck_ReturnsFalseWhenEntryAbsent verifies Check() returns false when
// the jira-mcp key is absent from mcpServers.
func TestCheck_ReturnsFalseWhenEntryAbsent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "claude_desktop_config.json")

	content := `{"mcpServers":{"other":{"command":"/bin/other","args":["mcp"]}}}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	ok, err := claudedesktop.Check(path)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if ok {
		t.Error("Check() = true, want false when jira-mcp entry is absent")
	}
}

// TestCheck_ReturnsFalseWhenNoMcpServersKey verifies Check() returns false
// when the file has no mcpServers key.
func TestCheck_ReturnsFalseWhenNoMcpServersKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "claude_desktop_config.json")

	content := `{"globalShortcut":""}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	ok, err := claudedesktop.Check(path)
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if ok {
		t.Error("Check() = true, want false when mcpServers key is absent")
	}
}

// TestSave_RoundTrip verifies Save() → Check() round-trip on a temp path.
func TestSave_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "claude_desktop_config.json")

	err := claudedesktop.SaveToPath(path, "/bin/jira-mcp")
	if err != nil {
		t.Fatalf("SaveToPath() error = %v", err)
	}

	ok, err := claudedesktop.Check(path)
	if err != nil {
		t.Fatalf("Check() after SaveToPath() error = %v", err)
	}
	if !ok {
		t.Error("Check() = false after SaveToPath(), expected true")
	}
}
