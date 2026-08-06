---
subcategory: "Compute Engine"
page_title: "Image"
description: |-
  Tools for listing, inspecting, and (opt-in) updating and deleting images in IONOS CLOUD.
---

# Images

The `list_*` and `get_*` tools are always available. The write tools register only when `IONOS_MCP_TOOL_SCOPE` opts in (`write` enables update; `destructive` also enables delete). `delete_*` uses a two-phase confirmation: call once **without** `confirmation_token` to get a preview plus a one-time token, then call again **with** that token to execute.

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

---

## There is no create_image

The API exposes no operation to create an image: public images come from IONOS, and private ones are uploaded out of band ([ionosctl](https://github.com/ionos-cloud/ionosctl), FTP upload, or the DCD). Only **private images you uploaded** can be updated or deleted — public IONOS images are read-only, and the API rejects attempts to modify them.

---

## update_image

Updates a private image's name, description, licence type, cloud-init compatibility or capability flags. Requires `write`. Partial update.

**Omit `licence_type` to keep the current value.** A blank `licence_type` is rejected rather than applied.

| Name | Type | Description |
|------|------|-------------|
| `image_id` | string | **Required.** A private image only. |
| `name`, `description` | string | Metadata. |
| `licence_type` | string | OS type. Omit to keep the current value. |
| `cloud_init` | string | `NONE` or `V1`. |
| `expose_serial`, `require_legacy_bios` | boolean | Applied to volumes created from this image. |
| the ten hot-plug flags | boolean | Same wider set as snapshots — see [snapshot.md](snapshot.md). |

**API Reference:** [imagesPatch](https://api.ionos.com/docs/cloud/v6/#tag/Images/operation/imagesPatch)

---

## delete_image

Deletes a private image. Irreversible. Requires `destructive`. Two-phase.

Anything that references the image by ID stops being able to create volumes from it — Terraform configurations, provisioning scripts and VM autoscaling templates included. The preview lists the image's **aliases**, since those are how most callers refer to it, so you can see what breaks.

If the image is public the preview says so up front and explains that the API will reject the deletion, rather than letting you spend a token round trip discovering it.

**API Reference:** [imagesDelete](https://api.ionos.com/docs/cloud/v6/#tag/Images/operation/imagesDelete)
