package compute

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

// RegisterServerWriteTools registers the create/update/delete server tools.
func RegisterServerWriteTools(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	registerCreateServer(server, client, scope, confirm)
	registerUpdateServer(server, client, scope)
	registerDeleteServer(server, client, scope, confirm)
}

func registerCreateServer(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterTool(server, scope, tools.MethodPost, &mcp.Tool{
		Name: "create_server",
		Description: "Create one server (virtual machine) in a data center. Creates exactly one server per call. Two-phase: call first WITHOUT confirmation_token to get a preview and a one-time token, then call again WITH the token (and the same datacenter_id, name and boot_volume) to create it. " +
			"Size it either with cores and ram (ENTERPRISE, the default, or VCPU) or with template_uuid plus type (CUBE or GPU) — not both. " +
			"boot_volume optionally creates the server's disk in the same request; for an ENTERPRISE or VCPU server you can just as well omit it and add a disk later with create_volume plus attach_server_volume, so do not ask the caller which they prefer. " +
			"It is required only for CUBE and GPU servers, which are template-sized: the API accepts their storage only as part of the create, so it rejects one made without it and attaching a volume afterwards does not work. Set boot_volume.type to DAS for CUBE; for GPU leave the type unset (the API chooses) or use SSD Premium. Neither takes a boot_volume.size — the template fixes it. " +
			"It is also required when creating a Confidential Computing server, because the API derives its cores and CPU family from the confidential image on that volume.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.CreateServerInput) (*mcp.CallToolResult, any, error) {
		dcID := strings.TrimSpace(input.DatacenterID)
		name := strings.TrimSpace(input.Name)
		if dcID == "" {
			return tools.ErrorText("datacenter_id is required to create a server"), nil, nil
		}
		if name == "" {
			return tools.ErrorText("name is required to create a server"), nil, nil
		}
		// The API reports these only after a round trip, without naming the field.
		sized := input.Cores != nil && input.Ram != nil
		templated := input.TemplateUuid != nil && *input.TemplateUuid != ""
		serverType := strings.ToUpper(strings.TrimSpace(tools.OptStr(input.Type)))
		if !sized && !templated {
			return tools.ErrorText("a server needs a size: provide both cores and ram (for ENTERPRISE or VCPU servers), or template_uuid plus type (for CUBE or GPU servers, see list_templates)"), nil, nil
		}
		if sized && templated {
			return tools.ErrorText("provide either cores and ram, or template_uuid, but not both: CUBE and GPU servers take their size from the template, ENTERPRISE and VCPU servers from cores and ram"), nil, nil
		}
		// type tells a CUBE template from a GPU one; their storage rules differ.
		if templated && serverType == "" {
			return tools.ErrorText("type is required when template_uuid is set: pass CUBE or GPU so the request can be validated (both need an inline boot_volume, with type DAS for CUBE)"), nil, nil
		}
		if msg := validateBootVolume(serverType, input.BootVolume); msg != "" {
			return tools.ErrorText(msg), nil, nil
		}
		// The storage type is in the target, so a token previewed with a disk
		// cannot execute without one.
		target := tools.Target(dcID, name, bootVolumeTargetPart(input.BootVolume))

		// Phase 2: token present -> validate and execute.
		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "create_server", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("create_server", "datacenter_id, name and the same boot_volume", err)), nil, nil
			}
			props := ionos.NewServerPropertiesWithDefaults()
			props.SetName(name)
			if input.Cores != nil {
				props.SetCores(*input.Cores)
			}
			if input.Ram != nil {
				props.SetRam(*input.Ram)
			}
			if input.Type != nil {
				props.SetType(*input.Type)
			}
			if input.TemplateUuid != nil {
				props.SetTemplateUuid(*input.TemplateUuid)
			}
			if input.CpuFamily != nil {
				props.SetCpuFamily(*input.CpuFamily)
			}
			if input.AvailabilityZone != nil {
				props.SetAvailabilityZone(*input.AvailabilityZone)
			}
			if input.Hostname != nil {
				props.SetHostname(*input.Hostname)
			}
			if input.NicMultiQueue != nil {
				props.SetNicMultiQueue(*input.NicMultiQueue)
			}
			body := ionos.NewServer(*props)
			// The boot volume rides in entities.volumes, making this a composite
			// create. Omitted entirely when there is no boot volume.
			if bv := buildBootVolume(input.BootVolume); bv != nil {
				body.SetEntities(ionos.ServerEntities{
					Volumes: &ionos.AttachedVolumes{Items: []ionos.Volume{*bv}},
				})
			}
			created, _, err := client.ServersApi.DatacentersServersPost(ctx, dcID).Server(*body).Execute()
			return tools.ToResult(created, err)
		}

		// Phase 1: no token -> preview and mint a one-time token.
		token, err := confirm.Mint("create_server", target)
		if err != nil {
			return nil, nil, err
		}
		fields := tools.Fields(
			"datacenter_id", dcID,
			"name", name,
			"cores", tools.OptInt32(input.Cores),
			"ram (MB)", tools.OptInt32(input.Ram),
			"type", tools.OptStr(input.Type),
			"template_uuid", tools.OptStr(input.TemplateUuid),
			"cpu_family", tools.OptStr(input.CpuFamily),
			"availability_zone", tools.OptStr(input.AvailabilityZone),
			"hostname", tools.OptStr(input.Hostname),
			"nic_multi_queue", tools.OptBool(input.NicMultiQueue),
		)
		headline := "About to CREATE one server:"
		note := "This creates exactly one server. "
		if bv := input.BootVolume; bv != nil {
			fields = append(fields, bootVolumePreviewFields(serverType, bv)...)
			note += "The boot volume listed above is created in the same request. "
			// Advisory only. In the headline so they are read before the fields.
			for _, w := range bootVolumeWarnings(serverType, bv) {
				headline += "\nNOTE: " + w
			}
		} else {
			note += "No boot_volume was given, so the server starts with no disk and no network: add storage with create_volume plus attach_server_volume, and networking with create_nic, before it can boot or be reached. "
		}
		return tools.TextResult(tools.Preview{
			Headline:  headline,
			Fields:    fields,
			Tool:      "create_server",
			Replay:    tools.Fields("datacenter_id", dcID, "name", name),
			TokenNote: note + "The token authorizes creating only this server, with this boot volume, in this data center",
		}.Render(token)), nil, nil
	})
}

