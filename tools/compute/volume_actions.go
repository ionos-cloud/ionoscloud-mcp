package compute

import (
	"context"
	"fmt"
	"strings"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

// Volume snapshot actions. create_volume_snapshot creates a resource; restoring
// overwrites a volume, so it registers as a destructive action verb instead.

// RegisterVolumeActionTools registers the volume snapshot create/restore actions.
func RegisterVolumeActionTools(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	registerCreateVolumeSnapshot(server, client, scope, confirm)
	registerRestoreVolumeSnapshot(server, client, scope, confirm)
}

func registerCreateVolumeSnapshot(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterTool(server, scope, tools.MethodPost, &mcp.Tool{
		Name: "create_volume_snapshot",
		Description: "Take a snapshot of a volume, capturing its contents at this moment. Requires IONOS_MCP_TOOL_SCOPE to include write. " +
			"Two-phase: call first WITHOUT confirmation_token to get a preview and a one-time token, then call again WITH the token (and the same datacenter_id, volume_id and name) to create it. " +
			"A snapshot is the ONLY way to recover a volume's data after delete_volume or restore_volume_snapshot, so take one before any destructive change you might need to undo. " +
			"Snapshots are billed for the storage they occupy. This is the only way to create a snapshot — there is no standalone create_snapshot.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.CreateVolumeSnapshotInput) (*mcp.CallToolResult, any, error) {
		dcID := strings.TrimSpace(input.DatacenterID)
		volumeID := strings.TrimSpace(input.VolumeID)
		name := strings.TrimSpace(input.Name)
		if dcID == "" {
			return tools.ErrorText("datacenter_id is required"), nil, nil
		}
		if volumeID == "" {
			return tools.ErrorText("volume_id is required"), nil, nil
		}
		if name == "" {
			return tools.ErrorText("name is required for the new snapshot"), nil, nil
		}
		target := tools.Target(dcID, volumeID, name)

		// Phase 2: token present -> validate and execute.
		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "create_volume_snapshot", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("create_volume_snapshot", "datacenter_id, volume_id and name", err)), nil, nil
			}
			props := &ionos.CreateSnapshotProperties{}
			props.SetName(name)
			if input.Description != nil {
				props.SetDescription(*input.Description)
			}
			if input.SecAuthProtection != nil {
				props.SetSecAuthProtection(*input.SecAuthProtection)
			}
			if input.LicenceType != nil {
				props.SetLicenceType(*input.LicenceType)
			}
			body := ionos.NewCreateSnapshot()
			body.SetProperties(*props)
			created, _, err := client.VolumesApi.DatacentersVolumesCreateSnapshotPost(ctx, dcID, volumeID).Snapshot(*body).Execute()
			return tools.ToResult(created, err)
		}

		// Phase 1: no token -> read the volume so the preview can name it.
		vol, _, err := client.VolumesApi.DatacentersVolumesFindById(ctx, dcID, volumeID).Depth(1).Execute()
		if err != nil {
			if tools.IsNotFound(err) {
				return tools.ErrorText(fmt.Sprintf("volume %s does not exist in data center %s; nothing to snapshot", volumeID, dcID)), nil, nil
			}
			return tools.ToResult(nil, err)
		}
		volProps := vol.GetProperties()
		token, mErr := confirm.Mint("create_volume_snapshot", target)
		if mErr != nil {
			return nil, nil, mErr
		}
		return tools.TextResult(tools.Preview{
			Headline: "About to CREATE one snapshot of a volume:",
			Fields: tools.Fields(
				"datacenter_id", dcID,
				"volume_id", volumeID,
				"volume name", volProps.GetName(),
				"volume size (GB)", fmt.Sprintf("%g", volProps.GetSize()),
				"snapshot name", name,
				"description", tools.OptStr(input.Description),
				"sec_auth_protection", tools.OptBool(input.SecAuthProtection),
				"licence_type", tools.OptStr(input.LicenceType),
			),
			Tool:      "create_volume_snapshot",
			Replay:    tools.Fields("datacenter_id", dcID, "volume_id", volumeID, "name", name),
			TokenNote: "The snapshot is billed for the storage it occupies. This token authorizes creating only this snapshot of only this volume",
		}.Render(token)), nil, nil
	})
}

