---
subcategory: "Compute Engine"
page_title: "LAN"
description: |-
  Tools for listing and inspecting LANs in IONOS CLOUD.
---

# LANs

## list_lans

Lists all LANs in a specific data center.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center |
| `depth` | integer | No | Nesting depth of returned objects (0–5, default `1`). |
| `filters` | object | No | Server-side property filters; e.g. `{"name":"prod"}`. If the result is empty, retry without filters — a filter typo or mismatch silently returns nothing. |

**Example:**

```json
{
  "name": "list_lans",
  "arguments": {
    "datacenter_id": "12345678-1234-1234-1234-123456789012"
  }
}
```

**API Reference:** [datacentersLansGet](https://api.ionos.com/docs/cloud/v6/#tag/LANs/operation/datacentersLansGet)

---

## get_lan

Gets detailed information about a specific LAN.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center |
| `lan_id` | string | Yes | The ID of the LAN |
| `depth` | integer | No | Nesting depth of returned objects (0–5). |

**Example:**

```json
{
  "name": "get_lan",
  "arguments": {
    "datacenter_id": "12345678-1234-1234-1234-123456789012",
    "lan_id": "1"
  }
}
```

**API Reference:** [datacentersLansFindById](https://api.ionos.com/docs/cloud/v6/#tag/LANs/operation/datacentersLansFindById)

---

## list_lan_nics

Lists all NICs attached to a specific LAN.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center |
| `lan_id` | string | Yes | The ID of the LAN |
| `depth` | integer | No | Nesting depth of returned objects (0–5). |
| `filters` | object | No | Server-side property filters; e.g. `{"name":"prod"}`. If the result is empty, retry without filters — a filter typo or mismatch silently returns nothing. |

**Example:**

```json
{
  "name": "list_lan_nics",
  "arguments": {
    "datacenter_id": "12345678-1234-1234-1234-123456789012",
    "lan_id": "1"
  }
}
```

**API Reference:** [datacentersLansNicsGet](https://api.ionos.com/docs/cloud/v6/#tag/LANs/operation/datacentersLansNicsGet)
