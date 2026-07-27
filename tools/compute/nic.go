package compute

import (
	"context"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterNicTools(server *mcp.Server, client *ionos.APIClient, scope tools.Scope) {
	tools.RegisterTool(server, scope, tools.MethodGet, &mcp.Tool{
		Name:        "list_nics",
		Description: "List all network interfaces (NICs) attached to a server",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ListInServerInput) (*mcp.CallToolResult, any, error) {
		depth := int32(1)
		if input.Depth != nil {
			depth = *input.Depth
		}
		r := client.NetworkInterfacesApi.DatacentersServersNicsGet(ctx, input.DatacenterID, input.ServerID).Depth(depth)
		for k, v := range input.Filters {
			r = r.Filter(k, v)
		}
		nics, _, err := r.Execute()
		return tools.ToResult(nics, err)
	})

	tools.RegisterTool(server, scope, tools.MethodGet, &mcp.Tool{
		Name:        "get_nic",
		Description: "Get details of a specific network interface (NIC)",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.NicIDInput) (*mcp.CallToolResult, any, error) {
		apiReq := client.NetworkInterfacesApi.DatacentersServersNicsFindById(ctx, input.DatacenterID, input.ServerID, input.NicID)
		if input.Depth != nil {
			apiReq = apiReq.Depth(*input.Depth)
		}
		nic, _, err := apiReq.Execute()
		return tools.ToResult(nic, err)
	})
}
