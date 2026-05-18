package validation

import (
	"fmt"

	jiraclient "github.com/isaiahrafael/jira-go-mcp/internal/jira"
)

// ruleAllIssuesDone fails if any issue is not in a done status.
func (e *Engine) ruleAllIssuesDone(issues []jiraclient.Issue, result *Result) {
	var open []string
	for _, iss := range issues {
		if !e.doneStatuses[iss.Fields.Status.Name] {
			open = append(open, iss.Key)
		}
	}
	if len(open) > 0 {
		result.Ready = false
		result.Errors = append(result.Errors, fmt.Sprintf("all_issues_done: %d issue(s) not done: %v", len(open), open))
	}
}

// ruleNoCriticalBugsOpen fails if any open Bug has a critical label.
func (e *Engine) ruleNoCriticalBugsOpen(issues []jiraclient.Issue, result *Result) {
	for _, iss := range issues {
		if iss.Fields.IssueType.Name != "Bug" {
			continue
		}
		if e.doneStatuses[iss.Fields.Status.Name] {
			continue
		}
		for _, label := range iss.Fields.Labels {
			if e.criticalLabels[label] {
				result.Ready = false
				result.Errors = append(result.Errors, fmt.Sprintf("no_critical_bugs_open: %s is an open critical bug", iss.Key))
				break
			}
		}
	}
}

// ruleNoBlockingIssues fails if any open issue has a "blocker" label.
func (e *Engine) ruleNoBlockingIssues(issues []jiraclient.Issue, result *Result) {
	for _, iss := range issues {
		if e.doneStatuses[iss.Fields.Status.Name] {
			continue
		}
		for _, label := range iss.Fields.Labels {
			if label == "blocker" {
				result.Ready = false
				result.Errors = append(result.Errors, fmt.Sprintf("no_blocking_issues: %s has blocking label", iss.Key))
				break
			}
		}
	}
}

// ruleMinIssuesCount fails if the release has no issues at all.
func (e *Engine) ruleMinIssuesCount(issues []jiraclient.Issue, result *Result) {
	if len(issues) == 0 {
		result.Ready = false
		result.Errors = append(result.Errors, "min_issues_count: release has no issues linked")
	}
}

// ruleStatusCheck warns if any issue has a non-standard status (not in doneStatuses and not "In Progress").
func (e *Engine) ruleStatusCheck(issues []jiraclient.Issue, result *Result) {
	known := map[string]bool{"In Progress": true, "To Do": true, "Open": true, "Reopened": true}
	for _, iss := range issues {
		s := iss.Fields.Status.Name
		if !e.doneStatuses[s] && !known[s] {
			result.Warnings = append(result.Warnings, fmt.Sprintf("status_check: %s has unusual status %q", iss.Key, s))
		}
	}
}

// ruleCustomJQL is a placeholder — custom JQL requires runtime evaluation against Jira.
// In this MVP, it produces an informational warning.
func (e *Engine) ruleCustomJQL(_ []jiraclient.Issue, result *Result) {
	result.Warnings = append(result.Warnings, "custom_jql: requires Jira query execution (not supported in offline validation)")
}
