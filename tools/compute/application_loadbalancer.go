package compute

import (
	"context"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterApplicationLoadBalancerTools(server *mcp.Server, client *ionos.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_application_loadbalancers",
		Annotations: tools.ReadOnly,
		Description: "List all application (layer-7/HTTP) load balancers in a data center. For layer-4 TCP/UDP balancing use list_network_loadbalancers instead.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.DatacenterIDInput) (*mcp.CallToolResult, any, error) {
		albs, _, err := client.ApplicationLoadBalancersApi.DatacentersApplicationloadbalancersGet(ctx, input.DatacenterID).Execute()
		return tools.ToResult(albs, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_application_loadbalancer",
		Annotations: tools.ReadOnly,
		Description: "Get details of a single application load balancer (ALB): listener/target LANs, IPs, and state. Use list_application_loadbalancers to find IDs; use list_alb_forwarding_rules for its HTTP routing rules.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ApplicationLoadBalancerIDInput) (*mcp.CallToolResult, any, error) {
		alb, _, err := client.ApplicationLoadBalancersApi.DatacentersApplicationloadbalancersFindByApplicationLoadBalancerId(ctx, input.DatacenterID, input.ApplicationLoadBalancerID).Execute()
		return tools.ToResult(alb, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_alb_forwarding_rules",
		Annotations: tools.ReadOnly,
		Description: "List all forwarding rules of an application load balancer: listener protocol/port, HTTP conditions, and target-group references (see list_target_groups). Use list_application_loadbalancers first to find the ALB ID.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ApplicationLoadBalancerIDInput) (*mcp.CallToolResult, any, error) {
		rules, _, err := client.ApplicationLoadBalancersApi.DatacentersApplicationloadbalancersForwardingrulesGet(ctx, input.DatacenterID, input.ApplicationLoadBalancerID).Execute()
		return tools.ToResult(rules, err)
	})
}
