---
subcategory: "Compute Engine"
page_title: "NAT Gateway"
description: |-
  Tools for listing, inspecting, and (opt-in) creating, updating and deleting NAT gateways and their rules in IONOS CLOUD.
---

# NAT Gateways

The `list_*` and `get_*` tools are always available. The write tools register only when `IONOS_MCP_TOOL_SCOPE` opts in (`write` enables create/update; `destructive` also enables delete). `create_*` and `delete_*` use a two-phase confirmation: call once **without** `confirmation_token` to get a preview plus a one-time token, then call again **with** that token to execute.

## list_nat_gateways

Lists all NAT gateways in a specific data center.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center |
| `depth` | integer | No | Nesting depth of returned objects (0–5, default `1`). |
| `filters` | object | No | Server-side property filters; e.g. `{"name":"prod"}`. If the result is empty, retry without filters — a filter typo or mismatch silently returns nothing. |

**Example:**

```json
{
  "name": "list_nat_gateways",
  "arguments": {
    "datacenter_id": "12345678-1234-1234-1234-123456789012"
  }
}
```

**API Reference:** [datacentersNatgatewaysGet](https://api.ionos.com/docs/cloud/v6/#tag/NATGateways/operation/datacentersNatgatewaysGet)

---

## get_nat_gateway

Gets detailed information about a specific NAT gateway.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center |
| `nat_gateway_id` | string | Yes | The ID of the NAT gateway |
| `depth` | integer | No | Nesting depth of returned objects (0–5). |

**Example:**

```json
{
  "name": "get_nat_gateway",
  "arguments": {
    "datacenter_id": "12345678-1234-1234-1234-123456789012",
    "nat_gateway_id": "87654321-4321-4321-4321-210987654321"
  }
}
```

**API Reference:** [datacentersNatgatewaysFindByNatGatewayId](https://api.ionos.com/docs/cloud/v6/#tag/NATGateways/operation/datacentersNatgatewaysFindByNatGatewayId)

---

## list_nat_gateway_rules

Lists all rules of a specific NAT gateway.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center |
| `nat_gateway_id` | string | Yes | The ID of the NAT gateway |
| `depth` | integer | No | Nesting depth of returned objects (0–5). |
| `filters` | object | No | Server-side property filters; e.g. `{"name":"prod","type":"SNAT"}`. If the result is empty, retry without filters — a filter typo or mismatch silently returns nothing. |

**Example:**

```json
{
  "name": "list_nat_gateway_rules",
  "arguments": {
    "datacenter_id": "12345678-1234-1234-1234-123456789012",
    "nat_gateway_id": "87654321-4321-4321-4321-210987654321"
  }
}
```

**API Reference:** [datacentersNatgatewaysRulesGet](https://api.ionos.com/docs/cloud/v6/#tag/NATGateways/operation/datacentersNatgatewaysRulesGet)

---

## Omitted fields and list fields

Every `update_*` tool here applies a **partial update** (an HTTP `PATCH`): send only the
properties you want to change, and anything you omit keeps its current value. That holds for the
properties the API marks required too — you do not need to repeat them just to change something
else.

Lists behave differently from single values. Supplying `public_ips` or `lans` **replaces** that
whole list, so include every entry the gateway should keep; omitting the field leaves the current
list untouched. An explicit empty `public_ips` is rejected rather than applied, since a gateway
with no public address has nothing to translate to. An empty `lans` is accepted — `lans` is
optional, so a gateway with no LANs is a valid state. It carries no traffic in that state, and
its rules keep reporting `state: AVAILABLE` even though nothing routes through them, so check
`lans` on the gateway itself rather than trusting a rule's state. Send the field only when you
mean to set the complete list.
