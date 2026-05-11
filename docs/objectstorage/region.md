---
subcategory: "Object Storage"
page_title: "Regions"
description: |-
  Tools for listing and inspecting Object Storage regions in IONOS CLOUD.
---

# Object Storage Regions

## list_object_storage_regions

Lists all available Object Storage regions.

**Parameters:** None

**Example:**

```json
{
  "name": "list_object_storage_regions",
  "arguments": {}
}
```

**API Reference:** [RegionsGet](https://api.ionos.com/docs/object-storage-management/v1/#tag/Regions/operation/regionsGet)

---

## get_object_storage_region

Get details of a specific Object Storage region by name (e.g. `eu-central-3`).

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `region` | string | Yes | The region name (e.g. `eu-central-3`) |

**Example:**

```json
{
  "name": "get_object_storage_region",
  "arguments": {
    "region": "eu-central-3"
  }
}
```

**API Reference:** [RegionsFindByRegion](https://api.ionos.com/docs/object-storage-management/v1/#tag/Regions/operation/regionsFindByRegion)
