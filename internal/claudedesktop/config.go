// Package claudedesktop provides functions to register jira-mcp in Claude Desktop's
// configuration file. Claude Desktop is only supported on Windows and macOS.
package claudedesktop

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
)

// ErrUnsupported is returned when Claude Desktop is not supported on the current OS.
var ErrUnsupported = errors.New("claude desktop is not supported on this operating system")

// Path returns the OS-resolved path to the Claude Desktop configuration file.
// Windows: %APPDATA%\Claude\claude_desktop_config.json
// macOS:   ~/Library/Application Support/Claude/claude_desktop_config.json
// Linux:   returns "", ErrUnsupported
func Path() (string, error) {
	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				return "", err
			}
			appData = filepath.Join(home, "AppData", "Roaming")
		}
		return filepath.Join(appData, "Claude", "claude_desktop_config.json"), nil
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, "Library", "Application Support", "Claude", "claude_desktop_config.json"), nil
	default:
		return "", ErrUnsupported
	}
}

// Save writes the jira-mcp mcpServers entry to the OS-resolved Claude Desktop
// configuration file. Returns ErrUnsupported on Linux.
func Save(binPath string) error {
	path, err := Path()
	if err != nil {
		return err
	}
	return SaveToPath(path, binPath)
}

// SaveToPath writes the jira-mcp mcpServers entry to a specific config file path.
// It reads the existing file (or starts from {}) and merges the new entry,
// preserving all existing keys using map[string]json.RawMessage.
func SaveToPath(path, binPath string) error {
	// Read existing file or start with empty object
	existing := map[string]json.RawMessage{}
	if data, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(data, &existing); err != nil {
			existing = map[string]json.RawMessage{}
		}
	}

	// Read or create the "mcpServers" sub-object
	mcpMap := map[string]json.RawMessage{}
	if raw, ok := existing["mcpServers"]; ok {
		if err := json.Unmarshal(raw, &mcpMap); err != nil {
			mcpMap = map[string]json.RawMessage{}
		}
	}

	// Build the jira-mcp entry
	entry := map[string]interface{}{
		"command": binPath,
		"args":    []string{"mcp"},
	}
	entryRaw, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	mcpMap["jira-mcp"] = json.RawMessage(entryRaw)

	mcpRaw, err := json.Marshal(mcpMap)
	if err != nil {
		return err
	}
	existing["mcpServers"] = json.RawMessage(mcpRaw)

	out, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		return err
	}

	// Ensure parent directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	return os.WriteFile(path, out, 0o644)
}

// Check returns whether the jira-mcp key is present in the mcpServers section
// of the Claude Desktop configuration file at the given path.
func Check(path string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}

	var outer map[string]json.RawMessage
	if err := json.Unmarshal(data, &outer); err != nil {
		return false, nil
	}

	mcpRaw, ok := outer["mcpServers"]
	if !ok {
		return false, nil
	}

	var mcpMap map[string]json.RawMessage
	if err := json.Unmarshal(mcpRaw, &mcpMap); err != nil {
		return false, nil
	}

	_, ok = mcpMap["jira-mcp"]
	return ok, nil
}
