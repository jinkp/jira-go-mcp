package jira

// Version represents a Jira project version (release).
type Version struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Released    bool   `json:"released"`
	Archived    bool   `json:"archived"`
	ReleaseDate string `json:"releaseDate,omitempty"`
	StartDate   string `json:"startDate,omitempty"`
	ProjectID   int    `json:"projectId,omitempty"`
}

// StatusField represents an issue status.
type StatusField struct {
	Name string `json:"name"`
}

// IssueTypeField represents an issue type.
type IssueTypeField struct {
	Name string `json:"name"`
}

// IssueFields holds the fields of a Jira issue.
type IssueFields struct {
	Summary   string         `json:"summary"`
	Status    StatusField    `json:"status"`
	IssueType IssueTypeField `json:"issuetype"`
	Labels    []string       `json:"labels,omitempty"`
	Priority  struct {
		Name string `json:"name"`
	} `json:"priority,omitempty"`
}

// Issue represents a Jira issue.
type Issue struct {
	Key    string      `json:"key"`
	Fields IssueFields `json:"fields"`
}

// SearchResult is the response from Jira's search/jql endpoint.
type SearchResult struct {
	Issues []Issue `json:"issues"`
	Total  int     `json:"total"`
}

// RelatedIssueCounts holds counts of issues related to a version.
type RelatedIssueCounts struct {
	IssuesUnresolved int `json:"issuesUnresolved"`
	IssuesFixedCount int `json:"issuesFixedCount"`
}

// CreateVersionInput holds parameters for creating a version.
type CreateVersionInput struct {
	ProjectKey  string `json:"project"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	ReleaseDate string `json:"releaseDate,omitempty"`
	StartDate   string `json:"startDate,omitempty"`
}

// UpdateVersionInput holds parameters for updating a version.
type UpdateVersionInput struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Released    *bool  `json:"released,omitempty"`
	Archived    *bool  `json:"archived,omitempty"`
	ReleaseDate string `json:"releaseDate,omitempty"`
}
