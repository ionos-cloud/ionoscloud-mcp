---
subcategory: "Compute Engine"
page_title: "Target Group"
description: |-
  Tools for listing and inspecting target groups in IONOS CLOUD.
---

# Target Groups

## list_target_groups

Lists all target groups in your IONOS CLOUD account.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `depth` | integer | No | Nesting depth of returned objects (0–5). |

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
