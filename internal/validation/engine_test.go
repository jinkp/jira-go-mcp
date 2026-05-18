package validation_test

import (
	"testing"

	jiraclient "github.com/isaiahrafael/jira-go-mcp/internal/jira"
	"github.com/isaiahrafael/jira-go-mcp/internal/validation"
)

func issue(key, status, itype string, labels ...string) jiraclient.Issue {
	return jiraclient.Issue{
		Key: key,
		Fields: jiraclient.IssueFields{
			Summary:   "Issue " + key,
			Status:    jiraclient.StatusField{Name: status},
			IssueType: jiraclient.IssueTypeField{Name: itype},
			Labels:    labels,
		},
	}
}

func TestEngine_Evaluate_AllIssuesDone(t *testing.T) {
	engine := validation.NewEngine([]string{"Done", "Closed"}, []string{"critical"})

	tests := []struct {
		name      string
		issues    []jiraclient.Issue
		wantReady bool
		wantErrs  int
	}{
		{
			name:      "all done — passes",
			issues:    []jiraclient.Issue{issue("P-1", "Done", "Bug"), issue("P-2", "Closed", "Story")},
			wantReady: true,
			wantErrs:  0,
		},
		{
			name:      "one open issue — fails",
			issues:    []jiraclient.Issue{issue("P-1", "Done", "Bug"), issue("P-2", "In Progress", "Story")},
			wantReady: false,
			wantErrs:  1,
		},
		{
			name:      "all open — fails with count",
			issues:    []jiraclient.Issue{issue("P-1", "Open", "Bug"), issue("P-2", "Open", "Story")},
			wantReady: false,
			wantErrs:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.Evaluate(tt.issues, []string{"all_issues_done"})
			if result.Ready != tt.wantReady {
				t.Errorf("Ready = %v, want %v", result.Ready, tt.wantReady)
			}
			if len(result.Errors) != tt.wantErrs {
				t.Errorf("Errors = %v (count %d), want %d errors", result.Errors, len(result.Errors), tt.wantErrs)
			}
		})
	}
}

func TestEngine_Evaluate_NoCriticalBugsOpen(t *testing.T) {
	engine := validation.NewEngine([]string{"Done"}, []string{"critical"})

	tests := []struct {
		name      string
		issues    []jiraclient.Issue
		wantReady bool
	}{
		{
			name:      "no critical bugs — passes",
			issues:    []jiraclient.Issue{issue("P-1", "Open", "Bug"), issue("P-2", "Done", "Bug", "critical")},
			wantReady: true,
		},
		{
			name:      "open critical bug — fails",
			issues:    []jiraclient.Issue{issue("P-1", "Open", "Bug", "critical")},
			wantReady: false,
		},
		{
			name:      "open critical story (not bug) — passes",
			issues:    []jiraclient.Issue{issue("P-1", "Open", "Story", "critical")},
			wantReady: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.Evaluate(tt.issues, []string{"no_critical_bugs_open"})
			if result.Ready != tt.wantReady {
				t.Errorf("Ready = %v, want %v", result.Ready, tt.wantReady)
			}
		})
	}
}

func TestEngine_Evaluate_NoBlockingIssues(t *testing.T) {
	engine := validation.NewEngine([]string{"Done"}, []string{"critical"})

	tests := []struct {
		name      string
		issues    []jiraclient.Issue
		wantReady bool
	}{
		{
			name:      "no blocking issues — passes",
			issues:    []jiraclient.Issue{issue("P-1", "Done", "Bug")},
			wantReady: true,
		},
		{
			name:      "blocking issue present — fails",
			issues:    []jiraclient.Issue{issue("P-1", "Open", "Bug", "blocker")},
			wantReady: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.Evaluate(tt.issues, []string{"no_blocking_issues"})
			if result.Ready != tt.wantReady {
				t.Errorf("Ready = %v, want %v", result.Ready, tt.wantReady)
			}
		})
	}
}

func TestEngine_Evaluate_MinIssuesCount(t *testing.T) {
	engine := validation.NewEngine([]string{"Done"}, []string{"critical"})

	tests := []struct {
		name      string
		issues    []jiraclient.Issue
		wantReady bool
	}{
		{
			name:      "has issues — passes min_issues_count",
			issues:    []jiraclient.Issue{issue("P-1", "Done", "Bug")},
			wantReady: true,
		},
		{
			name:      "empty issues — fails min_issues_count",
			issues:    []jiraclient.Issue{},
			wantReady: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.Evaluate(tt.issues, []string{"min_issues_count"})
			if result.Ready != tt.wantReady {
				t.Errorf("Ready = %v, want %v", result.Ready, tt.wantReady)
			}
		})
	}
}

func TestEngine_Evaluate_MultipleRules(t *testing.T) {
	engine := validation.NewEngine([]string{"Done"}, []string{"critical"})

	tests := []struct {
		name      string
		issues    []jiraclient.Issue
		rules     []string
		wantReady bool
	}{
		{
			name:      "all rules pass",
			issues:    []jiraclient.Issue{issue("P-1", "Done", "Bug"), issue("P-2", "Done", "Story")},
			rules:     []string{"all_issues_done", "min_issues_count"},
			wantReady: true,
		},
		{
			name:      "one rule fails — overall not ready",
			issues:    []jiraclient.Issue{issue("P-1", "Open", "Bug", "critical"), issue("P-2", "Done", "Story")},
			rules:     []string{"all_issues_done", "no_critical_bugs_open"},
			wantReady: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.Evaluate(tt.issues, tt.rules)
			if result.Ready != tt.wantReady {
				t.Errorf("Ready = %v, want %v, errors: %v", result.Ready, tt.wantReady, result.Errors)
			}
		})
	}
}

func TestEngine_Evaluate_StatusCheck(t *testing.T) {
	engine := validation.NewEngine([]string{"Done"}, []string{"critical"})

	tests := []struct {
		name         string
		issues       []jiraclient.Issue
		wantWarnings int
		wantReady    bool
	}{
		{
			name:         "standard statuses produce no warnings",
			issues:       []jiraclient.Issue{issue("P-1", "In Progress", "Bug"), issue("P-2", "Done", "Story")},
			wantWarnings: 0,
			wantReady:    true,
		},
		{
			name:         "unusual status triggers warning",
			issues:       []jiraclient.Issue{issue("P-1", "Weird Custom Status", "Bug")},
			wantWarnings: 1,
			wantReady:    true, // warning only — does not fail the release
		},
		{
			name:         "multiple unusual statuses produce multiple warnings",
			issues:       []jiraclient.Issue{issue("P-1", "Needs Review", "Bug"), issue("P-2", "Pending Approval", "Story")},
			wantWarnings: 2,
			wantReady:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.Evaluate(tt.issues, []string{"status_check"})
			if result.Ready != tt.wantReady {
				t.Errorf("Ready = %v, want %v", result.Ready, tt.wantReady)
			}
			if len(result.Warnings) != tt.wantWarnings {
				t.Errorf("Warnings = %v (count %d), want %d warnings", result.Warnings, len(result.Warnings), tt.wantWarnings)
			}
		})
	}
}

func TestEngine_Evaluate_UnknownRule(t *testing.T) {
	engine := validation.NewEngine([]string{"Done"}, []string{"critical"})
	issues := []jiraclient.Issue{issue("P-1", "Done", "Bug")}

	// Unknown rules should produce a warning, not crash
	result := engine.Evaluate(issues, []string{"nonexistent_rule"})
	if len(result.Warnings) == 0 {
		t.Error("expected warning for unknown rule, got none")
	}
}
