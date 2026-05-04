---
subcategory: "Certificate Manager"
page_title: "AutoCertificate"
description: |-
  Tools for listing and inspecting auto-certificates in IONOS Cloud Certificate Manager.
---

# Auto-Certificates

## list_auto_certificates

Lists all auto-certificates in your IONOS Cloud Certificate Manager account.

**Parameters:** None

**Example:**

```json
{
  "name": "list_auto_certificates",
  "arguments": {}
}
```

**API Reference:** [autoCertificatesGet](https://api.ionos.com/docs/certificatemanager/v2/#tag/AutoCertificates/operation/autoCertificatesGet)

---

## get_auto_certificate

Gets details of a specific auto-certificate by ID.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `auto_certificate_id` | string | Yes | The ID of the auto-certificate |

**Example:**

```json
{
  "name": "get_auto_certificate",
  "arguments": {
    "auto_certificate_id": "12345678-1234-1234-1234-123456789012"
  }
}
```

**API Reference:** [autoCertificatesFindById](https://api.ionos.com/docs/certificatemanager/v2/#tag/AutoCertificates/operation/autoCertificatesFindById)
