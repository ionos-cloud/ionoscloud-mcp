---
subcategory: "Object Storage"
page_title: "Buckets"
description: |-
  Tools for listing and inspecting Object Storage buckets in IONOS Cloud.
---

# Object Storage Buckets

## list_object_storage_buckets

Lists all Object Storage buckets owned by the contract.

**Parameters:** None

**Example:**

```json
{
  "name": "list_object_storage_buckets",
  "arguments": {}
}
```

**API Reference:** [ListAllBuckets](https://api.ionos.com/docs/object-storage-contract-owned-buckets/v2/#tag/Buckets/operation/GET_-_)

---

## get_object_storage_bucket_location

Gets the region/location of an Object Storage bucket.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `bucket` | string | Yes | The name of the object storage bucket |

**Example:**

```json
{
  "name": "get_object_storage_bucket_location",
  "arguments": {
    "bucket": "my-bucket"
  }
}
```

**API Reference:** [GetBucketLocation](https://api.ionos.com/docs/object-storage-contract-owned-buckets/v2/#tag/Buckets/operation/GET_bucket_?location)

---

## head_object_storage_bucket

Checks whether an Object Storage bucket exists and is accessible with the current credentials.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `bucket` | string | Yes | The name of the object storage bucket |

**Example:**

```json
{
  "name": "head_object_storage_bucket",
  "arguments": {
    "bucket": "my-bucket"
  }
}
```

**API Reference:** [HeadBucket](https://api.ionos.com/docs/object-storage-contract-owned-buckets/v2/#tag/Buckets/operation/HEAD_bucket_)
