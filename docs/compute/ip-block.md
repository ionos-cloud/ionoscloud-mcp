---
subcategory: "Compute Engine"
page_title: "IP Block"
description: |-
  Tools for listing and inspecting reserved IP blocks in IONOS CLOUD.
---

# IP Blocks

## list_ip_blocks

Lists all reserved IP blocks in your IONOS CLOUD account.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `depth` | integer | No | Nesting depth of returned objects (0–5). |
| `filters` | object | No | Server-side property filters; e.g. `{"name":"prod","location":"de/fra"}`. If the result is empty, retry without filters — a filter typo or mismatch silently returns nothing. |

**Example:**

```json
{
  "name": "list_ip_blocks",
  "arguments": {}
}
```

**API Reference:** [ipblocksGet](https://api.ionos.com/docs/cloud/v6/#tag/IPBlocks/operation/ipblocksGet)

---

## get_ip_block

Gets detailed information about a specific reserved IP block.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `ipblock_id` | string | Yes | The ID of the IP block |
| `depth` | integer | No | Nesting depth of returned objects (0–5). |

**Example:**

```json
{
  "name": "get_ip_block",
  "arguments": {
    "ipblock_id": "12345678-1234-1234-1234-123456789012"
  }
}
```

**API Reference:** [ipblocksFindById](https://api.ionos.com/docs/cloud/v6/#tag/IPBlocks/operation/ipblocksFindById)
