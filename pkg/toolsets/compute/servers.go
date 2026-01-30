package compute

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ionos-cloud/ionoscloud-mcp/pkg/api"
	ionoscloud "github.com/ionos-cloud/sdk-go/v6"
)

func initServers() []api.ServerTool {
	return []api.ServerTool{
		{
			Tool: api.Tool{
				Name:        "list_servers",
				Description: "List all servers in a data center",
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
				Annotations: api.ReadOnly("List Servers"),
			},
			Handler: listServers,
		},
		{
			Tool: api.Tool{
				Name:        "get_server",
				Description: "Get details of a specific server",
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
				Annotations: api.ReadOnly("Get Server"),
			},
			Handler: getServer,
		},
		{
			Tool: api.Tool{
				Name:        "create_server",
				Description: "Create a new server in a data center",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {
							"type": "string",
							"description": "The ID of the data center"
						},
						"name": {
							"type": "string",
							"description": "The name of the server"
						},
						"cores": {
							"type": "integer",
							"description": "The number of CPU cores"
						},
						"ram": {
							"type": "integer",
							"description": "The amount of RAM in MB"
						},
						"cpu_family": {
							"type": "string",
							"description": "The CPU family (e.g., INTEL_SKYLAKE, AMD_OPTERON)"
						},
						"availability_zone": {
							"type": "string",
							"description": "The availability zone (AUTO, ZONE_1, ZONE_2)"
						}
					},
					"required": ["datacenter_id", "name", "cores", "ram"]
				}`),
				Annotations: api.NonIdempotent("Create Server"),
			},
			Handler: createServer,
		},
		{
			Tool: api.Tool{
				Name:        "update_server",
				Description: "Update an existing server",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {
							"type": "string",
							"description": "The ID of the data center"
						},
						"server_id": {
							"type": "string",
							"description": "The ID of the server to update"
						},
						"name": {
							"type": "string",
							"description": "The new name for the server"
						},
						"cores": {
							"type": "integer",
							"description": "The new number of CPU cores"
						},
						"ram": {
							"type": "integer",
							"description": "The new amount of RAM in MB"
						}
					},
					"required": ["datacenter_id", "server_id"]
				}`),
				Annotations: api.Idempotent("Update Server"),
			},
			Handler: updateServer,
		},
		{
			Tool: api.Tool{
				Name:        "delete_server",
				Description: "Delete a server",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {
							"type": "string",
							"description": "The ID of the data center"
						},
						"server_id": {
							"type": "string",
							"description": "The ID of the server to delete"
						}
					},
					"required": ["datacenter_id", "server_id"]
				}`),
				Annotations: api.Destructive("Delete Server"),
			},
			Handler: deleteServer,
		},
		{
			Tool: api.Tool{
				Name:        "start_server",
				Description: "Start (power on) a server",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {
							"type": "string",
							"description": "The ID of the data center"
						},
						"server_id": {
							"type": "string",
							"description": "The ID of the server to start"
						}
					},
					"required": ["datacenter_id", "server_id"]
				}`),
				Annotations: api.Idempotent("Start Server"),
			},
			Handler: startServer,
		},
		{
			Tool: api.Tool{
				Name:        "stop_server",
				Description: "Stop (power off) a server",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {
							"type": "string",
							"description": "The ID of the data center"
						},
						"server_id": {
							"type": "string",
							"description": "The ID of the server to stop"
						}
					},
					"required": ["datacenter_id", "server_id"]
				}`),
				Annotations: api.Destructive("Stop Server"),
			},
			Handler: stopServer,
		},
		{
			Tool: api.Tool{
				Name:        "reboot_server",
				Description: "Reboot a server",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {
							"type": "string",
							"description": "The ID of the data center"
						},
						"server_id": {
							"type": "string",
							"description": "The ID of the server to reboot"
						}
					},
					"required": ["datacenter_id", "server_id"]
				}`),
				Annotations: api.Destructive("Reboot Server"),
			},
			Handler: rebootServer,
		},
	}
}

func listServers(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}

	servers, _, err := params.Client.ServersApi.DatacentersServersGet(ctx, datacenterID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to list servers: %w", err)
	}
	return api.MarshalResult(servers, "servers")
}

func getServer(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	serverID, ok := api.GetRequiredString(params.Arguments, "server_id")
	if !ok {
		return nil, fmt.Errorf("server_id is required")
	}

	server, _, err := params.Client.ServersApi.DatacentersServersFindById(ctx, datacenterID, serverID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get server: %w", err)
	}
	return api.MarshalResult(server, "server")
}

