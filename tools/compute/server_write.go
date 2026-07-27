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

// RegisterServerWriteTools registers the create/update/delete server tools. Each
// is gated by scope inside tools.RegisterTool (create/update need write, delete
// needs destructive); create and delete share the confirmation store so their
// two-phase preview->execute flow works identically across load modes.
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
		// Catch the mistakes the API can only report after a round trip, with
		// messages that name the field to fix.
		sized := input.Cores != nil && input.Ram != nil
		templated := input.TemplateUuid != nil && *input.TemplateUuid != ""
		serverType := strings.ToUpper(strings.TrimSpace(tools.OptStr(input.Type)))
		if !sized && !templated {
			return tools.ErrorText("a server needs a size: provide both cores and ram (for ENTERPRISE or VCPU servers), or template_uuid plus type (for CUBE or GPU servers, see list_templates)"), nil, nil
		}
		if sized && templated {
			return tools.ErrorText("provide either cores and ram, or template_uuid, but not both: CUBE and GPU servers take their size from the template, ENTERPRISE and VCPU servers from cores and ram"), nil, nil
		}
		// type is what distinguishes a CUBE template from a GPU one. Both then
		// require an inline boot volume, but with different storage-type rules, so
		// it cannot be left to be inferred.
		if templated && serverType == "" {
			return tools.ErrorText("type is required when template_uuid is set: pass CUBE or GPU so the request can be validated (both need an inline boot_volume, with type DAS for CUBE)"), nil, nil
		}
		if msg := validateBootVolume(serverType, input.BootVolume); msg != "" {
			return tools.ErrorText(msg), nil, nil
		}
		// The boot volume's storage type is part of the target so a token previewed
		// "with a disk" cannot execute as "with no disk at all", which would
		// silently produce an unbootable server — and for CUBE and GPU, one the API
		// rejects outright.
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
			// The boot volume rides along in entities.volumes, which is what makes
			// this a composite create. Server.Entities is nil-guarded, so a server
			// without a boot volume sends no entities key at all.
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
			// Advisory only — see bootVolumeWarnings for why these do not block.
			// Putting them in the headline means they are read before the fields,
			// while there is still a decision to make.
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
		// A zero-valued literal rather than NewServerPropertiesWithDefaults():
		// that constructor injects nothing today, but a PATCH must send only what
		// the caller asked for, and this cannot silently start applying defaults
		// if a future SDK bump adds one. See the "PATCH bodies" note in CLAUDE.md.
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
			// The boot device is a reference to an already-attached volume, which is
			// the API's only way to point a server at a different disk. Attaching a
			// volume does not make it bootable, and detaching the previous boot
			// volume clears this outright.
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
		// The delete_volumes choice is part of the target so a token previewed
		// as "keep the volumes" can never be replayed as "destroy them".
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

// dasVolumeType is the storage type of a CUBE server's Direct Attached Storage.
// The API accepts it only inside a composite server-creation request.
const dasVolumeType = "DAS"

// templateSizedTypes are the server types whose size comes from template_uuid.
// Both must be created with their boot volume in the same request, and neither
// accepts a boot_volume.size — the template fixes it.
//
// This mirrors what the team's other tools do against the real API: the Terraform
// provider marks the inline volume Required for both resource_cube_server and
// resource_gpu_server (and Optional for the ENTERPRISE/VCPU resource), and neither
// of those two exposes a volume size field at all. ionosctl likewise builds the
// volume into entities for CUBE and GPU, and sets an explicit DAS type only for
// CUBE — for GPU it sends no storage type and lets the API choose, which is why
// boot_volume.type is optional for GPU here.
func isTemplateSized(serverType string) bool {
	return serverType == "CUBE" || serverType == "GPU"
}

