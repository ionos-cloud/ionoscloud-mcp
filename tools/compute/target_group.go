package compute

import (
	"context"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterTargetGroupTools(server *mcp.Server, client *ionos.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_target_groups",
		Annotations: tools.ReadOnly,
		Description: "List all target groups in your IONOS CLOUD account — the backend pools (IP/port targets plus health checks) referenced by application load balancer forwarding rules.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		tgs, _, err := client.TargetGroupsApi.TargetgroupsGet(ctx).Execute()
		return tools.ToResult(tgs, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_target_group",
		Annotations: tools.ReadOnly,
		Description: "Get details of a single target group: balancing algorithm, protocol, targets, and health-check settings. Use list_target_groups to find IDs; see list_alb_forwarding_rules for where a group is used.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.TargetGroupIDInput) (*mcp.CallToolResult, any, error) {
		tg, _, err := client.TargetGroupsApi.TargetgroupsFindByTargetGroupId(ctx, input.TargetGroupID).Execute()
		return tools.ToResult(tg, err)
	})
}
