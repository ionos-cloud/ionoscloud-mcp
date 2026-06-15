# Changelog

## v1.0.1 — June 2026

### Added

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
