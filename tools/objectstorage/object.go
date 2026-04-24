package objectstorage

import (
	"context"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
	sdk "github.com/ionos-cloud/sdk-go-bundle/products/objectstorage/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterObjectTools(server *mcp.Server, client *sdk.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_object_storage_objects",
		Description: "List objects in an Object Storage bucket. Supports an optional prefix to filter by key path (e.g. 'images/' to list only objects under that prefix). Returns up to 1000 objects per call; use the next_continuation_token from the response to page through larger result sets.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ObjectStorageListObjectsInput) (*mcp.CallToolResult, any, error) {
		listReq := client.ObjectsApi.ListObjectsV2(ctx, input.Bucket)
		if input.Prefix != nil {
			listReq = listReq.Prefix(*input.Prefix)
		}
		result, _, err := listReq.Execute()
		return tools.ToResult(result, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "head_object_storage_object",
		Description: "Check whether an object exists in an Object Storage bucket and retrieve its user-defined metadata (x-amz-meta-* headers). Returns an error if the object does not exist or is not accessible.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ObjectStorageObjectInput) (*mcp.CallToolResult, any, error) {
		output, _, err := client.ObjectsApi.HeadObject(ctx, input.Bucket, input.Key).Execute()
		return tools.ToResult(output, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_object_storage_object_versions",
		Description: "List all versions of objects in an Object Storage bucket. Requires versioning to be enabled on the bucket. Supports an optional prefix filter.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ObjectStorageListObjectVersionsInput) (*mcp.CallToolResult, any, error) {
		listReq := client.VersionsApi.ListObjectVersions(ctx, input.Bucket)
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
		result, _, err := client.TaggingApi.GetObjectTagging(ctx, input.Bucket, input.Key).Execute()
		return tools.ToResult(result, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_object_storage_object_retention",
		Description: "Get the Object Lock retention configuration for an object in an Object Storage bucket.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ObjectStorageObjectInput) (*mcp.CallToolResult, any, error) {
		result, _, err := client.ObjectLockApi.GetObjectRetention(ctx, input.Bucket, input.Key).Execute()
		return tools.ToResult(result, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_object_storage_object_legal_hold",
		Description: "Get the Object Lock legal hold status for an object in an Object Storage bucket.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ObjectStorageObjectInput) (*mcp.CallToolResult, any, error) {
		result, _, err := client.ObjectLockApi.GetObjectLegalHold(ctx, input.Bucket, input.Key).Execute()
		return tools.ToResult(result, err)
	})
}
