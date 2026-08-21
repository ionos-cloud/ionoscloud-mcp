---
subcategory: "DNS"
page_title: "Reverse Record"
description: |-
  Tools for listing, inspecting, creating, updating and deleting reverse DNS records in IONOS CLOUD.
---

# Reverse DNS Records

## list_dns_reverse_records

Lists all reverse DNS records in your IONOS CLOUD account.

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

---

## create_dns_reverse_record

Creates one reverse DNS (PTR) record, making an IP resolve back to a hostname. Requires `IONOS_MCP_TOOL_SCOPE` to include `write`.

Two-phase: call first without `confirmation_token` to get a preview and a one-time token, then call again with the token and the same `ip`.

The IP must be one your contract owns: an IPv4 from one of your IP blocks (see `list_ip_blocks`) or an IPv6 from a VDC. The API answers `409` for an IP that is not eligible, and each IP can carry only one reverse record.

Unlike the other DNS creates, this one is **synchronous** — the returned record is the final state, and a reverse record carries no provisioning state to poll.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | string | Yes | The hostname the IP should resolve back to, e.g. `mail.example.com` |
| `ip` | string | Yes | The IPv4 or IPv6 address to create the reverse record for |
| `description` | string | No | Free-text description of what the record is for |
| `confirmation_token` | string | No | Omit on the first call; pass the minted token on the second |

**Example:**

```json
{
  "name": "create_dns_reverse_record",
  "arguments": {
    "name": "mail.example.com",
    "ip": "192.0.2.10"
  }
}
```

**API Reference:** [reverserecordsPost](https://api.ionos.com/docs/dns/v1/#tag/ReverseRecords/operation/reverserecordsPost)

---

## update_dns_reverse_record

Updates the hostname or description of a reverse DNS record. Requires `IONOS_MCP_TOOL_SCOPE` to include `write`. Single call, and synchronous.

The IP is immutable here — it identifies which address the record covers, so pointing a different IP at a hostname is a create, not an update. The endpoint is a PUT, so the current IP is read and sent back unchanged.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `reverse_record_id` | string | Yes | The ID of the reverse record to update |
| `name` | string | No | A new hostname for the IP to resolve back to. Omit to keep the current one. |
| `description` | string | No | A new description. Omit to keep the current one. |

**Example:**

```json
{
  "name": "update_dns_reverse_record",
  "arguments": {
    "reverse_record_id": "e74d0d15-f567-4b7b-9069-26ee1f93bae3",
    "name": "smtp.example.com"
  }
}
```

**API Reference:** [reverserecordsPut](https://api.ionos.com/docs/dns/v1/#tag/ReverseRecords/operation/reverserecordsPut)

---

## delete_dns_reverse_record

Deletes a reverse DNS record, so its IP stops resolving back to a hostname. Requires `IONOS_MCP_TOOL_SCOPE` to include `destructive`.

Two-phase: call first without `confirmation_token` to get a preview and a one-time token, then call again with the token. Irreversible. Mail servers commonly reject senders with no reverse DNS, so check what depends on the record first.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `reverse_record_id` | string | Yes | The ID of the reverse record to delete |
| `confirmation_token` | string | No | Omit on the first call; pass the minted token on the second |

**Example:**

```json
{
  "name": "delete_dns_reverse_record",
  "arguments": {
    "reverse_record_id": "e74d0d15-f567-4b7b-9069-26ee1f93bae3"
  }
}
```

**API Reference:** [reverserecordsDelete](https://api.ionos.com/docs/dns/v1/#tag/ReverseRecords/operation/reverserecordsDelete)
