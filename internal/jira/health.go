package jira

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ConnectionInfo contains the identity returned by Jira when auth succeeds.
type ConnectionInfo struct {
	DisplayName string
	AccountID   string
	Email       string
}

type myselfResponse struct {
	DisplayName  string `json:"displayName"`
	AccountID    string `json:"accountId"`
	EmailAddress string `json:"emailAddress"`
}

// ValidateConnection checks whether the provided Jira settings can authenticate
// successfully against Jira Cloud.
func ValidateConnection(ctx context.Context, baseURL, email, token string) (*ConnectionInfo, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		return nil, fmt.Errorf("jira base URL is required")
	}
	if strings.TrimSpace(email) == "" {
		return nil, fmt.Errorf("jira email is required")
	}
	if strings.TrimSpace(token) == "" {
		return nil, fmt.Errorf("jira API token is required")
	}

	url := baseURL + "/rest/api/3/myself"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("building request: %w", err)
	}
	creds := base64.StdEncoding.EncodeToString([]byte(email + ":" + token))
	req.Header.Set("Authorization", "Basic "+creds)
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("jira connection failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	var me myselfResponse
	if err := json.Unmarshal(body, &me); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &ConnectionInfo{
		DisplayName: me.DisplayName,
		AccountID:   me.AccountID,
		Email:       me.EmailAddress,
	}, nil
}
