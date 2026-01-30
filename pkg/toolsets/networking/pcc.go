package networking

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ionos-cloud/ionoscloud-mcp/pkg/api"
	ionoscloud "github.com/ionos-cloud/sdk-go/v6"
)

func initPcc() []api.ServerTool {
	return []api.ServerTool{
		{
			Tool: api.Tool{
				Name:        "list_pccs",
				Description: "List all Private Cross Connects in your IONOS Cloud account",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {},
					"required": []
				}`),
				Annotations: api.ReadOnly("List PCCs"),
			},
			Handler: listPccs,
		},
		{
			Tool: api.Tool{
				Name:        "get_pcc",
				Description: "Get details of a specific Private Cross Connect",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"pcc_id": {
							"type": "string",
							"description": "The ID of the Private Cross Connect"
						}
					},
					"required": ["pcc_id"]
				}`),
				Annotations: api.ReadOnly("Get PCC"),
			},
			Handler: getPcc,
		},
		{
			Tool: api.Tool{
				Name:        "create_pcc",
				Description: "Create a Private Cross Connect",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"name": {
							"type": "string",
							"description": "The name of the Private Cross Connect"
						},
						"description": {
							"type": "string",
							"description": "A description for the Private Cross Connect"
						}
					},
					"required": ["name"]
				}`),
				Annotations: api.NonIdempotent("Create PCC"),
			},
			Handler: createPcc,
		},
		{
			Tool: api.Tool{
				Name:        "update_pcc",
				Description: "Update a Private Cross Connect",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"pcc_id": {
							"type": "string",
							"description": "The ID of the Private Cross Connect"
						},
						"name": {
							"type": "string",
							"description": "The new name for the Private Cross Connect"
						},
						"description": {
							"type": "string",
							"description": "The new description for the Private Cross Connect"
						}
					},
					"required": ["pcc_id"]
				}`),
				Annotations: api.Idempotent("Update PCC"),
			},
			Handler: updatePcc,
		},
		{
			Tool: api.Tool{
				Name:        "delete_pcc",
				Description: "Delete a Private Cross Connect",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"pcc_id": {
							"type": "string",
							"description": "The ID of the Private Cross Connect to delete"
						}
					},
					"required": ["pcc_id"]
				}`),
				Annotations: api.Destructive("Delete PCC"),
			},
			Handler: deletePcc,
		},
	}
}

func listPccs(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	pccs, _, err := params.Client.PrivateCrossConnectsApi.PccsGet(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to list Private Cross Connects: %w", err)
	}
	return api.MarshalResult(pccs, "Private Cross Connects")
}

func getPcc(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	pccID, ok := api.GetRequiredString(params.Arguments, "pcc_id")
	if !ok {
		return nil, fmt.Errorf("pcc_id is required")
	}

	pcc, _, err := params.Client.PrivateCrossConnectsApi.PccsFindById(ctx, pccID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get Private Cross Connect: %w", err)
	}
	return api.MarshalResult(pcc, "Private Cross Connect")
}

func createPcc(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	name, ok := api.GetRequiredString(params.Arguments, "name")
	if !ok {
		return nil, fmt.Errorf("name is required")
	}

	description := api.GetOptionalString(params.Arguments, "description")

	properties := ionoscloud.PrivateCrossConnectProperties{
		Name: &name,
	}
	if description != "" {
		properties.Description = &description
	}

	pcc := ionoscloud.PrivateCrossConnect{
		Properties: &properties,
	}

	result, _, err := params.Client.PrivateCrossConnectsApi.PccsPost(ctx).Pcc(pcc).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to create Private Cross Connect: %w", err)
	}
	return api.MarshalResult(result, "Private Cross Connect")
}

func updatePcc(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	pccID, ok := api.GetRequiredString(params.Arguments, "pcc_id")
	if !ok {
		return nil, fmt.Errorf("pcc_id is required")
	}

	name := api.GetOptionalString(params.Arguments, "name")
	description := api.GetOptionalString(params.Arguments, "description")

	// Check if at least one field is provided
	if name == "" && description == "" {
		return nil, fmt.Errorf("at least one field must be provided for update")
	}

	properties := ionoscloud.PrivateCrossConnectProperties{}
	if name != "" {
		properties.Name = &name
	}
	if description != "" {
		properties.Description = &description
	}

	result, _, err := params.Client.PrivateCrossConnectsApi.PccsPatch(ctx, pccID).Pcc(properties).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to update Private Cross Connect: %w", err)
	}
	return api.MarshalResult(result, "Private Cross Connect")
}

func deletePcc(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	pccID, ok := api.GetRequiredString(params.Arguments, "pcc_id")
	if !ok {
		return nil, fmt.Errorf("pcc_id is required")
	}

	_, err := params.Client.PrivateCrossConnectsApi.PccsDelete(ctx, pccID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to delete Private Cross Connect: %w", err)
	}
	return api.StatusResult(map[string]string{"status": "deleted", "pcc_id": pccID})
}
