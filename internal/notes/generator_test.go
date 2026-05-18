package notes_test

import (
	"os"
	"path/filepath"
	"testing"

	jiraclient "github.com/jinkp/jira-go-mcp/internal/jira"
	"github.com/jinkp/jira-go-mcp/internal/notes"
)

func makeIssue(key, itype, summary string) jiraclient.Issue {
	return jiraclient.Issue{
		Key: key,
		Fields: jiraclient.IssueFields{
			Summary:   summary,
			IssueType: jiraclient.IssueTypeField{Name: itype},
		},
	}
}

func TestGenerate_MixedTypes(t *testing.T) {
	issues := []jiraclient.Issue{
		makeIssue("PROJ-1", "Bug", "Fix login crash"),
		makeIssue("PROJ-2", "Bug", "Fix null pointer"),
		makeIssue("PROJ-3", "Story", "User profile page"),
		makeIssue("PROJ-4", "Story", "Dashboard improvements"),
		makeIssue("PROJ-5", "Story", "Settings panel"),
	}

	got := notes.Generate(issues, "v1.0")

	golden := readGolden(t, "notes_mixed_types.golden")
	if got != golden {
		t.Errorf("output mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, golden)
	}
}

func TestGenerate_SingleType(t *testing.T) {
	issues := []jiraclient.Issue{
		makeIssue("PROJ-10", "Task", "Set up CI pipeline"),
		makeIssue("PROJ-11", "Task", "Add lint checks"),
	}

	got := notes.Generate(issues, "v2.0")

	golden := readGolden(t, "notes_single_type.golden")
	if got != golden {
		t.Errorf("output mismatch:\n--- got ---\n%s\n--- want ---\n%s", got, golden)
	}
}

func TestGenerate_EmptyIssues(t *testing.T) {
	got := notes.Generate([]jiraclient.Issue{}, "v0.0")
	if got == "" {
		t.Error("expected non-empty output even for empty issues (at minimum the header)")
	}
	// Must include the release name
	if !containsStr(got, "v0.0") {
		t.Errorf("output should contain release name v0.0, got: %s", got)
	}
}

func readGolden(t *testing.T, filename string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", filename))
	if err != nil {
		t.Fatalf("reading golden file %s: %v", filename, err)
	}
	return string(b)
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
