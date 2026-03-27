package compute

import (
	"context"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterNicTools(server *mcp.Server, client *ionos.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_nics",
		Description: "List all network interfaces (NICs) attached to a server",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ServerIDInput) (*mcp.CallToolResult, any, error) {
		nics, _, err := client.NetworkInterfacesApi.DatacentersServersNicsGet(ctx, input.DatacenterID, input.ServerID).Execute()
		return tools.ToResult(nics, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_nic",
		Description: "Get details of a specific network interface (NIC)",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.NicIDInput) (*mcp.CallToolResult, any, error) {
		nic, _, err := client.NetworkInterfacesApi.DatacentersServersNicsFindById(ctx, input.DatacenterID, input.ServerID, input.NicID).Execute()
		return tools.ToResult(nic, err)
	})
}
