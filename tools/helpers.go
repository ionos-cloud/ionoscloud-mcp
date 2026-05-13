package tools

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/ionos-cloud/sdk-go-bundle/shared"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// ValidateDate checks that date is a valid YYYY-MM-DD string (e.g. "2026-04-15").
// Leading/trailing whitespace is trimmed before parsing.
func ValidateDate(date string) error {
	if _, err := time.Parse("2006-01-02", strings.TrimSpace(date)); err != nil {
		return fmt.Errorf("invalid date %q: must be YYYY-MM-DD format (e.g. 2026-04-15)", date)
	}
	return nil
}

// ValidatePeriod checks that period is a valid YYYY-MM string (e.g. "2026-04").
// Leading/trailing whitespace is trimmed before parsing.
func ValidatePeriod(period string) error {
	return ValidateDate(strings.TrimSpace(period) + "-01")
}

// ToResult marshals an API response into an MCP text result.
func ToResult(data any, apiErr error) (*mcp.CallToolResult, any, error) {
	if apiErr != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: apiErr.Error()},
			},
			IsError: true,
		}, nil, nil
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal response: %w", err)
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(bytes)},
		},
	}, nil, nil
}

// TextResult wraps a plain string as an MCP text result.
func TextResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}

// ToRawResult returns the raw response payload as an MCP text result.
// Use this for API endpoints that return non-JSON content (e.g. zone files).
func ToRawResult(resp *shared.APIResponse, apiErr error) (*mcp.CallToolResult, any, error) {
	if apiErr != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: apiErr.Error()},
			},
			IsError: true,
		}, nil, nil
	}

	if resp == nil {
		return nil, nil, fmt.Errorf("empty response from API")
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(resp.Payload)},
		},
	}, nil, nil
}
