package compute

import (
	"context"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterSecurityGroupTools(server *mcp.Server, client *ionos.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_security_groups",
		Description: "List all security groups in a data center",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.DatacenterIDInput) (*mcp.CallToolResult, any, error) {
		sgs, _, err := client.SecurityGroupsApi.DatacentersSecuritygroupsGet(ctx, input.DatacenterID).Execute()
		return tools.ToResult(sgs, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_security_group",
		Description: "Get details of a specific security group",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.SecurityGroupIDInput) (*mcp.CallToolResult, any, error) {
		sg, _, err := client.SecurityGroupsApi.DatacentersSecuritygroupsFindById(ctx, input.DatacenterID, input.SecurityGroupID).Execute()
		return tools.ToResult(sg, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_security_group_rules",
		Description: "List all rules in a specific security group",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.SecurityGroupIDInput) (*mcp.CallToolResult, any, error) {
		rules, _, err := client.SecurityGroupsApi.DatacentersSecuritygroupsRulesGet(ctx, input.DatacenterID, input.SecurityGroupID).Execute()
		return tools.ToResult(rules, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_security_group_rule",
		Description: "Get details of a specific security group rule",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.SecurityGroupRuleIDInput) (*mcp.CallToolResult, any, error) {
		rule, _, err := client.SecurityGroupsApi.DatacentersSecuritygroupsRulesFindById(ctx, input.DatacenterID, input.SecurityGroupID, input.RuleID).Execute()
		return tools.ToResult(rule, err)
	})
}
