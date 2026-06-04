package objectstorage

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterObjectTools(server *mcp.Server, cache *clientCache) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_object_storage_objects",
		Description: "List objects in an Object Storage bucket. Supports an optional prefix to filter by key path (e.g. 'images/' to list only objects under that prefix), an optional continuation_token to continue from a previous page, and an optional max_keys to control page size. Returns up to 1000 objects per call by default; use the next_continuation_token from the response as continuation_token in a subsequent call to page through larger result sets.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ObjectStorageListObjectsInput) (*mcp.CallToolResult, any, error) {
		c, err := cache.forBucket(ctx, input.Bucket)
		if err != nil {
			return tools.ToResult(nil, err)
		}
		listReq := c.ObjectsApi.ListObjectsV2(ctx, input.Bucket)
		if input.Prefix != nil {
			listReq = listReq.Prefix(*input.Prefix)
		}
		if input.ContinuationToken != nil {
			listReq = listReq.ContinuationToken(*input.ContinuationToken)
		}
		if input.MaxKeys != nil {
			listReq = listReq.MaxKeys(*input.MaxKeys)
		}
		result, _, err := listReq.Execute()
		return tools.ToResult(result, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "head_object_storage_object",
		Description: "Check whether an object exists in an Object Storage bucket and retrieve its user-defined metadata (x-amz-meta-* headers). Returns an error if the object does not exist or is not accessible.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ObjectStorageObjectInput) (*mcp.CallToolResult, any, error) {
		c, err := cache.forBucket(ctx, input.Bucket)
		if err != nil {
			return tools.ToResult(nil, err)
		}
		output, _, err := c.ObjectsApi.HeadObject(ctx, input.Bucket, input.Key).Execute()
		if err != nil {
			return tools.ToResult(nil, err)
		}
		result := map[string]any{"bucket": input.Bucket, "key": input.Key, "accessible": true}
		if output != nil && output.Metadata != nil {
			result["metadata"] = *output.Metadata
		}
		return tools.ToResult(result, nil)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_object_storage_object_versions",
		Description: "List all versions of objects in an Object Storage bucket. Requires versioning to be enabled on the bucket. Supports an optional prefix filter.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ObjectStorageListObjectVersionsInput) (*mcp.CallToolResult, any, error) {
		c, err := cache.forBucket(ctx, input.Bucket)
		if err != nil {
			return tools.ToResult(nil, err)
		}
		listReq := c.VersionsApi.ListObjectVersions(ctx, input.Bucket)
		if input.Prefix != nil {
			listReq = listReq.Prefix(*input.Prefix)
		}
		result, _, err := listReq.Execute()
		return tools.ToResult(result, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_object_storage_object_tagging",
		Description: "Get the tags for an object in an Object Storage bucket.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ObjectStorageObjectInput) (*mcp.CallToolResult, any, error) {
		c, err := cache.forBucket(ctx, input.Bucket)
		if err != nil {
			return tools.ToResult(nil, err)
		}
		result, _, err := c.TaggingApi.GetObjectTagging(ctx, input.Bucket, input.Key).Execute()
		return tools.ToResult(result, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_object_storage_object_retention",
		Description: "Get the Object Lock retention configuration for an object in an Object Storage bucket.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ObjectStorageObjectInput) (*mcp.CallToolResult, any, error) {
		c, err := cache.forBucket(ctx, input.Bucket)
		if err != nil {
			return tools.ToResult(nil, err)
		}
		result, _, err := c.ObjectLockApi.GetObjectRetention(ctx, input.Bucket, input.Key).Execute()
		return tools.ToResult(result, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_object_storage_object_legal_hold",
		Description: "Get the Object Lock legal hold status for an object in an Object Storage bucket.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ObjectStorageObjectInput) (*mcp.CallToolResult, any, error) {
		c, err := cache.forBucket(ctx, input.Bucket)
		if err != nil {
			return tools.ToResult(nil, err)
		}
		result, _, err := c.ObjectLockApi.GetObjectLegalHold(ctx, input.Bucket, input.Key).Execute()
		return tools.ToResult(result, err)
	})
}
