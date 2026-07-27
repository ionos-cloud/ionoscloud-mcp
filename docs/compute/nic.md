---
subcategory: "Compute Engine"
page_title: "NIC"
description: |-
  Tools for listing, inspecting, and (opt-in) creating, updating, and deleting network interfaces in IONOS CLOUD.
---

# NICs

The `list_*` and `get_*` tools are always available. The write tools — `create_nic`, `update_nic`, and `delete_nic` — register only when `IONOS_MCP_TOOL_SCOPE` opts in (`write` enables create/update; `destructive` also enables delete). `create_nic` and `delete_nic` use a two-phase confirmation: call once **without** `confirmation_token` to get a preview plus a one-time token, then call again **with** that token to execute.

## list_nics

Lists all network interfaces (NICs) attached to a specific server.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center |
| `server_id` | string | Yes | The ID of the server |
| `depth` | integer | No | Nesting depth of returned objects (0–5, default `1`). |
| `filters` | object | No | Server-side property filters; e.g. `{"name":"prod"}`. If the result is empty, retry without filters — a filter typo or mismatch silently returns nothing. |

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
| `depth` | integer | No | Nesting depth of returned objects (0–5). |

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

---

## create_nic

Creates one network interface on a server and connects it to a LAN. Requires `IONOS_MCP_TOOL_SCOPE` to include `write`. Two-phase: the first call (no `confirmation_token`) returns a preview and a one-time token; a second call with that token creates exactly one NIC.

`lan` is the LAN's **small numeric ID** from `list_lans`, not a UUID. If no LAN with that ID exists the API creates one implicitly, so a wrong number silently puts the server on a new isolated LAN instead of the intended one — check `list_lans` first.

Setting `firewall_active` to `true` on a brand-new NIC blocks **all** incoming traffic until you add rules with `create_firewall_rule`; the preview warns about this.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center. |
| `server_id` | string | Yes | The ID of the server to attach the NIC to. |
| `lan` | integer | Yes | Numeric LAN ID from `list_lans`. A non-existent ID implicitly creates a new LAN. |
| `name` | string | No | The name of the new NIC. |
| `ips` | array of string | No | IPv4 addresses to assign. Public IPs must come from a reserved IP block (`list_ip_blocks`). Omit for automatic assignment. |
| `dhcp` | boolean | No | Whether the NIC reserves an IP using DHCP (API default `true`). |
| `firewall_active` | boolean | No | Activate the firewall. **With no rules defined this blocks all incoming traffic.** |
| `firewall_type` | string | No | `INGRESS` (default), `EGRESS` or `BIDIRECTIONAL`. |
| `confirmation_token` | string | No | Omit on the first call to receive a preview + token; pass the token (with the same `datacenter_id`, `server_id` and `lan`) on the second call to create. Expires after a few minutes. |

**Example (step 1 — preview):**

```json
{
  "name": "create_nic",
  "arguments": { "datacenter_id": "12345678-1234-1234-1234-123456789012", "server_id": "87654321-4321-4321-4321-210987654321", "lan": 1, "name": "eth0" }
}
```

**API Reference:** [datacentersServersNicsPost](https://api.ionos.com/docs/cloud/v6/#tag/Network-Interfaces/operation/datacentersServersNicsPost)

---

## update_nic

Updates a NIC's name, LAN, IP addresses or firewall settings. Requires `IONOS_MCP_TOOL_SCOPE` to include `write`. Applies a partial update (only the fields you provide).

**Omit `lan` to leave the NIC where it is.** The IONOS SDK always serializes a NIC's `lan` field, so this tool reads the NIC's current LAN and sends it back unchanged when you do not pass one — that read is why `update_nic` issues a `GET` before its `PATCH`. Pass `lan` explicitly only when you actually want to move the NIC.

Setting `ips` **replaces** the whole address list, so include every address the NIC should keep.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center. |
| `server_id` | string | Yes | The ID of the server the NIC belongs to. |
| `nic_id` | string | Yes | The ID of the NIC to update. |
| `name` | string | No | A new name. |
| `lan` | integer | No | Move the NIC to this LAN ID. Omit to keep it on its current LAN. |
| `ips` | array of string | No | **Replaces** the IPv4 address list. |
| `dhcp` | boolean | No | Whether the NIC reserves an IP using DHCP. |
| `firewall_active` | boolean | No | Activate or deactivate the firewall. Activating with no rules blocks all incoming traffic. |
| `firewall_type` | string | No | `INGRESS`, `EGRESS` or `BIDIRECTIONAL`. |

**Example:**

```json
{
  "name": "update_nic",
  "arguments": {
    "datacenter_id": "12345678-1234-1234-1234-123456789012",
    "server_id": "87654321-4321-4321-4321-210987654321",
    "nic_id": "11112222-3333-4444-5555-666677778888",
    "name": "eth0-renamed"
  }
}
```

**API Reference:** [datacentersServersNicsPatch](https://api.ionos.com/docs/cloud/v6/#tag/Network-Interfaces/operation/datacentersServersNicsPatch)

---

## delete_nic

Deletes a NIC along with its firewall rules and flow logs. Irreversible. Requires `IONOS_MCP_TOOL_SCOPE` to include `destructive`. Two-phase: the first call (no `confirmation_token`) returns a blast-radius preview and a one-time token; a second call with that token performs the deletion.

The server loses its connectivity on that LAN, and any public IP the NIC held is released back to its IP block.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center. |
| `server_id` | string | Yes | The ID of the server the NIC belongs to. |
| `nic_id` | string | Yes | The ID of the NIC to delete. |
| `confirmation_token` | string | No | Omit on the first call to receive a blast-radius preview + token; pass the token on the second call to delete. Bound to this NIC **and its parent server**; expires after a few minutes. |

**Example (step 1 — preview):**

```json
{
  "name": "delete_nic",
  "arguments": { "datacenter_id": "12345678-1234-1234-1234-123456789012", "server_id": "87654321-4321-4321-4321-210987654321", "nic_id": "11112222-3333-4444-5555-666677778888" }
}
```

**API Reference:** [datacentersServersNicsDelete](https://api.ionos.com/docs/cloud/v6/#tag/Network-Interfaces/operation/datacentersServersNicsDelete)
