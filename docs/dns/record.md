---
subcategory: "DNS"
page_title: "Record"
description: |-
  Tools for listing and inspecting DNS records in IONOS Cloud.
---

# DNS Records

## list_dns_records

Lists all DNS records across all zones.

**Parameters:** None

**Example:**

```json
{
  "name": "list_dns_records",
  "arguments": {}
}
```

**API Reference:** [recordsGet](https://api.ionos.com/docs/dns/v1/#tag/Records/operation/recordsGet)

---

## list_dns_zone_records

Lists all DNS records in a specific zone.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `zone_id` | string | Yes | The ID of the DNS zone |

**Example:**

```json
{
  "name": "list_dns_zone_records",
  "arguments": {
    "zone_id": "12345678-1234-1234-1234-123456789012"
  }
}
```

**API Reference:** [zonesRecordsGet](https://api.ionos.com/docs/dns/v1/#tag/Records/operation/zonesRecordsGet)

---

## get_dns_record

Gets detailed information about a specific DNS record in a zone.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `zone_id` | string | Yes | The ID of the DNS zone |
| `record_id` | string | Yes | The ID of the DNS record |

**Example:**

```json
{
  "name": "get_dns_record",
  "arguments": {
    "zone_id": "12345678-1234-1234-1234-123456789012",
    "record_id": "87654321-4321-4321-4321-210987654321"
  }
}
```

**API Reference:** [zonesRecordsFindById](https://api.ionos.com/docs/dns/v1/#tag/Records/operation/zonesRecordsFindById)

---

## list_dns_secondary_zone_records

Lists all DNS records in a specific secondary zone.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `secondary_zone_id` | string | Yes | The ID of the secondary DNS zone |

**Example:**

```json
{
  "name": "list_dns_secondary_zone_records",
  "arguments": {
    "secondary_zone_id": "33333333-3333-3333-3333-333333333333"
  }
}
```

**API Reference:** [secondaryzonesRecordsGet](https://api.ionos.com/docs/dns/v1/#tag/Records/operation/secondaryzonesRecordsGet)
