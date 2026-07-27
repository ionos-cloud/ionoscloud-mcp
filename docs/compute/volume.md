---
subcategory: "Compute Engine"
page_title: "Volume"
description: |-
  Tools for listing, inspecting, and (opt-in) creating, updating, and deleting storage volumes in IONOS CLOUD.
---

# Volumes

The `list_*` and `get_*` tools are always available. The write tools — `create_volume`, `update_volume`, and `delete_volume` — register only when `IONOS_MCP_TOOL_SCOPE` opts in (`write` enables create/update; `destructive` also enables delete). `create_volume` and `delete_volume` use a two-phase confirmation: call once **without** `confirmation_token` to get a preview plus a one-time token, then call again **with** that token to execute.

## list_volumes

Lists all volumes in a specific data center.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center |
| `depth` | integer | No | Nesting depth of returned objects (0–5, default `1`). |
| `filters` | object | No | Server-side property filters; e.g. `{"name":"prod","type":"HDD"}`. If the result is empty, retry without filters — a filter typo or mismatch silently returns nothing. |

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

---

## create_volume

Creates one storage volume. Requires `IONOS_MCP_TOOL_SCOPE` to include `write`. Two-phase: the first call (no `confirmation_token`) returns a preview and a one-time token; a second call with that token creates exactly one volume.

Provide `image` or `image_alias` for a bootable volume with an operating system. Without either you get an empty disk and must supply `licence_type` — the request is rejected before it reaches the API otherwise, because there is no image to infer the licence type from. For a bootable Linux volume also set `ssh_keys` or `image_password`, or you will not be able to log in.

The new volume is **not attached to any server** — use `attach_server_volume` afterwards.

`image_password` and `user_data` are reported as `(set, not shown)` in the preview rather than echoed, since previews are shown to the model and logged by clients.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center to create the volume in. |
| `name` | string | Yes | The name of the new volume. |
| `size` | number | Yes | Size in GB. |
| `type` | string | Yes | `HDD`, `SSD`, `SSD Standard`, `SSD Premium` or `DAS`. `DAS` works only inline with a CUBE server and ignores `size`. |
| `image` | string | No | Image or snapshot ID to use as the template. See `list_images`, `list_snapshots`. |
| `image_alias` | string | No | Image alias, e.g. `ubuntu:latest`. An alternative to `image`. |
| `image_password` | string | No | Initial root/administrator password; public images only. Characters `a-z`, `A-Z`, `0-9`, minimum 8. **Cannot be changed later.** |
| `ssh_keys` | array of string | No | Public SSH keys to authorize. Public Linux images only. Settable at creation only; reads always return `null`. |
| `licence_type` | string | No | `LINUX`, `WINDOWS`, `WINDOWS2016`, `WINDOWS2022`, `WINDOWS2025`, `UNKNOWN` or `OTHER`. Required when neither `image` nor `image_alias` is given. |
| `availability_zone` | string | No | `AUTO` (default), `ZONE_1`, `ZONE_2` or `ZONE_3`. Not available for `DAS`. |
| `bus` | string | No | `VIRTIO` (default, faster) or `IDE`. Use `IDE` only for images without VirtIO drivers. |
| `user_data` | string | No | Base64-encoded cloud-init configuration. Requires a cloud-init-capable image. Settable at creation only. |
| `backupunit_id` | string | No | Backup unit to associate. Requires `image` or `image_alias`. Settable at creation only. |
| `expose_serial` | boolean | No | Expose the disk serial ID to the server; some OSes and licensed software require it. |
| `confirmation_token` | string | No | Omit on the first call to receive a preview + token; pass the token (with the same `datacenter_id` and `name`) on the second call to create. Expires after a few minutes. |

**Example (step 1 — preview):**

```json
{
  "name": "create_volume",
  "arguments": { "datacenter_id": "12345678-1234-1234-1234-123456789012", "name": "data-1", "size": 50, "type": "SSD", "image_alias": "ubuntu:latest", "ssh_keys": ["ssh-ed25519 AAAA..."] }
}
```

