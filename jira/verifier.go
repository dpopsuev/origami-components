package jira

import (
	"context"
	"fmt"

	"github.com/dpopsuev/origami/ingest"
)

const verifierName = "jira"

// Verifier checks that a record's jira_id references a resolved Jira ticket.
type Verifier struct {
	Client *Client
	Field  string // record field containing the Jira key (default: "jira_id")
}

// Name returns the verifier name.
func (v *Verifier) Name() string { return verifierName }

// Verify fetches the Jira issue and checks if its status category is "done".
func (v *Verifier) Verify(ctx context.Context, record ingest.Record) (ingest.VerifyResult, error) {
	field := v.Field
	if field == "" {
		field = "jira_id"
	}

	jiraKey, _ := record.Fields[field].(string)
	if jiraKey == "" {
		return ingest.VerifyResult{Verified: false, Reason: fmt.Sprintf("field %q empty", field)}, nil
	}

	issue, err := v.Client.GetIssue(ctx, jiraKey)
	if err != nil {
		return ingest.VerifyResult{}, fmt.Errorf("jira verifier: %w", err)
	}

	if issue.IsDone() {
		return ingest.VerifyResult{
			Verified: true,
			Reason:   fmt.Sprintf("%s status=%s (done)", jiraKey, issue.StatusName()),
		}, nil
	}

	return ingest.VerifyResult{
		Verified: false,
		Reason:   fmt.Sprintf("%s status=%s (not done)", jiraKey, issue.StatusName()),
	}, nil
}

// Compile-time check.
var _ ingest.Verifier = (*Verifier)(nil)
