# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Model Context Protocol (MCP) server that enables LLMs to interact with IONOS Cloud infrastructure. Built with the [official MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk) and the IONOS Cloud Go SDK.

## Build Commands

```bash
make build      # Build the binary (outputs ./ionoscloud-mcp)
make test       # Run tests
make fmt        # Format code
make vet        # Run go vet
make check      # Run fmt and vet together
make deps       # Download and tidy dependencies
make clean      # Remove build artifacts
```

## Architecture

The codebase consists of two files:

- **main.go**: Entry point - initializes the IONOS Cloud client, creates the MCP server via `mcp.NewServer()`, registers tools, and runs over `mcp.StdioTransport`.
- **ionos.go**: Tool definitions and IONOS Cloud API implementations. Input types use Go structs with `jsonschema` tags for automatic schema inference. All tools are registered in `registerTools()` using the generic `mcp.AddTool()`.

### Request Flow

1. The official MCP SDK handles all JSON-RPC protocol framing over stdio
2. Tools are registered with typed Go structs - the SDK auto-generates JSON schemas and validates inputs
3. Each tool handler calls the IONOS Cloud SDK and returns results as `mcp.TextContent`

### Authentication

The server reads credentials from environment variables at startup:
- `IONOS_USERNAME` + `IONOS_PASSWORD` for username/password auth
- `IONOS_TOKEN` for token-based auth

### Adding New Tools

1. Define an input struct in ionos.go with `json` and `jsonschema` tags (non-pointer fields are automatically required)
2. Add a `mcp.AddTool()` call in `registerTools()` with the tool name, description, and handler function
3. The handler receives the typed input struct and returns `(*mcp.CallToolResult, any, error)`

## Testing MCP Protocol

```bash
# Test initialization (keep stdin open briefly for response)
{ echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"0.1.0"}}}'; sleep 1; } | ./ionoscloud-mcp

# List available tools
{ echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"0.1.0"}}}'; echo '{"jsonrpc":"2.0","method":"notifications/initialized"}'; echo '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}'; sleep 1; } | ./ionoscloud-mcp
```
