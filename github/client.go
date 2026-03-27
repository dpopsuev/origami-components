// Package github provides a minimal GitHub REST API v3 client for PR lookup.
// Used by the dataset pipeline to verify ground truth fix_prs references.
package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const defaultTimeout = 15 * time.Second

// Client is a GitHub REST API v3 client.
type Client struct {
	token string
	http  *http.Client
}

// Option configures a Client.
type Option func(*Client)

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(c *http.Client) Option {
	return func(cl *Client) { cl.http = c }
}

// New creates a GitHub client. Token is a PAT (optional for public repos).
func New(token string, opts ...Option) *Client {
	c := &Client{
		token: token,
		http:  &http.Client{Timeout: defaultTimeout},
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// PullRequest is the subset of GitHub PR fields we need for verification.
type PullRequest struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	State  string `json:"state"` // "open", "closed"
	Merged bool   `json:"merged"`
	HTMLURL string `json:"html_url"`
}

// ParsePRRef parses "org/repo#123" into (owner, repo, number).
func ParsePRRef(ref string) (owner, repo string, number int, err error) {
	// Split on "#" first.
	parts := strings.SplitN(ref, "#", 2)
	if len(parts) != 2 {
		return "", "", 0, fmt.Errorf("invalid PR ref %q: expected org/repo#number", ref)
	}

	number, err = strconv.Atoi(parts[1])
	if err != nil {
		return "", "", 0, fmt.Errorf("invalid PR number in %q: %w", ref, err)
	}

	// Split owner/repo.
	repoParts := strings.SplitN(parts[0], "/", 2)
	if len(repoParts) != 2 {
		return "", "", 0, fmt.Errorf("invalid PR ref %q: expected org/repo#number", ref)
	}

	return repoParts[0], repoParts[1], number, nil
}

// GetPR fetches a pull request by owner, repo, and number.
func (c *Client) GetPR(ctx context.Context, owner, repo string, number int) (*PullRequest, error) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls/%d", owner, repo, number)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("github: create request: %w", err)
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("github: request %s/%s#%d: %w", owner, repo, number, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github: %s/%s#%d returned %d", owner, repo, number, resp.StatusCode)
	}

	var pr PullRequest
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		return nil, fmt.Errorf("github: decode %s/%s#%d: %w", owner, repo, number, err)
	}
	return &pr, nil
}
