---
subcategory: "Billing"
page_title: "Traffic"
description: |-
  Tools for retrieving network traffic billing data from IONOS CLOUD.
---

# Traffic

Network traffic data shows inbound and outbound byte totals per datacenter and per NIC. The response returns the structured `trafficObj` field together with `metadata`, while CSV and array duplicate representations from the API are dropped.

## list_billing_traffic

Gets network traffic data for the current billing month.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `contract` | integer | Yes | Contract number from `get_billing_profile` |

**Example:**

```json
{
  "name": "list_billing_traffic",
  "arguments": {
    "contract": 12345678
  }
}
```

**API Reference:** [TrafficGet](https://api.ionos.com/docs/billing/v3/#tag/Traffic/operation/TrafficGet)

---

## list_billing_traffic_by_period

Gets network traffic data for a specific billing period. One month per call.

If the user requests a range longer than one month, calculate the number of monthly calls required, inform the user, and ask for permission before proceeding.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `contract` | integer | Yes | Contract number from `get_billing_profile` |
| `period` | string | Yes | Billing period in `YYYY-MM` format (e.g. `2026-04`) |

**Example:**

```json
{
  "name": "list_billing_traffic_by_period",
  "arguments": {
    "contract": 12345678,
    "period": "2026-04"
  }
}
```

**API Reference:** [TrafficFindByPeriod](https://api.ionos.com/docs/billing/v3/#tag/Traffic/operation/TrafficFindByPeriod)
