---
subcategory: "Compute Engine"
page_title: "IP Block"
description: |-
  Tools for listing, inspecting, and (opt-in) reserving and releasing public IP blocks in IONOS CLOUD.
---

# IP Blocks

The `list_*` and `get_*` tools are always available. The write tools register only when `IONOS_MCP_TOOL_SCOPE` opts in (`write` enables create; `destructive` also enables delete). `create_*` and `delete_*` use a two-phase confirmation: call once **without** `confirmation_token` to get a preview plus a one-time token, then call again **with** that token to execute.

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

---

## create_ip_block

Reserves a block of public IPv4 addresses. Requires `IONOS_MCP_TOOL_SCOPE` to include `write`. Two-phase.

An IP block belongs to the **account**, not to a data center — but its `location` must match the data center whose resources will use the addresses. Both `location` and `size` are **fixed once reserved**: to change either, reserve a new block and release this one. The block is billed from the moment it exists, whether or not its addresses are assigned to anything.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `location` | string | Yes | Where to reserve the addresses, e.g. `de/fra`. Must match the data center that will use them. Immutable. |
| `size` | integer | Yes | How many addresses to reserve. Immutable. |
| `name` | string | No | A name for the block — the only property you can change later. |
| `confirmation_token` | string | No | Omit on the first call for a preview + token; pass it (with the same `location` and `size`) on the second call. |

**API Reference:** [ipblocksPost](https://api.ionos.com/docs/cloud/v6/#tag/IP-Blocks/operation/ipblocksPost)


---

## delete_ip_block

Releases a block of public IPv4 addresses. Irreversible. Requires `IONOS_MCP_TOOL_SCOPE` to include `destructive`. Two-phase.

The preview lists every resource still using the addresses — servers, NICs and Kubernetes node pools — because releasing a block whose addresses are assigned breaks connectivity for exactly those resources. A block is requested by location and size only, so there is **no way to ask for the same addresses back**.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `ipblock_id` | string | Yes | The ID of the block to release. |
| `confirmation_token` | string | No | Omit on the first call for the in-use preview + token; pass it on the second call to release. |

**API Reference:** [ipblocksDelete](https://api.ionos.com/docs/cloud/v6/#tag/IP-Blocks/operation/ipblocksDelete)
