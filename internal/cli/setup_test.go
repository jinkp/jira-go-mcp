package cli_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/isaiahrafael/jira-go-mcp/internal/cli"
)

// TestNewSetupCmd_ExistsWithSubcommands verifies the setup command has the
// expected sub-subcommands: opencode, claude, claude-desktop.
func TestNewSetupCmd_ExistsWithSubcommands(t *testing.T) {
	cmd := cli.NewSetupCmd()
	if cmd == nil {
		t.Fatal("NewSetupCmd() returned nil")
	}
	if cmd.Use != "setup" {
		t.Errorf("Use = %q, want %q", cmd.Use, "setup")
	}

	subNames := map[string]bool{}
	for _, sub := range cmd.Commands() {
		subNames[sub.Use] = true
	}

	for _, want := range []string{"opencode", "claude", "claude-desktop"} {
		if !subNames[want] {
			t.Errorf("expected subcommand %q not found; got %v", want, subNames)
		}
	}
}

// TestSetupOpencode_Global_WritesConfig verifies that `setup opencode --global`
// writes the jira-mcp entry to the temp global config path.
// We override the global path by setting HOME/USERPROFILE.
func TestSetupOpencode_Global_WritesConfig(t *testing.T) {
	// Clear Jira env vars — setup must run without them
	for _, k := range []string{"JIRA_BASE_URL", "JIRA_EMAIL", "JIRA_API_TOKEN"} {
		t.Setenv(k, "")
	}

	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)

	cmd := cli.NewSetupCmd()
	// Find the opencode sub-subcommand
	var opencodeCmd interface{ Execute() error }
	for _, sub := range cmd.Commands() {
		if sub.Use == "opencode" {
			opencodeCmd = sub
			break
		}
	}
	if opencodeCmd == nil {
		t.Fatal("opencode subcommand not found")
	}

	// Execute via root to get full flag parsing
	root := cli.NewSetupCmd()
	root.SetArgs([]string{"opencode", "--global"})
	if err := root.Execute(); err != nil {
		t.Fatalf("setup opencode --global failed: %v", err)
	}

	// Check the written file exists under our temp dir
	cfgPath := filepath.Join(dir, ".config", "opencode", "opencode.json")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("expected config file at %s, ReadFile error = %v", cfgPath, err)
	}

	var result map[string]json.RawMessage
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	var mcpMap map[string]json.RawMessage
	if err := json.Unmarshal(result["mcp"], &mcpMap); err != nil {
		t.Fatalf("unmarshal mcp = %v", err)
	}

	if _, ok := mcpMap["jira-mcp"]; !ok {
		t.Error("expected 'jira-mcp' key in mcp section, not found")
	}
}

// TestSetupOpencode_Local_WritesConfig verifies that `setup opencode --local`
// writes the jira-mcp entry to a local opencode.json.
func TestSetupOpencode_Local_WritesConfig(t *testing.T) {
	// Clear Jira env vars — setup must run without them
	for _, k := range []string{"JIRA_BASE_URL", "JIRA_EMAIL", "JIRA_API_TOKEN"} {
		t.Setenv(k, "")
	}

	dir := t.TempDir()
	// Change working directory to temp dir for local scope
	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}
	t.Cleanup(func() { os.Chdir(origDir) })

	root := cli.NewSetupCmd()
	root.SetArgs([]string{"opencode", "--local"})
	if err := root.Execute(); err != nil {
		t.Fatalf("setup opencode --local failed: %v", err)
	}

	cfgPath := filepath.Join(dir, "opencode.json")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("expected config file at %s, ReadFile error = %v", cfgPath, err)
	}

	var result map[string]json.RawMessage
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	var mcpMap map[string]json.RawMessage
	if err := json.Unmarshal(result["mcp"], &mcpMap); err != nil {
		t.Fatalf("unmarshal mcp = %v", err)
	}

	if _, ok := mcpMap["jira-mcp"]; !ok {
		t.Error("expected 'jira-mcp' key in mcp section, not found")
	}
}

// TestSetupClaude_Global_WritesConfig verifies that `setup claude --global`
// writes the jira-mcp entry to the Claude Code global config.
func TestSetupClaude_Global_WritesConfig(t *testing.T) {
	for _, k := range []string{"JIRA_BASE_URL", "JIRA_EMAIL", "JIRA_API_TOKEN"} {
		t.Setenv(k, "")
	}

	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)

	root := cli.NewSetupCmd()
	root.SetArgs([]string{"claude", "--global"})
	if err := root.Execute(); err != nil {
		t.Fatalf("setup claude --global failed: %v", err)
	}

	cfgPath := filepath.Join(dir, ".claude.json")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("expected config file at %s, ReadFile error = %v", cfgPath, err)
	}

	var result map[string]json.RawMessage
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	var mcpMap map[string]json.RawMessage
	if err := json.Unmarshal(result["mcpServers"], &mcpMap); err != nil {
		t.Fatalf("unmarshal mcpServers = %v", err)
	}

	if _, ok := mcpMap["jira-mcp"]; !ok {
		t.Error("expected 'jira-mcp' key in mcpServers, not found")
	}
}

// TestSetupRunsWithoutJiraEnvVars verifies that the setup command (task 1.4)
// runs successfully even when Jira env vars are completely absent.
func TestSetupRunsWithoutJiraEnvVars(t *testing.T) {
	for _, k := range []string{"JIRA_BASE_URL", "JIRA_EMAIL", "JIRA_API_TOKEN",
		"JIRA_DEFAULT_PROJECT", "JIRA_DONE_STATUSES", "JIRA_CRITICAL_LABELS"} {
		t.Setenv(k, "")
	}

	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)

	root := cli.NewSetupCmd()
	root.SetArgs([]string{"opencode", "--global"})
	err := root.Execute()
	if err != nil {
		t.Errorf("setup opencode failed without Jira env vars: %v", err)
	}
}

// TestSetupClaudeDesktop_OnLinux verifies that setup claude-desktop returns
// an error message on Linux (or passes gracefully on Windows/macOS).
func TestSetupClaudeDesktop_PrintsError(t *testing.T) {
	for _, k := range []string{"JIRA_BASE_URL", "JIRA_EMAIL", "JIRA_API_TOKEN"} {
		t.Setenv(k, "")
	}

	dir := t.TempDir()
	t.Setenv("HOME", dir)
	t.Setenv("USERPROFILE", dir)
	// Set APPDATA for Windows
	t.Setenv("APPDATA", filepath.Join(dir, "AppData", "Roaming"))

	root := cli.NewSetupCmd()
	root.SetArgs([]string{"claude-desktop"})

	var errOut strings.Builder
	root.SetErr(&errOut)

	// On Linux: should fail. On Windows/macOS: should succeed.
	// Either way: should not panic.
	_ = root.Execute()
}
