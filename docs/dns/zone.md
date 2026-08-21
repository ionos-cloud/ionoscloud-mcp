---
subcategory: "DNS"
page_title: "Zone"
description: |-
  Tools for listing, inspecting, creating, updating and deleting DNS zones in IONOS CLOUD.
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

---

## create_dns_zone

Creates one primary DNS zone, with default NS and SOA records. Requires `IONOS_MCP_TOOL_SCOPE` to include `write`.

Two-phase: call first without `confirmation_token` to get a preview and a one-time token, then call again with the token and the same `zone_name`.

The zone does not serve your domain until you point the registrar at the nameservers in the response's `metadata.nameservers`.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `zone_name` | string | Yes | The zone name, e.g. `example.com` |
| `description` | string | No | Free-text description of what the zone is for |
| `enabled` | boolean | No | Whether the zone answers lookups (default `true`) |
| `confirmation_token` | string | No | Omit on the first call; pass the minted token on the second |

**Example:**

```json
{
  "name": "create_dns_zone",
  "arguments": {
    "zone_name": "example.com",
    "description": "Production zone"
  }
}
```

**API Reference:** [zonesPost](https://api.ionos.com/docs/dns/v1/#tag/Zones/operation/zonesPost)

---

## update_dns_zone

Updates a primary DNS zone's description, or enables/disables it. Requires `IONOS_MCP_TOOL_SCOPE` to include `write`. Single call.

The zone name is immutable: the endpoint is a PUT that replaces the zone's properties, so the tool reads the zone and sends the current name back unchanged. Disabling a zone stops it answering lookups for every record it contains, without deleting anything.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `zone_id` | string | Yes | The ID of the zone to update |
| `description` | string | No | A new description. Omit to keep the current one. |
| `enabled` | boolean | No | Set `false` to stop the zone answering lookups. Omit to keep the current setting. |

**Example:**

```json
{
  "name": "update_dns_zone",
  "arguments": {
    "zone_id": "12345678-1234-1234-1234-123456789012",
    "enabled": false
  }
}
```

**API Reference:** [zonesPut](https://api.ionos.com/docs/dns/v1/#tag/Zones/operation/zonesPut)

---

## delete_dns_zone

Deletes a primary DNS zone and every record it contains. Requires `IONOS_MCP_TOOL_SCOPE` to include `destructive`.

Two-phase: call first without `confirmation_token` to get a blast-radius preview (record count, and DNSSEC keys if the zone is signed) and a one-time token, then call again with the token. Irreversible — any domain still pointing at this zone's nameservers stops resolving.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `zone_id` | string | Yes | The ID of the zone to delete |
| `confirmation_token` | string | No | Omit on the first call; pass the minted token on the second |

**Example:**

```json
{
  "name": "delete_dns_zone",
  "arguments": {
    "zone_id": "12345678-1234-1234-1234-123456789012"
  }
}
```

**API Reference:** [zonesDelete](https://api.ionos.com/docs/dns/v1/#tag/Zones/operation/zonesDelete)

---

## import_dns_zone_file

Replaces **all** records in a primary DNS zone with the contents of a zone file in BIND format (RFC 1035). Requires `IONOS_MCP_TOOL_SCOPE` to include `destructive` — the HTTP method is a PUT, but every record currently in the zone is deleted, so the tool is classified by what it does rather than by its verb.

Two-phase: call first without `confirmation_token` to see how many existing records would be replaced and get a one-time token, then call again with the token. SOA and NS records in the file are accepted but ignored, so a file exported from another provider can be imported as-is. Returns the resulting record list.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `zone_id` | string | Yes | The ID of the zone to overwrite |
| `zone_file` | string | Yes | The zone file in BIND format (RFC 1035) |
| `confirmation_token` | string | No | Omit on the first call; pass the minted token on the second |

**Example:**

```json
{
  "name": "import_dns_zone_file",
  "arguments": {
    "zone_id": "12345678-1234-1234-1234-123456789012",
    "zone_file": "$ORIGIN example.com.\n$TTL 3600\nwww  IN  A  192.0.2.1\n"
  }
}
```

**API Reference:** [zonesZonefilePut](https://api.ionos.com/docs/dns/v1/#tag/ZoneFiles/operation/zonesZonefilePut)
