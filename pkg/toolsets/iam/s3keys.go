package iam

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ionos-cloud/ionoscloud-mcp/pkg/api"
)

func initS3Keys() []api.ServerTool {
	return []api.ServerTool{
		{
			Tool: api.Tool{
				Name:        "list_s3_keys",
				Description: "List all S3 keys for a user",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"user_id": {"type": "string", "description": "The ID of the user"}
					},
					"required": ["user_id"]
				}`),
				Annotations: api.ReadOnly("List S3 Keys"),
			},
			Handler: listS3Keys,
		},
		{
			Tool: api.Tool{
				Name:        "get_s3_key",
				Description: "Get details of a specific S3 key",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"user_id": {"type": "string", "description": "The ID of the user"},
						"key_id": {"type": "string", "description": "The ID of the S3 key"}
					},
					"required": ["user_id", "key_id"]
				}`),
				Annotations: api.ReadOnly("Get S3 Key"),
			},
			Handler: getS3Key,
		},
	}
}

func listS3Keys(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	userID, ok := api.GetRequiredString(params.Arguments, "user_id")
	if !ok {
		return nil, fmt.Errorf("user_id is required")
	}

	keys, _, err := params.Client.UserS3KeysApi.UmUsersS3keysGet(ctx, userID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to list S3 keys: %w", err)
	}
	return api.MarshalResult(keys, "S3 keys")
}

func getS3Key(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	userID, ok := api.GetRequiredString(params.Arguments, "user_id")
	if !ok {
		return nil, fmt.Errorf("user_id is required")
	}
	keyID, ok := api.GetRequiredString(params.Arguments, "key_id")
	if !ok {
		return nil, fmt.Errorf("key_id is required")
	}

	key, _, err := params.Client.UserS3KeysApi.UmUsersS3keysFindByKeyId(ctx, userID, keyID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get S3 key: %w", err)
	}
	return api.MarshalResult(key, "S3 key")
}
