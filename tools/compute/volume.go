package compute

import (
	"context"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterVolumeTools(server *mcp.Server, client *ionos.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_volumes",
		Description: "List all volumes in a data center",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.DatacenterIDInput) (*mcp.CallToolResult, any, error) {
		depth := int32(1)
		if input.Depth != nil {
			depth = *input.Depth
		}
		volumes, _, err := client.VolumesApi.DatacentersVolumesGet(ctx, input.DatacenterID).Depth(depth).Execute()
		return tools.ToResult(volumes, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_volume",
		Description: "Get details of a specific volume",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.VolumeIDInput) (*mcp.CallToolResult, any, error) {
		apiReq := client.VolumesApi.DatacentersVolumesFindById(ctx, input.DatacenterID, input.VolumeID)
		if input.Depth != nil {
			apiReq = apiReq.Depth(*input.Depth)
		}
		volume, _, err := apiReq.Execute()
		return tools.ToResult(volume, err)
	})
}
