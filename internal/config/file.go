package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const configEnvFileName = "jira-go-mcp.env"

// FilePath returns the external env-style config file used by jira-mcp.
// Example: ~/.mcp/jira-go-mcp.env
func FilePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".mcp", configEnvFileName)
}

// LoadFile reads the external env-style config file.
// Missing file is not an error and returns an empty map.
func LoadFile(path string) (map[string]string, error) {
	values := map[string]string{}
	if path == "" {
		return values, nil
	}

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return values, nil
		}
		return nil, fmt.Errorf("read config file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		val = strings.Trim(val, `"`)
		values[key] = val
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan config file: %w", err)
	}

	return values, nil
}

// SaveFile writes jira-mcp configuration to the external env-style config file.
func SaveFile(cfg *Config) error {
	path := FilePath()
	if path == "" {
		return fmt.Errorf("could not resolve config file path")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	lines := []string{
		fmt.Sprintf("JIRA_BASE_URL=%s", cfg.BaseURL),
		fmt.Sprintf("JIRA_EMAIL=%s", cfg.Email),
		fmt.Sprintf("JIRA_API_TOKEN=%s", cfg.APIToken),
		fmt.Sprintf("JIRA_DEFAULT_PROJECT=%s", cfg.DefaultProject),
		fmt.Sprintf("JIRA_DONE_STATUSES=%s", strings.Join(cfg.DoneStatuses, ",")),
		fmt.Sprintf("JIRA_CRITICAL_LABELS=%s", strings.Join(cfg.CriticalLabels, ",")),
	}

	content := strings.Join(lines, "\n") + "\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}
	return nil
}
