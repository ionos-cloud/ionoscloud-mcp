---
subcategory: "Compute Engine"
page_title: "Location"
description: |-
  Tools for listing available locations (regions) in IONOS CLOUD.
---

# Locations

## list_locations

Lists all available locations (regions) in IONOS CLOUD.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `depth` | integer | No | Nesting depth of returned objects (0–5, default `1`). Depth 1 includes names and basic properties. |
| `filters` | object | No | Server-side property filters; e.g. `{"name":"Frankfurt"}`. If the result is empty, retry without filters — a filter typo or mismatch silently returns nothing. |

**Example:**

```json
{
  "name": "list_locations",
  "arguments": {}
}
```

**API Reference:** [locationsGet](https://api.ionos.com/docs/cloud/v6/#tag/Locations/operation/locationsGet)
