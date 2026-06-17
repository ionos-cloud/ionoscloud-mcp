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
		Description: "List all security groups in a data center",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.DatacenterIDInput) (*mcp.CallToolResult, any, error) {
		depth := int32(1)
		if input.Depth != nil {
			depth = *input.Depth
		}
		sgs, _, err := client.SecurityGroupsApi.DatacentersSecuritygroupsGet(ctx, input.DatacenterID).Depth(depth).Execute()
		return tools.ToResult(sgs, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_security_group",
		Description: "Get details of a specific security group",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.SecurityGroupIDInput) (*mcp.CallToolResult, any, error) {
		apiReq := client.SecurityGroupsApi.DatacentersSecuritygroupsFindById(ctx, input.DatacenterID, input.SecurityGroupID)
		if input.Depth != nil {
			apiReq = apiReq.Depth(*input.Depth)
		}
		sg, _, err := apiReq.Execute()
		return tools.ToResult(sg, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_security_group_rules",
		Description: "List all rules in a specific security group",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.SecurityGroupIDInput) (*mcp.CallToolResult, any, error) {
		depth := int32(1)
		if input.Depth != nil {
			depth = *input.Depth
		}
		rules, _, err := client.SecurityGroupsApi.DatacentersSecuritygroupsRulesGet(ctx, input.DatacenterID, input.SecurityGroupID).Depth(depth).Execute()
		return tools.ToResult(rules, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_security_group_rule",
		Description: "Get details of a specific security group rule",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.SecurityGroupRuleIDInput) (*mcp.CallToolResult, any, error) {
		apiReq := client.SecurityGroupsApi.DatacentersSecuritygroupsRulesFindById(ctx, input.DatacenterID, input.SecurityGroupID, input.RuleID)
		if input.Depth != nil {
			apiReq = apiReq.Depth(*input.Depth)
		}
		rule, _, err := apiReq.Execute()
		return tools.ToResult(rule, err)
	})
}
