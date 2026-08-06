---
subcategory: "Compute Engine"
page_title: "Firewall Rule"
description: |-
  Tools for listing, inspecting, and (opt-in) creating, updating and deleting firewall rules in IONOS CLOUD.
---

# Firewall Rules

The `list_*` and `get_*` tools are always available. The write tools register only when `IONOS_MCP_TOOL_SCOPE` opts in (`write` enables create/update; `destructive` also enables delete). `create_*` and `delete_*` use a two-phase confirmation: call once **without** `confirmation_token` to get a preview plus a one-time token, then call again **with** that token to execute.

## list_firewall_rules

Lists all firewall rules on a specific network interface.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center |
| `server_id` | string | Yes | The ID of the server |
| `nic_id` | string | Yes | The ID of the network interface |
| `depth` | integer | No | Nesting depth of returned objects (0–5, default `1`). |
| `filters` | object | No | Server-side property filters; e.g. `{"name":"prod","type":"INGRESS"}`. If the result is empty, retry without filters — a filter typo or mismatch silently returns nothing. |

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
| `depth` | integer | No | Nesting depth of returned objects (0–5). |

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

---

## Rule fields

Firewall rules are **allow-rules**: an active firewall permits only what its rules match, so the first rule added to a NIC is what makes it reachable at all. The same fields apply whether the rule lives on a NIC or in a security group.

| Name | Type | Description |
|------|------|-------------|
| `protocol` | string | `TCP`, `UDP`, `ICMP`, `ICMPv6`, `GRE`, `VRRP`, `ESP`, `AH` or `ANY`. **Required on create.** |
| `name` | string | A name for the rule. |
| `type` | string | `INGRESS` (inbound, default) or `EGRESS` (outbound). |
| `source_ip` / `target_ip` | string | Restrict to an address or CIDR range. On create, omit to match any. |
| `source_mac` | string | Restrict to a MAC address. |
| `ip_version` | string | `IPv4` or `IPv6`. |
| `port_range_start` / `port_range_end` | integer | 1–65534, **given together**. TCP/UDP only. **Omitting both allows every port.** |
| `icmp_type` / `icmp_code` | integer | 0–254. ICMP/ICMPv6 only. On create, omit to allow all. |
| `clear` | array of string | **Update only.** Resets fields to "match anything" by sending an explicit `null`. Accepts `source_ip`, `target_ip`, `source_mac`, `ip_version`, `icmp_type`, `icmp_code`. |

Combinations the tools reject before sending, since the API's own message does not name the offending field: ICMP fields with TCP/UDP, ports with ICMP, a half-open port range, an inverted range, and out-of-range values.

### Widening a rule back to "any"

Omitting a field on **update** means "leave it unchanged", not "match anything" — so `clear` is the only way to reopen a field that already has a value:

```json
{"datacenter_id": "…", "server_id": "…", "nic_id": "…", "firewallrule_id": "…", "clear": ["source_ip"]}
```

Do not reach for `0.0.0.0/0` (or `::/0`) instead. The API accepts it and echoes it back unchanged, then stores the bare network address `0.0.0.0` once the request settles — a non-routable address that matches **no** traffic, so a rule written to open a port to the world silently closes it. Both create and update reject an all-addresses CIDR and point you at the right mechanism.

A field cannot be both given a value and listed in `clear`.

---

## NIC rules vs. security group rules

| | `create_firewall_rule` | `create_security_group_rule` |
|---|---|---|
| Applies to | one NIC | every server and NIC assigned to the group |
| Use when | the rule is specific to one interface | the rule should apply to several resources |

Prefer the group form for anything shared — the preview reports how many servers and NICs will inherit the rule, and warns when the group has no members yet (in which case the rule has no effect until you assign it with `assign_server_security_groups` or `assign_nic_security_groups`).

The security-group rule tools all operate on `/securitygroups/{id}/rules`, despite the SDK naming create and delete `…Firewallrules…` and update `…Rules…`.

---

## create_firewall_rule / update_firewall_rule / delete_firewall_rule

NIC-scoped. `create` and `delete` are two-phase; `update` is a single partial `PATCH`. Parent IDs: `datacenter_id`, `server_id`, `nic_id`, plus `firewallrule_id` for update and delete.

`delete_firewall_rule`'s preview describes exactly what the rule allows (protocol, direction, addresses, port range) so you can see which traffic stops being permitted, and warns that deleting a NIC's **last** rule while its firewall is active blocks all incoming traffic.

**API Reference:** [datacentersServersNicsFirewallrulesPost](https://api.ionos.com/docs/cloud/v6/#tag/Firewall-Rules/operation/datacentersServersNicsFirewallrulesPost), [datacentersServersNicsFirewallrulesPatch](https://api.ionos.com/docs/cloud/v6/#tag/Firewall-Rules/operation/datacentersServersNicsFirewallrulesPatch), [datacentersServersNicsFirewallrulesDelete](https://api.ionos.com/docs/cloud/v6/#tag/Firewall-Rules/operation/datacentersServersNicsFirewallrulesDelete)

---

## create_security_group_rule / update_security_group_rule / delete_security_group_rule

Group-scoped, so a change reaches every member at once — including `clear`, which widens the rule for every server and NIC in the group. Parent IDs: `datacenter_id`, `security_group_id`, plus `rule_id` for update and delete.

**API Reference:** [datacentersSecuritygroupsFirewallrulesPost](https://api.ionos.com/docs/cloud/v6/#tag/Security-Groups/operation/datacentersSecuritygroupsFirewallrulesPost), [datacentersSecuritygroupsRulesPatch](https://api.ionos.com/docs/cloud/v6/#tag/Security-Groups/operation/datacentersSecuritygroupsRulesPatch), [datacentersSecuritygroupsFirewallrulesDelete](https://api.ionos.com/docs/cloud/v6/#tag/Security-Groups/operation/datacentersSecuritygroupsFirewallrulesDelete)
