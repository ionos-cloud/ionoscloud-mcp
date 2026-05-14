---
subcategory: "Compute Engine"
page_title: "Application Load Balancer"
description: |-
  Tools for listing and inspecting application load balancers (ALB) in IONOS CLOUD.
---

# Application Load Balancers

## list_application_loadbalancers

Lists all application load balancers (ALB) in a specific data center.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center |

**Example:**

```json
{
  "name": "list_application_loadbalancers",
  "arguments": {
    "datacenter_id": "12345678-1234-1234-1234-123456789012"
  }
}
```

**API Reference:** [datacentersApplicationloadbalancersGet](https://api.ionos.com/docs/cloud/v6/#tag/ApplicationLoadBalancers/operation/datacentersApplicationloadbalancersGet)

---

## get_application_loadbalancer

Gets detailed information about a specific application load balancer (ALB).

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center |
| `application_loadbalancer_id` | string | Yes | The ID of the application load balancer |

**Example:**

```json
{
  "name": "get_application_loadbalancer",
  "arguments": {
    "datacenter_id": "12345678-1234-1234-1234-123456789012",
    "application_loadbalancer_id": "87654321-4321-4321-4321-210987654321"
  }
}
```

**API Reference:** [datacentersApplicationloadbalancersFindByApplicationLoadBalancerId](https://api.ionos.com/docs/cloud/v6/#tag/ApplicationLoadBalancers/operation/datacentersApplicationloadbalancersFindByApplicationLoadBalancerId)

---

## list_alb_forwarding_rules

Lists all forwarding rules of an application load balancer.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center |
| `application_loadbalancer_id` | string | Yes | The ID of the application load balancer |

**Example:**

```json
{
  "name": "list_alb_forwarding_rules",
  "arguments": {
    "datacenter_id": "12345678-1234-1234-1234-123456789012",
    "application_loadbalancer_id": "87654321-4321-4321-4321-210987654321"
  }
}
```

**API Reference:** [datacentersApplicationloadbalancersForwardingrulesGet](https://api.ionos.com/docs/cloud/v6/#tag/ApplicationLoadBalancers/operation/datacentersApplicationloadbalancersForwardingrulesGet)
