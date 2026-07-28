---
subcategory: "Compute Engine"
page_title: "Target Group"
description: |-
  Tools for listing, inspecting, and (opt-in) creating, updating and deleting target groups in IONOS CLOUD.
---

# Target Groups

The `list_*` and `get_*` tools are always available. The write tools register only when `IONOS_MCP_TOOL_SCOPE` opts in (`write` enables create/update; `destructive` also enables delete). `create_*` and `delete_*` use a two-phase confirmation: call once **without** `confirmation_token` to get a preview plus a one-time token, then call again **with** that token to execute.

## list_target_groups

Lists all target groups in your IONOS CLOUD account.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `depth` | integer | No | Nesting depth of returned objects (0–5). |
| `filters` | object | No | Server-side property filters; e.g. `{"name":"prod"}`. If the result is empty, retry without filters — a filter typo or mismatch silently returns nothing. |

**Example:**

```json
{
  "name": "list_target_groups",
  "arguments": {}
}
```

**API Reference:** [targetgroupsGet](https://api.ionos.com/docs/cloud/v6/#tag/TargetGroups/operation/targetgroupsGet)

---

## get_target_group

Gets detailed information about a specific target group.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `target_group_id` | string | Yes | The ID of the target group |
| `depth` | integer | No | Nesting depth of returned objects (0–5). |

**Example:**

```json
{
  "name": "get_target_group",
  "arguments": {
    "target_group_id": "12345678-1234-1234-1234-123456789012"
  }
}
```

**API Reference:** [targetgroupsFindByTargetGroupId](https://api.ionos.com/docs/cloud/v6/#tag/TargetGroups/operation/targetgroupsFindByTargetGroupId)

---

## Why the update tools issue a GET first

Every model in this area except the classic load balancer serializes its **required** fields unconditionally — the IONOS Go SDK sends them whether or not you set them. A partial update built the obvious way would therefore send empty values for fields you never mentioned:

| Resource | Fields always sent | What a naive partial update would do |
|---|---|---|
| network / application load balancer | `name`, `listenerLan`, `targetLan` | move it off **both** its client and backend LANs |
| NLB forwarding rule | + `targets` | **remove every backend** from the load balancer |
| ALB forwarding rule | `name`, `protocol`, `listenerIp`, `listenerPort` | break the listener |
| target group | `name`, `algorithm`, `protocol` | reset the pool's algorithm and protocol |
| NAT gateway | `name`, `publicIps` | leave it with no address to translate to |
| NAT gateway rule | `name`, `sourceSubnet`, `publicIp` | translate nothing |

So each `update_*` reads the current resource and sends those values back unchanged, overriding only what you supplied. That is why you see a `GET` before the `PATCH`, and why omitting a field is always safe.

Lists behave differently from scalars: supplying `targets`, `http_rules`, `server_certificates`, `public_ips` or `lans` **replaces** that list, so include every entry the resource should keep. Omitting the field entirely leaves the current list untouched. An explicit empty list is rejected rather than applied, so you cannot empty a backend pool by accident.
