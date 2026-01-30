package iam

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ionos-cloud/ionoscloud-mcp/pkg/api"
)

func initContract() []api.ServerTool {
	return []api.ServerTool{
		{
			Tool: api.Tool{
				Name:        "get_contract",
				Description: "Get contract information and resource limits",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {},
					"required": []
				}`),
				Annotations: api.ReadOnly("Get Contract"),
			},
			Handler: getContract,
		},
		{
			Tool: api.Tool{
				Name:        "list_resources",
				Description: "List all resources by type",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"resource_type": {"type": "string", "description": "The type of resource (optional, e.g., datacenter, image, snapshot, ipblock)"}
					},
					"required": []
				}`),
				Annotations: api.ReadOnly("List Resources"),
			},
			Handler: listResources,
		},
	}
}

func getContract(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	contract, _, err := params.Client.ContractResourcesApi.ContractsGet(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get contract: %w", err)
	}
	return api.MarshalResult(contract, "contract")
}

func listResources(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	resourceType := api.GetOptionalString(params.Arguments, "resource_type")

	if resourceType != "" {
		resources, _, err := params.Client.UserManagementApi.UmResourcesFindByType(ctx, resourceType).Execute()
		if err != nil {
			return nil, fmt.Errorf("failed to list resources: %w", err)
		}
		return api.MarshalResult(resources, "resources")
	}

	resources, _, err := params.Client.UserManagementApi.UmResourcesGet(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to list resources: %w", err)
	}
	return api.MarshalResult(resources, "resources")
}
