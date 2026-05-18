package config

// Config holds all runtime configuration loaded from environment variables.
type Config struct {
	BaseURL        string
	Email          string
	APIToken       string
	DefaultProject string
	DoneStatuses   []string
	CriticalLabels []string
}
