package compute

import (
	"context"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterTargetGroupTools(server *mcp.Server, client *ionos.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_target_groups",
		Description: "List all target groups in your IONOS Cloud account",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		tgs, _, err := client.TargetGroupsApi.TargetgroupsGet(ctx).Execute()
		return tools.ToResult(tgs, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_target_group",
		Description: "Get details of a specific target group",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.TargetGroupIDInput) (*mcp.CallToolResult, any, error) {
		tg, _, err := client.TargetGroupsApi.TargetgroupsFindByTargetGroupId(ctx, input.TargetGroupID).Execute()
		return tools.ToResult(tg, err)
	})
}
