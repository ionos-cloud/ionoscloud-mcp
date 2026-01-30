# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Model Context Protocol (MCP) server that enables LLMs to interact with IONOS Cloud infrastructure. The server communicates via JSON-RPC over stdio and uses the official IONOS Cloud Go SDK.

## Build Commands

```bash
make build          # Build the binary (outputs ./ionoscloud-mcp)
make test           # Run tests
make fmt            # Format code
make vet            # Run go vet
make check          # Run fmt and vet together
make deps           # Download and tidy dependencies
make clean          # Remove build artifacts
make list-toolsets  # List available toolsets
make list-tools     # List all available tools
```

## Architecture

The codebase uses a modular toolset-based plugin architecture:

```
cmd/
  ionoscloud-mcp/
    main.go                    # Cobra CLI entry point
pkg/
  api/
    toolsets.go                # Toolset, ServerTool, Tool interfaces
    types.go                   # ToolHandlerParams, ToolCallResult, helpers
  config/
    config.go                  # Configuration from env vars
  ionos/
    client.go                  # IONOS Cloud client wrapper (compute + DNS)
    helpers.go                 # Validation helpers
  mcp/
    mcp.go                     # MCP Server wrapping mark3labs/mcp-go
    gosdk.go                   # Type adapters (internal types <-> mcp-go types)
    middleware.go              # Logging and metrics hooks
  toolsets/
    toolsets.go                # Registry with Register(), Toolsets(), AllTools()
    compute/                   # 28 tools
      toolset.go, datacenters.go, servers.go, volumes.go, snapshots.go, images.go
    networking/                # 35 tools
      toolset.go, lans.go, nics.go, ipblocks.go, firewall.go, natgateways.go, pcc.go
    loadbalancing/             # 8 tools
      toolset.go, alb.go, nlb.go, targetgroups.go
    kubernetes/                # 8 tools
      toolset.go, clusters.go, nodepools.go
    iam/                       # 10 tools
      toolset.go, users.go, groups.go, s3keys.go, contract.go
    dns/                       # 4 tools
      toolset.go, zones.go, records.go
```

### Toolsets Summary

| Toolset | Tools | Description |
|---------|-------|-------------|
| compute | 28 | Datacenters, Servers, Volumes, Snapshots, Images |
| networking | 35 | LANs, NICs, IP Blocks, Firewall Rules, NAT Gateways, PCCs |
| loadbalancing | 8 | Application Load Balancers, Network Load Balancers, Target Groups |
| kubernetes | 8 | K8s Clusters, Node Pools, Nodes |
| iam | 10 | Users, Groups, S3 Keys, Contract, Resources |
| dns | 4 | DNS Zones, DNS Records |
| **Total** | **93** | |

### Key Architectural Patterns

#### 1. Toolset Plugin Architecture

Tools are organized into self-contained **toolsets** that self-register at init time:

```go
// pkg/toolsets/compute/toolset.go
type Toolset struct{}

func (t *Toolset) GetName() string { return "compute" }
func (t *Toolset) GetDescription() string {
    return "Compute resources (datacenters, servers, volumes, snapshots, images)"
}
func (t *Toolset) GetTools() []api.ServerTool { ... }

func init() {
    toolsets.Register(&Toolset{})  // Self-registration
}
```

#### 2. Tool Annotations

All tools have annotations for safety classification:

```go
type ToolAnnotations struct {
    Title           string
    ReadOnlyHint    *bool  // true = safe, no side effects (54 tools)
    DestructiveHint *bool  // true = can delete/destroy resources (14 tools)
    IdempotentHint  *bool  // true = same result on repeat
    OpenWorldHint   *bool  // true = external interactions
}
```

#### 3. Unified Handler Signature

```go
type ToolHandlerFunc func(ctx context.Context, params ToolHandlerParams) (*ToolCallResult, error)

type ToolHandlerParams struct {
    Context   context.Context
    Client    *ionoscloud.APIClient  // Compute API
    DNSClient *dns.APIClient         // DNS API
    Arguments map[string]interface{}
}
```

### Adding New Tools

1. Choose the appropriate toolset package (e.g., `pkg/toolsets/compute/`)
2. Add tool definition with handler in the appropriate file:
   ```go
   {
       Tool: api.Tool{
           Name:        "my_tool",
           Description: "Description of the tool",
           InputSchema: json.RawMessage(`{...}`),
           Annotations: api.ReadOnly("My Tool"),  // or api.Destructive(), api.Idempotent()
       },
       Handler: myToolHandler,
   }
   ```
3. Implement the handler function following existing patterns
4. Run `make build` to verify

### Authentication

The server reads credentials from environment variables at startup:
- `IONOS_USERNAME` + `IONOS_PASSWORD` for username/password auth
- `IONOS_TOKEN` for token-based auth

### CLI Options

```bash
./ionoscloud-mcp --help              # Show help
./ionoscloud-mcp --version           # Show version
./ionoscloud-mcp list-toolsets       # List available toolsets
./ionoscloud-mcp list-tools          # List all tools with annotations
./ionoscloud-mcp --read-only         # Only expose read-only tools
./ionoscloud-mcp --toolsets=compute  # Enable only specific toolsets
./ionoscloud-mcp --logging           # Enable request logging to stderr
./ionoscloud-mcp --metrics           # Enable timing metrics
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `IONOS_USERNAME` | IONOS Cloud username |
| `IONOS_PASSWORD` | IONOS Cloud password |
| `IONOS_TOKEN` | IONOS Cloud API token (alternative to username/password) |
| `IONOS_MCP_LOGGING` | Enable logging (set to "true" or "1") |
| `IONOS_MCP_METRICS` | Enable metrics (set to "true" or "1") |
| `IONOS_MCP_READ_ONLY` | Enable read-only mode (set to "true" or "1") |
| `IONOS_MCP_TOOLSETS` | Comma-separated list of enabled toolsets |

## Testing MCP Protocol

```bash
# Test initialization
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | ./ionoscloud-mcp

# List available tools
echo '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}' | ./ionoscloud-mcp

# Call a tool (requires valid credentials)
echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"list_datacenters","arguments":{}}}' | ./ionoscloud-mcp
```

## Dependencies

- [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) - MCP Go SDK
- [ionos-cloud/sdk-go/v6](https://github.com/ionos-cloud/sdk-go) - IONOS Cloud Compute SDK
- [ionos-cloud/sdk-go-bundle](https://github.com/ionos-cloud/sdk-go-bundle) - IONOS Cloud SDK Bundle (DNS)
- [spf13/cobra](https://github.com/spf13/cobra) - CLI framework

## Future Enhancements

See TODO.md for planned CRUD operations. Additional improvements could include:

- [ ] OpenTelemetry integration for distributed tracing
- [ ] TOML configuration file support
- [ ] HTTP transport (in addition to stdio)
- [ ] Graceful shutdown with SIGTERM handling
- [ ] Additional toolsets (Object Storage, DBaaS)
