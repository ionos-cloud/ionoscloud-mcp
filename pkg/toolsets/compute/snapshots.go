package compute

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ionos-cloud/ionoscloud-mcp/pkg/api"
	ionoscloud "github.com/ionos-cloud/sdk-go/v6"
)

func initSnapshots() []api.ServerTool {
	return []api.ServerTool{
		{
			Tool: api.Tool{
				Name:        "list_snapshots",
				Description: "List all snapshots in your IONOS Cloud account",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {},
					"required": []
				}`),
				Annotations: api.ReadOnly("List Snapshots"),
			},
			Handler: listSnapshots,
		},
		{
			Tool: api.Tool{
				Name:        "get_snapshot",
				Description: "Get details of a specific snapshot",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"snapshot_id": {
							"type": "string",
							"description": "The ID of the snapshot"
						}
					},
					"required": ["snapshot_id"]
				}`),
				Annotations: api.ReadOnly("Get Snapshot"),
			},
			Handler: getSnapshot,
		},
		{
			Tool: api.Tool{
				Name:        "create_snapshot",
				Description: "Create a snapshot of a volume",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {
							"type": "string",
							"description": "The ID of the data center"
						},
						"volume_id": {
							"type": "string",
							"description": "The ID of the volume to snapshot"
						},
						"name": {
							"type": "string",
							"description": "The name of the snapshot"
						},
						"description": {
							"type": "string",
							"description": "A description for the snapshot"
						}
					},
					"required": ["datacenter_id", "volume_id"]
				}`),
				Annotations: api.NonIdempotent("Create Snapshot"),
			},
			Handler: createSnapshot,
		},
		{
			Tool: api.Tool{
				Name:        "update_snapshot",
				Description: "Update snapshot metadata",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"snapshot_id": {
							"type": "string",
							"description": "The ID of the snapshot to update"
						},
						"name": {
							"type": "string",
							"description": "The new name for the snapshot"
						},
						"description": {
							"type": "string",
							"description": "The new description for the snapshot"
						}
					},
					"required": ["snapshot_id"]
				}`),
				Annotations: api.Idempotent("Update Snapshot"),
			},
			Handler: updateSnapshot,
		},
		{
			Tool: api.Tool{
				Name:        "delete_snapshot",
				Description: "Delete a snapshot",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"snapshot_id": {
							"type": "string",
							"description": "The ID of the snapshot to delete"
						}
					},
					"required": ["snapshot_id"]
				}`),
				Annotations: api.Destructive("Delete Snapshot"),
			},
			Handler: deleteSnapshot,
		},
		{
			Tool: api.Tool{
				Name:        "restore_snapshot",
				Description: "Restore a volume from a snapshot",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {
							"type": "string",
							"description": "The ID of the data center"
						},
						"volume_id": {
							"type": "string",
							"description": "The ID of the volume to restore"
						},
						"snapshot_id": {
							"type": "string",
							"description": "The ID of the snapshot to restore from"
						}
					},
					"required": ["datacenter_id", "volume_id", "snapshot_id"]
				}`),
				Annotations: api.Destructive("Restore Snapshot"),
			},
			Handler: restoreSnapshot,
		},
	}
}

func listSnapshots(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	snapshots, _, err := params.Client.SnapshotsApi.SnapshotsGet(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to list snapshots: %w", err)
	}
	return api.MarshalResult(snapshots, "snapshots")
}

func getSnapshot(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	snapshotID, ok := api.GetRequiredString(params.Arguments, "snapshot_id")
	if !ok {
		return nil, fmt.Errorf("snapshot_id is required")
	}

	snapshot, _, err := params.Client.SnapshotsApi.SnapshotsFindById(ctx, snapshotID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get snapshot: %w", err)
	}
	return api.MarshalResult(snapshot, "snapshot")
}

func createSnapshot(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	volumeID, ok := api.GetRequiredString(params.Arguments, "volume_id")
	if !ok {
		return nil, fmt.Errorf("volume_id is required")
	}

	name := api.GetOptionalString(params.Arguments, "name")
	description := api.GetOptionalString(params.Arguments, "description")

	properties := ionoscloud.CreateSnapshotProperties{}
	if name != "" {
		properties.Name = &name
	}
	if description != "" {
		properties.Description = &description
	}

	snapshot := ionoscloud.CreateSnapshot{
		Properties: &properties,
	}

	result, _, err := params.Client.VolumesApi.DatacentersVolumesCreateSnapshotPost(ctx, datacenterID, volumeID).Snapshot(snapshot).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to create snapshot: %w", err)
	}
	return api.MarshalResult(result, "snapshot")
}

func updateSnapshot(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	snapshotID, ok := api.GetRequiredString(params.Arguments, "snapshot_id")
	if !ok {
		return nil, fmt.Errorf("snapshot_id is required")
	}

	name := api.GetOptionalString(params.Arguments, "name")
	description := api.GetOptionalString(params.Arguments, "description")

	if name == "" && description == "" {
		return nil, fmt.Errorf("at least one of name or description must be provided")
	}

	properties := ionoscloud.SnapshotProperties{}
	if name != "" {
		properties.Name = &name
	}
	if description != "" {
		properties.Description = &description
	}

	result, _, err := params.Client.SnapshotsApi.SnapshotsPatch(ctx, snapshotID).Snapshot(properties).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to update snapshot: %w", err)
	}
	return api.MarshalResult(result, "snapshot")
}

func deleteSnapshot(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	snapshotID, ok := api.GetRequiredString(params.Arguments, "snapshot_id")
	if !ok {
		return nil, fmt.Errorf("snapshot_id is required")
	}

	_, err := params.Client.SnapshotsApi.SnapshotsDelete(ctx, snapshotID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to delete snapshot: %w", err)
	}
	return api.StatusResult(map[string]string{"status": "deleted", "snapshot_id": snapshotID})
}

func restoreSnapshot(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	volumeID, ok := api.GetRequiredString(params.Arguments, "volume_id")
	if !ok {
		return nil, fmt.Errorf("volume_id is required")
	}
	snapshotID, ok := api.GetRequiredString(params.Arguments, "snapshot_id")
	if !ok {
		return nil, fmt.Errorf("snapshot_id is required")
	}

	properties := ionoscloud.RestoreSnapshotProperties{
		SnapshotId: &snapshotID,
	}
	restoreSnapshot := ionoscloud.RestoreSnapshot{
		Properties: &properties,
	}

	_, err := params.Client.VolumesApi.DatacentersVolumesRestoreSnapshotPost(ctx, datacenterID, volumeID).RestoreSnapshot(restoreSnapshot).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to restore snapshot: %w", err)
	}
	return api.StatusResult(map[string]string{"status": "restoring", "volume_id": volumeID, "snapshot_id": snapshotID})
}
