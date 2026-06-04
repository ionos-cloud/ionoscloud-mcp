package compute

import (
	"context"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterLanTools(server *mcp.Server, client *ionos.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_lans",
		Description: "List all LANs in a data center",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.DatacenterIDInput) (*mcp.CallToolResult, any, error) {
		lans, _, err := client.LANsApi.DatacentersLansGet(ctx, input.DatacenterID).Execute()
		return tools.ToResult(lans, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_lan",
		Description: "Get details of a specific LAN",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.LanIDInput) (*mcp.CallToolResult, any, error) {
		lan, _, err := client.LANsApi.DatacentersLansFindById(ctx, input.DatacenterID, input.LanID).Execute()
		return tools.ToResult(lan, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_lan_nics",
		Description: "List all NICs attached to a specific LAN",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.LanIDInput) (*mcp.CallToolResult, any, error) {
		nics, _, err := client.LANsApi.DatacentersLansNicsGet(ctx, input.DatacenterID, input.LanID).Execute()
		return tools.ToResult(nics, err)
	})
}
