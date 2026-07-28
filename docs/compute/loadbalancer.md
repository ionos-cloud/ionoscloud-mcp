---
subcategory: "Compute Engine"
page_title: "Load Balancer"
description: |-
  Tools for listing, inspecting, and (opt-in) creating, updating and deleting classic load balancers in IONOS CLOUD.
---

# Load Balancers

The `list_*` and `get_*` tools are always available. The write tools register only when `IONOS_MCP_TOOL_SCOPE` opts in (`write` enables create/update; `destructive` also enables delete). `create_*` and `delete_*` use a two-phase confirmation: call once **without** `confirmation_token` to get a preview plus a one-time token, then call again **with** that token to execute.

## list_loadbalancers

Lists all load balancers in a specific data center.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center |
| `depth` | integer | No | Nesting depth of returned objects (0–5, default `1`). |
| `filters` | object | No | Server-side property filters; e.g. `{"name":"prod"}`. If the result is empty, retry without filters — a filter typo or mismatch silently returns nothing. |

**Example:**

```json
{
  "name": "list_loadbalancers",
  "arguments": {
    "datacenter_id": "12345678-1234-1234-1234-123456789012"
  }
}
```

**API Reference:** [datacentersLoadbalancersGet](https://api.ionos.com/docs/cloud/v6/#tag/LoadBalancers/operation/datacentersLoadbalancersGet)

---

## get_loadbalancer

Gets detailed information about a specific load balancer.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center |
| `loadbalancer_id` | string | Yes | The ID of the load balancer |
| `depth` | integer | No | Nesting depth of returned objects (0–5). |

**Example:**

```json
{
  "name": "get_loadbalancer",
  "arguments": {
    "datacenter_id": "12345678-1234-1234-1234-123456789012",
    "loadbalancer_id": "87654321-4321-4321-4321-210987654321"
  }
}
```

**API Reference:** [datacentersLoadbalancersFindById](https://api.ionos.com/docs/cloud/v6/#tag/LoadBalancers/operation/datacentersLoadbalancersFindById)

---

## list_loadbalancer_nics

Lists all NICs balanced by a specific load balancer.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center |
| `loadbalancer_id` | string | Yes | The ID of the load balancer |
| `depth` | integer | No | Nesting depth of returned objects (0–5). |
| `filters` | object | No | Server-side property filters; e.g. `{"name":"prod"}`. If the result is empty, retry without filters — a filter typo or mismatch silently returns nothing. |

**Example:**

```json
{
  "name": "list_loadbalancer_nics",
  "arguments": {
    "datacenter_id": "12345678-1234-1234-1234-123456789012",
    "loadbalancer_id": "87654321-4321-4321-4321-210987654321"
  }
}
```

**API Reference:** [datacentersLoadbalancersBalancednicsGet](https://api.ionos.com/docs/cloud/v6/#tag/LoadBalancers/operation/datacentersLoadbalancersBalancednicsGet)

---

## create_loadbalancer / update_loadbalancer / delete_loadbalancer

The **classic** load balancer balances traffic across NICs attached directly to it. That is a different model from the network and application load balancers, which forward to IP targets or target groups defined in forwarding rules. For new work prefer [network-loadbalancer.md](network-loadbalancer.md) (TCP/UDP) or [application-loadbalancer.md](application-loadbalancer.md) (HTTP), which offer health checks and finer routing.

| Tool | Scope | Confirmation | Parameters |
|------|-------|--------------|------------|
| `create_loadbalancer` | `write` | two-phase | `datacenter_id`, `name` (required), `ip`, `dhcp` |
| `update_loadbalancer` | `write` | none | `datacenter_id`, `loadbalancer_id` (required), `name`, `ip`, `dhcp` |
| `delete_loadbalancer` | `destructive` | two-phase | `datacenter_id`, `loadbalancer_id` (required), `confirmation_token` |

This is the **only** load balancer whose properties are all optional in the API, so `update_loadbalancer` is a genuine partial `PATCH` with no carry-forward read — one request, not two.

`delete_loadbalancer`'s preview counts the NICs it balances across. They are not deleted, but traffic stops being balanced to them.

### Not available: balanced-NIC attach/detach

`attach_loadbalancer_nic` is deliberately not exposed. Attaching an existing NIC is expressed as a body carrying only its id, but `Nic.Properties` is a non-pointer field in the Go SDK whose serializer runs unconditionally, so the smallest body the SDK can produce is `{"id":"…","properties":{"lan":0}}` — which asks the API to move the NIC to LAN 0. That is the same defect that keeps `attach_lan_nic` and `attach_server_cdrom` out, and it needs the same fix in the SDK templates (attach-by-reference should model the body as an id-only object).

**API Reference:** [datacentersLoadbalancersPost](https://api.ionos.com/docs/cloud/v6/#tag/Load-Balancers/operation/datacentersLoadbalancersPost), [datacentersLoadbalancersPatch](https://api.ionos.com/docs/cloud/v6/#tag/Load-Balancers/operation/datacentersLoadbalancersPatch), [datacentersLoadbalancersDelete](https://api.ionos.com/docs/cloud/v6/#tag/Load-Balancers/operation/datacentersLoadbalancersDelete)
