---
subcategory: "Compute Engine"
page_title: "LAN"
description: |-
  Tools for listing, inspecting, and (opt-in) creating, updating, and deleting LANs in IONOS CLOUD.
---

# LANs

The `list_*` and `get_*` tools are always available. The write tools — `create_lan`, `update_lan`, and `delete_lan` — register only when `IONOS_MCP_TOOL_SCOPE` opts in (`write` enables create/update; `destructive` also enables delete). `create_lan` and `delete_lan` use a two-phase confirmation: call once **without** `confirmation_token` to get a preview plus a one-time token, then call again **with** that token to execute.

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

---

## create_lan

Creates one LAN. Requires `IONOS_MCP_TOOL_SCOPE` to include `write`. Two-phase: the first call (no `confirmation_token`) returns a preview and a one-time token; a second call with that token creates exactly one LAN.

A **public** LAN is how servers reach the internet; a **private** LAN carries internal traffic only. The API assigns the LAN a small numeric ID — that number is the value you pass as `lan` to `create_nic`.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center to create the LAN in. |
| `name` | string | No | The name of the new LAN. |
| `public` | boolean | No | Whether the LAN is connected to the internet (default `false`). |
| `pcc` | string | No | ID of a private cross connect to attach the LAN to. All LANs on one cross connect must use non-overlapping ranges in the same subnet. |
| `ipv6_cidr_block` | string | No | `AUTO` for an automatically assigned `/64`, or an explicit `/64` inside the data center's IPv6 range and unique among its LANs. |
| `confirmation_token` | string | No | Omit on the first call to receive a preview + token; pass the token (with the same `datacenter_id` and `name`) on the second call to create. Expires after a few minutes. |

**Example (step 1 — preview):**

```json
{
  "name": "create_lan",
  "arguments": { "datacenter_id": "12345678-1234-1234-1234-123456789012", "name": "public-lan", "public": true }
}
```

**API Reference:** [datacentersLansPost](https://api.ionos.com/docs/cloud/v6/#tag/LANs/operation/datacentersLansPost)

---

## update_lan

Updates a LAN's name, public/private setting, cross connect or IPv6 block. Requires `IONOS_MCP_TOOL_SCOPE` to include `write`. Applies a partial update (only the fields you provide). Single call — no confirmation token.

Turning a public LAN private **removes internet access for every server connected to it**, and changing `ipv6_cidr_block` reassigns the `/80` blocks and addresses of every connected NIC. `ipv4_cidr_block` is read-only.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center the LAN is in. |
| `lan_id` | string | Yes | The ID of the LAN to update. |
| `name` | string | No | A new name. |
| `public` | boolean | No | Make the LAN public or private. |
| `pcc` | string | No | Attach to this cross connect ID, or pass an empty string to detach. |
| `ipv6_cidr_block` | string | No | A `/64` block, or `AUTO`. |

**Example:**

```json
{
  "name": "update_lan",
  "arguments": {
    "datacenter_id": "12345678-1234-1234-1234-123456789012",
    "lan_id": "1",
    "name": "renamed-lan"
  }
}
```

**API Reference:** [datacentersLansPatch](https://api.ionos.com/docs/cloud/v6/#tag/LANs/operation/datacentersLansPatch)

---

## delete_lan

Deletes a LAN. Irreversible. Requires `IONOS_MCP_TOOL_SCOPE` to include `destructive`. Two-phase: the first call (no `confirmation_token`) returns a blast-radius preview and a one-time token; a second call with that token performs the deletion.

The NICs on the LAN are not deleted, but every one of them **loses its network connection** — the preview reports how many, so check that count before proceeding.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center the LAN is in. |
| `lan_id` | string | Yes | The ID of the LAN to delete. |
| `confirmation_token` | string | No | Omit on the first call to receive a blast-radius preview + token; pass the token on the second call to delete. Bound to this LAN; expires after a few minutes. |

**Example (step 1 — preview):**

```json
{
  "name": "delete_lan",
  "arguments": { "datacenter_id": "12345678-1234-1234-1234-123456789012", "lan_id": "1" }
}
```

**API Reference:** [datacentersLansDelete](https://api.ionos.com/docs/cloud/v6/#tag/LANs/operation/datacentersLansDelete)
