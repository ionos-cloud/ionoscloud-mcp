package compute

import (
	"context"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterNatGatewayTools(server *mcp.Server, client *ionos.APIClient, scope tools.Scope) {
	tools.RegisterTool(server, scope, tools.MethodGet, &mcp.Tool{
		Name:        "list_nat_gateways",
		Description: "List all NAT gateways in a data center",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ListInDatacenterInput) (*mcp.CallToolResult, any, error) {
		depth := int32(1)
		if input.Depth != nil {
			depth = *input.Depth
		}
		r := client.NATGatewaysApi.DatacentersNatgatewaysGet(ctx, input.DatacenterID).Depth(depth)
		for k, v := range input.Filters {
			r = r.Filter(k, v)
		}
		nats, _, err := r.Execute()
		return tools.ToResult(nats, err)
	})

	tools.RegisterTool(server, scope, tools.MethodGet, &mcp.Tool{
		Name:        "get_nat_gateway",
		Description: "Get details of a specific NAT gateway",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.NatGatewayIDInput) (*mcp.CallToolResult, any, error) {
		apiReq := client.NATGatewaysApi.DatacentersNatgatewaysFindByNatGatewayId(ctx, input.DatacenterID, input.NatGatewayID)
		if input.Depth != nil {
			apiReq = apiReq.Depth(*input.Depth)
		}
		nat, _, err := apiReq.Execute()
		return tools.ToResult(nat, err)
	})

	tools.RegisterTool(server, scope, tools.MethodGet, &mcp.Tool{
		Name:        "list_nat_gateway_rules",
		Description: "List all rules of a specific NAT gateway",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ListNatGatewayRulesInput) (*mcp.CallToolResult, any, error) {
		depth := int32(1)
		if input.Depth != nil {
			depth = *input.Depth
		}
		r := client.NATGatewaysApi.DatacentersNatgatewaysRulesGet(ctx, input.DatacenterID, input.NatGatewayID).Depth(depth)
		for k, v := range input.Filters {
			r = r.Filter(k, v)
		}
		rules, _, err := r.Execute()
		return tools.ToResult(rules, err)
	})
}
