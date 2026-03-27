// Package jira provides a minimal Jira REST API v2 client for issue lookup.
// Used by the dataset pipeline to verify ground truth RCA jira_id references.
package jira

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const defaultTimeout = 15 * time.Second

// Client is a Jira REST API v2 client.
type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

// Option configures a Client.
type Option func(*Client)

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(c *http.Client) Option {
	return func(cl *Client) { cl.http = c }
}

// New creates a Jira client. Token is a PAT or API token.
// baseURL is the Jira instance (e.g. "https://issues.redhat.com").
func New(baseURL, token string, opts ...Option) (*Client, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("jira: baseURL is required")
	}
	if !strings.HasPrefix(baseURL, "https://") &&
		!strings.HasPrefix(baseURL, "http://localhost") &&
		!strings.HasPrefix(baseURL, "http://127.0.0.1") {
		return nil, fmt.Errorf("jira: baseURL must use HTTPS (got %q)", baseURL)
	}
	c := &Client{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		token:   token,
		http:    &http.Client{Timeout: defaultTimeout},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c, nil
}

// Issue is the subset of Jira issue fields we need for verification.
type Issue struct {
	Key    string `json:"key"`
	Fields struct {
		Summary    string      `json:"summary"`
		Status     IssueStatus `json:"status"`
		Resolution *struct {
			Name string `json:"name"`
		} `json:"resolution"`
		FixVersions []struct {
			Name string `json:"name"`
		} `json:"fixVersions"`
	} `json:"fields"`
}

// IssueStatus is the Jira status object.
type IssueStatus struct {
	Name           string `json:"name"`
	StatusCategory struct {
		Key string `json:"key"` // "done", "indeterminate", "new"
	} `json:"statusCategory"`
}

// StatusName returns the issue's current status name.
func (i *Issue) StatusName() string {
	return i.Fields.Status.Name
}

// IsDone returns true if the issue's status category is "done".
func (i *Issue) IsDone() bool {
	return i.Fields.Status.StatusCategory.Key == "done"
}

// GetIssue fetches a single issue by key (e.g. "OCPBUGS-70233").
func (c *Client) GetIssue(ctx context.Context, key string) (*Issue, error) {
	url := fmt.Sprintf("%s/rest/api/2/issue/%s?fields=summary,status,resolution,fixVersions", c.baseURL, key)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("jira: create request: %w", err)
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("jira: request %s: %w", key, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("jira: %s returned %d", key, resp.StatusCode)
	}

	var issue Issue
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, fmt.Errorf("jira: decode %s: %w", key, err)
	}
	return &issue, nil
}
