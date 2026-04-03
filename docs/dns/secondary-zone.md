---
subcategory: "DNS"
page_title: "Secondary Zone"
description: |-
  Tools for listing and inspecting secondary DNS zones in IONOS Cloud.
---

# Secondary DNS Zones

## list_dns_secondary_zones

Lists all secondary DNS zones in your IONOS Cloud account.

**Parameters:** None

**Example:**

```json
{
  "name": "list_dns_secondary_zones",
  "arguments": {}
}
```

**API Reference:** [secondaryzonesGet](https://api.ionos.com/docs/dns/v1/#tag/SecondaryZones/operation/secondaryzonesGet)

---

## get_dns_secondary_zone

Gets detailed information about a specific secondary DNS zone.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `secondary_zone_id` | string | Yes | The ID of the secondary DNS zone |

**Example:**

```json
{
  "name": "get_dns_secondary_zone",
  "arguments": {
    "secondary_zone_id": "33333333-3333-3333-3333-333333333333"
  }
}
```

**API Reference:** [secondaryzonesFindById](https://api.ionos.com/docs/dns/v1/#tag/SecondaryZones/operation/secondaryzonesFindById)

---

## get_dns_secondary_zone_axfr

Gets the zone transfer (AXFR) status of a specific secondary DNS zone.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `secondary_zone_id` | string | Yes | The ID of the secondary DNS zone |

**Example:**

```json
{
  "name": "get_dns_secondary_zone_axfr",
  "arguments": {
    "secondary_zone_id": "33333333-3333-3333-3333-333333333333"
  }
}
```

**API Reference:** [secondaryzonesAxfrGet](https://api.ionos.com/docs/dns/v1/#tag/SecondaryZones/operation/secondaryzonesAxfrGet)
