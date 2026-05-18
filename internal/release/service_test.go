package release_test

import (
	"context"
	"errors"
	"testing"

	jiraclient "github.com/jinkp/jira-go-mcp/internal/jira"
	"github.com/jinkp/jira-go-mcp/internal/release"
)

// mockJiraClient is a configurable test double for JiraClient.
type mockJiraClient struct {
	versions    []jiraclient.Version
	issues      []jiraclient.Issue
	created     *jiraclient.Version
	updated     *jiraclient.Version
	counts      *jiraclient.RelatedIssueCounts
	errVersions error
	errCreate   error
	errUpdate   error
	errSearch   error
	errCounts   error
}

func (m *mockJiraClient) GetProjectVersions(_ context.Context, _ string) ([]jiraclient.Version, error) {
	return m.versions, m.errVersions
}

func (m *mockJiraClient) CreateVersion(_ context.Context, _ jiraclient.CreateVersionInput) (*jiraclient.Version, error) {
	return m.created, m.errCreate
}

func (m *mockJiraClient) UpdateVersion(_ context.Context, _ string, _ jiraclient.UpdateVersionInput) (*jiraclient.Version, error) {
	return m.updated, m.errUpdate
}

func (m *mockJiraClient) SearchIssues(_ context.Context, _ string, _ []string) ([]jiraclient.Issue, error) {
	return m.issues, m.errSearch
}

func (m *mockJiraClient) GetRelatedIssueCounts(_ context.Context, _ string) (*jiraclient.RelatedIssueCounts, error) {
	return m.counts, m.errCounts
}

// ---- List tests ----

