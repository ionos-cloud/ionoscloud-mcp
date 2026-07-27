---
subcategory: "Compute Engine"
page_title: "Server"
description: |-
  Tools for listing, inspecting, and (opt-in) creating, updating, and deleting servers in IONOS CLOUD.
---

# Servers

The `list_*` and `get_*` tools are always available. The write tools — `create_server`, `update_server`, and `delete_server` — register only when `IONOS_MCP_TOOL_SCOPE` opts in (`write` enables create/update; `destructive` also enables delete). `create_server` and `delete_server` use a two-phase confirmation: call once **without** `confirmation_token` to get a preview plus a one-time token, then call again **with** that token to execute.

## list_servers

Lists all servers in a specific data center.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center |
| `depth` | integer | No | Nesting depth of returned objects (0–5, default `1`). |
| `filters` | object | No | Server-side property filters; e.g. `{"name":"prod","vmState":"RUNNING"}`. If the result is empty, retry without filters — a filter typo or mismatch silently returns nothing. |

**Example:**

```json
{
  "name": "list_servers",
  "arguments": {
    "datacenter_id": "12345678-1234-1234-1234-123456789012"
  }
}
```

**API Reference:** [datacentersServersGet](https://api.ionos.com/docs/cloud/v6/#tag/Servers/operation/datacentersServersGet)

---

## get_server

Gets detailed information about a specific server.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center |
| `server_id` | string | Yes | The ID of the server |
| `depth` | integer | No | Nesting depth of returned objects (0–5). |

**Example:**

```json
{
  "name": "get_server",
  "arguments": {
    "datacenter_id": "12345678-1234-1234-1234-123456789012",
    "server_id": "87654321-4321-4321-4321-210987654321"
  }
}
```

**API Reference:** [datacentersServersFindById](https://api.ionos.com/docs/cloud/v6/#tag/Servers/operation/datacentersServersFindById)

---

## create_server

Creates one server (virtual machine). Requires `IONOS_MCP_TOOL_SCOPE` to include `write`. Two-phase: the first call (no `confirmation_token`) returns a preview and a one-time token; a second call with that token creates exactly one server (there is no bulk/batch parameter).

Size the server **either** with `cores` + `ram` (ENTERPRISE, VCPU) **or** with `template_uuid` plus `type` (CUBE, GPU) — supplying neither, or both, is rejected before the request reaches the API. `type` is required alongside `template_uuid` because it is what distinguishes a CUBE template from a GPU one, and CUBE has an extra requirement.

### boot_volume: required for CUBE and GPU servers

Pass `boot_volume` to create the server's disk **in the same request**. For the template-sized types — CUBE and GPU — this is mandatory, not a convenience: the API accepts their storage *only* as part of a composite server-creation call. Such a server created without it is rejected, and attaching a volume afterwards does not work, so there is no recovery path. The tool therefore rejects the combination up front and says what to pass.

| Server type | `boot_volume` | `boot_volume.type` | `boot_volume.size` |
|---|---|---|---|
| CUBE | **required** | **required**, must be `DAS` | must be **omitted** (fixed by `template_uuid`) |
| GPU | **required** | *optional* — omit to let the API choose, or `SSD Premium` | must be **omitted** (fixed by `template_uuid`) |
| ENTERPRISE, VCPU | optional (recommended) | usually `HDD`, `SSD`, `SSD Standard`, `SSD Premium` | usually required |

Only the CUBE and GPU rules above are enforced, because they are documented and a
server created the wrong way cannot be repaired by attaching a volume afterwards.
For ENTERPRISE and VCPU the storage type and size rules are *inferred* rather than
documented, so the tool does not reject them — it adds a `NOTE:` line to the
preview and lets the API decide. A wrong storage type is trivially recoverable
(retry with another value), whereas a mistaken block would break a valid request
with no way around it.

`boot_volume` is also required for a **Confidential Computing** server: the API derives its core count and CPU family from the confidential image on that volume, so it must be present in the same request.

`boot_volume` fields: `type`, `name`, `size`, `image`, `image_alias`, `image_password`, `ssh_keys`, `licence_type`, `bus`, `user_data`. It needs `image` or `image_alias` to install an operating system, or `licence_type` for an empty disk. `image_password` and `user_data` are shown as `(set, not shown)` in the preview rather than echoed.

The confirmation token is bound to the boot volume's storage type, so a preview shown "with a disk" cannot be executed as "with no disk".

Without `boot_volume` the server has **no storage and no network**: follow up with `attach_server_volume` (or `create_volume` then attach) and `create_nic`, or it has nothing to boot from and no way to be reached.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center to create the server in. |
| `name` | string | Yes | The name of the new server. |
| `cores` | integer | No | Total CPU cores. Required for ENTERPRISE/VCPU; must not be set for CUBE/GPU. |
| `ram` | integer | No | Memory in MB, in multiples of 256 (minimum 256; at least 1024 if RAM hot-plug is enabled). Required for ENTERPRISE/VCPU; must not be set for CUBE/GPU. |
| `type` | string | No | `ENTERPRISE` (default), `VCPU`, `CUBE` or `GPU`. |
| `template_uuid` | string | No | Template that fixes the size of a CUBE or GPU server. See `list_templates`. |
| `cpu_family` | string | No | CPU architecture, e.g. `INTEL_SKYLAKE`, `AMD_OPTERON`. Must not be set for CUBE/VCPU. Omit to have one chosen automatically; availability varies per location (`list_locations`). |
| `availability_zone` | string | No | `AUTO` (default), `ZONE_1` or `ZONE_2`. CUBE and GPU accept only `AUTO`. |
| `hostname` | string | No | Characters `a-z`, `0-9`, `-`; must not start with `-`; max 63 characters. |
| `nic_multi_queue` | boolean | No | Enable Multi Queue on all NICs (helps with low NIC throughput). Not allowed for CUBE. **Toggling this restarts the server.** |
| `boot_volume` | object | CUBE, GPU: **yes** | Create the server's disk in the same request. See the table above. |
| `confirmation_token` | string | No | Omit on the first call to receive a preview + token; pass the token (with the same `datacenter_id` and `name`) on the second call to create. Expires after a few minutes. |

**Example (step 1 — preview):**

```json
{
  "name": "create_server",
  "arguments": { "datacenter_id": "12345678-1234-1234-1234-123456789012", "name": "web-1", "cores": 4, "ram": 2048 }
}
```

**Example (step 2 — execute):**

```json
{
  "name": "create_server",
  "arguments": { "datacenter_id": "12345678-1234-1234-1234-123456789012", "name": "web-1", "cores": 4, "ram": 2048, "confirmation_token": "<token-from-step-1>" }
}
```

**Example (CUBE server — the DAS volume must be inline):**

```json
{
  "name": "create_server",
  "arguments": {
    "datacenter_id": "12345678-1234-1234-1234-123456789012",
    "name": "cube-01",
    "type": "CUBE",
    "template_uuid": "<id from list_templates>",
    "boot_volume": {
      "type": "DAS",
      "name": "cube-01-das",
      "image_alias": "ubuntu:latest",
      "ssh_keys": ["ssh-ed25519 AAAA..."]
    }
  }
}
```

**Example (GPU server — also template-sized, so the volume is inline and carries no size):**

```json
{
  "name": "create_server",
  "arguments": {
    "datacenter_id": "12345678-1234-1234-1234-123456789012",
    "name": "gpu-01",
    "type": "GPU",
    "template_uuid": "<id from list_templates, GPU template>",
    "availability_zone": "AUTO",
    "boot_volume": {
      "type": "SSD Premium",
      "name": "system",
      "licence_type": "LINUX",
      "bus": "VIRTIO"
    }
  }
}
```

GPU servers do not accept `cpu_family`, and `availability_zone` must be `AUTO`.

**API Reference:** [datacentersServersPost](https://api.ionos.com/docs/cloud/v6/#tag/Servers/operation/datacentersServersPost)

---

## update_server

Updates a server's name, size, CPU family, hostname or Multi Queue setting. Requires `IONOS_MCP_TOOL_SCOPE` to include `write`. Applies a partial update (only the fields you provide). Single call — no confirmation token.

Resizing may need a reboot before the guest OS sees the change, and toggling `nic_multi_queue` restarts the server. On CUBE and GPU servers the size comes from the template; on servers with Confidential Computing enabled `cores`, `ram` and `cpu_family` are immutable and the API rejects changes to them.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center the server is in. |
| `server_id` | string | Yes | The ID of the server to update. |
| `name` | string | No | A new name. |
| `cores` | integer | No | A new CPU core count. |
| `ram` | integer | No | A new memory size in MB, in multiples of 256. |
| `cpu_family` | string | No | A new CPU architecture. |
| `hostname` | string | No | A new hostname. |
| `nic_multi_queue` | boolean | No | Turn Multi Queue on or off. **Restarts the server.** |

**Example:**

```json
{
  "name": "update_server",
  "arguments": {
    "datacenter_id": "12345678-1234-1234-1234-123456789012",
    "server_id": "87654321-4321-4321-4321-210987654321",
    "cores": 8
  }
}
```

**API Reference:** [datacentersServersPatch](https://api.ionos.com/docs/cloud/v6/#tag/Servers/operation/datacentersServersPatch)

---

## delete_server

Deletes a server. Irreversible. Requires `IONOS_MCP_TOOL_SCOPE` to include `destructive`. Two-phase: the first call (no `confirmation_token`) returns a preview of what is attached and what happens to it, plus a one-time token; a second call with that token performs the deletion.

By default the attached volumes are **not** deleted — they survive as unattached volumes and **keep incurring cost** until removed with `delete_volume`. Set `delete_volumes` to `true` to destroy them with the server, which makes their data unrecoverable. Because that choice changes what is destroyed, the token is bound to it: if you change `delete_volumes`, you must preview again.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center the server is in. |
| `server_id` | string | Yes | The ID of the server to delete. |
| `delete_volumes` | boolean | No | Also delete the attached volumes (default `false`). `false` leaves them as unattached, still-billed volumes; `true` destroys their data irrecoverably. |
| `confirmation_token` | string | No | Omit on the first call to receive the preview + token; pass the token (with the same `delete_volumes` value) on the second call to delete. Expires after a few minutes. |

**Example (step 1 — preview):**

```json
{
  "name": "delete_server",
  "arguments": { "datacenter_id": "12345678-1234-1234-1234-123456789012", "server_id": "87654321-4321-4321-4321-210987654321" }
}
```

**Example (step 2 — execute, also removing the volumes):**

```json
{
  "name": "delete_server",
  "arguments": { "datacenter_id": "12345678-1234-1234-1234-123456789012", "server_id": "87654321-4321-4321-4321-210987654321", "delete_volumes": true, "confirmation_token": "<token-from-a-delete_volumes-true-preview>" }
}
```

**API Reference:** [datacentersServersDelete](https://api.ionos.com/docs/cloud/v6/#tag/Servers/operation/datacentersServersDelete)

---

## Power actions

`start_server` and `resume_server` bring a server up and are **single call** — nothing is interrupted, so there is no confirmation step. `stop_server`, `reboot_server`, `suspend_server` and `upgrade_server` interrupt a running workload and are classified **destructive**: they require `IONOS_MCP_TOOL_SCOPE` to include `destructive` and use the same two-phase confirmation as `delete_*`. Their preview names the server and reports its current `vmState`, which is what lets you catch "wrong server" before the action lands.

Note that a destructive action is not always a `delete_`: `stop_server` is an HTTP `POST` that carries `destructiveHint: true`. All of these endpoints return no body, so the tools report that the request was accepted; poll `get_server` and watch `vmState` for progress.

| Tool | Scope | Confirmation | Notes |
|------|-------|--------------|-------|
| `start_server` | `write` | none | Starts a stopped ENTERPRISE, VCPU or GPU server. **Rejected for CUBE** — use `resume_server`. No effect if already running. For ENTERPRISE this re-allocates cores and RAM and resumes charging. |
| `resume_server` | `write` | none | CUBE servers only; counterpart to `suspend_server`. |
| `stop_server` | `destructive` | two-phase | Like pulling the power — unwritten data is lost. **Rejected for CUBE** — use `suspend_server`. Frees ENTERPRISE cores and RAM and stops charging for them; volumes are kept and still billed. |
| `reboot_server` | `destructive` | two-phase | Hard reset, not a graceful restart. |
| `suspend_server` | `destructive` | two-phase | **CUBE servers only** — rejected for other types. Storage is retained. |
| `upgrade_server` | `destructive` | two-phase | Upgrades to the latest hardware generation and restarts the server. |

All six take `datacenter_id` and `server_id`; the four destructive ones also take `confirmation_token`.

CUBE servers are suspended and resumed; every other type is stopped and started. The API rejects the wrong pairing, but its message is late and generic, so `stop_server` and `suspend_server` check the server's type during their preview — before a token is minted — and name the tool to use instead.

**Example (stop, step 1 — preview):**

```json
{
  "name": "stop_server",
  "arguments": { "datacenter_id": "12345678-1234-1234-1234-123456789012", "server_id": "87654321-4321-4321-4321-210987654321" }
}
```

**API Reference:** [datacentersServersStartPost](https://api.ionos.com/docs/cloud/v6/#tag/Servers/operation/datacentersServersStartPost), [datacentersServersStopPost](https://api.ionos.com/docs/cloud/v6/#tag/Servers/operation/datacentersServersStopPost), [datacentersServersRebootPost](https://api.ionos.com/docs/cloud/v6/#tag/Servers/operation/datacentersServersRebootPost), [datacentersServersSuspendPost](https://api.ionos.com/docs/cloud/v6/#tag/Servers/operation/datacentersServersSuspendPost), [datacentersServersResumePost](https://api.ionos.com/docs/cloud/v6/#tag/Servers/operation/datacentersServersResumePost), [datacentersServersUpgradePost](https://api.ionos.com/docs/cloud/v6/#tag/Servers/operation/datacentersServersUpgradePost)

---

## attach_server_volume

Attaches an existing volume to a server. Requires `IONOS_MCP_TOOL_SCOPE` to include `write`. Single call.

The volume must already exist (`create_volume`) and be in the same data center as the server. A newly created server has no storage, so this is what gives it a disk to boot from.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center. |
| `server_id` | string | Yes | The ID of the server. |
| `volume_id` | string | Yes | The ID of an existing volume in the same data center. |

**API Reference:** [datacentersServersVolumesPost](https://api.ionos.com/docs/cloud/v6/#tag/Servers/operation/datacentersServersVolumesPost)

---

## detach_server_volume

Detaches a volume from a server. Requires `IONOS_MCP_TOOL_SCOPE` to include `destructive`. Two-phase.

The volume is **not** deleted — it survives as an unattached volume and **keeps incurring cost** until removed with `delete_volume`. Detaching a server's boot volume leaves it unable to boot.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center. |
| `server_id` | string | Yes | The ID of the server. |
| `volume_id` | string | Yes | The ID of the volume to detach. |
| `confirmation_token` | string | No | Omit on the first call for a preview + token; pass it on the second call to detach. |

**API Reference:** [datacentersServersVolumesDelete](https://api.ionos.com/docs/cloud/v6/#tag/Servers/operation/datacentersServersVolumesDelete)

---

## Not currently available: CD-ROM attach/detach

`attach_server_cdrom` is deliberately not exposed. Attaching an existing resource is expressed as a request body carrying only its id, but `Image.Properties` is a non-pointer field in the Go SDK whose serializer runs unconditionally, so the smallest body the SDK can produce is `{"id":"…","properties":{"licenceType":""}}` — property values the caller never supplied. The request builder accepts only the typed struct, so there is no way to send a correct body without hand-rolling the HTTP call. This needs a fix in the SDK templates (attach-by-reference should model the body as an id-only object) before the tool can ship.

The same applies to `attach_lan_nic`, which would send `{"id":"…","properties":{"lan":0}}`. That one is redundant in any case: use `update_nic` with an explicit `lan` to move a NIC onto a LAN.

---

## Changing which volume a server boots from

Attaching a volume does **not** make it the boot device, and detaching the current boot volume **clears the server's boot setting** — leaving a server that will not boot until you set it again. Two tools control this:

| Tool | Field | Use |
|------|-------|-----|
| `update_server` | `boot_volume_id` | Point the server at an already-attached volume. One call, no coordination needed. **Prefer this.** |
| `update_volume` | `boot_order` | `PRIMARY`, `NONE` or `AUTO` on the volume itself. `PRIMARY` requires *every* other volume on the server to be `NONE` first, so it takes several calls in the right order. |

`boot_volume_id` maps to the API's `properties.bootVolume` reference on a server `PATCH`. The volume must already be attached to the server. Reboot the server for the change to take effect.

### Swapping a server's disk

Do it in this order, which never leaves the server unbootable:

1. `create_volume` — create the replacement (with the image you want)
2. `attach_server_volume` — attach it to the server
3. `update_server` with `boot_volume_id` — make it the boot device
4. `reboot_server` — boot from it
5. `detach_server_volume` — remove the old volume, once you are satisfied
6. `delete_volume` — the detached volume survives and keeps incurring cost until you delete it

Detaching **before** step 3 clears the boot setting, so the server cannot boot until you complete it.

**API Reference:** [datacentersServersPatch](https://api.ionos.com/docs/cloud/v6/#tag/Servers/operation/datacentersServersPatch), [datacentersVolumesPatch](https://api.ionos.com/docs/cloud/v6/#tag/Volumes/operation/datacentersVolumesPatch)
