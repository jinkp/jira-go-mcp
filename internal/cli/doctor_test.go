package cli_test

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jinkp/jira-go-mcp/internal/cli"
)

// TestNewDoctorCmd_Exists verifies the doctor command is well-formed.
func TestNewDoctorCmd_Exists(t *testing.T) {
	cmd := cli.NewDoctorCmd()
	if cmd == nil {
		t.Fatal("NewDoctorCmd() returned nil")
	}
	if cmd.Use != "doctor" {
		t.Errorf("Use = %q, want %q", cmd.Use, "doctor")
	}
}

// TestDoctor_NoInstallations_ShowsAllFail verifies that when no config files
// contain the jira-mcp entry, all targets show ✗ and the command exits 0.
func TestDoctor_NoInstallations_ShowsAllFail(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)
	t.Setenv("APPDATA", filepath.Join(dir, "AppData", "Roaming"))

	// Change CWD to temp dir so local paths also point there
	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}
	t.Cleanup(func() { os.Chdir(origDir) })

	var out bytes.Buffer
	cmd := cli.NewDoctorCmd()
	cmd.SetOut(&out)

	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("doctor RunE() error = %v", err)
	}

	output := out.String()
	if output == "" {
		t.Fatal("doctor produced no output")
	}

	// All 5 targets should appear with a ✗ symbol
	failCount := strings.Count(output, "✗")
	if failCount < 4 {
		// On Linux there may be only 4 (claude-desktop unsupported)
		// On Windows/macOS there should be 5
		t.Errorf("expected at least 4 ✗ symbols (no installations), got %d\nOutput:\n%s", failCount, output)
	}
}

// TestDoctor_WithOpenCodeGlobal_ShowsPass verifies that when the OpenCode
// global config has the jira-mcp entry, that target shows ✓.
func TestDoctor_WithOpenCodeGlobal_ShowsPass(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)
	t.Setenv("APPDATA", filepath.Join(dir, "AppData", "Roaming"))

	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}
	t.Cleanup(func() { os.Chdir(origDir) })

	// Write a valid OpenCode global config with jira-mcp entry
	cfgDir := filepath.Join(dir, ".config", "opencode")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() error = %v", err)
	}

	entry := map[string]interface{}{
		"mcp": map[string]interface{}{
			"jira-mcp": map[string]interface{}{
				"type":    "local",
				"command": "/usr/local/bin/jira-mcp",
				"args":    []string{"mcp"},
			},
		},
	}
	data, _ := json.MarshalIndent(entry, "", "  ")
	cfgPath := filepath.Join(cfgDir, "opencode.json")
	if err := os.WriteFile(cfgPath, data, 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	var out bytes.Buffer
	cmd := cli.NewDoctorCmd()
	cmd.SetOut(&out)

	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("doctor RunE() error = %v", err)
	}

	output := out.String()
	passCount := strings.Count(output, "✓")
	if passCount < 1 {
		t.Errorf("expected at least 1 ✓ symbol (OpenCode global installed), got %d\nOutput:\n%s", passCount, output)
	}
}
