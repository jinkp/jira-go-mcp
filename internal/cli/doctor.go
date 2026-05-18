package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/jinkp/jira-go-mcp/internal/claude"
	"github.com/jinkp/jira-go-mcp/internal/claudedesktop"
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
	return &cobra.Command{
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
