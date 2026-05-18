package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/jinkp/jira-go-mcp/internal/claude"
	"github.com/jinkp/jira-go-mcp/internal/claudedesktop"
	"github.com/jinkp/jira-go-mcp/internal/config"
	"github.com/jinkp/jira-go-mcp/internal/opencode"
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
// It has subcommands for Jira runtime config plus MCP client registration.
// None of these require Jira environment variables.
func NewSetupCmd() *cobra.Command {
	setupCmd := &cobra.Command{
		Use:   "setup",
		Short: "Configure Jira credentials and register jira-mcp in AI clients",
		Long:  "Configure jira-mcp runtime settings for Jira, then register the binary as an MCP server in OpenCode, Claude Code, or Claude Desktop.",
	}

	setupCmd.AddCommand(newSetupJiraCmd())
	setupCmd.AddCommand(newSetupOpencodeCmd())
	setupCmd.AddCommand(newSetupClaudeCmd())
	setupCmd.AddCommand(newSetupClaudeDesktopCmd())

	return setupCmd
}

func newSetupJiraCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "jira",
		Short: "Configure Jira connection settings in an external file",
		RunE: func(cmd *cobra.Command, args []string) error {
			existing := existingJiraConfig()
			reader := bufio.NewReader(cmd.InOrStdin())

			baseURL, err := promptValue(cmd.OutOrStdout(), reader, "Jira base URL", existing.BaseURL, false)
			if err != nil {
				return err
			}
			email, err := promptValue(cmd.OutOrStdout(), reader, "Jira email", existing.Email, false)
			if err != nil {
				return err
			}
			token, err := promptValue(cmd.OutOrStdout(), reader, "Jira API token", existing.APIToken, false)
			if err != nil {
				return err
			}
			project, err := promptValue(cmd.OutOrStdout(), reader, "Default project key (optional)", existing.DefaultProject, true)
			if err != nil {
				return err
			}
			doneStatuses, err := promptValue(cmd.OutOrStdout(), reader, "Done statuses (comma-separated)", strings.Join(existing.DoneStatuses, ","), true)
			if err != nil {
				return err
			}
			criticalLabels, err := promptValue(cmd.OutOrStdout(), reader, "Critical labels (comma-separated)", strings.Join(existing.CriticalLabels, ","), true)
			if err != nil {
				return err
			}

			next := &config.Config{
				BaseURL:        strings.TrimSpace(baseURL),
				Email:          strings.TrimSpace(email),
				APIToken:       strings.TrimSpace(token),
				DefaultProject: strings.TrimSpace(project),
				DoneStatuses:   splitCSVInput(doneStatuses, "Done"),
				CriticalLabels: splitCSVInput(criticalLabels, "critical"),
			}

			if err := config.SaveFile(next); err != nil {
				return fmt.Errorf("failed to save Jira config: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "✓ Jira config saved to %s\n", config.FilePath())
			return nil
		},
	}
}

func existingJiraConfig() *config.Config {
	cfg, err := config.Load()
	if err == nil && cfg != nil {
		return cfg
	}
	return &config.Config{
		DoneStatuses:   []string{"Done"},
		CriticalLabels: []string{"critical"},
	}
}

func promptValue(w interface{ Write([]byte) (int, error) }, reader *bufio.Reader, label, current string, optional bool) (string, error) {
	prompt := label
	if current != "" {
		prompt += fmt.Sprintf(" [%s]", current)
	}
	prompt += ": "
	if _, err := fmt.Fprint(w, prompt); err != nil {
		return "", err
	}
	line, err := reader.ReadString('\n')
	if err != nil && err.Error() != "EOF" {
		return "", err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		if current != "" {
			return current, nil
		}
		if optional {
			return "", nil
		}
		return "", fmt.Errorf("%s is required", label)
	}
	return line, nil
}

func splitCSVInput(val, fallback string) []string {
	val = strings.TrimSpace(val)
	if val == "" {
		return []string{fallback}
	}
	parts := strings.Split(val, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	if len(out) == 0 {
		return []string{fallback}
	}
	return out
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
	cmd.Flags().BoolVar(&local, "local", false, "Write to local OpenCode config (./.opencode/opencode.json)")

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
