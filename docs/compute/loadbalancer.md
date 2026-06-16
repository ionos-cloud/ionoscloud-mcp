---
subcategory: "Compute Engine"
page_title: "Load Balancer"
description: |-
  Tools for listing and inspecting load balancers in IONOS CLOUD.
---

# Load Balancers

## list_loadbalancers

Lists all load balancers in a specific data center.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center |
| `depth` | integer | No | Nesting depth of returned objects (0–5, default `1`). |

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
