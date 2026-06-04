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
		Description: "List all virtual data centers in your IONOS CLOUD account",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		datacenters, _, err := client.DataCentersApi.DatacentersGet(ctx).Execute()
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
