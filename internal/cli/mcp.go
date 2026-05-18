package cli

import (
	"fmt"
	"os"

	"github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"

	"github.com/isaiahrafael/jira-go-mcp/internal/config"
	jiraclient "github.com/isaiahrafael/jira-go-mcp/internal/jira"
	internalmcp "github.com/isaiahrafael/jira-go-mcp/internal/mcp"
	"github.com/isaiahrafael/jira-go-mcp/internal/release"
)

// NewMCPCmd returns the cobra command for the mcp subcommand.
// Config is loaded lazily inside RunE so that setup/tui/doctor commands
// can run without Jira environment variables being set.
// CRITICAL: This command MUST NOT write anything to stdout before server.ServeStdio().
// All diagnostic output goes to stderr.
func NewMCPCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "mcp",
		Short: "Start the MCP stdio server",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Lazy config load — only triggered when `mcp` subcommand executes
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("configuration error: %w", err)
			}

			// Wire dependencies locally
			jiraClient := jiraclient.NewHTTPClient(cfg.BaseURL, cfg.Email, cfg.APIToken)
			svc := release.NewService(jiraClient, cfg)

			s := server.NewMCPServer("jira-mcp", Version,
				server.WithToolCapabilities(true),
			)
			internalmcp.RegisterTools(s, svc)

			// All logs MUST go to stderr to avoid polluting the MCP stdio protocol
			fmt.Fprintln(os.Stderr, "jira-mcp MCP server starting")

			if err := server.ServeStdio(s); err != nil {
				return fmt.Errorf("MCP server error: %w", err)
			}
			return nil
		},
	}
}
