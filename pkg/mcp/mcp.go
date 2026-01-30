package mcp

import (
	"context"
	"fmt"
	"log"

	"github.com/ionos-cloud/ionoscloud-mcp/pkg/api"
	"github.com/ionos-cloud/ionoscloud-mcp/pkg/ionos"
	"github.com/ionos-cloud/ionoscloud-mcp/pkg/toolsets"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	// ServerName is the name of the MCP server.
	ServerName = "ionoscloud-mcp"
	// ServerVersion is the version of the MCP server.
	ServerVersion = "0.2.0"
)

// Server wraps the MCP server with IONOS Cloud client integration.
type Server struct {
	mcpServer *server.MCPServer
	client    *ionos.Client
}

// Config holds configuration options for the MCP server.
type Config struct {
	// EnableLogging enables request/response logging
	EnableLogging bool
	// EnableMetrics enables basic timing metrics
	EnableMetrics bool
	// ReadOnly if true, only registers read-only tools
	ReadOnly bool
	// EnabledToolsets limits which toolsets are enabled (empty = all)
	EnabledToolsets []string
}

// NewServer creates a new MCP server with the given IONOS Cloud client.
func NewServer(client *ionos.Client, cfg *Config) *Server {
	if cfg == nil {
		cfg = &Config{}
	}

	// Build server options
	opts := []server.ServerOption{
		server.WithToolCapabilities(false), // We don't support tools/list_changed notifications
	}

	// Add hooks for logging/metrics
	hooks := buildHooks(cfg)
	if hooks != nil {
		opts = append(opts, server.WithHooks(hooks))
	}

	// Create the MCP server
	mcpServer := server.NewMCPServer(ServerName, ServerVersion, opts...)

	s := &Server{
		mcpServer: mcpServer,
		client:    client,
	}

	// Register all tools from toolsets
	s.registerTools(cfg)

	return s
}

// buildHooks creates hooks based on config.
func buildHooks(cfg *Config) *server.Hooks {
	if !cfg.EnableLogging && !cfg.EnableMetrics {
		return nil
	}

	hooks := &server.Hooks{}

	if cfg.EnableLogging {
		loggingHooks := LoggingHooks()
		hooks.OnBeforeAny = append(hooks.OnBeforeAny, loggingHooks.OnBeforeAny...)
		hooks.OnAfterCallTool = append(hooks.OnAfterCallTool, loggingHooks.OnAfterCallTool...)
		hooks.OnError = append(hooks.OnError, loggingHooks.OnError...)
	}

	if cfg.EnableMetrics {
		metricsHooks := MetricsHooks()
		hooks.OnBeforeCallTool = append(hooks.OnBeforeCallTool, metricsHooks.OnBeforeCallTool...)
		hooks.OnAfterCallTool = append(hooks.OnAfterCallTool, metricsHooks.OnAfterCallTool...)
	}

	return hooks
}

// registerTools registers all tools from enabled toolsets.
func (s *Server) registerTools(cfg *Config) {
	enabledSet := make(map[string]bool)
	for _, name := range cfg.EnabledToolsets {
		enabledSet[name] = true
	}

	for _, ts := range toolsets.Toolsets() {
		// Skip if toolset not in enabled list (when list is not empty)
		if len(enabledSet) > 0 && !enabledSet[ts.GetName()] {
			continue
		}

		for _, tool := range ts.GetTools() {
			// Skip non-read-only tools if in read-only mode
			if cfg.ReadOnly && !isReadOnly(tool.Tool) {
				continue
			}

			s.registerTool(tool)
		}
	}
}

// isReadOnly checks if a tool has the ReadOnlyHint set to true.
func isReadOnly(tool api.Tool) bool {
	if tool.Annotations == nil {
		return false
	}
	return tool.Annotations.ReadOnlyHint != nil && *tool.Annotations.ReadOnlyHint
}

// registerTool registers a single tool with the MCP server.
func (s *Server) registerTool(tool api.ServerTool) {
	mcpTool := ConvertTool(tool.Tool)

	// Create handler that bridges our handler signature to mcp-go's
	handler := s.createHandler(tool)

	s.mcpServer.AddTool(mcpTool, handler)
}

// createHandler creates an mcp-go ToolHandlerFunc from our api.ToolHandlerFunc.
func (s *Server) createHandler(tool api.ServerTool) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Create our handler params
		params := api.ToolHandlerParams{
			Context:   ctx,
			Client:    s.client.Compute,
			DNSClient: s.client.DNS,
			Arguments: request.GetArguments(),
		}

		// Call our handler
		result, err := tool.Handler(ctx, params)
		if err != nil {
			// Return errors in the result, not as protocol errors
			return mcp.NewToolResultError(err.Error()), nil
		}

		return ConvertResult(result), nil
	}
}

// Run starts the MCP server on stdio (standard input/output).
func (s *Server) Run() error {
	log.Printf("Starting %s v%s", ServerName, ServerVersion)
	return server.ServeStdio(s.mcpServer)
}

// MCPServer returns the underlying mcp-go server for advanced usage.
func (s *Server) MCPServer() *server.MCPServer {
	return s.mcpServer
}

// ToolCount returns the number of registered tools.
func (s *Server) ToolCount() int {
	return len(s.mcpServer.ListTools())
}

// NewServerFromEnv creates a new MCP server with IONOS client initialized from environment.
func NewServerFromEnv(cfg *Config) (*Server, error) {
	client, err := ionos.NewClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create IONOS client: %w", err)
	}
	return NewServer(client, cfg), nil
}
