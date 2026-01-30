// Package mcp provides the MCP server implementation using mark3labs/mcp-go.
package mcp

import (
	"github.com/ionos-cloud/ionoscloud-mcp/pkg/api"
	"github.com/mark3labs/mcp-go/mcp"
)

// ConvertTool converts our internal api.Tool to the mcp-go library's mcp.Tool type.
func ConvertTool(t api.Tool) mcp.Tool {
	tool := mcp.Tool{
		Name:           t.Name,
		Description:    t.Description,
		RawInputSchema: t.InputSchema,
	}
	if t.Annotations != nil {
		tool.Annotations = convertAnnotations(*t.Annotations)
	}
	return tool
}

// convertAnnotations converts our internal api.ToolAnnotations to mcp.ToolAnnotation.
func convertAnnotations(a api.ToolAnnotations) mcp.ToolAnnotation {
	return mcp.ToolAnnotation{
		Title:           a.Title,
		ReadOnlyHint:    a.ReadOnlyHint,
		DestructiveHint: a.DestructiveHint,
		IdempotentHint:  a.IdempotentHint,
		OpenWorldHint:   a.OpenWorldHint,
	}
}

// ConvertResult converts our internal api.ToolCallResult to mcp.CallToolResult.
func ConvertResult(r *api.ToolCallResult) *mcp.CallToolResult {
	if r == nil {
		return mcp.NewToolResultText("")
	}
	if r.IsError {
		return mcp.NewToolResultError(r.Content)
	}
	return mcp.NewToolResultText(r.Content)
}
