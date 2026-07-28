---
subcategory: "Compute Engine"
page_title: "Private Cross-Connect"
description: |-
  Tools for listing, inspecting, and (opt-in) creating, updating and deleting private cross connects in IONOS CLOUD.
---

# Private Cross-Connects

The `list_*` and `get_*` tools are always available. The write tools register only when `IONOS_MCP_TOOL_SCOPE` opts in (`write` enables create/update; `destructive` also enables delete). `create_*` and `delete_*` use a two-phase confirmation: call once **without** `confirmation_token` to get a preview plus a one-time token, then call again **with** that token to execute.

## list_private_cross_connects

Lists all private cross-connects in your IONOS CLOUD account.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `depth` | integer | No | Nesting depth of returned objects (0–5). |
| `filters` | object | No | Server-side property filters; e.g. `{"name":"prod"}`. If the result is empty, retry without filters — a filter typo or mismatch silently returns nothing. |

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
| `depth` | integer | No | Nesting depth of returned objects (0–5). |

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

---

## create_pcc / update_pcc / delete_pcc

A private cross connect links private LANs together — including across data centers — without traffic leaving the IONOS network. It is account-level, so these tools take no `datacenter_id`.

**Which LANs are peered is controlled from the LAN side**, with `update_lan`'s `pcc` field — not here. These tools manage only the cross connect's own `name` and `description`.

A new cross connect connects nothing until you attach LANs to it. Every LAN on one cross connect must use non-overlapping IP ranges within the same subnet, so plan addressing before attaching the second LAN.

| Tool | Scope | Confirmation | Parameters |
|------|-------|--------------|------------|
| `create_pcc` | `write` | two-phase | `name` (required), `description` |
| `update_pcc` | `write` | none | `pcc_id` (required), `name`, `description` |
| `delete_pcc` | `destructive` | two-phase | `pcc_id` (required), `confirmation_token` |

`delete_pcc`'s preview names the peered LANs and their data centers, so you can see which private connections break. The LANs themselves survive.

**API Reference:** [pccsPost](https://api.ionos.com/docs/cloud/v6/#tag/Private-Cross-Connects/operation/pccsPost), [pccsPatch](https://api.ionos.com/docs/cloud/v6/#tag/Private-Cross-Connects/operation/pccsPatch), [pccsDelete](https://api.ionos.com/docs/cloud/v6/#tag/Private-Cross-Connects/operation/pccsDelete)
