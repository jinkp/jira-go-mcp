package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jinkp/jira-go-mcp/internal/config"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name        string
		env         map[string]string
		wantErr     bool
		errContains string
		wantCfg     *config.Config
	}{
		{
			name: "all required vars present",
			env: map[string]string{
				"JIRA_BASE_URL":   "https://acme.atlassian.net",
				"JIRA_EMAIL":      "dev@acme.com",
				"JIRA_API_TOKEN":  "token123",
				"JIRA_DEFAULT_PROJECT": "PROJ",
			},
			wantErr: false,
			wantCfg: &config.Config{
				BaseURL:         "https://acme.atlassian.net",
				Email:           "dev@acme.com",
				APIToken:        "token123",
				DefaultProject:  "PROJ",
				DoneStatuses:    []string{"Done"},
				CriticalLabels:  []string{"critical"},
			},
		},
		{
			name: "missing JIRA_BASE_URL returns error naming the var",
			env: map[string]string{
				"JIRA_EMAIL":     "dev@acme.com",
				"JIRA_API_TOKEN": "token123",
			},
			wantErr:     true,
			errContains: "JIRA_BASE_URL",
		},
		{
			name: "missing JIRA_EMAIL returns error naming the var",
			env: map[string]string{
				"JIRA_BASE_URL":  "https://acme.atlassian.net",
				"JIRA_API_TOKEN": "token123",
			},
			wantErr:     true,
			errContains: "JIRA_EMAIL",
		},
		{
			name: "missing JIRA_API_TOKEN returns error naming the var",
			env: map[string]string{
				"JIRA_BASE_URL": "https://acme.atlassian.net",
				"JIRA_EMAIL":    "dev@acme.com",
			},
			wantErr:     true,
			errContains: "JIRA_API_TOKEN",
		},
		{
			name: "optional vars default when not set",
			env: map[string]string{
				"JIRA_BASE_URL":  "https://acme.atlassian.net",
				"JIRA_EMAIL":     "dev@acme.com",
				"JIRA_API_TOKEN": "token123",
			},
			wantErr: false,
			wantCfg: &config.Config{
				BaseURL:        "https://acme.atlassian.net",
				Email:          "dev@acme.com",
				APIToken:       "token123",
				DoneStatuses:   []string{"Done"},
				CriticalLabels: []string{"critical"},
			},
		},
		{
			name: "custom JIRA_DONE_STATUSES comma-separated",
			env: map[string]string{
				"JIRA_BASE_URL":        "https://acme.atlassian.net",
				"JIRA_EMAIL":           "dev@acme.com",
				"JIRA_API_TOKEN":       "token123",
				"JIRA_DONE_STATUSES":   "Done,Closed,Released",
				"JIRA_CRITICAL_LABELS": "critical,blocker",
			},
			wantErr: false,
			wantCfg: &config.Config{
				BaseURL:        "https://acme.atlassian.net",
				Email:          "dev@acme.com",
				APIToken:       "token123",
				DoneStatuses:   []string{"Done", "Closed", "Released"},
				CriticalLabels: []string{"critical", "blocker"},
			},
		},
		{
			name:    "loads required vars from external file when env is absent",
			env:     map[string]string{},
			wantErr: false,
			wantCfg: &config.Config{
				BaseURL:        "https://file.atlassian.net",
				Email:          "file@acme.com",
				APIToken:       "filetoken",
				DefaultProject: "FILE",
				DoneStatuses:   []string{"Done", "Closed"},
				CriticalLabels: []string{"critical", "blocker"},
			},
		},
	}

	for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				// isolate env per test
				setEnv(t, tt.env)
				if tt.name == "loads required vars from external file when env is absent" {
					home := t.TempDir()
					t.Setenv("HOME", home)
					t.Setenv("USERPROFILE", home)
					if err := os.MkdirAll(filepath.Join(home, ".mcp"), 0o755); err != nil {
						t.Fatalf("MkdirAll() error = %v", err)
					}
					content := "JIRA_BASE_URL=https://file.atlassian.net\nJIRA_EMAIL=file@acme.com\nJIRA_API_TOKEN=filetoken\nJIRA_DEFAULT_PROJECT=FILE\nJIRA_DONE_STATUSES=Done,Closed\nJIRA_CRITICAL_LABELS=critical,blocker\n"
					if err := os.WriteFile(filepath.Join(home, ".mcp", "jira-go-mcp.env"), []byte(content), 0o600); err != nil {
						t.Fatalf("WriteFile() error = %v", err)
					}
				}

				got, err := config.Load()

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.errContains)
				}
				if tt.errContains != "" {
					if !contains(err.Error(), tt.errContains) {
						t.Errorf("error = %q, want it to contain %q", err.Error(), tt.errContains)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.BaseURL != tt.wantCfg.BaseURL {
				t.Errorf("BaseURL = %q, want %q", got.BaseURL, tt.wantCfg.BaseURL)
			}
			if got.Email != tt.wantCfg.Email {
				t.Errorf("Email = %q, want %q", got.Email, tt.wantCfg.Email)
			}
			if got.APIToken != tt.wantCfg.APIToken {
				t.Errorf("APIToken = %q, want %q", got.APIToken, tt.wantCfg.APIToken)
			}
			if !equalSlices(got.DoneStatuses, tt.wantCfg.DoneStatuses) {
				t.Errorf("DoneStatuses = %v, want %v", got.DoneStatuses, tt.wantCfg.DoneStatuses)
			}
			if !equalSlices(got.CriticalLabels, tt.wantCfg.CriticalLabels) {
				t.Errorf("CriticalLabels = %v, want %v", got.CriticalLabels, tt.wantCfg.CriticalLabels)
			}
		})
	}
}

// setEnv clears all Jira env vars and sets only those in env.
func setEnv(t *testing.T, env map[string]string) {
	t.Helper()
	keys := []string{
		"JIRA_BASE_URL", "JIRA_EMAIL", "JIRA_API_TOKEN",
		"JIRA_DEFAULT_PROJECT", "JIRA_DONE_STATUSES", "JIRA_CRITICAL_LABELS",
	}
	for _, k := range keys {
		t.Setenv(k, "")
	}
	for k, v := range env {
		t.Setenv(k, v)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}

func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
