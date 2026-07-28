---
subcategory: "Compute Engine"
page_title: "Snapshot"
description: |-
  Tools for listing, inspecting, and (opt-in) updating and deleting volume snapshots in IONOS CLOUD.
---

# Snapshots

The `list_*` and `get_*` tools are always available. The write tools register only when `IONOS_MCP_TOOL_SCOPE` opts in (`write` enables update; `destructive` also enables delete). `delete_*` uses a two-phase confirmation: call once **without** `confirmation_token` to get a preview plus a one-time token, then call again **with** that token to execute.

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

---

## There is no create_snapshot

Snapshots are produced from a volume, not created directly. Use [`create_volume_snapshot`](volume.md) — it takes the volume you want to capture and is two-phase confirmed. `restore_volume_snapshot` writes one back onto a volume.

A snapshot's `size` and `location` are fixed by the volume it came from and cannot be changed.

---

## update_snapshot

Updates a snapshot's name, description, licence type, protection flag or capability flags. Requires `write`. Partial update — one request, no carry-forward read needed.

The hot-plug and BIOS flags describe what a volume **restored from** this snapshot will support, so changing them affects future restores rather than the snapshot's stored data.

| Name | Type | Description |
|------|------|-------------|
| `snapshot_id` | string | **Required.** |
| `name`, `description` | string | Metadata. |
| `licence_type` | string | `LINUX`, `WINDOWS`, `WINDOWS2016`, `WINDOWS2022`, `WINDOWS2025`, `UNKNOWN`, `OTHER`. |
| `sec_auth_protection` | boolean | Require extra protection before the snapshot can be deleted. |
| `expose_serial`, `require_legacy_bios` | boolean | Applied to volumes restored from this snapshot. |
| `cpu_hot_plug`, `cpu_hot_unplug`, `ram_hot_plug`, `ram_hot_unplug`, `nic_hot_plug`, `nic_hot_unplug`, `disc_virtio_hot_plug`, `disc_virtio_hot_unplug`, `disc_scsi_hot_plug`, `disc_scsi_hot_unplug` | boolean | Capability flags. Snapshots support a **wider set than volumes**, which have neither the CPU/RAM unplug nor the SCSI flags. |

**API Reference:** [snapshotsPatch](https://api.ionos.com/docs/cloud/v6/#tag/Snapshots/operation/snapshotsPatch)

---

## delete_snapshot

Deletes a snapshot. Irreversible. Requires `destructive`. Two-phase.

A snapshot is frequently the only copy of a volume's earlier contents, so deleting it can remove the only way back — `restore_volume_snapshot` will have nothing to restore from. The preview shows the snapshot's name, description, size and licence type, and flags one marked with `sec_auth_protection`, since that usually means someone deliberately marked it important.

**API Reference:** [snapshotsDelete](https://api.ionos.com/docs/cloud/v6/#tag/Snapshots/operation/snapshotsDelete)
