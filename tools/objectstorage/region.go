package objectstorage

import (
	"context"

	mgmtSDK "github.com/ionos-cloud/sdk-go-bundle/products/objectstoragemanagement/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterRegionTools(server *mcp.Server, client *mgmtSDK.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_object_storage_regions",
		Annotations: tools.ReadOnly,
		Description: "List all available Object Storage regions.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		result, _, err := client.RegionsApi.RegionsGet(ctx).Execute()
		return tools.ToResult(result, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_object_storage_region",
		Annotations: tools.ReadOnly,
		Description: "Get details of a specific Object Storage region by name (e.g. eu-central-3).",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ObjectStorageRegionInput) (*mcp.CallToolResult, any, error) {
		result, _, err := client.RegionsApi.RegionsFindByRegion(ctx, input.Region).Execute()
		return tools.ToResult(result, err)
	})
}
