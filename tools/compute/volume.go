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
		Annotations: tools.ReadOnly,
		Description: "List all storage volumes in a data center — attached or not — with size, type, and source image. Use list_server_volumes instead to see only the volumes attached to one server.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.DatacenterIDInput) (*mcp.CallToolResult, any, error) {
		volumes, _, err := client.VolumesApi.DatacentersVolumesGet(ctx, input.DatacenterID).Execute()
		return tools.ToResult(volumes, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_volume",
		Annotations: tools.ReadOnly,
		Description: "Get details of a single storage volume: size, type (HDD or SSD), source image, boot flags, and bus. Use list_volumes to find volume IDs.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.VolumeIDInput) (*mcp.CallToolResult, any, error) {
		volume, _, err := client.VolumesApi.DatacentersVolumesFindById(ctx, input.DatacenterID, input.VolumeID).Execute()
		return tools.ToResult(volume, err)
	})
}
