---
subcategory: "Compute Engine"
page_title: "Data Center"
description: |-
  Tools for listing, inspecting, and (opt-in) creating, updating, and deleting virtual data centers in IONOS CLOUD.
---

# Data Centers

The `list_*` and `get_*` tools are always available. The write tools — `create_datacenter`, `update_datacenter`, and `delete_datacenter` — register only when `IONOS_MCP_TOOL_SCOPE` opts in (`write` enables create/update; `destructive` also enables delete). `create_datacenter` and `delete_datacenter` use a two-phase confirmation: call once **without** `confirmation_token` to get a preview (delete returns a blast-radius summary) plus a one-time token, then call again **with** that token to execute.

## list_datacenters

Lists all virtual data centers in your IONOS CLOUD account.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `depth` | integer | No | Nesting depth of returned objects (0–5, default `1`). Depth 1 includes names and basic properties. |
| `filters` | object | No | Server-side property filters; e.g. `{"name":"prod","location":"de/fra"}`. If the result is empty, retry without filters — a filter typo or mismatch silently returns nothing. |

**Example:**

```json
{
  "name": "list_datacenters",
  "arguments": {}
}
```

**API Reference:** [datacentersGet](https://api.ionos.com/docs/cloud/v6/#tag/Data-Centers/operation/datacentersGet)

---

## get_datacenter

Gets detailed information about a specific virtual data center.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center |
| `depth` | integer | No | Nesting depth of returned objects (0–5). |

**Example:**

```json
{
  "name": "get_datacenter",
  "arguments": {
    "datacenter_id": "12345678-1234-1234-1234-123456789012"
  }
}
```

**API Reference:** [datacentersFindById](https://api.ionos.com/docs/cloud/v6/#tag/Data-Centers/operation/datacentersFindById)

---

## create_datacenter

Creates one virtual data center. Requires `IONOS_MCP_TOOL_SCOPE` to include `write`. Two-phase: the first call (no `confirmation_token`) returns a preview and a one-time token; a second call with that token creates exactly one data center (there is no bulk/batch parameter).

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | string | Yes | The name of the new data center. |
| `location` | string | Yes | Physical location, e.g. `de/fra`, `de/txl`, `us/las`, `us/ewr`, `gb/lhr`, `es/vit`, `fr/par`. Cannot be changed after creation. |
| `description` | string | No | An optional description (e.g. staging, production). |
| `sec_auth_protection` | boolean | No | If `true`, the data center requires extra protection (two-step verification). |
| `confirmation_token` | string | No | Omit on the first call to receive a preview + token; pass the token (with the same name and location) on the second call to create. Expires after a few minutes. |

**Example (step 1 — preview):**

```json
{
  "name": "create_datacenter",
  "arguments": { "name": "my-dc", "location": "de/fra" }
}
```

**Example (step 2 — execute):**

```json
{
  "name": "create_datacenter",
  "arguments": { "name": "my-dc", "location": "de/fra", "confirmation_token": "<token-from-step-1>" }
}
```

**API Reference:** [datacentersPost](https://api.ionos.com/docs/cloud/v6/#tag/Data-Centers/operation/datacentersPost)

---

## update_datacenter

Updates a data center's `name`, `description`, or protection flag. Requires `IONOS_MCP_TOOL_SCOPE` to include `write`. Applies a partial update (only the fields you provide); the `location` is immutable. Single call — no confirmation token.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center to update. |
| `name` | string | No | A new name for the data center. |
| `description` | string | No | A new description for the data center. |
| `sec_auth_protection` | boolean | No | Set the extra-protection (two-step verification) flag. |

**Example:**

```json
{
  "name": "update_datacenter",
  "arguments": {
    "datacenter_id": "12345678-1234-1234-1234-123456789012",
    "name": "renamed-dc"
  }
}
```

**API Reference:** [datacentersPatch](https://api.ionos.com/docs/cloud/v6/#tag/Data-Centers/operation/datacentersPatch)

---

## delete_datacenter

Deletes a data center **and every resource inside it** (servers, volumes, LANs, load balancers, and more). This is irreversible. Requires `IONOS_MCP_TOOL_SCOPE` to include `destructive`. Two-phase: the first call (no `confirmation_token`) returns a blast-radius preview and a one-time token; a second call with that token performs the deletion.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `datacenter_id` | string | Yes | The ID of the data center to delete. |
| `confirmation_token` | string | No | Omit on the first call to receive a blast-radius preview + token; pass the token on the second call to delete. The token is bound to this data center and expires after a few minutes. |

**Example (step 1 — preview):**

```json
{
  "name": "delete_datacenter",
  "arguments": { "datacenter_id": "12345678-1234-1234-1234-123456789012" }
}
```

**Example (step 2 — execute):**

```json
{
  "name": "delete_datacenter",
  "arguments": { "datacenter_id": "12345678-1234-1234-1234-123456789012", "confirmation_token": "<token-from-step-1>" }
}
```

**API Reference:** [datacentersDelete](https://api.ionos.com/docs/cloud/v6/#tag/Data-Centers/operation/datacentersDelete)
