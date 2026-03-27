package tools

import (
	"encoding/json"
	"fmt"

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
