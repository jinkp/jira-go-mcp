package config_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/jinkp/jira-go-mcp/internal/config"
)

func TestLoadFile_MissingFileReturnsEmptyMap(t *testing.T) {
	got, err := config.LoadFile(filepath.Join(t.TempDir(), "missing.env"))
	if err != nil {
		t.Fatalf("LoadFile() error = %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("LoadFile() len = %d, want 0", len(got))
	}
}

func TestLoadFile_ParsesEnvStyleConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "jira-go-mcp.env")
	content := "# comment\nJIRA_BASE_URL=https://acme.atlassian.net\nJIRA_EMAIL=dev@acme.com\nJIRA_API_TOKEN=token123\nJIRA_DONE_STATUSES=Done,Closed\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	got, err := config.LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile() error = %v", err)
	}
	if got["JIRA_BASE_URL"] != "https://acme.atlassian.net" {
		t.Errorf("JIRA_BASE_URL = %q", got["JIRA_BASE_URL"])
	}
	if got["JIRA_EMAIL"] != "dev@acme.com" {
		t.Errorf("JIRA_EMAIL = %q", got["JIRA_EMAIL"])
	}
	if got["JIRA_API_TOKEN"] != "token123" {
		t.Errorf("JIRA_API_TOKEN = %q", got["JIRA_API_TOKEN"])
	}
	if got["JIRA_DONE_STATUSES"] != "Done,Closed" {
		t.Errorf("JIRA_DONE_STATUSES = %q", got["JIRA_DONE_STATUSES"])
	}
}

func TestSaveFile_WritesExternalConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	cfg := &config.Config{
		BaseURL:        "https://acme.atlassian.net",
		Email:          "dev@acme.com",
		APIToken:       "token123",
		DefaultProject: "PROJ",
		DoneStatuses:   []string{"Done", "Closed"},
		CriticalLabels: []string{"critical", "blocker"},
	}

	if err := config.SaveFile(cfg); err != nil {
		t.Fatalf("SaveFile() error = %v", err)
	}

	path := filepath.Join(home, ".mcp", "jira-go-mcp.env")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	text := string(data)
	for _, want := range []string{
		"JIRA_BASE_URL=https://acme.atlassian.net",
		"JIRA_EMAIL=dev@acme.com",
		"JIRA_API_TOKEN=token123",
		"JIRA_DEFAULT_PROJECT=PROJ",
		"JIRA_DONE_STATUSES=Done,Closed",
		"JIRA_CRITICAL_LABELS=critical,blocker",
	} {
		if !contains(text, want) {
			t.Errorf("written file missing %q", want)
		}
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat() error = %v", err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0o600 {
		t.Errorf("file mode = %o, want 600", info.Mode().Perm())
	}
}
