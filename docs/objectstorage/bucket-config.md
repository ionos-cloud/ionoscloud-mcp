---
subcategory: "Object Storage"
page_title: "Bucket Configuration"
description: |-
  Tools for inspecting Object Storage bucket configuration in IONOS CLOUD.
---

# Object Storage Bucket Configuration

## get_object_storage_bucket_cors

Gets the CORS configuration for an Object Storage bucket.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `bucket` | string | Yes | The name of the object storage bucket |

**Example:**

```json
{
  "name": "get_object_storage_bucket_cors",
  "arguments": {
    "bucket": "my-bucket"
  }
}
```

**API Reference:** [GetBucketCors](https://api.ionos.com/docs/object-storage-contract-owned-buckets/v2/#tag/CORS/operation/GET_bucket_?cors)

---

## get_object_storage_bucket_encryption

Gets the server-side encryption configuration for an Object Storage bucket.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `bucket` | string | Yes | The name of the object storage bucket |

**Example:**

```json
{
  "name": "get_object_storage_bucket_encryption",
  "arguments": {
    "bucket": "my-bucket"
  }
}
```

**API Reference:** [GetBucketEncryption](https://api.ionos.com/docs/object-storage-contract-owned-buckets/v2/#tag/Encryption/operation/GET_bucket_?encryption)

---

## get_object_storage_bucket_lifecycle

Gets the lifecycle configuration for an Object Storage bucket.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `bucket` | string | Yes | The name of the object storage bucket |

**Example:**

```json
{
  "name": "get_object_storage_bucket_lifecycle",
  "arguments": {
    "bucket": "my-bucket"
  }
}
```

**API Reference:** [GetBucketLifecycle](https://api.ionos.com/docs/object-storage-contract-owned-buckets/v2/#tag/Lifecycle/operation/GET_bucket_?lifecycle)

---

## get_object_storage_bucket_policy

Gets the bucket policy for an Object Storage bucket.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `bucket` | string | Yes | The name of the object storage bucket |

**Example:**

```json
{
  "name": "get_object_storage_bucket_policy",
  "arguments": {
    "bucket": "my-bucket"
  }
}
```

**API Reference:** [GetBucketPolicy](https://api.ionos.com/docs/object-storage-contract-owned-buckets/v2/#tag/Policy/operation/GET_bucket_?policy)

---

## get_object_storage_bucket_policy_status

Gets the policy status for an Object Storage bucket, indicating whether the bucket is public.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `bucket` | string | Yes | The name of the object storage bucket |

**Example:**

```json
{
  "name": "get_object_storage_bucket_policy_status",
  "arguments": {
    "bucket": "my-bucket"
  }
}
```

**API Reference:** [GetBucketPolicyStatus](https://api.ionos.com/docs/object-storage-contract-owned-buckets/v2/#tag/Policy/operation/GET_bucket_?policyStatus)

---

## get_object_storage_bucket_replication

Gets the replication configuration for an Object Storage bucket.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `bucket` | string | Yes | The name of the object storage bucket |

**Example:**

```json
{
  "name": "get_object_storage_bucket_replication",
  "arguments": {
    "bucket": "my-bucket"
  }
}
```

**API Reference:** [GetBucketReplication](https://api.ionos.com/docs/object-storage-contract-owned-buckets/v2/#tag/Replication/operation/GET_bucket_?replication)

---

## get_object_storage_bucket_tagging

Gets the tags for an Object Storage bucket.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `bucket` | string | Yes | The name of the object storage bucket |

**Example:**

```json
{
  "name": "get_object_storage_bucket_tagging",
  "arguments": {
    "bucket": "my-bucket"
  }
}
```

**API Reference:** [GetBucketTagging](https://api.ionos.com/docs/object-storage-contract-owned-buckets/v2/#tag/Tagging/operation/GET_bucket_?tagging)

---

## get_object_storage_bucket_versioning

Gets the versioning configuration for an Object Storage bucket.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `bucket` | string | Yes | The name of the object storage bucket |

**Example:**

```json
{
  "name": "get_object_storage_bucket_versioning",
  "arguments": {
    "bucket": "my-bucket"
  }
}
```

**API Reference:** [GetBucketVersioning](https://api.ionos.com/docs/object-storage-contract-owned-buckets/v2/#tag/Versioning/operation/GET_bucket_?versioning)

---

## get_object_storage_bucket_public_access_block

Gets the public access block configuration for an Object Storage bucket.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `bucket` | string | Yes | The name of the object storage bucket |

**Example:**

```json
{
  "name": "get_object_storage_bucket_public_access_block",
  "arguments": {
    "bucket": "my-bucket"
  }
}
```

**API Reference:** [GetPublicAccessBlock](https://api.ionos.com/docs/object-storage-contract-owned-buckets/v2/#tag/PublicAccessBlock/operation/GET_bucket_?publicAccessBlock)

---

## get_object_storage_bucket_lock_configuration

Gets the Object Lock configuration for an Object Storage bucket.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `bucket` | string | Yes | The name of the object storage bucket |

**Example:**

```json
{
  "name": "get_object_storage_bucket_lock_configuration",
  "arguments": {
    "bucket": "my-bucket"
  }
}
```

**API Reference:** [GetObjectLockConfiguration](https://api.ionos.com/docs/object-storage-contract-owned-buckets/v2/#tag/ObjectLock/operation/GET_bucket_?object-lock)
