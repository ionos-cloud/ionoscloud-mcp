package compute

import (
	"context"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterLanTools(server *mcp.Server, client *ionos.APIClient, scope tools.Scope) {
	tools.RegisterTool(server, scope, tools.MethodGet, &mcp.Tool{
		Name:        "list_lans",
		Description: "List all LANs in a data center",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ListInDatacenterInput) (*mcp.CallToolResult, any, error) {
		depth := int32(1)
		if input.Depth != nil {
			depth = *input.Depth
		}
		r := client.LANsApi.DatacentersLansGet(ctx, input.DatacenterID).Depth(depth)
		for k, v := range input.Filters {
			r = r.Filter(k, v)
		}
		lans, _, err := r.Execute()
		return tools.ToResult(lans, err)
	})

	tools.RegisterTool(server, scope, tools.MethodGet, &mcp.Tool{
		Name:        "get_lan",
		Description: "Get details of a specific LAN",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.LanIDInput) (*mcp.CallToolResult, any, error) {
		apiReq := client.LANsApi.DatacentersLansFindById(ctx, input.DatacenterID, input.LanID)
		if input.Depth != nil {
			apiReq = apiReq.Depth(*input.Depth)
		}
		lan, _, err := apiReq.Execute()
		return tools.ToResult(lan, err)
	})

	tools.RegisterTool(server, scope, tools.MethodGet, &mcp.Tool{
		Name:        "list_lan_nics",
		Description: "List all NICs attached to a specific LAN",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ListLanNicsInput) (*mcp.CallToolResult, any, error) {
		depth := int32(1)
		if input.Depth != nil {
			depth = *input.Depth
		}
		r := client.LANsApi.DatacentersLansNicsGet(ctx, input.DatacenterID, input.LanID).Depth(depth)
		for k, v := range input.Filters {
			r = r.Filter(k, v)
		}
		nics, _, err := r.Execute()
		return tools.ToResult(nics, err)
	})
}
