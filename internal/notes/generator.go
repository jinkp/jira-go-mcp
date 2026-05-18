package notes

import (
	"fmt"
	"sort"
	"strings"

	jiraclient "github.com/isaiahrafael/jira-go-mcp/internal/jira"
)

// Generate produces Markdown release notes grouped by issue type.
// Groups are sorted alphabetically; issues within a group maintain input order.
func Generate(issues []jiraclient.Issue, releaseName string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Release Notes: %s\n", releaseName))

	if len(issues) == 0 {
		sb.WriteString("\n_No issues linked to this release._\n")
		return sb.String()
	}

	// Group issues by type
	groups := make(map[string][]jiraclient.Issue)
	order := []string{}
	seen := make(map[string]bool)

	for _, iss := range issues {
		t := iss.Fields.IssueType.Name
		if !seen[t] {
			seen[t] = true
			order = append(order, t)
		}
		groups[t] = append(groups[t], iss)
	}

	// Sort group names alphabetically for deterministic output
	sort.Strings(order)

	for _, typeName := range order {
		sb.WriteString(fmt.Sprintf("\n## %s\n\n", typeName))
		for _, iss := range groups[typeName] {
			sb.WriteString(fmt.Sprintf("- %s: %s\n", iss.Key, iss.Fields.Summary))
		}
	}

	return sb.String()
}
