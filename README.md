# IONOS CLOUD MCP Server

This project implements a Model Context Protocol (MCP) server that allows LLMs to interact with IONOS CLOUD resources. The server is written in Go and uses the official IONOS CLOUD SDK.

## What is MCP?

The [Model Context Protocol](https://modelcontextprotocol.io/) (MCP) is an open standard that allows AI assistants to connect to external tools and data sources. It defines a JSON-RPC 2.0 interface over stdio (or HTTP) through which an LLM client can discover and invoke tools provided by a server. This MCP server exposes IONOS CLOUD infrastructure operations as tools, enabling AI assistants like Claude to list, inspect, and manage your cloud resources through natural language. It is designed for developers and platform engineers who want to interact with IONOS CLOUD programmatically through an AI-powered workflow.

## Supported Products

| Product | Tools | Resources | Docs |
|---------|-------|-----------|------|
| [Compute Engine](docs/compute/) | 50 | Data Centers, Servers, Volumes, NICs, LANs, Firewall Rules, IP Blocks, Load Balancers, NAT Gateways, Security Groups, and more | [docs/compute/](docs/compute/) |
| [DNS](docs/dns/) | 14 | Zones, Zone Files, Records, Reverse Records, Secondary Zones, DNSSEC, Quota | [docs/dns/](docs/dns/) |
| [Billing](docs/billing/) | 14 | Profile, Invoices, EVN (provisioning intervals), Traffic, Usage, Utilization, Product pricing catalog | [docs/billing/](docs/billing/) |
| [Object Storage](docs/objectstorage/) | 23 | Buckets, Bucket Configuration (CORS, encryption, lifecycle, policy, replication, tagging, versioning, Object Lock), Objects, Access Keys, Regions | [docs/objectstorage/](docs/objectstorage/) |
| [Certificate Manager](docs/cert/) | 6 | Certificates, Auto-Certificates, Providers | [docs/cert/](docs/cert/) |
| [Activity Log](docs/activitylog/) | 2 | Contracts, Events | [docs/activitylog/](docs/activitylog/) |

## Installation

Pick whichever fits your workflow.

### Homebrew (macOS, Linux)

```bash
brew install ionos-cloud/ionos-cloud/ionoscloud-mcp
```

### Docker (linux/amd64, linux/arm64)

```bash
docker pull ghcr.io/ionos-cloud/ionoscloud-mcp:latest
```

Run with the MCP stdio transport:

```bash
docker run -i --rm -e IONOS_TOKEN="$IONOS_TOKEN" ghcr.io/ionos-cloud/ionoscloud-mcp
```

### `go install`

```bash
go install github.com/ionos-cloud/ionoscloud-mcp@latest
```

### Pre-built binaries

Download the archive for your OS/arch from the [latest release](https://github.com/ionos-cloud/ionoscloud-mcp/releases/latest) (Linux, macOS, Windows × amd64, arm64).

### From source

```bash
git clone https://github.com/ionos-cloud/ionoscloud-mcp.git
cd ionoscloud-mcp
make build
```

## Configuration

You need an IONOS CLOUD account with API credentials. Set the required environment variables:

```bash
# Required: API token for management/control-plane APIs (Compute, DNS, Billing, Certificate Manager, Object Storage Management)
export IONOS_TOKEN="your-api-token"

# Optional: S3 credentials for Object Storage data-plane operations
# Only required if using Object Storage tools (list/inspect buckets, objects, access keys, etc.)
export IONOS_S3_ACCESS_KEY="your-s3-access-key"
export IONOS_S3_SECRET_KEY="your-s3-secret-key"
```

You can generate a token from the [IONOS CLOUD DCD](https://dcd.ionos.com/) under Management > Token Management. S3 credentials for Object Storage can be created in the same interface under Object Storage > Access Keys.

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

### Tool loading mode

The `IONOS_MCP_LOAD_MODE` environment variable selects how tools are exposed:

- **`eager`** (default): all tools register at startup. Recommended for Claude Code (which defers full schemas client-side via ToolSearch, paying ~1–3k tokens for names only) and the only working mode for clients that ignore `notifications/tools/list_changed` (Claude Desktop, claude.ai connectors, Claude in Chrome, Smithery scanner).

- **`lazy`**: Compute and Object Storage register only on demand. Two sentinel tools (`ionos_load_compute_tools`, `ionos_load_objectstorage_tools`) appear at startup; calling either registers the full product set and emits `notifications/tools/list_changed`. Use only if your MCP client honours that notification AND lacks client-side schema deferral — otherwise eager mode is cheaper.

Parsing is case-insensitive. Unknown values fall back to `eager`.

```json
{
  "mcpServers": {
    "ionoscloud": {
      "command": "/path/to/ionoscloud-mcp",
      "env": {
        "IONOS_TOKEN": "your-api-token",
        "IONOS_MCP_LOAD_MODE": "lazy"
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
- `make test` - Run tests
- `make clean` - Remove build artifacts
- `make fmt` - Format code
- `make vet` - Run go vet
- `make check` - Run fmt and vet
- `make deps` - Download and tidy dependencies

### Dependencies

This project uses minimal external dependencies:
- [ionos-cloud/sdk-go-bundle](https://github.com/ionos-cloud/sdk-go-bundle) - IONOS CLOUD Go SDK Bundle

## API Documentation

For more information about the IONOS CLOUD API, refer to:
- [IONOS CLOUD API Documentation](https://api.ionos.com/docs/)
- [API Specifications](https://github.com/ionos-cloud/rest-api/tree/main/public)
- [SDK Documentation](https://github.com/ionos-cloud/sdk-go-bundle)

## License

Apache License 2.0 - See LICENSE file for details.
