# Changelog

## Unreleased

### Added

- **Write operations for Certificate Manager**, off by default and behind the same `IONOS_MCP_TOOL_SCOPE` gate: create, update and delete for certificates, auto-certificates and ACME providers. Every create and delete takes two calls — the first returns a preview and a one-time token, the second carries it. The 6 existing Certificate Manager read tools now carry the `readOnlyHint` annotation.

  All three PATCH endpoints accept only the spec's `PatchName`, so `update_cert_*` renames the resource and nothing else — a certificate's material, an auto-certificate's provider and DNS names, and a provider's email, directory URL and external account binding are all immutable, and each tool's description says so. There is deliberately **no** `create_cert_certificate`: `certificatesPost` requires the private key in the body, so the tool would have to take it as an argument, putting it in the model's context, the on-disk session transcript, and every later request to the model provider — none of which redaction reaches, since it guards the outbound response and preview. Issue certificates with `create_cert_auto_certificate`, where IONOS generates and holds the key, or upload outside this server. Write-only secrets that *do* cross (a provider's `external_account_binding.key_secret`) are never echoed in a preview, never quoted back in a validation error, and stripped from every response. Each create's confirmation token is bound to every input that changes what it does, so a field cannot be swapped between the preview a human approved and the call that executes. `delete_cert_auto_certificate` counts the certificates it has issued and `delete_cert_provider` the auto-certificates issuing through it; neither reports zero when the count could not be read.

- **Write operations for DNS**, off by default and behind the same `IONOS_MCP_TOOL_SCOPE` gate: create, update and delete for primary zones, records, secondary zones and reverse records, plus DNSSEC enable/disable, `start_dns_zone_transfer` and `import_dns_zone_file`. Every create and delete takes two calls — the first returns a preview and a one-time token, the second carries it. The 14 existing DNS read tools now carry the `readOnlyHint` annotation.

  `import_dns_zone_file` replaces *every* record in a zone, so it requires `destructive` rather than `write` despite being a `PUT`, and its preview counts the records it would replace. `delete_dns_zone_dnssec_key` warns that the registrar's DS record must be removed first — a stale DS record makes validating resolvers answer SERVFAIL for the whole zone. Updates are PUT, so each one reads the resource and carries every field it does not change forward, including the identity fields (a zone's name, a record's name and type, a reverse record's IP), which are therefore not tool inputs.

## 1.1.0

### Added

- **Write operations for Compute Engine**, off by default. Opt in with `IONOS_MCP_TOOL_SCOPE`: `write` enables create and update, `destructive` also enables delete. Covers data centers, servers, volumes, NICs, LANs, firewall rules, security groups and their rules, IP blocks, classic/network/application load balancers and their forwarding rules, NAT gateways, target groups, cross connects, snapshots and images — plus server power control (start, stop, reboot, suspend, resume, upgrade), volume snapshot create and restore, and volume attach/detach.

  Creating and deleting takes two calls: the first returns a preview of exactly what will change plus a one-time token, the second carries that token and executes. A delete preview also lists what else it affects — the servers and volumes inside a data center, the NICs on a LAN, the backends behind a forwarding rule.

  > ⚠️ These tools create real resources that cost money, and deletions cannot be undone. Read the warning in the README before enabling them. Enabling them is done at your own risk; IONOS Cloud is not responsible for actions the model takes as a result.

- **Write operations for Managed Kubernetes**, off by default and behind the same `IONOS_MCP_TOOL_SCOPE` gate: create, update and delete for clusters and node pools, plus `recreate_k8s_node` and `delete_k8s_node`. Node pool creation covers autoscaling, attached LANs with static routes, labels, annotations and reserved public IPs. Every create and delete takes two calls — the first returns a preview and a one-time token, the second carries it. All of these are asynchronous, so poll `get_k8s_cluster` or `get_k8s_nodepool` and read `metadata.state` before chaining a dependent call.

  Three IONOS API limits to be aware of: a node pool cannot be renamed, an autoscaler's bounds can be changed but it cannot be switched off, and `delete_k8s_node` is not a reliable way to replace a node — it removes the node first, and an active autoscaler may hold the pool at the smaller size, so prefer `recreate_k8s_node`.


- **Read-only annotations on the Kubernetes read tools**: the 8 existing `list_*`/`get_*` Kubernetes tools now carry the `readOnlyHint` annotation that clients use to decide whether to prompt before a call. No behaviour or signature change for callers.

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
