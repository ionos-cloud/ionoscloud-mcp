# Handoff — Write operations for IONOS Cloud MCP (Compute)

## Purpose

The **data-center write tools** (`create_/update_/delete_datacenter`) are merged. They are the **blueprint** for the next goal: **write operations for all of Compute**. This doc captures the architecture, the exact research method, the decisions to keep consistent, and the open design questions — so a fresh session can extend the pattern without rediscovering everything.

Repo: `workspace/ionoscloud-mcp/` · branch that landed the blueprint: `feat/write-ops-datacenters`.

---

## 1. What already exists (reuse it — do not rebuild)

All the security machinery is generic and lives in the `tools` package. New resources only add per-resource handlers.

| File | What it gives you |
|---|---|
| `tools/scope.go` | `Scope`, `Class`, `Method`, `ParseScope`, `ClassFromName`, `nameMatchesMethod`, and **`RegisterTool[In,Out]`** — the single gate. Also `Method.annotations()`. |
| `tools/confirm.go` | `ConfirmationStore` (`Mint`/`Consume`): single-use, target-bound, 5-min TTL token store for two-phase confirm. |
| `tools/helpers.go` | `ToResult`, `TextResult`, **`ErrorText`** (IsError result), **`IsNotFound`** (404 check), `enrichSDKError`. |
| `tools/compute/datacenter_write.go` | **The blueprint** — copy this file's shape per resource. |
| `tools/compute/datacenter.go` | Reads migrated to `RegisterTool` (annotations). |

**Enablement:** `IONOS_MCP_TOOL_SCOPE` env, resolved once in `main.go` via `tools.ParseScope`. Hierarchical: `read` ⊂ `write` ⊂ `destructive` (destructive implies write). Read-only by default.

**How scope + the confirmation store reach handlers (already wired — no `main.go` change needed for new compute resources):**
- `main.go` builds `scope` and `confirm := tools.NewConfirmationStore()` once, and captures both in the compute `Register` closure: `compute.RegisterAll(s, client, scope, confirm)`.
- `tools/compute/register.go` `RegisterAll(server, client, scope, confirm)` already receives them. **Add your `RegisterXWriteTools(server, client, scope, confirm)` call here** next to `RegisterDatacenterWriteTools`.
- Lazy mode (`tools/loader/loader.go`) and dynamic mode (`tools/dynamic/`) already thread/re-check scope. Dynamic classifies by name prefix via `tools.ClassFromName` and re-gates in `callHandler`; `ionos_call_tool`'s annotation is derived from scope in `callToolAnnotations`. **As long as you follow the naming convention, dynamic mode needs no changes.**

**The gate (`tools.RegisterTool`):** classify by HTTP method → skip registration if scope disallows → set annotations from method → `mcp.AddTool`. It **panics at boot** if the tool name prefix doesn't match the method (`create_`↔POST, `update_`↔PUT/PATCH, `delete_`↔DELETE, `list_`/`get_`/`head_`↔GET/HEAD).

---

## 2. The blueprint pattern (copy `datacenter_write.go`)

Per resource, create `tools/compute/<resource>_write.go` with `RegisterXWriteTools(server, client, scope, confirm)` that registers:

- **create_<resource>** — `tools.MethodPost`, **two-phase confirmed**. Phase 1 (no token): validate + `confirm.Mint("create_<resource>", <target>)` + return preview (non-error `TextResult`). Phase 2 (token): `confirm.Consume(...)` → `XxxPost(...).Execute()` → `ToResult`. Target = the identifying inputs (datacenter uses `name|location`; nested resources should include the parent IDs, e.g. `datacenterID|name`). **One resource per call — no count/batch field.**
- **update_<resource>** — `tools.MethodPatch`, **single call** (not confirmed). Partial: only set provided fields; **omit immutable fields**. Use `NewXPropertiesPutWithDefaults()` + `SetX(...)`.
- **delete_<resource>** — `tools.MethodDelete`, **two-phase confirmed**. Phase 1: `FindById` (handle `IsNotFound` → "does not exist"), compute a preview (blast radius where relevant), `Mint("delete_<resource>", id)`. Phase 2: `Consume` → `XxxDelete(...).Execute()` → **`TextResult`** (DELETE returns no body — do **not** `ToResult` a struct).

Input structs go in `tools/inputs.go` (non-pointer = required; add a `ConfirmationToken *string` to create/delete inputs). Error mapping helpers: mirror `createConfirmError`/`deleteConfirmError` (map `ErrToken{Unknown,Expired,Mismatch}` → actionable text).

