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
		Annotations: tools.ReadOnly,
		Description: "List all firewall rules configured on a specific NIC, with protocol, port ranges, and source/target restrictions. Firewall rules are scoped per NIC — use list_nics to find the NIC ID. For datacenter-level security groups use list_security_group_rules instead.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.NicIDInput) (*mcp.CallToolResult, any, error) {
		rules, _, err := client.FirewallRulesApi.DatacentersServersNicsFirewallrulesGet(ctx, input.DatacenterID, input.ServerID, input.NicID).Execute()
		return tools.ToResult(rules, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_firewall_rule",
		Annotations: tools.ReadOnly,
		Description: "Get the full definition of a single firewall rule on a NIC: protocol, port/ICMP ranges, source MAC/IP, and target IP. Use list_firewall_rules first to find the rule ID.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.FirewallRuleIDInput) (*mcp.CallToolResult, any, error) {
		rule, _, err := client.FirewallRulesApi.DatacentersServersNicsFirewallrulesFindById(ctx, input.DatacenterID, input.ServerID, input.NicID, input.FirewallRuleID).Execute()
		return tools.ToResult(rule, err)
	})
}
