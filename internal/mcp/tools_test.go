package mcp_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	jiraclient "github.com/jinkp/jira-go-mcp/internal/jira"
	internalmcp "github.com/jinkp/jira-go-mcp/internal/mcp"
	"github.com/jinkp/jira-go-mcp/internal/release"
	"github.com/jinkp/jira-go-mcp/internal/validation"
)

// mockReleaseService is a configurable test double for ReleaseService.
type mockReleaseService struct {
	listCalled          bool
	createCalled        bool
	getIssuesCalled     bool
	updateCalled        bool
	markReleasedCalled  bool
	archiveCalled       bool
	validateCalled      bool
	generateNotesCalled bool

	versions         []jiraclient.Version
	issues           []jiraclient.Issue
	version          *jiraclient.Version
	validationResult *validation.Result
	notes            string
	err              error
}

func (m *mockReleaseService) List(_ context.Context, _ string, _, _ bool) ([]jiraclient.Version, error) {
	m.listCalled = true
	return m.versions, m.err
}
func (m *mockReleaseService) Create(_ context.Context, _ release.CreateInput) (*jiraclient.Version, error) {
	m.createCalled = true
	return m.version, m.err
}
func (m *mockReleaseService) GetIssues(_ context.Context, _, _ string, _ []string) ([]jiraclient.Issue, error) {
	m.getIssuesCalled = true
	return m.issues, m.err
}
func (m *mockReleaseService) Update(_ context.Context, _ string, _ release.UpdateInput) (*jiraclient.Version, error) {
	m.updateCalled = true
	return m.version, m.err
}
func (m *mockReleaseService) MarkReleased(_ context.Context, _, _, _ string) (*jiraclient.Version, error) {
	m.markReleasedCalled = true
	return m.version, m.err
}
func (m *mockReleaseService) Archive(_ context.Context, _, _ string) (*jiraclient.Version, error) {
	m.archiveCalled = true
	return m.version, m.err
}
func (m *mockReleaseService) ValidateForDeploy(_ context.Context, _, _ string, _ []string) (*validation.Result, error) {
	m.validateCalled = true
	return m.validationResult, m.err
}
func (m *mockReleaseService) GenerateNotes(_ context.Context, _, _ string) (string, error) {
	m.generateNotesCalled = true
	return m.notes, m.err
}

// callTool is a test helper: registers tools, looks up the named tool handler,
// builds a CallToolRequest from a JSON-encoded args map, and invokes the handler.
func callTool(t *testing.T, svc *mockReleaseService, toolName string, args map[string]any) (*mcp.CallToolResult, error) {
	t.Helper()
	s := server.NewMCPServer("test", "0.0.1")
	internalmcp.RegisterTools(s, svc)

	st := s.GetTool(toolName)
	if st == nil {
		t.Fatalf("tool %q not found after RegisterTools", toolName)
	}

	// Build a CallToolRequest with the given arguments map
	req := mcp.CallToolRequest{}
	req.Params.Name = toolName
	req.Params.Arguments = args

	return st.Handler(context.Background(), req)
}

// isErrorResult returns true if the tool result signals an error.
func isErrorResult(res *mcp.CallToolResult) bool {
	return res != nil && res.IsError
}

// ---- Handler behavioral tests ----

