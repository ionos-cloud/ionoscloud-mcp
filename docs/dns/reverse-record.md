---
subcategory: "DNS"
page_title: "Reverse Record"
description: |-
  Tools for listing and inspecting reverse DNS records in IONOS Cloud.
---

# Reverse DNS Records

## list_dns_reverse_records

Lists all reverse DNS records in your IONOS Cloud account.

**Parameters:** None

**Example:**

```json
{
  "name": "list_dns_reverse_records",
  "arguments": {}
}
```

**API Reference:** [reverserecordsGet](https://api.ionos.com/docs/dns/v1/#tag/ReverseRecords/operation/reverserecordsGet)

---

## get_dns_reverse_record

Gets detailed information about a specific reverse DNS record.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `reverse_record_id` | string | Yes | The ID of the reverse DNS record |

**Example:**

```json
{
  "name": "get_dns_reverse_record",
  "arguments": {
    "reverse_record_id": "44444444-4444-4444-4444-444444444444"
  }
}
```

**API Reference:** [reverserecordsFindById](https://api.ionos.com/docs/dns/v1/#tag/ReverseRecords/operation/reverserecordsFindById)
