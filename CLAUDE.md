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
├── k8s/            # Managed Kubernetes tools (clusters, node pools, nodes, versions)
├── dns/            # DNS tools (zones, records, reverse records, secondary zones, DNSSEC, quota)
├── billing/        # Billing tools (invoices, usage, utilization, traffic, EVN)
├── cert/           # Certificate Manager tools (certificates, auto-certs, providers)
├── activitylog/    # Activity Log tools (contracts, events)
└── objectstorage/  # Object Storage tools (buckets, objects, access keys, regions)
docs/
├── billing/        # One doc per tool group + focus-v1.3.md (embedded as MCP resource)
├── compute/        # One doc per resource
├── k8s/            # Managed Kubernetes docs
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
- **Fields the spec marks `x-internal` are never exposed as tool inputs.** Only seven exist in the CloudAPI spec — `vnet`, `placementGroupId`, `vni`, `KubernetesNodePoolLan.datacenterId` and node pool `taints` (on all three property models) — and most are documented "requires system privileges, for internal usage only". The generated SDK models them anyway, so presence in the SDK is not evidence a field is callable. Node pool `taints` is the worked example: no tool accepts it, `TestK8sNodepoolToolsDoNotExposeTaints` asserts it stays absent, and `update_k8s_nodepool` nonetheless carries existing taints forward — declining to accept a field is not a reason to destroy it on a replacing PUT.
- **An async endpoint must say so in its description.** Every mutating Kubernetes endpoint answers `202 Accepted`, so the tool returns before the change has taken effect. Each write tool appends `asyncResourceNote` or `asyncNodeNote` (`tools/k8s/async.go`), which name the polling tool, the `metadata.state` values that mean done/working/failed, and the fact that a `BUSY` resource *queues* further modifications rather than rejecting them. Cluster/node pool state and node state are different enums — point at the right one. `TestK8sWriteToolsDeclareAsyncBehaviour` enforces this.

  **Completion semantics are per-operation, not per-product.** DNS is the counter-example: most of its mutations answer 202 and expose `metadata.state`, but `reverserecordsPost` answers 201 and `reverserecordsPut` 200 with the finished record, `zonesZonefilePut` answers 200, and `ReverseRecordRead` carries the bare `Metadata` type with **no** state field — so a tool promising a polling step there would send the model looking for a field that does not exist. `tools/dns/async.go` holds one note per case and says why reverse records differ; `TestDnsWriteToolsDeclareCompletionSemantics` asserts the async tools say "Asynchronous (202)" and the synchronous ones say "Synchronous" and never claim otherwise.

  Certificate Manager splits the other way and is the cleaner illustration: its POSTs answer `201` and its PATCHes `200` with the stored resource, and only its DELETEs are `202`. `tools/cert/async.go` therefore has a note per verb rather than per resource, plus `stateNote` (every cert resource carries `metadata.state`) and `issuanceNote` — a `201` from `autoCertificatesPost` means the renewal *configuration* exists, not that a certificate has been issued, so the tool points at `metadata.lastIssuedCertificate` instead of letting the model assume the certificate is ready. `TestCertWriteToolsDeclareCompletionSemantics` pins the split.

**Every request goes through the IONOS SDK.** Never hand-build an HTTP call, even to work around an SDK defect and even when the SDK's own configuration would supply the base URL, credentials and HTTP client. If an operation cannot be expressed through a typed SDK call, the tool does not ship: document why in the register function and add a test asserting the tool stays absent. `update_ip_block` is the worked example — `IpBlockProperties` models `location` and `size` as non-pointer fields and serializes both unconditionally, while the API documents `location` as "disallowed in update requests", so no typed call can produce an acceptable body.

**Return the API's response, not a summary of it, wherever the response carries information the caller needs.** `delete_dns_zone_dnssec_key` is the worked example: it used to answer with a hand-written "requested, accepted" message, which hid a real `409` (`paas-dns-rest-0513`, "the zone has too many operations in progress") behind prose and led to a wrong diagnosis that the endpoint silently ignored deletes. It now returns the SDK's typed body through `tools.ToResult`, so the model reads the API's own words. Prose is fine where the endpoint genuinely returns nothing (`tools.DeletedAsync` for the compute deletes); it is not fine as a substitute for a body that exists.

**A write-only field is write-only in both directions.** A tool result and a two-phase preview are both transcript that the model and its MCP client keep, so a secret must not appear in either — nor in a validation error, which is as loggable as a preview. Certificate Manager is the product with write-only *inputs*: `Certificate.privateKey` and `Provider.externalAccountBinding.keySecret`. The spec marks both `writeOnly`, but the generated SDK models them on the read types as ordinary fields, so nothing stops a response echoing one. `tools/cert/redact.go` is the single choke point — every cert read and write handler returns its body through it, so a newly added secret field has one place to go — previews render secrets through `tools.Redacted`, and `validatePEM` names the offending field without quoting its value back. `TestCreateCertCertificatePreviewNeverEchoesThePrivateKey`, `TestCreateCertCertificateResultNeverReturnsThePrivateKey`, `TestCreateCertCertificateValidationNeverQuotesTheKey` and `TestUpdateCertProviderNeverReturnsTheEabSecret` are the guards.

