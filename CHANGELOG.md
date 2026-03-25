## Upcoming

### Features
- Implement MCP server with 10 Compute Engine tools: datacenters, servers, volumes, images, locations, snapshots @copilot
### Improvements
- Refactor from hand-rolled JSON-RPC to official MCP Go SDK with typed tool handlers via `mcp.AddTool()` @avirtopeanu-ionos
### Dependencies
- Migrate from `sdk-go/v6` to `sdk-go-bundle` with `shared.NewConfigurationFromEnv()` @cavramoniu-ionos
