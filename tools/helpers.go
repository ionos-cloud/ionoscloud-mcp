package tools

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ionos-cloud/sdk-go-bundle/shared"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const maxErrorBodyBytes = 2048

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

// sdkAPIError is what enrichSDKError needs from an SDK error. Matched by interface
// so every product's generated error type works without importing them all.
type sdkAPIError interface {
	error
	StatusCode() int
	Body() []byte
}

// enrichSDKError turns an IONOS SDK error into actionable guidance, since the raw
// message rarely says what to do about it.
func enrichSDKError(apiErr error) string {
	var sdkErr sdkAPIError
	if !errors.As(apiErr, &sdkErr) {
		return apiErr.Error()
	}

	if sdkErr.StatusCode() != 401 {
		return apiErr.Error()
	}

	body := sdkErr.Body()
	truncated := false
	if len(body) > maxErrorBodyBytes {
		body = body[:maxErrorBodyBytes]
		truncated = true
	}
	bodyStr := string(body)
	if truncated {
		bodyStr += "..."
	}

	return "IONOS API 401 Unauthorized. IONOS_TOKEN is missing, expired, or revoked. " +
		"Fix: set IONOS_TOKEN in your MCP client config (env block of .mcp.json / " +
		"claude_desktop_config.json) then restart the MCP client (restarting only your " +
		"shell does not propagate env to an already-running client). Original response: " +
		bodyStr
}

// ToResult marshals an API response into an MCP text result.
func ToResult(data any, apiErr error) (*mcp.CallToolResult, any, error) {
	if apiErr != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: enrichSDKError(apiErr)},
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

// ErrorText wraps a message as an MCP result flagged as an error.
func ErrorText(msg string) *mcp.CallToolResult {
	r := TextResult(msg)
	r.IsError = true
	return r
}

// IsNotFound reports whether err is an SDK error carrying HTTP 404.
func IsNotFound(err error) bool {
	var sdkErr sdkAPIError
	if !errors.As(err, &sdkErr) {
		return false
	}
	return sdkErr.StatusCode() == 404
}

// ToRawResult returns the raw response payload as an MCP text result.
// Use this for API endpoints that return non-JSON content (e.g. zone files).
func ToRawResult(resp *shared.APIResponse, apiErr error) (*mcp.CallToolResult, any, error) {
	if apiErr != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: enrichSDKError(apiErr)},
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
