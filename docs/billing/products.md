---
subcategory: "Billing"
page_title: "Products"
description: |-
  Tool for searching the IONOS CLOUD product and pricing catalog.
---

# Products

## list_billing_products

Searches the IONOS CLOUD product and pricing catalog by keyword. Returns all non-deprecated products whose description matches the filter keyword.

**Only call this tool when the user has explicitly specified a product or category.** If the user asks a broad question like "what are the prices" or "show me all products", ask which specific product or category they are interested in before calling. Use a specific filter (e.g. `RAM` rather than `server`) to keep the result set small.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `contract` | integer | Yes | Contract number from `get_billing_profile` |
| `filter` | string | Yes | Keyword to filter products by description. Examples: `RAM`, `core`, `storage`, `Kubernetes`, `Postgres`, `network`, `Windows` |

**Example:**

```json
{
  "name": "list_billing_products",
  "arguments": {
    "contract": 12345678,
    "filter": "RAM"
  }
}
```

**Response fields:**

| Field | Description |
|-------|-------------|
| `appliedFilter` | The filter keyword used |
| `matchCount` | Number of matching products returned |
| `products` | Array of matching product entries |

**API Reference:** [ProductsGet](https://api.ionos.com/docs/billing/v3/#tag/Products/operation/ProductsGet)
