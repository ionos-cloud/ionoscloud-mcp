---
subcategory: "Certificate Manager"
page_title: "Provider"
description: |-
  Tools for listing and inspecting certificate providers in IONOS Cloud Certificate Manager.
---

# Providers

## list_cert_providers

Lists all certificate providers in your IONOS Cloud Certificate Manager account.

**Parameters:** None

**Example:**

```json
{
  "name": "list_cert_providers",
  "arguments": {}
}
```

**API Reference:** [providersGet](https://api.ionos.com/docs/certificatemanager/v2/#tag/Providers/operation/providersGet)

---

## get_cert_provider

Gets details of a specific certificate provider by ID.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `provider_id` | string | Yes | The ID of the certificate provider |

**Example:**

```json
{
  "name": "get_cert_provider",
  "arguments": {
    "provider_id": "12345678-1234-1234-1234-123456789012"
  }
}
```

**API Reference:** [providersFindById](https://api.ionos.com/docs/certificatemanager/v2/#tag/Providers/operation/providersFindById)
