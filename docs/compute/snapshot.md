---
subcategory: "Compute Engine"
page_title: "Snapshot"
description: |-
  Tools for listing and inspecting snapshots in IONOS CLOUD.
---

# Snapshots

## list_snapshots

Lists all snapshots in your IONOS CLOUD account.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `depth` | integer | No | Nesting depth of returned objects (0–5, default `1`). Depth 1 includes names and basic properties. |
| `filters` | object | No | Server-side property filters; e.g. `{"name":"prod","location":"de/fra"}`. If the result is empty, retry without filters — a filter typo or mismatch silently returns nothing. |

**Example:**

```json
{
  "name": "list_snapshots",
  "arguments": {}
}
```

**API Reference:** [snapshotsGet](https://api.ionos.com/docs/cloud/v6/#tag/Snapshots/operation/snapshotsGet)

---

## get_snapshot

Gets detailed information about a specific snapshot.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `snapshot_id` | string | Yes | The ID of the snapshot |
| `depth` | integer | No | Nesting depth of returned objects (0–5). |

**Example:**

```json
{
  "name": "get_snapshot",
  "arguments": {
    "snapshot_id": "22222222-2222-2222-2222-222222222222"
  }
}
```

**API Reference:** [snapshotsFindById](https://api.ionos.com/docs/cloud/v6/#tag/Snapshots/operation/snapshotsFindById)
