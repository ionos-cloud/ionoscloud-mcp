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
		Description: "List all virtual data centers. Returns names and basic properties by default (depth=1). Use name to filter by datacenter name.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ListDatacentersInput) (*mcp.CallToolResult, any, error) {
		depth := int32(1)
		if input.Depth != nil {
			depth = *input.Depth
		}
		apiReq := client.DataCentersApi.DatacentersGet(ctx).Depth(depth)
		if input.Name != nil {
			apiReq = apiReq.Filter("properties.name", *input.Name)
		}
		datacenters, _, err := apiReq.Execute()
		return tools.ToResult(datacenters, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_datacenter",
		Description: "Get details of a specific virtual data center",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.DatacenterIDInput) (*mcp.CallToolResult, any, error) {
		datacenter, _, err := client.DataCentersApi.DatacentersFindById(ctx, input.DatacenterID).Execute()
		return tools.ToResult(datacenter, err)
	})
}
