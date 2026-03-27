# IONOS Cloud MCP Server

This project implements a Model Context Protocol (MCP) server that allows LLMs to interact with IONOS Cloud resources. The server is written in Go and uses the official IONOS Cloud SDK.

## What is MCP?

The [Model Context Protocol](https://modelcontextprotocol.io/) (MCP) is an open standard that allows AI assistants to connect to external tools and data sources. It defines a JSON-RPC 2.0 interface over stdio (or HTTP) through which an LLM client can discover and invoke tools provided by a server. This MCP server exposes IONOS Cloud infrastructure operations as tools, enabling AI assistants like Claude to list, inspect, and manage your cloud resources through natural language. It is designed for developers and platform engineers who want to interact with IONOS Cloud programmatically through an AI-powered workflow.

## Features

The server provides the following tools for interacting with IONOS Cloud:

### [Compute Engine](docs/compute/)

| Resource | Tools | Docs |
|----------|-------|------|
| Data Center | `list_datacenters`, `get_datacenter` | [datacenter.md](docs/compute/datacenter.md) |
| Server | `list_servers`, `get_server` | [server.md](docs/compute/server.md) |
| Server Sub-Resources | `list_server_volumes`, `list_server_cdroms`, `list_server_gpus`, `get_server_gpu`, `get_server_remote_console` | [server-subresources.md](docs/compute/server-subresources.md) |
| Volume | `list_volumes`, `get_volume` | [volume.md](docs/compute/volume.md) |
| NIC | `list_nics`, `get_nic` | [nic.md](docs/compute/nic.md) |
| LAN | `list_lans`, `get_lan`, `list_lan_nics` | [lan.md](docs/compute/lan.md) |
| Firewall Rule | `list_firewall_rules`, `get_firewall_rule` | [firewall-rule.md](docs/compute/firewall-rule.md) |
| IP Block | `list_ip_blocks`, `get_ip_block` | [ip-block.md](docs/compute/ip-block.md) |
| Load Balancer | `list_loadbalancers`, `get_loadbalancer`, `list_loadbalancer_nics` | [loadbalancer.md](docs/compute/loadbalancer.md) |
| Network Load Balancer | `list_network_loadbalancers`, `get_network_loadbalancer`, `list_nlb_forwarding_rules` | [network-loadbalancer.md](docs/compute/network-loadbalancer.md) |
| Application Load Balancer | `list_application_loadbalancers`, `get_application_loadbalancer`, `list_alb_forwarding_rules` | [application-loadbalancer.md](docs/compute/application-loadbalancer.md) |
| Target Group | `list_target_groups`, `get_target_group` | [target-group.md](docs/compute/target-group.md) |
| NAT Gateway | `list_nat_gateways`, `get_nat_gateway`, `list_nat_gateway_rules` | [nat-gateway.md](docs/compute/nat-gateway.md) |
| Private Cross-Connect | `list_private_cross_connects`, `get_private_cross_connect` | [private-cross-connect.md](docs/compute/private-cross-connect.md) |
| Security Group | `list_security_groups`, `get_security_group`, `list_security_group_rules`, `get_security_group_rule` | [security-group.md](docs/compute/security-group.md) |
| Image | `list_images` | [image.md](docs/compute/image.md) |
| Location | `list_locations` | [location.md](docs/compute/location.md) |
| Snapshot | `list_snapshots`, `get_snapshot` | [snapshot.md](docs/compute/snapshot.md) |
| Contract | `get_contract` | [contract.md](docs/compute/contract.md) |
| Request | `list_requests`, `get_request`, `get_request_status` | [request.md](docs/compute/request.md) |
| Template | `list_templates`, `get_template` | [template.md](docs/compute/template.md) |

## Installation

1. Clone the repository:
```bash
git clone https://github.com/ionos-cloud/ionoscloud-mcp.git
cd ionoscloud-mcp
```

2. Build the server:
```bash
make build
# or
go build -o ionoscloud-mcp .
```

## Configuration

You need an IONOS Cloud account with API credentials. Set the token as an environment variable:

```bash
# Token-based authentication
export IONOS_TOKEN="your-api-token"
```

You can generate a token from the [IONOS Cloud DCD](https://dcd.ionos.com/) under Management > Token Management.

## Usage

The server uses stdio for communication following the MCP protocol. To run the server:

```bash
./ionoscloud-mcp
```

### Integration with MCP Clients

To use this server with an MCP client (like Claude Desktop), add it to your MCP settings:

```json
{
  "mcpServers": {
    "ionoscloud": {
      "command": "/path/to/ionoscloud-mcp",
      "env": {
        "IONOS_TOKEN": "your-api-token"
      }
    }
  }
}
```

## Development

### Testing the MCP Protocol

You can test the server's MCP protocol implementation using stdin/stdout:

```bash
# Initialize and list tools
{
  echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"0.1.0"}}}'
  echo '{"jsonrpc":"2.0","method":"notifications/initialized"}'
  echo '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}'
  sleep 1
} | ./ionoscloud-mcp

# Call a tool (requires valid IONOS_TOKEN)
{
  echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"0.1.0"}}}'
  echo '{"jsonrpc":"2.0","method":"notifications/initialized"}'
  echo '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"list_datacenters","arguments":{}}}'
  sleep 1
} | ./ionoscloud-mcp
```

### Building from Source

```bash
make build
# or
go build -o ionoscloud-mcp .
```

### Available Make Targets

- `make build` - Build the binary
- `make clean` - Remove build artifacts
- `make fmt` - Format code
- `make vet` - Run go vet
- `make check` - Run fmt and vet
- `make deps` - Download and tidy dependencies

### Dependencies

This project uses minimal external dependencies:
- [ionos-cloud/sdk-go-bundle](https://github.com/ionos-cloud/sdk-go-bundle) - IONOS Cloud Go SDK Bundle

## API Documentation

For more information about the IONOS Cloud API, refer to:
- [IONOS Cloud API Documentation](https://api.ionos.com/docs/)
- [API Specifications](https://github.com/ionos-cloud/rest-api/tree/main/public)
- [SDK Documentation](https://github.com/ionos-cloud/sdk-go-bundle)

## License

Apache License 2.0 - See LICENSE file for details.
