---
subcategory: "Activity Log"
page_title: "Contracts"
description: |-
  Tool for listing contracts accessible via the IONOS CLOUD Activity Log API.
---

# Contracts

## list_activitylog_contracts

Lists contracts accessible for IONOS CLOUD activity log queries. Primarily useful for reseller and partner users managing multiple contracts. Single-contract users can skip this — the contract number is embedded in the JWT token and is also returned by `get_billing_profile`.

**Parameters:** none

**Example:**

```json
{
  "name": "list_activitylog_contracts",
  "arguments": {}
}
```

**Response fields:**

| Field | Type | Description |
|-------|------|-------------|
| `id` | integer | Contract number |
| `type` | string | Always `"contracts"` |
| `href` | string | API URL for this contract |

**API Reference:** [GetAvailableContracts](https://api.ionos.com/docs/activitylog/v1/#tag/Contracts/operation/GetAvailableContracts)
