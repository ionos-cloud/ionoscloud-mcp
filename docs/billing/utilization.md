---
subcategory: "Billing"
page_title: "Utilization"
description: |-
  Tools for retrieving high-granularity resource utilization data from IONOS CLOUD.
---

# Utilization

Utilization data provides per-resource metrics (CPU, RAM, storage, DNS, DBaaS, etc.) grouped by datacenter. More granular than usage — each record tracks an individual resource instance.

## Compacted response shape

All three utilization tools return the same compacted shape (raw SDK output is ~1.27 MB on a typical contract — unusable for LLMs).

**Defaults that shape the response:**

- Zero-quantity meters are dropped (set `include_zero=true` to keep them).
- `meter_definitions` is hoisted to the top level — one description per `meter_id` instead of repeated on every meter.
- Empty datacenters (after meter filtering) are omitted.
- Redundant fields removed: `exists` (always null), per-meter `from`/`to` (already at top level), meter event UUIDs, null `server_id`, empty `name`.
- `quantity.{quantity, unit}` is flattened to `quantity` and `unit`.

**Response fields:**

| Field | Type | Description |
|-------|------|-------------|
| `start_date` | string | Window start (YYYY-MM-DD) |
| `end_date` | string | Window end (YYYY-MM-DD) |
| `contract_id` | string | Contract number |
| `meter_definitions` | object | Map of `meter_id` → human description (e.g. `"DBMP1000": "1h of MongoDB Playground first instance"`) |
| `datacenters` | array | Per-datacenter meter rows (see below) |

**Per-datacenter:**

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Datacenter UUID |
| `name` | string | Datacenter name |
| `meters` | array | Per-resource meter rows |

**Per-meter:**

| Field | Type | Description |
|-------|------|-------------|
| `meter_id` | string | Product code (look up description in `meter_definitions`) |
| `type` | string | Category (`SERVER`, `DBAAS`, `DNS`, `DB`, …) |
| `region` | string | Region code (e.g. `de/fra`) |
| `resource_id` | string | Resource UUID |
| `server_id` | string | Server UUID — omitted when null |
| `name` | string | Resource name — omitted when empty |
| `quantity` | number | Consumed amount |
| `unit` | string | Unit string (e.g. `1G*30Days`, `1hour`) |

## Compaction flags (all three tools)

| Name | Type | Default | Description |
|------|------|---------|-------------|
| `include_zero` | boolean | `false` | Include meters with quantity 0. Use `true` to find existing resources that didn't consume in the window. |
| `group_by` | string | `""` | Aggregation level: `""` (per-resource, default), `"meter"` (sum per SKU per datacenter), `"datacenter"` (sum per type per datacenter). Coarser groupings shrink output but lose detail. |
| `datacenter_id` | string | — | Scope to a single datacenter (VDC UUID). |
| `meter_types` | string[] | — | Filter to these meter type categories only (client-side). E.g. `["DBAAS","DNS"]`. |
| `regions` | string[] | — | Filter to these regions only (client-side). E.g. `["de/fra","es/vit"]`. |

**Expected sizes** (a representative ~225-datacenter contract):

| Mode | Size |
|------|------|
| Default (compacted, zero-filter) | ~325 KB |
| `group_by=meter` | ~80 KB |
| `group_by=datacenter` | ~10 KB |
| `datacenter_id` scope | 5–50 KB depending on DC |

---

## list_billing_utilization

Gets resource utilization for the current billing period.

**Required parameters:** `contract`.

**Example — find idle resources across the contract:**

```json
{
  "name": "list_billing_utilization",
  "arguments": {
    "contract": 12345678,
    "include_zero": true
  }
}
```

**Example — coarse cost breakdown:**

```json
{
  "name": "list_billing_utilization",
  "arguments": {
    "contract": 12345678,
    "group_by": "datacenter"
  }
}
```

**API Reference:** [UtilizationGet](https://api.ionos.com/docs/billing/v3/#tag/Utilization/operation/UtilizationGet)

---

## list_billing_utilization_by_period

Gets resource utilization for a specific billing period. One month per call.

If the user requests a range longer than one month, calculate the number of monthly calls required, inform the user, and ask for permission before proceeding.

**Required parameters:** `contract`, `period` (`YYYY-MM`).

**Example:**

```json
{
  "name": "list_billing_utilization_by_period",
  "arguments": {
    "contract": 12345678,
    "period": "2026-04",
    "group_by": "meter"
  }
}
```

**API Reference:** [UtilizationFindByPeriod](https://api.ionos.com/docs/billing/v3/#tag/Utilization/operation/UtilizationFindByPeriod)

---

## get_billing_utilization_daily

Gets resource utilization for a specific date. Use for day-level analysis within a month.

**Required parameters:** `contract`, `date` (`YYYY-MM-DD`).

**Example:**

```json
{
  "name": "get_billing_utilization_daily",
  "arguments": {
    "contract": 12345678,
    "date": "2026-04-15",
    "meter_types": ["DBAAS"]
  }
}
```

**API Reference:** [UtilizationDailyFindByDate](https://api.ionos.com/docs/billing/v3/#tag/Utilization/operation/UtilizationDailyFindByDate)
