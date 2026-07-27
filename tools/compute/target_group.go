package compute

import (
	"context"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterTargetGroupTools(server *mcp.Server, client *ionos.APIClient, scope tools.Scope) {
	tools.RegisterTool(server, scope, tools.MethodGet, &mcp.Tool{
		Name:        "list_target_groups",
		Description: "List all target groups in your IONOS CLOUD account",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ListTargetGroupsInput) (*mcp.CallToolResult, any, error) {
		depth := int32(1)
		if input.Depth != nil {
			depth = *input.Depth
		}
		r := client.TargetGroupsApi.TargetgroupsGet(ctx).Depth(depth)
		for k, v := range input.Filters {
			r = r.Filter(k, v)
		}
		tgs, _, err := r.Execute()
		return tools.ToResult(tgs, err)
	})

	tools.RegisterTool(server, scope, tools.MethodGet, &mcp.Tool{
		Name:        "get_target_group",
		Description: "Get details of a specific target group",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.TargetGroupIDInput) (*mcp.CallToolResult, any, error) {
		apiReq := client.TargetGroupsApi.TargetgroupsFindByTargetGroupId(ctx, input.TargetGroupID)
		if input.Depth != nil {
			apiReq = apiReq.Depth(*input.Depth)
		}
		tg, _, err := apiReq.Execute()
		return tools.ToResult(tg, err)
	})
}
