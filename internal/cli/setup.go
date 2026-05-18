package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/isaiahrafael/jira-go-mcp/internal/claude"
	"github.com/isaiahrafael/jira-go-mcp/internal/claudedesktop"
	"github.com/isaiahrafael/jira-go-mcp/internal/opencode"
)

// resolveBinPath returns the absolute path of the current executable,
// resolving any symlinks (e.g. Homebrew shims).
func resolveBinPath() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("could not determine executable path: %w", err)
	}
	resolved, err := filepath.EvalSymlinks(exe)
	if err != nil {
		// Fallback to unresolved path if EvalSymlinks fails (e.g. tests)
		return exe, nil
	}
	return resolved, nil
}

// NewSetupCmd returns the Cobra command for the setup subcommand.
// It has three sub-subcommands: opencode, claude, claude-desktop.
// None of these require Jira environment variables.
func NewSetupCmd() *cobra.Command {
	setupCmd := &cobra.Command{
		Use:   "setup",
		Short: "Register jira-mcp in an AI client configuration",
		Long:  "Register the jira-mcp binary as an MCP server in OpenCode, Claude Code, or Claude Desktop.",
	}

	setupCmd.AddCommand(newSetupOpencodeCmd())
	setupCmd.AddCommand(newSetupClaudeCmd())
	setupCmd.AddCommand(newSetupClaudeDesktopCmd())

	return setupCmd
}

func newSetupOpencodeCmd() *cobra.Command {
	var global bool
	var local bool

	cmd := &cobra.Command{
		Use:   "opencode",
		Short: "Register jira-mcp in OpenCode",
		RunE: func(cmd *cobra.Command, args []string) error {
			binPath, err := resolveBinPath()
			if err != nil {
				return err
			}

			scope := opencode.GlobalScope
			if local {
				scope = opencode.LocalScope
			}

			if err := opencode.Save(scope, binPath); err != nil {
				return fmt.Errorf("failed to register in OpenCode: %w", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "✓ OpenCode registered")
			return nil
		},
	}

	cmd.Flags().BoolVar(&global, "global", true, "Write to global OpenCode config (~/.config/opencode/opencode.json)")
	cmd.Flags().BoolVar(&local, "local", false, "Write to local OpenCode config (./opencode.json)")

	return cmd
}

func newSetupClaudeCmd() *cobra.Command {
	var global bool
	var local bool

	cmd := &cobra.Command{
		Use:   "claude",
		Short: "Register jira-mcp in Claude Code",
		RunE: func(cmd *cobra.Command, args []string) error {
			binPath, err := resolveBinPath()
			if err != nil {
				return err
			}

			scope := claude.GlobalScope
			if local {
				scope = claude.LocalScope
			}

			if err := claude.Save(scope, binPath); err != nil {
				return fmt.Errorf("failed to register in Claude Code: %w", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "✓ Claude Code registered")
			return nil
		},
	}

	cmd.Flags().BoolVar(&global, "global", true, "Write to global Claude Code config (~/.claude.json)")
	cmd.Flags().BoolVar(&local, "local", false, "Write to local Claude Code config (./.claude/settings.json)")

	return cmd
}

func newSetupClaudeDesktopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "claude-desktop",
		Short: "Register jira-mcp in Claude Desktop",
		RunE: func(cmd *cobra.Command, args []string) error {
			binPath, err := resolveBinPath()
			if err != nil {
				return err
			}

			if err := claudedesktop.Save(binPath); err != nil {
				return fmt.Errorf("failed to register in Claude Desktop: %w", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "✓ Claude Desktop registered")
			return nil
		},
	}
}