Naming: strict `create_<r>` / `update_<r>` / `delete_<r>` so the gate and client allow-rules pattern-match.

---

## 3. Where I researched the params (do this per resource)

The IONOS SDK is the source of truth. Module: `github.com/ionos-cloud/sdk-go-bundle/products/compute/v2 v2.0.7` (see `go.mod`). **Both v2.0.5 and v2.0.7 sit in the module cache — always read v2.0.7.** The bump was purely additive (no removals or signature changes, so nothing broke), but v2.0.5 is missing `ServerProperties.EnabledFeatures`, `ImageProperties.RequiredFeatures`, `TemplateProperties.StorageType`, `LocationProperties.MetroRegion` and `CpuArchitectureProperties.EnabledFeatures`. It's read-only generated code in the module cache:

```
/home/cavramoniu/go/pkg/mod/github.com/ionos-cloud/sdk-go-bundle/products/compute/v2@v2.0.7/
```

For each resource:

1. **API methods & request builders** → `api_<resource>.go` (e.g. `api_servers.go`, `api_volumes.go`, `api_lans.go`).
   - Method signatures: `XxxPost(ctx) ApiXxxPostRequest`, `XxxPut(ctx, ids...)`, `XxxPatch(...)`, `XxxDelete(...)`, `XxxFindById(...)`.
   - Body setter on the request builder: `.Server(...)`, `.Volume(...)`, `.Datacenter(...)`, etc.
   - `.Execute()` return types (note: **DELETE returns only `(*shared.APIResponse, error)` — no body**).
2. **Request/response models** → `model_<resource>_post.go`, `model_<resource>_properties_post.go`, `model_<resource>_put.go`, `model_<resource>_properties_put.go`.
   - Field list + **required vs optional**: non-pointer field = required (e.g. datacenter `Location string`); pointer = optional.
   - Constructors: `NewXPost(props)`, `NewXPropertiesPost(<requiredArgs>)`, `NewXPropertiesPutWithDefaults()`, plus `SetField(...)` setters.
   - Immutability notes are in the field doc comments (e.g. datacenter `Location` "cannot be modified after creation").
3. **Container blast radius** (only for resources that contain others) → `model_data_center_entities.go` pattern: nested collections each have `Items []X`; count via `len(coll.Items)` at depth 2.
4. **The official API docs** (cross-check required/optional/immutable, operationIds for doc links): `https://api.ionos.com/docs/cloud/v6/#tag/<Resource>`. Used for the `**API Reference:**` links in `docs/compute/*.md`.
5. **Existing READ tool** for the resource (`tools/compute/<resource>.go` + its input struct in `tools/inputs.go`) — shows the exact client call path (`client.<Resource>Api.<Method>`) and the parent-ID inputs to mirror.

Handy alternatives to reading cache files: `go doc github.com/ionos-cloud/sdk-go-bundle/products/compute/v2.ServerProperties` (etc.), and to enumerate mutating methods:
```
SDK=/home/cavramoniu/go/pkg/mod/github.com/ionos-cloud/sdk-go-bundle/products/compute/v2@v2.0.7
grep -rhoE '[A-Za-z]+(Post|Put|Patch|Delete)\(ctx' "$SDK"/api_*.go | sort -u
```

---

## 4. Scope of "all compute" — mutating API surface (from the SDK)

Full CRUD exists for (top-level and nested): **datacenters (done)**, servers, volumes, NICs, LANs, IP blocks, load balancers (+ balanced NICs), network load balancers (+ forwarding rules, flowlogs), application load balancers (+ forwarding rules, flowlogs), NAT gateways (+ rules, flowlogs), security groups (+ rules), firewall rules (NIC- and security-group-scoped), private cross-connects (`Pccs`), target groups, snapshots (Patch/Put/Delete only — created from a volume), images (Patch/Put/Delete only — no create), plus **labels** sub-resources on most types.

**Non-CRUD mutating actions (need a design decision — see §5):**
- Server power/lifecycle: `…ServersStartPost`, `…StopPost`, `…RebootPost`, `…SuspendPost`, `…ResumePost`, `…UpgradePost`.
- Volume snapshots: `…VolumesCreateSnapshotPost`, `…VolumesRestoreSnapshotPost`.

Read-only (no writes): locations, templates, contracts, requests, images (create), CPU/version discovery.

