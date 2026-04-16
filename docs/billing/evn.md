---
subcategory: "Billing"
page_title: "EVN"
description: |-
  Tools for retrieving provisioning itemized data (EVN) from IONOS Cloud.
---

# EVN (Provisioning Itemized Data)

EVN (Provisioning Itemized Data) provides per-resource usage intervals grouped by datacenter. Each record tracks when a resource (server, volume, IP, etc.) was provisioned and for how long.

The response drops the `evnCSV` duplicate field from the API and returns the structured `metadata` and `datacenters` fields.

## list_billing_evn

Gets EVN data for the current billing month.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `contract` | integer | Yes | Contract number from `get_billing_profile` |

**Example:**

```json
{
  "name": "list_billing_evn",
  "arguments": {
    "contract": 12345678
  }
}
```

**API Reference:** [EvnGet](https://api.ionos.com/docs/billing/v3/#tag/EVN/operation/EvnGet)

---

## list_billing_evn_by_period

Gets EVN data for a specific billing period. One month per call.

If the user requests a range longer than one month, calculate the number of monthly calls required, inform the user, and ask for permission before proceeding.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `contract` | integer | Yes | Contract number from `get_billing_profile` |
| `period` | string | Yes | Billing period in `YYYY-MM` format (e.g. `2026-04`) |

**Example:**

```json
{
  "name": "list_billing_evn_by_period",
  "arguments": {
    "contract": 12345678,
    "period": "2026-04"
  }
}
```

**API Reference:** [EvnFindByPeriod](https://api.ionos.com/docs/billing/v3/#tag/EVN/operation/EvnFindByPeriod)
