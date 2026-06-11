package objectstorage

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterBucketConfigTools(server *mcp.Server, cache *clientCache) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_object_storage_bucket_cors",
		Annotations: tools.ReadOnly,
		Description: "Get the CORS rules configured on an Object Storage bucket. Returns an error if the bucket has no CORS configuration.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ObjectStorageBucketInput) (*mcp.CallToolResult, any, error) {
		c, err := cache.forBucket(ctx, input.Bucket)
		if err != nil {
			return tools.ToResult(nil, err)
		}
		result, _, err := c.CORSApi.GetBucketCors(ctx, input.Bucket).Execute()
		return tools.ToResult(result, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_object_storage_bucket_encryption",
		Annotations: tools.ReadOnly,
		Description: "Get the server-side encryption configuration of an Object Storage bucket — the algorithm applied to newly written objects.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ObjectStorageBucketInput) (*mcp.CallToolResult, any, error) {
		c, err := cache.forBucket(ctx, input.Bucket)
		if err != nil {
			return tools.ToResult(nil, err)
		}
		result, _, err := c.EncryptionApi.GetBucketEncryption(ctx, input.Bucket).Execute()
		return tools.ToResult(result, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_object_storage_bucket_lifecycle",
		Annotations: tools.ReadOnly,
		Description: "Get the lifecycle rules of an Object Storage bucket, e.g. object expiration and incomplete-upload cleanup. Returns an error if the bucket has no lifecycle configuration.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ObjectStorageBucketInput) (*mcp.CallToolResult, any, error) {
		c, err := cache.forBucket(ctx, input.Bucket)
		if err != nil {
			return tools.ToResult(nil, err)
		}
		result, _, err := c.LifecycleApi.GetBucketLifecycle(ctx, input.Bucket).Execute()
		return tools.ToResult(result, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_object_storage_bucket_policy",
		Annotations: tools.ReadOnly,
		Description: "Get the bucket policy (JSON access-policy document) of an Object Storage bucket. Returns an error if the bucket has no policy. For a quick public/private verdict use get_object_storage_bucket_policy_status instead.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ObjectStorageBucketInput) (*mcp.CallToolResult, any, error) {
		c, err := cache.forBucket(ctx, input.Bucket)
		if err != nil {
			return tools.ToResult(nil, err)
		}
		result, _, err := c.PolicyApi.GetBucketPolicy(ctx, input.Bucket).Execute()
		return tools.ToResult(result, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_object_storage_bucket_policy_status",
		Annotations: tools.ReadOnly,
		Description: "Report whether an Object Storage bucket is public, derived from its policy. Lighter than get_object_storage_bucket_policy — use it when you only need the public/private answer.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ObjectStorageBucketInput) (*mcp.CallToolResult, any, error) {
		c, err := cache.forBucket(ctx, input.Bucket)
		if err != nil {
			return tools.ToResult(nil, err)
		}
		result, _, err := c.PolicyApi.GetBucketPolicyStatus(ctx, input.Bucket).Execute()
		return tools.ToResult(result, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_object_storage_bucket_replication",
		Annotations: tools.ReadOnly,
		Description: "Get the cross-region replication rules of an Object Storage bucket. Returns an error if replication is not configured.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ObjectStorageBucketInput) (*mcp.CallToolResult, any, error) {
		c, err := cache.forBucket(ctx, input.Bucket)
		if err != nil {
			return tools.ToResult(nil, err)
		}
		result, _, err := c.ReplicationApi.GetBucketReplication(ctx, input.Bucket).Replication(true).Execute()
		return tools.ToResult(result, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_object_storage_bucket_tagging",
		Annotations: tools.ReadOnly,
		Description: "Get the tag set of an Object Storage bucket. For tags on an individual object use get_object_storage_object_tagging instead.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ObjectStorageBucketInput) (*mcp.CallToolResult, any, error) {
		c, err := cache.forBucket(ctx, input.Bucket)
		if err != nil {
			return tools.ToResult(nil, err)
		}
		result, _, err := c.TaggingApi.GetBucketTagging(ctx, input.Bucket).Execute()
		return tools.ToResult(result, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_object_storage_bucket_versioning",
		Annotations: tools.ReadOnly,
		Description: "Get the versioning state of an Object Storage bucket: Enabled, Suspended, or unset. Versioning must be enabled for list_object_storage_object_versions to return data.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ObjectStorageBucketInput) (*mcp.CallToolResult, any, error) {
		c, err := cache.forBucket(ctx, input.Bucket)
		if err != nil {
			return tools.ToResult(nil, err)
		}
		result, _, err := c.VersioningApi.GetBucketVersioning(ctx, input.Bucket).Execute()
		return tools.ToResult(result, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_object_storage_bucket_public_access_block",
		Annotations: tools.ReadOnly,
		Description: "Get the public-access-block settings of an Object Storage bucket — the flags that override ACLs and policies to keep a bucket private.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ObjectStorageBucketInput) (*mcp.CallToolResult, any, error) {
		c, err := cache.forBucket(ctx, input.Bucket)
		if err != nil {
			return tools.ToResult(nil, err)
		}
		result, _, err := c.PublicAccessBlockApi.GetPublicAccessBlock(ctx, input.Bucket).Execute()
		return tools.ToResult(result, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_object_storage_bucket_lock_configuration",
		Annotations: tools.ReadOnly,
		Description: "Get the Object Lock (WORM retention) defaults of an Object Storage bucket. Returns an error if Object Lock is not enabled on the bucket; for per-object settings use get_object_storage_object_retention.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ObjectStorageBucketInput) (*mcp.CallToolResult, any, error) {
		c, err := cache.forBucket(ctx, input.Bucket)
		if err != nil {
			return tools.ToResult(nil, err)
		}
		result, _, err := c.ObjectLockApi.GetObjectLockConfiguration(ctx, input.Bucket).Execute()
		return tools.ToResult(result, err)
	})
}
