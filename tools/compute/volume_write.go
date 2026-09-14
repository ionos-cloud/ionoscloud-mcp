package compute

import (
	"context"
	"fmt"
	"strings"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

// RegisterVolumeWriteTools registers the create/update/delete volume tools.
// create and delete are two-phase confirmed; update is a single partial PATCH.
func RegisterVolumeWriteTools(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	registerCreateVolume(server, client, scope, confirm)
	registerUpdateVolume(server, client, scope)
	registerDeleteVolume(server, client, scope, confirm)
}

func registerCreateVolume(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterTool(server, scope, tools.MethodPost, &mcp.Tool{
		Name: "create_volume",
		Description: "Create one storage volume in a data center. Two-phase: call first WITHOUT confirmation_token to get a preview and a one-time token, then call again WITH the token (and the same datacenter_id and name) to create it. " +
			"Give image or image_alias to get a bootable volume with an operating system; without either you get an empty disk and must also give licence_type. " +
			"For a bootable Linux volume set ssh_keys or image_password, or you will not be able to log in. The new volume is NOT attached to any server — use attach_server_volume afterwards. Creates exactly one volume per call.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.CreateVolumeInput) (*mcp.CallToolResult, any, error) {
		dcID := strings.TrimSpace(input.DatacenterID)
		name := strings.TrimSpace(input.Name)
		volType := strings.TrimSpace(input.Type)
		if dcID == "" {
			return tools.ErrorText("datacenter_id is required to create a volume"), nil, nil
		}
		if name == "" {
			return tools.ErrorText("name is required to create a volume"), nil, nil
		}
		if volType == "" {
			return tools.ErrorText("type is required to create a volume (HDD, SSD, SSD Standard, SSD Premium, or DAS)"), nil, nil
		}
		if input.Size <= 0 {
			return tools.ErrorText("size must be greater than 0 GB"), nil, nil
		}
		// An empty volume has no image to infer a licence type from, so the API
		// rejects it. Catching that here saves a round trip and says what to do.
		hasImage := (input.Image != nil && *input.Image != "") || (input.ImageAlias != nil && *input.ImageAlias != "")
		hasLicence := input.LicenceType != nil && *input.LicenceType != ""
		if !hasImage && !hasLicence {
			return tools.ErrorText("provide image or image_alias to create a bootable volume (see list_images), or licence_type to create an empty one"), nil, nil
		}
		target := tools.Target(req, dcID, name)

		// Phase 2: token present -> validate and execute.
		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "create_volume", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("create_volume", "datacenter_id and name", err)), nil, nil
			}
			// A literal, not a generated constructor: those inject bootOrder, exposeSerial
			// and requireLegacyBios, which would change a server's boot behaviour.
			props := &ionos.VolumeProperties{}
			props.SetName(name)
			props.SetType(volType)
			props.SetSize(input.Size)
			if input.Image != nil {
				props.SetImage(*input.Image)
			}
			if input.ImageAlias != nil {
				props.SetImageAlias(*input.ImageAlias)
			}
			if input.ImagePassword != nil {
				props.SetImagePassword(*input.ImagePassword)
			}
			if len(input.SshKeys) > 0 {
				props.SetSshKeys(input.SshKeys)
			}
			if input.LicenceType != nil {
				props.SetLicenceType(*input.LicenceType)
			}
			if input.AvailabilityZone != nil {
				props.SetAvailabilityZone(*input.AvailabilityZone)
			}
			if input.Bus != nil {
				props.SetBus(*input.Bus)
			}
			if input.UserData != nil {
				props.SetUserData(*input.UserData)
			}
			if input.BackupunitId != nil {
				props.SetBackupunitId(*input.BackupunitId)
			}
			if input.ExposeSerial != nil {
				props.SetExposeSerial(*input.ExposeSerial)
			}
			applyHotPlugFlags(props, input.HotPlugFlags)
			body := ionos.NewVolume()
			body.SetProperties(*props)
			created, _, err := client.VolumesApi.DatacentersVolumesPost(ctx, dcID).Volume(*body).Execute()
			return tools.ToResult(created, err)
		}

		// Phase 1: no token -> preview and mint a one-time token.
		token, err := confirm.Mint("create_volume", target)
		if err != nil {
			return nil, nil, err
		}
		fields := tools.Fields(
			"datacenter_id", dcID,
			"name", name,
			"size (GB)", fmt.Sprintf("%g", input.Size),
			"type", volType,
			"image", tools.OptStr(input.Image),
			"image_alias", tools.OptStr(input.ImageAlias),
			"image_password", tools.Redacted(input.ImagePassword),
			"ssh_keys", sshKeySummary(input.SshKeys),
			"licence_type", tools.OptStr(input.LicenceType),
			"availability_zone", tools.OptStr(input.AvailabilityZone),
			"bus", tools.OptStr(input.Bus),
			"user_data", tools.Redacted(input.UserData),
			"backupunit_id", tools.OptStr(input.BackupunitId),
			"expose_serial", tools.OptBool(input.ExposeSerial),
		)
		fields = append(fields, hotPlugPreviewFields("", input.HotPlugFlags)...)
		return tools.TextResult(tools.Preview{
			Headline:  "About to CREATE one storage volume:",
			Fields:    fields,
			Tool:      "create_volume",
			Replay:    tools.Fields("datacenter_id", dcID, "name", name),
			TokenNote: "This creates exactly one volume, not attached to any server. The token authorizes creating only this volume in this data center",
		}.Render(token)), nil, nil
	})
}