func registerUpdateServer(server *mcp.Server, client *ionos.APIClient, scope tools.Scope) {
	tools.RegisterTool(server, scope, tools.MethodPatch, &mcp.Tool{
		Name: "update_server",
		Description: "Update a server's name, size (cores, ram), CPU family, hostname, Multi Queue setting, or which volume it boots from. Applies a partial update (only the fields you provide). " +
			"boot_volume_id is how you point a server at a different disk: attaching a volume with attach_server_volume does NOT make it bootable, and detaching the previous boot volume clears the setting, leaving a server that cannot boot until you set it again. The volume must already be attached. " +
			"Resizing may require the server to be rebooted before the guest OS sees the change, and toggling nic_multi_queue restarts it. " +
			"On CUBE and GPU servers the size comes from the template and cannot be changed here; on servers with Confidential Computing enabled cores, ram and cpu_family are immutable.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.UpdateServerInput) (*mcp.CallToolResult, any, error) {
		dcID := strings.TrimSpace(input.DatacenterID)
		id := strings.TrimSpace(input.ServerID)
		if dcID == "" {
			return tools.ErrorText("datacenter_id is required"), nil, nil
		}
		if id == "" {
			return tools.ErrorText("server_id is required"), nil, nil
		}
		if input.Name == nil && input.Cores == nil && input.Ram == nil &&
			input.CpuFamily == nil && input.Hostname == nil && input.NicMultiQueue == nil &&
			input.BootVolumeID == nil {
			return tools.ErrorText("nothing to update: provide at least one of name, cores, ram, cpu_family, hostname, nic_multi_queue, boot_volume_id"), nil, nil
		}
		// A literal, not a generated constructor: a PATCH must carry only the
		// fields the caller supplied.
		props := &ionos.ServerProperties{}
		if input.Name != nil {
			props.SetName(*input.Name)
		}
		if input.Cores != nil {
			props.SetCores(*input.Cores)
		}
		if input.Ram != nil {
			props.SetRam(*input.Ram)
		}
		if input.CpuFamily != nil {
			props.SetCpuFamily(*input.CpuFamily)
		}
		if input.Hostname != nil {
			props.SetHostname(*input.Hostname)
		}
		if input.NicMultiQueue != nil {
			props.SetNicMultiQueue(*input.NicMultiQueue)
		}
		if input.BootVolumeID != nil {
			bootVolID := strings.TrimSpace(*input.BootVolumeID)
			if bootVolID == "" {
				return tools.ErrorText("boot_volume_id must be a volume ID; omit the field entirely to leave the boot device unchanged"), nil, nil
			}
			// The only way to change a server's boot disk. The volume must already
			// be attached; attaching alone does not make it bootable.
			props.SetBootVolume(*ionos.NewResourceReference(bootVolID))
		}
		updated, _, err := client.ServersApi.DatacentersServersPatch(ctx, dcID, id).Server(*props).Execute()
		return tools.ToResult(updated, err)
	})
}

