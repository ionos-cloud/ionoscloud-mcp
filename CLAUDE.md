# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Model Context Protocol (MCP) server that enables LLMs to interact with IONOS CLOUD infrastructure. Built with the [official MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk) and the IONOS CLOUD Go SDK.

## Build Commands

```bash
make build      # Build the binary (outputs ./ionoscloud-mcp)
make run        # Build and run the MCP server
make test       # Run unit tests
make fmt        # Format code with gofmt
make vet        # Run go vet
make check      # Run fmt + vet together
make lint       # Run golangci-lint (read-only)
make lintfix    # Run golangci-lint with --fix
make vuln       # Run govulncheck against all packages
make docker     # Build local Docker image (IMAGE= to override tag)
make snapshot   # Dry-run GoReleaser pipeline locally (no publish)
make deps       # go mod download + tidy
make dev        # check + build + run
make clean      # Remove build artifacts and dist/
```

Pass `VERSION=<tag>` to `make build` or `make docker` to override the version string (defaults to `dev`).

## Architecture

```
main.go             # Entry point: all SDK clients init, MCP server, stdio transport
server_config.go    # Version resolution (ldflags > go install > vcs), eagerLoad()
resources.go        # MCP resources (embedded docs served to LLM clients)
tools/
├── helpers.go      # Shared helpers (TextResult)
├── inputs.go       # Shared input structs with json/jsonschema tags
├── ionosclient/    # User-Agent string builder
├── loader/         # Lazy loaders for compute and object storage
├── compute/        # Compute Engine tools (servers, datacenters, volumes, NICs, etc.)
├── dns/            # DNS tools (zones, records, DNSSEC, quota)
├── billing/        # Billing tools (invoices, usage, utilization, traffic, EVN)
├── cert/           # Certificate Manager tools (certificates, auto-certs, providers)
├── activitylog/    # Activity Log tools (contracts, events)
└── objectstorage/  # Object Storage tools (buckets, objects, access keys, regions)
docs/
├── billing/        # One doc per tool group + focus-v1.3.md (embedded as MCP resource)
├── compute/        # One doc per resource
├── dns/            # One doc per resource
├── cert/           # Certificate Manager docs
├── activitylog/    # Activity Log docs
└── objectstorage/  # Object Storage docs
```

- **main.go**: Initializes all SDK clients (compute, DNS, billing, cert, object storage base + management, activity log), creates the MCP server, and runs over `mcp.StdioTransport`. All clients share a single `*http.Client` with the custom User-Agent `RoundTripper` installed.
- **server_config.go**: Resolves `serverVersion` from ldflags (release builds), `go install` module version, or VCS revision (local builds). Also contains `eagerLoad()` which reads `IONOS_MCP_EAGER_LOAD`.
- **resources.go**: Registers MCP _resources_ (distinct from tools) — structured documents served to LLM clients. Currently exposes `ionos://billing/focus-v1.3` (the FOCUS v1.3 billing spec, embedded from `docs/billing/focus-v1.3.md`).
- **tools/ionosclient/**: Builds the User-Agent string for all outbound IONOS API calls, including product name, server version, SDK bundle version, transport mode, and Go OS/arch.
- **tools/loader/**: Registers `ionos_load_compute_tools` and `ionos_load_objectstorage_tools` — sentinel tools that dynamically register the full product tool set on first call. Used in lazy mode (default). Once called, the tool list is updated and MCP clients receive a `notifications/tools/list_changed` signal.

### Request Flow

1. The official MCP SDK handles all JSON-RPC protocol framing over stdio
2. Tools are registered with typed Go structs — the SDK auto-generates JSON schemas and validates inputs
3. Each tool handler calls the IONOS CLOUD SDK and returns results as `mcp.TextContent`

### Authentication

Required environment variables:
- `IONOS_TOKEN` — IONOS Cloud API token (all products). The server exits with an error if not set.
- `IONOS_S3_ACCESS_KEY` + `IONOS_S3_SECRET_KEY` — Required only for Object Storage tools (per-region S3 endpoint authentication).

### Lazy vs Eager Loading

By default, **Compute** and **Object Storage** tools are not registered at startup. Instead, two loader tools are available: `ionos_load_compute_tools` and `ionos_load_objectstorage_tools`. Calling either tool registers the full product set and notifies the MCP client.

Set `IONOS_MCP_EAGER_LOAD=true` to register all tools at startup. Use this for MCP clients that do not handle `notifications/tools/list_changed` (e.g. some Claude Desktop configurations, claude.ai connectors, Claude in Chrome).

All other products (DNS, Billing, Cert, Activity Log) are always registered eagerly.

### Adding New Tools

1. Define an input struct in `tools/inputs.go` with `json` and `jsonschema` tags (non-pointer fields are automatically required)
2. Add a `mcp.AddTool()` call in the appropriate resource file under `tools/<product>/`
3. If it's a new resource within an existing product, create a new file and register it in `tools/<product>/register.go`
4. If it's a new product:
   - Create a new sub-package under `tools/<product>/` with a `RegisterAll()` function
   - Add the SDK import and client initialization to `main.go`
   - Call `RegisterAll()` from `main()` (either directly for eager products, or via a new loader for lazy products)
   - Add docs under `docs/<product>/`
5. If it needs to be a lazy-loaded product, add a loader function in `tools/loader/loader.go`
6. The handler receives the typed input struct and returns `(*mcp.CallToolResult, any, error)`

### Adding MCP Resources

Resources are registered in `resources.go` via `server.AddResource()`. Use `//go:embed` to inline static content (e.g. spec documents). Resources are served to LLM clients that call `resources/read` and are useful for reference documents the LLM should consult when generating output.

## Testing MCP Protocol

```bash
# Test initialization (keep stdin open briefly for response)
{ echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"0.1.0"}}}'; sleep 1; } | ./ionoscloud-mcp

# List available tools
{ echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"0.1.0"}}}'; echo '{"jsonrpc":"2.0","method":"notifications/initialized"}'; echo '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}'; sleep 1; } | ./ionoscloud-mcp
```
