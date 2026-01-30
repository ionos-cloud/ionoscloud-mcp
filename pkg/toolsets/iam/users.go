package iam

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ionos-cloud/ionoscloud-mcp/pkg/api"
)

func initUsers() []api.ServerTool {
	return []api.ServerTool{
		{
			Tool: api.Tool{
				Name:        "list_users",
				Description: "List all users in your IONOS Cloud account",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {},
					"required": []
				}`),
				Annotations: api.ReadOnly("List Users"),
			},
			Handler: listUsers,
		},
		{
			Tool: api.Tool{
				Name:        "get_user",
				Description: "Get details of a specific user",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"user_id": {"type": "string", "description": "The ID of the user"}
					},
					"required": ["user_id"]
				}`),
				Annotations: api.ReadOnly("Get User"),
			},
			Handler: getUser,
		},
		{
			Tool: api.Tool{
				Name:        "list_user_groups",
				Description: "List all groups a user belongs to",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"user_id": {"type": "string", "description": "The ID of the user"}
					},
					"required": ["user_id"]
				}`),
				Annotations: api.ReadOnly("List User Groups"),
			},
			Handler: listUserGroups,
		},
	}
}

func listUsers(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	users, _, err := params.Client.UserManagementApi.UmUsersGet(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return api.MarshalResult(users, "users")
}

func getUser(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	userID, ok := api.GetRequiredString(params.Arguments, "user_id")
	if !ok {
		return nil, fmt.Errorf("user_id is required")
	}

	user, _, err := params.Client.UserManagementApi.UmUsersFindById(ctx, userID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return api.MarshalResult(user, "user")
}

func listUserGroups(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	userID, ok := api.GetRequiredString(params.Arguments, "user_id")
	if !ok {
		return nil, fmt.Errorf("user_id is required")
	}

	groups, _, err := params.Client.UserManagementApi.UmUsersGroupsGet(ctx, userID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to list user groups: %w", err)
	}
	return api.MarshalResult(groups, "user groups")
}
