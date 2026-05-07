---
subcategory: "Certificate Manager"
page_title: "Certificate"
description: |-
  Tools for listing and inspecting SSL/TLS certificates in IONOS Cloud Certificate Manager.
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
