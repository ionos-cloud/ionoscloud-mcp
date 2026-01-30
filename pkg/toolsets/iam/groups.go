package iam

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ionos-cloud/ionoscloud-mcp/pkg/api"
)

func initGroups() []api.ServerTool {
	return []api.ServerTool{
		{
			Tool: api.Tool{
				Name:        "list_groups",
				Description: "List all groups in your IONOS Cloud account",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {},
					"required": []
				}`),
				Annotations: api.ReadOnly("List Groups"),
			},
			Handler: listGroups,
		},
		{
			Tool: api.Tool{
				Name:        "get_group",
				Description: "Get details of a specific group",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"group_id": {"type": "string", "description": "The ID of the group"}
					},
					"required": ["group_id"]
				}`),
				Annotations: api.ReadOnly("Get Group"),
			},
			Handler: getGroup,
		},
		{
			Tool: api.Tool{
				Name:        "list_group_members",
				Description: "List all users in a group",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"group_id": {"type": "string", "description": "The ID of the group"}
					},
					"required": ["group_id"]
				}`),
				Annotations: api.ReadOnly("List Group Members"),
			},
			Handler: listGroupMembers,
		},
	}
}

func listGroups(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	groups, _, err := params.Client.UserManagementApi.UmGroupsGet(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to list groups: %w", err)
	}
	return api.MarshalResult(groups, "groups")
}

func getGroup(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	groupID, ok := api.GetRequiredString(params.Arguments, "group_id")
	if !ok {
		return nil, fmt.Errorf("group_id is required")
	}

	group, _, err := params.Client.UserManagementApi.UmGroupsFindById(ctx, groupID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get group: %w", err)
	}
	return api.MarshalResult(group, "group")
}

func listGroupMembers(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	groupID, ok := api.GetRequiredString(params.Arguments, "group_id")
	if !ok {
		return nil, fmt.Errorf("group_id is required")
	}

	members, _, err := params.Client.UserManagementApi.UmGroupsUsersGet(ctx, groupID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to list group members: %w", err)
	}
	return api.MarshalResult(members, "group members")
}
