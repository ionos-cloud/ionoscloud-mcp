---
subcategory: "Compute Engine"
page_title: "Data Center"
description: |-
  Tools for listing and inspecting virtual data centers in IONOS CLOUD.
---

# Data Centers

## list_datacenters

Lists all virtual data centers in your IONOS CLOUD account. Returns names and basic properties by default (`depth=1`), so datacenter names are available without follow-up calls.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `depth` | integer | No | Nesting depth of returned objects (0–5, default `1`). Depth 1 includes datacenter name, description, location, and version. |
| `name` | string | No | Filter datacenters by name. Forwarded as `filter.properties.name` to the API (server-side match). |

**Examples:**

List all datacenters with names included (default):

```json
{
  "name": "list_datacenters",
  "arguments": {}
}
```

Find a datacenter by name:

```json
{
  "name": "list_datacenters",
  "arguments": {
    "name": "production-dc"
  }
}
```

**API Reference:** [datacentersGet](https://api.ionos.com/docs/cloud/v6/#tag/Data-Centers/operation/datacentersGet)

---

## get_datacenter

Gets detailed information about a specific virtual data center.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center |

**Example:**

```json
{
  "name": "get_datacenter",
  "arguments": {
    "datacenter_id": "12345678-1234-1234-1234-123456789012"
  }
}
```

**API Reference:** [datacentersFindById](https://api.ionos.com/docs/cloud/v6/#tag/Data-Centers/operation/datacentersFindById)
