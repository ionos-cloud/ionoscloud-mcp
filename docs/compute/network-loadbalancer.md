---
subcategory: "Compute Engine"
page_title: "Network Load Balancer"
description: |-
  Tools for listing, inspecting, and (opt-in) creating, updating and deleting network load balancers and their forwarding rules in IONOS CLOUD.
---

# Network Load Balancers

The `list_*` and `get_*` tools are always available. The write tools register only when `IONOS_MCP_TOOL_SCOPE` opts in (`write` enables create/update; `destructive` also enables delete). `create_*` and `delete_*` use a two-phase confirmation: call once **without** `confirmation_token` to get a preview plus a one-time token, then call again **with** that token to execute.

## list_network_loadbalancers

Lists all network load balancers (NLB) in a specific data center.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center |
| `depth` | integer | No | Nesting depth of returned objects (0–5, default `1`). |
| `filters` | object | No | Server-side property filters; e.g. `{"name":"prod"}`. If the result is empty, retry without filters — a filter typo or mismatch silently returns nothing. |

**Example:**

```json
{
  "name": "list_network_loadbalancers",
  "arguments": {
    "datacenter_id": "12345678-1234-1234-1234-123456789012"
  }
}
```

**API Reference:** [datacentersNetworkloadbalancersGet](https://api.ionos.com/docs/cloud/v6/#tag/NetworkLoadBalancers/operation/datacentersNetworkloadbalancersGet)

---

## get_network_loadbalancer

Gets detailed information about a specific network load balancer (NLB).

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center |
| `network_loadbalancer_id` | string | Yes | The ID of the network load balancer |
| `depth` | integer | No | Nesting depth of returned objects (0–5). |

**Example:**

```json
{
  "name": "get_network_loadbalancer",
  "arguments": {
    "datacenter_id": "12345678-1234-1234-1234-123456789012",
    "network_loadbalancer_id": "87654321-4321-4321-4321-210987654321"
  }
}
```

**API Reference:** [datacentersNetworkloadbalancersFindByNetworkLoadBalancerId](https://api.ionos.com/docs/cloud/v6/#tag/NetworkLoadBalancers/operation/datacentersNetworkloadbalancersFindByNetworkLoadBalancerId)

---

## list_nlb_forwarding_rules

Lists all forwarding rules of a network load balancer.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center |
| `network_loadbalancer_id` | string | Yes | The ID of the network load balancer |
| `depth` | integer | No | Nesting depth of returned objects (0–5). |
| `filters` | object | No | Server-side property filters; e.g. `{"name":"prod"}`. If the result is empty, retry without filters — a filter typo or mismatch silently returns nothing. |

**Example:**

```json
{
  "name": "list_nlb_forwarding_rules",
  "arguments": {
    "datacenter_id": "12345678-1234-1234-1234-123456789012",
    "network_loadbalancer_id": "87654321-4321-4321-4321-210987654321"
  }
}
```

**API Reference:** [datacentersNetworkloadbalancersForwardingrulesGet](https://api.ionos.com/docs/cloud/v6/#tag/NetworkLoadBalancers/operation/datacentersNetworkloadbalancersForwardingrulesGet)

---

## Omitted fields and list fields

Every `update_*` tool here applies a **partial update** (an HTTP `PATCH`): send only the
properties you want to change, and anything you omit keeps its current value. That holds for the
properties the API marks required too — you do not need to repeat them just to change something
else.

Lists behave differently from single values. Supplying `targets` on a forwarding rule
**replaces** the whole backend list, so include every backend the rule should keep; omitting the
field leaves the current list untouched. An explicit empty list is rejected rather than applied,
so a backend pool cannot be emptied by accident.
