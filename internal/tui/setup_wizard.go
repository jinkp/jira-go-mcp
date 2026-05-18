// Package tui provides the Bubbletea-based interactive setup wizard for jira-mcp.
package tui

import (
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/isaiahrafael/jira-go-mcp/internal/claude"
	"github.com/isaiahrafael/jira-go-mcp/internal/claudedesktop"
	"github.com/isaiahrafael/jira-go-mcp/internal/opencode"
)

// screenID represents the current screen in the wizard flow.
type screenID int

const (
	scopeScreen   screenID = iota // Select target + scope
	confirmScreen                 // Show resolved path, ask to proceed
	doneScreen                    // Success
	errorScreen                   // Write failure
)

// targetOption represents a selectable install target.
type targetOption struct {
	label  string
	target string // "opencode" | "claude" | "claude-desktop"
	scope  string // "global" | "local"
}

var targetOptions = []targetOption{
	{label: "OpenCode (global)", target: "opencode", scope: "global"},
	{label: "OpenCode (local)", target: "opencode", scope: "local"},
	{label: "Claude Code (global)", target: "claude", scope: "global"},
	{label: "Claude Code (local)", target: "claude", scope: "local"},
	{label: "Claude Desktop (global)", target: "claude-desktop", scope: "global"},
}

// wizardModel is the Bubbletea model for the setup wizard.
type wizardModel struct {
	screen    screenID
	cursor    int    // current selection cursor on scopeScreen
	target    string // selected target
	scope     string // selected scope
	binPath   string // resolved binary path
	resolved  string // absolute config file path shown on confirmScreen
	err       error
	cancelled bool
}

// SetupWizard is the exported interface for the wizard model,
// allowing tests to inspect the final state.
type SetupWizard interface {
	tea.Model
	Cancelled() bool
}

// NewSetupWizard creates and returns a new wizard model ready to run.
func NewSetupWizard() SetupWizard {
	exe, _ := os.Executable()
	binPath, err := filepath.EvalSymlinks(exe)
	if err != nil {
		binPath = exe
	}
	return &wizardModel{
		screen:  scopeScreen,
		binPath: binPath,
	}
}

// Cancelled returns true if the user cancelled the wizard (Ctrl+C or 'n').
func (m *wizardModel) Cancelled() bool {
	return m.cancelled
}

// Init implements tea.Model.
func (m *wizardModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (m *wizardModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			m.cancelled = true
			return m, tea.Quit

		case tea.KeyUp:
			if m.screen == scopeScreen && m.cursor > 0 {
				m.cursor--
			}

		case tea.KeyDown:
			if m.screen == scopeScreen && m.cursor < len(targetOptions)-1 {
				m.cursor++
			}

		case tea.KeyEnter:
			return m.handleEnter()

		default:
			if m.screen == confirmScreen {
				switch msg.String() {
				case "y", "Y":
					return m.handleConfirm()
				case "n", "N":
					m.cancelled = true
					return m, tea.Quit
				}
			}
		}
	}
	return m, nil
}

// handleEnter advances the wizard to the next screen.
func (m *wizardModel) handleEnter() (tea.Model, tea.Cmd) {
	switch m.screen {
	case scopeScreen:
		selected := targetOptions[m.cursor]
		m.target = selected.target
		m.scope = selected.scope
		m.resolved = m.resolveConfigPath()
		m.screen = confirmScreen
	case confirmScreen:
		return m.handleConfirm()
	}
	return m, nil
}

// handleConfirm writes the config and transitions to done or error screen.
func (m *wizardModel) handleConfirm() (tea.Model, tea.Cmd) {
	err := m.writeConfig()
	if err != nil {
		m.err = err
		m.screen = errorScreen
	} else {
		m.screen = doneScreen
	}
	return m, tea.Quit
}

// resolveConfigPath returns the absolute config file path for the selected target+scope.
func (m *wizardModel) resolveConfigPath() string {
	switch m.target {
	case "opencode":
		if m.scope == "local" {
			return opencode.LocalPath()
		}
		return opencode.GlobalPath()
	case "claude":
		if m.scope == "local" {
			return claude.LocalPath()
		}
		return claude.GlobalPath()
	case "claude-desktop":
		path, _ := claudedesktop.Path()
		return path
	}
	return ""
}

// writeConfig performs the actual config file write for the selected target+scope.
func (m *wizardModel) writeConfig() error {
	switch m.target {
	case "opencode":
		scope := opencode.GlobalScope
		if m.scope == "local" {
			scope = opencode.LocalScope
		}
		return opencode.Save(scope, m.binPath)
	case "claude":
		scope := claude.GlobalScope
		if m.scope == "local" {
			scope = claude.LocalScope
		}
		return claude.Save(scope, m.binPath)
	case "claude-desktop":
		return claudedesktop.Save(m.binPath)
	}
	return fmt.Errorf("unknown target: %s", m.target)
}

// View implements tea.Model.
func (m *wizardModel) View() string {
	switch m.screen {
	case scopeScreen:
		return m.viewScopeScreen()
	case confirmScreen:
		return m.viewConfirmScreen()
	case doneScreen:
		return m.viewDoneScreen()
	case errorScreen:
		return m.viewErrorScreen()
	}
	return ""
}

func (m *wizardModel) viewScopeScreen() string {
	s := "jira-mcp Setup Wizard\n"
	s += "─────────────────────\n\n"
	s += "Select where to register jira-mcp:\n\n"

	for i, opt := range targetOptions {
		cursor := "  "
		if i == m.cursor {
			cursor = "▶ "
		}
		s += fmt.Sprintf("%s%s\n", cursor, opt.label)
	}

	s += "\n↑/↓ to navigate  Enter to select  Ctrl+C to cancel\n"
	return s
}

func (m *wizardModel) viewConfirmScreen() string {
	s := "jira-mcp Setup Wizard\n"
	s += "─────────────────────\n\n"
	s += fmt.Sprintf("Target:  %s (%s)\n", m.target, m.scope)
	s += fmt.Sprintf("Config:  %s\n\n", m.resolved)
	s += "Proceed? (y/n)  Ctrl+C to cancel\n"
	return s
}

func (m *wizardModel) viewDoneScreen() string {
	return fmt.Sprintf("\n✓ %s registered to:\n  %s\n\nSetup complete!\n", m.target, m.resolved)
}

func (m *wizardModel) viewErrorScreen() string {
	return fmt.Sprintf("\n✗ Failed to register %s:\n  %v\n\nPlease check permissions and try again.\n", m.target, m.err)
}
