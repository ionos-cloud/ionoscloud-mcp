## Upcoming

### Features
- `[FEA][ionoscloud-mcp]` Implement MCP server with 10 Compute Engine tools: datacenters, servers, volumes, images, locations, snapshots - @copilot
- `[FEA][ionoscloud-mcp]` Add 40 new read-only compute tools covering all ionosctl compute engine resources (networking, load balancers, NAT gateways, security groups, etc.) — 50 tools total - @cavramoniu-ionos
### Improvements
- `[IMP][ionoscloud-mcp]` Refactor from hand-rolled JSON-RPC to official MCP Go SDK with typed tool handlers via `mcp.AddTool()` - @avirtopeanu-ionos
- `[IMP][ionoscloud-mcp]` Refactor monolithic `ionos.go` into per-resource files under `tools/compute/`, move shared helpers and input structs to `tools/` package for cross-product reuse - @cavramoniu-ionos
### Dependencies
- `[IMP][ionoscloud-mcp]` Migrate from `sdk-go/v6` to `sdk-go-bundle` with `shared.NewConfigurationFromEnv()` - @cavramoniu-ionos
### Documentation
- `[IMP][ionoscloud-mcp]` Restructure tool docs into `docs/compute/` grouped by resource, fix CONTRIBUTING.md for `mcp.AddTool()` pattern, add "What is MCP?" section and CHANGELOG.md - @cavramoniu-ionos
- `[IMP][ionoscloud-mcp]` Add documentation for all 50 compute tools in `docs/compute/`, update CONTRIBUTING.md for new project structure - @cavramoniu-ionos
