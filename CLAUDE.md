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

```
main.go                     # Entry point: SDK client init, MCP server, stdio transport
tools/
├── helpers.go              # Shared helpers (ToResult) — reusable across products
├── inputs.go               # Shared input structs with json/jsonschema tags
└── compute/
    ├── register.go         # RegisterAll() — calls all per-resource Register*Tools()
    ├── datacenter.go       # Datacenter tools (list, get)
    ├── server.go           # Server tools (list, get, sub-resources)
    ├── volume.go           # Volume tools
    └── ...                 # One file per resource (20 files total, 50 tools)
docs/compute/               # One doc file per resource
```

- **main.go**: Initializes the IONOS Cloud client, creates the MCP server via `mcp.NewServer()`, calls `compute.RegisterAll()`, and runs over `mcp.StdioTransport`.
- **tools/**: Shared input structs and helpers in the parent package, product-specific tools in sub-packages.
- **tools/compute/**: One file per resource. Each file exports a `Register*Tools()` function that adds tools via `mcp.AddTool()`.

### Request Flow

1. The official MCP SDK handles all JSON-RPC protocol framing over stdio
2. Tools are registered with typed Go structs — the SDK auto-generates JSON schemas and validates inputs
3. Each tool handler calls the IONOS Cloud SDK and returns results as `mcp.TextContent`

### Authentication

The server requires `IONOS_TOKEN` environment variable at startup. It exits with an error if the token is not set.

### Adding New Tools

1. Define an input struct in `tools/inputs.go` with `json` and `jsonschema` tags (non-pointer fields are automatically required)
2. Add a `mcp.AddTool()` call in the appropriate resource file under `tools/compute/` (e.g., `tools/compute/server.go`)
3. If it's a new resource, create a new file and register it in `tools/compute/register.go`
4. The handler receives the typed input struct and returns `(*mcp.CallToolResult, any, error)`

## Testing MCP Protocol

```bash
# Test initialization (keep stdin open briefly for response)
{ echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"0.1.0"}}}'; sleep 1; } | ./ionoscloud-mcp

# List available tools
{ echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"0.1.0"}}}'; echo '{"jsonrpc":"2.0","method":"notifications/initialized"}'; echo '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}'; sleep 1; } | ./ionoscloud-mcp
```
