# IONOS Cloud MCP Server

This project implements a Model Context Protocol (MCP) server that allows LLMs to interact with IONOS Cloud resources. The server is written in Go and uses the official IONOS Cloud SDK.

## What is MCP?

The [Model Context Protocol](https://modelcontextprotocol.io/) (MCP) is an open standard that allows AI assistants to connect to external tools and data sources. It defines a JSON-RPC 2.0 interface over stdio (or HTTP) through which an LLM client can discover and invoke tools provided by a server. This MCP server exposes IONOS Cloud infrastructure operations as tools, enabling AI assistants like Claude to list, inspect, and manage your cloud resources through natural language. It is designed for developers and platform engineers who want to interact with IONOS Cloud programmatically through an AI-powered workflow.

## Supported Products

| Product | Tools | Resources | Docs |
|---------|-------|-----------|------|
| [Compute Engine](docs/compute/) | 50 | Data Centers, Servers, Volumes, NICs, LANs, Firewall Rules, IP Blocks, Load Balancers, NAT Gateways, Security Groups, and more | [docs/compute/](docs/compute/) |
| [DNS](docs/dns/) | 14 | Zones, Zone Files, Records, Reverse Records, Secondary Zones, DNSSEC, Quota | [docs/dns/](docs/dns/) |
| [Billing](docs/billing/) | 14 | Profile, Invoices, EVN (provisioning intervals), Traffic, Usage, Utilization, Product pricing catalog | [docs/billing/](docs/billing/) |
| [Object Storage](docs/objectstorage/) | 23 | Buckets, Bucket Configuration (CORS, encryption, lifecycle, policy, replication, tagging, versioning, Object Lock), Objects, Access Keys, Regions | [docs/objectstorage/](docs/objectstorage/) |
| [Certificate Manager](docs/cert/) | 6 | Certificates, Auto-Certificates, Providers | [docs/cert/](docs/cert/) |

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

You need an IONOS Cloud account with API credentials. Set the required environment variables:

```bash
# Required: API token for management/control-plane APIs (Compute, DNS, Billing, Certificate Manager, Object Storage Management)
export IONOS_TOKEN="your-api-token"

# Optional: S3 credentials for Object Storage data-plane operations
# Only required if using Object Storage tools (list/inspect buckets, objects, access keys, etc.)
export IONOS_S3_ACCESS_KEY="your-s3-access-key"
export IONOS_S3_SECRET_KEY="your-s3-secret-key"
```

You can generate a token from the [IONOS Cloud DCD](https://dcd.ionos.com/) under Management > Token Management. S3 credentials for Object Storage can be created in the same interface under Object Storage > Access Keys.

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
        "IONOS_TOKEN": "your-api-token",
        "IONOS_S3_ACCESS_KEY": "your-s3-access-key",
        "IONOS_S3_SECRET_KEY": "your-s3-secret-key"
      }
    }
  }
}
```

**Note:** `IONOS_TOKEN` is required. The Object Storage credentials are only needed if you plan to use Object Storage tools.

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
- `make test` - Run tests
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
