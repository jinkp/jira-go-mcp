package jira_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	jiraclient "github.com/isaiahrafael/jira-go-mcp/internal/jira"
)

// newTestClient builds an HTTPClient pointed at the given test server base URL.
func newTestClient(baseURL, email, token string) jiraclient.JiraClient {
	return jiraclient.NewHTTPClient(baseURL, email, token)
}

func TestHTTPClient_GetProjectVersions(t *testing.T) {
	tests := []struct {
		name       string
		serverResp func(w http.ResponseWriter, r *http.Request)
		wantCount  int
		wantErr    bool
		errContains string
	}{
		{
			name: "returns versions on success",
			serverResp: func(w http.ResponseWriter, r *http.Request) {
				// Verify auth header present
				auth := r.Header.Get("Authorization")
				if !strings.HasPrefix(auth, "Basic ") {
					http.Error(w, "missing auth", http.StatusUnauthorized)
					return
				}
				versions := []jiraclient.Version{
					{ID: "10001", Name: "v1.0", Released: false, Archived: false},
					{ID: "10002", Name: "v1.1", Released: true, Archived: false},
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(versions)
			},
			wantCount: 2,
			wantErr:   false,
		},
		{
			name: "returns error on non-2xx",
			serverResp: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Forbidden", http.StatusForbidden)
			},
			wantErr:     true,
			errContains: "403",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(tt.serverResp))
			defer srv.Close()

			client := newTestClient(srv.URL, "test@test.com", "token")
			versions, err := client.GetProjectVersions(context.Background(), "PROJ")

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error = %q, want to contain %q", err.Error(), tt.errContains)
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

func TestHTTPClient_CreateVersion(t *testing.T) {
	tests := []struct {
		name        string
		input       jiraclient.CreateVersionInput
		serverResp  func(w http.ResponseWriter, r *http.Request)
		wantVersion *jiraclient.Version
		wantErr     bool
	}{
		{
			name:  "creates version successfully",
			input: jiraclient.CreateVersionInput{ProjectKey: "PROJ", Name: "v2.0"},
			serverResp: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					http.Error(w, "wrong method", http.StatusMethodNotAllowed)
					return
				}
				v := jiraclient.Version{ID: "10003", Name: "v2.0"}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				_ = json.NewEncoder(w).Encode(v)
			},
			wantVersion: &jiraclient.Version{ID: "10003", Name: "v2.0"},
		},
		{
			name:  "returns error on server failure",
			input: jiraclient.CreateVersionInput{ProjectKey: "PROJ", Name: "v2.0"},
			serverResp: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(tt.serverResp))
			defer srv.Close()

			client := newTestClient(srv.URL, "test@test.com", "token")
			got, err := client.CreateVersion(context.Background(), tt.input)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.ID != tt.wantVersion.ID {
				t.Errorf("ID = %q, want %q", got.ID, tt.wantVersion.ID)
			}
			if got.Name != tt.wantVersion.Name {
				t.Errorf("Name = %q, want %q", got.Name, tt.wantVersion.Name)
			}
		})
	}
}

func TestHTTPClient_SearchIssues(t *testing.T) {
	tests := []struct {
		name       string
		jql        string
		serverResp func(w http.ResponseWriter, r *http.Request)
		wantCount  int
		wantErr    bool
	}{
		{
			name: "returns issues matching JQL",
			jql:  `fixVersion="v1.0"`,
			serverResp: func(w http.ResponseWriter, r *http.Request) {
				result := jiraclient.SearchResult{
					Issues: []jiraclient.Issue{
						{Key: "PROJ-1", Fields: jiraclient.IssueFields{Summary: "Bug fix", Status: jiraclient.StatusField{Name: "Done"}, IssueType: jiraclient.IssueTypeField{Name: "Bug"}}},
						{Key: "PROJ-2", Fields: jiraclient.IssueFields{Summary: "Feature", Status: jiraclient.StatusField{Name: "Done"}, IssueType: jiraclient.IssueTypeField{Name: "Story"}}},
					},
				}
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(result)
			},
			wantCount: 2,
		},
		{
			name: "returns error on non-2xx",
			jql:  `fixVersion="bad"`,
			serverResp: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "Bad Request", http.StatusBadRequest)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(tt.serverResp))
			defer srv.Close()

			client := newTestClient(srv.URL, "test@test.com", "token")
			issues, err := client.SearchIssues(context.Background(), tt.jql, []string{"summary", "status", "issuetype"})

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
