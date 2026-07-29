package compute

import (
	"context"
	"fmt"
	"strings"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

// A snapshot has no create tool: it is produced from a volume by
// create_volume_snapshot, so a create here would be a second, wrong way to do it.
//
// SnapshotProperties injects defaults through its WithDefaults constructor
// (exposeSerial=false, requireLegacyBios=true), so update_snapshot builds a
// zero-valued literal instead — see the "generated constructors" rule in CLAUDE.md.

// RegisterSnapshotWriteTools registers the snapshot update and delete tools.
func RegisterSnapshotWriteTools(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	registerUpdateSnapshot(server, client, scope)
	registerDeleteSnapshot(server, client, scope, confirm)
}

func registerUpdateSnapshot(server *mcp.Server, client *ionos.APIClient, scope tools.Scope) {
	tools.RegisterTool(server, scope, tools.MethodPatch, &mcp.Tool{
		Name: "update_snapshot",
		Description: "Update a snapshot's name, description, licence type, protection flag or capability flags. Applies a partial update (only the fields you provide). " +
			"The hot-plug and BIOS flags describe what a volume RESTORED from this snapshot will support, so changing them affects future restores rather than the snapshot's stored data. " +
			"There is no create_snapshot: take a snapshot from a volume with create_volume_snapshot. A snapshot's size and location are fixed by the volume it came from.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.UpdateSnapshotInput) (*mcp.CallToolResult, any, error) {
		id := strings.TrimSpace(input.SnapshotID)
		if id == "" {
			return tools.ErrorText("snapshot_id is required"), nil, nil
		}
		if input.Name == nil && input.Description == nil && input.LicenceType == nil &&
			input.SecAuthProtection == nil && input.ExposeSerial == nil && input.RequireLegacyBios == nil &&
			!anyHotPlugFlagSet(input.HotPlugFlags) && !anyExtraHotPlugFlagSet(input.ExtraHotPlugFlags) {
			return tools.ErrorText("nothing to update: provide at least one of name, description, licence_type, sec_auth_protection, expose_serial, require_legacy_bios, or one of the hot-plug flags"), nil, nil
		}

		// A zero-valued literal, NOT NewSnapshotPropertiesWithDefaults(): that
		// constructor pre-sets exposeSerial=false and requireLegacyBios=true, which a
		// PATCH would apply as though the caller had asked. See the "PATCH bodies"
		// note in CLAUDE.md. SnapshotProperties has no unconditionally serialized
		// fields, so nothing needs carrying forward.
		props := &ionos.SnapshotProperties{}
		if input.Name != nil {
			props.SetName(*input.Name)
		}
		if input.Description != nil {
			props.SetDescription(*input.Description)
		}
		if input.LicenceType != nil {
			props.SetLicenceType(*input.LicenceType)
		}
		if input.SecAuthProtection != nil {
			props.SetSecAuthProtection(*input.SecAuthProtection)
		}
		if input.ExposeSerial != nil {
			props.SetExposeSerial(*input.ExposeSerial)
		}
		if input.RequireLegacyBios != nil {
			props.SetRequireLegacyBios(*input.RequireLegacyBios)
		}
		applySnapshotHotPlugFlags(props, input.HotPlugFlags, input.ExtraHotPlugFlags)

		updated, _, err := client.SnapshotsApi.SnapshotsPatch(ctx, id).Snapshot(*props).Execute()
		return tools.ToResult(updated, err)
	})
}

func registerDeleteSnapshot(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterTool(server, scope, tools.MethodDelete, &mcp.Tool{
		Name: "delete_snapshot",
		Description: "Delete a snapshot. Two-phase: call first WITHOUT confirmation_token to get a preview of the snapshot plus a one-time token, then call again WITH the token to delete. " +
			"A snapshot is frequently the only copy of a volume's earlier contents, so deleting it can remove the only way to recover that state — restore_volume_snapshot will no longer have anything to restore from. This is irreversible.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.DeleteSnapshotInput) (*mcp.CallToolResult, any, error) {
		id := strings.TrimSpace(input.SnapshotID)
		if id == "" {
			return tools.ErrorText("snapshot_id is required"), nil, nil
		}
		target := tools.Target(id)

		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "delete_snapshot", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("delete_snapshot", "snapshot_id", err)), nil, nil
			}
			_, err := client.SnapshotsApi.SnapshotsDelete(ctx, id).Execute()
			if err != nil {
				return tools.ToResult(nil, err)
			}
			return tools.TextResult(tools.DeletedAsync("snapshot", id)), nil, nil
		}

		snap, _, err := client.SnapshotsApi.SnapshotsFindById(ctx, id).Depth(1).Execute()
		if err != nil {
			if tools.IsNotFound(err) {
				return tools.ErrorText(fmt.Sprintf("snapshot %s does not exist; nothing to delete", id)), nil, nil
			}
			return tools.ToResult(nil, err)
		}
		props := snap.GetProperties()
		token, mErr := confirm.Mint("delete_snapshot", target)
		if mErr != nil {
			return nil, nil, mErr
		}
		headline := "About to DELETE a snapshot. This is IRREVERSIBLE.\n" +
			"If this is the only snapshot of that volume's earlier state, deleting it removes the only way back — restore_volume_snapshot will have nothing to restore from."
		if props.GetSecAuthProtection() {
			headline += "\nNOTE: this snapshot is marked as requiring extra protection (sec_auth_protection), which suggests it was flagged as important."
		}
		return tools.TextResult(tools.Preview{
			Headline: headline,
			Fields: tools.Fields(
				"snapshot_id", id,
				"name", props.GetName(),
				"description", props.GetDescription(),
				"location", props.GetLocation(),
				"size (GB)", fmt.Sprintf("%g", props.GetSize()),
				"licence_type", props.GetLicenceType(),
			),
			Tool:      "delete_snapshot",
			Replay:    tools.Fields("snapshot_id", id),
			TokenNote: "This token authorizes deleting ONLY this snapshot",
		}.Render(token)), nil, nil
	})
}