func registerDeleteServer(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterTool(server, scope, tools.MethodDelete, &mcp.Tool{
		Name: "delete_server",
		Description: "Delete a server. Two-phase: call first WITHOUT confirmation_token to get a preview of what is attached and what happens to it, plus a one-time token, then call again WITH the token to delete. " +
			"By default the attached volumes are NOT deleted — they survive as unattached volumes and keep incurring cost. Set delete_volumes to true to destroy them with the server (their data is then unrecoverable). " +
			"Because that choice changes what is destroyed, the token is bound to it: preview again if you change delete_volumes. This is irreversible.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.DeleteServerInput) (*mcp.CallToolResult, any, error) {
		dcID := strings.TrimSpace(input.DatacenterID)
		id := strings.TrimSpace(input.ServerID)
		if dcID == "" {
			return tools.ErrorText("datacenter_id is required"), nil, nil
		}
		if id == "" {
			return tools.ErrorText("server_id is required"), nil, nil
		}
		deleteVolumes := input.DeleteVolumes != nil && *input.DeleteVolumes
		// delete_volumes is in the target, so a token previewed as "keep" cannot
		// execute as "destroy".
		target := tools.Target(dcID, id, strconv.FormatBool(deleteVolumes))

		// Phase 2: token present -> validate and execute.
		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "delete_server", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("delete_server", "datacenter_id, server_id and the same delete_volumes value", err)), nil, nil
			}
			_, err := client.ServersApi.DatacentersServersDelete(ctx, dcID, id).DeleteVolumes(deleteVolumes).Execute()
			if err != nil {
				return tools.ToResult(nil, err)
			}
			msg := tools.DeletedAsync("server", id)
			if deleteVolumes {
				msg += " Its attached volumes were deleted with it."
			} else {
				msg += " Its attached volumes were kept and are now unattached; delete them with delete_volume if they are no longer needed."
			}
			return tools.TextResult(msg), nil, nil
		}

		// Phase 1: no token -> report what is attached, then mint a token.
		srv, _, err := client.ServersApi.DatacentersServersFindById(ctx, dcID, id).Depth(2).Execute()
		if err != nil {
			if tools.IsNotFound(err) {
				return tools.ErrorText(fmt.Sprintf("server %s does not exist in data center %s; nothing to delete", id, dcID)), nil, nil
			}
			return tools.ToResult(nil, err)
		}
		radius, volumeCount := serverBlastRadius(srv, deleteVolumes)
		token, mErr := confirm.Mint("delete_server", target)
		if mErr != nil {
			return nil, nil, mErr
		}
		props := srv.GetProperties()
		return tools.TextResult(tools.Preview{
			Headline: "About to DELETE a server. This is IRREVERSIBLE.",
			Fields: tools.Fields(
				"datacenter_id", dcID,
				"server_id", id,
				"name", props.GetName(),
				"type", props.GetType(),
				"vm_state", props.GetVmState(),
				"delete_volumes", strconv.FormatBool(deleteVolumes),
			),
			Radius:    radius,
			EmptyNote: "This server has nothing attached; deleting removes only the server itself.",
			Tool:      "delete_server",
			Replay:    tools.Fields("datacenter_id", dcID, "server_id", id, "delete_volumes", strconv.FormatBool(deleteVolumes)),
			TokenNote: volumeFateNote(deleteVolumes, volumeCount) + " This token authorizes deleting ONLY this server with this delete_volumes value",
		}.Render(token)), nil, nil
	})
}

// dasVolumeType is a CUBE server's Direct Attached Storage. Accepted only inside
// a composite server-creation request.
const dasVolumeType = "DAS"

