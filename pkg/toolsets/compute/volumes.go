package compute

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ionos-cloud/ionoscloud-mcp/pkg/api"
	"github.com/ionos-cloud/ionoscloud-mcp/pkg/ionos"
	ionoscloud "github.com/ionos-cloud/sdk-go/v6"
)

func initVolumes() []api.ServerTool {
	return []api.ServerTool{
		{
			Tool: api.Tool{
				Name:        "list_volumes",
				Description: "List all volumes in a data center",
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
				Annotations: api.ReadOnly("List Volumes"),
			},
			Handler: listVolumes,
		},
		{
			Tool: api.Tool{
				Name:        "get_volume",
				Description: "Get details of a specific volume",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {
							"type": "string",
							"description": "The ID of the data center"
						},
						"volume_id": {
							"type": "string",
							"description": "The ID of the volume"
						}
					},
					"required": ["datacenter_id", "volume_id"]
				}`),
				Annotations: api.ReadOnly("Get Volume"),
			},
			Handler: getVolume,
		},
		{
			Tool: api.Tool{
				Name:        "create_volume",
				Description: "Create a new volume in a data center",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {
							"type": "string",
							"description": "The ID of the data center"
						},
						"name": {
							"type": "string",
							"description": "The name of the volume"
						},
						"size": {
							"type": "integer",
							"description": "The size of the volume in GB"
						},
						"type": {
							"type": "string",
							"description": "The volume type (HDD, SSD, SSD_STANDARD, SSD_PREMIUM)"
						},
						"bus": {
							"type": "string",
							"description": "The bus type (VIRTIO, IDE)"
						},
						"availability_zone": {
							"type": "string",
							"description": "The availability zone (AUTO, ZONE_1, ZONE_2, ZONE_3)"
						},
						"image": {
							"type": "string",
							"description": "The image or snapshot ID to use"
						},
						"image_password": {
							"type": "string",
							"description": "Password for the image (required for some images)"
						},
						"licence_type": {
							"type": "string",
							"description": "The licence type (LINUX, WINDOWS, WINDOWS2016, OTHER, UNKNOWN)"
						}
					},
					"required": ["datacenter_id", "name", "size"]
				}`),
				Annotations: api.NonIdempotent("Create Volume"),
			},
			Handler: createVolume,
		},
		{
			Tool: api.Tool{
				Name:        "update_volume",
				Description: "Update an existing volume",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {
							"type": "string",
							"description": "The ID of the data center"
						},
						"volume_id": {
							"type": "string",
							"description": "The ID of the volume to update"
						},
						"name": {
							"type": "string",
							"description": "The new name for the volume"
						},
						"size": {
							"type": "integer",
							"description": "The new size in GB (can only increase)"
						}
					},
					"required": ["datacenter_id", "volume_id"]
				}`),
				Annotations: api.Idempotent("Update Volume"),
			},
			Handler: updateVolume,
		},
		{
			Tool: api.Tool{
				Name:        "delete_volume",
				Description: "Delete a volume",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {
							"type": "string",
							"description": "The ID of the data center"
						},
						"volume_id": {
							"type": "string",
							"description": "The ID of the volume to delete"
						}
					},
					"required": ["datacenter_id", "volume_id"]
				}`),
				Annotations: api.Destructive("Delete Volume"),
			},
			Handler: deleteVolume,
		},
		{
			Tool: api.Tool{
				Name:        "attach_volume",
				Description: "Attach a volume to a server",
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
						"volume_id": {
							"type": "string",
							"description": "The ID of the volume to attach"
						}
					},
					"required": ["datacenter_id", "server_id", "volume_id"]
				}`),
				Annotations: api.Idempotent("Attach Volume"),
			},
			Handler: attachVolume,
		},
		{
			Tool: api.Tool{
				Name:        "detach_volume",
				Description: "Detach a volume from a server",
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
						"volume_id": {
							"type": "string",
							"description": "The ID of the volume to detach"
						}
					},
					"required": ["datacenter_id", "server_id", "volume_id"]
				}`),
				Annotations: api.Idempotent("Detach Volume"),
			},
			Handler: detachVolume,
		},
	}
}

func listVolumes(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}

	volumes, _, err := params.Client.VolumesApi.DatacentersVolumesGet(ctx, datacenterID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to list volumes: %w", err)
	}
	return api.MarshalResult(volumes, "volumes")
}

func getVolume(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	volumeID, ok := api.GetRequiredString(params.Arguments, "volume_id")
	if !ok {
		return nil, fmt.Errorf("volume_id is required")
	}

	volume, _, err := params.Client.VolumesApi.DatacentersVolumesFindById(ctx, datacenterID, volumeID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get volume: %w", err)
	}
	return api.MarshalResult(volume, "volume")
}

