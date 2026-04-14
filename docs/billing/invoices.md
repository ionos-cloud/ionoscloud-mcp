---
subcategory: "Billing"
page_title: "Invoices"
description: |-
  Tools for listing and inspecting IONOS Cloud invoices.
---

# Invoices

## billing_invoices

Lists all invoices for your IONOS Cloud contract. Returns invoice IDs, dates, and totals. Use the returned IDs with `billing_invoice` to get line-item detail.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `contract` | integer | Yes | Contract number from `billing_profile` |

**Example:**

```json
{
  "name": "billing_invoices",
  "arguments": {
    "contract": 12345678
  }
}
```

**API Reference:** [InvoicesGet](https://api.ionos.com/docs/billing/v3/#tag/Invoices/operation/InvoicesGet)

---

## billing_invoices_by_period

Lists invoices for a specific billing period (`YYYY-MM`). One month per call.

If the user requests a range longer than one month, calculate the number of monthly calls required, inform the user, and ask for permission before proceeding.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `period` | string | Yes | Billing period in `YYYY-MM` format (e.g. `2026-04`) |

**Example:**

```json
{
  "name": "billing_invoices_by_period",
  "arguments": {
    "period": "2026-04"
  }
}
```

**API Reference:** [InvoicesFindByPeriod](https://api.ionos.com/docs/billing/v3/#tag/Invoices/operation/InvoicesFindByPeriod)

---

## billing_invoice

Gets the detailed line-item breakdown for a specific invoice. Returns per-datacenter meter entries with quantities and amounts.

Use `billing_invoices` first to find available invoice IDs.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `contract` | integer | Yes | Contract number from `billing_profile` |
| `invoice_id` | string | Yes | The invoice ID (e.g. `GY00350536`) |

**Example:**

```json
{
  "name": "billing_invoice",
  "arguments": {
    "contract": 12345678,
    "invoice_id": "GY00350536"
  }
}
```

**API Reference:** [InvoicesFindById](https://api.ionos.com/docs/billing/v3/#tag/Invoices/operation/InvoicesFindById)
