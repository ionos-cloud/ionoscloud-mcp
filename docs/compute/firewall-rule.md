---
subcategory: "Compute Engine"
page_title: "Firewall Rule"
description: |-
  Tools for listing and inspecting firewall rules in IONOS Cloud.
---

# Firewall Rules

## list_firewall_rules

Lists all firewall rules on a specific network interface.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center |
| `server_id` | string | Yes | The ID of the server |
| `nic_id` | string | Yes | The ID of the network interface |

**Example:**

```json
{
  "name": "list_firewall_rules",
  "arguments": {
    "datacenter_id": "12345678-1234-1234-1234-123456789012",
    "server_id": "87654321-4321-4321-4321-210987654321",
    "nic_id": "abcdef12-3456-7890-abcd-ef1234567890"
  }
}
```

**API Reference:** [datacentersServersNicsFirewallrulesGet](https://api.ionos.com/docs/cloud/v6/#tag/FirewallRules/operation/datacentersServersNicsFirewallrulesGet)

---

## get_firewall_rule

Gets detailed information about a specific firewall rule.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center |
| `server_id` | string | Yes | The ID of the server |
| `nic_id` | string | Yes | The ID of the network interface |
| `firewallrule_id` | string | Yes | The ID of the firewall rule |

**Example:**

```json
{
  "name": "get_firewall_rule",
  "arguments": {
    "datacenter_id": "12345678-1234-1234-1234-123456789012",
    "server_id": "87654321-4321-4321-4321-210987654321",
    "nic_id": "abcdef12-3456-7890-abcd-ef1234567890",
    "firewallrule_id": "11111111-2222-3333-4444-555555555555"
  }
}
```

**API Reference:** [datacentersServersNicsFirewallrulesFindById](https://api.ionos.com/docs/cloud/v6/#tag/FirewallRules/operation/datacentersServersNicsFirewallrulesFindById)
