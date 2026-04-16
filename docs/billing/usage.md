---
subcategory: "Billing"
page_title: "Usage"
description: |-
  Tools for retrieving aggregated resource usage from IONOS Cloud.
---

# Usage

Usage data provides aggregated metered quantities (CPU hours, GB-hours, etc.) per datacenter for a billing period. This is a higher-level view than EVN — it shows totals rather than individual provisioning intervals.

## list_billing_usage

Gets aggregated resource usage for the current billing period.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `contract` | integer | Yes | Contract number from `get_billing_profile` |

**Example:**

```json
{
  "name": "list_billing_usage",
  "arguments": {
    "contract": 12345678
  }
}
```

**API Reference:** [UsageGet](https://api.ionos.com/docs/billing/v3/#tag/Usage/operation/UsageGet)

---

## get_billing_usage_by_datacenter

Gets aggregated resource usage for a specific datacenter in the current billing period. Use `list_billing_usage` first to find datacenter IDs.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `contract` | integer | Yes | Contract number from `get_billing_profile` |
| `datacenter_id` | string | Yes | The VDC UUID (from `list_billing_usage`) |

**Example:**

```json
{
  "name": "get_billing_usage_by_datacenter",
  "arguments": {
    "contract": 12345678,
    "datacenter_id": "12345678-1234-1234-1234-123456789012"
  }
}
```

**API Reference:** [UsageFindByDatacenter](https://api.ionos.com/docs/billing/v3/#tag/Usage/operation/UsageFindByDatacenter)
