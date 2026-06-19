---
subcategory: "Compute Engine"
page_title: "Image"
description: |-
  Tools for listing available images (OS templates) in IONOS CLOUD.
---

# Images

## list_images

Lists all available images (OS templates) in IONOS CLOUD.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `depth` | integer | No | Nesting depth of returned objects (0–5, default `1`). Depth 1 includes names and basic properties. |
| `filters` | object | No | Server-side property filters; e.g. `{"name":"ubuntu","imageType":"HDD","licenceType":"LINUX"}`. If the result is empty, retry without filters — a filter typo or mismatch silently returns nothing. |

**Example:**

```json
{
  "name": "list_images",
  "arguments": {}
}
```

**API Reference:** [imagesGet](https://api.ionos.com/docs/cloud/v6/#tag/Images/operation/imagesGet)
