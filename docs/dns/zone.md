---
subcategory: "DNS"
page_title: "Zone"
description: |-
  Tools for listing and inspecting DNS zones in IONOS CLOUD.
---

# DNS Zones

## list_dns_zones

Lists all DNS zones in your IONOS CLOUD account.

**Parameters:** None

**Example:**

```json
{
  "name": "list_dns_zones",
  "arguments": {}
}
```

**API Reference:** [zonesGet](https://api.ionos.com/docs/dns/v1/#tag/Zones/operation/zonesGet)

---

## get_dns_zone

Gets detailed information about a specific DNS zone.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `zone_id` | string | Yes | The ID of the DNS zone |

**Example:**

```json
{
  "name": "get_dns_zone",
  "arguments": {
    "zone_id": "12345678-1234-1234-1234-123456789012"
  }
}
```

**API Reference:** [zonesFindById](https://api.ionos.com/docs/dns/v1/#tag/Zones/operation/zonesFindById)

---

## get_dns_zone_file

Gets the zone file (BIND format) for a specific DNS zone.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `zone_id` | string | Yes | The ID of the DNS zone |

**Example:**

```json
{
  "name": "get_dns_zone_file",
  "arguments": {
    "zone_id": "12345678-1234-1234-1234-123456789012"
  }
}
```

**API Reference:** [zonesZonefileGet](https://api.ionos.com/docs/dns/v1/#tag/ZoneFiles/operation/zonesZonefileGet)
