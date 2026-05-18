// Package opencode provides functions to register jira-mcp in OpenCode's
// configuration file using a safe JSON merge that preserves all existing keys.
package opencode

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Scope represents the registration scope (global or local).
type Scope string

const (
	GlobalScope Scope = "global"
	LocalScope  Scope = "local"
)

// GlobalPath returns the path to the OpenCode global configuration file.
// On Unix: ~/.config/opencode/opencode.json
// On Windows: %USERPROFILE%/.config/opencode/opencode.json (os.UserHomeDir)
func GlobalPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "opencode", "opencode.json")
}

// LocalPath returns the path to the OpenCode local configuration file
// in the current working directory under .opencode/opencode.json.
func LocalPath() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return filepath.Join(cwd, ".opencode", "opencode.json")
}

// Save writes the jira-mcp MCP entry to the OpenCode configuration file
// at the path determined by scope (global or local).
func Save(scope Scope, binPath string) error {
	var path string
	if scope == LocalScope {
		path = LocalPath()
	} else {
		path = GlobalPath()
	}
	return SaveToPath(path, binPath)
}

// SaveToPath writes the jira-mcp MCP entry to a specific config file path.
// It reads the existing file (or starts from {}) and merges the new entry,
// preserving all existing keys using map[string]json.RawMessage.
func SaveToPath(path, binPath string) error {
	// Read existing file or start with empty object
	existing := map[string]json.RawMessage{}
	if data, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(data, &existing); err != nil {
			// If the file is malformed, start fresh
			existing = map[string]json.RawMessage{}
		}
	}

	// Read or create the "mcp" sub-object
	mcpMap := map[string]json.RawMessage{}
	if raw, ok := existing["mcp"]; ok {
		if err := json.Unmarshal(raw, &mcpMap); err != nil {
			mcpMap = map[string]json.RawMessage{}
		}
	}

	// Build the jira-mcp entry
	entry := map[string]interface{}{
		"type":    "local",
		"command": binPath,
		"args":    []string{"mcp"},
	}
	entryRaw, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	mcpMap["jira-mcp"] = json.RawMessage(entryRaw)

	// Re-marshal the mcp sub-object and set it back
	mcpRaw, err := json.Marshal(mcpMap)
	if err != nil {
		return err
	}
	existing["mcp"] = json.RawMessage(mcpRaw)

	// Marshal the final result with indentation for readability
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

// Check returns true if the jira-mcp key is present in the mcp section
// of the OpenCode configuration file at the given path.
func Check(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}

	var outer map[string]json.RawMessage
	if err := json.Unmarshal(data, &outer); err != nil {
		return false
	}

	mcpRaw, ok := outer["mcp"]
	if !ok {
		return false
	}

	var mcpMap map[string]json.RawMessage
	if err := json.Unmarshal(mcpRaw, &mcpMap); err != nil {
		return false
	}

	_, ok = mcpMap["jira-mcp"]
	return ok
}