// applySnapshotHotPlugFlags sets the capability flags a snapshot supports. Snapshots
// and images accept a wider set than volumes, so the shared HotPlugFlags cover the
// common six and ExtraHotPlugFlags the remaining four.
func applySnapshotHotPlugFlags(props *ionos.SnapshotProperties, f tools.HotPlugFlags, x tools.ExtraHotPlugFlags) {
	if f.CpuHotPlug != nil {
		props.SetCpuHotPlug(*f.CpuHotPlug)
	}
	if f.RamHotPlug != nil {
		props.SetRamHotPlug(*f.RamHotPlug)
	}
	if f.NicHotPlug != nil {
		props.SetNicHotPlug(*f.NicHotPlug)
	}
	if f.NicHotUnplug != nil {
		props.SetNicHotUnplug(*f.NicHotUnplug)
	}
	if f.DiscVirtioHotPlug != nil {
		props.SetDiscVirtioHotPlug(*f.DiscVirtioHotPlug)
	}
	if f.DiscVirtioHotUnplug != nil {
		props.SetDiscVirtioHotUnplug(*f.DiscVirtioHotUnplug)
	}
	if x.CpuHotUnplug != nil {
		props.SetCpuHotUnplug(*x.CpuHotUnplug)
	}
	if x.RamHotUnplug != nil {
		props.SetRamHotUnplug(*x.RamHotUnplug)
	}
	if x.DiscScsiHotPlug != nil {
		props.SetDiscScsiHotPlug(*x.DiscScsiHotPlug)
	}
	if x.DiscScsiHotUnplug != nil {
		props.SetDiscScsiHotUnplug(*x.DiscScsiHotUnplug)
	}
}

// anyExtraHotPlugFlagSet reports whether any of the snapshot/image-only capability
// flags was supplied, so an otherwise-empty update is still rejected.
func anyExtraHotPlugFlagSet(f tools.ExtraHotPlugFlags) bool {
	return f.CpuHotUnplug != nil || f.RamHotUnplug != nil ||
		f.DiscScsiHotPlug != nil || f.DiscScsiHotUnplug != nil
}
