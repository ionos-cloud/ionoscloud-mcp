## Upcoming

### Features
- `[FEA][ionoscloud-mcp]` Add 2 read-only Activity Log tools (list contracts, query events) — 109 tools total; events response is compacted (strips `_source` wrapper, `auditVersion`, redundant `contractNumber`, duplicate `sourceService`/`initiator` fields) for ~35% smaller output - @avirtopeanu-ionos
- `[FEA][ionoscloud-mcp]` Implement MCP server with 10 Compute Engine tools: datacenters, servers, volumes, images, locations, snapshots - @copilot
- `[FEA][ionoscloud-mcp]` Add 40 new read-only compute tools covering all ionosctl compute engine resources (networking, load balancers, NAT gateways, security groups, etc.) — 50 tools total - @cavramoniu-ionos
- `[FEA][ionoscloud-mcp]` Add 14 read-only DNS tools (zones, zone files, records, reverse records, secondary zones, DNSSEC, quota) — 64 tools total - @cavramoniu-ionos
- `[FEA][ionoscloud-mcp]` Add 14 read-only Billing tools (profile, EVN, invoices, products, traffic, usage, utilization) — 78 tools total; EVN and traffic responses drop CSV/array duplicate fields to reduce output size - @cavramoniu-ionos
- `[FEA][ionoscloud-mcp]` Add 23 read-only Object Storage tools (buckets, bucket config, objects, access keys, regions) — 101 tools total - @cavramoniu-ionos
- `[FEA][ionoscloud-mcp]` Add 6 read-only Certificate Manager v2 tools: list/get certificates, auto-certificates, and providers — 107 tools total - @cavramoniu-ionos
### Testing
- `[FEA][ionoscloud-mcp]` Add integration tests for all 64 tools using httptest + MCP in-memory transport — verifies correct API endpoint routing without real credentials - @cavramoniu-ionos
- `[FEA][ionoscloud-mcp]` Add integration tests for all 14 billing tools — verifies correct API endpoint routing and period validation - @cavramoniu-ionos
- `[FEA][ionoscloud-mcp]` Add integration tests for all 23 Object Storage tools - @cavramoniu-ionos
- `[FEA][ionoscloud-mcp]` Add integration tests for all 6 Certificate Manager tools — verifies correct API endpoint routing without real credentials - @cavramoniu-ionos
### Improvements
- `[IMP][ionoscloud-mcp]` Refactor from hand-rolled JSON-RPC to official MCP Go SDK with typed tool handlers via `mcp.AddTool()` - @avirtopeanu-ionos
- `[IMP][ionoscloud-mcp]` Refactor monolithic `ionos.go` into per-resource files under `tools/compute/`, move shared helpers and input structs to `tools/` package for cross-product reuse - @cavramoniu-ionos
### Dependencies
- `[IMP][ionoscloud-mcp]` Migrate from `sdk-go/v6` to `sdk-go-bundle` with `shared.NewConfigurationFromEnv()` - @cavramoniu-ionos
### Documentation
- `[IMP][ionoscloud-mcp]` Restructure tool docs into `docs/compute/` grouped by resource, fix CONTRIBUTING.md for `mcp.AddTool()` pattern, add "What is MCP?" section and CHANGELOG.md - @cavramoniu-ionos
- `[IMP][ionoscloud-mcp]` Add documentation for all 50 compute tools in `docs/compute/`, update CONTRIBUTING.md for new project structure - @cavramoniu-ionos
- `[IMP][ionoscloud-mcp]` Switch JSON responses from `MarshalIndent` to `Marshal` (compact JSON) — ~30% reduction in output size across all tools - @cavramoniu-ionos
- `[IMP][ionoscloud-mcp]` Add documentation for Certificate Manager tools (certificate, auto-certificate, provider) in `docs/cert/` - @cavramoniu-ionos
- `[IMP][ionoscloud-mcp]` Align brand to "IONOS CLOUD" across user-facing content (docs, README, tool descriptions) - @mimihalescu
