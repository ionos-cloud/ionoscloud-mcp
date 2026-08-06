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
main.go             # Entry point: arg parsing, load-mode resolution, SDK clients init, MCP server, stdio transport
server_config.go    # Version resolution (ldflags > go install > vcs), resolveLoadMode()
resources.go        # MCP resources (embedded docs served to LLM clients)
tools/
├── helpers.go      # Shared helpers (TextResult)
├── inputs.go       # Shared input structs with json/jsonschema tags
├── ionosclient/    # User-Agent string builder
├── loader/         # Lazy loaders for compute and object storage
├── dynamic/        # 'dynamic' load mode: catalog + search/describe/call meta-tools
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
- **server_config.go**: Resolves `serverVersion` from ldflags (release builds), `go install` module version, or VCS revision (local builds). Also contains `resolveLoadMode(flagVal, envVal)` — the pure precedence resolver (flag > env > default) for the load mode.
- **resources.go**: Registers MCP _resources_ (distinct from tools) — structured documents served to LLM clients. Currently exposes `ionos://billing/focus-v1.3` (the FOCUS v1.3 billing spec, embedded from `docs/billing/focus-v1.3.md`).
- **tools/ionosclient/**: Builds the User-Agent string for all outbound IONOS API calls, including product name, server version, SDK bundle version, transport mode, and Go OS/arch.
- **tools/loader/**: Registers `ionos_load_compute_tools` and `ionos_load_objectstorage_tools` — sentinel tools that dynamically register the full product tool set on first call. Used in `lazy` mode. Once called, the tool list is updated and MCP clients receive a `notifications/tools/list_changed` signal.
- **tools/dynamic/**: Implements `dynamic` load mode. Builds a private in-memory "catalog" server with every product's tools (reusing their `RegisterAll`), self-connects to snapshot the tool metadata, and registers three meta-tools on the public server — `ionos_search_tools` (keyword search over the catalog), `ionos_describe_tools` (full input schemas), `ionos_call_tool` (forwards an invocation to the catalog server). The public tool list never changes.

### Request Flow

1. The official MCP SDK handles all JSON-RPC protocol framing over stdio
2. Tools are registered with typed Go structs — the SDK auto-generates JSON schemas and validates inputs
3. Each tool handler calls the IONOS CLOUD SDK and returns results as `mcp.TextContent`

### Authentication

Environment variables (read from the MCP server process — typically inherited from the MCP client that spawns it, e.g. via `env` block in `.mcp.json` or the shell running Claude Code):
- `IONOS_TOKEN` — IONOS Cloud API token (all products). Not required to start the server. If unset, expired, or revoked, the IONOS API returns 401 which `tools.enrichSDKError` wraps with an actionable message (where to set the token in the MCP client config + restart hint) before it reaches the LLM.
- `IONOS_S3_ACCESS_KEY` + `IONOS_S3_SECRET_KEY` — Required only for Object Storage tools (per-region S3 endpoint authentication).
- `IONOS_MCP_TOOL_SCOPE` — Opt-in write access; read-only by default. `write` enables `create_*`/`update_*`; `destructive` (which implies `write`) also enables `delete_*`. Comma-separated, case-insensitive, hierarchical; unrecognised values stay read-only. Parsed by `tools.ParseScope`, resolved once in `main.go`. See "Write operations & tool scope".

### Load modes

The server supports three tool-registration strategies, selectable via the `--load-mode` flag or the `IONOS_MCP_LOAD_MODE` env var. Precedence: flag > env > default (`eager`). Resolution is a pure function `resolveLoadMode(flagVal, envVal)` in `server_config.go` (returns the mode + its source); `main.go` resolves once at startup and logs `load mode: <mode> (source: ...)` to stderr.

- **`eager`** (default): all tools register at startup. Optimal for Claude Code (ToolSearch defers schemas client-side, ~1–3k tokens for names only) and required for clients without `notifications/tools/list_changed` support (Claude Desktop, claude.ai connectors, Claude in Chrome, Smithery scanner).

- **`lazy`**: defer Compute and Object Storage behind `ionos_load_compute_tools` / `ionos_load_objectstorage_tools` sentinel tools. Calling either registers the full product set and emits `notifications/tools/list_changed`. Only useful for clients that honour the notification AND lack client-side schema deferral. (Note: this runtime list-mutation pattern is the one the GitHub MCP server retired in 2026; `dynamic` is generally preferable for cap-limited clients.)

- **`dynamic`** (alias: `search`): exposes only three meta-tools — `ionos_search_tools`, `ionos_describe_tools`, `ionos_call_tool` (see `tools/dynamic/`). The full catalogue is registered onto a private in-memory "catalog" server (reusing each product's `RegisterAll` unchanged); the dynamic package self-connects over an in-memory transport, snapshots tool metadata at startup, and forwards `ionos_call_tool` to it. The public tool list never changes (no `list_changed` needed). For clients with hard tool caps and no tool search of their own (Cursor ~40, Windsurf 100). Not for Claude Code — keep it `eager`.

Parsing is case-insensitive. Any unknown value warns on stderr and falls back to `eager`. Empty / unset = eager.

In `eager` and `lazy` modes, the small products (DNS, Billing, Cert, Activity Log, k8s) always register eagerly. In `dynamic` mode every product — including those — is hidden behind the meta-tools. The product list is defined once as a `[]dynamic.Product` slice in `main.go` and shared across all three modes so they cannot drift.

### Write operations & tool scope

Read-only is the default; write tools are gated behind `IONOS_MCP_TOOL_SCOPE` (see Authentication). Key invariants:

- **Gate at registration, in one place.** Tools register through `tools.RegisterTool(server, scope, method, tool, handler)` (`tools/scope.go`). It classifies by HTTP `method`, sets MCP annotations, panics if the name prefix disagrees with the method (`create_`/`update_`/`delete_` vs `list_`/`get_`/`head_`), and skips registration entirely when the scope disallows the class — a skipped tool never enters `tools/list`.
- **No bypass across load modes.** `scope` is captured in the `main.go` compute closure and threaded into `compute.RegisterAll` and the lazy `RegisterComputeLoader`, so eager, lazy, and dynamic all apply the same filter. The dynamic dispatcher re-checks scope in `callHandler` (`tools/dynamic/`) as defense-in-depth, classifying catalog entries by name via `tools.ClassFromName`.
- **Two-phase confirmation** for create/delete lives in `tools.ConfirmationStore` (`tools/confirm.go`): a single-use, target-bound, TTL'd token minted on the first (preview) call and consumed on the second (execute) call. One shared store is built in `main.go` and threaded through registration, so both calls — including via `ionos_call_tool` in dynamic mode — hit the same store.
- **Tokens are session-scoped.** Build every target with `tools.Target(req, parts...)` — never by hand — because it prefixes `tools.CallerID(req)`. One process serves many clients over HTTP, so without it a token previewed by one client can be executed by another. `CallerID` reads the transport's session id first and only falls back to the `_meta` value the dynamic dispatcher forwards; that order is deliberate, since `_meta` is client-supplied and trusting it first would let a client claim another's id. On stdio the id is empty, which is correct — one client per process. `TestConfirmationTokensAreSessionScoped` covers eager and dynamic over real HTTP; an in-memory harness cannot tell sessions apart and would pass regardless.
- **No bulk creation:** `create_*` tools take no count/batch parameter (one resource per call).

**Every request goes through the IONOS SDK.** Never hand-build an HTTP call, even to work around an SDK defect and even when the SDK's own configuration would supply the base URL, credentials and HTTP client. If an operation cannot be expressed through a typed SDK call, the tool does not ship: document why in the register function and add a test asserting the tool stays absent. `update_ip_block` is the worked example — `IpBlockProperties` models `location` and `size` as non-pointer fields and serializes both unconditionally, while the API documents `location` as "disallowed in update requests", so no typed call can produce an acceptable body.

Two SDK serialization hazards govern every write tool. Both have caused real bugs; both are now covered by tests, but new resources must be checked against them.

- **PATCH bodies must contain only the fields the caller supplied.** Build the properties struct as a keyed or zero-valued literal (`&ionos.VolumeProperties{}`, `&ionos.ImageProperties{LicenceType: lt}`), never with a generated constructor. **Any** generated constructor may inject documented API defaults — not just the `...WithDefaults()` ones: `NewImageProperties(licenceType)` takes a required argument *and* sets `exposeSerial=false`, `requireLegacyBios=true`. Known injectors: `NicProperties` (`dhcp=true`), `VolumeProperties` (`exposeSerial`, `requireLegacyBios`, `bootOrder="AUTO"`), `SnapshotProperties` and `ImageProperties` (`exposeSerial`, `requireLegacyBios`). A PATCH built from one sends those as though the caller had asked, so renaming a volume would also force legacy BIOS on and reset its boot order — which can stop a server booting. Check with `sed -n '/^func New<Type>(/,/^}/p'` and look for `this.X = &`.
- **Non-pointer property fields are serialized unconditionally**, regardless of the above, so an update that omits them sends zero values. The update tool must read the resource and carry those fields forward. Known cases: `NicProperties.Lan`, `SecurityGroupProperties.Name`, `ImageProperties.LicenceType`, and across load balancing — `NetworkLoadBalancer`/`ApplicationLoadBalancerProperties` (`name`, `listenerLan`, `targetLan`), the two forwarding-rule models (including **`targets`**, whose loss empties a load balancer's backend pool), `TargetGroupProperties` (`name`, `algorithm`, `protocol`), `NatGatewayProperties` (`name`, `publicIps`), `NatGatewayRuleProperties` (`name`, `sourceSubnet`, `publicIp`). Check `ToMap` in the model file for any `toSerialize[...]` line not behind an `IsNil` guard.

  Carry-forward is not always the right answer, and assuming it is caused a bug. When the API forbids a field in update requests outright, sending the current value is still wrong — the objection is to its presence, not its value. `IpBlockProperties.Location`+`Size` are the case in point: both are unguarded non-pointer fields, but `location` is documented "disallowed in update requests", so there is no correct body and the tool cannot exist. Read the field comment before reaching for carry-forward.

`TestUpdateBodiesContainOnlyRequestedFields` asserts the exact JSON key set of every update body, and each carry-forward has a test that fails when it is removed. When adding a resource, add it to that table.

The presentation half of the two-phase flow lives in `tools/preview.go` (`tools.Preview`, `tools.Fields`, `tools.BlastRadius`, `tools.ConfirmErrorText`, `tools.Target`, `tools.HasToken`, `tools.DeletedAsync`, `tools.Opt*`). It is product-agnostic on purpose — DNS, cert and object storage write tools should reuse it rather than growing their own preview format, so a model only has to learn one shape.

The data-center write tools (`tools/compute/datacenter_write.go`) are the blueprint for extending writes to other resources; `tools/compute/server_write.go` additionally shows pre-flight validation, a delete flag bound into the confirmation target, and secret redaction in previews.

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
7. **Write tools** (`create_*`/`update_*`/`delete_*`) must register via `tools.RegisterTool(server, scope, tools.Method*, ...)` instead of `mcp.AddTool`, so they are scope-gated and annotated. Thread `scope` (and, for create/delete, the shared `*tools.ConfirmationStore`) down from `RegisterAll`, and put destructive ops behind the two-phase confirmation flow. See `tools/compute/datacenter_write.go`.

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
