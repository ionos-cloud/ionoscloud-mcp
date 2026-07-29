---
subcategory: "Compute Engine"
page_title: "Application Load Balancer"
description: |-
  Tools for listing, inspecting, and (opt-in) creating, updating and deleting application load balancers and their forwarding rules in IONOS CLOUD.
---

# Application Load Balancers

The `list_*` and `get_*` tools are always available. The write tools register only when `IONOS_MCP_TOOL_SCOPE` opts in (`write` enables create/update; `destructive` also enables delete). `create_*` and `delete_*` use a two-phase confirmation: call once **without** `confirmation_token` to get a preview plus a one-time token, then call again **with** that token to execute.

## list_application_loadbalancers

Lists all application load balancers (ALB) in a specific data center.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center |
| `depth` | integer | No | Nesting depth of returned objects (0–5, default `1`). |
| `filters` | object | No | Server-side property filters; e.g. `{"name":"prod"}`. If the result is empty, retry without filters — a filter typo or mismatch silently returns nothing. |

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
| `depth` | integer | No | Nesting depth of returned objects (0–5). |

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
| `depth` | integer | No | Nesting depth of returned objects (0–5). |
| `filters` | object | No | Server-side property filters; e.g. `{"name":"prod"}`. If the result is empty, retry without filters — a filter typo or mismatch silently returns nothing. |

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

---

## Omitted fields and list fields

Every `update_*` tool here applies a **partial update**: send only the properties you want to
change, and anything you omit keeps its current value. That holds for the properties the API
marks required too — you do not need to repeat them just to change something else.

Lists behave differently from single values. Supplying `targets`, `http_rules`,
`server_certificates`, `public_ips` or `lans` **replaces** that list, so include every entry
the resource should keep. Omitting the field leaves the current list untouched. An explicit
empty list is rejected rather than applied, so a backend pool cannot be emptied by accident.
