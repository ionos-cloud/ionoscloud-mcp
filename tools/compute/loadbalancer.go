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
		Annotations: tools.ReadOnly,
		Description: "List all classic (layer-4) load balancers in a data center. For the newer products use list_network_loadbalancers (NLB) or list_application_loadbalancers (ALB) instead.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.DatacenterIDInput) (*mcp.CallToolResult, any, error) {
		lbs, _, err := client.LoadBalancersApi.DatacentersLoadbalancersGet(ctx, input.DatacenterID).Execute()
		return tools.ToResult(lbs, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_loadbalancer",
		Annotations: tools.ReadOnly,
		Description: "Get details of a single classic load balancer: listener IP, DHCP setting, and state. Use list_loadbalancers to find IDs; for NLB/ALB use get_network_loadbalancer or get_application_loadbalancer instead.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.LoadBalancerIDInput) (*mcp.CallToolResult, any, error) {
		lb, _, err := client.LoadBalancersApi.DatacentersLoadbalancersFindById(ctx, input.DatacenterID, input.LoadBalancerID).Execute()
		return tools.ToResult(lb, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_loadbalancer_nics",
		Annotations: tools.ReadOnly,
		Description: "List the server NICs a classic load balancer distributes traffic across. Use list_loadbalancers first to find the load balancer ID.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.LoadBalancerIDInput) (*mcp.CallToolResult, any, error) {
		nics, _, err := client.LoadBalancersApi.DatacentersLoadbalancersBalancednicsGet(ctx, input.DatacenterID, input.LoadBalancerID).Execute()
		return tools.ToResult(nics, err)
	})
}
