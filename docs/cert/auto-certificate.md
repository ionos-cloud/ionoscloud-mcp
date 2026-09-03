---
subcategory: "Certificate Manager"
page_title: "AutoCertificate"
description: |-
  Tools for listing, inspecting, creating, renaming and deleting auto-certificates in IONOS Cloud Certificate Manager.
---

# Auto-Certificates

## list_cert_auto_certificates

Lists all auto-certificates in your IONOS Cloud Certificate Manager account.

**Parameters:** None

**Example:**

```json
{
  "name": "list_cert_auto_certificates",
  "arguments": {}
}
```

**API Reference:** [autoCertificatesGet](https://api.ionos.com/docs/certificatemanager/v2/#tag/AutoCertificates/operation/autoCertificatesGet)

---

## get_cert_auto_certificate

Gets details of a specific auto-certificate by ID.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `auto_certificate_id` | string | Yes | The ID of the auto-certificate |

**Example:**

```json
{
  "name": "get_cert_auto_certificate",
  "arguments": {
    "auto_certificate_id": "12345678-1234-1234-1234-123456789012"
  }
}
```

**API Reference:** [autoCertificatesFindById](https://api.ionos.com/docs/certificatemanager/v2/#tag/AutoCertificates/operation/autoCertificatesFindById)

---

## create_cert_auto_certificate

Creates one auto-certificate: a standing instruction to issue a certificate for a DNS name through a certificate provider and renew it before expiry. Requires `IONOS_MCP_TOOL_SCOPE` to include `write`.

Two-phase: call first without `confirmation_token` to get a preview and a one-time token, then call again with the token. The preview resolves `provider_id` and shows which CA will issue, so an unknown provider is rejected with a named field error rather than a 422 on the execute call.

`common_name` and every `subject_alternative_names` entry must belong to a zone hosted in IONOS Cloud DNS: the provider validates ownership through a DNS challenge, so a name outside IONOS DNS fails to issue.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `provider_id` | string | Yes | The ID of the provider that will issue the certificates ([`list_cert_providers`](provider.md)) |
| `common_name` | string | Yes | The DNS name to issue for, e.g. `www.example.com` |
| `key_algorithm` | string | Yes | `rsa2048`, `rsa3072` or `rsa4096` (case-insensitive) |
| `name` | string | Yes | A name for the auto-certificate, for management purposes only |
| `subject_alternative_names` | string[] | No | Additional DNS names to add to the issued certificate |
| `confirmation_token` | string | No | Omit on the first call; pass the minted token on the second |

**Example:**

```json
{
  "name": "create_cert_auto_certificate",
  "arguments": {
    "provider_id": "74edc770-5cc6-5976-ac99-013ddb4af403",
    "common_name": "www.example.com",
    "key_algorithm": "rsa4096",
    "name": "www.example.com auto-renew",
    "subject_alternative_names": ["app.example.com"]
  }
}
```

Synchronous (201) for the auto-certificate itself; issuing the certificate happens afterwards. Poll `get_cert_auto_certificate` until `metadata.lastIssuedCertificate` names a certificate, then read it with `get_cert_certificate`.

**API Reference:** [autoCertificatesPost](https://api.ionos.com/docs/certificatemanager/v2/#tag/AutoCertificates/operation/autoCertificatesPost)

---

## update_cert_auto_certificate

Renames an auto-certificate. Requires `IONOS_MCP_TOOL_SCOPE` to include `write`. Single call.

**Only the name can be changed.** The PATCH endpoint accepts the spec's `PatchName` and nothing else, so the provider, common name, subject alternative names and key algorithm are immutable. To change what is issued, delete the auto-certificate and create a new one.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `auto_certificate_id` | string | Yes | The ID of the auto-certificate to rename |
| `name` | string | Yes | The new auto-certificate name |

**Example:**

```json
{
  "name": "update_cert_auto_certificate",
  "arguments": {
    "auto_certificate_id": "f88467f8-a2d6-5871-83b9-e10f23d0a48a",
    "name": "www.example.com auto-renew (prod)"
  }
}
```

Synchronous (200): the returned body is the updated auto-certificate.

**API Reference:** [autoCertificatesPatch](https://api.ionos.com/docs/certificatemanager/v2/#tag/AutoCertificates/operation/autoCertificatesPatch)

---

## delete_cert_auto_certificate

Deletes one auto-certificate, stopping automatic renewal for its DNS name. Requires `IONOS_MCP_TOOL_SCOPE` to include `destructive`.

Two-phase: call first without `confirmation_token` to get a blast-radius preview and a one-time token, then call again with the token. **This is irreversible.**

The preview counts the certificates this auto-certificate has issued (via `filter.autoCertificate`). Those certificates keep working until they expire, and nothing renews them afterwards — so a load balancer using one starts serving an expired certificate on that date. If the count cannot be read, the preview says the radius is incomplete rather than reporting zero.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `auto_certificate_id` | string | Yes | The ID of the auto-certificate to delete |
| `confirmation_token` | string | No | Omit on the first call; pass the minted token on the second |

**Example:**

```json
{
  "name": "delete_cert_auto_certificate",
  "arguments": {
    "auto_certificate_id": "f88467f8-a2d6-5871-83b9-e10f23d0a48a"
  }
}
```

Asynchronous (202): the API accepts the request; `get_cert_auto_certificate` answers 404 once it is gone.

**API Reference:** [autoCertificatesDelete](https://api.ionos.com/docs/certificatemanager/v2/#tag/AutoCertificates/operation/autoCertificatesDelete)
