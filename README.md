# IONOS Cloud MCP Server

This project implements a Model Context Protocol (MCP) server that allows LLMs to interact with IONOS Cloud resources. The server is written in Go and uses the official IONOS Cloud SDK.

## Features

The server provides the following tools for interacting with IONOS Cloud:

- **list_datacenters**: List all virtual data centers in your IONOS Cloud account
- **get_datacenter**: Get details of a specific virtual data center
- **list_servers**: List all servers in a data center
- **get_server**: Get details of a specific server
- **list_volumes**: List all volumes in a data center
- **get_volume**: Get details of a specific volume

## Prerequisites

- Go 1.16 or higher
- IONOS Cloud account with API credentials

## Installation

1. Clone the repository:
```bash
git clone https://github.com/ionos-cloud/ionoscloud-mcp.git
cd ionoscloud-mcp
```

2. Build the server:
```bash
go build -o ionoscloud-mcp .
```

## Configuration

The server requires IONOS Cloud API credentials to be set as environment variables. You can use either:

- Username and password authentication:
  ```bash
  export IONOS_USERNAME="your-username"
  export IONOS_PASSWORD="your-password"
  ```

- Token-based authentication:
  ```bash
  export IONOS_TOKEN="your-api-token"
  ```

You can obtain your API credentials from the [IONOS Cloud DCD](https://dcd.ionos.com/).

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
        "IONOS_USERNAME": "your-username",
        "IONOS_PASSWORD": "your-password"
      }
    }
  }
}
```

Or with token authentication:

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

## Available Tools

### list_datacenters

Lists all virtual data centers in your IONOS Cloud account.

**Parameters:** None

**Example:**
```json
{
  "name": "list_datacenters",
  "arguments": {}
}
```

### get_datacenter

Gets detailed information about a specific data center.

**Parameters:**
- `datacenter_id` (string, required): The ID of the data center

**Example:**
```json
{
  "name": "get_datacenter",
  "arguments": {
    "datacenter_id": "12345678-1234-1234-1234-123456789012"
  }
}
```

### list_servers

Lists all servers in a specific data center.

**Parameters:**
- `datacenter_id` (string, required): The ID of the data center

**Example:**
```json
{
  "name": "list_servers",
  "arguments": {
    "datacenter_id": "12345678-1234-1234-1234-123456789012"
  }
}
```

### get_server

Gets detailed information about a specific server.

**Parameters:**
- `datacenter_id` (string, required): The ID of the data center
- `server_id` (string, required): The ID of the server

**Example:**
```json
{
  "name": "get_server",
  "arguments": {
    "datacenter_id": "12345678-1234-1234-1234-123456789012",
    "server_id": "87654321-4321-4321-4321-210987654321"
  }
}
```

### list_volumes

Lists all volumes in a specific data center.

**Parameters:**
- `datacenter_id` (string, required): The ID of the data center

**Example:**
```json
{
  "name": "list_volumes",
  "arguments": {
    "datacenter_id": "12345678-1234-1234-1234-123456789012"
  }
}
```

### get_volume

Gets detailed information about a specific volume.

**Parameters:**
- `datacenter_id` (string, required): The ID of the data center
- `volume_id` (string, required): The ID of the volume

**Example:**
```json
{
  "name": "get_volume",
  "arguments": {
    "datacenter_id": "12345678-1234-1234-1234-123456789012",
    "volume_id": "11111111-1111-1111-1111-111111111111"
  }
}
```

## Development

### Building from Source

```bash
go build -o ionoscloud-mcp .
```

### Dependencies

This project uses minimal external dependencies:
- [ionos-cloud/sdk-go](https://github.com/ionos-cloud/sdk-go) - Official IONOS Cloud Go SDK

## API Documentation

For more information about the IONOS Cloud API, refer to:
- [IONOS Cloud API Documentation](https://api.ionos.com/docs/)
- [API Specifications](https://github.com/ionos-cloud/rest-api/tree/main/public)
- [SDK Documentation](https://github.com/ionos-cloud/sdk-go)

## License

Apache License 2.0 - See LICENSE file for details.
