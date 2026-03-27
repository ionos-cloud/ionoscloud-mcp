---
subcategory: "Compute Engine"
page_title: "Private Cross-Connect"
description: |-
  Tools for listing and inspecting private cross-connects in IONOS Cloud.
---

# Private Cross-Connects

## list_private_cross_connects

Lists all private cross-connects in your IONOS Cloud account.

**Parameters:**

None.

**Example:**

```json
{
  "name": "list_private_cross_connects",
  "arguments": {}
}
```

**API Reference:** [pccsGet](https://api.ionos.com/docs/cloud/v6/#tag/PrivateCrossConnects/operation/pccsGet)

---

## get_private_cross_connect

Gets detailed information about a specific private cross-connect.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `pcc_id` | string | Yes | The ID of the private cross-connect |

**Example:**

```json
{
  "name": "get_private_cross_connect",
  "arguments": {
    "pcc_id": "12345678-1234-1234-1234-123456789012"
  }
}
```

**API Reference:** [pccsFindById](https://api.ionos.com/docs/cloud/v6/#tag/PrivateCrossConnects/operation/pccsFindById)
