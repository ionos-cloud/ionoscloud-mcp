---
subcategory: "Compute Engine"
page_title: "Template"
description: |-
  Tools for listing and inspecting server templates in IONOS Cloud.
---

# Templates

## list_templates

Lists all available server templates in IONOS Cloud.

**Parameters:**

None.

**Example:**

```json
{
  "name": "list_templates",
  "arguments": {}
}
```

**API Reference:** [templatesGet](https://api.ionos.com/docs/cloud/v6/#tag/Templates/operation/templatesGet)

---

## get_template

Gets detailed information about a specific server template.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `template_id` | string | Yes | The ID of the template |

**Example:**

```json
{
  "name": "get_template",
  "arguments": {
    "template_id": "12345678-1234-1234-1234-123456789012"
  }
}
```

**API Reference:** [templatesFindById](https://api.ionos.com/docs/cloud/v6/#tag/Templates/operation/templatesFindById)
