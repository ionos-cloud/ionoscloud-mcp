---
subcategory: "DNS"
page_title: "DNSSEC"
description: |-
  Tools for inspecting, enabling and disabling DNSSEC on DNS zones in IONOS CLOUD.
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

---

## create_dns_zone_dnssec_key

Enables DNSSEC on a primary DNS zone, generating its signing keys and DNSKEY records. Requires `IONOS_MCP_TOOL_SCOPE` to include `write`.

Two-phase: call first without `confirmation_token` to get a preview and a one-time token, then call again with the token. A zone can hold only one DNSSEC configuration, so the API answers `409` if it is already signed.

After enabling, read the DS digest from `list_dns_zone_dnssec_keys` and add it at your domain's registrar. Until you do, the zone is signed but the chain of trust is not established.

Only `validity` is required; the rest default to the values below. `RSASHA256` is the only algorithm the API accepts, and `ksk_bits` must be greater than or equal to `zsk_bits`. The three `nsec3_*` fields are always sent, even in `NSEC` mode, because the API requires them.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `zone_id` | string | Yes | The ID of the zone to sign |
| `validity` | integer | Yes | Signature validity in days, 90–365 |
| `algorithm` | string | No | `RSASHA256` (default, and the only accepted value) |
| `ksk_bits` | integer | No | Key signing key length: 1024, 2048 or 4096 (default 4096). Must be ≥ `zsk_bits`. |
| `zsk_bits` | integer | No | Zone signing key length: 1024, 2048 or 4096 (default 2048) |
| `nsec_mode` | string | No | `NSEC` or `NSEC3` (default `NSEC3`) |
| `nsec3_iterations` | integer | No | 0–50 (default 0, as RFC 9276 recommends) |
| `nsec3_salt_bits` | integer | No | 64–128, a multiple of 8 (default 64) |
| `confirmation_token` | string | No | Omit on the first call; pass the minted token on the second |

**Example:**

```json
{
  "name": "create_dns_zone_dnssec_key",
  "arguments": {
    "zone_id": "12345678-1234-1234-1234-123456789012",
    "validity": 120
  }
}
```

**API Reference:** [zonesKeysPost](https://api.ionos.com/docs/dns/v1/#tag/DNSSEC/operation/zonesKeysPost)

---

## delete_dns_zone_dnssec_key

Disables DNSSEC on a primary DNS zone, removing its signing keys and DNSKEY records. Requires `IONOS_MCP_TOOL_SCOPE` to include `destructive`.

Two-phase: call first without `confirmation_token` to get a preview and a one-time token, then call again with the token. Irreversible — re-enabling generates new keys.

> **Remove the DS record at your registrar first.** If the parent zone still publishes a DS record for keys that no longer exist, validating resolvers answer `SERVFAIL` and the whole zone stops resolving.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `zone_id` | string | Yes | The ID of the zone to stop signing |
| `confirmation_token` | string | No | Omit on the first call; pass the minted token on the second |

**Example:**

```json
{
  "name": "delete_dns_zone_dnssec_key",
  "arguments": {
    "zone_id": "12345678-1234-1234-1234-123456789012"
  }
}
```

**API Reference:** [zonesKeysDelete](https://api.ionos.com/docs/dns/v1/#tag/DNSSEC/operation/zonesKeysDelete)
