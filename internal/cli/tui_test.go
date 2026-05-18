package cli_test

import (
	"testing"

	"github.com/isaiahrafael/jira-go-mcp/internal/cli"
)

// TestNewTUICmd_ExistsAndHasCorrectUse verifies the TUI command is well-formed.
func TestNewTUICmd_ExistsAndHasCorrectUse(t *testing.T) {
	cmd := cli.NewTUICmd()
	if cmd == nil {
		t.Fatal("NewTUICmd() returned nil")
	}
	if cmd.Use != "tui" {
		t.Errorf("Use = %q, want %q", cmd.Use, "tui")
	}
	if cmd.Short == "" {
		t.Error("Short description must not be empty")
	}
}
