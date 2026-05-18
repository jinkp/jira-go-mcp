package cli

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	internaltui "github.com/jinkp/jira-go-mcp/internal/tui"
)

// NewTUICmd returns the cobra command that launches the interactive setup wizard.
// Does not require Jira environment variables.
func NewTUICmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tui",
		Short: "Launch the interactive setup wizard",
		Long:  "Launch the Bubbletea interactive wizard to register jira-mcp in an AI client.",
		RunE: func(cmd *cobra.Command, args []string) error {
			p := tea.NewProgram(internaltui.NewSetupWizard())
			finalModel, err := p.Run()
			if err != nil {
				return fmt.Errorf("wizard error: %w", err)
			}

			// Check if the user cancelled
			if wizard, ok := finalModel.(internaltui.SetupWizard); ok {
				if wizard.Cancelled() {
					fmt.Fprintln(os.Stderr, "Setup cancelled.")
					return nil
				}
			}
			return nil
		},
	}
}
