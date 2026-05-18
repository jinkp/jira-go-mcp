package config

import (
	"fmt"
	"os"
	"strings"
)

// Load reads configuration from environment variables and returns a Config.
// It fails fast if any required variable is missing.
func Load() (*Config, error) {
	fileValues, err := LoadFile(FilePath())
	if err != nil {
		return nil, err
	}

	required := []string{"JIRA_BASE_URL", "JIRA_EMAIL", "JIRA_API_TOKEN"}
	for _, key := range required {
		if envOrFile(key, fileValues) == "" {
			return nil, fmt.Errorf("required environment variable %s is not set", key)
		}
	}

	doneStatuses := splitCSV(envOrFile("JIRA_DONE_STATUSES", fileValues), "Done")
	criticalLabels := splitCSV(envOrFile("JIRA_CRITICAL_LABELS", fileValues), "critical")

	return &Config{
		BaseURL:        envOrFile("JIRA_BASE_URL", fileValues),
		Email:          envOrFile("JIRA_EMAIL", fileValues),
		APIToken:       envOrFile("JIRA_API_TOKEN", fileValues),
		DefaultProject: envOrFile("JIRA_DEFAULT_PROJECT", fileValues),
		DoneStatuses:   doneStatuses,
		CriticalLabels: criticalLabels,
	}, nil
}

func envOrFile(key string, fileValues map[string]string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fileValues[key]
}

// splitCSV splits a comma-separated value string; returns defaultVal slice if empty.
func splitCSV(val, defaultVal string) []string {
	if val == "" {
		return []string{defaultVal}
	}
	parts := strings.Split(val, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}
