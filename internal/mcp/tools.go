package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/isaiahrafael/jira-go-mcp/internal/release"
)

// RegisterTools registers all 8 Jira release MCP tools on the given server.
// All business logic is delegated to the ReleaseService — handlers are thin wrappers.
func RegisterTools(s *server.MCPServer, svc release.ReleaseService) {
	registerListByProject(s, svc)
	registerCreate(s, svc)
	registerGetIssues(s, svc)
	registerUpdate(s, svc)
	registerMarkReleased(s, svc)
	registerArchive(s, svc)
	registerValidateForDeploy(s, svc)
	registerGenerateNotes(s, svc)
}

func registerListByProject(s *server.MCPServer, svc release.ReleaseService) {
	tool := mcp.NewTool("jira_release_list_by_project",
		mcp.WithDescription("List releases (versions) for a Jira project"),
		mcp.WithString("project_key", mcp.Required(), mcp.Description("Jira project key, e.g. PROJ")),
		mcp.WithBoolean("include_archived", mcp.Description("Include archived releases (default false)")),
		mcp.WithBoolean("include_released", mcp.Description("Include already-released versions (default false)")),
	)
	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		projectKey := mcp.ParseString(req, "project_key", "")
		includeArchived := mcp.ParseBoolean(req, "include_archived", false)
		includeReleased := mcp.ParseBoolean(req, "include_released", false)

		versions, err := svc.List(ctx, projectKey, includeArchived, includeReleased)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("listing releases: %v", err)), nil
		}
		return toolJSON(versions), nil
	})
}

func registerCreate(s *server.MCPServer, svc release.ReleaseService) {
	tool := mcp.NewTool("jira_release_create",
		mcp.WithDescription("Create a new release version in a Jira project"),
		mcp.WithString("project_key", mcp.Required(), mcp.Description("Jira project key")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Release name, e.g. v1.0.0")),
		mcp.WithString("description", mcp.Description("Optional release description")),
		mcp.WithString("release_date", mcp.Description("Optional release date (YYYY-MM-DD)")),
		mcp.WithString("start_date", mcp.Description("Optional start date (YYYY-MM-DD)")),
	)
	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		input := release.CreateInput{
			ProjectKey:  mcp.ParseString(req, "project_key", ""),
			Name:        mcp.ParseString(req, "name", ""),
			Description: mcp.ParseString(req, "description", ""),
			ReleaseDate: mcp.ParseString(req, "release_date", ""),
			StartDate:   mcp.ParseString(req, "start_date", ""),
		}
		version, err := svc.Create(ctx, input)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("creating release: %v", err)), nil
		}
		return toolJSON(version), nil
	})
}

func registerGetIssues(s *server.MCPServer, svc release.ReleaseService) {
	tool := mcp.NewTool("jira_release_get_issues",
		mcp.WithDescription("Get issues linked to a Jira release via fixVersion"),
		mcp.WithString("project_key", mcp.Required(), mcp.Description("Jira project key")),
		mcp.WithString("release_name", mcp.Required(), mcp.Description("Release name")),
	)
	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		projectKey := mcp.ParseString(req, "project_key", "")
		releaseName := mcp.ParseString(req, "release_name", "")

		issues, err := svc.GetIssues(ctx, projectKey, releaseName, nil)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("getting issues: %v", err)), nil
		}
		return toolJSON(issues), nil
	})
}

func registerUpdate(s *server.MCPServer, svc release.ReleaseService) {
	tool := mcp.NewTool("jira_release_update",
		mcp.WithDescription("Update an existing Jira release version"),
		mcp.WithString("version_id", mcp.Required(), mcp.Description("Version ID to update")),
		mcp.WithString("name", mcp.Description("New name")),
		mcp.WithString("description", mcp.Description("New description")),
		mcp.WithString("release_date", mcp.Description("New release date (YYYY-MM-DD)")),
	)
	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		versionID := mcp.ParseString(req, "version_id", "")
		input := release.UpdateInput{
			Name:        mcp.ParseString(req, "name", ""),
			Description: mcp.ParseString(req, "description", ""),
			ReleaseDate: mcp.ParseString(req, "release_date", ""),
		}
		version, err := svc.Update(ctx, versionID, input)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("updating release: %v", err)), nil
		}
		return toolJSON(version), nil
	})
}

