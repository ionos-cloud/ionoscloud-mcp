# Changelog

## Unreleased

### Added

- **Write operations for Compute Engine**, off by default. Opt in with `IONOS_MCP_TOOL_SCOPE`: `write` enables create and update, `destructive` also enables delete. Covers data centers, servers, volumes, NICs, LANs, firewall rules, security groups and their rules, IP blocks, classic/network/application load balancers and their forwarding rules, NAT gateways, target groups, cross connects, snapshots and images — plus server power control (start, stop, reboot, suspend, resume, upgrade), volume snapshot create and restore, and volume attach/detach.

  Creating and deleting takes two calls: the first returns a preview of exactly what will change plus a one-time token, the second carries that token and executes. A delete preview also lists what else it affects — the servers and volumes inside a data center, the NICs on a LAN, the backends behind a forwarding rule.

  > ⚠️ These tools create real resources that cost money, and deletions cannot be undone. Read the warning in the README before enabling them. Enabling them is done at your own risk; IONOS Cloud is not responsible for actions the model takes as a result.

- **Write operations for Managed Kubernetes**, off by default and behind the same `IONOS_MCP_TOOL_SCOPE` gate: `create_k8s_cluster`, `update_k8s_cluster`, `delete_k8s_cluster`, `create_k8s_nodepool`, `update_k8s_nodepool`, `delete_k8s_nodepool`, `recreate_k8s_node` and `delete_k8s_node`. Node pool creation covers autoscaling, attached LANs with static routes, labels, annotations and reserved public IPs.

  `recreate_k8s_node` is a `POST` but is classified **destructive**, because it throws the existing node away; it brings the replacement up and joins it to the cluster before draining the old one, so it is the safe way to replace an unhealthy node — at the cost of one extra billable node while both are running. `delete_k8s_node` is the opposite: it removes the node first, leaving the pool a node short, and is **not** a reliable replace operation — with an active autoscaler the autoscaler owns the count and can hold the pool at the smaller size (observed live: 2 nodes → 1, and it stayed there). Its preview warns when an autoscaler is active, and refuses outright when the delete would breach the autoscaler minimum, which the API otherwise reports as the misleading "last node can not be deleted from nodepool". Scale on purpose with `update_k8s_nodepool`. Both node tools, plus every create and delete, are two-phase confirmed.

  Both Kubernetes update endpoints are `PUT` rather than `PATCH`, meaning the API replaces the resource's properties. Both update tools therefore read the resource first and send back every field the caller omitted, so a rename cannot reset the node count to zero and drain a pool, drop a node pool's LANs, labels or taints, or clear a cluster's `api_subnet_allow_list` and expose its Kubernetes API server. Supplying `lans`, `labels`, `annotations`, `public_ips`, `api_subnet_allow_list` or `s3_buckets` replaces that list outright — the tool descriptions and docs say so, and an empty array is the explicit way to clear one.

  **Every mutating Kubernetes endpoint answers `202 Accepted`**, so all eight tools return before the change has taken effect. Each description now says so and names what to poll: `metadata.state` on the cluster or node pool (`ACTIVE` done, `DEPLOYING`/`UPDATING`/`BUSY` working, `FAILED_*` not), or the separate node state enum (`PROVISIONING`, `PROVISIONED`, `READY`, `TERMINATING`, `REBUILDING`, `BUSY`) for the single-node tools. They also warn that a `BUSY` resource *queues* further modifications rather than rejecting them, and that a node pool `k8s_version` upgrade replaces every node one at a time.

  Two node pool fields are deliberately never sent back on an update, both found by testing against a live account: the pool **`name`** (the API rejects it as immutable, so there is no `name` parameter on `update_k8s_nodepool` — recreate the pool to rename it) and an **inactive autoscaler** (a pool without one reads back as `{minNodeCount: 0, maxNodeCount: 0}`, which the API rejects on a write, so resending it would have broken every update to such a pool). Turning an autoscaler **off** is not supported: the API answers `422` for zero bounds and silently ignores an omitted `autoScaling`, so no working request exists. Zero bounds are rejected with that explanation rather than accepted as a no-op — delete and recreate the pool without `auto_scaling` instead. (`ionosctl` has the same gap.)

  Node pool **taints** are deliberately not exposed: the API spec marks them `x-internal`, the same marker it uses for `vnet`, `placementGroupId` and other system-privilege fields, so they are not part of the customer-facing surface. `update_k8s_nodepool` still carries existing taints forward, because a replacing `PUT` that omitted them would drop taints applied out of band. `cpu_family` is flagged as deprecated by IONOS in favour of `server_type`, and the private-cluster fields (`public`, `location`, `nat_gateway_ip`, `node_subnet`) are flagged prerelease.

- **Read-only annotations on the Kubernetes read tools**: the 8 existing `list_*`/`get_*` Kubernetes tools now register through the same scope-gated path as every other product, so they carry the `readOnlyHint` annotation that clients use to decide whether to prompt before a call. No behaviour or signature change for callers.