// validateBootVolume returns an error message, or "" when the combination is
// valid. It exists because these rules are only reported by the API after a round
// trip, as a generic rejection that does not say which field to fix — and for CUBE
// and GPU the caller cannot recover by attaching a volume afterwards, so a clear
// up-front message is the difference between a fixable mistake and a dead end.
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

	// CUBE's storage type is pinned by documentation: "If you want to create a
	// CUBE server, the type of the inline volume must be set to DAS. In this case,
	// you can not set the size argument since it is taken from the template_uuid"
	// (Terraform provider docs/resources/volume.md).
	switch {
	case isCube && volType == "":
		return "a CUBE server's boot_volume.type must be set to DAS: CUBE storage is Direct Attached Storage"
	case isCube && !isDAS:
		return fmt.Sprintf("a CUBE server's boot_volume.type must be DAS, not %q: CUBE storage is Direct Attached Storage and the API accepts no other type for it", volType)
	case templateSized && bv.Size != nil:
		return fmt.Sprintf("boot_volume.size must be omitted for a %s server: its storage size is fixed by template_uuid", serverType)
	}

	// Without an image or a licence type the volume has no operating system, so the
	// server cannot boot from it. Both of the team's other tools always send one of
	// the three.
	hasImage := (bv.Image != nil && *bv.Image != "") || (bv.ImageAlias != nil && *bv.ImageAlias != "")
	hasLicence := bv.LicenceType != nil && *bv.LicenceType != ""
	if !hasImage && !hasLicence {
		return "boot_volume needs image or image_alias to install an operating system (see list_images), or licence_type for an empty disk"
	}
	return ""
}

// bootVolumeWarnings returns advisory notes for combinations that look wrong but
// are NOT rejected here.
//
// The distinction from validateBootVolume is deliberate. A hard error is right
// when the rule is documented AND blocking replaces an opaque API rejection with
// a precise instruction — CUBE's DAS type, and the missing inline volume that
// makes a CUBE or GPU server uncreatable. It is wrong when the rule is inferred
// rather than documented, because a mistaken block breaks a valid request with no
// way around it, whereas an unwanted storage type is trivially recoverable: the
// API rejects it and the caller retries with a different value.
//
// These three are inferred, so they warn:
//   - DAS on a non-template-sized server. Every DAS example in the IONOS docs is a
//     CUBE server and the spec says DAS "could be used only in a composite call
//     with a Cube server", which implies but does not state that ENTERPRISE
//     refuses it.
//   - A missing size on a non-template-sized server. The Terraform provider marks
//     the inline volume's size Optional+Computed, so it evidently does not always
//     have to be supplied.
//   - A missing storage type on a non-template-sized server. The Terraform
//     provider marks disk_type Required, but that may be its own UX choice rather
//     than an API constraint.
//
// The warnings surface in the two-phase preview, so the model sees them before
// committing — which is what the preview is for.
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

// bootVolumeTargetPart renders the boot volume's contribution to the confirmation
// target. Only the storage type is included: enough to stop a token previewed
// with a disk from executing without one (which would silently produce an
// unbootable server), without making the token brittle to cosmetic changes.
func bootVolumeTargetPart(bv *tools.BootVolumeInput) string {
	if bv == nil {
		return "no-boot-volume"
	}
	volType := strings.ToUpper(strings.TrimSpace(tools.OptStr(bv.Type)))
	if volType == "" {
		// A template-sized server may leave the type to the API; the marker still
		// records that a boot volume was previewed.
		volType = "api-default"
	}
	return "boot-volume:" + volType
}

// buildBootVolume converts the input into an SDK Volume for the composite create,
// or nil when no boot volume was requested. Only the fields the caller supplied
// are set, for the same reason update handlers avoid the WithDefaults
// constructors — see the "PATCH bodies" note in CLAUDE.md.
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
	vol := ionos.NewVolume()
	vol.SetProperties(*props)
	return vol
}

// bootVolumePreviewFields renders the boot volume for the create preview. Secrets
// are acknowledged rather than echoed, since previews are shown to the model and
// logged by clients.
func bootVolumePreviewFields(serverType string, bv *tools.BootVolumeInput) []tools.KV {
	size := tools.OptFloat32(bv.Size)
	if size == "" {
		// Only a template-sized server takes its size from the template; saying so
		// for an ENTERPRISE server would be plainly wrong.
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
	return tools.Fields(
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
}

// firstNonEmpty returns the first non-empty string, used to keep error messages
// readable when an optional field was not supplied.
func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}

// serverBlastRadius counts what a server delete affects, from a server fetched at
// depth 2. Volumes are only listed as destroyed when delete_volumes is set —
// otherwise they survive the delete, which volumeFateNote spells out instead. It
// also returns the attached volume count so the caller can describe their fate.
func serverBlastRadius(srv ionos.Server, deleteVolumes bool) (*tools.BlastRadius, int) {
	r := &tools.BlastRadius{}
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

// volumeFateNote states what happens to the attached volumes, because the
// default (keep them, keep billing) is the surprising one.
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
