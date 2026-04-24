package objectstorage

import (
	"context"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
	sdk "github.com/ionos-cloud/sdk-go-bundle/products/objectstorage/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterBucketTools(server *mcp.Server, client *sdk.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_object_storage_buckets",
		Description: "List all Object Storage buckets owned by the contract.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		result, _, err := client.BucketsApi.ListBuckets(ctx).Execute()
		return tools.ToResult(result, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_object_storage_bucket_location",
		Description: "Get the region/location of an Object Storage bucket.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ObjectStorageBucketInput) (*mcp.CallToolResult, any, error) {
		result, _, err := client.BucketsApi.GetBucketLocation(ctx, input.Bucket).Execute()
		return tools.ToResult(result, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "head_object_storage_bucket",
		Description: "Check whether an Object Storage bucket exists and is accessible with the current credentials.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ObjectStorageBucketInput) (*mcp.CallToolResult, any, error) {
		_, err := client.BucketsApi.HeadBucket(ctx, input.Bucket).Execute()
		if err != nil {
			return tools.ToResult(nil, err)
		}
		return tools.ToResult(map[string]any{"bucket": input.Bucket, "accessible": true}, nil)
	})
}