- **`filters` parameter on all compute `list_*` tools**: every compute list tool now accepts an optional `filters` object (`{"property": "value"}` pairs) that is forwarded to the API as server-side query params, so only matching resources are returned. Useful for scoping large result sets by name, location, state, image type, and more without client-side post-processing. If a filter returns an empty result, retry without it — a typo or value mismatch silently returns nothing. Filterable properties vary by resource (e.g. `name`, `location`, `vmState`, `cpuFamily`, `imageType`).

## v1.0.1 — June 2026

### Added

- **optional `depth` parameter on compute and Kubernetes tools**: every `list_*` tool now defaults to `depth=1`, returning resource names and properties in a single API call instead of only UUIDs. Every `get_*` tool accepts an optional `depth` to control nesting. Eliminates the agent loop of N follow-up `get_*` calls when resolving names.

- **`dynamic` load mode** (alias `search`): exposes only three meta-tools — `ionos_search_tools`, `ionos_describe_tools`, `ionos_call_tool` — through which the model discovers and invokes the full catalogue at runtime. The public tool list never changes (no `notifications/tools/list_changed` needed), so it fits clients with hard tool caps and no tool search of their own (Cursor ~40, Windsurf 100). Select with `--load-mode dynamic` or `IONOS_MCP_LOAD_MODE=dynamic`.

- **`--load-mode` flag**: selects the tool-registration strategy (`eager` | `lazy` | `dynamic`) from the command line. Takes precedence over `IONOS_MCP_LOAD_MODE`; the effective mode and its source are logged to stderr at startup.

- **Managed Kubernetes**: 8 new read-only tools as a dedicated product (eagerly loaded) — `list_k8s_clusters`, `get_k8s_cluster`, `list_k8s_nodepools`, `get_k8s_nodepool`, `list_k8s_nodepool_nodes`, `get_k8s_node`, `list_k8s_versions`, `get_k8s_default_version`. Covers the full cluster/node-pool/node hierarchy plus available version discovery.

## v1.0.0 — June 2026

First release. An MCP server that lets LLM clients explore an IONOS CLOUD account read-only.

### Added

- **Homebrew install**: `brew install ionos-cloud/ionos-cloud/ionoscloud-mcp` (macOS + Linux, Intel + Apple Silicon).
- **Docker image**: `ghcr.io/ionos-cloud/ionoscloud-mcp` (multi-arch: linux/amd64, linux/arm64). Pull and run with `docker run -i --rm -e IONOS_TOKEN=… ghcr.io/ionos-cloud/ionoscloud-mcp`.
- **Pre-built binaries** on every GitHub release for linux, macOS, and Windows (amd64 + arm64).
- **`--version` flag**: prints the server version. Useful for bug reports and version pinning.
- **Compute Engine**: Browse datacenters, servers, volumes, NICs, LANs, IP blocks, load balancers (ALB/NLB), NAT gateways, security groups, firewall rules, target groups, private cross-connects, snapshots, images, templates, locations, server GPUs and CD-ROMs, remote console URLs, and provisioning request status.
- **DNS**: Inspect zones, zone files, records, reverse records, secondary zones (incl. AXFR transfer status), DNSSEC keys, and account quota.
- **Billing**: View your contract and billing profile, invoices, product catalog, traffic and usage breakdowns (per datacenter, per period), and daily utilization.
- **Object Storage**: List buckets and objects across regions, read every bucket configuration surface (CORS, encryption, lifecycle, locking, policy, public access, replication, tagging, versioning), object metadata, retention and legal hold state, access keys, and regions.
- **Certificate Manager**: List certificates and auto-certificates, look up issuance providers.
- **Activity Log**: Look up accessible contracts, query the audit trail of API requests (who did what, when, on which resource). Defaults to the last 7 days and excludes noisy async-provisioning echoes — both overridable.

### Try

- **Cost audit**: "Audit my IONOS CLOUD account, find the top 5 cost-inducing resources this month, and suggest cost-efficiency tips."
- **Security sweep**: "List every bucket whose public access block is off or whose policy is public — flag anything that looks unintentional."
- **Audit trail**: "Show me every failed API request on my contract in the last 30 days, grouped by user."
- **Forgotten resources**: "Find unattached volumes, unused IP blocks, and stopped servers across all my datacenters."
- **DNS sanity check**: "List all zones on my account and flag any without DNSSEC enabled or with records pointing to IPs I no longer own."
- **Certificate expiry**: "Which certificates on my account expire in the next 60 days?"
- **Traffic spike investigation**: "My last invoice was higher than usual — show me daily traffic and utilization for the previous billing period and tell me what changed."
- **Onboarding tour**: "Walk me through what I have running on IONOS CLOUD — datacenters, servers, storage, DNS — like you're explaining it to a new teammate."

### Notes

- Auth via the `IONOS_TOKEN` environment variable. Server refuses to start without it.
- All responses are compacted before being returned to the LLM (no pretty-print, redundant audit envelopes stripped) so a single tool call uses far fewer tokens.
- Brand: tool descriptions and docs consistently use "IONOS CLOUD".
