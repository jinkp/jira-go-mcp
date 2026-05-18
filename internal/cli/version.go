package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version is the current binary version. Set via -ldflags at build time.
var Version = "dev"

// NewVersionCmd returns a cobra command that prints the binary version.
func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version of jira-mcp",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(Version)
		},
	}
}
