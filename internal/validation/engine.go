package validation

import (
	"fmt"

	jiraclient "github.com/isaiahrafael/jira-go-mcp/internal/jira"
)

// Result holds the outcome of a validation run.
type Result struct {
	Ready    bool     `json:"ready"`
	Errors   []string `json:"errors"`
	Warnings []string `json:"warnings"`
}

// Rule is a validation function that checks issues and appends errors/warnings.
type Rule func(issues []jiraclient.Issue, result *Result)

// Engine holds registered rules and evaluates them against issue sets.
type Engine struct {
	rules          map[string]Rule
	doneStatuses   map[string]bool
	criticalLabels map[string]bool
}

// NewEngine creates an Engine with the given done statuses and critical labels.
func NewEngine(doneStatuses, criticalLabels []string) *Engine {
	e := &Engine{
		rules:          make(map[string]Rule),
		doneStatuses:   toSet(doneStatuses),
		criticalLabels: toSet(criticalLabels),
	}
	e.registerBuiltins()
	return e
}

// Evaluate runs the named rules against the given issues and returns the result.
func (e *Engine) Evaluate(issues []jiraclient.Issue, ruleNames []string) *Result {
	result := &Result{Ready: true}
	for _, name := range ruleNames {
		rule, ok := e.rules[name]
		if !ok {
			result.Warnings = append(result.Warnings, fmt.Sprintf("unknown rule: %s", name))
			continue
		}
		rule(issues, result)
	}
	return result
}

// registerBuiltins registers all built-in validation rules.
func (e *Engine) registerBuiltins() {
	e.rules["all_issues_done"] = e.ruleAllIssuesDone
	e.rules["no_critical_bugs_open"] = e.ruleNoCriticalBugsOpen
	e.rules["no_blocking_issues"] = e.ruleNoBlockingIssues
	e.rules["min_issues_count"] = e.ruleMinIssuesCount
	e.rules["status_check"] = e.ruleStatusCheck
	e.rules["custom_jql"] = e.ruleCustomJQL
}

func toSet(values []string) map[string]bool {
	m := make(map[string]bool, len(values))
	for _, v := range values {
		m[v] = true
	}
	return m
}
