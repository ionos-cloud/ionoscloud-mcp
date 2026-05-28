# Changelog

## v1.0.0 - June 2026

Initial release. MCP server exposing 109 read-only tools across IONOS CLOUD products: Compute, DNS, Billing, Object Storage, Certificate Manager, and Activity Log.

### Added

- **Compute Engine tools (50)**: Datacenters, servers, volumes, images, locations, snapshots, NICs, LANs, IP blocks, load balancers (ALB/NLB), NAT gateways, security groups, firewall rules, target groups, private cross-connects, templates, server GPUs, server CD-ROMs, remote console, and request status — full coverage of ionosctl compute resources.
- **DNS tools (14)**: Zones, zone files, records, reverse records, secondary zones (incl. AXFR status), DNSSEC keys, and quota.
- **Billing tools (14)**: Profile, contract, invoices (incl. by period), products, traffic (incl. by period), usage (incl. by datacenter), utilization (daily + by period), and EVN (incl. by period). EVN and traffic responses drop CSV/array duplicate fields to reduce output size.
- **Object Storage tools (23)**: Buckets (list, head, location, CORS, encryption, lifecycle, lock, policy, policy status, public access block, replication, tagging, versioning), objects (list, head, versions, legal hold, retention, tagging), access keys, and regions.
- **Certificate Manager v2 tools (6)**: List/get certificates, auto-certificates, and providers.
- **Activity Log tools (2)**: List contracts, query events. Events response strips `_source` wrapper, `auditVersion`, redundant `contractNumber`, and duplicate `sourceService`/`initiator` fields for ~35% smaller output.
- **`IONOS_TOKEN` auth**: Required at startup; server exits if unset.
- **stdio transport**: Standard MCP transport for integration with MCP-capable LLM clients.

### Changed

- **Compact JSON responses**: All tools use `json.Marshal` instead of `MarshalIndent` — ~30% reduction in output size.
- **Brand alignment**: "IONOS CLOUD" used consistently across docs, README, and tool descriptions.
