package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// ListPRs returns pull requests for a repo, optionally filtered by state and base branch.
func (c *Client) ListPRs(ctx context.Context, owner, repo string, state, base string, perPage int) ([]PullRequest, error) {
	if perPage <= 0 {
		perPage = 30
	}

	params := url.Values{
		"per_page": {fmt.Sprintf("%d", perPage)},
		"sort":     {"updated"},
		"direction": {"desc"},
	}
	if state != "" {
		params.Set("state", state) // "open", "closed", "all"
	}
	if base != "" {
		params.Set("base", base) // e.g. "release-4.18"
	}

	reqURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls?%s", owner, repo, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("github list PRs: create request: %w", err)
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("github list PRs %s/%s: %w", owner, repo, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github list PRs %s/%s: status %d", owner, repo, resp.StatusCode)
	}

	var prs []PullRequest
	if err := json.NewDecoder(resp.Body).Decode(&prs); err != nil {
		return nil, fmt.Errorf("github list PRs: decode: %w", err)
	}
	return prs, nil
}