**Bind a confirmation token to the content, not just the name, when the content is what matters.** `tools.Target` takes as many parts as the operation needs. `create_cert_certificate` includes a SHA-256 digest of the certificate, chain and key (`materialDigest`), so a token minted while previewing one certificate cannot upload a different key — the same reasoning as `import_dns_zone_file` hashing the zone file. `TestCreateCertCertificateTokenIsBoundToTheMaterial` asserts a swapped key is refused and mutates nothing.

**Never state a status code a tool's endpoint does not return.** Success codes come from the spec per `operationId`, not from the HTTP method: of the sixteen DNS mutations, `reverserecordsPost` is `201` and `reverserecordsPut` and `zonesZonefilePut` are `200`, while the other thirteen are `202`. `TestDnsWriteToolsDeclareCompletionSemantics` pins which tools may claim to be asynchronous; re-derive the codes from the spec rather than from the method when adding one.

Two SDK serialization hazards govern every write tool. Both have caused real bugs; both are now covered by tests, but new resources must be checked against them.

- **PATCH bodies must contain only the fields the caller supplied.** Build the properties struct as a keyed or zero-valued literal (`&ionos.VolumeProperties{}`, `&ionos.ImageProperties{LicenceType: lt}`), never with a generated constructor. **Any** generated constructor may inject documented API defaults — not just the `...WithDefaults()` ones: `NewImageProperties(licenceType)` takes a required argument *and* sets `exposeSerial=false`, `requireLegacyBios=true`. Known injectors: `NicProperties` (`dhcp=true`), `VolumeProperties` (`exposeSerial`, `requireLegacyBios`, `bootOrder="AUTO"`), `SnapshotProperties` and `ImageProperties` (`exposeSerial`, `requireLegacyBios`). A PATCH built from one sends those as though the caller had asked, so renaming a volume would also force legacy BIOS on and reset its boot order — which can stop a server booting. Check with `sed -n '/^func New<Type>(/,/^}/p'` and look for `this.X = &`.
- **Non-pointer property fields are serialized unconditionally**, regardless of the above, so an update that omits them sends zero values. The update tool must read the resource and carry those fields forward. Known cases: `NicProperties.Lan`, `SecurityGroupProperties.Name`, `ImageProperties.LicenceType`, and across load balancing — `NetworkLoadBalancer`/`ApplicationLoadBalancerProperties` (`name`, `listenerLan`, `targetLan`), the two forwarding-rule models (including **`targets`**, whose loss empties a load balancer's backend pool), `TargetGroupProperties` (`name`, `algorithm`, `protocol`), `NatGatewayProperties` (`name`, `publicIps`), `NatGatewayRuleProperties` (`name`, `sourceSubnet`, `publicIp`). Check `ToMap` in the model file for any `toSerialize[...]` line not behind an `IsNil` guard.

- **A PUT endpoint replaces the resource's properties, so carry-forward covers every mutable field, not just the unguarded ones.** Managed Kubernetes (`K8sPut`, `K8sNodepoolsPut`) and DNS (every update) are the products here with PUT updates, and all of them are affected. `KubernetesNodePoolPropertiesForPut.NodeCount` is the sharpest case — a non-pointer field, so a plain rename that does not carry it forward sends `nodeCount: 0` and drains the pool; its constructor additionally injects `serverType=DedicatedCore`, which would rebuild a VCPU pool's nodes as dedicated-core ones. `KubernetesClusterPropertiesForPut.Name` is unguarded the same way, and dropping the cluster's `apiSubnetAllowList` silently exposes the Kubernetes API server to every source address. `tools/k8s/{cluster,nodepool}_write.go` read the resource and override only the supplied fields; `TestUpdateK8sClusterCarriesFieldsForward` and `TestUpdateK8sNodepoolCarriesNodeCountForward` assert the exact PUT key set and fail if a carry-forward is removed.

  **A GET response is not necessarily a legal PUT body**, so blanket carry-forward needs a per-field sanity check. Two live rejections on Managed Kubernetes made the point, neither of which the mocked suite could have caught: a node pool with no autoscaler reads back as `autoScaling: {minNodeCount: 0, maxNodeCount: 0}`, but writing that shape is rejected with "autoScaling.minNodeCount must be > 0" — so resending it broke *every* update to such a pool; and the node pool `name` is immutable, so sending it back at all is refused. Carry-forward now skips an inactive autoscaler (`autoScalingActive`) and never sends the pool name. `TestUpdateK8sNodepoolDropsInactiveAutoScaling` and `TestUpdateK8sNodepoolNeverSendsName` are the guards, and `poolFixture` carries `autoScaling: {0,0}` on purpose so the first fails in CI rather than against an account.

  Carry-forward is not always the right answer, and assuming it is caused a bug. When the API forbids a field in update requests outright, sending the current value is still wrong — the objection is to its presence, not its value. `IpBlockProperties.Location`+`Size` are the case in point: both are unguarded non-pointer fields, but `location` is documented "disallowed in update requests", so there is no correct body and the tool cannot exist. Read the field comment before reaching for carry-forward.

  DNS adds two variants of the same hazard. `SecondaryZone.PrimaryIps` is an unguarded **slice**, so a nil value serializes as `"primaryIps":null` rather than being omitted, and the API rejects it — carry-forward is the only legal body. And `Record.Ttl`/`Priority`/`Enabled` are guarded pointers, which looks safe but is not: the spec gives `ttl` a default of 3600 and `enabled` a default of `true`, so on a replacing PUT an omitted `ttl` cannot be relied on to preserve the record's current value. `tools/dns/record_write.go` therefore sends all three explicitly, carrying the current value forward when the caller omits one. Because every DNS update is a PUT, the identity fields (`Zone.ZoneName`, `Record.Name`/`Type`, `ReverseRecord.Ip`) are carried forward and deliberately **not** exposed as tool inputs — `TestUpdateDnsToolsDoNotExposeIdentityFields` asserts they stay absent, and `TestDnsUpdateBodiesContainOnlyExpectedFields` pins each PUT's exact key set.

- **Where a PATCH accepts only one field, the update tool must say so.** All three Certificate Manager PATCH endpoints take the spec's `PatchName` — `{"properties": {"name": "..."}}` — so `update_cert_*` renames and nothing else. There is no carry-forward hazard (the body has one required field, which the caller always supplies) but there is a description hazard: a model that assumes an update can rotate a certificate's key, or repoint an auto-certificate at another provider, would believe it had made a change the endpoint cannot make. Each description ends with `renameNote` and names what is immutable and what to do instead; `TestUpdateCertToolsSayOnlyTheNameChanges` enforces the note and `TestUpdateCertToolsSendOnlyTheName` pins each PATCH body to exactly `{properties: {name}}` with no `metadata` key. The three tools share one `rename` helper (`tools/cert/write.go`) because only the SDK call differs.

`TestUpdateBodiesContainOnlyRequestedFields` asserts the exact JSON key set of every update body, and each carry-forward has a test that fails when it is removed. When adding a resource, add it to that table.

- **A tool whose HTTP method understates what it does needs an action verb, not a CRUD prefix.** `import_dns_zone_file` is a `PUT`, so `Method.Class()` would call it `ClassWrite` and a `write`-scoped server could wipe every record in a zone. It registers through `tools.RegisterActionTool` with the `import_` verb, which `actionVerbs` maps to `ClassDestructive`; `TestImportDnsZoneFileIsGatedAsDestructive` asserts a write-only scope does not expose it. Adding a verb to `actionVerbs` is mandatory, not optional — `RegisterActionTool` panics at startup on an unknown one.

The presentation half of the two-phase flow lives in `tools/preview.go` (`tools.Preview`, `tools.Fields`, `tools.BlastRadius`, `tools.ConfirmErrorText`, `tools.Target`, `tools.HasToken`, `tools.DeletedAsync`, `tools.Redacted`, `tools.IncompleteRadiusNote`, `tools.CappedCountNote`, `tools.ErrLabel`, `tools.Opt*`). It is product-agnostic on purpose — DNS and cert write tools use it unchanged, and object storage should do the same rather than growing its own preview format, so a model only has to learn one shape. The last four started as private helpers (or inline text) in `tools/compute` and `tools/dns`; a second product needing one is the signal to promote rather than copy.

`tools.IncompleteRadiusNote` carries an invariant worth restating: an unreadable collection must never render as an empty one. A preview that says "this auto-certificate has issued no certificates" when the list call actually returned 500 is a false statement in the one place a caller looks before authorizing a delete. Every counting helper therefore returns `(count, capped, error)` — `capped` because these list endpoints report no total, so a full page is a floor — and the caller clears its `EmptyNote` when the note fires. `TestDeleteCertAutoCertificatePreviewNeverClaimsEmptyOnError` and `TestDeleteCertProviderPreviewNeverClaimsEmptyOnError` mirror the DNS pair.

The data-center write tools (`tools/compute/datacenter_write.go`) are the blueprint for extending writes to other resources; `tools/compute/server_write.go` additionally shows pre-flight validation, a delete flag bound into the confirmation target, and secret redaction in previews. For a resource whose update is a PUT rather than a PATCH, `tools/k8s/nodepool_write.go` is the worked example: read-then-override carry-forward across every mutable field, with the enum/shape validation factored into `tools/k8s/validate.go` so both the create and the update normalise inputs the same way.

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