func createServer(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	name, ok := api.GetRequiredString(params.Arguments, "name")
	if !ok {
		return nil, fmt.Errorf("name is required")
	}
	coresFloat, ok := api.GetOptionalFloat(params.Arguments, "cores")
	if !ok {
		return nil, fmt.Errorf("cores is required")
	}
	ramFloat, ok := api.GetOptionalFloat(params.Arguments, "ram")
	if !ok {
		return nil, fmt.Errorf("ram is required")
	}

	cores := int32(coresFloat)
	ram := int32(ramFloat)

	// Validate inputs
	if cores < 1 {
		return nil, fmt.Errorf("cores must be at least 1, got %d", cores)
	}
	if ram < 256 {
		return nil, fmt.Errorf("ram must be at least 256 MB, got %d", ram)
	}
	if ram%256 != 0 {
		return nil, fmt.Errorf("ram must be a multiple of 256 MB, got %d", ram)
	}

	cpuFamily := api.GetOptionalString(params.Arguments, "cpu_family")
	availabilityZone := api.GetOptionalString(params.Arguments, "availability_zone")

	properties := ionoscloud.ServerProperties{
		Name:  &name,
		Cores: &cores,
		Ram:   &ram,
	}
	if cpuFamily != "" {
		properties.CpuFamily = &cpuFamily
	}
	if availabilityZone != "" {
		properties.AvailabilityZone = &availabilityZone
	}

	server := ionoscloud.Server{
		Properties: &properties,
	}

	result, _, err := params.Client.ServersApi.DatacentersServersPost(ctx, datacenterID).Server(server).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to create server: %w", err)
	}
	return api.MarshalResult(result, "server")
}

func updateServer(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	serverID, ok := api.GetRequiredString(params.Arguments, "server_id")
	if !ok {
		return nil, fmt.Errorf("server_id is required")
	}

	name := api.GetOptionalString(params.Arguments, "name")
	cores, coresSet := api.GetOptionalInt32(params.Arguments, "cores")
	ram, ramSet := api.GetOptionalInt32(params.Arguments, "ram")

	if name == "" && !coresSet && !ramSet {
		return nil, fmt.Errorf("at least one of name, cores, or ram must be provided")
	}

	properties := ionoscloud.ServerProperties{}
	if name != "" {
		properties.Name = &name
	}
	if coresSet && cores > 0 {
		properties.Cores = &cores
	}
	if ramSet && ram > 0 {
		properties.Ram = &ram
	}

	result, _, err := params.Client.ServersApi.DatacentersServersPatch(ctx, datacenterID, serverID).Server(properties).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to update server: %w", err)
	}
	return api.MarshalResult(result, "server")
}

func deleteServer(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	serverID, ok := api.GetRequiredString(params.Arguments, "server_id")
	if !ok {
		return nil, fmt.Errorf("server_id is required")
	}

	_, err := params.Client.ServersApi.DatacentersServersDelete(ctx, datacenterID, serverID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to delete server: %w", err)
	}
	return api.StatusResult(map[string]string{"status": "deleted", "server_id": serverID})
}

func startServer(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	serverID, ok := api.GetRequiredString(params.Arguments, "server_id")
	if !ok {
		return nil, fmt.Errorf("server_id is required")
	}

	_, err := params.Client.ServersApi.DatacentersServersStartPost(ctx, datacenterID, serverID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to start server: %w", err)
	}
	return api.StatusResult(map[string]string{"status": "starting", "server_id": serverID})
}

func stopServer(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	serverID, ok := api.GetRequiredString(params.Arguments, "server_id")
	if !ok {
		return nil, fmt.Errorf("server_id is required")
	}

	_, err := params.Client.ServersApi.DatacentersServersStopPost(ctx, datacenterID, serverID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to stop server: %w", err)
	}
	return api.StatusResult(map[string]string{"status": "stopping", "server_id": serverID})
}

func rebootServer(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	serverID, ok := api.GetRequiredString(params.Arguments, "server_id")
	if !ok {
		return nil, fmt.Errorf("server_id is required")
	}

	_, err := params.Client.ServersApi.DatacentersServersRebootPost(ctx, datacenterID, serverID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to reboot server: %w", err)
	}
	return api.StatusResult(map[string]string{"status": "rebooting", "server_id": serverID})
}
