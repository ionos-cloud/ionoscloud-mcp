# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Model Context Protocol (MCP) server that enables LLMs to interact with IONOS CLOUD infrastructure. Built with the [official MCP Go SDK](https://github.com/modelcontextprotocol/go-sdk) and the IONOS CLOUD Go SDK.

## Build Commands

```bash
make build      # Build the binary (outputs ./ionoscloud-mcp)
make install    # Install the binary to $GOBIN so MCP clients on PATH pick it up
make test       # Run unit tests
make test-e2e   # Binary-over-stdio (mocked API) + read-only live API checks
make fmt        # Format code with gofmt
make vet        # Run go vet
make lint       # Run golangci-lint (read-only)
make lintfix    # Run golangci-lint with --fix
make vuln       # Run govulncheck against all packages
make docker     # Build local Docker image (IMAGE= to override tag)
make deps       # go mod download + tidy
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

Environment variables (read from the MCP server process — typically inherited from the MCP client that spawns it, e.g. via `env` block in `.mcp.json` or the shell running Claude Code):
- `IONOS_TOKEN` — IONOS Cloud API token (all products). Not required to start the server. If unset, expired, or revoked, the IONOS API returns 401 which `tools.enrichSDKError` wraps with an actionable message (where to set the token in the MCP client config + restart hint) before it reaches the LLM.
- `IONOS_S3_ACCESS_KEY` + `IONOS_S3_SECRET_KEY` — Required only for Object Storage tools (per-region S3 endpoint authentication).

### Load modes

The server supports three tool-registration strategies via `IONOS_MCP_LOAD_MODE`:

- **`eager`** (default): all tools register at startup. Optimal for Claude Code (ToolSearch defers schemas client-side, ~1–3k tokens for names only) and required for clients without `notifications/tools/list_changed` support (Claude Desktop, claude.ai connectors, Claude in Chrome, Smithery scanner).

- **`lazy`**: defer Compute and Object Storage behind `ionos_load_compute_tools` / `ionos_load_objectstorage_tools` sentinel tools. Calling either registers the full product set and emits `notifications/tools/list_changed`. Only useful for clients that honour the notification AND lack client-side schema deferral.

- **`router`** (reserved, not yet implemented): single `ionos_search_tools` + `ionos_invoke` pair. Designed for clients with hard tool caps (Cursor 40, Windsurf 100) or no schema deferral. Currently logs a warning and falls back to `eager`. Implementation tracked separately.

Parsing is case-insensitive. Unknown values fall back to `eager` with a stderr warning. Empty / unset env var = eager.

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

## Testing

Three tiers, all runnable locally:

- **Unit** (`tools/**/*_test.go`, `server_config_test.go`): pure logic — error enrichment, validation, version/load-mode resolution, the Object Storage regional client cache, billing/activitylog transforms.
- **In-memory protocol** (`test/`): wires the full MCP server to an in-memory client over `mcp.NewInMemoryTransports`, with an `httptest` backend standing in for the IONOS API. The shared `h.run(t, tests)` runner asserts HTTP method/path (always) plus query params and tool output (per-case). `test/errors_test.go` covers protocol-level failure paths.
- **Binary e2e** (`test/e2e/`, `e2e` build tag): builds the real binary and drives it over stdio JSON-RPC against a mocked API injected via `IONOS_API_URL`. Exercises the shipped artifact, both load modes, resources, and the User-Agent.
- **Live e2e** (`test/live/`, `e2e_live` build tag): read-only discovery tests against the REAL IONOS API. Skips entirely without `IONOS_TOKEN`; object-storage chains additionally need `IONOS_S3_ACCESS_KEY`/`IONOS_S3_SECRET_KEY`. Each chain lists then drills in only if the account has the resource, so it stays green on an empty/reset account.

The `e2e` and `e2e_live` suites run locally only — not yet wired into CI. `make test-e2e` runs both (the binary suite and the live suite).

```bash
make test        # unit + in-memory (race)
make test-e2e    # binary-over-stdio (mocked API) + read-only live API
```

When adding a tool: add a `toolTest` case in the product's `test/*_test.go`.

## Testing MCP Protocol

```bash
# Test initialization (keep stdin open briefly for response)
{ echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"0.1.0"}}}'; sleep 1; } | ./ionoscloud-mcp

# List available tools
{ echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"0.1.0"}}}'; echo '{"jsonrpc":"2.0","method":"notifications/initialized"}'; echo '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}'; sleep 1; } | ./ionoscloud-mcp
```
