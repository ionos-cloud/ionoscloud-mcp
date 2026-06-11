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
		Annotations: tools.ReadOnly,
		Description: "List all network (layer-4) load balancers in a data center. For HTTP-aware layer-7 balancing use list_application_loadbalancers instead.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.DatacenterIDInput) (*mcp.CallToolResult, any, error) {
		nlbs, _, err := client.NetworkLoadBalancersApi.DatacentersNetworkloadbalancersGet(ctx, input.DatacenterID).Execute()
		return tools.ToResult(nlbs, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_network_loadbalancer",
		Annotations: tools.ReadOnly,
		Description: "Get details of a single network load balancer (NLB): listener/target LANs, IPs, and state. Use list_network_loadbalancers to find IDs; use list_nlb_forwarding_rules for its forwarding rules.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.NetworkLoadBalancerIDInput) (*mcp.CallToolResult, any, error) {
		nlb, _, err := client.NetworkLoadBalancersApi.DatacentersNetworkloadbalancersFindByNetworkLoadBalancerId(ctx, input.DatacenterID, input.NetworkLoadBalancerID).Execute()
		return tools.ToResult(nlb, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_nlb_forwarding_rules",
		Annotations: tools.ReadOnly,
		Description: "List all forwarding rules of a network load balancer: listener IP/port, balancing algorithm, health check, and backend targets. Use list_network_loadbalancers first to find the NLB ID.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.NetworkLoadBalancerIDInput) (*mcp.CallToolResult, any, error) {
		rules, _, err := client.NetworkLoadBalancersApi.DatacentersNetworkloadbalancersForwardingrulesGet(ctx, input.DatacenterID, input.NetworkLoadBalancerID).Execute()
		return tools.ToResult(rules, err)
	})
}
