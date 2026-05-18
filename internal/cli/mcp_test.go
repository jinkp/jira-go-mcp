package cli_test

import (
	"testing"

	"github.com/jinkp/jira-go-mcp/internal/cli"
)

// TestNewMCPCmd_NoArgs verifies that the MCP command factory is callable
// without pre-wired service dependencies (lazy config refactor).
// The command must not call config.Load() at construction time.
func TestNewMCPCmd_NoArgs(t *testing.T) {
	// RED: NewMCPCmd() should accept no arguments after the lazy refactor.
	// Currently it requires a release.ReleaseService — this test drives the API change.
	cmd := cli.NewMCPCmd()
	if cmd == nil {
		t.Fatal("NewMCPCmd() returned nil")
	}
	if cmd.Use != "mcp" {
		t.Errorf("Use = %q, want %q", cmd.Use, "mcp")
	}
}

// TestNewMCPCmd_RunE_MissingConfig verifies that the mcp RunE returns an error
// when Jira env vars are missing — config is loaded lazily inside RunE.
func TestNewMCPCmd_RunE_MissingConfig(t *testing.T) {
	// Clear all Jira env vars
	jiraVars := []string{"JIRA_BASE_URL", "JIRA_EMAIL", "JIRA_API_TOKEN",
		"JIRA_DEFAULT_PROJECT", "JIRA_DONE_STATUSES", "JIRA_CRITICAL_LABELS"}
	for _, k := range jiraVars {
		t.Setenv(k, "")
	}

	cmd := cli.NewMCPCmd()
	// Execute RunE directly — it should fail because config.Load() returns error
	err := cmd.RunE(cmd, []string{})
	if err == nil {
		t.Fatal("expected error when Jira env vars are missing, got nil")
	}
}
