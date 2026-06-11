package objectstorage

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterBucketTools(server *mcp.Server, cache *clientCache) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_object_storage_buckets",
		Annotations: tools.ReadOnly,
		Description: "List all Object Storage buckets owned by the contract, with names and creation dates. Use get_object_storage_bucket_location to find the region a bucket lives in.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		result, _, err := cache.base.BucketsApi.ListBuckets(ctx).Execute()
		return tools.ToResult(result, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_object_storage_bucket_location",
		Annotations: tools.ReadOnly,
		Description: "Get the region (e.g. eu-central-3) where an Object Storage bucket is stored. Use it when the user asks where their data resides; other bucket tools resolve the regional endpoint automatically.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ObjectStorageBucketInput) (*mcp.CallToolResult, any, error) {
		result, _, err := cache.base.BucketsApi.GetBucketLocation(ctx, input.Bucket).Execute()
		return tools.ToResult(result, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "head_object_storage_bucket",
		Annotations: tools.ReadOnly,
		Description: "Check whether an Object Storage bucket exists and is accessible with the configured credentials. Returns an error when missing or forbidden. Cheaper than list_object_storage_buckets for a single-bucket existence check.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ObjectStorageBucketInput) (*mcp.CallToolResult, any, error) {
		c, err := cache.forBucket(ctx, input.Bucket)
		if err != nil {
			return tools.ToResult(nil, err)
		}
		_, err = c.BucketsApi.HeadBucket(ctx, input.Bucket).Execute()
		if err != nil {
			return tools.ToResult(nil, err)
		}
		return tools.ToResult(map[string]any{"bucket": input.Bucket, "accessible": true}, nil)
	})
}
