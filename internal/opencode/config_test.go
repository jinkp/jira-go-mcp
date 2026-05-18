package opencode_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/jinkp/jira-go-mcp/internal/opencode"
)

// TestSave_CreatesFileWithEntry verifies that Save creates the config file
// with the jira-mcp MCP entry when the file does not exist.
func TestSave_CreatesFileWithEntry(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "opencode.json")

	err := opencode.SaveToPath(path, "/usr/local/bin/jira-mcp")
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

	mcpRaw, ok := result["mcp"]
	if !ok {
		t.Fatal("expected 'mcp' key in result, not found")
	}

	var mcpMap map[string]json.RawMessage
	if err := json.Unmarshal(mcpRaw, &mcpMap); err != nil {
		t.Fatalf("unmarshal mcp = %v", err)
	}

	entryRaw, ok := mcpMap["jira-mcp"]
	if !ok {
		t.Fatal("expected 'jira-mcp' key in mcp, not found")
	}

	var entry map[string]json.RawMessage
	if err := json.Unmarshal(entryRaw, &entry); err != nil {
		t.Fatalf("unmarshal entry = %v", err)
	}

	var entryType string
	if err := json.Unmarshal(entry["type"], &entryType); err != nil {
		t.Fatalf("unmarshal type = %v", err)
	}
	if entryType != "local" {
		t.Errorf("entry.type = %q, want %q", entryType, "local")
	}

	var cmd string
	if err := json.Unmarshal(entry["command"], &cmd); err != nil {
		t.Fatalf("unmarshal command = %v", err)
	}
	if cmd != "/usr/local/bin/jira-mcp" {
		t.Errorf("entry.command = %q, want %q", cmd, "/usr/local/bin/jira-mcp")
	}
}

// TestSave_PreservesExistingKeys verifies that Save preserves keys that were
// already in the file when adding the jira-mcp entry.
func TestSave_PreservesExistingKeys(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "opencode.json")

	// Write a file with an existing entry
	initial := `{"mcp":{"other-tool":{"type":"local","command":"/bin/other","args":["mcp"]}},"theme":"dark"}`
	if err := os.WriteFile(path, []byte(initial), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	err := opencode.SaveToPath(path, "/usr/local/bin/jira-mcp")
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

	// 'theme' key must be preserved
	if _, ok := result["theme"]; !ok {
		t.Error("expected 'theme' key to be preserved, not found")
	}

	// 'other-tool' must still be in mcp
	var mcpMap map[string]json.RawMessage
	if err := json.Unmarshal(result["mcp"], &mcpMap); err != nil {
		t.Fatalf("unmarshal mcp = %v", err)
	}
	if _, ok := mcpMap["other-tool"]; !ok {
		t.Error("expected 'other-tool' key to be preserved in mcp, not found")
	}
	if _, ok := mcpMap["jira-mcp"]; !ok {
		t.Error("expected 'jira-mcp' key to be added to mcp, not found")
	}
}

// TestCheck_ReturnsTrueWhenEntryPresent verifies Check() returns true when
// the jira-mcp key exists in the mcp section of the config file.
func TestCheck_ReturnsTrueWhenEntryPresent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "opencode.json")

	content := `{"mcp":{"jira-mcp":{"type":"local","command":"/bin/jira-mcp","args":["mcp"]}}}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if !opencode.Check(path) {
		t.Error("Check() = false, want true when jira-mcp entry is present")
	}
}

// TestCheck_ReturnsFalseWhenEntryAbsent verifies Check() returns false when
// the jira-mcp key is not in the config file.
func TestCheck_ReturnsFalseWhenEntryAbsent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "opencode.json")

	content := `{"mcp":{"other-tool":{"type":"local","command":"/bin/other","args":["mcp"]}}}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if opencode.Check(path) {
		t.Error("Check() = true, want false when jira-mcp entry is absent")
	}
}

// TestCheck_ReturnsFalseWhenFileAbsent verifies Check() returns false when
// the config file does not exist at all.
func TestCheck_ReturnsFalseWhenFileAbsent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.json")
	if opencode.Check(path) {
		t.Error("Check() = true, want false when file does not exist")
	}
}

// TestLocalPath verifies LocalPath returns a path under the current directory.
func TestLocalPath(t *testing.T) {
	path := opencode.LocalPath()
	if path == "" {
		t.Error("LocalPath() returned empty string")
	}
	if filepath.Base(path) != "opencode.json" {
		t.Errorf("LocalPath() base = %q, want %q", filepath.Base(path), "opencode.json")
	}
}

// TestSave_GlobalScope uses Save() with GlobalScope and verifies the file is created.
func TestSave_GlobalScope(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)

	err := opencode.Save(opencode.GlobalScope, "/usr/local/bin/jira-mcp")
	if err != nil {
		t.Fatalf("Save(GlobalScope) error = %v", err)
	}

	cfgPath := filepath.Join(dir, ".config", "opencode", "opencode.json")
	if !opencode.Check(cfgPath) {
		t.Error("Check() = false after Save(GlobalScope), expected true")
	}
}

// TestSave_LocalScope uses Save() with LocalScope and verifies the file is created.
func TestSave_LocalScope(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}
	t.Cleanup(func() { os.Chdir(origDir) })

	err := opencode.Save(opencode.LocalScope, "/usr/local/bin/jira-mcp")
	if err != nil {
		t.Fatalf("Save(LocalScope) error = %v", err)
	}

	cfgPath := filepath.Join(dir, "opencode.json")
	if !opencode.Check(cfgPath) {
		t.Error("Check() = false after Save(LocalScope), expected true")
	}
}

// TestGlobalPath verifies GlobalPath returns a path under the user config dir.
func TestGlobalPath(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir) // Windows fallback

	path := opencode.GlobalPath()
	if path == "" {
		t.Error("GlobalPath() returned empty string")
	}
	// Must end with opencode.json
	if filepath.Base(path) != "opencode.json" {
		t.Errorf("GlobalPath() base = %q, want %q", filepath.Base(path), "opencode.json")
	}
}