**API Reference:** [datacentersVolumesPost](https://api.ionos.com/docs/cloud/v6/#tag/Volumes/operation/datacentersVolumesPost)

---

## update_volume

Updates a volume's name, size, bus type or serial exposure. Requires `IONOS_MCP_TOOL_SCOPE` to include `write`. Applies a partial update (only the fields you provide). Single call — no confirmation token.

`size` can only be **increased** — the API rejects shrinking a volume — and after growing it the guest OS still has to extend its own filesystem to use the new space. `image_password`, `user_data` and `backupunit_id` are creation-only and cannot be changed here.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center the volume is in. |
| `volume_id` | string | Yes | The ID of the volume to update. |
| `name` | string | No | A new name. |
| `size` | number | No | A new size in GB. **Increase only.** |
| `bus` | string | No | `VIRTIO` or `IDE`. Requires a server restart to take effect. |
| `expose_serial` | boolean | No | Expose the disk serial ID, or stop exposing it. |

**Example:**

```json
{
  "name": "update_volume",
  "arguments": {
    "datacenter_id": "12345678-1234-1234-1234-123456789012",
    "volume_id": "aabbccdd-1122-3344-5566-778899aabbcc",
    "size": 100
  }
}
```

**API Reference:** [datacentersVolumesPatch](https://api.ionos.com/docs/cloud/v6/#tag/Volumes/operation/datacentersVolumesPatch)

---

## delete_volume

Deletes a volume and all data on it. Irreversible without a snapshot. Requires `IONOS_MCP_TOOL_SCOPE` to include `destructive`. Two-phase: the first call (no `confirmation_token`) returns a preview and a one-time token; a second call with that token performs the deletion.

If the volume is still attached to a server the preview says so and warns that it may be that server's boot disk, in which case deleting it can leave the server unable to boot. Take a snapshot with `create_volume_snapshot` first if you might need the data.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center the volume is in. |
| `volume_id` | string | Yes | The ID of the volume to delete. |
| `confirmation_token` | string | No | Omit on the first call to receive the preview + token; pass the token on the second call to delete. Bound to this volume; expires after a few minutes. |

**Example (step 1 — preview):**

```json
{
  "name": "delete_volume",
  "arguments": { "datacenter_id": "12345678-1234-1234-1234-123456789012", "volume_id": "aabbccdd-1122-3344-5566-778899aabbcc" }
}
```

**API Reference:** [datacentersVolumesDelete](https://api.ionos.com/docs/cloud/v6/#tag/Volumes/operation/datacentersVolumesDelete)

---

## create_volume_snapshot

Takes a snapshot of a volume, capturing its contents at that moment. Requires `IONOS_MCP_TOOL_SCOPE` to include `write`. Two-phase.

A snapshot is the **only** way to recover a volume's data after `delete_volume` or `restore_volume_snapshot`, so take one before any destructive change you might need to undo. Snapshots are billed for the storage they occupy.

This is the only way to create a snapshot — there is no standalone `create_snapshot`.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center the volume is in. |
| `volume_id` | string | Yes | The ID of the volume to snapshot. |
| `name` | string | Yes | The name of the new snapshot. |
| `description` | string | No | An optional description, e.g. what state the volume was in. |
| `sec_auth_protection` | boolean | No | Require extra protection (two-step verification) before the snapshot can be deleted. |
| `licence_type` | string | No | OS type. Defaults to the source volume's licence type. |
| `confirmation_token` | string | No | Omit on the first call for a preview + token; pass it (with the same `datacenter_id`, `volume_id` and `name`) on the second call. |

**API Reference:** [datacentersVolumesCreateSnapshotPost](https://api.ionos.com/docs/cloud/v6/#tag/Volumes/operation/datacentersVolumesCreateSnapshotPost)

---

## restore_volume_snapshot

Restores a snapshot onto a volume, **overwriting everything currently on it**. Requires `IONOS_MCP_TOOL_SCOPE` to include `destructive`. Two-phase.

The volume's current contents are unrecoverable afterwards unless you snapshot them first with `create_volume_snapshot`. If the volume is attached to a server, stop that server with `stop_server` before restoring — otherwise the running guest and the restored disk disagree. The preview says so when it detects an attachment.

The confirmation token is bound to **both** the volume and the snapshot, so a token previewed for one snapshot cannot be replayed to restore a different one.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center the volume is in. |
| `volume_id` | string | Yes | The volume to restore **into**. Its current contents are overwritten. |
| `snapshot_id` | string | Yes | The snapshot to restore from; find it with `list_snapshots`. |
| `confirmation_token` | string | No | Omit on the first call for a preview + token; pass it on the second call to restore. |

**API Reference:** [datacentersVolumesRestoreSnapshotPost](https://api.ionos.com/docs/cloud/v6/#tag/Volumes/operation/datacentersVolumesRestoreSnapshotPost)
