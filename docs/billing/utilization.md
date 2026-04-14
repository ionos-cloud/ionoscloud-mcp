---
subcategory: "Billing"
page_title: "Utilization"
description: |-
  Tools for retrieving high-granularity resource utilization data from IONOS Cloud.
---

# Utilization

Utilization data provides per-resource metrics (CPU, RAM, storage, DNS) grouped by datacenter. It is more granular than usage — each record tracks an individual resource instance rather than aggregated totals.

## billing_utilization

Gets resource utilization for the current billing period.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `contract` | integer | Yes | Contract number from `billing_profile` |

**Example:**

```json
{
  "name": "billing_utilization",
  "arguments": {
    "contract": 12345678
  }
}
```

**API Reference:** [UtilizationGet](https://api.ionos.com/docs/billing/v3/#tag/Utilization/operation/UtilizationGet)

---

## billing_utilization_by_period

Gets resource utilization for a specific billing period. One month per call.

If the user requests a range longer than one month, calculate the number of monthly calls required, inform the user, and ask for permission before proceeding.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `contract` | integer | Yes | Contract number from `billing_profile` |
| `period` | string | Yes | Billing period in `YYYY-MM` format (e.g. `2026-04`) |

**Example:**

```json
{
  "name": "billing_utilization_by_period",
  "arguments": {
    "contract": 12345678,
    "period": "2026-04"
  }
}
```

**API Reference:** [UtilizationFindByPeriod](https://api.ionos.com/docs/billing/v3/#tag/Utilization/operation/UtilizationFindByPeriod)

---

## billing_utilization_daily

Gets resource utilization for a specific date. Use this for day-level analysis within a month — it returns less data than a full-month query.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `contract` | integer | Yes | Contract number from `billing_profile` |
| `date` | string | Yes | Date in `YYYY-MM-DD` format (e.g. `2026-04-15`) |

**Example:**

```json
{
  "name": "billing_utilization_daily",
  "arguments": {
    "contract": 12345678,
    "date": "2026-04-15"
  }
}
```

**API Reference:** [UtilizationDailyFindByDate](https://api.ionos.com/docs/billing/v3/#tag/Utilization/operation/UtilizationDailyFindByDate)
