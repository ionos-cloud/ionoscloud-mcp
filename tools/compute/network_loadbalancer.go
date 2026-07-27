package compute

import (
	"context"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterNetworkLoadBalancerTools(server *mcp.Server, client *ionos.APIClient, scope tools.Scope) {
	tools.RegisterTool(server, scope, tools.MethodGet, &mcp.Tool{
		Name:        "list_network_loadbalancers",
		Description: "List all network load balancers (NLB) in a data center",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ListInDatacenterInput) (*mcp.CallToolResult, any, error) {
		depth := int32(1)
		if input.Depth != nil {
			depth = *input.Depth
		}
		r := client.NetworkLoadBalancersApi.DatacentersNetworkloadbalancersGet(ctx, input.DatacenterID).Depth(depth)
		for k, v := range input.Filters {
			r = r.Filter(k, v)
		}
		nlbs, _, err := r.Execute()
		return tools.ToResult(nlbs, err)
	})

	tools.RegisterTool(server, scope, tools.MethodGet, &mcp.Tool{
		Name:        "get_network_loadbalancer",
		Description: "Get details of a specific network load balancer (NLB)",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.NetworkLoadBalancerIDInput) (*mcp.CallToolResult, any, error) {
		apiReq := client.NetworkLoadBalancersApi.DatacentersNetworkloadbalancersFindByNetworkLoadBalancerId(ctx, input.DatacenterID, input.NetworkLoadBalancerID)
		if input.Depth != nil {
			apiReq = apiReq.Depth(*input.Depth)
		}
		nlb, _, err := apiReq.Execute()
		return tools.ToResult(nlb, err)
	})

	tools.RegisterTool(server, scope, tools.MethodGet, &mcp.Tool{
		Name:        "list_nlb_forwarding_rules",
		Description: "List all forwarding rules of a network load balancer",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ListNlbForwardingRulesInput) (*mcp.CallToolResult, any, error) {
		depth := int32(1)
		if input.Depth != nil {
			depth = *input.Depth
		}
		r := client.NetworkLoadBalancersApi.DatacentersNetworkloadbalancersForwardingrulesGet(ctx, input.DatacenterID, input.NetworkLoadBalancerID).Depth(depth)
		for k, v := range input.Filters {
			r = r.Filter(k, v)
		}
		rules, _, err := r.Execute()
		return tools.ToResult(rules, err)
	})
}
