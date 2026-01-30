package networking

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ionos-cloud/ionoscloud-mcp/pkg/api"
	"github.com/ionos-cloud/ionoscloud-mcp/pkg/ionos"
	ionoscloud "github.com/ionos-cloud/sdk-go/v6"
)

func initIpblocks() []api.ServerTool {
	return []api.ServerTool{
		{
			Tool: api.Tool{
				Name:        "list_ipblocks",
				Description: "List all IP blocks in your IONOS Cloud account",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {},
					"required": []
				}`),
				Annotations: api.ReadOnly("List IP Blocks"),
			},
			Handler: listIpBlocks,
		},
		{
			Tool: api.Tool{
				Name:        "get_ipblock",
				Description: "Get details of a specific IP block",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"ipblock_id": {
							"type": "string",
							"description": "The ID of the IP block"
						}
					},
					"required": ["ipblock_id"]
				}`),
				Annotations: api.ReadOnly("Get IP Block"),
			},
			Handler: getIpBlock,
		},
		{
			Tool: api.Tool{
				Name:        "create_ipblock",
				Description: "Reserve a block of public IP addresses",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"location": {
							"type": "string",
							"description": "The location/region for the IP block (e.g., us/las, de/fra)"
						},
						"size": {
							"type": "integer",
							"description": "The number of IP addresses to reserve"
						},
						"name": {
							"type": "string",
							"description": "The name of the IP block"
						}
					},
					"required": ["location", "size"]
				}`),
				Annotations: api.NonIdempotent("Create IP Block"),
			},
			Handler: createIpBlock,
		},
		{
			Tool: api.Tool{
				Name:        "update_ipblock",
				Description: "Update an IP block name",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"ipblock_id": {
							"type": "string",
							"description": "The ID of the IP block to update"
						},
						"name": {
							"type": "string",
							"description": "The new name for the IP block"
						}
					},
					"required": ["ipblock_id", "name"]
				}`),
				Annotations: api.Idempotent("Update IP Block"),
			},
			Handler: updateIpBlock,
		},
		{
			Tool: api.Tool{
				Name:        "delete_ipblock",
				Description: "Release an IP block",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"ipblock_id": {
							"type": "string",
							"description": "The ID of the IP block to delete"
						}
					},
					"required": ["ipblock_id"]
				}`),
				Annotations: api.Destructive("Delete IP Block"),
			},
			Handler: deleteIpBlock,
		},
	}
}

func listIpBlocks(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	ipblocks, _, err := params.Client.IPBlocksApi.IpblocksGet(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to list IP blocks: %w", err)
	}
	return api.MarshalResult(ipblocks, "IP blocks")
}

func getIpBlock(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	ipblockID, ok := api.GetRequiredString(params.Arguments, "ipblock_id")
	if !ok {
		return nil, fmt.Errorf("ipblock_id is required")
	}

	ipblock, _, err := params.Client.IPBlocksApi.IpblocksFindById(ctx, ipblockID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get IP block: %w", err)
	}
	return api.MarshalResult(ipblock, "IP block")
}

func createIpBlock(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	location, ok := api.GetRequiredString(params.Arguments, "location")
	if !ok {
		return nil, fmt.Errorf("location is required")
	}
	size, ok := api.GetOptionalInt32(params.Arguments, "size")
	if !ok {
		return nil, fmt.Errorf("size is required")
	}

	// Validate location
	if err := ionos.ValidateLocation(location); err != nil {
		return nil, err
	}
	// Validate size
	if size < 1 {
		return nil, fmt.Errorf("size must be at least 1, got %d", size)
	}

	name := api.GetOptionalString(params.Arguments, "name")

	properties := ionoscloud.IpBlockProperties{
		Location: &location,
		Size:     &size,
	}
	if name != "" {
		properties.Name = &name
	}

	ipblock := ionoscloud.IpBlock{
		Properties: &properties,
	}

	result, _, err := params.Client.IPBlocksApi.IpblocksPost(ctx).Ipblock(ipblock).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to create IP block: %w", err)
	}
	return api.MarshalResult(result, "IP block")
}

func updateIpBlock(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	ipblockID, ok := api.GetRequiredString(params.Arguments, "ipblock_id")
	if !ok {
		return nil, fmt.Errorf("ipblock_id is required")
	}
	name, ok := api.GetRequiredString(params.Arguments, "name")
	if !ok {
		return nil, fmt.Errorf("name is required")
	}

	properties := ionoscloud.IpBlockProperties{
		Name: &name,
	}

	result, _, err := params.Client.IPBlocksApi.IpblocksPatch(ctx, ipblockID).Ipblock(properties).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to update IP block: %w", err)
	}
	return api.MarshalResult(result, "IP block")
}

func deleteIpBlock(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	ipblockID, ok := api.GetRequiredString(params.Arguments, "ipblock_id")
	if !ok {
		return nil, fmt.Errorf("ipblock_id is required")
	}

	_, err := params.Client.IPBlocksApi.IpblocksDelete(ctx, ipblockID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to delete IP block: %w", err)
	}
	return api.StatusResult(map[string]string{"status": "deleted", "ipblock_id": ipblockID})
}
