package loadbalancing

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ionos-cloud/ionoscloud-mcp/pkg/api"
)

func initTargetgroups() []api.ServerTool {
	return []api.ServerTool{
		{
			Tool: api.Tool{
				Name:        "list_target_groups",
				Description: "List all target groups in your IONOS Cloud account",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {},
					"required": []
				}`),
				Annotations: api.ReadOnly("List Target Groups"),
			},
			Handler: listTargetGroups,
		},
		{
			Tool: api.Tool{
				Name:        "get_target_group",
				Description: "Get details of a specific target group",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"target_group_id": {"type": "string", "description": "The ID of the target group"}
					},
					"required": ["target_group_id"]
				}`),
				Annotations: api.ReadOnly("Get Target Group"),
			},
			Handler: getTargetGroup,
		},
	}
}

func listTargetGroups(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	targetGroups, _, err := params.Client.TargetGroupsApi.TargetgroupsGet(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to list target groups: %w", err)
	}
	return api.MarshalResult(targetGroups, "target groups")
}

func getTargetGroup(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	targetGroupID, ok := api.GetRequiredString(params.Arguments, "target_group_id")
	if !ok {
		return nil, fmt.Errorf("target_group_id is required")
	}

	targetGroup, _, err := params.Client.TargetGroupsApi.TargetgroupsFindByTargetGroupId(ctx, targetGroupID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get target group: %w", err)
	}
	return api.MarshalResult(targetGroup, "target group")
}
