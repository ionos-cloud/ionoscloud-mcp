---
subcategory: "Compute Engine"
page_title: "NIC"
description: |-
  Tools for listing and inspecting network interfaces (NICs) in IONOS Cloud.
---

# NICs

## list_nics

Lists all network interfaces (NICs) attached to a specific server.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center |
| `server_id` | string | Yes | The ID of the server |

**Example:**

```json
{
  "name": "list_nics",
  "arguments": {
    "datacenter_id": "12345678-1234-1234-1234-123456789012",
    "server_id": "87654321-4321-4321-4321-210987654321"
  }
}
```

**API Reference:** [datacentersServersNicsGet](https://api.ionos.com/docs/cloud/v6/#tag/NetworkInterfaces/operation/datacentersServersNicsGet)

---

## get_nic

Gets detailed information about a specific network interface (NIC).

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center |
| `server_id` | string | Yes | The ID of the server |
| `nic_id` | string | Yes | The ID of the network interface |

**Example:**

```json
{
  "name": "get_nic",
  "arguments": {
    "datacenter_id": "12345678-1234-1234-1234-123456789012",
    "server_id": "87654321-4321-4321-4321-210987654321",
    "nic_id": "abcdef12-3456-7890-abcd-ef1234567890"
  }
}
```

**API Reference:** [datacentersServersNicsFindById](https://api.ionos.com/docs/cloud/v6/#tag/NetworkInterfaces/operation/datacentersServersNicsFindById)
