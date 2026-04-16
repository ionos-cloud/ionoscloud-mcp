---
subcategory: "Billing"
page_title: "Profile"
description: |-
  Tool for retrieving the billing profile of your IONOS Cloud account.
---

# Profile

## get_billing_profile

Gets the billing profile for your IONOS Cloud account, including contract numbers and customer IDs.

Call this first before any other billing tool — the contract number in the response is required by all subsequent billing tools.

**Parameters:** None

**Example:**

```json
{
  "name": "get_billing_profile",
  "arguments": {}
}
```

**API Reference:** [ProfilesGet](https://api.ionos.com/docs/billing/v3/#tag/Profile/operation/ProfilesGet)
