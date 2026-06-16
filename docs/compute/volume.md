---
subcategory: "Compute Engine"
page_title: "Volume"
description: |-
  Tools for listing and inspecting volumes in IONOS CLOUD.
---

# Volumes

## list_volumes

Lists all volumes in a specific data center.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center |
| `depth` | integer | No | Nesting depth of returned objects (0–5, default `1`). |

**Example:**

```json
{
  "name": "list_volumes",
  "arguments": {
    "datacenter_id": "12345678-1234-1234-1234-123456789012"
  }
}
```

**API Reference:** [datacentersVolumesGet](https://api.ionos.com/docs/cloud/v6/#tag/Volumes/operation/datacentersVolumesGet)

---

## get_volume

Gets detailed information about a specific volume.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center |
| `volume_id` | string | Yes | The ID of the volume |
| `depth` | integer | No | Nesting depth of returned objects (0–5). |

**Example:**

```json
{
  "name": "get_volume",
  "arguments": {
    "datacenter_id": "12345678-1234-1234-1234-123456789012",
    "volume_id": "11111111-1111-1111-1111-111111111111"
  }
}
```

**API Reference:** [datacentersVolumesFindById](https://api.ionos.com/docs/cloud/v6/#tag/Volumes/operation/datacentersVolumesFindById)
