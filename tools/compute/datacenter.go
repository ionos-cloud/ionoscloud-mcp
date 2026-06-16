package compute

import (
	"context"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterDatacenterTools(server *mcp.Server, client *ionos.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_datacenters",
		Description: "List all virtual data centers. Returns names and basic properties by default (depth=1).",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ListDatacentersInput) (*mcp.CallToolResult, any, error) {
		depth := int32(1)
		if input.Depth != nil {
			depth = *input.Depth
		}
		datacenters, _, err := client.DataCentersApi.DatacentersGet(ctx).Depth(depth).Execute()
		return tools.ToResult(datacenters, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_datacenter",
		Description: "Get details of a specific virtual data center",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.DatacenterIDInput) (*mcp.CallToolResult, any, error) {
		apiReq := client.DataCentersApi.DatacentersFindById(ctx, input.DatacenterID)
		if input.Depth != nil {
			apiReq = apiReq.Depth(*input.Depth)
		}
		datacenter, _, err := apiReq.Execute()
		return tools.ToResult(datacenter, err)
	})
}
