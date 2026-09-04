---
subcategory: "Certificate Manager"
page_title: "Provider"
description: |-
  Tools for listing, inspecting, registering, renaming and deleting certificate providers in IONOS Cloud Certificate Manager.
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

---

## create_cert_provider

Registers one ACME certificate provider (a certificate authority) that auto-certificates can issue and renew certificates through. Requires `IONOS_MCP_TOOL_SCOPE` to include `write`.

Two-phase: call first without `confirmation_token` to get a preview and a one-time token, then call again with the token.

`server` must be an `https` ACME directory URL. Supply `external_account_binding` only for a CA that requires a pre-registered account (ZeroSSL, Google Trust Services) — both halves are required together, so a key ID can never be sent without its secret. The secret is write-only: it is never echoed in a preview and never returned by a read tool.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | string | Yes | A name for the provider, for management purposes only |
| `email` | string | Yes | The email registered with the CA as the certificate requester |
| `server` | string | Yes | The ACME directory URL, e.g. `https://acme-v02.api.letsencrypt.org/directory` |
| `external_account_binding` | object | No | `{ "key_id": "...", "key_secret": "..." }` — both required when present |
| `confirmation_token` | string | No | Omit on the first call; pass the minted token on the second |

**Example:**

```json
{
  "name": "create_cert_provider",
  "arguments": {
    "name": "Let's Encrypt",
    "email": "ops@example.com",
    "server": "https://acme-v02.api.letsencrypt.org/directory"
  }
}
```

Synchronous (201): the returned body is the stored provider. Read `metadata.state` to see whether the ACME account registered successfully.

**API Reference:** [providersPost](https://api.ionos.com/docs/certificatemanager/v2/#tag/Providers/operation/providersPost)

---

## update_cert_provider

Renames a certificate provider. Requires `IONOS_MCP_TOOL_SCOPE` to include `write`. Single call.

**Only the name can be changed.** The PATCH endpoint accepts the spec's `PatchName` and nothing else, so the email, the ACME directory URL and the external account binding are immutable. To change any of them, create a new provider, repoint the auto-certificates at it, and delete this one.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `provider_id` | string | Yes | The ID of the provider to rename |
| `name` | string | Yes | The new provider name |

**Example:**

```json
{
  "name": "update_cert_provider",
  "arguments": {
    "provider_id": "74edc770-5cc6-5976-ac99-013ddb4af403",
    "name": "Let's Encrypt (production)"
  }
}
```

Synchronous (200): the returned body is the updated provider, with the binding secret redacted.

**API Reference:** [providersPatch](https://api.ionos.com/docs/certificatemanager/v2/#tag/Providers/operation/providersPatch)

---

## delete_cert_provider

Deletes one ACME certificate provider. Requires `IONOS_MCP_TOOL_SCOPE` to include `destructive`.

Two-phase: call first without `confirmation_token` to get a blast-radius preview and a one-time token, then call again with the token. **This is irreversible**, and the external account binding secret cannot be recovered.

Auto-certificates issue through a provider, so any that name this one lose the ability to renew. The preview counts them — `autoCertificatesGet` has no provider filter, so the match is made client-side over the listed auto-certificates. If the list cannot be read, the preview says the radius is incomplete rather than reporting zero.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `provider_id` | string | Yes | The ID of the provider to delete |
| `confirmation_token` | string | No | Omit on the first call; pass the minted token on the second |

**Example:**

```json
{
  "name": "delete_cert_provider",
  "arguments": {
    "provider_id": "74edc770-5cc6-5976-ac99-013ddb4af403"
  }
}
```

Asynchronous (202): the API accepts the request; `get_cert_provider` answers 404 once it is gone.

**API Reference:** [providersDelete](https://api.ionos.com/docs/certificatemanager/v2/#tag/Providers/operation/providersDelete)
