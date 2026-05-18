package release

import (
	"context"
	"fmt"

	"github.com/jinkp/jira-go-mcp/internal/config"
	jiraclient "github.com/jinkp/jira-go-mcp/internal/jira"
	"github.com/jinkp/jira-go-mcp/internal/notes"
	"github.com/jinkp/jira-go-mcp/internal/validation"
)

// CreateInput holds parameters for creating a release.
type CreateInput struct {
	ProjectKey  string
	Name        string
	Description string
	ReleaseDate string
	StartDate   string
}

// UpdateInput holds parameters for updating a release.
type UpdateInput struct {
	Name        string
	Description string
	ReleaseDate string
}

// ReleaseService defines all release management operations.
type ReleaseService interface {
	List(ctx context.Context, projectKey string, includeArchived, includeReleased bool) ([]jiraclient.Version, error)
	Create(ctx context.Context, input CreateInput) (*jiraclient.Version, error)
	GetIssues(ctx context.Context, projectKey, releaseName string, fields []string) ([]jiraclient.Issue, error)
	Update(ctx context.Context, versionID string, input UpdateInput) (*jiraclient.Version, error)
	MarkReleased(ctx context.Context, projectKey, releaseName, releaseDate string) (*jiraclient.Version, error)
	Archive(ctx context.Context, projectKey, releaseName string) (*jiraclient.Version, error)
	ValidateForDeploy(ctx context.Context, projectKey, releaseName string, rules []string) (*validation.Result, error)
	GenerateNotes(ctx context.Context, projectKey, releaseName string) (string, error)
}

// Service implements ReleaseService.
type Service struct {
	jira   jiraclient.JiraClient
	config *config.Config
}

// NewService constructs a Service with the given dependencies.
func NewService(jira jiraclient.JiraClient, cfg *config.Config) ReleaseService {
	return &Service{jira: jira, config: cfg}
}

// List returns versions for a project, with optional filtering.
func (s *Service) List(ctx context.Context, projectKey string, includeArchived, includeReleased bool) ([]jiraclient.Version, error) {
	all, err := s.jira.GetProjectVersions(ctx, projectKey)
	if err != nil {
		return nil, fmt.Errorf("fetching versions: %w", err)
	}

	result := make([]jiraclient.Version, 0, len(all))
	for _, v := range all {
		if v.Archived && !includeArchived {
			continue
		}
		if v.Released && !includeReleased {
			continue
		}
		result = append(result, v)
	}
	return result, nil
}

// Create creates a new version after validating uniqueness.
func (s *Service) Create(ctx context.Context, input CreateInput) (*jiraclient.Version, error) {
	existing, err := s.jira.GetProjectVersions(ctx, input.ProjectKey)
	if err != nil {
		return nil, fmt.Errorf("checking existing versions: %w", err)
	}
	for _, v := range existing {
		if v.Name == input.Name {
			return nil, fmt.Errorf("version %q already exists in project %s", input.Name, input.ProjectKey)
		}
	}

	payload := jiraclient.CreateVersionInput{
		ProjectKey:  input.ProjectKey,
		Name:        input.Name,
		Description: input.Description,
		ReleaseDate: input.ReleaseDate,
		StartDate:   input.StartDate,
	}
	version, err := s.jira.CreateVersion(ctx, payload)
	if err != nil {
		return nil, fmt.Errorf("creating version: %w", err)
	}
	return version, nil
}

// GetIssues returns issues linked to a fixVersion via JQL.
func (s *Service) GetIssues(ctx context.Context, projectKey, releaseName string, fields []string) ([]jiraclient.Issue, error) {
	if len(fields) == 0 {
		fields = []string{"summary", "status", "issuetype", "labels", "priority"}
	}
	jql := fmt.Sprintf(`project = "%s" AND fixVersion = "%s"`, projectKey, releaseName)
	issues, err := s.jira.SearchIssues(ctx, jql, fields)
	if err != nil {
		return nil, fmt.Errorf("searching issues: %w", err)
	}
	return issues, nil
}

// Update modifies name, description, or dates of an existing version.
func (s *Service) Update(ctx context.Context, versionID string, input UpdateInput) (*jiraclient.Version, error) {
	payload := jiraclient.UpdateVersionInput{
		Name:        input.Name,
		Description: input.Description,
		ReleaseDate: input.ReleaseDate,
	}
	version, err := s.jira.UpdateVersion(ctx, versionID, payload)
	if err != nil {
		return nil, fmt.Errorf("updating version: %w", err)
	}
	return version, nil
}

// MarkReleased sets released=true and releaseDate on the named version.
func (s *Service) MarkReleased(ctx context.Context, projectKey, releaseName, releaseDate string) (*jiraclient.Version, error) {
	id, err := s.findVersionID(ctx, projectKey, releaseName)
	if err != nil {
		return nil, err
	}

	released := true
	payload := jiraclient.UpdateVersionInput{
		Released:    &released,
		ReleaseDate: releaseDate,
	}
	version, err := s.jira.UpdateVersion(ctx, id, payload)
	if err != nil {
		return nil, fmt.Errorf("marking version released: %w", err)
	}
	return version, nil
}

// Archive sets archived=true on the named version.
func (s *Service) Archive(ctx context.Context, projectKey, releaseName string) (*jiraclient.Version, error) {
	id, err := s.findVersionID(ctx, projectKey, releaseName)
	if err != nil {
		return nil, err
	}

	archived := true
	payload := jiraclient.UpdateVersionInput{
		Archived: &archived,
	}
	version, err := s.jira.UpdateVersion(ctx, id, payload)
	if err != nil {
		return nil, fmt.Errorf("archiving version: %w", err)
	}
	return version, nil
}

// ValidateForDeploy runs validation rules against the release's issues.
func (s *Service) ValidateForDeploy(ctx context.Context, projectKey, releaseName string, rules []string) (*validation.Result, error) {
	issues, err := s.GetIssues(ctx, projectKey, releaseName, nil)
	if err != nil {
		return nil, err
	}

	cfg := s.config
	var doneStatuses []string
	var criticalLabels []string
	if cfg != nil {
		doneStatuses = cfg.DoneStatuses
		criticalLabels = cfg.CriticalLabels
	}
	if len(doneStatuses) == 0 {
		doneStatuses = []string{"Done"}
	}
	if len(criticalLabels) == 0 {
		criticalLabels = []string{"critical"}
	}

	engine := validation.NewEngine(doneStatuses, criticalLabels)
	return engine.Evaluate(issues, rules), nil
}

// GenerateNotes produces Markdown release notes grouped by issue type.
func (s *Service) GenerateNotes(ctx context.Context, projectKey, releaseName string) (string, error) {
	issues, err := s.GetIssues(ctx, projectKey, releaseName, nil)
	if err != nil {
		return "", err
	}

	return notes.Generate(issues, releaseName), nil
}

// findVersionID looks up a version by name and returns its ID.
func (s *Service) findVersionID(ctx context.Context, projectKey, name string) (string, error) {
	versions, err := s.jira.GetProjectVersions(ctx, projectKey)
	if err != nil {
		return "", fmt.Errorf("fetching versions: %w", err)
	}
	for _, v := range versions {
		if v.Name == name {
			return v.ID, nil
		}
	}
	return "", fmt.Errorf("version %q not found in project %s", name, projectKey)
}
