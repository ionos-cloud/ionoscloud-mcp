---
subcategory: "Compute Engine"
page_title: "NAT Gateway"
description: |-
  Tools for listing and inspecting NAT gateways in IONOS CLOUD.
---

# NAT Gateways

## list_nat_gateways

Lists all NAT gateways in a specific data center.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center |

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
