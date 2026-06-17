---
subcategory: "Compute Engine"
page_title: "Template"
description: |-
  Tools for listing and inspecting server templates in IONOS CLOUD.
---

# Templates

## list_templates

Lists all available server templates in IONOS CLOUD.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `depth` | integer | No | Nesting depth of returned objects (0–5). |

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
| `depth` | integer | No | Nesting depth of returned objects (0–5). |

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