func registerUpdateVolume(server *mcp.Server, client *ionos.APIClient, scope tools.Scope) {
	tools.RegisterTool(server, scope, tools.MethodPatch, &mcp.Tool{
		Name: "update_volume",
		Description: "Update a volume's name, size, bus type, serial exposure or boot order. Applies a partial update (only the fields you provide). " +
			"boot_order controls whether the volume is the server's boot disk (PRIMARY, NONE or AUTO), but PRIMARY requires every other volume on the server to be NONE — to just point a server at a different disk, use update_server with boot_volume_id instead. " +
			"size can only be INCREASED — the API rejects shrinking a volume — and after growing it the guest operating system still has to extend its own filesystem to use the new space. " +
			"The cpu_hot_plug, ram_hot_plug, nic_hot_plug, nic_hot_unplug, disc_virtio_hot_plug and disc_virtio_hot_unplug flags control whether the server this volume is attached to can change that hardware without rebooting — set them before you need to resize a running server. " +
			"image_password, user_data and backupunit_id are set at creation only and cannot be changed here.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.UpdateVolumeInput) (*mcp.CallToolResult, any, error) {
		dcID := strings.TrimSpace(input.DatacenterID)
		id := strings.TrimSpace(input.VolumeID)
		if dcID == "" {
			return tools.ErrorText("datacenter_id is required"), nil, nil
		}
		if id == "" {
			return tools.ErrorText("volume_id is required"), nil, nil
		}
		if input.Name == nil && input.Size == nil && input.Bus == nil &&
			input.ExposeSerial == nil && input.BootOrder == nil && !anyHotPlugFlagSet(input.HotPlugFlags) {
			return tools.ErrorText("nothing to update: provide at least one of name, size, bus, expose_serial, boot_order, cpu_hot_plug, ram_hot_plug, nic_hot_plug, nic_hot_unplug, disc_virtio_hot_plug, disc_virtio_hot_unplug"), nil, nil
		}
		// A literal, not a generated constructor: a PATCH must carry only the fields
		// the caller supplied, or a rename would also force legacy BIOS on.
		props := &ionos.VolumeProperties{}
		if input.Name != nil {
			props.SetName(*input.Name)
		}
		if input.Size != nil {
			props.SetSize(*input.Size)
		}
		if input.Bus != nil {
			props.SetBus(*input.Bus)
		}
		if input.ExposeSerial != nil {
			props.SetExposeSerial(*input.ExposeSerial)
		}
		if input.BootOrder != nil {
			order := strings.ToUpper(strings.TrimSpace(*input.BootOrder))
			switch order {
			case "PRIMARY", "NONE", "AUTO":
			default:
				return tools.ErrorText(fmt.Sprintf("boot_order must be PRIMARY, NONE or AUTO, got %q", *input.BootOrder)), nil, nil
			}
			props.SetBootOrder(order)
		}
		applyHotPlugFlags(props, input.HotPlugFlags)
		updated, _, err := client.VolumesApi.DatacentersVolumesPatch(ctx, dcID, id).Volume(*props).Execute()
		return tools.ToResult(updated, err)
	})
}