func createVolume(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	name, ok := api.GetRequiredString(params.Arguments, "name")
	if !ok {
		return nil, fmt.Errorf("name is required")
	}
	sizeFloat, ok := api.GetOptionalFloat(params.Arguments, "size")
	if !ok {
		return nil, fmt.Errorf("size is required")
	}

	size := float32(sizeFloat)
	if size < 1 {
		return nil, fmt.Errorf("size must be at least 1 GB, got %.1f", size)
	}

	volumeType := api.GetOptionalString(params.Arguments, "type")
	bus := api.GetOptionalString(params.Arguments, "bus")
	availabilityZone := api.GetOptionalString(params.Arguments, "availability_zone")
	image := api.GetOptionalString(params.Arguments, "image")
	imagePassword := api.GetOptionalString(params.Arguments, "image_password")
	licenceType := api.GetOptionalString(params.Arguments, "licence_type")

	// Validate volume type if provided
	if err := ionos.ValidateVolumeType(volumeType); err != nil {
		return nil, err
	}
	// Validate bus type if provided
	if err := ionos.ValidateBusType(bus); err != nil {
		return nil, err
	}

	properties := ionoscloud.VolumeProperties{
		Name: &name,
		Size: &size,
	}
	if volumeType != "" {
		properties.Type = &volumeType
	}
	if bus != "" {
		properties.Bus = &bus
	}
	if availabilityZone != "" {
		properties.AvailabilityZone = &availabilityZone
	}
	if image != "" {
		properties.Image = &image
	}
	if imagePassword != "" {
		properties.ImagePassword = &imagePassword
	}
	if licenceType != "" {
		properties.LicenceType = &licenceType
	}

	volume := ionoscloud.Volume{
		Properties: &properties,
	}

	result, _, err := params.Client.VolumesApi.DatacentersVolumesPost(ctx, datacenterID).Volume(volume).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to create volume: %w", err)
	}
	return api.MarshalResult(result, "volume")
}

func updateVolume(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	volumeID, ok := api.GetRequiredString(params.Arguments, "volume_id")
	if !ok {
		return nil, fmt.Errorf("volume_id is required")
	}

	name := api.GetOptionalString(params.Arguments, "name")
	sizeFloat, sizeSet := api.GetOptionalFloat(params.Arguments, "size")
	size := float32(sizeFloat)

	if name == "" && !sizeSet {
		return nil, fmt.Errorf("at least one of name or size must be provided")
	}

	properties := ionoscloud.VolumeProperties{}
	if name != "" {
		properties.Name = &name
	}
	if sizeSet && size > 0 {
		properties.Size = &size
	}

	result, _, err := params.Client.VolumesApi.DatacentersVolumesPatch(ctx, datacenterID, volumeID).Volume(properties).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to update volume: %w", err)
	}
	return api.MarshalResult(result, "volume")
}

func deleteVolume(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	volumeID, ok := api.GetRequiredString(params.Arguments, "volume_id")
	if !ok {
		return nil, fmt.Errorf("volume_id is required")
	}

	_, err := params.Client.VolumesApi.DatacentersVolumesDelete(ctx, datacenterID, volumeID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to delete volume: %w", err)
	}
	return api.StatusResult(map[string]string{"status": "deleted", "volume_id": volumeID})
}

func attachVolume(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	serverID, ok := api.GetRequiredString(params.Arguments, "server_id")
	if !ok {
		return nil, fmt.Errorf("server_id is required")
	}
	volumeID, ok := api.GetRequiredString(params.Arguments, "volume_id")
	if !ok {
		return nil, fmt.Errorf("volume_id is required")
	}

	volume := ionoscloud.Volume{
		Id: &volumeID,
	}

	result, _, err := params.Client.ServersApi.DatacentersServersVolumesPost(ctx, datacenterID, serverID).Volume(volume).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to attach volume: %w", err)
	}
	return api.MarshalResult(result, "volume")
}

func detachVolume(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	serverID, ok := api.GetRequiredString(params.Arguments, "server_id")
	if !ok {
		return nil, fmt.Errorf("server_id is required")
	}
	volumeID, ok := api.GetRequiredString(params.Arguments, "volume_id")
	if !ok {
		return nil, fmt.Errorf("volume_id is required")
	}

	_, err := params.Client.ServersApi.DatacentersServersVolumesDelete(ctx, datacenterID, serverID, volumeID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to detach volume: %w", err)
	}
	return api.StatusResult(map[string]string{"status": "detached", "volume_id": volumeID, "server_id": serverID})
}
