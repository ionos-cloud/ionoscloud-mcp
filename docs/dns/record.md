---
subcategory: "DNS"
page_title: "Record"
description: |-
  Tools for listing, inspecting, creating, updating and deleting DNS records in IONOS CLOUD.
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

---

## create_dns_record

Creates one record in a primary DNS zone. Requires `IONOS_MCP_TOOL_SCOPE` to include `write`.

Two-phase: call first without `confirmation_token` to get a preview and a one-time token, then call again with the token and the same `name`, `type` and `content`.

Pass an empty `name` for a record on the zone apex. The API answers `409` when the record conflicts with an existing one — a CNAME cannot coexist with another record of the same name (RFC 1034 §3.6.2), and a TXT record cannot carry a second SPF entry.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `zone_id` | string | Yes | The ID of the zone to add the record to |
| `name` | string | Yes | The record name relative to the zone, e.g. `www`. Empty string for the zone apex. |
| `type` | string | Yes | `A`, `AAAA`, `CNAME`, `ALIAS`, `MX`, `NS`, `SRV`, `TXT`, `CAA`, `SSHFP`, `TLSA`, `SMIMEA`, `DS`, `HTTPS`, `SVCB`, `OPENPGPKEY`, `CERT`, `URI`, `RP` or `LOC` |
| `content` | string | Yes | The record value, e.g. `192.0.2.1` |
| `ttl` | integer | No | Time to live in seconds, 60–604800 (default 3600) |
| `priority` | integer | No | 0–65535. Required for `MX`, `SRV` and `URI`; ignored for every other type. |
| `enabled` | boolean | No | Whether the record is visible for lookup (default `true`) |
| `confirmation_token` | string | No | Omit on the first call; pass the minted token on the second |

**Example:**

```json
{
  "name": "create_dns_record",
  "arguments": {
    "zone_id": "12345678-1234-1234-1234-123456789012",
    "name": "www",
    "type": "A",
    "content": "192.0.2.1",
    "ttl": 3600
  }
}
```

**API Reference:** [zonesRecordsPost](https://api.ionos.com/docs/dns/v1/#tag/Records/operation/zonesRecordsPost)

---

## update_dns_record

Updates a DNS record's content, TTL, priority, or whether it is published. Requires `IONOS_MCP_TOOL_SCOPE` to include `write`. Single call.

The record's `name` and `type` are immutable here — changing either is a delete plus a create, and can trip the API's conflict rules. Because the endpoint is a PUT that replaces the record's properties, every field you omit is read from the record and sent back explicitly rather than left out, which would let the API re-apply its own defaults (a missing `ttl` would reset to 3600).

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `zone_id` | string | Yes | The ID of the zone the record belongs to |
| `record_id` | string | Yes | The ID of the record to update |
| `content` | string | No | A new record value. Omit to keep the current one. |
| `ttl` | integer | No | A new TTL in seconds, 60–604800. Omit to keep the current one. |
| `priority` | integer | No | A new priority, 0–65535. Omit to keep the current one. |
| `enabled` | boolean | No | Set `false` to hide the record from lookups. Omit to keep the current setting. |

**Example:**

```json
{
  "name": "update_dns_record",
  "arguments": {
    "zone_id": "12345678-1234-1234-1234-123456789012",
    "record_id": "90d81ac0-3a30-44d4-95a5-12959effa6ee",
    "content": "192.0.2.2"
  }
}
```

**API Reference:** [zonesRecordsPut](https://api.ionos.com/docs/dns/v1/#tag/Records/operation/zonesRecordsPut)

---

## delete_dns_record

Deletes one record from a primary DNS zone. Requires `IONOS_MCP_TOOL_SCOPE` to include `destructive`.

Two-phase: call first without `confirmation_token` to get a preview and a one-time token, then call again with the token. Irreversible. Resolvers may keep serving the old answer until its TTL expires.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `zone_id` | string | Yes | The ID of the zone the record belongs to |
| `record_id` | string | Yes | The ID of the record to delete |
| `confirmation_token` | string | No | Omit on the first call; pass the minted token on the second |

**Example:**

```json
{
  "name": "delete_dns_record",
  "arguments": {
    "zone_id": "12345678-1234-1234-1234-123456789012",
    "record_id": "90d81ac0-3a30-44d4-95a5-12959effa6ee"
  }
}
```

**API Reference:** [zonesRecordsDelete](https://api.ionos.com/docs/dns/v1/#tag/Records/operation/zonesRecordsDelete)
