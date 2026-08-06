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

## Omitted fields and list fields

Every `update_*` tool here applies a **partial update** (an HTTP `PATCH`): send only the
properties you want to change, and anything you omit keeps its current value. That holds for the
properties the API marks required too — you do not need to repeat them just to change something
else.

Lists behave differently from single values. Supplying `targets` **replaces** the whole backend
list, so include every backend the group should keep; omitting the field leaves the current list
untouched. An explicit empty list is rejected rather than applied, so a backend pool cannot be
emptied by accident.
