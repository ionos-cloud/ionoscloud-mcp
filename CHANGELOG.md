## Upcoming

### Features
- `[FEA][ionoscloud-mcp]` Implement MCP server with 10 Compute Engine tools: datacenters, servers, volumes, images, locations, snapshots - @copilot
### Improvements
- `[IMP][ionoscloud-mcp]` Refactor from hand-rolled JSON-RPC to official MCP Go SDK with typed tool handlers via `mcp.AddTool()` - @avirtopeanu-ionos
### Dependencies
- `[IMP][ionoscloud-mcp]` Migrate from `sdk-go/v6` to `sdk-go-bundle` with `shared.NewConfigurationFromEnv()` - @cavramoniu-ionos
### Documentation
- `[IMP][ionoscloud-mcp]` Restructure tool docs into `docs/compute/` grouped by resource, fix CONTRIBUTING.md for `mcp.AddTool()` pattern, add "What is MCP?" section and CHANGELOG.md - @cavramoniu-ionos
