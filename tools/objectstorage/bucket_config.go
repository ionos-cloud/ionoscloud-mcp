package objectstorage

import (
	"context"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
	sdk "github.com/ionos-cloud/sdk-go-bundle/products/objectstorage/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterBucketConfigTools(server *mcp.Server, client *sdk.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_object_storage_bucket_cors",
		Description: "Get the CORS configuration for an Object Storage bucket.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ObjectStorageBucketInput) (*mcp.CallToolResult, any, error) {
		result, _, err := client.CORSApi.GetBucketCors(ctx, input.Bucket).Execute()
		return tools.ToResult(result, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_object_storage_bucket_encryption",
		Description: "Get the server-side encryption configuration for an Object Storage bucket.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ObjectStorageBucketInput) (*mcp.CallToolResult, any, error) {
		result, _, err := client.EncryptionApi.GetBucketEncryption(ctx, input.Bucket).Execute()
		return tools.ToResult(result, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_object_storage_bucket_lifecycle",
		Description: "Get the lifecycle configuration for an Object Storage bucket.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ObjectStorageBucketInput) (*mcp.CallToolResult, any, error) {
		result, _, err := client.LifecycleApi.GetBucketLifecycle(ctx, input.Bucket).Execute()
		return tools.ToResult(result, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_object_storage_bucket_policy",
		Description: "Get the bucket policy for an Object Storage bucket.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ObjectStorageBucketInput) (*mcp.CallToolResult, any, error) {
		result, _, err := client.PolicyApi.GetBucketPolicy(ctx, input.Bucket).Execute()
		return tools.ToResult(result, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_object_storage_bucket_policy_status",
		Description: "Get the policy status for an Object Storage bucket, indicating whether the bucket is public.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ObjectStorageBucketInput) (*mcp.CallToolResult, any, error) {
		result, _, err := client.PolicyApi.GetBucketPolicyStatus(ctx, input.Bucket).Execute()
		return tools.ToResult(result, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_object_storage_bucket_replication",
		Description: "Get the replication configuration for an Object Storage bucket.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ObjectStorageBucketInput) (*mcp.CallToolResult, any, error) {
		result, _, err := client.ReplicationApi.GetBucketReplication(ctx, input.Bucket).Replication(true).Execute()
		return tools.ToResult(result, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_object_storage_bucket_tagging",
		Description: "Get the tags for an Object Storage bucket.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ObjectStorageBucketInput) (*mcp.CallToolResult, any, error) {
		result, _, err := client.TaggingApi.GetBucketTagging(ctx, input.Bucket).Execute()
		return tools.ToResult(result, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_object_storage_bucket_versioning",
		Description: "Get the versioning configuration for an Object Storage bucket.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ObjectStorageBucketInput) (*mcp.CallToolResult, any, error) {
		result, _, err := client.VersioningApi.GetBucketVersioning(ctx, input.Bucket).Execute()
		return tools.ToResult(result, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_object_storage_bucket_public_access_block",
		Description: "Get the public access block configuration for an Object Storage bucket.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ObjectStorageBucketInput) (*mcp.CallToolResult, any, error) {
		result, _, err := client.PublicAccessBlockApi.GetPublicAccessBlock(ctx, input.Bucket).Execute()
		return tools.ToResult(result, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_object_storage_bucket_lock_configuration",
		Description: "Get the Object Lock configuration for an Object Storage bucket.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ObjectStorageBucketInput) (*mcp.CallToolResult, any, error) {
		result, _, err := client.ObjectLockApi.GetObjectLockConfiguration(ctx, input.Bucket).Execute()
		return tools.ToResult(result, err)
	})
}
