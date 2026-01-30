// Package api defines the core interfaces and types for the IONOS Cloud MCP server.
package api

import (
	"context"
	"encoding/json"
)

// ToolAnnotations provides hints about tool behavior for MCP clients.
// These annotations help clients understand the nature and impact of each tool.
type ToolAnnotations struct {
	// Title is a human-readable name for the tool
	Title string `json:"title,omitempty"`
	// ReadOnlyHint indicates the tool only reads data (no side effects)
	ReadOnlyHint *bool `json:"readOnlyHint,omitempty"`
	// DestructiveHint indicates the tool may delete or destroy resources
	DestructiveHint *bool `json:"destructiveHint,omitempty"`
	// IdempotentHint indicates repeated calls produce the same result
	IdempotentHint *bool `json:"idempotentHint,omitempty"`
	// OpenWorldHint indicates the tool may interact with external systems
	OpenWorldHint *bool `json:"openWorldHint,omitempty"`
}

// Tool represents an MCP tool definition with its schema and annotations.
type Tool struct {
	// Name is the unique identifier for the tool
	Name string `json:"name"`
	// Description explains what the tool does
	Description string `json:"description"`
	// InputSchema is the JSON Schema for the tool's parameters
	InputSchema json.RawMessage `json:"inputSchema"`
	// Annotations provides behavioral hints about the tool
	Annotations *ToolAnnotations `json:"annotations,omitempty"`
}

// ToolHandlerFunc is the signature for tool handler functions.
// It receives parameters and returns a result or error.
type ToolHandlerFunc func(ctx context.Context, params ToolHandlerParams) (*ToolCallResult, error)

// ServerTool combines a Tool definition with its handler function.
type ServerTool struct {
	Tool    Tool
	Handler ToolHandlerFunc
}

// Toolset is an interface for groups of related tools.
// Each toolset provides tools for a specific domain (compute, networking, etc.).
type Toolset interface {
	// GetName returns the unique name of the toolset
	GetName() string
	// GetDescription returns a human-readable description of the toolset
	GetDescription() string
	// GetTools returns all tools provided by this toolset
	GetTools() []ServerTool
}

// Helper functions to create pointer bools for annotations
func BoolPtr(b bool) *bool {
	return &b
}

// ReadOnly creates annotations marking a tool as read-only.
func ReadOnly(title string) *ToolAnnotations {
	return &ToolAnnotations{
		Title:        title,
		ReadOnlyHint: BoolPtr(true),
	}
}

// Destructive creates annotations marking a tool as destructive.
func Destructive(title string) *ToolAnnotations {
	return &ToolAnnotations{
		Title:           title,
		DestructiveHint: BoolPtr(true),
	}
}

// Idempotent creates annotations marking a tool as idempotent (create/update operations).
func Idempotent(title string) *ToolAnnotations {
	return &ToolAnnotations{
		Title:          title,
		IdempotentHint: BoolPtr(true),
	}
}

// NonIdempotent creates annotations for tools that are not idempotent (create operations).
func NonIdempotent(title string) *ToolAnnotations {
	return &ToolAnnotations{
		Title:          title,
		IdempotentHint: BoolPtr(false),
	}
}