func TestService_List(t *testing.T) {
	tests := []struct {
		name             string
		mock             *mockJiraClient
		includeArchived  bool
		includeReleased  bool
		wantCount        int
		wantErr          bool
	}{
		{
			name: "returns all active versions",
			mock: &mockJiraClient{
				versions: []jiraclient.Version{
					{ID: "1", Name: "v1.0", Released: false, Archived: false},
					{ID: "2", Name: "v2.0", Released: false, Archived: false},
				},
				counts: &jiraclient.RelatedIssueCounts{IssuesUnresolved: 3},
			},
			wantCount: 2,
		},
		{
			name: "excludes archived when includeArchived=false",
			mock: &mockJiraClient{
				versions: []jiraclient.Version{
					{ID: "1", Name: "v1.0", Released: false, Archived: false},
					{ID: "2", Name: "old", Released: false, Archived: true},
				},
				counts: &jiraclient.RelatedIssueCounts{},
			},
			includeArchived: false,
			wantCount:       1,
		},
		{
			name: "excludes released when includeReleased=false",
			mock: &mockJiraClient{
				versions: []jiraclient.Version{
					{ID: "1", Name: "v1.0", Released: false, Archived: false},
					{ID: "2", Name: "v0.9", Released: true, Archived: false},
				},
				counts: &jiraclient.RelatedIssueCounts{},
			},
			includeReleased: false,
			wantCount:       1,
		},
		{
			name: "propagates jira client error",
			mock: &mockJiraClient{
				errVersions: errors.New("network error"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := release.NewService(tt.mock, nil)
			versions, err := svc.List(context.Background(), "PROJ", tt.includeArchived, tt.includeReleased)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(versions) != tt.wantCount {
				t.Errorf("got %d versions, want %d", len(versions), tt.wantCount)
			}
		})
	}
}

// ---- Create tests ----

func TestService_Create(t *testing.T) {
	tests := []struct {
		name     string
		existing []jiraclient.Version
		input    release.CreateInput
		created  *jiraclient.Version
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "creates a new version successfully",
			existing: []jiraclient.Version{{ID: "1", Name: "v1.0"}},
			input:    release.CreateInput{ProjectKey: "PROJ", Name: "v2.0"},
			created:  &jiraclient.Version{ID: "99", Name: "v2.0"},
			wantErr:  false,
		},
		{
			name:     "rejects duplicate version name",
			existing: []jiraclient.Version{{ID: "1", Name: "v1.0"}},
			input:    release.CreateInput{ProjectKey: "PROJ", Name: "v1.0"},
			wantErr:  true,
			errMsg:   "v1.0",
		},
		{
			name:     "propagates create error from jira",
			existing: []jiraclient.Version{},
			input:    release.CreateInput{ProjectKey: "PROJ", Name: "v3.0"},
			created:  nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var createErr error
			if tt.wantErr && tt.errMsg == "" && len(tt.existing) == 0 {
				createErr = errors.New("jira error")
			}
			mock := &mockJiraClient{
				versions: tt.existing,
				created:  tt.created,
				errCreate: createErr,
			}
			svc := release.NewService(mock, nil)
			got, err := svc.Create(context.Background(), tt.input)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errMsg != "" {
					found := false
					for i := 0; i <= len(err.Error())-len(tt.errMsg); i++ {
						if err.Error()[i:i+len(tt.errMsg)] == tt.errMsg {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("error %q should contain %q", err.Error(), tt.errMsg)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Name != tt.created.Name {
				t.Errorf("Name = %q, want %q", got.Name, tt.created.Name)
			}
		})
	}
}

// ---- GetIssues tests ----

func TestService_GetIssues(t *testing.T) {
	tests := []struct {
		name      string
		mock      *mockJiraClient
		wantCount int
		wantErr   bool
	}{
		{
			name: "returns issues for a version",
			mock: &mockJiraClient{
				issues: []jiraclient.Issue{
					{Key: "PROJ-1", Fields: jiraclient.IssueFields{Summary: "Bug fix"}},
					{Key: "PROJ-2", Fields: jiraclient.IssueFields{Summary: "Feature"}},
					{Key: "PROJ-3", Fields: jiraclient.IssueFields{Summary: "Chore"}},
				},
			},
			wantCount: 3,
		},
		{
			name:    "propagates search error",
			mock:    &mockJiraClient{errSearch: errors.New("jql error")},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := release.NewService(tt.mock, nil)
			issues, err := svc.GetIssues(context.Background(), "PROJ", "v1.0", nil)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(issues) != tt.wantCount {
				t.Errorf("got %d issues, want %d", len(issues), tt.wantCount)
			}
		})
	}
}

// ---- MarkReleased tests ----

func TestService_MarkReleased(t *testing.T) {
	tests := []struct {
		name        string
		mock        *mockJiraClient
		releaseName string
		wantErr     bool
	}{
		{
			name: "marks version as released",
			mock: &mockJiraClient{
				versions: []jiraclient.Version{{ID: "10", Name: "v1.0", Released: false}},
				updated:  &jiraclient.Version{ID: "10", Name: "v1.0", Released: true, ReleaseDate: "2026-05-17"},
			},
			releaseName: "v1.0",
		},
		{
			name: "returns error when version not found",
			mock: &mockJiraClient{
				versions: []jiraclient.Version{{ID: "10", Name: "v1.0"}},
			},
			releaseName: "v99.0",
			wantErr:     true,
		},
		{
			name: "propagates update error",
			mock: &mockJiraClient{
				versions:  []jiraclient.Version{{ID: "10", Name: "v1.0"}},
				errUpdate: errors.New("update failed"),
			},
			releaseName: "v1.0",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := release.NewService(tt.mock, nil)
			got, err := svc.MarkReleased(context.Background(), "PROJ", tt.releaseName, "2026-05-17")

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !got.Released {
				t.Errorf("expected Released=true, got false")
			}
		})
	}
}

// ---- Update tests ----

func TestService_Update(t *testing.T) {
	tests := []struct {
		name      string
		mock      *mockJiraClient
		versionID string
		input     release.UpdateInput
		wantName  string
		wantErr   bool
	}{
		{
			name: "updates description successfully",
			mock: &mockJiraClient{
				updated: &jiraclient.Version{ID: "42", Name: "v1.0", Description: "Updated desc"},
			},
			versionID: "42",
			input:     release.UpdateInput{Name: "v1.0", Description: "Updated desc"},
			wantName:  "v1.0",
		},
		{
			name: "updates release date",
			mock: &mockJiraClient{
				updated: &jiraclient.Version{ID: "42", Name: "v1.0", ReleaseDate: "2026-06-01"},
			},
			versionID: "42",
			input:     release.UpdateInput{Name: "v1.0", ReleaseDate: "2026-06-01"},
			wantName:  "v1.0",
		},
		{
			name: "propagates update error from jira",
			mock: &mockJiraClient{
				errUpdate: errors.New("jira update failed"),
			},
			versionID: "42",
			input:     release.UpdateInput{Name: "v1.0"},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := release.NewService(tt.mock, nil)
			got, err := svc.Update(context.Background(), tt.versionID, tt.input)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", got.Name, tt.wantName)
			}
		})
	}
}