func registerRestoreVolumeSnapshot(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	// Destructive despite creating nothing: it overwrites the target volume, and a
	// second restore discards whatever the guest wrote after the first.
	tools.RegisterActionTool(server, scope,
		tools.Action{Verb: "restore_", Method: tools.MethodPost, Idempotent: false},
		&mcp.Tool{
			Name: "restore_volume_snapshot",
			Description: "Restore a snapshot onto a volume, OVERWRITING everything currently on that volume. Requires IONOS_MCP_TOOL_SCOPE to include destructive. " +
				"Two-phase: call first WITHOUT confirmation_token to see which volume will be overwritten, plus a one-time token, then call again WITH the token to restore. " +
				"The volume's current contents are unrecoverable afterwards unless you snapshot them first with create_volume_snapshot. Stop the server before restoring its boot volume, or the running guest and the restored disk will disagree.",
		},
		func(ctx context.Context, req *mcp.CallToolRequest, input tools.RestoreVolumeSnapshotInput) (*mcp.CallToolResult, any, error) {
			dcID := strings.TrimSpace(input.DatacenterID)
			volumeID := strings.TrimSpace(input.VolumeID)
			snapshotID := strings.TrimSpace(input.SnapshotID)
			if dcID == "" {
				return tools.ErrorText("datacenter_id is required"), nil, nil
			}
			if volumeID == "" {
				return tools.ErrorText("volume_id is required"), nil, nil
			}
			if snapshotID == "" {
				return tools.ErrorText("snapshot_id is required; list available snapshots with list_snapshots"), nil, nil
			}
			// Volume and snapshot are both in the target, so a token cannot be
			// replayed to restore a different snapshot.
			target := tools.Target(dcID, volumeID, snapshotID)

			// Phase 2: token present -> validate and execute.
			if tools.HasToken(input.ConfirmationToken) {
				if err := confirm.Consume(*input.ConfirmationToken, "restore_volume_snapshot", target); err != nil {
					return tools.ErrorText(tools.ConfirmErrorText("restore_volume_snapshot", "datacenter_id, volume_id and snapshot_id", err)), nil, nil
				}
				props := &ionos.RestoreSnapshotProperties{}
				props.SetSnapshotId(snapshotID)
				body := ionos.NewRestoreSnapshot()
				body.SetProperties(*props)
				_, err := client.VolumesApi.DatacentersVolumesRestoreSnapshotPost(ctx, dcID, volumeID).RestoreSnapshot(*body).Execute()
				if err != nil {
					return tools.ToResult(nil, err)
				}
				// This endpoint returns no body.
				return tools.TextResult(fmt.Sprintf("Restoring snapshot %s onto volume %s. The restore is asynchronous; the API has accepted the request. The volume's previous contents are gone.", snapshotID, volumeID)), nil, nil
			}

			// Phase 1: no token -> name the volume about to be overwritten.
			vol, _, err := client.VolumesApi.DatacentersVolumesFindById(ctx, dcID, volumeID).Depth(1).Execute()
			if err != nil {
				if tools.IsNotFound(err) {
					return tools.ErrorText(fmt.Sprintf("volume %s does not exist in data center %s; nothing to restore into", volumeID, dcID)), nil, nil
				}
				return tools.ToResult(nil, err)
			}
			volProps := vol.GetProperties()
			token, mErr := confirm.Mint("restore_volume_snapshot", target)
			if mErr != nil {
				return nil, nil, mErr
			}
			headline := "About to RESTORE a snapshot onto a volume, OVERWRITING everything on it. This is IRREVERSIBLE.\n" +
				"Take a snapshot of the volume's current contents first with create_volume_snapshot if you might need them."
			if bootServer := volProps.GetBootServer(); bootServer != "" {
				headline += fmt.Sprintf("\nWARNING: this volume is attached to server %s. Stop that server with stop_server before restoring, or the running guest and the restored disk will disagree.", bootServer)
			}
			return tools.TextResult(tools.Preview{
				Headline: headline,
				Fields: tools.Fields(
					"datacenter_id", dcID,
					"volume_id", volumeID,
					"volume name", volProps.GetName(),
					"volume size (GB)", fmt.Sprintf("%g", volProps.GetSize()),
					"attached to server", volProps.GetBootServer(),
					"snapshot_id (source)", snapshotID,
				),
				Tool:      "restore_volume_snapshot",
				Replay:    tools.Fields("datacenter_id", dcID, "volume_id", volumeID, "snapshot_id", snapshotID),
				TokenNote: "This token authorizes restoring ONLY this snapshot onto ONLY this volume",
			}.Render(token)), nil, nil
		})
}
