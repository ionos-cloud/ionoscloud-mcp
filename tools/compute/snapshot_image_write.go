package compute

import (
	"context"
	"fmt"
	"strings"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

// Snapshots and images are the two compute resources with no create tool, for
// different reasons: a snapshot is produced from a volume by create_volume_snapshot,
// and the API exposes no way to create an image at all — public images come from
// IONOS and private ones are uploaded out of band.
//
// Both models inject defaults through their WithDefaults constructors
// (exposeSerial=false, requireLegacyBios=true), so both updates build a zero-valued
// literal instead. ImageProperties additionally serializes licenceType
// unconditionally, so update_image carries the current value forward.

// RegisterSnapshotImageWriteTools registers the snapshot and image update/delete
// tools.
func RegisterSnapshotImageWriteTools(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	registerUpdateSnapshot(server, client, scope)
	registerDeleteSnapshot(server, client, scope, confirm)
	registerUpdateImage(server, client, scope)
	registerDeleteImage(server, client, scope, confirm)
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

func registerUpdateImage(server *mcp.Server, client *ionos.APIClient, scope tools.Scope) {
	tools.RegisterTool(server, scope, tools.MethodPatch, &mcp.Tool{
		Name: "update_image",
		Description: "Update a private image's name, description, licence type, cloud-init compatibility or capability flags. Applies a partial update (only the fields you provide). " +
			"Only an image you uploaded can be changed — public IONOS images are read-only and the API rejects attempts to modify them. " +
			"Omit licence_type to keep the current value; it is read and sent back unchanged because the API always receives it. " +
			"The hot-plug and BIOS flags describe what a volume CREATED from this image will support. There is no create_image: the API exposes no way to create one.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.UpdateImageInput) (*mcp.CallToolResult, any, error) {
		id := strings.TrimSpace(input.ImageID)
		if id == "" {
			return tools.ErrorText("image_id is required"), nil, nil
		}
		if input.Name == nil && input.Description == nil && input.LicenceType == nil &&
			input.CloudInit == nil && input.ExposeSerial == nil && input.RequireLegacyBios == nil &&
			!anyHotPlugFlagSet(input.HotPlugFlags) && !anyExtraHotPlugFlagSet(input.ExtraHotPlugFlags) {
			return tools.ErrorText("nothing to update: provide at least one of name, description, licence_type, cloud_init, expose_serial, require_legacy_bios, or one of the hot-plug flags"), nil, nil
		}

		// ImageProperties.LicenceType is a non-pointer field the SDK ALWAYS
		// serializes, so a PATCH built without it would send an empty licence type.
		// Read the current value and let the caller override only if they asked to.
		licenceType := ""
		if input.LicenceType != nil {
			licenceType = strings.TrimSpace(*input.LicenceType)
			if licenceType == "" {
				return tools.ErrorText("licence_type must not be empty; omit it entirely to keep the image's current licence type"), nil, nil
			}
		} else {
			current, _, err := client.ImagesApi.ImagesFindById(ctx, id).Depth(1).Execute()
			if err != nil {
				if tools.IsNotFound(err) {
					return tools.ErrorText(fmt.Sprintf("image %s does not exist; nothing to update", id)), nil, nil
				}
				return tools.ToResult(nil, err)
			}
			currentProps := current.GetProperties()
			licenceType = currentProps.GetLicenceType()
		}

		// A keyed literal, NOT NewImageProperties(licenceType): even that
		// required-argument constructor injects exposeSerial=false and
		// requireLegacyBios=true, so a PATCH built from it would force the legacy
		// BIOS on and stop exposing the serial on every volume made from this image.
		// It is not only the WithDefaults constructors that do this — see the "PATCH
		// bodies" note in CLAUDE.md.
		props := &ionos.ImageProperties{LicenceType: licenceType}
		if input.Name != nil {
			props.SetName(*input.Name)
		}
		if input.Description != nil {
			props.SetDescription(*input.Description)
		}
		if input.CloudInit != nil {
			props.SetCloudInit(*input.CloudInit)
		}
		if input.ExposeSerial != nil {
			props.SetExposeSerial(*input.ExposeSerial)
		}
		if input.RequireLegacyBios != nil {
			props.SetRequireLegacyBios(*input.RequireLegacyBios)
		}
		applyImageHotPlugFlags(props, input.HotPlugFlags, input.ExtraHotPlugFlags)

		updated, _, err := client.ImagesApi.ImagesPatch(ctx, id).Image(*props).Execute()
		return tools.ToResult(updated, err)
	})
}

func registerDeleteImage(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterTool(server, scope, tools.MethodDelete, &mcp.Tool{
		Name: "delete_image",
		Description: "Delete a private image. Two-phase: call first WITHOUT confirmation_token to get a preview of the image plus a one-time token, then call again WITH the token to delete. " +
			"Only an image you uploaded can be deleted — public IONOS images cannot. Anything that references the image by ID stops being able to create volumes from it, including Terraform configurations, scripts and VM autoscaling templates, so check for references before deleting. This is irreversible.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.DeleteImageInput) (*mcp.CallToolResult, any, error) {
		id := strings.TrimSpace(input.ImageID)
		if id == "" {
			return tools.ErrorText("image_id is required"), nil, nil
		}
		target := tools.Target(id)

		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "delete_image", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("delete_image", "image_id", err)), nil, nil
			}
			_, err := client.ImagesApi.ImagesDelete(ctx, id).Execute()
			if err != nil {
				return tools.ToResult(nil, err)
			}
			return tools.TextResult(tools.DeletedAsync("image", id)), nil, nil
		}

		img, _, err := client.ImagesApi.ImagesFindById(ctx, id).Depth(1).Execute()
		if err != nil {
			if tools.IsNotFound(err) {
				return tools.ErrorText(fmt.Sprintf("image %s does not exist; nothing to delete", id)), nil, nil
			}
			return tools.ToResult(nil, err)
		}
		props := img.GetProperties()
		token, mErr := confirm.Mint("delete_image", target)
		if mErr != nil {
			return nil, nil, mErr
		}
		headline := "About to DELETE an image. This is IRREVERSIBLE.\n" +
			"Anything referencing this image by ID — Terraform configurations, scripts, VM autoscaling templates — stops being able to create volumes from it."
		if props.GetPublic() {
			// The API will refuse this, but saying so up front is more useful than a
			// generic rejection after the token round trip.
			headline = "About to attempt DELETING a PUBLIC image.\n" +
				"Public IONOS images cannot be deleted and the API will reject this. Only images you uploaded yourself can be removed."
		}
		return tools.TextResult(tools.Preview{
			Headline: headline,
			Fields: tools.Fields(
				"image_id", id,
				"name", props.GetName(),
				"description", props.GetDescription(),
				"location", props.GetLocation(),
				"size (GB)", fmt.Sprintf("%g", props.GetSize()),
				"licence_type", props.GetLicenceType(),
				"image_type", props.GetImageType(),
				"public", fmt.Sprintf("%t", props.GetPublic()),
				"image_aliases", aliasSummary(props.GetImageAliases()),
			),
			Tool:      "delete_image",
			Replay:    tools.Fields("image_id", id),
			TokenNote: "This token authorizes deleting ONLY this image",
		}.Render(token)), nil, nil
	})
}

// anyExtraHotPlugFlagSet reports whether any of the snapshot/image-only capability
// flags was supplied, so an otherwise-empty update is still rejected.
func anyExtraHotPlugFlagSet(f tools.ExtraHotPlugFlags) bool {
	return f.CpuHotUnplug != nil || f.RamHotUnplug != nil ||
		f.DiscScsiHotPlug != nil || f.DiscScsiHotUnplug != nil
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

// applyImageHotPlugFlags is applySnapshotHotPlugFlags for images. The two generated
// types have identical setters but no shared interface, so the bodies cannot be
// merged without reflection.
func applyImageHotPlugFlags(props *ionos.ImageProperties, f tools.HotPlugFlags, x tools.ExtraHotPlugFlags) {
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

// aliasSummary lists an image's aliases, which are how most callers refer to it —
// deleting the image breaks any script using one of these.
func aliasSummary(aliases []string) string {
	if len(aliases) == 0 {
		return ""
	}
	return strings.Join(aliases, ", ")
}
