package mcp

import (
	"context"
	"log"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// LoggingHooks returns hooks that log tool calls for debugging/monitoring.
func LoggingHooks() *server.Hooks {
	hooks := &server.Hooks{}

	// Log before any request
	hooks.AddBeforeAny(func(ctx context.Context, id any, method mcp.MCPMethod, message any) {
		log.Printf("[MCP] Request: method=%s id=%v", method, id)
	})

	// Log after tool calls
	hooks.AddAfterCallTool(func(ctx context.Context, id any, message *mcp.CallToolRequest, result *mcp.CallToolResult) {
		status := "success"
		if result != nil && result.IsError {
			status = "error"
		}
		log.Printf("[MCP] Tool call: name=%s status=%s", message.Params.Name, status)
	})

	// Log errors
	hooks.AddOnError(func(ctx context.Context, id any, method mcp.MCPMethod, message any, err error) {
		log.Printf("[MCP] Error: method=%s id=%v error=%v", method, id, err)
	})

	return hooks
}

// MetricsHooks returns hooks that collect metrics for tool calls.
// This is a placeholder for OpenTelemetry integration.
func MetricsHooks() *server.Hooks {
	hooks := &server.Hooks{}

	startTimes := make(map[any]time.Time)

	hooks.AddBeforeCallTool(func(ctx context.Context, id any, message *mcp.CallToolRequest) {
		startTimes[id] = time.Now()
	})

	hooks.AddAfterCallTool(func(ctx context.Context, id any, message *mcp.CallToolRequest, result *mcp.CallToolResult) {
		if start, ok := startTimes[id]; ok {
			duration := time.Since(start)
			log.Printf("[Metrics] Tool %s took %v", message.Params.Name, duration)
			delete(startTimes, id)
		}
	})

	return hooks
}
