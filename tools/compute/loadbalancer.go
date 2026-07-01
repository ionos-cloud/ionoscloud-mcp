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
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ListInDatacenterInput) (*mcp.CallToolResult, any, error) {
		depth := int32(1)
		if input.Depth != nil {
			depth = *input.Depth
		}
		r := client.LoadBalancersApi.DatacentersLoadbalancersGet(ctx, input.DatacenterID).Depth(depth)
		for k, v := range input.Filters {
			r = r.Filter(k, v)
		}
		lbs, _, err := r.Execute()
		return tools.ToResult(lbs, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_loadbalancer",
		Description: "Get details of a specific load balancer",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.LoadBalancerIDInput) (*mcp.CallToolResult, any, error) {
		apiReq := client.LoadBalancersApi.DatacentersLoadbalancersFindById(ctx, input.DatacenterID, input.LoadBalancerID)
		if input.Depth != nil {
			apiReq = apiReq.Depth(*input.Depth)
		}
		lb, _, err := apiReq.Execute()
		return tools.ToResult(lb, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_loadbalancer_nics",
		Description: "List all NICs balanced by a specific load balancer",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ListLoadBalancerNicsInput) (*mcp.CallToolResult, any, error) {
		depth := int32(1)
		if input.Depth != nil {
			depth = *input.Depth
		}
		r := client.LoadBalancersApi.DatacentersLoadbalancersBalancednicsGet(ctx, input.DatacenterID, input.LoadBalancerID).Depth(depth)
		for k, v := range input.Filters {
			r = r.Filter(k, v)
		}
		nics, _, err := r.Execute()
		return tools.ToResult(nics, err)
	})
}