// templateSizedTypes take their size from template_uuid. Both must be created with
// their boot volume in the same request and neither accepts a boot_volume.size.
// GPU leaves the storage type to the API; CUBE requires DAS.
func isTemplateSized(serverType string) bool {
	return serverType == "CUBE" || serverType == "GPU"
}

// validateBootVolume returns an error message, or "" if the combination is valid.
// A CUBE or GPU server created without its inline volume cannot be repaired by
// attaching one afterwards, so these are caught before the request.
func validateBootVolume(serverType string, bv *tools.BootVolumeInput) string {
	isCube := serverType == "CUBE"
	templateSized := isTemplateSized(serverType)

	if bv == nil {
		if templateSized {
			what := "with type DAS"
			if !isCube {
				what = "leaving type unset, or set it to SSD Premium"
			}
			return fmt.Sprintf("a %s server must be created together with its storage: pass boot_volume %s (leave size out, the template fixes it). "+
				"The API rejects a %s server created without a volume in the same request, and attaching one afterwards does not work.",
				serverType, what, serverType)
		}
		return ""
	}

	volType := strings.TrimSpace(tools.OptStr(bv.Type))
	isDAS := strings.EqualFold(volType, dasVolumeType)

	// Documented: a CUBE server's inline volume must be DAS, with no size.
	switch {
	case isCube && volType == "":
		return "a CUBE server's boot_volume.type must be set to DAS: CUBE storage is Direct Attached Storage"
	case isCube && !isDAS:
		return fmt.Sprintf("a CUBE server's boot_volume.type must be DAS, not %q: CUBE storage is Direct Attached Storage and the API accepts no other type for it", volType)
	case templateSized && bv.Size != nil:
		return fmt.Sprintf("boot_volume.size must be omitted for a %s server: its storage size is fixed by template_uuid", serverType)
	}

	// Without an image or licence type the volume has no OS to boot.
	hasImage := (bv.Image != nil && *bv.Image != "") || (bv.ImageAlias != nil && *bv.ImageAlias != "")
	hasLicence := bv.LicenceType != nil && *bv.LicenceType != ""
	if !hasImage && !hasLicence {
		return "boot_volume needs image or image_alias to install an operating system (see list_images), or licence_type for an empty disk"
	}
	return ""
}

// bootVolumeWarnings returns advisory notes for combinations that look wrong but are
// not rejected. These rules are inferred rather than documented, and a wrong storage
// type is recoverable by retrying, so they warn instead of blocking.
func bootVolumeWarnings(serverType string, bv *tools.BootVolumeInput) []string {
	if bv == nil || isTemplateSized(serverType) {
		return nil
	}
	shown := firstNonEmpty(serverType, "ENTERPRISE (the default)")
	volType := strings.TrimSpace(tools.OptStr(bv.Type))

	var out []string
	if strings.EqualFold(volType, dasVolumeType) {
		out = append(out, fmt.Sprintf("boot_volume.type is DAS on a server of type %s. DAS storage is documented for template-sized CUBE servers, so the API may reject this; HDD, SSD, SSD Standard and SSD Premium are the usual types here.", shown))
	}
	if volType == "" {
		out = append(out, fmt.Sprintf("boot_volume.type is not set on a server of type %s. Only template-sized CUBE and GPU servers are known to let the API choose, so consider naming a type (HDD, SSD, SSD Standard, SSD Premium).", shown))
	}
	if bv.Size == nil || *bv.Size <= 0 {
		out = append(out, fmt.Sprintf("boot_volume.size is not set on a server of type %s. Only template-sized CUBE and GPU servers take their size from the template, so the API may require one here.", shown))
	}
	return out
}

// bootVolumeTargetPart renders the boot volume's part of the confirmation target.
// Only the storage type, so the token is not brittle to cosmetic changes.
func bootVolumeTargetPart(bv *tools.BootVolumeInput) string {
	if bv == nil {
		return "no-boot-volume"
	}
	volType := strings.ToUpper(strings.TrimSpace(tools.OptStr(bv.Type)))
	if volType == "" {
		// The type may be left to the API; the marker still records a disk.
		volType = "api-default"
	}
	return "boot-volume:" + volType
}

