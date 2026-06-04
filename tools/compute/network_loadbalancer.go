package compute

import (
	"context"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterNetworkLoadBalancerTools(server *mcp.Server, client *ionos.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_network_loadbalancers",
		Description: "List all network load balancers (NLB) in a data center",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.DatacenterIDInput) (*mcp.CallToolResult, any, error) {
		nlbs, _, err := client.NetworkLoadBalancersApi.DatacentersNetworkloadbalancersGet(ctx, input.DatacenterID).Execute()
		return tools.ToResult(nlbs, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_network_loadbalancer",
		Description: "Get details of a specific network load balancer (NLB)",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.NetworkLoadBalancerIDInput) (*mcp.CallToolResult, any, error) {
		nlb, _, err := client.NetworkLoadBalancersApi.DatacentersNetworkloadbalancersFindByNetworkLoadBalancerId(ctx, input.DatacenterID, input.NetworkLoadBalancerID).Execute()
		return tools.ToResult(nlb, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_nlb_forwarding_rules",
		Description: "List all forwarding rules of a network load balancer",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.NetworkLoadBalancerIDInput) (*mcp.CallToolResult, any, error) {
		rules, _, err := client.NetworkLoadBalancersApi.DatacentersNetworkloadbalancersForwardingrulesGet(ctx, input.DatacenterID, input.NetworkLoadBalancerID).Execute()
		return tools.ToResult(rules, err)
	})
}
