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
		Annotations: tools.ReadOnly,
		Description: "List all NAT gateways in a data center, with their public IPs and LAN attachments. Use list_datacenters first to find the datacenter ID.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.DatacenterIDInput) (*mcp.CallToolResult, any, error) {
		nats, _, err := client.NATGatewaysApi.DatacentersNatgatewaysGet(ctx, input.DatacenterID).Execute()
		return tools.ToResult(nats, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_nat_gateway",
		Annotations: tools.ReadOnly,
		Description: "Get details of a single NAT gateway: public IPs and the LANs it serves. Use list_nat_gateways to find IDs; use list_nat_gateway_rules for its translation rules.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.NatGatewayIDInput) (*mcp.CallToolResult, any, error) {
		nat, _, err := client.NATGatewaysApi.DatacentersNatgatewaysFindByNatGatewayId(ctx, input.DatacenterID, input.NatGatewayID).Execute()
		return tools.ToResult(nat, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_nat_gateway_rules",
		Annotations: tools.ReadOnly,
		Description: "List all NAT rules of a specific NAT gateway: type, protocol, source/target subnets, and translated public IP. Use list_nat_gateways first to find the gateway ID.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.NatGatewayIDInput) (*mcp.CallToolResult, any, error) {
		rules, _, err := client.NATGatewaysApi.DatacentersNatgatewaysRulesGet(ctx, input.DatacenterID, input.NatGatewayID).Execute()
		return tools.ToResult(rules, err)
	})
}