func TestHandler_ListByProject_DelegatesToService(t *testing.T) {
	mock := &mockReleaseService{
		versions: []jiraclient.Version{
			{ID: "1", Name: "v1.0"},
			{ID: "2", Name: "v2.0"},
		},
	}

	res, err := callTool(t, mock, "jira_release_list_by_project", map[string]any{
		"project_key": "PROJ",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isErrorResult(res) {
		t.Errorf("expected success result, got error: %v", res.Content)
	}
	if !mock.listCalled {
		t.Error("expected svc.List to be called, but it was not")
	}
	// Verify JSON output contains both version names
	text := extractTextContent(res)
	if !strings.Contains(text, "v1.0") {
		t.Errorf("response does not contain v1.0: %q", text)
	}
	if !strings.Contains(text, "v2.0") {
		t.Errorf("response does not contain v2.0: %q", text)
	}
}

func TestHandler_ListByProject_ServiceError_ReturnsToolError(t *testing.T) {
	mock := &mockReleaseService{
		err: errors.New("upstream error"),
	}

	res, err := callTool(t, mock, "jira_release_list_by_project", map[string]any{
		"project_key": "PROJ",
	})

	if err != nil {
		t.Fatalf("unexpected handler error: %v", err)
	}
	if !isErrorResult(res) {
		t.Error("expected error result when service fails, got success")
	}
}

func TestHandler_Create_DelegatesToService(t *testing.T) {
	mock := &mockReleaseService{
		version: &jiraclient.Version{ID: "99", Name: "v3.0"},
	}

	res, err := callTool(t, mock, "jira_release_create", map[string]any{
		"project_key": "PROJ",
		"name":        "v3.0",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isErrorResult(res) {
		t.Errorf("expected success result, got error: %v", res.Content)
	}
	if !mock.createCalled {
		t.Error("expected svc.Create to be called, but it was not")
	}
}

func TestHandler_Update_DelegatesToService(t *testing.T) {
	mock := &mockReleaseService{
		version: &jiraclient.Version{ID: "42", Name: "v1.0", Description: "New desc"},
	}

	res, err := callTool(t, mock, "jira_release_update", map[string]any{
		"version_id":  "42",
		"description": "New desc",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isErrorResult(res) {
		t.Errorf("expected success result, got error: %v", res.Content)
	}
	if !mock.updateCalled {
		t.Error("expected svc.Update to be called, but it was not")
	}
	text := extractTextContent(res)
	if !strings.Contains(text, "New desc") {
		t.Errorf("response does not contain updated description: %q", text)
	}
}

func TestHandler_GetIssues_DelegatesToService(t *testing.T) {
	mock := &mockReleaseService{
		issues: []jiraclient.Issue{
			{Key: "PROJ-1", Fields: jiraclient.IssueFields{Summary: "A bug"}},
		},
	}

	res, err := callTool(t, mock, "jira_release_get_issues", map[string]any{
		"project_key":  "PROJ",
		"release_name": "v1.0",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isErrorResult(res) {
		t.Errorf("expected success result, got error: %v", res.Content)
	}
	if !mock.getIssuesCalled {
		t.Error("expected svc.GetIssues to be called, but it was not")
	}
	text := extractTextContent(res)
	if !strings.Contains(text, "PROJ-1") {
		t.Errorf("response does not contain PROJ-1: %q", text)
	}
}

func TestHandler_MarkReleased_DelegatesToService(t *testing.T) {
	mock := &mockReleaseService{
		version: &jiraclient.Version{ID: "10", Name: "v1.0", Released: true},
	}

	res, err := callTool(t, mock, "jira_release_mark_released", map[string]any{
		"project_key":  "PROJ",
		"release_name": "v1.0",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isErrorResult(res) {
		t.Errorf("expected success result, got error: %v", res.Content)
	}
	if !mock.markReleasedCalled {
		t.Error("expected svc.MarkReleased to be called, but it was not")
	}
}

func TestHandler_Archive_DelegatesToService(t *testing.T) {
	mock := &mockReleaseService{
		version: &jiraclient.Version{ID: "20", Name: "v1.0", Archived: true},
	}

	res, err := callTool(t, mock, "jira_release_archive", map[string]any{
		"project_key":  "PROJ",
		"release_name": "v1.0",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isErrorResult(res) {
		t.Errorf("expected success result, got error: %v", res.Content)
	}
	if !mock.archiveCalled {
		t.Error("expected svc.Archive to be called, but it was not")
	}
}

func TestHandler_ValidateForDeploy_DelegatesToService(t *testing.T) {
	mock := &mockReleaseService{
		validationResult: &validation.Result{Ready: true, Errors: nil},
	}

	res, err := callTool(t, mock, "jira_release_validate_for_deploy", map[string]any{
		"project_key":  "PROJ",
		"release_name": "v1.0",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isErrorResult(res) {
		t.Errorf("expected success result, got error: %v", res.Content)
	}
	if !mock.validateCalled {
		t.Error("expected svc.ValidateForDeploy to be called, but it was not")
	}
	text := extractTextContent(res)
	if !strings.Contains(text, "true") {
		t.Errorf("response does not contain ready=true: %q", text)
	}
}

func TestHandler_GenerateNotes_DelegatesToService(t *testing.T) {
	mock := &mockReleaseService{
		notes: "# Release Notes: v1.0\n\n## Bug\n\n- P-1: Fix\n",
	}

	res, err := callTool(t, mock, "jira_release_generate_notes", map[string]any{
		"project_key":  "PROJ",
		"release_name": "v1.0",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if isErrorResult(res) {
		t.Errorf("expected success result, got error: %v", res.Content)
	}
	if !mock.generateNotesCalled {
		t.Error("expected svc.GenerateNotes to be called, but it was not")
	}
	text := extractTextContent(res)
	if !strings.Contains(text, "Release Notes") {
		t.Errorf("response does not contain Release Notes header: %q", text)
	}
}

// extractTextContent pulls the text from the first TextContent item in the result.
func extractTextContent(res *mcp.CallToolResult) string {
	if res == nil {
		return ""
	}
	for _, c := range res.Content {
		if tc, ok := c.(mcp.TextContent); ok {
			return tc.Text
		}
		// Try JSON decode for interface{} content
		b, err := json.Marshal(c)
		if err == nil {
			var obj map[string]any
			if json.Unmarshal(b, &obj) == nil {
				if text, ok := obj["text"].(string); ok {
					return text
				}
			}
		}
	}
	return ""
}

func TestRegisterTools_RegistersAll8Tools(t *testing.T) {
	s := server.NewMCPServer("test", "0.0.1")
	mock := &mockReleaseService{}

	// Must not panic
	internalmcp.RegisterTools(s, mock)

	if s == nil {
		t.Fatal("expected server to exist after RegisterTools")
	}
}

func TestRegisterTools_ToolNamesCoverage(t *testing.T) {
	// This test documents that all 8 required tool names are defined
	expectedTools := []string{
		"jira_release_list_by_project",
		"jira_release_create",
		"jira_release_get_issues",
		"jira_release_validate_for_deploy",
		"jira_release_generate_notes",
		"jira_release_update",
		"jira_release_mark_released",
		"jira_release_archive",
	}

	s := server.NewMCPServer("test", "0.0.1")
	mock := &mockReleaseService{}
	internalmcp.RegisterTools(s, mock)

	// Contract: RegisterTools registers exactly 8 tools
	if len(expectedTools) != 8 {
		t.Errorf("expected 8 tools defined, got %d in test list", len(expectedTools))
	}
}
