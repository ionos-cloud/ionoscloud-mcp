---
subcategory: "Certificate Manager"
page_title: "Certificate"
description: |-
  Tools for listing, inspecting, renaming and deleting SSL/TLS certificates in IONOS Cloud Certificate Manager.
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

## No create tool — by design

**Certificate material cannot be uploaded through this server.** `certificatesPost` requires the private key in the request body, so a tool would have to accept it as an argument. That puts the key in the model's context, in the session transcript on disk in plaintext, and in every subsequent request to your model provider for the rest of that session. Redaction does not reach any of it — it guards the response and the two-phase preview, which are the *outbound* paths.

Use one of these instead:

| Goal | Route |
|---|---|
| A certificate that issues and renews itself | [`create_cert_auto_certificate`](auto-certificate.md) — IONOS generates and holds the key, so nothing sensitive crosses |
| Upload material you already hold | `ionosctl`, the REST API directly, or the [DCD](https://dcd.ionos.com/) |

`update_cert_certificate` and `delete_cert_certificate` remain available and work on certificates from either route. `TestCreateCertCertificateIsNotRegistered` asserts the tool stays absent at every scope.

---

## update_cert_certificate

Renames a certificate. Requires `IONOS_MCP_TOOL_SCOPE` to include `write`. Single call.

**Only the name can be changed.** The PATCH endpoint accepts the spec's `PatchName` and nothing else, so the certificate body, chain and private key are immutable. To rotate, issue a replacement with [`create_cert_auto_certificate`](auto-certificate.md) (or upload outside this server), repoint the load balancer at it, then delete the old one.

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

Two-phase: call first without `confirmation_token` to get a preview and a one-time token, then call again with the token. **This is irreversible** — the private key is gone, and this server cannot upload a replacement, so re-creating it means issuing a new certificate or uploading outside this server.

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