// buildBootVolume converts the input into a Volume for the composite create, or
// nil when none was requested. Sets only the fields the caller supplied.
func buildBootVolume(bv *tools.BootVolumeInput) *ionos.Volume {
	if bv == nil {
		return nil
	}
	props := &ionos.VolumeProperties{}
	if volType := strings.TrimSpace(tools.OptStr(bv.Type)); volType != "" {
		props.SetType(volType)
	}
	if bv.Name != nil {
		props.SetName(*bv.Name)
	}
	if bv.Size != nil {
		props.SetSize(*bv.Size)
	}
	if bv.Image != nil {
		props.SetImage(*bv.Image)
	}
	if bv.ImageAlias != nil {
		props.SetImageAlias(*bv.ImageAlias)
	}
	if bv.ImagePassword != nil {
		props.SetImagePassword(*bv.ImagePassword)
	}
	if len(bv.SshKeys) > 0 {
		props.SetSshKeys(bv.SshKeys)
	}
	if bv.LicenceType != nil {
		props.SetLicenceType(*bv.LicenceType)
	}
	if bv.Bus != nil {
		props.SetBus(*bv.Bus)
	}
	if bv.UserData != nil {
		props.SetUserData(*bv.UserData)
	}
	applyHotPlugFlags(props, bv.HotPlugFlags)
	vol := ionos.NewVolume()
	vol.SetProperties(*props)
	return vol
}

// bootVolumePreviewFields renders the boot volume for the create preview.
// Secrets are acknowledged, never echoed — clients log previews.
func bootVolumePreviewFields(serverType string, bv *tools.BootVolumeInput) []tools.KV {
	size := tools.OptFloat32(bv.Size)
	if size == "" {
		// Only template-sized servers take their size from the template.
		if isTemplateSized(serverType) {
			size = "fixed by the template"
		} else {
			size = "not set"
		}
	}
	volType := tools.OptStr(bv.Type)
	if strings.TrimSpace(volType) == "" {
		if isTemplateSized(serverType) {
			volType = "(chosen by the API)"
		} else {
			volType = "(not set)"
		}
	}
	out := tools.Fields(
		"boot_volume.type", volType,
		"boot_volume.name", tools.OptStr(bv.Name),
		"boot_volume.size (GB)", size,
		"boot_volume.image", tools.OptStr(bv.Image),
		"boot_volume.image_alias", tools.OptStr(bv.ImageAlias),
		"boot_volume.image_password", redacted(bv.ImagePassword),
		"boot_volume.ssh_keys", sshKeySummary(bv.SshKeys),
		"boot_volume.licence_type", tools.OptStr(bv.LicenceType),
		"boot_volume.bus", tools.OptStr(bv.Bus),
		"boot_volume.user_data", redacted(bv.UserData),
	)
	return append(out, hotPlugPreviewFields("boot_volume.", bv.HotPlugFlags)...)
}

// firstNonEmpty returns the first non-empty string.
func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}

// serverBlastRadius counts what a delete affects, from a server fetched at depth 2.
// Volumes count as destroyed only when delete_volumes is set. Also returns the
// attached volume count.
func serverBlastRadius(srv ionos.Server, deleteVolumes bool) (*tools.BlastRadius, int) {
	r := tools.DestroyedRadius()
	e := srv.Entities
	if e == nil {
		return r, 0
	}
	volumeCount := 0
	if e.Volumes != nil {
		volumeCount = len(e.Volumes.Items)
	}
	if deleteVolumes {
		r.Add("attached volumes (data destroyed)", volumeCount)
	}
	if e.Nics != nil {
		r.Add("NICs (with their firewall rules and flow logs)", len(e.Nics.Items))
	}
	if e.Cdroms != nil {
		r.Add("attached CD-ROMs (detached, images not deleted)", len(e.Cdroms.Items))
	}
	return r, volumeCount
}

// volumeFateNote states what happens to the attached volumes. The default —
// keep them, keep billing — is the surprising one.
func volumeFateNote(deleteVolumes bool, volumeCount int) string {
	switch {
	case volumeCount == 0:
		return "No volumes are attached."
	case deleteVolumes:
		return fmt.Sprintf("delete_volumes is TRUE: the %d attached volume(s) will be DESTROYED with the server and their data cannot be recovered.", volumeCount)
	default:
		return fmt.Sprintf("delete_volumes is FALSE: the %d attached volume(s) will SURVIVE as unattached volumes and keep incurring cost until deleted separately.", volumeCount)
	}
}
