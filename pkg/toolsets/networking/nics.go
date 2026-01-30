package networking

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ionos-cloud/ionoscloud-mcp/pkg/api"
	"github.com/ionos-cloud/ionoscloud-mcp/pkg/ionos"
	ionoscloud "github.com/ionos-cloud/sdk-go/v6"
)

func initNics() []api.ServerTool {
	return []api.ServerTool{
		{
			Tool: api.Tool{
				Name:        "list_nics",
				Description: "List all NICs attached to a server",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {
							"type": "string",
							"description": "The ID of the data center"
						},
						"server_id": {
							"type": "string",
							"description": "The ID of the server"
						}
					},
					"required": ["datacenter_id", "server_id"]
				}`),
				Annotations: api.ReadOnly("List NICs"),
			},
			Handler: listNics,
		},
		{
			Tool: api.Tool{
				Name:        "get_nic",
				Description: "Get details of a specific NIC",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {
							"type": "string",
							"description": "The ID of the data center"
						},
						"server_id": {
							"type": "string",
							"description": "The ID of the server"
						},
						"nic_id": {
							"type": "string",
							"description": "The ID of the NIC"
						}
					},
					"required": ["datacenter_id", "server_id", "nic_id"]
				}`),
				Annotations: api.ReadOnly("Get NIC"),
			},
			Handler: getNic,
		},
		{
			Tool: api.Tool{
				Name:        "create_nic",
				Description: "Create a new NIC attached to a server",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {
							"type": "string",
							"description": "The ID of the data center"
						},
						"server_id": {
							"type": "string",
							"description": "The ID of the server"
						},
						"lan": {
							"type": "integer",
							"description": "The LAN ID to connect the NIC to"
						},
						"name": {
							"type": "string",
							"description": "The name of the NIC"
						},
						"dhcp": {
							"type": "boolean",
							"description": "Whether DHCP is enabled (default: true)"
						},
						"ips": {
							"type": "array",
							"items": {"type": "string"},
							"description": "List of IP addresses to assign to the NIC"
						}
					},
					"required": ["datacenter_id", "server_id", "lan"]
				}`),
				Annotations: api.NonIdempotent("Create NIC"),
			},
			Handler: createNic,
		},
		{
			Tool: api.Tool{
				Name:        "update_nic",
				Description: "Update an existing NIC",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {
							"type": "string",
							"description": "The ID of the data center"
						},
						"server_id": {
							"type": "string",
							"description": "The ID of the server"
						},
						"nic_id": {
							"type": "string",
							"description": "The ID of the NIC to update"
						},
						"name": {
							"type": "string",
							"description": "The new name for the NIC"
						},
						"lan": {
							"type": "integer",
							"description": "The new LAN ID"
						},
						"dhcp": {
							"type": "boolean",
							"description": "Whether DHCP is enabled"
						},
						"ips": {
							"type": "array",
							"items": {"type": "string"},
							"description": "New list of IP addresses"
						}
					},
					"required": ["datacenter_id", "server_id", "nic_id"]
				}`),
				Annotations: api.Idempotent("Update NIC"),
			},
			Handler: updateNic,
		},
		{
			Tool: api.Tool{
				Name:        "delete_nic",
				Description: "Delete a NIC from a server",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {
							"type": "string",
							"description": "The ID of the data center"
						},
						"server_id": {
							"type": "string",
							"description": "The ID of the server"
						},
						"nic_id": {
							"type": "string",
							"description": "The ID of the NIC to delete"
						}
					},
					"required": ["datacenter_id", "server_id", "nic_id"]
				}`),
				Annotations: api.Destructive("Delete NIC"),
			},
			Handler: deleteNic,
		},
	}
}

func listNics(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	serverID, ok := api.GetRequiredString(params.Arguments, "server_id")
	if !ok {
		return nil, fmt.Errorf("server_id is required")
	}

	nics, _, err := params.Client.NetworkInterfacesApi.DatacentersServersNicsGet(ctx, datacenterID, serverID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to list NICs: %w", err)
	}
	return api.MarshalResult(nics, "NICs")
}

func getNic(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	serverID, ok := api.GetRequiredString(params.Arguments, "server_id")
	if !ok {
		return nil, fmt.Errorf("server_id is required")
	}
	nicID, ok := api.GetRequiredString(params.Arguments, "nic_id")
	if !ok {
		return nil, fmt.Errorf("nic_id is required")
	}

	nic, _, err := params.Client.NetworkInterfacesApi.DatacentersServersNicsFindById(ctx, datacenterID, serverID, nicID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get NIC: %w", err)
	}
	return api.MarshalResult(nic, "NIC")
}

func createNic(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	serverID, ok := api.GetRequiredString(params.Arguments, "server_id")
	if !ok {
		return nil, fmt.Errorf("server_id is required")
	}
	lan, ok := api.GetOptionalInt32(params.Arguments, "lan")
	if !ok {
		return nil, fmt.Errorf("lan is required")
	}

	// Validate LAN ID
	if lan < 1 {
		return nil, fmt.Errorf("lan must be at least 1, got %d", lan)
	}

	name := api.GetOptionalString(params.Arguments, "name")
	dhcp, _ := api.GetOptionalBool(params.Arguments, "dhcp")
	ips := api.GetStringSlice(params.Arguments, "ips")

	// Validate IP addresses
	for _, ip := range ips {
		if err := ionos.ValidateIP(ip); err != nil {
			return nil, fmt.Errorf("invalid IP in ips list: %w", err)
		}
	}

	properties := ionoscloud.NicProperties{
		Lan:  &lan,
		Dhcp: &dhcp,
	}
	if name != "" {
		properties.Name = &name
	}
	if len(ips) > 0 {
		properties.Ips = &ips
	}

	nic := ionoscloud.Nic{
		Properties: &properties,
	}

	result, _, err := params.Client.NetworkInterfacesApi.DatacentersServersNicsPost(ctx, datacenterID, serverID).Nic(nic).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to create NIC: %w", err)
	}
	return api.MarshalResult(result, "NIC")
}

func updateNic(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	serverID, ok := api.GetRequiredString(params.Arguments, "server_id")
	if !ok {
		return nil, fmt.Errorf("server_id is required")
	}
	nicID, ok := api.GetRequiredString(params.Arguments, "nic_id")
	if !ok {
		return nil, fmt.Errorf("nic_id is required")
	}

	name := api.GetOptionalString(params.Arguments, "name")
	lan, lanSet := api.GetOptionalInt32(params.Arguments, "lan")
	dhcp, dhcpSet := api.GetOptionalBool(params.Arguments, "dhcp")
	ips := api.GetStringSlice(params.Arguments, "ips")

	if name == "" && !lanSet && !dhcpSet && len(ips) == 0 {
		return nil, fmt.Errorf("at least one of name, lan, dhcp, or ips must be provided")
	}

	// Validate IP addresses
	for _, ip := range ips {
		if err := ionos.ValidateIP(ip); err != nil {
			return nil, fmt.Errorf("invalid IP in ips list: %w", err)
		}
	}

	properties := ionoscloud.NicProperties{}
	if name != "" {
		properties.Name = &name
	}
	if lanSet && lan > 0 {
		properties.Lan = &lan
	}
	if dhcpSet {
		properties.Dhcp = &dhcp
	}
	if len(ips) > 0 {
		properties.Ips = &ips
	}

	result, _, err := params.Client.NetworkInterfacesApi.DatacentersServersNicsPatch(ctx, datacenterID, serverID, nicID).Nic(properties).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to update NIC: %w", err)
	}
	return api.MarshalResult(result, "NIC")
}

func deleteNic(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	serverID, ok := api.GetRequiredString(params.Arguments, "server_id")
	if !ok {
		return nil, fmt.Errorf("server_id is required")
	}
	nicID, ok := api.GetRequiredString(params.Arguments, "nic_id")
	if !ok {
		return nil, fmt.Errorf("nic_id is required")
	}

	_, err := params.Client.NetworkInterfacesApi.DatacentersServersNicsDelete(ctx, datacenterID, serverID, nicID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to delete NIC: %w", err)
	}
	return api.StatusResult(map[string]string{"status": "deleted", "nic_id": nicID})
}
