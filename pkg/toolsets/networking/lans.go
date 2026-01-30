package networking

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ionos-cloud/ionoscloud-mcp/pkg/api"
	ionoscloud "github.com/ionos-cloud/sdk-go/v6"
)

func initLans() []api.ServerTool {
	return []api.ServerTool{
		{
			Tool: api.Tool{
				Name:        "list_lans",
				Description: "List all LANs in a data center",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {
							"type": "string",
							"description": "The ID of the data center"
						}
					},
					"required": ["datacenter_id"]
				}`),
				Annotations: api.ReadOnly("List LANs"),
			},
			Handler: listLans,
		},
		{
			Tool: api.Tool{
				Name:        "get_lan",
				Description: "Get details of a specific LAN",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {
							"type": "string",
							"description": "The ID of the data center"
						},
						"lan_id": {
							"type": "string",
							"description": "The ID of the LAN"
						}
					},
					"required": ["datacenter_id", "lan_id"]
				}`),
				Annotations: api.ReadOnly("Get LAN"),
			},
			Handler: getLan,
		},
		{
			Tool: api.Tool{
				Name:        "create_lan",
				Description: "Create a new LAN in a data center",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {
							"type": "string",
							"description": "The ID of the data center"
						},
						"name": {
							"type": "string",
							"description": "The name of the LAN"
						},
						"public": {
							"type": "boolean",
							"description": "Whether the LAN is public (has internet access)"
						}
					},
					"required": ["datacenter_id"]
				}`),
				Annotations: api.NonIdempotent("Create LAN"),
			},
			Handler: createLan,
		},
		{
			Tool: api.Tool{
				Name:        "update_lan",
				Description: "Update an existing LAN",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {
							"type": "string",
							"description": "The ID of the data center"
						},
						"lan_id": {
							"type": "string",
							"description": "The ID of the LAN to update"
						},
						"name": {
							"type": "string",
							"description": "The new name for the LAN"
						},
						"public": {
							"type": "boolean",
							"description": "Whether the LAN is public"
						}
					},
					"required": ["datacenter_id", "lan_id"]
				}`),
				Annotations: api.Idempotent("Update LAN"),
			},
			Handler: updateLan,
		},
		{
			Tool: api.Tool{
				Name:        "delete_lan",
				Description: "Delete a LAN",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {
							"type": "string",
							"description": "The ID of the data center"
						},
						"lan_id": {
							"type": "string",
							"description": "The ID of the LAN to delete"
						}
					},
					"required": ["datacenter_id", "lan_id"]
				}`),
				Annotations: api.Destructive("Delete LAN"),
			},
			Handler: deleteLan,
		},
	}
}

func listLans(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}

	lans, _, err := params.Client.LANsApi.DatacentersLansGet(ctx, datacenterID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to list LANs: %w", err)
	}
	return api.MarshalResult(lans, "LANs")
}

func getLan(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	lanID, ok := api.GetRequiredString(params.Arguments, "lan_id")
	if !ok {
		return nil, fmt.Errorf("lan_id is required")
	}

	lan, _, err := params.Client.LANsApi.DatacentersLansFindById(ctx, datacenterID, lanID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get LAN: %w", err)
	}
	return api.MarshalResult(lan, "LAN")
}

func createLan(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}

	name := api.GetOptionalString(params.Arguments, "name")
	public, _ := api.GetOptionalBool(params.Arguments, "public")

	properties := ionoscloud.LanProperties{
		Public: &public,
	}
	if name != "" {
		properties.Name = &name
	}

	lan := ionoscloud.Lan{
		Properties: &properties,
	}

	result, _, err := params.Client.LANsApi.DatacentersLansPost(ctx, datacenterID).Lan(lan).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to create LAN: %w", err)
	}
	return api.MarshalResult(result, "LAN")
}

func updateLan(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	lanID, ok := api.GetRequiredString(params.Arguments, "lan_id")
	if !ok {
		return nil, fmt.Errorf("lan_id is required")
	}

	name := api.GetOptionalString(params.Arguments, "name")
	public, publicSet := api.GetOptionalBool(params.Arguments, "public")

	if name == "" && !publicSet {
		return nil, fmt.Errorf("at least one of name or public must be provided")
	}

	properties := ionoscloud.LanProperties{}
	if name != "" {
		properties.Name = &name
	}
	if publicSet {
		properties.Public = &public
	}

	result, _, err := params.Client.LANsApi.DatacentersLansPatch(ctx, datacenterID, lanID).Lan(properties).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to update LAN: %w", err)
	}
	return api.MarshalResult(result, "LAN")
}

func deleteLan(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	lanID, ok := api.GetRequiredString(params.Arguments, "lan_id")
	if !ok {
		return nil, fmt.Errorf("lan_id is required")
	}

	_, err := params.Client.LANsApi.DatacentersLansDelete(ctx, datacenterID, lanID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to delete LAN: %w", err)
	}
	return api.StatusResult(map[string]string{"status": "deleted", "lan_id": lanID})
}
