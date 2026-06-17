---
subcategory: "Compute Engine"
page_title: "Contract"
description: |-
  Tools for inspecting contract and resource limit information in IONOS CLOUD.
---

# Contract

## get_contract

Gets contract and resource limit information for your IONOS CLOUD account.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `depth` | integer | No | Nesting depth of returned objects (0–5). |

**Example:**

```json
{
  "name": "get_contract",
  "arguments": {}
}
```

**API Reference:** [contractsGet](https://api.ionos.com/docs/cloud/v6/#tag/ContractResources/operation/contractsGet)
