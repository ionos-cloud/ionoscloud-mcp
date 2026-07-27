---
subcategory: "Compute Engine"
page_title: "Security Group"
description: |-
  Tools for listing and inspecting security groups and their rules, and (opt-in) assigning them to servers and NICs, in IONOS CLOUD.
---

# Security Groups

## list_security_groups

Lists all security groups in a specific data center.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center |
| `depth` | integer | No | Nesting depth of returned objects (0–5, default `1`). |
| `filters` | object | No | Server-side property filters; e.g. `{"name":"prod"}`. If the result is empty, retry without filters — a filter typo or mismatch silently returns nothing. |

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
| `depth` | integer | No | Nesting depth of returned objects (0–5). |

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
| `depth` | integer | No | Nesting depth of returned objects (0–5). |
| `filters` | object | No | Server-side property filters; e.g. `{"name":"prod","type":"INGRESS"}`. If the result is empty, retry without filters — a filter typo or mismatch silently returns nothing. |

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
| `depth` | integer | No | Nesting depth of returned objects (0–5). |
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

---

## assign_server_security_groups / assign_nic_security_groups

Sets which security groups a server or a NIC has. Requires `IONOS_MCP_TOOL_SCOPE` to include `write`. Single call.

Both tools **replace the entire set**: any group you omit is unassigned, and an empty list unassigns all of them, removing the protection those groups provided. To *add* a group, first read the current set (`get_server` or `get_nic` at depth 2), then pass that set plus the new ID.

Only the assignment changes — no security group is created or deleted, and the change is reversible by assigning the previous set back. The underlying API call is a `PUT`, so the tools carry `idempotentHint: true`.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center. |
| `server_id` | string | Yes | The ID of the server (both tools). |
| `nic_id` | string | Yes (NIC tool only) | The ID of the NIC. |
| `security_group_ids` | array of string | Yes | The **complete** list of security group IDs the resource should have. Replaces the current set. An empty list unassigns all groups. |

**Example (replace the set with two groups):**

```json
{
  "name": "assign_server_security_groups",
  "arguments": {
    "datacenter_id": "12345678-1234-1234-1234-123456789012",
    "server_id": "87654321-4321-4321-4321-210987654321",
    "security_group_ids": ["sg-aaaa", "sg-bbbb"]
  }
}
```

**API Reference:** [datacentersServersSecuritygroupsPut](https://api.ionos.com/docs/cloud/v6/#tag/Security-Groups/operation/datacentersServersSecuritygroupsPut), [datacentersServersNicsSecuritygroupsPut](https://api.ionos.com/docs/cloud/v6/#tag/Security-Groups/operation/datacentersServersNicsSecuritygroupsPut)
