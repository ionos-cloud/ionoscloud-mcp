---
subcategory: "Compute Engine"
page_title: "Network Load Balancer"
description: |-
  Tools for listing and inspecting network load balancers (NLB) in IONOS CLOUD.
---

# Network Load Balancers

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