func registerDeleteVolume(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterTool(server, scope, tools.MethodDelete, &mcp.Tool{
		Name: "delete_volume",
		Description: "Delete a storage volume and all data on it. Two-phase: call first WITHOUT confirmation_token to get a preview (including whether the volume is still attached to a server) and a one-time token, then call again WITH the token to delete. " +
			"The data is unrecoverable unless a snapshot exists — take one with create_volume_snapshot first if you might need it. This is irreversible.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.DeleteVolumeInput) (*mcp.CallToolResult, any, error) {
		dcID := strings.TrimSpace(input.DatacenterID)
		id := strings.TrimSpace(input.VolumeID)
		if dcID == "" {
			return tools.ErrorText("datacenter_id is required"), nil, nil
		}
		if id == "" {
			return tools.ErrorText("volume_id is required"), nil, nil
		}
		target := tools.Target(req, dcID, id)

		// Phase 2: token present -> validate and execute.
		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "delete_volume", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("delete_volume", "datacenter_id and volume_id", err)), nil, nil
			}
			_, err := client.VolumesApi.DatacentersVolumesDelete(ctx, dcID, id).Execute()
			if err != nil {
				return tools.ToResult(nil, err)
			}
			return tools.TextResult(tools.DeletedAsync("volume", id)), nil, nil
		}

		// Phase 1: no token -> read the volume so the preview can name it.
		vol, _, err := client.VolumesApi.DatacentersVolumesFindById(ctx, dcID, id).Depth(1).Execute()
		if err != nil {
			if tools.IsNotFound(err) {
				return tools.ErrorText(fmt.Sprintf("volume %s does not exist in data center %s; nothing to delete", id, dcID)), nil, nil
			}
			return tools.ToResult(nil, err)
		}
		props := vol.GetProperties()
		token, mErr := confirm.Mint("delete_volume", target)
		if mErr != nil {
			return nil, nil, mErr
		}
		note := "All data on this volume is destroyed and cannot be recovered without a snapshot."
		if bootServer := props.GetBootServer(); bootServer != "" {
			note = fmt.Sprintf("WARNING: this volume is still attached to server %s and may be its boot disk — deleting it can leave that server unable to boot. %s", bootServer, note)
		}
		return tools.TextResult(tools.Preview{
			Headline: "About to DELETE a storage volume. This is IRREVERSIBLE.\n" + note,
			Fields: tools.Fields(
				"datacenter_id", dcID,
				"volume_id", id,
				"name", props.GetName(),
				"size (GB)", fmt.Sprintf("%g", props.GetSize()),
				"type", props.GetType(),
				"attached to server", props.GetBootServer(),
			),
			Tool:      "delete_volume",
			Replay:    tools.Fields("datacenter_id", dcID, "volume_id", id),
			TokenNote: "This token authorizes deleting ONLY this volume",
		}.Render(token)), nil, nil
	})
}

// applyHotPlugFlags copies the supplied capability flags onto volume properties.
func applyHotPlugFlags(props *ionos.VolumeProperties, f tools.HotPlugFlags) {
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
}

// anyHotPlugFlagSet reports whether the caller supplied any capability flag, so
// update_volume's "nothing to update" guard counts them as a real change.
func anyHotPlugFlagSet(f tools.HotPlugFlags) bool {
	return f.CpuHotPlug != nil || f.RamHotPlug != nil || f.NicHotPlug != nil ||
		f.NicHotUnplug != nil || f.DiscVirtioHotPlug != nil || f.DiscVirtioHotUnplug != nil
}

// hotPlugPreviewFields renders the capability flags for a preview, omitting unset ones.
func hotPlugPreviewFields(prefix string, f tools.HotPlugFlags) []tools.KV {
	return tools.Fields(
		prefix+"cpu_hot_plug", tools.OptBool(f.CpuHotPlug),
		prefix+"ram_hot_plug", tools.OptBool(f.RamHotPlug),
		prefix+"nic_hot_plug", tools.OptBool(f.NicHotPlug),
		prefix+"nic_hot_unplug", tools.OptBool(f.NicHotUnplug),
		prefix+"disc_virtio_hot_plug", tools.OptBool(f.DiscVirtioHotPlug),
		prefix+"disc_virtio_hot_unplug", tools.OptBool(f.DiscVirtioHotUnplug),
	)
}

// sshKeySummary reports how many SSH keys were supplied rather than listing them,
// which keeps a long key list out of the preview.
func sshKeySummary(keys []string) string {
	if len(keys) == 0 {
		return ""
	}
	return fmt.Sprintf("%d key(s) supplied", len(keys))
}