// ---- ValidateForDeploy tests ----

func TestService_ValidateForDeploy(t *testing.T) {
	tests := []struct {
		name      string
		mock      *mockJiraClient
		rules     []string
		wantReady bool
		wantErr   bool
	}{
		{
			name: "all issues done — ready",
			mock: &mockJiraClient{
				issues: []jiraclient.Issue{
					{Key: "P-1", Fields: jiraclient.IssueFields{Summary: "Fix", Status: jiraclient.StatusField{Name: "Done"}, IssueType: jiraclient.IssueTypeField{Name: "Bug"}}},
				},
			},
			rules:     []string{"all_issues_done"},
			wantReady: true,
		},
		{
			name: "open issue — not ready",
			mock: &mockJiraClient{
				issues: []jiraclient.Issue{
					{Key: "P-1", Fields: jiraclient.IssueFields{Summary: "Unresolved", Status: jiraclient.StatusField{Name: "In Progress"}, IssueType: jiraclient.IssueTypeField{Name: "Story"}}},
				},
			},
			rules:     []string{"all_issues_done"},
			wantReady: false,
		},
		{
			name: "propagates search error",
			mock: &mockJiraClient{
				errSearch: errors.New("jql failed"),
			},
			rules:   []string{"all_issues_done"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := release.NewService(tt.mock, nil)
			result, err := svc.ValidateForDeploy(context.Background(), "PROJ", "v1.0", tt.rules)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Ready != tt.wantReady {
				t.Errorf("Ready = %v, want %v, errors: %v", result.Ready, tt.wantReady, result.Errors)
			}
		})
	}
}

// ---- GenerateNotes tests ----

func TestService_GenerateNotes(t *testing.T) {
	tests := []struct {
		name        string
		mock        *mockJiraClient
		wantContain string
		wantErr     bool
	}{
		{
			name: "generates notes with issues",
			mock: &mockJiraClient{
				issues: []jiraclient.Issue{
					{Key: "P-1", Fields: jiraclient.IssueFields{Summary: "Fix login", IssueType: jiraclient.IssueTypeField{Name: "Bug"}}},
					{Key: "P-2", Fields: jiraclient.IssueFields{Summary: "Add dashboard", IssueType: jiraclient.IssueTypeField{Name: "Story"}}},
				},
			},
			wantContain: "# Release Notes: v1.0",
		},
		{
			name: "generates empty notes when no issues",
			mock: &mockJiraClient{
				issues: []jiraclient.Issue{},
			},
			wantContain: "_No issues linked to this release._",
		},
		{
			name: "propagates search error",
			mock: &mockJiraClient{
				errSearch: errors.New("jql failed"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := release.NewService(tt.mock, nil)
			md, err := svc.GenerateNotes(context.Background(), "PROJ", "v1.0")

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !containsStr(md, tt.wantContain) {
				t.Errorf("output %q does not contain %q", md, tt.wantContain)
			}
		})
	}
}

// containsStr is a simple substring check (strings.Contains without import pollution).
func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (substr == "" || findSubstring(s, substr))
}

func findSubstring(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// ---- Archive tests ----

func TestService_Archive(t *testing.T) {
	tests := []struct {
		name        string
		mock        *mockJiraClient
		releaseName string
		wantErr     bool
	}{
		{
			name: "archives a version",
			mock: &mockJiraClient{
				versions: []jiraclient.Version{{ID: "20", Name: "v1.0", Released: true}},
				updated:  &jiraclient.Version{ID: "20", Name: "v1.0", Archived: true},
			},
			releaseName: "v1.0",
		},
		{
			name: "returns error when version not found",
			mock: &mockJiraClient{
				versions: []jiraclient.Version{{ID: "20", Name: "v1.0"}},
			},
			releaseName: "v99.0",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := release.NewService(tt.mock, nil)
			got, err := svc.Archive(context.Background(), "PROJ", tt.releaseName)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !got.Archived {
				t.Errorf("expected Archived=true, got false")
			}
		})
	}
}
