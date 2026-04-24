---
subcategory: "Object Storage"
page_title: "Objects"
description: |-
  Tools for listing and inspecting objects in IONOS Object Storage buckets.
---

# Object Storage Objects

## list_object_storage_objects

Lists objects in an Object Storage bucket. Supports an optional prefix to filter by key path (e.g. `images/` to list only objects under that prefix), an optional `continuation_token` to continue from a previous page, and an optional `max_keys` to control page size. Returns up to 1000 objects per call by default; use the `next_continuation_token` from the response as `continuation_token` in a subsequent call to page through larger result sets.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `bucket` | string | Yes | The name of the object storage bucket |
| `prefix` | string | No | Key prefix to filter results (e.g. `images/`) |
| `continuation_token` | string | No | Pagination token from a previous response's `next_continuation_token` |
| `max_keys` | integer | No | Maximum number of objects to return per page (default 1000) |

**Example — first page:**

```json
{
  "name": "list_object_storage_objects",
  "arguments": {
    "bucket": "my-bucket",
    "prefix": "images/"
  }
}
```

**Example — next page:**

```json
{
  "name": "list_object_storage_objects",
  "arguments": {
    "bucket": "my-bucket",
    "prefix": "images/",
    "continuation_token": "<token from previous response>"
  }
}
```

**API Reference:** [ListObjectsV2](https://api.ionos.com/docs/object-storage-contract-owned-buckets/v2/#tag/Objects/operation/GET_bucket_?list-type=2)

---

## head_object_storage_object

Check whether an object exists in an Object Storage bucket and retrieve its user-defined metadata (x-amz-meta-* headers). Returns an error if the object does not exist or is not accessible.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `bucket` | string | Yes | The name of the object storage bucket |
| `key` | string | Yes | The object key (path within the bucket) |

**Example:**

```json
{
  "name": "head_object_storage_object",
  "arguments": {
    "bucket": "my-bucket",
    "key": "images/photo.jpg"
  }
}
```

**API Reference:** [HeadObject](https://api.ionos.com/docs/object-storage-contract-owned-buckets/v2/#tag/Objects/operation/HEAD_bucket_key_)

---

## list_object_storage_object_versions

Lists all versions of objects in an Object Storage bucket. Requires versioning to be enabled on the bucket. Supports an optional prefix filter.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `bucket` | string | Yes | The name of the object storage bucket |
| `prefix` | string | No | Key prefix to filter versions |

**Example:**

```json
{
  "name": "list_object_storage_object_versions",
  "arguments": {
    "bucket": "my-bucket",
    "prefix": "logs/"
  }
}
```

**API Reference:** [ListObjectVersions](https://api.ionos.com/docs/object-storage-contract-owned-buckets/v2/#tag/Versioning/operation/GET_bucket_?versions)

---

## get_object_storage_object_tagging

Gets the tags for an object in an Object Storage bucket.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `bucket` | string | Yes | The name of the object storage bucket |
| `key` | string | Yes | The object key (path within the bucket) |

**Example:**

```json
{
  "name": "get_object_storage_object_tagging",
  "arguments": {
    "bucket": "my-bucket",
    "key": "images/photo.jpg"
  }
}
```

**API Reference:** [GetObjectTagging](https://api.ionos.com/docs/object-storage-contract-owned-buckets/v2/#tag/Tagging/operation/GET_bucket_key_?tagging)

---

## get_object_storage_object_retention

Gets the Object Lock retention configuration for an object in an Object Storage bucket.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `bucket` | string | Yes | The name of the object storage bucket |
| `key` | string | Yes | The object key (path within the bucket) |

**Example:**

```json
{
  "name": "get_object_storage_object_retention",
  "arguments": {
    "bucket": "my-bucket",
    "key": "images/photo.jpg"
  }
}
```

**API Reference:** [GetObjectRetention](https://api.ionos.com/docs/object-storage-contract-owned-buckets/v2/#tag/ObjectLock/operation/GET_bucket_key_?retention)

---

## get_object_storage_object_legal_hold

Gets the Object Lock legal hold status for an object in an Object Storage bucket.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `bucket` | string | Yes | The name of the object storage bucket |
| `key` | string | Yes | The object key (path within the bucket) |

**Example:**

```json
{
  "name": "get_object_storage_object_legal_hold",
  "arguments": {
    "bucket": "my-bucket",
    "key": "images/photo.jpg"
  }
}
```

**API Reference:** [GetObjectLegalHold](https://api.ionos.com/docs/object-storage-contract-owned-buckets/v2/#tag/ObjectLock/operation/GET_bucket_key_?legal-hold)
