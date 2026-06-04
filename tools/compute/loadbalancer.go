package compute

import (
	"context"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterLoadBalancerTools(server *mcp.Server, client *ionos.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_loadbalancers",
		Description: "List all load balancers in a data center",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.DatacenterIDInput) (*mcp.CallToolResult, any, error) {
		lbs, _, err := client.LoadBalancersApi.DatacentersLoadbalancersGet(ctx, input.DatacenterID).Execute()
		return tools.ToResult(lbs, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_loadbalancer",
		Description: "Get details of a specific load balancer",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.LoadBalancerIDInput) (*mcp.CallToolResult, any, error) {
		lb, _, err := client.LoadBalancersApi.DatacentersLoadbalancersFindById(ctx, input.DatacenterID, input.LoadBalancerID).Execute()
		return tools.ToResult(lb, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_loadbalancer_nics",
		Description: "List all NICs balanced by a specific load balancer",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.LoadBalancerIDInput) (*mcp.CallToolResult, any, error) {
		nics, _, err := client.LoadBalancersApi.DatacentersLoadbalancersBalancednicsGet(ctx, input.DatacenterID, input.LoadBalancerID).Execute()
		return tools.ToResult(nics, err)
	})
}
