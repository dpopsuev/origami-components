package github

import (
	"context"
	"fmt"

	"github.com/dpopsuev/origami/ingest"
)

const verifierName = "github"

// Verifier checks that a record's fix_prs references are merged.
type Verifier struct {
	Client *Client
	Field  string // record field containing PR refs (default: "fix_prs")
}

// Name returns the verifier name.
func (v *Verifier) Name() string { return verifierName }

// Verify checks each PR ref in the record's fix_prs field.
// All PRs must be merged for verification to pass.
func (v *Verifier) Verify(ctx context.Context, record ingest.Record) (ingest.VerifyResult, error) {
	field := v.Field
	if field == "" {
		field = "fix_prs"
	}

	refs, ok := record.Fields[field].([]any)
	if !ok || len(refs) == 0 {
		// No fix PRs — can't verify via GitHub. Not a failure, just inconclusive.
		return ingest.VerifyResult{Verified: false, Reason: fmt.Sprintf("field %q empty or missing", field)}, nil
	}

	for _, ref := range refs {
		refStr, ok := ref.(string)
		if !ok {
			continue
		}

		owner, repo, number, parseErr := ParsePRRef(refStr)
		if parseErr != nil {
			return ingest.VerifyResult{Verified: false, Reason: parseErr.Error()}, nil
		}

		pr, err := v.Client.GetPR(ctx, owner, repo, number)
		if err != nil {
			return ingest.VerifyResult{}, fmt.Errorf("github verifier: %w", err)
		}

		if !pr.Merged {
			return ingest.VerifyResult{
				Verified: false,
				Reason:   fmt.Sprintf("%s not merged (state=%s)", refStr, pr.State),
			}, nil
		}
	}

	return ingest.VerifyResult{
		Verified: true,
		Reason:   fmt.Sprintf("all %d fix PRs merged", len(refs)),
	}, nil
}

// Compile-time check.
var _ ingest.Verifier = (*Verifier)(nil)