func registerMarkReleased(s *server.MCPServer, svc release.ReleaseService) {
	tool := mcp.NewTool("jira_release_mark_released",
		mcp.WithDescription("Mark a Jira release version as released with today's date"),
		mcp.WithString("project_key", mcp.Required(), mcp.Description("Jira project key")),
		mcp.WithString("release_name", mcp.Required(), mcp.Description("Release name")),
		mcp.WithString("release_date", mcp.Description("Release date override (YYYY-MM-DD); defaults to today")),
	)
	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		projectKey := mcp.ParseString(req, "project_key", "")
		releaseName := mcp.ParseString(req, "release_name", "")
		releaseDate := mcp.ParseString(req, "release_date", "")
		if releaseDate == "" {
			releaseDate = time.Now().Format("2006-01-02")
		}
		version, err := svc.MarkReleased(ctx, projectKey, releaseName, releaseDate)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("marking released: %v", err)), nil
		}
		return toolJSON(version), nil
	})
}

func registerArchive(s *server.MCPServer, svc release.ReleaseService) {
	tool := mcp.NewTool("jira_release_archive",
		mcp.WithDescription("Archive a Jira release version"),
		mcp.WithString("project_key", mcp.Required(), mcp.Description("Jira project key")),
		mcp.WithString("release_name", mcp.Required(), mcp.Description("Release name to archive")),
	)
	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		projectKey := mcp.ParseString(req, "project_key", "")
		releaseName := mcp.ParseString(req, "release_name", "")

		version, err := svc.Archive(ctx, projectKey, releaseName)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("archiving release: %v", err)), nil
		}
		return toolJSON(version), nil
	})
}

func registerValidateForDeploy(s *server.MCPServer, svc release.ReleaseService) {
	tool := mcp.NewTool("jira_release_validate_for_deploy",
		mcp.WithDescription("Validate a Jira release is ready to deploy by running rules against its issues"),
		mcp.WithString("project_key", mcp.Required(), mcp.Description("Jira project key")),
		mcp.WithString("release_name", mcp.Required(), mcp.Description("Release name")),
		mcp.WithString("rules", mcp.Description("Comma-separated list of rules; defaults to all built-in rules")),
	)
	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		projectKey := mcp.ParseString(req, "project_key", "")
		releaseName := mcp.ParseString(req, "release_name", "")
		rulesStr := mcp.ParseString(req, "rules", "")

		var rules []string
		if rulesStr != "" {
			for _, r := range strings.Split(rulesStr, ",") {
				if trimmed := strings.TrimSpace(r); trimmed != "" {
					rules = append(rules, trimmed)
				}
			}
		} else {
			rules = []string{"all_issues_done", "no_critical_bugs_open", "no_blocking_issues", "min_issues_count"}
		}

		result, err := svc.ValidateForDeploy(ctx, projectKey, releaseName, rules)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("validating release: %v", err)), nil
		}
		return toolJSON(result), nil
	})
}

func registerGenerateNotes(s *server.MCPServer, svc release.ReleaseService) {
	tool := mcp.NewTool("jira_release_generate_notes",
		mcp.WithDescription("Generate Markdown release notes for a Jira release grouped by issue type"),
		mcp.WithString("project_key", mcp.Required(), mcp.Description("Jira project key")),
		mcp.WithString("release_name", mcp.Required(), mcp.Description("Release name")),
	)
	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		projectKey := mcp.ParseString(req, "project_key", "")
		releaseName := mcp.ParseString(req, "release_name", "")

		md, err := svc.GenerateNotes(ctx, projectKey, releaseName)
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("generating notes: %v", err)), nil
		}
		return mcp.NewToolResultText(md), nil
	})
}

// toolJSON serializes v as JSON and returns it as a tool result text.
func toolJSON(v interface{}) *mcp.CallToolResult {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("serializing response: %v", err))
	}
	return mcp.NewToolResultText(string(b))
}
