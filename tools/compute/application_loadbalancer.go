package compute

import (
	"context"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterApplicationLoadBalancerTools(server *mcp.Server, client *ionos.APIClient, scope tools.Scope) {
	tools.RegisterTool(server, scope, tools.MethodGet, &mcp.Tool{
		Name:        "list_application_loadbalancers",
		Description: "List all application load balancers (ALB) in a data center",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ListInDatacenterInput) (*mcp.CallToolResult, any, error) {
		depth := int32(1)
		if input.Depth != nil {
			depth = *input.Depth
		}
		r := client.ApplicationLoadBalancersApi.DatacentersApplicationloadbalancersGet(ctx, input.DatacenterID).Depth(depth)
		for k, v := range input.Filters {
			r = r.Filter(k, v)
		}
		albs, _, err := r.Execute()
		return tools.ToResult(albs, err)
	})

	tools.RegisterTool(server, scope, tools.MethodGet, &mcp.Tool{
		Name:        "get_application_loadbalancer",
		Description: "Get details of a specific application load balancer (ALB)",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ApplicationLoadBalancerIDInput) (*mcp.CallToolResult, any, error) {
		apiReq := client.ApplicationLoadBalancersApi.DatacentersApplicationloadbalancersFindByApplicationLoadBalancerId(ctx, input.DatacenterID, input.ApplicationLoadBalancerID)
		if input.Depth != nil {
			apiReq = apiReq.Depth(*input.Depth)
		}
		alb, _, err := apiReq.Execute()
		return tools.ToResult(alb, err)
	})

	tools.RegisterTool(server, scope, tools.MethodGet, &mcp.Tool{
		Name:        "list_alb_forwarding_rules",
		Description: "List all forwarding rules of an application load balancer",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ListAlbForwardingRulesInput) (*mcp.CallToolResult, any, error) {
		depth := int32(1)
		if input.Depth != nil {
			depth = *input.Depth
		}
		r := client.ApplicationLoadBalancersApi.DatacentersApplicationloadbalancersForwardingrulesGet(ctx, input.DatacenterID, input.ApplicationLoadBalancerID).Depth(depth)
		for k, v := range input.Filters {
			r = r.Filter(k, v)
		}
		rules, _, err := r.Execute()
		return tools.ToResult(rules, err)
	})
}
