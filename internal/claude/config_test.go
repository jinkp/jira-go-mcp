package claude_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/jinkp/jira-go-mcp/internal/claude"
)

// TestSave_CreatesFileWithEntry verifies that SaveToPath creates the config file
// with the jira-mcp mcpServers entry when the file does not exist.
func TestSave_CreatesFileWithEntry(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".claude.json")

	err := claude.SaveToPath(path, "/usr/local/bin/jira-mcp")
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

	entryRaw, ok := mcpMap["jira-mcp"]
	if !ok {
		t.Fatal("expected 'jira-mcp' key in mcpServers, not found")
	}

	var entry map[string]json.RawMessage
	if err := json.Unmarshal(entryRaw, &entry); err != nil {
		t.Fatalf("unmarshal entry = %v", err)
	}

	var cmd string
	if err := json.Unmarshal(entry["command"], &cmd); err != nil {
		t.Fatalf("unmarshal command = %v", err)
	}
	if cmd != "/usr/local/bin/jira-mcp" {
		t.Errorf("entry.command = %q, want %q", cmd, "/usr/local/bin/jira-mcp")
	}
}

// TestSave_PreservesExistingKeys verifies that SaveToPath preserves all existing
// keys in the config file when adding the jira-mcp entry.
func TestSave_PreservesExistingKeys(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".claude.json")

	initial := `{"mcpServers":{"existing-tool":{"command":"/bin/existing","args":["mcp"]}},"configVersion":1}`
	if err := os.WriteFile(path, []byte(initial), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	err := claude.SaveToPath(path, "/usr/local/bin/jira-mcp")
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

	if _, ok := result["configVersion"]; !ok {
		t.Error("expected 'configVersion' key to be preserved, not found")
	}

	var mcpMap map[string]json.RawMessage
	if err := json.Unmarshal(result["mcpServers"], &mcpMap); err != nil {
		t.Fatalf("unmarshal mcpServers = %v", err)
	}
	if _, ok := mcpMap["existing-tool"]; !ok {
		t.Error("expected 'existing-tool' to be preserved in mcpServers, not found")
	}
	if _, ok := mcpMap["jira-mcp"]; !ok {
		t.Error("expected 'jira-mcp' to be added to mcpServers, not found")
	}
}

// TestCheck_ReturnsTrueWhenEntryPresent verifies Check() returns true when
// the jira-mcp key is present in mcpServers.
func TestCheck_ReturnsTrueWhenEntryPresent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".claude.json")

	content := `{"mcpServers":{"jira-mcp":{"command":"/bin/jira-mcp","args":["mcp"]}}}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if !claude.Check(path) {
		t.Error("Check() = false, want true when jira-mcp entry is present")
	}
}

// TestCheck_ReturnsFalseWhenEntryAbsent verifies Check() returns false when
// the jira-mcp key is not in mcpServers.
func TestCheck_ReturnsFalseWhenEntryAbsent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".claude.json")

	content := `{"mcpServers":{"other-tool":{"command":"/bin/other","args":["mcp"]}}}`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if claude.Check(path) {
		t.Error("Check() = true, want false when jira-mcp entry is absent")
	}
}

// TestSave_GlobalScope verifies Save() with GlobalScope writes to the correct path.
func TestSave_GlobalScope(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)

	err := claude.Save(claude.GlobalScope, "/usr/local/bin/jira-mcp")
	if err != nil {
		t.Fatalf("Save(GlobalScope) error = %v", err)
	}

	cfgPath := filepath.Join(dir, ".claude.json")
	if !claude.Check(cfgPath) {
		t.Error("Check() = false after Save(GlobalScope), expected true")
	}
}

// TestSave_LocalScope verifies Save() with LocalScope writes to .claude/settings.json.
func TestSave_LocalScope(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}
	t.Cleanup(func() { os.Chdir(origDir) })

	err := claude.Save(claude.LocalScope, "/usr/local/bin/jira-mcp")
	if err != nil {
		t.Fatalf("Save(LocalScope) error = %v", err)
	}

	cfgPath := filepath.Join(dir, ".claude", "settings.json")
	if !claude.Check(cfgPath) {
		t.Error("Check() = false after Save(LocalScope), expected true")
	}
}

// TestCheck_ReturnsFalseWhenFileAbsent verifies Check() returns false when
// the config file does not exist.
func TestCheck_ReturnsFalseWhenFileAbsent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent.json")
	if claude.Check(path) {
		t.Error("Check() = true on absent file, want false")
	}
}

// TestGlobalPath verifies GlobalPath returns a path that ends with .claude.json.
func TestGlobalPath(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)

	path := claude.GlobalPath()
	if path == "" {
		t.Error("GlobalPath() returned empty string")
	}
	if filepath.Base(path) != ".claude.json" {
		t.Errorf("GlobalPath() base = %q, want %q", filepath.Base(path), ".claude.json")
	}
}

// TestLocalPath verifies LocalPath returns a path under .claude/settings.json.
func TestLocalPath(t *testing.T) {
	path := claude.LocalPath()
	if path == "" {
		t.Error("LocalPath() returned empty string")
	}
	if filepath.Base(path) != "settings.json" {
		t.Errorf("LocalPath() base = %q, want %q", filepath.Base(path), "settings.json")
	}
}
