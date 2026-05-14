---
subcategory: "Object Storage"
page_title: "Access Keys"
description: |-
  Tools for listing and inspecting Object Storage access keys in IONOS CLOUD.
---

# Object Storage Access Keys

## list_object_storage_access_keys

Lists all Object Storage access keys for the contract. Returns key IDs and metadata but not the secret keys.

**Parameters:** None

**Example:**

```json
{
  "name": "list_object_storage_access_keys",
  "arguments": {}
}
```

**API Reference:** [AccesskeysGet](https://api.ionos.com/docs/object-storage-management/v1/#tag/AccessKeys/operation/accesskeysGet)

---

## get_object_storage_access_key

Gets details of a specific Object Storage access key by its ID. Returns key metadata but not the secret key.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `access_key_id` | string | Yes | The ID of the object storage access key |

**Example:**

```json
{
  "name": "get_object_storage_access_key",
  "arguments": {
    "access_key_id": "ak-1"
  }
}
```

**API Reference:** [AccesskeysFindById](https://api.ionos.com/docs/object-storage-management/v1/#tag/AccessKeys/operation/accesskeysFindById)