---

## 5. Decisions to keep + open questions for the next session

**Keep consistent with the blueprint:**
- Hierarchical scope; read-only default; no rate limiter / no caps / no env knobs.
- One resource per call (no batch param).
- Two-phase confirm for **create + delete**; **single-call update**.
- Annotations derived from HTTP method.
- Known limitation (documented): two-phase confirm does **not** stop an auto-approving client (the agent can walk both calls). The real boundary for unattended sessions is not granting `destructive`. (which is ok)

**Open decisions to make before coding all-compute:**
1. **Power/action verbs** (`start_server`, `reboot_server`, …). These are POST but not "create", so **`RegisterTool`'s `nameMatchesMethod` will panic** on a non-`create_` POST. Decide: add an action class/prefix (e.g. a new `Method` or a `create_`-exempt "action" path) or defer power actions. This is the biggest gate change needed.
2. **Which deletes warrant two-phase.** Datacenter delete cascades everything (big blast radius). Most leaf resources (a NIC, a firewall rule) delete only themselves. Decide: two-phase on every delete for consistency, or only on high-blast containers (server/volume) with a trivial preview otherwise.
3. **PUT vs PATCH for update.** Blueprint uses PATCH (partial, safer — avoids wiping omitted fields). Keep PATCH-only, or also expose PUT full-replace? (We chose PATCH-only for datacenter.)
4. **Labels & deep sub-resources** — in scope or a later pass? They multiply the tool count fast.
5. **Nested targets** — create/delete tokens must bind to parent IDs too (e.g. `datacenterID|serverID|nicName`) so a token can't act on the wrong parent.
6. **Tool-count blowup** — full compute write CRUD is dozens of tools; consider whether `dynamic`/`lazy` modes and the READ annotation backfill should happen alongside.

---

## 6. Registration & testing recipe (per resource)

1. Add input structs to `tools/inputs.go` (create/delete get `ConfirmationToken *string`).
2. Add `tools/compute/<resource>_write.go` (copy `datacenter_write.go`).
3. Wire `RegisterXWriteTools(server, client, scope, confirm)` into `tools/compute/register.go` `RegisterAll`.
4. Docs: extend the existing `docs/compute/<resource>.md` (same file as the read tools — that's the convention; see `docs/compute/datacenter.md`).
5. Tests (mirror `test/compute_write_test.go`): two-phase create/delete, single-call update PATCH body (`wantBody`), scope gating (tool hidden under read-only), bad/mismatched token, 404, and dynamic-mode parity via `ionos_call_tool`. Use `setupWithScope(t, scope)` / `setupDynamicWithScope(t, scope)` (already in `test/setup_test.go`, which records request bodies and supports `wantBody`).

**Verification (all must pass):**
```
go build ./...
go vet ./...
gofmt -l .            # must be empty
make test             # unit + in-memory, -race
go test -tags e2e ./test/e2e/...   # mocked binary over stdio (no real API, no token)
make lint             # golangci-lint, expect 0 issues
```
`make vuln` currently fails on a **stdlib** `crypto/tls` advisory (GO-2026-5856) — unrelated to this code; fixed by building with Go **1.26.5+** (`GOTOOLCHAIN=go1.26.5`), a separate toolchain chore.

---

## 7. Key gotchas (learned the hard way)

- **DELETE returns no body** → success path uses `TextResult`, error path uses `ToResult(nil, err)` (for 401 enrichment).
- **Immutable fields** (datacenter `Location`) must be omitted from update; a full PUT that omits a field can null it — another reason PATCH is safer.
- **`IsNotFound`** relies on the SDK error matching the `sdkAPIError` interface (the SDK returns errors *by value*, matched via `errors.As` on the interface — see `tools/helpers.go`). Reuse it, don't reinvent 404 detection.
- **Two calls share the store** even in dynamic mode because the private in-memory catalog server is a **single process-lifetime instance** whose handlers captured the one shared `*ConfirmationStore` — both `ionos_call_tool` forwards hit it. (In-memory server = same Go process, not a subprocess; the store is a shared pointer.)
- **Async ops**: create/delete are accepted asynchronously (202); return the created object (create) or an "accepted" `TextResult` (delete). Don't wait for provisioning in the handler.
- **`variable shadowing`**: in `tools/dynamic/catalog.go` the loop var was renamed `prodTools` because a local `tools` shadowed the `tools` package — watch for that if you add code touching the `tools` import there.
