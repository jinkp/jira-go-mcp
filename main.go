package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/isaiahrafael/jira-go-mcp/internal/cli"
)

func main() {
	root := &cobra.Command{
		Use:   "jira-mcp",
		Short: "Jira release management via MCP protocol",
		Long:  "jira-mcp exposes Jira release management as MCP tools for AI assistants.",
		// SilenceUsage prevents cobra from printing usage on error — keeps MCP stdio clean
		SilenceUsage: true,
	}

	// Register subcommands — config is loaded lazily inside each RunE that needs it
	root.AddCommand(cli.NewVersionCmd())
	root.AddCommand(cli.NewMCPCmd())
	root.AddCommand(cli.NewSetupCmd())
	root.AddCommand(cli.NewTUICmd())
	root.AddCommand(cli.NewDoctorCmd())

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
