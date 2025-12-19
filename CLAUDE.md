# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Model Context Protocol (MCP) server that enables LLMs to interact with IONOS Cloud infrastructure. The server communicates via JSON-RPC over stdio and uses the official IONOS Cloud Go SDK.

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

The codebase consists of three files with clear separation:

- **main.go**: MCP server core - JSON-RPC protocol handling, request routing, and tool registration. Contains the `Server` struct that holds the IONOS client and tool definitions.
- **ionos.go**: IONOS Cloud API implementations - all tool execution logic that calls the SDK (`listDatacenters`, `getServer`, etc.)
- **tool.go**: Tool type definition

### Request Flow

1. Server reads JSON-RPC requests line-by-line from stdin
2. `handleRequest()` routes to appropriate handler based on method (`initialize`, `tools/list`, `tools/call`)
3. For `tools/call`, `executeTool()` dispatches to the corresponding IONOS API function in ionos.go
4. Results are returned as JSON-RPC responses to stdout

### Authentication

The server reads credentials from environment variables at startup:
- `IONOS_USERNAME` + `IONOS_PASSWORD` for username/password auth
- `IONOS_TOKEN` for token-based auth

### Adding New Tools

1. Add tool definition to `registerTools()` in main.go (name, description, JSON schema)
2. Add case to `executeTool()` switch in ionos.go
3. Implement the API function in ionos.go following existing patterns

## Testing MCP Protocol

```bash
# Test initialization
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | ./ionoscloud-mcp

# List available tools
echo '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}' | ./ionoscloud-mcp
```
