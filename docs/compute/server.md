---
subcategory: "Compute Engine"
page_title: "Server"
description: |-
  Tools for listing and inspecting servers in IONOS CLOUD.
---

# Servers

## list_servers

Lists all servers in a specific data center.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center |
| `depth` | integer | No | Nesting depth of returned objects (0–5, default `1`). |

**Example:**

```json
{
  "name": "list_servers",
  "arguments": {
    "datacenter_id": "12345678-1234-1234-1234-123456789012"
  }
}
```

**API Reference:** [datacentersServersGet](https://api.ionos.com/docs/cloud/v6/#tag/Servers/operation/datacentersServersGet)

---

## get_server

Gets detailed information about a specific server.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center |
| `server_id` | string | Yes | The ID of the server |
| `depth` | integer | No | Nesting depth of returned objects (0–5). |

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

**API Reference:** [datacentersServersFindById](https://api.ionos.com/docs/cloud/v6/#tag/Servers/operation/datacentersServersFindById)
