---
subcategory: "Billing"
page_title: "Usage"
description: |-
  Tools for retrieving aggregated resource usage from IONOS CLOUD.
---

# Usage

Usage data provides aggregated metered quantities (CPU hours, GB-hours, etc.) per datacenter for a billing period. Higher-level view than EVN — totals rather than individual provisioning intervals.

## Compacted response shape

Both usage tools return the same compacted shape. Zero-quantity meters are dropped by default; `meter_definitions` is hoisted to the top level.

**Response fields:**

| Field | Type | Description |
|-------|------|-------------|
| `start_date` | string | Window start (YYYY-MM-DD) |
| `end_date` | string | Window end (YYYY-MM-DD) |
| `contract_id` | string | Contract number |
| `meter_definitions` | object | Map of `meter_id` → human description. Trimmed to only `meter_id`s present in the emitted rows. |
| `datacenters` | array | Per-datacenter meter totals |

**Per-datacenter:**

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Datacenter UUID |
| `name` | string | Datacenter name |
| `location` | string | Region code (e.g. `de/fra`) |
| `meters` | array | Per-SKU totals |

**Per-meter:**

| Field | Type | Description |
|-------|------|-------------|
| `meter_id` | string | Product code (look up description in `meter_definitions`) |
| `quantity` | string | Total quantity (string to preserve precision; numeric strings filter as 0 when zero) |
| `unit` | string | Unit string |

## Compaction flags

| Name | Type | Default | Description |
|------|------|---------|-------------|
| `include_zero` | boolean | `false` | Include meters with quantity 0 |
| `datacenter_id` | string | — | (list_billing_usage only) scope to a single datacenter |

Type/region/group_by flags do not apply — usage meters carry no `type`, `region`, or per-resource breakdown.

---

## list_billing_usage

Gets aggregated resource usage for the current billing period.

**Required parameters:** `contract`.

**Example:**

```json
{
  "name": "list_billing_usage",
  "arguments": {
    "contract": 12345678,
    "datacenter_id": "12345678-1234-1234-1234-123456789012"
  }
}
```

**API Reference:** [UsageGet](https://api.ionos.com/docs/billing/v3/#tag/Usage/operation/UsageGet)

---

## get_billing_usage_by_datacenter

Gets aggregated resource usage for a specific datacenter in the current billing period. Use `list_billing_usage` first to find datacenter IDs.

**Required parameters:** `contract`, `datacenter_id`.

**Example:**

```json
{
  "name": "get_billing_usage_by_datacenter",
  "arguments": {
    "contract": 12345678,
    "datacenter_id": "12345678-1234-1234-1234-123456789012",
    "include_zero": true
  }
}
```

**API Reference:** [UsageFindByDatacenter](https://api.ionos.com/docs/billing/v3/#tag/Usage/operation/UsageFindByDatacenter)
