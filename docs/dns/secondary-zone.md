---
subcategory: "DNS"
page_title: "Secondary Zone"
description: |-
  Tools for listing, inspecting, creating, updating and deleting secondary DNS zones in IONOS CLOUD.
---

# Secondary DNS Zones

## list_dns_secondary_zones

Lists all secondary DNS zones in your IONOS CLOUD account.

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

---

## create_dns_secondary_zone

Creates one secondary DNS zone, which mirrors a zone hosted on nameservers you run elsewhere. Requires `IONOS_MCP_TOOL_SCOPE` to include `write`.

Two-phase: call first without `confirmation_token` to get a preview and a one-time token, then call again with the token and the same `zone_name`.

Its records are transferred from the primary IPs, not authored through the API — use `start_dns_zone_transfer` to trigger a transfer and `get_dns_secondary_zone_axfr` to check it.

> Whitelist IONOS's notify sources on your primary nameservers, or the transfer will not work: IPv4 `212.227.123.25` and IPv6 `2001:8d8:fe:53::5cd:25`.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `zone_name` | string | Yes | The zone name, e.g. `example.com` |
| `primary_ips` | array of strings | Yes | IPv4 or IPv6 addresses of the primary nameservers. At least one, no duplicates. |
| `description` | string | No | Free-text description of what the zone is for |
| `confirmation_token` | string | No | Omit on the first call; pass the minted token on the second |

**Example:**

```json
{
  "name": "create_dns_secondary_zone",
  "arguments": {
    "zone_name": "example.com",
    "primary_ips": ["1.2.3.4", "5.6.7.8"]
  }
}
```

**API Reference:** [secondaryzonesPost](https://api.ionos.com/docs/dns/v1/#tag/SecondaryZones/operation/secondaryzonesPost)

---

## update_dns_secondary_zone

Updates a secondary DNS zone's primary nameserver IPs or its description. Requires `IONOS_MCP_TOOL_SCOPE` to include `write`. Single call.

The zone name is immutable and is read and sent back unchanged. `primary_ips` **replaces** the current list when supplied, so include every IP that should remain.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `secondary_zone_id` | string | Yes | The ID of the secondary zone to update |
| `primary_ips` | array of strings | No | Replaces the primary nameserver IPs. Omit to keep the current list. |
| `description` | string | No | A new description. Omit to keep the current one. |

**Example:**

```json
{
  "name": "update_dns_secondary_zone",
  "arguments": {
    "secondary_zone_id": "e74d0d15-f567-4b7b-9069-26ee1f93bae3",
    "primary_ips": ["1.2.3.4", "9.9.9.9"]
  }
}
```

**API Reference:** [secondaryzonesPut](https://api.ionos.com/docs/dns/v1/#tag/SecondaryZones/operation/secondaryzonesPut)

---

## delete_dns_secondary_zone

Deletes a secondary DNS zone. Requires `IONOS_MCP_TOOL_SCOPE` to include `destructive`.

Two-phase: call first without `confirmation_token` to get a preview and a one-time token, then call again with the token. Irreversible, but the zone's records live on your primary nameservers, so only the IONOS-side mirror is removed.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `secondary_zone_id` | string | Yes | The ID of the secondary zone to delete |
| `confirmation_token` | string | No | Omit on the first call; pass the minted token on the second |

**Example:**

```json
{
  "name": "delete_dns_secondary_zone",
  "arguments": {
    "secondary_zone_id": "e74d0d15-f567-4b7b-9069-26ee1f93bae3"
  }
}
```

**API Reference:** [secondaryzonesDelete](https://api.ionos.com/docs/dns/v1/#tag/SecondaryZones/operation/secondaryzonesDelete)

---

## start_dns_zone_transfer

Triggers a zone transfer (AXFR) for a secondary DNS zone, pulling its records from the primary nameservers. Requires `IONOS_MCP_TOOL_SCOPE` to include `write`.

Single call — a transfer only refreshes the IONOS-side copy, so there is no confirmation step. Follow it with `get_dns_secondary_zone_axfr`, which reports the status of each primary IP separately.

> Whitelist IONOS's notify sources on your primary nameservers: IPv4 `212.227.123.25` and IPv6 `2001:8d8:fe:53::5cd:25`.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `secondary_zone_id` | string | Yes | The ID of the secondary zone to transfer |

**Example:**

```json
{
  "name": "start_dns_zone_transfer",
  "arguments": {
    "secondary_zone_id": "e74d0d15-f567-4b7b-9069-26ee1f93bae3"
  }
}
```

**API Reference:** [secondaryzonesAxfrPut](https://api.ionos.com/docs/dns/v1/#tag/SecondaryZones/operation/secondaryzonesAxfrPut)
