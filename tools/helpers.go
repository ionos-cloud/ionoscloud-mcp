package tools

import (
	"encoding/json"
	"fmt"

	"github.com/ionos-cloud/sdk-go-bundle/shared"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

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
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal response: %w", err)
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(bytes)},
		},
	}, nil, nil
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
