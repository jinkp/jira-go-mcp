package config

import (
	"fmt"
	"os"
	"strings"
)

// Load reads configuration from environment variables and returns a Config.
// It fails fast if any required variable is missing.
func Load() (*Config, error) {
	required := []string{"JIRA_BASE_URL", "JIRA_EMAIL", "JIRA_API_TOKEN"}
	for _, key := range required {
		if os.Getenv(key) == "" {
			return nil, fmt.Errorf("required environment variable %s is not set", key)
		}
	}

	doneStatuses := splitCSV(os.Getenv("JIRA_DONE_STATUSES"), "Done")
	criticalLabels := splitCSV(os.Getenv("JIRA_CRITICAL_LABELS"), "critical")

	return &Config{
		BaseURL:        os.Getenv("JIRA_BASE_URL"),
		Email:          os.Getenv("JIRA_EMAIL"),
		APIToken:       os.Getenv("JIRA_API_TOKEN"),
		DefaultProject: os.Getenv("JIRA_DEFAULT_PROJECT"),
		DoneStatuses:   doneStatuses,
		CriticalLabels: criticalLabels,
	}, nil
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
