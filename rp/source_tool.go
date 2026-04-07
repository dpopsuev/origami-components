package rp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/dpopsuev/battery/tool"
)

// sourceReadSchema is the JSON Schema for the source_read tool input.
var sourceReadSchema = json.RawMessage(`{
	"type": "object",
	"properties": {
		"launch_id": {"type": "string", "description": "ReportPortal launch ID"}
	},
	"required": ["launch_id"]
}`)

// SourceTool returns a battery.Tool that reads test failure data from ReportPortal.
// Config comes from environment variables (12-factor).
func SourceTool() tool.Tool {
	return &sourceReadTool{
		baseURL:    os.Getenv("RP_BASE_URL"),
		apiKeyPath: os.Getenv("RP_API_KEY_PATH"),
		project:    os.Getenv("RP_PROJECT"),
	}
}

type sourceReadTool struct {
	baseURL    string
	apiKeyPath string
	project    string
}

func (t *sourceReadTool) Name() string             { return "source_read" }
func (t *sourceReadTool) Description() string      { return "Fetch test failure data from ReportPortal" }
func (t *sourceReadTool) InputSchema() json.RawMessage { return sourceReadSchema }

func (t *sourceReadTool) Execute(ctx context.Context, input json.RawMessage) (string, error) {
	var params struct {
		LaunchID string `json:"launch_id"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return "", fmt.Errorf("source_read: parse input: %w", err)
	}
	if params.LaunchID == "" {
		return "", fmt.Errorf("source_read: launch_id is required")
	}

	reader, err := NewSourceReader(t.baseURL, t.apiKeyPath, t.project)
	if err != nil {
		return "", fmt.Errorf("source_read: create reader: %w", err)
	}

	env, err := reader.FetchEnvelope(params.LaunchID)
	if err != nil {
		return "", fmt.Errorf("source_read: fetch: %w", err)
	}

	data, err := json.Marshal(env)
	if err != nil {
		return "", fmt.Errorf("source_read: marshal: %w", err)
	}
	return string(data), nil
}

// Compile-time check.
var _ tool.Tool = (*sourceReadTool)(nil)
