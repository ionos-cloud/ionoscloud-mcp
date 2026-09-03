---
subcategory: "Certificate Manager"
page_title: "Certificate"
description: |-
  Tools for listing, inspecting, uploading, renaming and deleting SSL/TLS certificates in IONOS Cloud Certificate Manager.
---

# Certificates

## list_cert_certificates

Lists all SSL/TLS certificates in your IONOS Cloud Certificate Manager account.

**Parameters:** None

**Example:**

```json
{
  "name": "list_cert_certificates",
  "arguments": {}
}
```

**API Reference:** [certificatesGet](https://api.ionos.com/docs/certificatemanager/v2/#tag/Certificates/operation/certificatesGet)

---

## get_cert_certificate

Gets details of a specific SSL/TLS certificate by ID.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `certificate_id` | string | Yes | The ID of the certificate |

**Example:**

```json
{
  "name": "get_cert_certificate",
  "arguments": {
    "certificate_id": "12345678-1234-1234-1234-123456789012"
  }
}
```

**API Reference:** [certificatesFindById](https://api.ionos.com/docs/certificatemanager/v2/#tag/Certificates/operation/certificatesFindById)

---

## create_cert_certificate

Uploads one SSL/TLS certificate with its chain and private key, for use by an Application Load Balancer HTTPS listener. Requires `IONOS_MCP_TOOL_SCOPE` to include `write`.

Two-phase: call first without `confirmation_token` to get a preview and a one-time token, then call again with the token and the same material. The token is bound to a digest of the material, so changing any of the three PEM fields invalidates it and a fresh preview is needed.

`private_key` is write-only. It is never echoed in the preview (the preview shows `(set, not shown)`), never quoted back in a validation error, and never returned by a read tool. For a certificate that should renew itself, use [`create_cert_provider`](provider.md) and [`create_cert_auto_certificate`](auto-certificate.md) instead of uploading here.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | string | Yes | A name for the certificate, for management purposes only |
| `certificate` | string | Yes | The certificate body in PEM format |
| `certificate_chain` | string | Yes | The intermediate CA certificates in PEM format, leaf-issuer first |
| `private_key` | string | Yes | The unencrypted private key in PEM format |
| `confirmation_token` | string | No | Omit on the first call; pass the minted token on the second |

**Example:**

```json
{
  "name": "create_cert_certificate",
  "arguments": {
    "name": "www.example.com 2026",
    "certificate": "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----\n",
    "certificate_chain": "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----\n",
    "private_key": "-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----\n"
  }
}
```

Synchronous (201): the returned body is the stored certificate. Read `metadata.state` to see whether it is `AVAILABLE`.

**API Reference:** [certificatesPost](https://api.ionos.com/docs/certificatemanager/v2/#tag/Certificates/operation/certificatesPost)

---

## update_cert_certificate

Renames a certificate. Requires `IONOS_MCP_TOOL_SCOPE` to include `write`. Single call.

**Only the name can be changed.** The PATCH endpoint accepts the spec's `PatchName` and nothing else, so the certificate body, chain and private key are immutable. To rotate the material, upload a new certificate, repoint the load balancer at it, then delete the old one.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `certificate_id` | string | Yes | The ID of the certificate to rename |
| `name` | string | Yes | The new certificate name |

**Example:**

```json
{
  "name": "update_cert_certificate",
  "arguments": {
    "certificate_id": "12345678-1234-1234-1234-123456789012",
    "name": "www.example.com 2026 (retiring)"
  }
}
```

Synchronous (200): the returned body is the updated certificate.

**API Reference:** [certificatesPatch](https://api.ionos.com/docs/certificatemanager/v2/#tag/Certificates/operation/certificatesPatch)

---

## delete_cert_certificate

Deletes one SSL/TLS certificate and its private key. Requires `IONOS_MCP_TOOL_SCOPE` to include `destructive`.

Two-phase: call first without `confirmation_token` to get a preview and a one-time token, then call again with the token. **This is irreversible** — the private key is gone, so re-uploading needs the original key file.

Any Application Load Balancer HTTPS listener still referencing the certificate stops serving TLS. Certificate Manager cannot list those references, so check the ALB forwarding rules with `list_alb_forwarding_rules` first. If the certificate was issued by an auto-certificate, the preview says so: that auto-certificate survives and issues a replacement at the next renewal.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `certificate_id` | string | Yes | The ID of the certificate to delete |
| `confirmation_token` | string | No | Omit on the first call; pass the minted token on the second |

**Example:**

```json
{
  "name": "delete_cert_certificate",
  "arguments": {
    "certificate_id": "12345678-1234-1234-1234-123456789012"
  }
}
```

Asynchronous (202): the API accepts the request; `get_cert_certificate` answers 404 once it is gone.

**API Reference:** [certificatesDelete](https://api.ionos.com/docs/certificatemanager/v2/#tag/Certificates/operation/certificatesDelete)
