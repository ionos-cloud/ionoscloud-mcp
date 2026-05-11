# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Model Context Protocol (MCP) server that enables LLMs to interact with IONOS CLOUD infrastructure. Built with the [official MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk) and the IONOS CLOUD Go SDK.

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
├── compute/
│   ├── register.go         # RegisterAll() — calls all per-resource Register*Tools()
│   ├── datacenter.go       # Datacenter tools (list, get)
│   ├── server.go           # Server tools (list, get, sub-resources)
│   ├── volume.go           # Volume tools
│   └── ...                 # One file per resource (20 files, 50 tools)
└── dns/
    ├── register.go         # RegisterAll() — calls all per-resource Register*Tools()
    ├── zone.go             # Zone tools (list, get)
    ├── record.go           # Record tools (list all, list by zone, get, list secondary)
    ├── reverse_record.go   # Reverse record tools (list, get)
    ├── secondary_zone.go   # Secondary zone tools (list, get, AXFR status)
    ├── dnssec.go           # DNSSEC key tools (list by zone)
    └── quota.go            # Quota tools (get)
docs/
├── compute/                # One doc file per compute resource
└── dns/                    # One doc file per DNS resource
```

- **main.go**: Initializes IONOS CLOUD clients (compute + DNS), creates the MCP server via `mcp.NewServer()`, calls `compute.RegisterAll()` and `dns.RegisterAll()`, and runs over `mcp.StdioTransport`.
- **tools/**: Shared input structs and helpers in the parent package, product-specific tools in sub-packages.
- **tools/compute/**: One file per resource. Each file exports a `Register*Tools()` function that adds tools via `mcp.AddTool()`.
- **tools/dns/**: Same pattern as compute. Each product has its own SDK client (`dns.APIClient` vs `compute.APIClient`), both initialized from the same `shared.Configuration`.

### Request Flow

1. The official MCP SDK handles all JSON-RPC protocol framing over stdio
2. Tools are registered with typed Go structs — the SDK auto-generates JSON schemas and validates inputs
3. Each tool handler calls the IONOS CLOUD SDK and returns results as `mcp.TextContent`

### Authentication

The server requires `IONOS_TOKEN` environment variable at startup. It exits with an error if the token is not set.

### Adding New Tools

1. Define an input struct in `tools/inputs.go` with `json` and `jsonschema` tags (non-pointer fields are automatically required)
2. Add a `mcp.AddTool()` call in the appropriate resource file under `tools/<product>/` (e.g., `tools/compute/server.go`, `tools/dns/zone.go`)
3. If it's a new resource, create a new file and register it in `tools/<product>/register.go`
4. If it's a new product, create a new sub-package, add a `RegisterAll()` function, create a new SDK client in `main.go`, and call `RegisterAll()` from `main()`
5. The handler receives the typed input struct and returns `(*mcp.CallToolResult, any, error)`

## Testing MCP Protocol

```bash
# Test initialization (keep stdin open briefly for response)
{ echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"0.1.0"}}}'; sleep 1; } | ./ionoscloud-mcp

# List available tools
{ echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"0.1.0"}}}'; echo '{"jsonrpc":"2.0","method":"notifications/initialized"}'; echo '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}'; sleep 1; } | ./ionoscloud-mcp
```
