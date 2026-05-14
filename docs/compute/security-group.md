---
subcategory: "Compute Engine"
page_title: "Security Group"
description: |-
  Tools for listing and inspecting security groups and their rules in IONOS CLOUD.
---

# Security Groups

## list_security_groups

Lists all security groups in a specific data center.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center |

**Example:**

```json
{
  "name": "list_security_groups",
  "arguments": {
    "datacenter_id": "12345678-1234-1234-1234-123456789012"
  }
}
```

**API Reference:** [datacentersSecuritygroupsGet](https://api.ionos.com/docs/cloud/v6/#tag/SecurityGroups/operation/datacentersSecuritygroupsGet)

---

## get_security_group

Gets detailed information about a specific security group.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center |
| `security_group_id` | string | Yes | The ID of the security group |

**Example:**

```json
{
  "name": "get_security_group",
  "arguments": {
    "datacenter_id": "12345678-1234-1234-1234-123456789012",
    "security_group_id": "87654321-4321-4321-4321-210987654321"
  }
}
```

**API Reference:** [datacentersSecuritygroupsFindById](https://api.ionos.com/docs/cloud/v6/#tag/SecurityGroups/operation/datacentersSecuritygroupsFindById)

---

## list_security_group_rules

Lists all rules in a specific security group.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center |
| `security_group_id` | string | Yes | The ID of the security group |

**Example:**

```json
{
  "name": "list_security_group_rules",
  "arguments": {
    "datacenter_id": "12345678-1234-1234-1234-123456789012",
    "security_group_id": "87654321-4321-4321-4321-210987654321"
  }
}
```

**API Reference:** [datacentersSecuritygroupsRulesGet](https://api.ionos.com/docs/cloud/v6/#tag/SecurityGroups/operation/datacentersSecuritygroupsRulesGet)

---

## get_security_group_rule

Gets details of a specific security group rule.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center |
| `security_group_id` | string | Yes | The ID of the security group |
| `rule_id` | string | Yes | The ID of the security group rule |

**Example:**

```json
{
  "name": "get_security_group_rule",
  "arguments": {
    "datacenter_id": "12345678-1234-1234-1234-123456789012",
    "security_group_id": "87654321-4321-4321-4321-210987654321",
    "rule_id": "abcdef12-3456-7890-abcd-ef1234567890"
  }
}
```

**API Reference:** [datacentersSecuritygroupsRulesFindById](https://api.ionos.com/docs/cloud/v6/#tag/SecurityGroups/operation/datacentersSecuritygroupsRulesFindById)
