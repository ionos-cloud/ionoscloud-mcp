---
subcategory: "DNS"
page_title: "DNSSEC"
description: |-
  Tools for inspecting DNSSEC keys on DNS zones in IONOS Cloud.
---

# DNSSEC

## list_dns_zone_dnssec_keys

Lists DNSSEC keys for a specific DNS zone.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `zone_id` | string | Yes | The ID of the DNS zone |

**Example:**

```json
{
  "name": "list_dns_zone_dnssec_keys",
  "arguments": {
    "zone_id": "12345678-1234-1234-1234-123456789012"
  }
}
```

**API Reference:** [zonesKeysGet](https://api.ionos.com/docs/dns/v1/#tag/DNSSEC/operation/zonesKeysGet)
