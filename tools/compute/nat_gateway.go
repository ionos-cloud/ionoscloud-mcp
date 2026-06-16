package compute

import (
	"context"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterNatGatewayTools(server *mcp.Server, client *ionos.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_nat_gateways",
		Description: "List all NAT gateways in a data center",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.DatacenterIDInput) (*mcp.CallToolResult, any, error) {
		depth := int32(1)
		if input.Depth != nil {
			depth = *input.Depth
		}
		nats, _, err := client.NATGatewaysApi.DatacentersNatgatewaysGet(ctx, input.DatacenterID).Depth(depth).Execute()
		return tools.ToResult(nats, err)
	})

	mcp.AddTool(server, &mcp.Tool{
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

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_nat_gateway_rules",
		Description: "List all rules of a specific NAT gateway",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.NatGatewayIDInput) (*mcp.CallToolResult, any, error) {
		depth := int32(1)
		if input.Depth != nil {
			depth = *input.Depth
		}
		rules, _, err := client.NATGatewaysApi.DatacentersNatgatewaysRulesGet(ctx, input.DatacenterID, input.NatGatewayID).Depth(depth).Execute()
		return tools.ToResult(rules, err)
	})
}
