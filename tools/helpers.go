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

// sdkAPIError is the behavioural contract enrichSDKError needs from an IONOS
// SDK error. It is matched by interface (not concrete type) on purpose: the
// product SDKs return shared.GenericOpenAPIError *by value* from their API
// methods (e.g. dns api_zones.go), so a *shared.GenericOpenAPIError target
// would never bind via errors.As and 401s would silently pass through
// un-enriched. Matching the interface binds both value and pointer forms.
type sdkAPIError interface {
	error
	StatusCode() int
	Body() []byte
}

// enrichSDKError augments IONOS SDK errors with actionable guidance for the
// LLM. Only 401 is enriched — 403 is intentionally left alone because IONOS
// returns 403 for several distinct cases (wrong contract, missing role,
// resource-level ACL) that need separate guidance not yet researched.
//
// Non-SDK errors and other status codes pass through unchanged.
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

// ErrorText wraps a message as an MCP text result flagged as an error
// (IsError). Use it for actionable, user-facing failures (missing input,
// confirmation-token problems) that are surfaced as tool content rather than a
// transport-level error.
func ErrorText(msg string) *mcp.CallToolResult {
	r := TextResult(msg)
	r.IsError = true
	return r
}

// IsNotFound reports whether err is an IONOS SDK error carrying HTTP 404. It
// matches by the same behavioural interface as enrichSDKError (the SDK returns
// its error by value), so both value and pointer forms bind.
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
