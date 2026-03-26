---
subcategory: "Compute Engine"
page_title: "Snapshot"
description: |-
  Tools for listing and inspecting snapshots in IONOS Cloud.
---

# Snapshots

## list_snapshots

Lists all snapshots in your IONOS Cloud account.

**Parameters:** None

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
