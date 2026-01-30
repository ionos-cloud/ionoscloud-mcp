# IONOS Cloud MCP Server

This project implements a Model Context Protocol (MCP) server that allows LLMs to interact with IONOS Cloud resources. The server is written in Go and uses the official IONOS Cloud SDK.

## Features

- **93 tools** for comprehensive IONOS Cloud infrastructure management
- **6 toolsets**: Compute, Networking, Load Balancing, Kubernetes, IAM, DNS
- **Tool annotations** for safety (read-only, destructive hints)
- **CLI options** for toolset filtering and read-only mode
- **Modular architecture** with self-registering toolsets

### Available Toolsets

| Toolset | Tools | Resources |
|---------|-------|-----------|
| compute | 28 | Datacenters, Servers, Volumes, Snapshots, Images, Locations |
| networking | 35 | LANs, NICs, IP Blocks, Firewall Rules, NAT Gateways, PCCs |
| loadbalancing | 8 | Application Load Balancers, Network Load Balancers, Target Groups |
| kubernetes | 8 | K8s Clusters, Node Pools, Nodes, Kubeconfig, Versions |
| iam | 10 | Users, Groups, S3 Keys, Contract, Resources |
| dns | 4 | DNS Zones, DNS Records |

## Prerequisites

- Go 1.20 or higher (tested with Go 1.24)
- IONOS Cloud account with API credentials

## Installation

1. Clone the repository:
```bash
git clone https://github.com/ionos-cloud/ionoscloud-mcp.git
cd ionoscloud-mcp
```

2. Build the server:
```bash
make build
```

## Configuration

The server requires IONOS Cloud API credentials to be set as environment variables:

```bash
# Username and password authentication
export IONOS_USERNAME="your-username"
export IONOS_PASSWORD="your-password"

# Or token-based authentication
export IONOS_TOKEN="your-api-token"
```

You can obtain your API credentials from the [IONOS Cloud DCD](https://dcd.ionos.com/).

### Optional Environment Variables

| Variable | Description |
|----------|-------------|
| `IONOS_MCP_LOGGING` | Enable request logging (`true` or `1`) |
| `IONOS_MCP_METRICS` | Enable timing metrics (`true` or `1`) |
| `IONOS_MCP_READ_ONLY` | Only expose read-only tools (`true` or `1`) |
| `IONOS_MCP_TOOLSETS` | Comma-separated list of enabled toolsets |

## Usage

### Running the Server

The server uses stdio for communication following the MCP protocol:

```bash
./ionoscloud-mcp
```

### CLI Options

```bash
./ionoscloud-mcp --help              # Show help
./ionoscloud-mcp --version           # Show version
./ionoscloud-mcp list-toolsets       # List available toolsets
./ionoscloud-mcp list-tools          # List all tools with annotations
./ionoscloud-mcp --read-only         # Only expose read-only tools (54 tools)
./ionoscloud-mcp --toolsets=compute,dns  # Enable only specific toolsets
./ionoscloud-mcp --logging           # Enable request logging to stderr
./ionoscloud-mcp --metrics           # Enable timing metrics
```

### Integration with MCP Clients

To use this server with an MCP client (like Claude Desktop), add it to your MCP settings:

```json
{
  "mcpServers": {
    "ionoscloud": {
      "command": "/path/to/ionoscloud-mcp",
      "env": {
        "IONOS_USERNAME": "your-username",
        "IONOS_PASSWORD": "your-password"
      }
    }
  }
}
```

For read-only mode (recommended for safety):

```json
{
  "mcpServers": {
    "ionoscloud": {
      "command": "/path/to/ionoscloud-mcp",
      "args": ["--read-only"],
      "env": {
        "IONOS_TOKEN": "your-api-token"
      }
    }
  }
}
```

## Tool Categories

### Read-Only Tools (54 tools)
Safe operations that only read data: `list_*`, `get_*`

### Destructive Tools (14 tools)
Operations that can delete or destroy resources: `delete_*`, `stop_server`, `reboot_server`, `restore_snapshot`

### Write Tools (25 tools)
Operations that create or modify resources: `create_*`, `update_*`, `attach_*`, `detach_*`, `start_server`

## Development

### Testing the MCP Protocol

```bash
# Test initialization
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | ./ionoscloud-mcp

# List available tools
echo '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}' | ./ionoscloud-mcp

# Call a tool (requires valid credentials)
echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"list_datacenters","arguments":{}}}' | ./ionoscloud-mcp
```

### Available Make Targets

```bash
make build          # Build the binary
make test           # Run tests
make clean          # Remove build artifacts
make fmt            # Format code
make vet            # Run go vet
make check          # Run fmt and vet
make deps           # Download and tidy dependencies
make list-toolsets  # List available toolsets
make list-tools     # List all available tools
```

### Project Structure

```
cmd/ionoscloud-mcp/     # CLI entry point
pkg/
  api/                  # Core interfaces and types
  config/               # Configuration management
  ionos/                # IONOS Cloud client wrapper
  mcp/                  # MCP server implementation
  toolsets/             # Tool implementations by category
    compute/            # Datacenters, Servers, Volumes, etc.
    networking/         # LANs, NICs, Firewall Rules, etc.
    loadbalancing/      # ALB, NLB, Target Groups
    kubernetes/         # K8s Clusters, Node Pools
    iam/                # Users, Groups, S3 Keys
    dns/                # DNS Zones, Records
```

## Dependencies

- [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) - MCP Go SDK
- [ionos-cloud/sdk-go/v6](https://github.com/ionos-cloud/sdk-go) - IONOS Cloud Compute SDK
- [ionos-cloud/sdk-go-bundle](https://github.com/ionos-cloud/sdk-go-bundle) - IONOS Cloud SDK Bundle (DNS)
- [spf13/cobra](https://github.com/spf13/cobra) - CLI framework

## API Documentation

For more information about the IONOS Cloud API, refer to:
- [IONOS Cloud API Documentation](https://api.ionos.com/docs/)
- [API Specifications](https://github.com/ionos-cloud/rest-api/tree/main/public)
- [SDK Documentation](https://github.com/ionos-cloud/sdk-go)

## License

Apache License 2.0 - See LICENSE file for details.
