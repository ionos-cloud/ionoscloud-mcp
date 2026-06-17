package compute

import (
	"context"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterFirewallRuleTools(server *mcp.Server, client *ionos.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_firewall_rules",
		Description: "List all firewall rules on a network interface",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.NicIDInput) (*mcp.CallToolResult, any, error) {
		depth := int32(1)
		if input.Depth != nil {
			depth = *input.Depth
		}
		rules, _, err := client.FirewallRulesApi.DatacentersServersNicsFirewallrulesGet(ctx, input.DatacenterID, input.ServerID, input.NicID).Depth(depth).Execute()
		return tools.ToResult(rules, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_firewall_rule",
		Description: "Get details of a specific firewall rule",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.FirewallRuleIDInput) (*mcp.CallToolResult, any, error) {
		apiReq := client.FirewallRulesApi.DatacentersServersNicsFirewallrulesFindById(ctx, input.DatacenterID, input.ServerID, input.NicID, input.FirewallRuleID)
		if input.Depth != nil {
			apiReq = apiReq.Depth(*input.Depth)
		}
		rule, _, err := apiReq.Execute()
		return tools.ToResult(rule, err)
	})
}
