package compute

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ionos-cloud/ionoscloud-mcp/pkg/api"
	ionoscloud "github.com/ionos-cloud/sdk-go/v6"
)

func initDatacenters() []api.ServerTool {
	return []api.ServerTool{
		{
			Tool: api.Tool{
				Name:        "list_datacenters",
				Description: "List all virtual data centers in your IONOS Cloud account",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {},
					"required": []
				}`),
				Annotations: api.ReadOnly("List Datacenters"),
			},
			Handler: listDatacenters,
		},
		{
			Tool: api.Tool{
				Name:        "get_datacenter",
				Description: "Get details of a specific virtual data center",
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
				Annotations: api.ReadOnly("Get Datacenter"),
			},
			Handler: getDatacenter,
		},
		{
			Tool: api.Tool{
				Name:        "create_datacenter",
				Description: "Create a new virtual data center",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"name": {
							"type": "string",
							"description": "The name of the data center"
						},
						"location": {
							"type": "string",
							"description": "The location/region for the data center (e.g., us/las, de/fra, de/txl)"
						},
						"description": {
							"type": "string",
							"description": "A description for the data center"
						}
					},
					"required": ["name", "location"]
				}`),
				Annotations: api.NonIdempotent("Create Datacenter"),
			},
			Handler: createDatacenter,
		},
		{
			Tool: api.Tool{
				Name:        "update_datacenter",
				Description: "Update an existing virtual data center",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {
							"type": "string",
							"description": "The ID of the data center to update"
						},
						"name": {
							"type": "string",
							"description": "The new name for the data center"
						},
						"description": {
							"type": "string",
							"description": "The new description for the data center"
						}
					},
					"required": ["datacenter_id"]
				}`),
				Annotations: api.Idempotent("Update Datacenter"),
			},
			Handler: updateDatacenter,
		},
		{
			Tool: api.Tool{
				Name:        "delete_datacenter",
				Description: "Delete a virtual data center",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {
							"type": "string",
							"description": "The ID of the data center to delete"
						}
					},
					"required": ["datacenter_id"]
				}`),
				Annotations: api.Destructive("Delete Datacenter"),
			},
			Handler: deleteDatacenter,
		},
	}
}

func listDatacenters(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenters, _, err := params.Client.DataCentersApi.DatacentersGet(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to list datacenters: %w", err)
	}
	return api.MarshalResult(datacenters, "datacenters")
}

func getDatacenter(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}

	datacenter, _, err := params.Client.DataCentersApi.DatacentersFindById(ctx, datacenterID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get datacenter: %w", err)
	}
	return api.MarshalResult(datacenter, "datacenter")
}

func createDatacenter(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	name, ok := api.GetRequiredString(params.Arguments, "name")
	if !ok {
		return nil, fmt.Errorf("name is required")
	}
	location, ok := api.GetRequiredString(params.Arguments, "location")
	if !ok {
		return nil, fmt.Errorf("location is required")
	}
	description := api.GetOptionalString(params.Arguments, "description")

	properties := ionoscloud.DatacenterPropertiesPost{
		Name:     &name,
		Location: &location,
	}
	if description != "" {
		properties.Description = &description
	}

	datacenter := ionoscloud.DatacenterPost{
		Properties: &properties,
	}

	result, _, err := params.Client.DataCentersApi.DatacentersPost(ctx).Datacenter(datacenter).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to create datacenter: %w", err)
	}
	return api.MarshalResult(result, "datacenter")
}

func updateDatacenter(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	name := api.GetOptionalString(params.Arguments, "name")
	description := api.GetOptionalString(params.Arguments, "description")

	if name == "" && description == "" {
		return nil, fmt.Errorf("at least one of name or description must be provided")
	}

	properties := ionoscloud.DatacenterPropertiesPut{}
	if name != "" {
		properties.Name = &name
	}
	if description != "" {
		properties.Description = &description
	}

	result, _, err := params.Client.DataCentersApi.DatacentersPatch(ctx, datacenterID).Datacenter(properties).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to update datacenter: %w", err)
	}
	return api.MarshalResult(result, "datacenter")
}

func deleteDatacenter(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}

	_, err := params.Client.DataCentersApi.DatacentersDelete(ctx, datacenterID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to delete datacenter: %w", err)
	}
	return api.StatusResult(map[string]string{"status": "deleted", "datacenter_id": datacenterID})
}
