package jira

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// JiraClient defines the operations this application makes against the Jira API.
type JiraClient interface {
	GetProjectVersions(ctx context.Context, projectKey string) ([]Version, error)
	CreateVersion(ctx context.Context, payload CreateVersionInput) (*Version, error)
	UpdateVersion(ctx context.Context, versionID string, payload UpdateVersionInput) (*Version, error)
	SearchIssues(ctx context.Context, jql string, fields []string) ([]Issue, error)
	GetRelatedIssueCounts(ctx context.Context, versionID string) (*RelatedIssueCounts, error)
}

// HTTPClient implements JiraClient using Jira Cloud REST API v3 with Basic Auth.
type HTTPClient struct {
	baseURL string
	auth    string
	http    *http.Client
}

// NewHTTPClient constructs an HTTPClient with the given credentials.
func NewHTTPClient(baseURL, email, token string) JiraClient {
	creds := base64.StdEncoding.EncodeToString([]byte(email + ":" + token))
	return &HTTPClient{
		baseURL: baseURL,
		auth:    "Basic " + creds,
		http:    &http.Client{},
	}
}

// GetProjectVersions fetches all versions for a given project key.
func (c *HTTPClient) GetProjectVersions(ctx context.Context, projectKey string) ([]Version, error) {
	url := fmt.Sprintf("%s/rest/api/3/project/%s/versions", c.baseURL, projectKey)
	var versions []Version
	if err := c.doGet(ctx, url, &versions); err != nil {
		return nil, err
	}
	return versions, nil
}

// CreateVersion creates a new version in Jira.
func (c *HTTPClient) CreateVersion(ctx context.Context, payload CreateVersionInput) (*Version, error) {
	url := fmt.Sprintf("%s/rest/api/3/version", c.baseURL)
	var version Version
	if err := c.doPost(ctx, url, payload, &version); err != nil {
		return nil, err
	}
	return &version, nil
}

// UpdateVersion updates an existing version by ID.
func (c *HTTPClient) UpdateVersion(ctx context.Context, versionID string, payload UpdateVersionInput) (*Version, error) {
	url := fmt.Sprintf("%s/rest/api/3/version/%s", c.baseURL, versionID)
	var version Version
	if err := c.doPut(ctx, url, payload, &version); err != nil {
		return nil, err
	}
	return &version, nil
}

// SearchIssues executes a JQL query and returns matching issues.
func (c *HTTPClient) SearchIssues(ctx context.Context, jql string, fields []string) ([]Issue, error) {
	url := fmt.Sprintf("%s/rest/api/3/search/jql", c.baseURL)
	body := map[string]interface{}{
		"jql":    jql,
		"fields": fields,
	}
	var result SearchResult
	if err := c.doPost(ctx, url, body, &result); err != nil {
		return nil, err
	}
	return result.Issues, nil
}

// GetRelatedIssueCounts fetches issue counts related to a version.
func (c *HTTPClient) GetRelatedIssueCounts(ctx context.Context, versionID string) (*RelatedIssueCounts, error) {
	url := fmt.Sprintf("%s/rest/api/3/version/%s/relatedIssueCounts", c.baseURL, versionID)
	var counts RelatedIssueCounts
	if err := c.doGet(ctx, url, &counts); err != nil {
		return nil, err
	}
	return &counts, nil
}

// doGet performs an authenticated GET request and decodes the JSON response.
func (c *HTTPClient) doGet(ctx context.Context, url string, out interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}
	return c.do(req, out)
}

// doPost performs an authenticated POST request with a JSON body.
func (c *HTTPClient) doPost(ctx context.Context, url string, body interface{}, out interface{}) error {
	return c.doWithBody(ctx, http.MethodPost, url, body, out)
}

// doPut performs an authenticated PUT request with a JSON body.
func (c *HTTPClient) doPut(ctx context.Context, url string, body interface{}, out interface{}) error {
	return c.doWithBody(ctx, http.MethodPut, url, body, out)
}

func (c *HTTPClient) doWithBody(ctx context.Context, method, url string, body interface{}, out interface{}) error {
	encoded, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("encoding request body: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(encoded))
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	return c.do(req, out)
}

// do adds auth headers, executes the request, and decodes the response.
func (c *HTTPClient) do(req *http.Request, out interface{}) error {
	req.Header.Set("Authorization", c.auth)
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("jira API error: status %d, body: %s", resp.StatusCode, string(respBody))
	}

	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}
	}
	return nil
}
