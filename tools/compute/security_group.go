package compute

import (
	"context"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterSecurityGroupTools(server *mcp.Server, client *ionos.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_security_groups",
		Annotations: tools.ReadOnly,
		Description: "List all security groups in a data center, with their rules and attached resources. Use list_datacenters first to find the datacenter ID.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.DatacenterIDInput) (*mcp.CallToolResult, any, error) {
		sgs, _, err := client.SecurityGroupsApi.DatacentersSecuritygroupsGet(ctx, input.DatacenterID).Execute()
		return tools.ToResult(sgs, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_security_group",
		Annotations: tools.ReadOnly,
		Description: "Get a single security group with its rules and the servers/NICs it is applied to. Use list_security_groups to find group IDs; use get_security_group_rule for one rule's full definition.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.SecurityGroupIDInput) (*mcp.CallToolResult, any, error) {
		sg, _, err := client.SecurityGroupsApi.DatacentersSecuritygroupsFindById(ctx, input.DatacenterID, input.SecurityGroupID).Execute()
		return tools.ToResult(sg, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_security_group_rules",
		Annotations: tools.ReadOnly,
		Description: "List all rules of a specific security group: direction, protocol, port ranges, and remote IP. Distinct from list_firewall_rules, which covers per-NIC firewall rules.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.SecurityGroupIDInput) (*mcp.CallToolResult, any, error) {
		rules, _, err := client.SecurityGroupsApi.DatacentersSecuritygroupsRulesGet(ctx, input.DatacenterID, input.SecurityGroupID).Execute()
		return tools.ToResult(rules, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_security_group_rule",
		Annotations: tools.ReadOnly,
		Description: "Get the full definition of a single security group rule: direction, protocol, port/ICMP ranges, and remote IP. Use list_security_group_rules first to find the rule ID.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.SecurityGroupRuleIDInput) (*mcp.CallToolResult, any, error) {
		rule, _, err := client.SecurityGroupsApi.DatacentersSecuritygroupsRulesFindById(ctx, input.DatacenterID, input.SecurityGroupID, input.RuleID).Execute()
		return tools.ToResult(rule, err)
	})
}
