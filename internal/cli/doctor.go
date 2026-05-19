package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jinkp/jira-go-mcp/internal/claude"
	"github.com/jinkp/jira-go-mcp/internal/claudedesktop"
	"github.com/jinkp/jira-go-mcp/internal/config"
	"github.com/jinkp/jira-go-mcp/internal/jira"
	"github.com/jinkp/jira-go-mcp/internal/opencode"
)

// installTarget represents one potential install location to check.
type installTarget struct {
	label string
	scope string
	check func() (bool, error)
}

// NewDoctorCmd returns the Cobra command that checks whether jira-mcp is
// registered in each known AI client configuration.
func NewDoctorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Check jira-mcp installation status across all AI clients",
		Long:  "Reads each known AI client config and reports ✓ if jira-mcp is registered, ✗ if not.",
		RunE: func(cmd *cobra.Command, args []string) error {
			w := cmd.OutOrStdout()

			fmt.Fprintf(w, "%-30s %-10s %s\n", "Target", "Scope", "Status")
			fmt.Fprintf(w, "%-30s %-10s %s\n", "──────────────────────────────", "──────────", "──────")

			targets := buildTargets()
			for _, t := range targets {
				ok, err := t.check()
				status := "✓ registered"
				if err != nil {
					status = fmt.Sprintf("✗ error: %v", err)
				} else if !ok {
					status = "✗ not registered"
				}
				fmt.Fprintf(w, "%-30s %-10s %s\n", t.label, t.scope, status)
			}

			return nil
		},
	}
	cmd.AddCommand(newDoctorJiraCmd())
	return cmd
}

func newDoctorJiraCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "jira",
		Short: "Validate Jira configuration and connectivity",
		RunE: func(cmd *cobra.Command, args []string) error {
			w := cmd.OutOrStdout()
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("jira config invalid: %w", err)
			}

			fmt.Fprintln(w, "Jira configuration")
			fmt.Fprintln(w, "──────────────────")
			fmt.Fprintf(w, "Base URL   : %s\n", cfg.BaseURL)
			fmt.Fprintf(w, "Email      : %s\n", cfg.Email)
			fmt.Fprintf(w, "Token      : %s\n", maskedToken(cfg.APIToken))

			info, err := jira.ValidateConnection(context.Background(), cfg.BaseURL, cfg.Email, cfg.APIToken)
			if err != nil {
				fmt.Fprintf(w, "Connection : ✗ %v\n", err)
				return nil
			}

			name := info.DisplayName
			if name == "" {
				name = info.Email
			}
			fmt.Fprintf(w, "Connection : ✓ authenticated as %s\n", name)
			return nil
		},
	}
}

func maskedToken(token string) string {
	if token == "" {
		return "(missing)"
	}
	if len(token) <= 4 {
		return "****"
	}
	return token[:2] + "****" + token[len(token)-2:]
}

// buildTargets returns the ordered list of install targets to check.
func buildTargets() []installTarget {
	targets := []installTarget{
		{
			label: "OpenCode",
			scope: "global",
			check: func() (bool, error) {
				return opencode.Check(opencode.GlobalPath()), nil
			},
		},
		{
			label: "OpenCode",
			scope: "local",
			check: func() (bool, error) {
				return opencode.Check(opencode.LocalPath()), nil
			},
		},
		{
			label: "Claude Code",
			scope: "global",
			check: func() (bool, error) {
				return claude.Check(claude.GlobalPath()), nil
			},
		},
		{
			label: "Claude Code",
			scope: "local",
			check: func() (bool, error) {
				return claude.Check(claude.LocalPath()), nil
			},
		},
	}

	// Claude Desktop — OS-aware, may be unsupported
	targets = append(targets, installTarget{
		label: "Claude Desktop",
		scope: "global",
		check: func() (bool, error) {
			path, err := claudedesktop.Path()
			if err != nil {
				return false, err
			}
			return claudedesktop.Check(path)
		},
	})

	return targets
}
