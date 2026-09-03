package tools

// Input types for tool parameters.
// Each struct maps to the JSON schema that the MCP SDK auto-generates
// for a tool's input. Non-pointer fields are treated as required.

// Compute input types

// ListDatacentersInput is the input for list_datacenters (no required ID fields).
type ListDatacentersInput struct {
	Depth   *int32            `json:"depth,omitempty"   jsonschema:"nesting depth of returned objects (0-5, default 1); depth 1 includes names and basic properties"`
	Filters map[string]string `json:"filters,omitempty" jsonschema:"server-side filters as property name to value pairs (contains match); e.g. {\"name\":\"prod\",\"location\":\"de/fra\"}. Filterable properties include: name, description, location, version. If the result is empty, retry without filters — a filter typo or mismatch silently returns nothing."`
}

type DatacenterIDInput struct {
	DatacenterID string `json:"datacenter_id" jsonschema:"the ID of the data center"`
	Depth        *int32 `json:"depth,omitempty" jsonschema:"nesting depth of returned objects (0-5, default 1 for list operations)"`
}

// CreateDatacenterInput is the input for create_datacenter. Two-phase confirmed;
// creates exactly one data center per call (no batch/count field).
type CreateDatacenterInput struct {
	Name              string  `json:"name" jsonschema:"the name of the new data center"`
	Location          string  `json:"location" jsonschema:"the physical location where the data center will be created, e.g. de/fra, de/txl, us/las, us/ewr, gb/lhr, es/vit, fr/par. Cannot be changed after creation."`
	Description       *string `json:"description,omitempty" jsonschema:"an optional description, such as staging or production"`
	SecAuthProtection *bool   `json:"sec_auth_protection,omitempty" jsonschema:"if true, the data center requires extra protection such as two-step verification"`
	ConfirmationToken *string `json:"confirmation_token,omitempty" jsonschema:"leave empty on the FIRST call to receive a preview of what will be created plus a one-time token; pass that token on the SECOND call (with the same name and location) to actually create the data center. The token expires after a few minutes."`
}

// UpdateDatacenterInput is the input for update_datacenter. Partial update;
// location is immutable and cannot be set here.
type UpdateDatacenterInput struct {
	DatacenterID      string  `json:"datacenter_id" jsonschema:"the ID of the data center to update"`
	Name              *string `json:"name,omitempty" jsonschema:"a new name for the data center"`
	Description       *string `json:"description,omitempty" jsonschema:"a new description for the data center"`
	SecAuthProtection *bool   `json:"sec_auth_protection,omitempty" jsonschema:"set the extra-protection (two-step verification) flag"`
}

// DeleteDatacenterInput is the input for delete_datacenter. Two-phase confirmed:
// deleting removes the data center and every resource inside it.
type DeleteDatacenterInput struct {
	DatacenterID      string  `json:"datacenter_id" jsonschema:"the ID of the data center to delete"`
	ConfirmationToken *string `json:"confirmation_token,omitempty" jsonschema:"leave empty on the FIRST call to receive a blast-radius preview (what will be destroyed) plus a one-time token; pass that token on the SECOND call to actually delete. The token authorizes deleting only the data center it was issued for and expires after a few minutes."`
}

type ServerIDInput struct {
	DatacenterID string `json:"datacenter_id" jsonschema:"the ID of the data center"`
	ServerID     string `json:"server_id" jsonschema:"the ID of the server"`
	Depth        *int32 `json:"depth,omitempty" jsonschema:"nesting depth of returned objects (0-5, default 1 for list operations)"`
}

// HotPlugFlags are the volume capability flags that decide whether the server it
// is attached to can change hardware without a reboot. They are embedded in every
// input that carries volume properties — create_volume, update_volume and the
// inline boot volume of create_server — because all three IONOS tooling paths
// expose them on all three (ionosctl's volume create and update, and the volume
// block of the Terraform server, cube_server and gpu_server resources).
//
// Embedded rather than repeated so the wording cannot drift between the three
// tools; Go flattens the fields into the parent object's JSON schema.
type HotPlugFlags struct {
	CpuHotPlug          *bool `json:"cpu_hot_plug,omitempty" jsonschema:"whether CPU cores can be added to the server without rebooting it. Set this before you need to resize a running server, since turning it on later is itself only picked up on a restart."`
	RamHotPlug          *bool `json:"ram_hot_plug,omitempty" jsonschema:"whether memory can be added to the server without rebooting it. A volume with RAM hot-plug enabled requires the server to have at least 1024 MB of RAM."`
	NicHotPlug          *bool `json:"nic_hot_plug,omitempty" jsonschema:"whether a NIC can be attached to the server without rebooting it"`
	NicHotUnplug        *bool `json:"nic_hot_unplug,omitempty" jsonschema:"whether a NIC can be detached from the server without rebooting it"`
	DiscVirtioHotPlug   *bool `json:"disc_virtio_hot_plug,omitempty" jsonschema:"whether a VirtIO disk can be attached to the server without rebooting it"`
	DiscVirtioHotUnplug *bool `json:"disc_virtio_hot_unplug,omitempty" jsonschema:"whether a VirtIO disk can be detached from the server without rebooting it. Not supported on Windows guests."`
}

// BootVolumeInput describes a volume to create together with a server, in the
// same API request. This is not merely a convenience: CUBE and GPU servers are
// template-sized and the API accepts their storage ONLY as part of a composite
// server-creation call, so neither can be created without it and attaching a
// volume afterwards is not equivalent. A Confidential Computing server likewise
// must be created with its boot volume inline, because the API derives the core
// count and CPU family from the confidential image on that volume.
type BootVolumeInput struct {
	Type          *string  `json:"type,omitempty" jsonschema:"storage type of the boot volume. Must be DAS for a CUBE server, whose storage exists only as part of the server. Required and must NOT be DAS for ENTERPRISE and VCPU servers: use HDD, SSD, SSD Standard or SSD Premium. Optional for a GPU server — omit it to let the API choose, or pass SSD Premium."`
	Name          *string  `json:"name,omitempty" jsonschema:"the name of the boot volume"`
	Size          *float32 `json:"size,omitempty" jsonschema:"size in GB. Required for ENTERPRISE and VCPU servers. Must be OMITTED for CUBE and GPU servers, whose storage size is fixed by template_uuid."`
	Image         *string  `json:"image,omitempty" jsonschema:"ID of an image or snapshot to install. Provide exactly one of image, image_alias or licence_type; without one of the first two the volume has no operating system. See list_images."`
	ImageAlias    *string  `json:"image_alias,omitempty" jsonschema:"alias of an image to install, e.g. ubuntu:latest. An alternative to image."`
	ImagePassword *string  `json:"image_password,omitempty" jsonschema:"initial root/administrator password for the installed OS; public images only. Characters a-z, A-Z, 0-9, minimum 8. Cannot be changed later. Prefer ssh_keys for Linux images."`
	SshKeys       []string `json:"ssh_keys,omitempty" jsonschema:"public SSH keys to authorize for login. Public Linux images only."`
	LicenceType   *string  `json:"licence_type,omitempty" jsonschema:"OS type: LINUX, WINDOWS, WINDOWS2016, WINDOWS2022, WINDOWS2025, UNKNOWN or OTHER. Required when neither image nor image_alias is given."`
	Bus           *string  `json:"bus,omitempty" jsonschema:"bus type: VIRTIO (default, faster) or IDE"`
	UserData      *string  `json:"user_data,omitempty" jsonschema:"cloud-init configuration as a base64-encoded string; requires a cloud-init-capable image"`
	HotPlugFlags
}

// CreateServerInput is the input for create_server. Two-phase confirmed; creates
// exactly one server per call.
//
// boot_volume creates a volume in the same request. It is REQUIRED for the
// template-sized types, CUBE and GPU, whose storage the API accepts only in a
// composite call. Without a boot_volume an ENTERPRISE or VCPU server comes up with
// no storage and no NIC, so it has nothing to boot from and no network — follow up
// with attach_server_volume and create_nic, or supply boot_volume here and save
// the round trips.
type CreateServerInput struct {
	DatacenterID      string           `json:"datacenter_id" jsonschema:"the ID of the data center to create the server in"`
	Name              string           `json:"name" jsonschema:"the name of the new server"`
	Cores             *int32           `json:"cores,omitempty" jsonschema:"total number of CPU cores. Required for ENTERPRISE and VCPU servers; must NOT be set for CUBE or GPU servers, which take their size from template_uuid instead."`
	Ram               *int32           `json:"ram,omitempty" jsonschema:"memory size in MB, in multiples of 256 with a minimum of 256 (use at least 1024 if you enable RAM hot-plug). Required for ENTERPRISE and VCPU servers; must NOT be set for CUBE or GPU servers, which take their size from template_uuid instead."`
	Type              *string          `json:"type,omitempty" jsonschema:"server type: ENTERPRISE (default, cores+ram), VCPU (cores+ram), CUBE (fixed size from template_uuid) or GPU (template_uuid, GPU template only)"`
	TemplateUuid      *string          `json:"template_uuid,omitempty" jsonschema:"the template ID that fixes the size of a CUBE or GPU server; required for those types and forbidden for ENTERPRISE and VCPU. List available templates with list_templates."`
	CpuFamily         *string          `json:"cpu_family,omitempty" jsonschema:"CPU architecture, e.g. INTEL_SKYLAKE or AMD_OPTERON. Must not be set for CUBE or VCPU servers. Omit to have an available family chosen automatically; available families per location come from list_locations. Availability varies by data center region."`
	AvailabilityZone  *string          `json:"availability_zone,omitempty" jsonschema:"availability zone to provision in: AUTO (default), ZONE_1 or ZONE_2. CUBE and GPU servers accept only AUTO."`
	Hostname          *string          `json:"hostname,omitempty" jsonschema:"the hostname of the server; allowed characters are a-z, 0-9 and - (minus), must not start with a minus and must be at most 63 characters"`
	NicMultiQueue     *bool            `json:"nic_multi_queue,omitempty" jsonschema:"activate Multi Queue on all NICs of this server, which helps when NICs show low throughput. Not allowed for CUBE servers. Toggling this restarts the server."`
	BootVolume        *BootVolumeInput `json:"boot_volume,omitempty" jsonschema:"optional for ENTERPRISE and VCPU servers (the default type): supply it to create the server's disk in the same request, or omit it and add a disk later with create_volume plus attach_server_volume. Do not ask which the caller wants — omitting it is fine. Required only in three cases: a CUBE server, a GPU server (both are template-sized, and the API accepts their storage only as part of the create, so it rejects one made without it and attaching a volume afterwards does not work), and a Confidential Computing server (the API derives its cores and CPU family from the confidential image on this volume)."`
	ConfirmationToken *string          `json:"confirmation_token,omitempty" jsonschema:"leave empty on the FIRST call to receive a preview of what will be created plus a one-time token; pass that token on the SECOND call (with the same datacenter_id, name and boot_volume) to actually create the server. The token expires after a few minutes."`
}

// UpdateServerInput is the input for update_server. Partial update: only the
// fields you provide are sent. Resizing cores or ram on a running server may
// require a reboot, and on a Confidential Computing server cores, ram,
// cpu_family and availability_zone are immutable and will be rejected.
type UpdateServerInput struct {
	DatacenterID  string  `json:"datacenter_id" jsonschema:"the ID of the data center the server is in"`
	ServerID      string  `json:"server_id" jsonschema:"the ID of the server to update"`
	Name          *string `json:"name,omitempty" jsonschema:"a new name for the server"`
	Cores         *int32  `json:"cores,omitempty" jsonschema:"a new CPU core count. Rejected on CUBE and GPU servers and on servers with Confidential Computing enabled."`
	Ram           *int32  `json:"ram,omitempty" jsonschema:"a new memory size in MB, in multiples of 256. Rejected on CUBE and GPU servers and on servers with Confidential Computing enabled."`
	CpuFamily     *string `json:"cpu_family,omitempty" jsonschema:"a new CPU architecture. Rejected on CUBE and VCPU servers and on servers with Confidential Computing enabled."`
	Hostname      *string `json:"hostname,omitempty" jsonschema:"a new hostname; allowed characters are a-z, 0-9 and - (minus)"`
	NicMultiQueue *bool   `json:"nic_multi_queue,omitempty" jsonschema:"turn Multi Queue on all NICs of this server on or off. Toggling this restarts the server."`
	BootVolumeID  *string `json:"boot_volume_id,omitempty" jsonschema:"set which attached volume the server boots from, by volume ID. This is how you point a server at a different disk: after attach_server_volume, the new volume is attached but NOT the boot device, and detaching the previous boot volume clears the setting entirely, leaving a server that cannot boot until you set this. The volume must already be attached to this server (attach_server_volume first). Reboot the server for the change to take effect."`
}

// DeleteServerInput is the input for delete_server. Two-phase confirmed. Whether
// attached volumes are destroyed with the server depends on delete_volumes, so
// that flag is part of the confirmation target: a token minted for one choice
// cannot be replayed with the other.
type DeleteServerInput struct {
	DatacenterID      string  `json:"datacenter_id" jsonschema:"the ID of the data center the server is in"`
	ServerID          string  `json:"server_id" jsonschema:"the ID of the server to delete"`
	DeleteVolumes     *bool   `json:"delete_volumes,omitempty" jsonschema:"also delete the volumes attached to this server (default false). If false, the volumes survive as unattached volumes in the data center and KEEP INCURRING COST until deleted separately with delete_volume. Deleting the volumes destroys their data irrecoverably."`
	ConfirmationToken *string `json:"confirmation_token,omitempty" jsonschema:"leave empty on the FIRST call to receive a preview of what is attached and what happens to it, plus a one-time token; pass that token on the SECOND call (with the same delete_volumes value) to actually delete. The token authorizes deleting only the server it was issued for, with only the delete_volumes choice it was previewed with, and expires after a few minutes."`
}

type VolumeIDInput struct {
	DatacenterID string `json:"datacenter_id" jsonschema:"the ID of the data center"`
	VolumeID     string `json:"volume_id" jsonschema:"the ID of the volume"`
	Depth        *int32 `json:"depth,omitempty" jsonschema:"nesting depth of returned objects (0-5)"`
}

// CreateVolumeInput is the input for create_volume. Two-phase confirmed; creates
// exactly one volume per call. A standalone volume is not attached to any server
// — use attach_server_volume afterwards.
type CreateVolumeInput struct {
	DatacenterID     string   `json:"datacenter_id" jsonschema:"the ID of the data center to create the volume in"`
	Name             string   `json:"name" jsonschema:"the name of the new volume"`
	Size             float32  `json:"size" jsonschema:"the size of the volume in GB"`
	Type             string   `json:"type" jsonschema:"storage type: HDD, SSD, SSD Standard, SSD Premium, or DAS. DAS (Direct Attached Storage) works only inline with a CUBE server and ignores size."`
	Image            *string  `json:"image,omitempty" jsonschema:"ID of an image or snapshot to use as the template for this volume. Provide exactly one of image, image_alias or licence_type; without one of the first two the volume is created empty and has no operating system. Find IDs with list_images or list_snapshots."`
	ImageAlias       *string  `json:"image_alias,omitempty" jsonschema:"alias of an image to use as the template, e.g. ubuntu:latest. An alternative to image."`
	ImagePassword    *string  `json:"image_password,omitempty" jsonschema:"initial root/administrator password for the installed OS; works with public images only. Allowed characters are a-z, A-Z and 0-9, minimum 8 characters. Cannot be changed later. Prefer ssh_keys for Linux images."`
	SshKeys          []string `json:"ssh_keys,omitempty" jsonschema:"public SSH keys to authorize for login. Supported only when creating from a public Linux image. Can only be set at creation; reads always return null."`
	LicenceType      *string  `json:"licence_type,omitempty" jsonschema:"OS type for the volume: LINUX, WINDOWS, WINDOWS2016, WINDOWS2022, WINDOWS2025, UNKNOWN or OTHER. Required when neither image nor image_alias is given, since the licence type cannot then be inferred."`
	AvailabilityZone *string  `json:"availability_zone,omitempty" jsonschema:"availability zone to provision in: AUTO (default), ZONE_1, ZONE_2 or ZONE_3. Not available for DAS."`
	Bus              *string  `json:"bus,omitempty" jsonschema:"bus type: VIRTIO (default, faster) or IDE. Use IDE only for images without VirtIO drivers."`
	UserData         *string  `json:"user_data,omitempty" jsonschema:"cloud-init configuration as a base64-encoded string. Requires a cloud-init-capable image or image_alias. Can only be set at creation."`
	BackupunitId     *string  `json:"backupunit_id,omitempty" jsonschema:"ID of a backup unit to associate. Requires image or image_alias. Can only be set at creation."`
	ExposeSerial     *bool    `json:"expose_serial,omitempty" jsonschema:"expose the disk serial ID to the server; some operating systems and licensed software require it"`
	HotPlugFlags
	ConfirmationToken *string `json:"confirmation_token,omitempty" jsonschema:"leave empty on the FIRST call to receive a preview of what will be created plus a one-time token; pass that token on the SECOND call (with the same datacenter_id and name) to actually create the volume. The token expires after a few minutes."`
}

// UpdateVolumeInput is the input for update_volume. Partial update. Note that
// size can only ever be increased — shrinking a volume is rejected by the API —
// and image_password, user_data and backupunit_id are immutable after creation.
type UpdateVolumeInput struct {
	DatacenterID string   `json:"datacenter_id" jsonschema:"the ID of the data center the volume is in"`
	VolumeID     string   `json:"volume_id" jsonschema:"the ID of the volume to update"`
	Name         *string  `json:"name,omitempty" jsonschema:"a new name for the volume"`
	Size         *float32 `json:"size,omitempty" jsonschema:"a new size in GB. Can only be INCREASED; the API rejects any attempt to shrink a volume. The guest OS must then grow its filesystem to use the extra space."`
	Bus          *string  `json:"bus,omitempty" jsonschema:"a new bus type: VIRTIO or IDE. Changing this requires a server restart to take effect."`
	ExposeSerial *bool    `json:"expose_serial,omitempty" jsonschema:"expose the disk serial ID to the server, or stop exposing it"`
	BootOrder    *string  `json:"boot_order,omitempty" jsonschema:"whether this volume is used as a boot volume: PRIMARY, NONE or AUTO. PRIMARY makes it the boot volume, and requires EVERY other volume on the same server to be set to NONE first, so set the others before this one. AUTO (the default) is the legacy behaviour and requires all volumes on the server to be AUTO. On a server with Confidential Computing the confidential volume is the only one allowed to be PRIMARY, and it must never be set to NONE. To point a server at a different disk, prefer update_server with boot_volume_id — it takes one call and needs no coordination between volumes."`
	HotPlugFlags
}

// DeleteVolumeInput is the input for delete_volume. Two-phase confirmed; the data
// on the volume is destroyed and cannot be recovered without a snapshot.
type DeleteVolumeInput struct {
	DatacenterID      string  `json:"datacenter_id" jsonschema:"the ID of the data center the volume is in"`
	VolumeID          string  `json:"volume_id" jsonschema:"the ID of the volume to delete"`
	ConfirmationToken *string `json:"confirmation_token,omitempty" jsonschema:"leave empty on the FIRST call to receive a preview (including whether the volume is attached to a server) plus a one-time token; pass that token on the SECOND call to actually delete. All data on the volume is lost. The token authorizes deleting only the volume it was issued for and expires after a few minutes."`
}

// ExtraHotPlugFlags are the capability flags that snapshots and images carry but
// volumes do not. They are kept separate from HotPlugFlags rather than merged so
// each input exposes exactly the set its resource supports — VolumeProperties has
// no cpuHotUnplug, ramHotUnplug or SCSI flags at all.
type ExtraHotPlugFlags struct {
	CpuHotUnplug      *bool `json:"cpu_hot_unplug,omitempty" jsonschema:"whether CPU cores can be removed from a server without rebooting it"`
	RamHotUnplug      *bool `json:"ram_hot_unplug,omitempty" jsonschema:"whether memory can be removed from a server without rebooting it"`
	DiscScsiHotPlug   *bool `json:"disc_scsi_hot_plug,omitempty" jsonschema:"whether a SCSI disk can be attached without rebooting the server"`
	DiscScsiHotUnplug *bool `json:"disc_scsi_hot_unplug,omitempty" jsonschema:"whether a SCSI disk can be detached without rebooting the server"`
}

// UpdateSnapshotInput is the input for update_snapshot. There is no create_snapshot:
// snapshots are taken from a volume with create_volume_snapshot.
//
// The hot-plug flags describe what a volume restored from this snapshot will
// support, so changing them affects future restores rather than the snapshot's data.
type UpdateSnapshotInput struct {
	SnapshotID        string  `json:"snapshot_id" jsonschema:"the ID of the snapshot to update"`
	Name              *string `json:"name,omitempty" jsonschema:"a new name for the snapshot"`
	Description       *string `json:"description,omitempty" jsonschema:"a new description"`
	LicenceType       *string `json:"licence_type,omitempty" jsonschema:"OS type recorded on the snapshot: LINUX, WINDOWS, WINDOWS2016, WINDOWS2022, WINDOWS2025, UNKNOWN or OTHER"`
	SecAuthProtection *bool   `json:"sec_auth_protection,omitempty" jsonschema:"require extra protection such as two-step verification before the snapshot can be deleted"`
	ExposeSerial      *bool   `json:"expose_serial,omitempty" jsonschema:"expose the disk serial ID on volumes restored from this snapshot"`
	RequireLegacyBios *bool   `json:"require_legacy_bios,omitempty" jsonschema:"whether volumes restored from this snapshot need the legacy BIOS"`
	HotPlugFlags
	ExtraHotPlugFlags
}

// DeleteSnapshotInput is the input for delete_snapshot. Two-phase confirmed: a
// snapshot is often the only copy of a volume's earlier state.
type DeleteSnapshotInput struct {
	SnapshotID        string  `json:"snapshot_id" jsonschema:"the ID of the snapshot to delete"`
	ConfirmationToken *string `json:"confirmation_token,omitempty" jsonschema:"leave empty on the FIRST call to receive a preview of the snapshot plus a one-time token; pass that token on the SECOND call to actually delete. A snapshot is frequently the only copy of a volume's earlier contents, so deleting it can remove your only way back. The token expires after a few minutes."`
}

// UpdateImageInput is the input for update_image. There is no create_image — the API
// exposes no way to create one, and only private images you uploaded can be changed
// or deleted; public IONOS images are read-only.
//
// licence_type is read and carried forward when omitted, because the API always
// receives it and an empty value would be rejected or would clear the image's OS type.
type UpdateImageInput struct {
	ImageID           string  `json:"image_id" jsonschema:"the ID of the image to update. Only a private image you uploaded can be changed; public IONOS images are read-only."`
	Name              *string `json:"name,omitempty" jsonschema:"a new name for the image"`
	Description       *string `json:"description,omitempty" jsonschema:"a new description"`
	LicenceType       *string `json:"licence_type,omitempty" jsonschema:"OS type: LINUX, WINDOWS, WINDOWS2016, WINDOWS2022, WINDOWS2025, UNKNOWN or OTHER. Omit to keep the current value — it is read and sent back unchanged."`
	CloudInit         *string `json:"cloud_init,omitempty" jsonschema:"cloud-init compatibility: NONE or V1"`
	ExposeSerial      *bool   `json:"expose_serial,omitempty" jsonschema:"expose the disk serial ID on volumes created from this image"`
	RequireLegacyBios *bool   `json:"require_legacy_bios,omitempty" jsonschema:"whether volumes created from this image need the legacy BIOS"`
	HotPlugFlags
	ExtraHotPlugFlags
}

// DeleteImageInput is the input for delete_image. Two-phase confirmed.
type DeleteImageInput struct {
	ImageID           string  `json:"image_id" jsonschema:"the ID of the image to delete. Only a private image you uploaded can be deleted; public IONOS images cannot."`
	ConfirmationToken *string `json:"confirmation_token,omitempty" jsonschema:"leave empty on the FIRST call to receive a preview of the image plus a one-time token; pass that token on the SECOND call to actually delete. Anything that references this image by ID — Terraform configurations, scripts, autoscaling templates — stops being able to create volumes from it. The token expires after a few minutes."`
}

type SnapshotIDInput struct {
	SnapshotID string `json:"snapshot_id" jsonschema:"the ID of the snapshot"`
	Depth      *int32 `json:"depth,omitempty" jsonschema:"nesting depth of returned objects (0-5)"`
}

type NicIDInput struct {
	DatacenterID string `json:"datacenter_id" jsonschema:"the ID of the data center"`
	ServerID     string `json:"server_id" jsonschema:"the ID of the server"`
	NicID        string `json:"nic_id" jsonschema:"the ID of the network interface"`
	Depth        *int32 `json:"depth,omitempty" jsonschema:"nesting depth of returned objects (0-5, default 1 for list operations)"`
}

// CreateNicInput is the input for create_nic. Two-phase confirmed; creates
// exactly one NIC per call.
type CreateNicInput struct {
	DatacenterID      string   `json:"datacenter_id" jsonschema:"the ID of the data center"`
	ServerID          string   `json:"server_id" jsonschema:"the ID of the server to attach the new NIC to"`
	Lan               int32    `json:"lan" jsonschema:"the numeric ID of the LAN to connect the NIC to (the lan field from list_lans, not a UUID). If no LAN with this ID exists it is created implicitly, so a typo silently creates a new isolated LAN — check list_lans first."`
	Name              *string  `json:"name,omitempty" jsonschema:"the name of the new NIC"`
	Ips               []string `json:"ips,omitempty" jsonschema:"IPv4 addresses to assign. Public IPs must come from a reserved IP block (see list_ip_blocks). Omit or pass an empty list to have an address assigned automatically."`
	Dhcp              *bool    `json:"dhcp,omitempty" jsonschema:"whether the NIC reserves an IP using DHCP (default true)"`
	FirewallActive    *bool    `json:"firewall_active,omitempty" jsonschema:"activate the firewall on this NIC. Careful: an active firewall with no rules blocks ALL incoming traffic, so create the firewall rules you need (create_firewall_rule) before or right after turning it on."`
	FirewallType      *string  `json:"firewall_type,omitempty" jsonschema:"which rule directions are allowed on this NIC: INGRESS (default), EGRESS or BIDIRECTIONAL"`
	ConfirmationToken *string  `json:"confirmation_token,omitempty" jsonschema:"leave empty on the FIRST call to receive a preview of what will be created plus a one-time token; pass that token on the SECOND call (with the same datacenter_id, server_id and lan) to actually create the NIC. The token expires after a few minutes."`
}

// UpdateNicInput is the input for update_nic. Partial update, with one important
// exception: the IONOS SDK always serializes the NIC's lan field, so an update
// that omitted it would move the NIC to LAN 0. update_nic therefore reads the
// NIC's current lan and sends it back unchanged unless lan is given explicitly.
type UpdateNicInput struct {
	DatacenterID   string   `json:"datacenter_id" jsonschema:"the ID of the data center"`
	ServerID       string   `json:"server_id" jsonschema:"the ID of the server the NIC belongs to"`
	NicID          string   `json:"nic_id" jsonschema:"the ID of the NIC to update"`
	Name           *string  `json:"name,omitempty" jsonschema:"a new name for the NIC"`
	Lan            *int32   `json:"lan,omitempty" jsonschema:"move the NIC to this LAN ID. Omit to leave the NIC on its current LAN — the current value is read and preserved automatically, so omitting this never moves the NIC."`
	Ips            []string `json:"ips,omitempty" jsonschema:"replace the NIC's IPv4 addresses. This REPLACES the whole list, so include every address the NIC should keep. Public IPs must come from a reserved IP block."`
	Dhcp           *bool    `json:"dhcp,omitempty" jsonschema:"whether the NIC reserves an IP using DHCP"`
	FirewallActive *bool    `json:"firewall_active,omitempty" jsonschema:"activate or deactivate the firewall on this NIC. Activating it with no rules defined blocks ALL incoming traffic."`
	FirewallType   *string  `json:"firewall_type,omitempty" jsonschema:"which rule directions are allowed on this NIC: INGRESS, EGRESS or BIDIRECTIONAL"`
}

// DeleteNicInput is the input for delete_nic. Two-phase confirmed; deleting a NIC
// also removes its firewall rules and flow logs, and drops the server's
// connectivity on that LAN.
type DeleteNicInput struct {
	DatacenterID      string  `json:"datacenter_id" jsonschema:"the ID of the data center"`
	ServerID          string  `json:"server_id" jsonschema:"the ID of the server the NIC belongs to"`
	NicID             string  `json:"nic_id" jsonschema:"the ID of the NIC to delete"`
	ConfirmationToken *string `json:"confirmation_token,omitempty" jsonschema:"leave empty on the FIRST call to receive a blast-radius preview (firewall rules and flow logs that go with it) plus a one-time token; pass that token on the SECOND call to actually delete. The token authorizes deleting only the NIC it was issued for and expires after a few minutes."`
}

type LanIDInput struct {
	DatacenterID string `json:"datacenter_id" jsonschema:"the ID of the data center"`
	LanID        string `json:"lan_id" jsonschema:"the ID of the LAN"`
	Depth        *int32 `json:"depth,omitempty" jsonschema:"nesting depth of returned objects (0-5, default 1 for list operations)"`
}

// CreateLanInput is the input for create_lan. Two-phase confirmed; creates
// exactly one LAN per call.
type CreateLanInput struct {
	DatacenterID      string  `json:"datacenter_id" jsonschema:"the ID of the data center to create the LAN in"`
	Name              *string `json:"name,omitempty" jsonschema:"the name of the new LAN"`
	Public            *bool   `json:"public,omitempty" jsonschema:"whether the LAN is connected to the internet (default false). A public LAN is how servers reach the internet; a private LAN is internal only."`
	Pcc               *string `json:"pcc,omitempty" jsonschema:"the ID of a private cross connect to attach this LAN to. All LANs on one cross connect must use non-overlapping IP ranges in the same subnet."`
	Ipv6CidrBlock     *string `json:"ipv6_cidr_block,omitempty" jsonschema:"enable IPv6 on this LAN. Pass AUTO to have a /64 block assigned automatically, or an explicit /64 block that sits inside the data center's IPv6 range and is unique among its LANs."`
	ConfirmationToken *string `json:"confirmation_token,omitempty" jsonschema:"leave empty on the FIRST call to receive a preview of what will be created plus a one-time token; pass that token on the SECOND call (with the same datacenter_id and name) to actually create the LAN. The token expires after a few minutes."`
}

// UpdateLanInput is the input for update_lan. Partial update. ipv4_cidr_block is
// read-only and cannot be set.
type UpdateLanInput struct {
	DatacenterID  string  `json:"datacenter_id" jsonschema:"the ID of the data center the LAN is in"`
	LanID         string  `json:"lan_id" jsonschema:"the ID of the LAN to update"`
	Name          *string `json:"name,omitempty" jsonschema:"a new name for the LAN"`
	Public        *bool   `json:"public,omitempty" jsonschema:"make the LAN public (internet-connected) or private. Making a public LAN private removes internet access for every server on it."`
	Pcc           *string `json:"pcc,omitempty" jsonschema:"attach the LAN to this private cross connect ID. This can only SET the connection, not clear it — omitting the field leaves the current value alone. Detaching is supported by the API (the DCD web console does it, keeping both the LAN and the cross connect) but is not yet exposed here; use the DCD for that."`
	Ipv6CidrBlock *string `json:"ipv6_cidr_block,omitempty" jsonschema:"set the LAN's /64 IPv6 block, or AUTO to have one assigned. Changing it reassigns the /80 blocks and addresses of every connected NIC."`
}

// DeleteLanInput is the input for delete_lan. Two-phase confirmed; the preview
// reports how many NICs are still attached, since deleting the LAN disconnects
// every one of them.
type DeleteLanInput struct {
	DatacenterID      string  `json:"datacenter_id" jsonschema:"the ID of the data center the LAN is in"`
	LanID             string  `json:"lan_id" jsonschema:"the ID of the LAN to delete"`
	ConfirmationToken *string `json:"confirmation_token,omitempty" jsonschema:"leave empty on the FIRST call to receive a blast-radius preview (how many NICs are attached) plus a one-time token; pass that token on the SECOND call to actually delete. The token authorizes deleting only the LAN it was issued for and expires after a few minutes."`
}

// Compute action input types.
//
// These back the non-CRUD tools registered through RegisterActionTool: server
// power control, volume snapshot actions, and attach/detach/assign relations.
// The ones whose verb is destructive carry a ConfirmationToken and use the same
// two-phase preview flow as delete_*.

// ServerPowerInput is the input for the non-disruptive power actions
// (start_server, resume_server). Single call — they bring a server up rather
// than interrupting it, so there is nothing to preview.
type ServerPowerInput struct {
	DatacenterID string `json:"datacenter_id" jsonschema:"the ID of the data center the server is in"`
	ServerID     string `json:"server_id" jsonschema:"the ID of the server"`
}

// ServerDisruptiveActionInput is the input for the power actions that interrupt a
// running server (stop_server, reboot_server, suspend_server, upgrade_server).
// Two-phase confirmed: the preview shows the server's current name and state so
// the caller can confirm it is about to hit the right machine.
type ServerDisruptiveActionInput struct {
	DatacenterID      string  `json:"datacenter_id" jsonschema:"the ID of the data center the server is in"`
	ServerID          string  `json:"server_id" jsonschema:"the ID of the server"`
	ConfirmationToken *string `json:"confirmation_token,omitempty" jsonschema:"leave empty on the FIRST call to receive a preview of the server and its current state plus a one-time token; pass that token on the SECOND call to actually perform the action. The token authorizes this one action on this one server and expires after a few minutes."`
}

// CreateVolumeSnapshotInput is the input for create_volume_snapshot. Two-phase
// confirmed. A snapshot is the only way to recover a volume's data after
// delete_volume, so this is the tool to reach for before a destructive change.
type CreateVolumeSnapshotInput struct {
	DatacenterID      string  `json:"datacenter_id" jsonschema:"the ID of the data center the volume is in"`
	VolumeID          string  `json:"volume_id" jsonschema:"the ID of the volume to snapshot"`
	Name              string  `json:"name" jsonschema:"the name of the new snapshot"`
	Description       *string `json:"description,omitempty" jsonschema:"an optional description, e.g. what state the volume was in"`
	SecAuthProtection *bool   `json:"sec_auth_protection,omitempty" jsonschema:"if true, the snapshot requires extra protection such as two-step verification before it can be deleted"`
	LicenceType       *string `json:"licence_type,omitempty" jsonschema:"OS type of the snapshot: LINUX, WINDOWS, WINDOWS2016, WINDOWS2022, WINDOWS2025, UNKNOWN or OTHER. Defaults to the source volume's licence type."`
	ConfirmationToken *string `json:"confirmation_token,omitempty" jsonschema:"leave empty on the FIRST call to receive a preview plus a one-time token; pass that token on the SECOND call (with the same datacenter_id, volume_id and name) to actually create the snapshot. The token expires after a few minutes."`
}

// RestoreVolumeSnapshotInput is the input for restore_volume_snapshot. Two-phase
// confirmed and destructive: restoring OVERWRITES the volume's current contents.
type RestoreVolumeSnapshotInput struct {
	DatacenterID      string  `json:"datacenter_id" jsonschema:"the ID of the data center the volume is in"`
	VolumeID          string  `json:"volume_id" jsonschema:"the ID of the volume to restore INTO. Its current contents are overwritten."`
	SnapshotID        string  `json:"snapshot_id" jsonschema:"the ID of the snapshot to restore from; find it with list_snapshots"`
	ConfirmationToken *string `json:"confirmation_token,omitempty" jsonschema:"leave empty on the FIRST call to receive a preview of the volume that will be overwritten plus a one-time token; pass that token on the SECOND call to actually restore. The token is bound to this volume and snapshot pair and expires after a few minutes."`
}

// AttachServerVolumeInput is the input for attach_server_volume. Single call:
// attaching is additive and reversible with detach_server_volume.
type AttachServerVolumeInput struct {
	DatacenterID string `json:"datacenter_id" jsonschema:"the ID of the data center"`
	ServerID     string `json:"server_id" jsonschema:"the ID of the server to attach the volume to"`
	VolumeID     string `json:"volume_id" jsonschema:"the ID of an EXISTING volume to attach; create one first with create_volume. The volume must be in the same data center as the server."`
}

// DetachServerVolumeInput is the input for detach_server_volume. Two-phase
// confirmed: detaching a boot volume leaves the server unable to boot.
type DetachServerVolumeInput struct {
	DatacenterID      string  `json:"datacenter_id" jsonschema:"the ID of the data center"`
	ServerID          string  `json:"server_id" jsonschema:"the ID of the server to detach the volume from"`
	VolumeID          string  `json:"volume_id" jsonschema:"the ID of the volume to detach"`
	ConfirmationToken *string `json:"confirmation_token,omitempty" jsonschema:"leave empty on the FIRST call to receive a preview plus a one-time token; pass that token on the SECOND call to actually detach. Detaching does NOT delete the volume — it survives as an unattached volume and keeps incurring cost. The token expires after a few minutes."`
}

// AssignServerSecurityGroupsInput is the input for assign_server_security_groups.
// The API call is a PUT that REPLACES the whole assignment set, so the field is
// named and documented as a full list rather than an addition.
type AssignServerSecurityGroupsInput struct {
	DatacenterID     string   `json:"datacenter_id" jsonschema:"the ID of the data center"`
	ServerID         string   `json:"server_id" jsonschema:"the ID of the server"`
	SecurityGroupIDs []string `json:"security_group_ids" jsonschema:"the COMPLETE list of security group IDs the server should have. This REPLACES the current set: any group you omit is unassigned, and an empty list unassigns all of them. Read the current set with get_server at depth 2 first if you mean to add one."`
}

// AssignNicSecurityGroupsInput is the input for assign_nic_security_groups. Also
// a full-set replacing PUT.
type AssignNicSecurityGroupsInput struct {
	DatacenterID     string   `json:"datacenter_id" jsonschema:"the ID of the data center"`
	ServerID         string   `json:"server_id" jsonschema:"the ID of the server the NIC belongs to"`
	NicID            string   `json:"nic_id" jsonschema:"the ID of the NIC"`
	SecurityGroupIDs []string `json:"security_group_ids" jsonschema:"the COMPLETE list of security group IDs the NIC should have. This REPLACES the current set: any group you omit is unassigned, and an empty list unassigns all of them. Read the current set with get_nic at depth 2 first if you mean to add one."`
}

type FirewallRuleIDInput struct {
	DatacenterID   string `json:"datacenter_id" jsonschema:"the ID of the data center"`
	ServerID       string `json:"server_id" jsonschema:"the ID of the server"`
	NicID          string `json:"nic_id" jsonschema:"the ID of the network interface"`
	FirewallRuleID string `json:"firewallrule_id" jsonschema:"the ID of the firewall rule"`
	Depth          *int32 `json:"depth,omitempty" jsonschema:"nesting depth of returned objects (0-5)"`
}

type IpBlockIDInput struct {
	IpBlockID string `json:"ipblock_id" jsonschema:"the ID of the IP block"`
	Depth     *int32 `json:"depth,omitempty" jsonschema:"nesting depth of returned objects (0-5)"`
}

type LoadBalancerIDInput struct {
	DatacenterID   string `json:"datacenter_id" jsonschema:"the ID of the data center"`
	LoadBalancerID string `json:"loadbalancer_id" jsonschema:"the ID of the load balancer"`
	Depth          *int32 `json:"depth,omitempty" jsonschema:"nesting depth of returned objects (0-5, default 1 for list operations)"`
}

type NetworkLoadBalancerIDInput struct {
	DatacenterID          string `json:"datacenter_id" jsonschema:"the ID of the data center"`
	NetworkLoadBalancerID string `json:"network_loadbalancer_id" jsonschema:"the ID of the network load balancer"`
	Depth                 *int32 `json:"depth,omitempty" jsonschema:"nesting depth of returned objects (0-5, default 1 for list operations)"`
}

type ApplicationLoadBalancerIDInput struct {
	DatacenterID              string `json:"datacenter_id" jsonschema:"the ID of the data center"`
	ApplicationLoadBalancerID string `json:"application_loadbalancer_id" jsonschema:"the ID of the application load balancer"`
	Depth                     *int32 `json:"depth,omitempty" jsonschema:"nesting depth of returned objects (0-5, default 1 for list operations)"`
}

type TargetGroupIDInput struct {
	TargetGroupID string `json:"target_group_id" jsonschema:"the ID of the target group"`
	Depth         *int32 `json:"depth,omitempty" jsonschema:"nesting depth of returned objects (0-5)"`
}

type NatGatewayIDInput struct {
	DatacenterID string `json:"datacenter_id" jsonschema:"the ID of the data center"`
	NatGatewayID string `json:"nat_gateway_id" jsonschema:"the ID of the NAT gateway"`
	Depth        *int32 `json:"depth,omitempty" jsonschema:"nesting depth of returned objects (0-5, default 1 for list operations)"`
}

// Networking write input types: IP blocks, security groups and their rules,
// NIC firewall rules, and private cross connects.

// CreateIpBlockInput is the input for create_ip_block. Two-phase confirmed. An IP
// block is account-level rather than inside a data center, and it is billed from
// creation whether or not its addresses are in use.
type CreateIpBlockInput struct {
	Location          string  `json:"location" jsonschema:"the physical location to reserve the addresses in, e.g. de/fra, de/txl, us/las, us/ewr, gb/lhr, es/vit, fr/par. Must match the location of the data center whose resources will use them, and cannot be changed afterwards."`
	Size              int32   `json:"size" jsonschema:"how many public IPv4 addresses to reserve. Cannot be changed afterwards — reserve a new block if you need more."`
	Name              *string `json:"name,omitempty" jsonschema:"a name for the block, which is the only property you can change later"`
	ConfirmationToken *string `json:"confirmation_token,omitempty" jsonschema:"leave empty on the FIRST call to receive a preview plus a one-time token; pass that token on the SECOND call (with the same location and size) to actually reserve the block. The token expires after a few minutes."`
}

// DeleteIpBlockInput is the input for delete_ip_block. Two-phase confirmed; the
// preview lists which resources are still using the addresses.
type DeleteIpBlockInput struct {
	IpBlockID         string  `json:"ipblock_id" jsonschema:"the ID of the IP block to release"`
	ConfirmationToken *string `json:"confirmation_token,omitempty" jsonschema:"leave empty on the FIRST call to receive a preview listing every resource still using these addresses, plus a one-time token; pass that token on the SECOND call to actually release the block. Releasing addresses that are still assigned breaks connectivity for those resources, and a new block can only be requested by location and size, so there is no way to ask for the same addresses back. The token expires after a few minutes."`
}

// CreateSecurityGroupInput is the input for create_security_group. Two-phase
// confirmed. A new group has no rules, so it permits nothing until rules are added.
type CreateSecurityGroupInput struct {
	DatacenterID      string  `json:"datacenter_id" jsonschema:"the ID of the data center to create the security group in"`
	Name              string  `json:"name" jsonschema:"the name of the new security group"`
	Description       *string `json:"description,omitempty" jsonschema:"an optional description of what the group is for"`
	ConfirmationToken *string `json:"confirmation_token,omitempty" jsonschema:"leave empty on the FIRST call to receive a preview plus a one-time token; pass that token on the SECOND call (with the same datacenter_id and name) to actually create the group. The token expires after a few minutes."`
}

// UpdateSecurityGroupInput is the input for update_security_group. name is
// required because the IONOS SDK always serializes it: an update that omitted it
// would send an empty name and wipe the group's name as a side effect.
type UpdateSecurityGroupInput struct {
	DatacenterID    string  `json:"datacenter_id" jsonschema:"the ID of the data center the security group is in"`
	SecurityGroupID string  `json:"security_group_id" jsonschema:"the ID of the security group to update"`
	Name            *string `json:"name,omitempty" jsonschema:"a new name for the group. Omit to keep the current name — it is read and sent back unchanged, because the API always receives this field and an empty value would clear it."`
	Description     *string `json:"description,omitempty" jsonschema:"a new description for the group"`
}

// DeleteSecurityGroupInput is the input for delete_security_group. Two-phase
// confirmed; the preview counts the rules deleted with it and the servers and NICs
// that lose the protection it provided.
type DeleteSecurityGroupInput struct {
	DatacenterID      string  `json:"datacenter_id" jsonschema:"the ID of the data center the security group is in"`
	SecurityGroupID   string  `json:"security_group_id" jsonschema:"the ID of the security group to delete"`
	ConfirmationToken *string `json:"confirmation_token,omitempty" jsonschema:"leave empty on the FIRST call to receive a blast-radius preview (its rules, and the servers and NICs it is assigned to) plus a one-time token; pass that token on the SECOND call to actually delete. Every server and NIC using the group loses the protection its rules provided. The token expires after a few minutes."`
}

// RuleFields are the properties of a firewall rule. Embedded by both the
// NIC-scoped firewall rule tools and the security-group rule tools, because the
// API uses one FirewallruleProperties model for both and the semantics are
// identical — only the parent chain differs.
type RuleFields struct {
	Protocol       *string `json:"protocol,omitempty" jsonschema:"the protocol the rule matches: TCP, UDP, ICMP, ICMPv6, GRE, VRRP, ESP, AH or ANY. Required when creating a rule. ANY matches every protocol and forbids ports and ICMP fields."`
	Name           *string `json:"name,omitempty" jsonschema:"a name for the rule"`
	Type           *string `json:"type,omitempty" jsonschema:"direction the rule applies to: INGRESS (inbound, the default) or EGRESS (outbound)"`
	SourceMac      *string `json:"source_mac,omitempty" jsonschema:"match only traffic from this MAC address, e.g. aa:bb:cc:dd:ee:ff. On create, omitting it matches any source MAC. On update, omitting it leaves the current value unchanged — to widen the rule back to any MAC, list source_mac in the clear field instead."`
	SourceIp       *string `json:"source_ip,omitempty" jsonschema:"match only traffic from this IP address or CIDR range. On create, omitting it matches any source. On update, omitting it leaves the current value unchanged — to widen the rule back to any source, list source_ip in the clear field instead. Do NOT pass 0.0.0.0/0 to mean anywhere: the API stores that as the literal address 0.0.0.0, which matches no real traffic."`
	TargetIp       *string `json:"target_ip,omitempty" jsonschema:"match only traffic to this IP address or CIDR range; for an INGRESS rule this is usually one of the NIC's own IPs. On create, omitting it matches any destination. On update, omitting it leaves the current value unchanged — to widen it back to any destination, list target_ip in the clear field instead. Do NOT pass 0.0.0.0/0."`
	IpVersion      *string `json:"ip_version,omitempty" jsonschema:"IPv4 or IPv6. Defaults to the version implied by the addresses given, or IPv4. On update, list ip_version in the clear field to go back to that automatic behaviour."`
	PortRangeStart *int32  `json:"port_range_start,omitempty" jsonschema:"first port in the allowed range (1-65534). Only valid with protocol TCP or UDP, and must be given together with port_range_end. Omit both to allow all ports."`
	PortRangeEnd   *int32  `json:"port_range_end,omitempty" jsonschema:"last port in the allowed range (1-65534). Only valid with protocol TCP or UDP, and must be given together with port_range_start."`
	IcmpType       *int32  `json:"icmp_type,omitempty" jsonschema:"ICMP type to allow (0-254). Only valid with protocol ICMP or ICMPv6. On create, omitting it allows all types. On update, omitting it leaves the current value unchanged — list icmp_type in the clear field to allow all types again."`
	IcmpCode       *int32  `json:"icmp_code,omitempty" jsonschema:"ICMP code to allow (0-254). Only valid with protocol ICMP or ICMPv6. On create, omitting it allows all codes. On update, omitting it leaves the current value unchanged — list icmp_code in the clear field to allow all codes again."`
}

// CreateFirewallRuleInput is the input for create_firewall_rule, which adds a rule
// to a single NIC. Two-phase confirmed.
type CreateFirewallRuleInput struct {
	DatacenterID string `json:"datacenter_id" jsonschema:"the ID of the data center"`
	ServerID     string `json:"server_id" jsonschema:"the ID of the server the NIC belongs to"`
	NicID        string `json:"nic_id" jsonschema:"the ID of the NIC to add the rule to"`
	RuleFields
	ConfirmationToken *string `json:"confirmation_token,omitempty" jsonschema:"leave empty on the FIRST call to receive a preview plus a one-time token; pass that token on the SECOND call (with the same parent IDs and protocol) to actually create the rule. The token expires after a few minutes."`
}

// UpdateFirewallRuleInput is the input for update_firewall_rule. Partial update.
type UpdateFirewallRuleInput struct {
	DatacenterID   string `json:"datacenter_id" jsonschema:"the ID of the data center"`
	ServerID       string `json:"server_id" jsonschema:"the ID of the server the NIC belongs to"`
	NicID          string `json:"nic_id" jsonschema:"the ID of the NIC the rule is on"`
	FirewallRuleID string `json:"firewallrule_id" jsonschema:"the ID of the firewall rule to update"`
	RuleFields
	Clear []string `json:"clear,omitempty" jsonschema:"field names to reset so the rule stops matching on them, i.e. to WIDEN it: source_ip, target_ip, source_mac, ip_version, icmp_type, icmp_code. This is the only way to go back to 'any' — omitting a field leaves it unchanged, and there is no value that means anywhere (0.0.0.0/0 is stored by the API as the literal address 0.0.0.0 and matches nothing). Clearing source_ip on an INGRESS rule opens it to the whole internet, so say so before you do it. Listing a field you also set in the same call is rejected."`
}

// DeleteFirewallRuleInput is the input for delete_firewall_rule. Two-phase
// confirmed: removing a rule closes the traffic it was allowing.
type DeleteFirewallRuleInput struct {
	DatacenterID      string  `json:"datacenter_id" jsonschema:"the ID of the data center"`
	ServerID          string  `json:"server_id" jsonschema:"the ID of the server the NIC belongs to"`
	NicID             string  `json:"nic_id" jsonschema:"the ID of the NIC the rule is on"`
	FirewallRuleID    string  `json:"firewallrule_id" jsonschema:"the ID of the firewall rule to delete"`
	ConfirmationToken *string `json:"confirmation_token,omitempty" jsonschema:"leave empty on the FIRST call to receive a preview of the rule plus a one-time token; pass that token on the SECOND call to actually delete it. Traffic the rule was allowing will be blocked, and if this is the NIC's last rule while its firewall is active, ALL incoming traffic is blocked. The token expires after a few minutes."`
}

// CreateSecurityGroupRuleInput is the input for create_security_group_rule, which
// adds a rule to a security group so every server and NIC assigned to that group
// inherits it. Two-phase confirmed.
type CreateSecurityGroupRuleInput struct {
	DatacenterID    string `json:"datacenter_id" jsonschema:"the ID of the data center"`
	SecurityGroupID string `json:"security_group_id" jsonschema:"the ID of the security group to add the rule to"`
	RuleFields
	ConfirmationToken *string `json:"confirmation_token,omitempty" jsonschema:"leave empty on the FIRST call to receive a preview (including how many servers and NICs will inherit the rule) plus a one-time token; pass that token on the SECOND call to actually create it. The token expires after a few minutes."`
}

// UpdateSecurityGroupRuleInput is the input for update_security_group_rule.
// Partial update; the change applies to every member of the group at once.
type UpdateSecurityGroupRuleInput struct {
	DatacenterID    string `json:"datacenter_id" jsonschema:"the ID of the data center"`
	SecurityGroupID string `json:"security_group_id" jsonschema:"the ID of the security group the rule belongs to"`
	RuleID          string `json:"rule_id" jsonschema:"the ID of the rule to update"`
	RuleFields
	Clear []string `json:"clear,omitempty" jsonschema:"field names to reset so the rule stops matching on them, i.e. to WIDEN it: source_ip, target_ip, source_mac, ip_version, icmp_type, icmp_code. This is the only way to go back to 'any' — omitting a field leaves it unchanged, and there is no value that means anywhere (0.0.0.0/0 is stored by the API as the literal address 0.0.0.0 and matches nothing). Clearing source_ip on an INGRESS rule opens it to the whole internet, so say so before you do it. Listing a field you also set in the same call is rejected."`
}

// DeleteSecurityGroupRuleInput is the input for delete_security_group_rule.
// Two-phase confirmed: the rule is removed for every member of the group.
type DeleteSecurityGroupRuleInput struct {
	DatacenterID      string  `json:"datacenter_id" jsonschema:"the ID of the data center"`
	SecurityGroupID   string  `json:"security_group_id" jsonschema:"the ID of the security group the rule belongs to"`
	RuleID            string  `json:"rule_id" jsonschema:"the ID of the rule to delete"`
	ConfirmationToken *string `json:"confirmation_token,omitempty" jsonschema:"leave empty on the FIRST call to receive a preview (including how many servers and NICs lose the rule) plus a one-time token; pass that token on the SECOND call to actually delete it. Every member of the group stops allowing the traffic this rule permitted. The token expires after a few minutes."`
}

// CreatePccInput is the input for create_pcc. Two-phase confirmed. A private cross
// connect is account-level: it links private LANs across data centers.
type CreatePccInput struct {
	Name              string  `json:"name" jsonschema:"the name of the new private cross connect"`
	Description       *string `json:"description,omitempty" jsonschema:"an optional description of what it connects"`
	ConfirmationToken *string `json:"confirmation_token,omitempty" jsonschema:"leave empty on the FIRST call to receive a preview plus a one-time token; pass that token on the SECOND call (with the same name) to actually create it. The token expires after a few minutes."`
}

// UpdatePccInput is the input for update_pcc. Partial update; the connected LANs
// are managed from the LAN side with update_lan's pcc field, not here.
type UpdatePccInput struct {
	PccID       string  `json:"pcc_id" jsonschema:"the ID of the private cross connect to update"`
	Name        *string `json:"name,omitempty" jsonschema:"a new name"`
	Description *string `json:"description,omitempty" jsonschema:"a new description"`
}

// DeletePccInput is the input for delete_pcc. Two-phase confirmed; the preview
// counts the LANs still peered through it, all of which lose that connection.
type DeletePccInput struct {
	PccID             string  `json:"pcc_id" jsonschema:"the ID of the private cross connect to delete"`
	ConfirmationToken *string `json:"confirmation_token,omitempty" jsonschema:"leave empty on the FIRST call to receive a blast-radius preview (the LANs peered through it) plus a one-time token; pass that token on the SECOND call to actually delete it. Every peered LAN loses its cross-data-center connection. The token expires after a few minutes."`
}

// Load-balancing write input types.
//
// Every model in this area except the classic load balancer serializes its
// required fields unconditionally, so each update tool reads the resource first
// and carries those values forward — see the "PATCH bodies" note in CLAUDE.md.
// The worst case is a forwarding rule's targets list: a partial update built
// without it would send an empty targets array and wipe the load balancer's
// entire backend pool.

// ManagedLoadBalancerFields are the properties shared by network and application
// load balancers, whose API models are field-for-field identical.
type ManagedLoadBalancerFields struct {
	Ips            []string `json:"ips,omitempty" jsonschema:"public IPv4 addresses the load balancer listens on. They must come from a reserved IP block in the same location (see list_ip_blocks and create_ip_block). Omit to have addresses assigned automatically."`
	LbPrivateIps   []string `json:"lb_private_ips,omitempty" jsonschema:"private IPs the load balancer uses to reach its targets on the target LAN, in CIDR form. Omit to have them assigned automatically."`
	CentralLogging *bool    `json:"central_logging,omitempty" jsonschema:"send the load balancer's logs to the central logging service"`
	LoggingFormat  *string  `json:"logging_format,omitempty" jsonschema:"the log line format to use when central_logging is on"`
}

// CreateManagedLoadBalancerInput is the input for create_network_loadbalancer and
// create_application_loadbalancer. Two-phase confirmed.
type CreateManagedLoadBalancerInput struct {
	DatacenterID string `json:"datacenter_id" jsonschema:"the ID of the data center to create the load balancer in"`
	Name         string `json:"name" jsonschema:"the name of the new load balancer"`
	ListenerLan  int32  `json:"listener_lan" jsonschema:"the numeric ID of the LAN the load balancer listens on — usually a PUBLIC LAN, since this is the side clients connect to. Use the lan value from list_lans, not a UUID."`
	TargetLan    int32  `json:"target_lan" jsonschema:"the numeric ID of the LAN holding the backend servers — usually a PRIVATE LAN. Must be different from listener_lan."`
	ManagedLoadBalancerFields
	ConfirmationToken *string `json:"confirmation_token,omitempty" jsonschema:"leave empty on the FIRST call to receive a preview plus a one-time token; pass that token on the SECOND call (with the same datacenter_id and name) to actually create it. The token expires after a few minutes."`
}

// UpdateManagedLoadBalancerInput is the input for update_network_loadbalancer and
// update_application_loadbalancer. Partial update: omitted fields are read from
// the current resource and sent back unchanged, because the API always receives
// name, listener_lan and target_lan.
type UpdateManagedLoadBalancerInput struct {
	DatacenterID   string  `json:"datacenter_id" jsonschema:"the ID of the data center the load balancer is in"`
	LoadBalancerID string  `json:"loadbalancer_id" jsonschema:"the ID of the load balancer to update"`
	Name           *string `json:"name,omitempty" jsonschema:"a new name. Omit to keep the current one — it is read and sent back unchanged."`
	ListenerLan    *int32  `json:"listener_lan,omitempty" jsonschema:"move the listener to this LAN ID. Omit to keep the current LAN; changing it moves where clients connect."`
	TargetLan      *int32  `json:"target_lan,omitempty" jsonschema:"move the target side to this LAN ID. Omit to keep the current LAN; changing it repoints the load balancer at a different backend network."`
	ManagedLoadBalancerFields
}

// DeleteManagedLoadBalancerInput is the input for delete_network_loadbalancer and
// delete_application_loadbalancer. Two-phase confirmed.
type DeleteManagedLoadBalancerInput struct {
	DatacenterID      string  `json:"datacenter_id" jsonschema:"the ID of the data center the load balancer is in"`
	LoadBalancerID    string  `json:"loadbalancer_id" jsonschema:"the ID of the load balancer to delete"`
	ConfirmationToken *string `json:"confirmation_token,omitempty" jsonschema:"leave empty on the FIRST call to receive a blast-radius preview (its forwarding rules, and the addresses that stop serving traffic) plus a one-time token; pass that token on the SECOND call to actually delete. Clients connecting to its listener IPs will no longer reach the backends. The token expires after a few minutes."`
}

// CreateLoadBalancerInput is the input for create_loadbalancer, the classic
// load balancer. Two-phase confirmed. It balances traffic across NICs attached to
// it rather than across IP targets, so attach NICs with attach_loadbalancer_nic.
type CreateLoadBalancerInput struct {
	DatacenterID      string  `json:"datacenter_id" jsonschema:"the ID of the data center to create the load balancer in"`
	Name              string  `json:"name" jsonschema:"the name of the new load balancer"`
	Ip                *string `json:"ip,omitempty" jsonschema:"the IPv4 address to listen on, from a reserved IP block. Omit to have one assigned automatically."`
	Dhcp              *bool   `json:"dhcp,omitempty" jsonschema:"whether the load balancer reserves its IP using DHCP (default true)"`
	ConfirmationToken *string `json:"confirmation_token,omitempty" jsonschema:"leave empty on the FIRST call to receive a preview plus a one-time token; pass that token on the SECOND call (with the same datacenter_id and name) to actually create it. The token expires after a few minutes."`
}

// UpdateLoadBalancerInput is the input for update_loadbalancer. All of the classic
// load balancer's properties are optional in the API model, so this is a genuine
// partial update with no carry-forward read.
type UpdateLoadBalancerInput struct {
	DatacenterID   string  `json:"datacenter_id" jsonschema:"the ID of the data center the load balancer is in"`
	LoadBalancerID string  `json:"loadbalancer_id" jsonschema:"the ID of the load balancer to update"`
	Name           *string `json:"name,omitempty" jsonschema:"a new name"`
	Ip             *string `json:"ip,omitempty" jsonschema:"a new IPv4 address to listen on, from a reserved IP block"`
	Dhcp           *bool   `json:"dhcp,omitempty" jsonschema:"whether the load balancer reserves its IP using DHCP"`
}

// DeleteLoadBalancerInput is the input for delete_loadbalancer. Two-phase
// confirmed; the preview counts the NICs it is balancing across.
type DeleteLoadBalancerInput struct {
	DatacenterID      string  `json:"datacenter_id" jsonschema:"the ID of the data center the load balancer is in"`
	LoadBalancerID    string  `json:"loadbalancer_id" jsonschema:"the ID of the load balancer to delete"`
	ConfirmationToken *string `json:"confirmation_token,omitempty" jsonschema:"leave empty on the FIRST call to receive a blast-radius preview (the NICs it balances across) plus a one-time token; pass that token on the SECOND call to actually delete. The NICs themselves are not deleted, but traffic stops being balanced to them. The token expires after a few minutes."`
}

// NlbTargetInput is one backend behind a network load balancer forwarding rule.
type NlbTargetInput struct {
	Ip                  string  `json:"ip" jsonschema:"the IPv4 or IPv6 address of the backend, usually a server's private IP on the load balancer's target LAN"`
	Port                int32   `json:"port" jsonschema:"the port the backend listens on (1-65535)"`
	Weight              int32   `json:"weight" jsonschema:"relative share of traffic this backend receives (0-256). A weight of 0 takes it out of rotation without removing it."`
	ProxyProtocol       *string `json:"proxy_protocol,omitempty" jsonschema:"PROXY protocol version used to pass the client address to the backend: none, v1, v2 or v2ssl"`
	HealthCheck         *bool   `json:"health_check,omitempty" jsonschema:"whether the load balancer health-checks this backend"`
	HealthCheckInterval *int32  `json:"health_check_interval,omitempty" jsonschema:"how often to health-check this backend, in milliseconds"`
	Maintenance         *bool   `json:"maintenance,omitempty" jsonschema:"put this backend into maintenance so it receives no traffic while staying configured"`
}

// NlbHealthCheckInput is the rule-level health check for a network load balancer
// forwarding rule. All values are milliseconds except retries.
type NlbHealthCheckInput struct {
	ClientTimeout  *int32 `json:"client_timeout,omitempty" jsonschema:"how long an inactive client connection is kept open, in milliseconds"`
	ConnectTimeout *int32 `json:"connect_timeout,omitempty" jsonschema:"how long to wait when connecting to a backend, in milliseconds"`
	TargetTimeout  *int32 `json:"target_timeout,omitempty" jsonschema:"how long a backend has to respond before the connection is considered dead, in milliseconds"`
	Retries        *int32 `json:"retries,omitempty" jsonschema:"how many times to retry a failed connection attempt (0-65535)"`
}

// CreateNlbForwardingRuleInput is the input for create_nlb_forwarding_rule.
// Two-phase confirmed. A network load balancer carries no traffic until it has at
// least one forwarding rule.
type CreateNlbForwardingRuleInput struct {
	DatacenterID      string               `json:"datacenter_id" jsonschema:"the ID of the data center"`
	LoadBalancerID    string               `json:"loadbalancer_id" jsonschema:"the ID of the network load balancer to add the rule to"`
	Name              string               `json:"name" jsonschema:"the name of the new forwarding rule"`
	Algorithm         string               `json:"algorithm" jsonschema:"how connections are distributed across the targets: ROUND_ROBIN, LEAST_CONNECTION, RANDOM or SOURCE_IP"`
	Protocol          string               `json:"protocol" jsonschema:"the transport protocol to forward: TCP or UDP"`
	ListenerIp        string               `json:"listener_ip" jsonschema:"the address clients connect to. Must be one of the load balancer's own listener IPs."`
	ListenerPort      int32                `json:"listener_port" jsonschema:"the port clients connect to (1-65535)"`
	Targets           []NlbTargetInput     `json:"targets" jsonschema:"the backends to forward to. At least one is required — a rule with no targets accepts connections and has nowhere to send them."`
	HealthCheck       *NlbHealthCheckInput `json:"health_check,omitempty" jsonschema:"rule-level timeouts and retry behaviour"`
	ConfirmationToken *string              `json:"confirmation_token,omitempty" jsonschema:"leave empty on the FIRST call to receive a preview plus a one-time token; pass that token on the SECOND call (with the same parent IDs and name) to actually create the rule. The token expires after a few minutes."`
}

// UpdateNlbForwardingRuleInput is the input for update_nlb_forwarding_rule.
//
// Partial update, but with an important caveat: the API always receives name,
// algorithm, protocol, listener_ip, listener_port AND targets, so all of them are
// read from the current rule and carried forward when omitted. Without that, a
// rename would send an empty targets list and remove every backend from the load
// balancer.
type UpdateNlbForwardingRuleInput struct {
	DatacenterID   string               `json:"datacenter_id" jsonschema:"the ID of the data center"`
	LoadBalancerID string               `json:"loadbalancer_id" jsonschema:"the ID of the network load balancer"`
	RuleID         string               `json:"rule_id" jsonschema:"the ID of the forwarding rule to update"`
	Name           *string              `json:"name,omitempty" jsonschema:"a new name. Omit to keep the current one."`
	Algorithm      *string              `json:"algorithm,omitempty" jsonschema:"a new distribution algorithm. Omit to keep the current one."`
	Protocol       *string              `json:"protocol,omitempty" jsonschema:"a new transport protocol: TCP or UDP. Omit to keep the current one."`
	ListenerIp     *string              `json:"listener_ip,omitempty" jsonschema:"a new listener address. Omit to keep the current one; changing it moves where clients connect."`
	ListenerPort   *int32               `json:"listener_port,omitempty" jsonschema:"a new listener port. Omit to keep the current one."`
	Targets        []NlbTargetInput     `json:"targets,omitempty" jsonschema:"REPLACE the rule's backends with this list. Include every backend the rule should keep — any you omit stops receiving traffic. Omit the field entirely to leave the current backends untouched."`
	HealthCheck    *NlbHealthCheckInput `json:"health_check,omitempty" jsonschema:"replace the rule-level timeouts and retry behaviour"`
}

// DeleteNlbForwardingRuleInput is the input for delete_nlb_forwarding_rule.
type DeleteNlbForwardingRuleInput struct {
	DatacenterID      string  `json:"datacenter_id" jsonschema:"the ID of the data center"`
	LoadBalancerID    string  `json:"loadbalancer_id" jsonschema:"the ID of the network load balancer"`
	RuleID            string  `json:"rule_id" jsonschema:"the ID of the forwarding rule to delete"`
	ConfirmationToken *string `json:"confirmation_token,omitempty" jsonschema:"leave empty on the FIRST call to receive a preview of the listener and its backends plus a one-time token; pass that token on the SECOND call to actually delete. Clients connecting to that listener address and port stop being served. The token expires after a few minutes."`
}

// AlbHttpRuleConditionInput is one condition that decides whether an ALB HTTP rule
// applies to a request.
type AlbHttpRuleConditionInput struct {
	Type      string  `json:"type" jsonschema:"what to match on: HEADER, PATH, QUERY, METHOD, HOST, COOKIE or SOURCE_IP"`
	Condition string  `json:"condition" jsonschema:"how to match: EQUALS, LESS_THAN, GREATER_THAN, STARTS_WITH, ENDS_WITH, CONTAINS or MATCHES"`
	Negate    *bool   `json:"negate,omitempty" jsonschema:"invert the match, so the rule applies when the condition does NOT hold"`
	Key       *string `json:"key,omitempty" jsonschema:"the name to match against, e.g. the header or cookie name. Not used for PATH, METHOD, HOST or SOURCE_IP."`
	Value     *string `json:"value,omitempty" jsonschema:"the value to match against"`
}

// AlbHttpRuleInput is one HTTP routing rule on an application load balancer
// forwarding rule. Its type decides which of the other fields apply.
type AlbHttpRuleInput struct {
	Name            string                      `json:"name" jsonschema:"the name of this HTTP rule"`
	Type            string                      `json:"type" jsonschema:"what the rule does: FORWARD (send to a target group), REDIRECT (send a redirect response) or STATIC (return a fixed response)"`
	TargetGroup     *string                     `json:"target_group,omitempty" jsonschema:"for FORWARD: the ID of the target group to send matching requests to. See create_target_group."`
	DropQuery       *bool                       `json:"drop_query,omitempty" jsonschema:"for REDIRECT: drop the original query string instead of carrying it over"`
	Location        *string                     `json:"location,omitempty" jsonschema:"for REDIRECT: the URL to redirect to"`
	StatusCode      *int32                      `json:"status_code,omitempty" jsonschema:"the HTTP status to return: 301, 302, 303, 307 or 308 for REDIRECT; 200, 503 or 599 for STATIC"`
	ResponseMessage *string                     `json:"response_message,omitempty" jsonschema:"for STATIC: the body to return"`
	ContentType     *string                     `json:"content_type,omitempty" jsonschema:"for STATIC: the Content-Type of the response body"`
	Conditions      []AlbHttpRuleConditionInput `json:"conditions,omitempty" jsonschema:"conditions that must all hold for this rule to apply. A rule with no conditions matches every request."`
}

// CreateAlbForwardingRuleInput is the input for create_alb_forwarding_rule.
// Two-phase confirmed.
type CreateAlbForwardingRuleInput struct {
	DatacenterID       string             `json:"datacenter_id" jsonschema:"the ID of the data center"`
	LoadBalancerID     string             `json:"loadbalancer_id" jsonschema:"the ID of the application load balancer to add the rule to"`
	Name               string             `json:"name" jsonschema:"the name of the new forwarding rule"`
	Protocol           string             `json:"protocol" jsonschema:"the protocol to serve: HTTP or HTTPS"`
	ListenerIp         string             `json:"listener_ip" jsonschema:"the address clients connect to. Must be one of the load balancer's own listener IPs."`
	ListenerPort       int32              `json:"listener_port" jsonschema:"the port clients connect to (1-65535)"`
	ClientTimeout      *int32             `json:"client_timeout,omitempty" jsonschema:"how long an inactive client connection is kept open, in milliseconds"`
	ServerCertificates []string           `json:"server_certificates,omitempty" jsonschema:"certificate IDs to serve for HTTPS. Required in practice when protocol is HTTPS; manage the certificates with the Certificate Manager tools."`
	HttpRules          []AlbHttpRuleInput `json:"http_rules,omitempty" jsonschema:"the HTTP routing rules applied to matching requests. Without any, the listener accepts connections but routes nothing."`
	ConfirmationToken  *string            `json:"confirmation_token,omitempty" jsonschema:"leave empty on the FIRST call to receive a preview plus a one-time token; pass that token on the SECOND call (with the same parent IDs and name) to actually create the rule. The token expires after a few minutes."`
}

// UpdateAlbForwardingRuleInput is the input for update_alb_forwarding_rule. The API
// always receives name, protocol, listener_ip and listener_port, so those are read
// and carried forward when omitted.
type UpdateAlbForwardingRuleInput struct {
	DatacenterID       string             `json:"datacenter_id" jsonschema:"the ID of the data center"`
	LoadBalancerID     string             `json:"loadbalancer_id" jsonschema:"the ID of the application load balancer"`
	RuleID             string             `json:"rule_id" jsonschema:"the ID of the forwarding rule to update"`
	Name               *string            `json:"name,omitempty" jsonschema:"a new name. Omit to keep the current one."`
	Protocol           *string            `json:"protocol,omitempty" jsonschema:"a new protocol: HTTP or HTTPS. Omit to keep the current one."`
	ListenerIp         *string            `json:"listener_ip,omitempty" jsonschema:"a new listener address. Omit to keep the current one."`
	ListenerPort       *int32             `json:"listener_port,omitempty" jsonschema:"a new listener port. Omit to keep the current one."`
	ClientTimeout      *int32             `json:"client_timeout,omitempty" jsonschema:"a new client timeout in milliseconds"`
	ServerCertificates []string           `json:"server_certificates,omitempty" jsonschema:"REPLACE the served certificates with this list. Omit the field entirely to leave them untouched."`
	HttpRules          []AlbHttpRuleInput `json:"http_rules,omitempty" jsonschema:"REPLACE the HTTP routing rules with this list. Include every rule the listener should keep. Omit the field entirely to leave them untouched."`
}

// DeleteAlbForwardingRuleInput is the input for delete_alb_forwarding_rule.
type DeleteAlbForwardingRuleInput struct {
	DatacenterID      string  `json:"datacenter_id" jsonschema:"the ID of the data center"`
	LoadBalancerID    string  `json:"loadbalancer_id" jsonschema:"the ID of the application load balancer"`
	RuleID            string  `json:"rule_id" jsonschema:"the ID of the forwarding rule to delete"`
	ConfirmationToken *string `json:"confirmation_token,omitempty" jsonschema:"leave empty on the FIRST call to receive a preview of the listener and its HTTP rules plus a one-time token; pass that token on the SECOND call to actually delete. Clients connecting to that listener address and port stop being served. The token expires after a few minutes."`
}

// TargetGroupTargetInput is one backend behind a target group.
type TargetGroupTargetInput struct {
	Ip                 string  `json:"ip" jsonschema:"the IPv4 or IPv6 address of the backend"`
	Port               int32   `json:"port" jsonschema:"the port the backend listens on (1-65535)"`
	Weight             int32   `json:"weight" jsonschema:"relative share of traffic this backend receives (0-256). A weight of 0 takes it out of rotation without removing it."`
	ProxyProtocol      *string `json:"proxy_protocol,omitempty" jsonschema:"PROXY protocol version to announce the client address with: none, v1, v2 or v2ssl"`
	HealthCheckEnabled *bool   `json:"health_check_enabled,omitempty" jsonschema:"whether the target group health-checks this backend"`
	MaintenanceEnabled *bool   `json:"maintenance_enabled,omitempty" jsonschema:"put this backend into maintenance so it receives no traffic while staying configured"`
}

// CreateTargetGroupInput is the input for create_target_group. Two-phase confirmed.
// A target group is account-level and is referenced by application load balancer
// HTTP rules, so it takes no datacenter_id.
type CreateTargetGroupInput struct {
	Name              string                   `json:"name" jsonschema:"the name of the new target group"`
	Algorithm         string                   `json:"algorithm" jsonschema:"how traffic is distributed across the targets: ROUND_ROBIN, LEAST_CONNECTION, RANDOM or SOURCE_IP"`
	Protocol          string                   `json:"protocol" jsonschema:"the protocol the targets speak: HTTP or TCP"`
	ProtocolVersion   *string                  `json:"protocol_version,omitempty" jsonschema:"for HTTP: HTTP1 or HTTP2"`
	Targets           []TargetGroupTargetInput `json:"targets,omitempty" jsonschema:"the backends behind this group. A group with no targets accepts no traffic."`
	ConfirmationToken *string                  `json:"confirmation_token,omitempty" jsonschema:"leave empty on the FIRST call to receive a preview plus a one-time token; pass that token on the SECOND call (with the same name) to actually create it. The token expires after a few minutes."`
}

// UpdateTargetGroupInput is the input for update_target_group. Partial update:
// name, algorithm and protocol are read and carried forward when omitted, because
// the API always receives them.
//
// targets REPLACES the whole backend list rather than adding to it, so omit it
// unless you intend to redefine the set — that is why it is not carried forward
// silently the way the scalar fields are.
type UpdateTargetGroupInput struct {
	TargetGroupID   string                   `json:"target_group_id" jsonschema:"the ID of the target group to update"`
	Name            *string                  `json:"name,omitempty" jsonschema:"a new name. Omit to keep the current one."`
	Algorithm       *string                  `json:"algorithm,omitempty" jsonschema:"a new distribution algorithm: ROUND_ROBIN, LEAST_CONNECTION, RANDOM or SOURCE_IP. Omit to keep the current one."`
	Protocol        *string                  `json:"protocol,omitempty" jsonschema:"a new protocol: HTTP or TCP. Omit to keep the current one."`
	ProtocolVersion *string                  `json:"protocol_version,omitempty" jsonschema:"for HTTP: HTTP1 or HTTP2"`
	Targets         []TargetGroupTargetInput `json:"targets,omitempty" jsonschema:"REPLACE the group's backends with this list. Include every backend the group should keep — any you omit is removed. Omit the field entirely to leave the current backends untouched."`
}

// DeleteTargetGroupInput is the input for delete_target_group. Two-phase confirmed.
type DeleteTargetGroupInput struct {
	TargetGroupID     string  `json:"target_group_id" jsonschema:"the ID of the target group to delete"`
	ConfirmationToken *string `json:"confirmation_token,omitempty" jsonschema:"leave empty on the FIRST call to receive a preview of the group and its backends plus a one-time token; pass that token on the SECOND call to actually delete. Any application load balancer HTTP rule forwarding to this group stops working. The token expires after a few minutes."`
}

// NatGatewayLanInput is one LAN a NAT gateway serves.
type NatGatewayLanInput struct {
	ID         int32    `json:"id" jsonschema:"the numeric LAN ID the gateway serves (the lan value from list_lans)"`
	GatewayIps []string `json:"gateway_ips,omitempty" jsonschema:"the gateway's addresses on that LAN, in CIDR form. Omit to have them assigned automatically."`
}

// CreateNatGatewayInput is the input for create_nat_gateway. Two-phase confirmed.
type CreateNatGatewayInput struct {
	DatacenterID      string               `json:"datacenter_id" jsonschema:"the ID of the data center to create the NAT gateway in"`
	Name              string               `json:"name" jsonschema:"the name of the new NAT gateway"`
	PublicIps         []string             `json:"public_ips" jsonschema:"the public IPv4 addresses the gateway translates to. They must come from a reserved IP block in the same location (see list_ip_blocks). At least one is required."`
	Lans              []NatGatewayLanInput `json:"lans,omitempty" jsonschema:"the private LANs the gateway serves. Without any, nothing routes through it."`
	ConfirmationToken *string              `json:"confirmation_token,omitempty" jsonschema:"leave empty on the FIRST call to receive a preview plus a one-time token; pass that token on the SECOND call (with the same datacenter_id and name) to actually create it. The token expires after a few minutes."`
}

// UpdateNatGatewayInput is the input for update_nat_gateway. Partial update: name
// and public_ips are read and carried forward when omitted, because the API always
// receives them.
type UpdateNatGatewayInput struct {
	DatacenterID string               `json:"datacenter_id" jsonschema:"the ID of the data center the NAT gateway is in"`
	NatGatewayID string               `json:"natgateway_id" jsonschema:"the ID of the NAT gateway to update"`
	Name         *string              `json:"name,omitempty" jsonschema:"a new name. Omit to keep the current one."`
	PublicIps    []string             `json:"public_ips,omitempty" jsonschema:"REPLACE the gateway's public addresses with this list. Include every address it should keep. Omit the field entirely to leave them untouched."`
	Lans         []NatGatewayLanInput `json:"lans,omitempty" jsonschema:"REPLACE the LANs the gateway serves with this list. Omit the field entirely to leave them untouched."`
}

// CreateNatGatewayRuleInput is the input for create_nat_gateway_rule. Two-phase
// confirmed. A NAT gateway does not translate anything until it has a rule.
type CreateNatGatewayRuleInput struct {
	DatacenterID         string  `json:"datacenter_id" jsonschema:"the ID of the data center"`
	NatGatewayID         string  `json:"natgateway_id" jsonschema:"the ID of the NAT gateway to add the rule to"`
	Name                 string  `json:"name" jsonschema:"the name of the new rule"`
	SourceSubnet         string  `json:"source_subnet" jsonschema:"the private source range whose outbound traffic is translated, in CIDR form, e.g. 10.0.1.0/24. Traffic from outside this range is not translated by this rule."`
	PublicIp             string  `json:"public_ip" jsonschema:"the public address to translate to. It must be one of the gateway's own public_ips."`
	Type                 *string `json:"type,omitempty" jsonschema:"the translation type. Only SNAT (source NAT, for outbound traffic) is supported."`
	Protocol             *string `json:"protocol,omitempty" jsonschema:"restrict the rule to one protocol: TCP, UDP, ICMP or ALL. Omit to match every protocol. A port range may only be set for TCP or UDP."`
	TargetSubnet         *string `json:"target_subnet,omitempty" jsonschema:"restrict the rule to traffic bound for this destination range, in CIDR form. Omit to match any destination."`
	TargetPortRangeStart *int32  `json:"target_port_range_start,omitempty" jsonschema:"first destination port the rule applies to (1-65535). Only valid with protocol TCP or UDP, and must be given together with target_port_range_end."`
	TargetPortRangeEnd   *int32  `json:"target_port_range_end,omitempty" jsonschema:"last destination port the rule applies to (1-65535). Only valid with protocol TCP or UDP, and must be given together with target_port_range_start."`
	ConfirmationToken    *string `json:"confirmation_token,omitempty" jsonschema:"leave empty on the FIRST call to receive a preview plus a one-time token; pass that token on the SECOND call (with the same parent IDs and name) to actually create the rule. The token expires after a few minutes."`
}

// UpdateNatGatewayRuleInput is the input for update_nat_gateway_rule. The API always
// receives name, source_subnet and public_ip, so those are read and carried forward
// when omitted.
type UpdateNatGatewayRuleInput struct {
	DatacenterID         string  `json:"datacenter_id" jsonschema:"the ID of the data center"`
	NatGatewayID         string  `json:"natgateway_id" jsonschema:"the ID of the NAT gateway"`
	RuleID               string  `json:"rule_id" jsonschema:"the ID of the rule to update"`
	Name                 *string `json:"name,omitempty" jsonschema:"a new name. Omit to keep the current one."`
	SourceSubnet         *string `json:"source_subnet,omitempty" jsonschema:"a new private source range in CIDR form. Omit to keep the current one; changing it changes which traffic is translated."`
	PublicIp             *string `json:"public_ip,omitempty" jsonschema:"a new public address to translate to; must be one of the gateway's public_ips. Omit to keep the current one."`
	Type                 *string `json:"type,omitempty" jsonschema:"the translation type. Only SNAT is supported."`
	Protocol             *string `json:"protocol,omitempty" jsonschema:"restrict the rule to TCP, UDP, ICMP or ALL"`
	TargetSubnet         *string `json:"target_subnet,omitempty" jsonschema:"restrict the rule to this destination range, in CIDR form"`
	TargetPortRangeStart *int32  `json:"target_port_range_start,omitempty" jsonschema:"first destination port (1-65535). TCP or UDP only, given together with the end."`
	TargetPortRangeEnd   *int32  `json:"target_port_range_end,omitempty" jsonschema:"last destination port (1-65535). TCP or UDP only, given together with the start."`
}

// DeleteNatGatewayRuleInput is the input for delete_nat_gateway_rule.
type DeleteNatGatewayRuleInput struct {
	DatacenterID      string  `json:"datacenter_id" jsonschema:"the ID of the data center"`
	NatGatewayID      string  `json:"natgateway_id" jsonschema:"the ID of the NAT gateway"`
	RuleID            string  `json:"rule_id" jsonschema:"the ID of the rule to delete"`
	ConfirmationToken *string `json:"confirmation_token,omitempty" jsonschema:"leave empty on the FIRST call to receive a preview of what the rule translates plus a one-time token; pass that token on the SECOND call to actually delete. Servers in the rule's source range lose the outbound translation it provided, which usually means losing internet access. The token expires after a few minutes."`
}

// DeleteNatGatewayInput is the input for delete_nat_gateway. Two-phase confirmed.
type DeleteNatGatewayInput struct {
	DatacenterID      string  `json:"datacenter_id" jsonschema:"the ID of the data center the NAT gateway is in"`
	NatGatewayID      string  `json:"natgateway_id" jsonschema:"the ID of the NAT gateway to delete"`
	ConfirmationToken *string `json:"confirmation_token,omitempty" jsonschema:"leave empty on the FIRST call to receive a blast-radius preview (its rules and the LANs it serves) plus a one-time token; pass that token on the SECOND call to actually delete. Servers on those LANs lose their outbound internet access. The token expires after a few minutes."`
}

type PccIDInput struct {
	PccID string `json:"pcc_id" jsonschema:"the ID of the private cross-connect"`
	Depth *int32 `json:"depth,omitempty" jsonschema:"nesting depth of returned objects (0-5)"`
}

type GpuIDInput struct {
	DatacenterID string `json:"datacenter_id" jsonschema:"the ID of the data center"`
	ServerID     string `json:"server_id" jsonschema:"the ID of the server"`
	GpuID        string `json:"gpu_id" jsonschema:"the ID of the GPU"`
	Depth        *int32 `json:"depth,omitempty" jsonschema:"nesting depth of returned objects (0-5)"`
}

type SecurityGroupIDInput struct {
	DatacenterID    string `json:"datacenter_id" jsonschema:"the ID of the data center"`
	SecurityGroupID string `json:"security_group_id" jsonschema:"the ID of the security group"`
	Depth           *int32 `json:"depth,omitempty" jsonschema:"nesting depth of returned objects (0-5, default 1 for list operations)"`
}

type SecurityGroupRuleIDInput struct {
	DatacenterID    string `json:"datacenter_id" jsonschema:"the ID of the data center"`
	SecurityGroupID string `json:"security_group_id" jsonschema:"the ID of the security group"`
	RuleID          string `json:"rule_id" jsonschema:"the ID of the security group rule"`
	Depth           *int32 `json:"depth,omitempty" jsonschema:"nesting depth of returned objects (0-5)"`
}

type RequestIDInput struct {
	RequestID string `json:"request_id" jsonschema:"the ID of the request"`
	Depth     *int32 `json:"depth,omitempty" jsonschema:"nesting depth of returned objects (0-5)"`
}

type TemplateIDInput struct {
	TemplateID string `json:"template_id" jsonschema:"the ID of the template"`
	Depth      *int32 `json:"depth,omitempty" jsonschema:"nesting depth of returned objects (0-5)"`
}

// GetContractInput is the input for get_contract (no required ID fields).
type GetContractInput struct {
	Depth *int32 `json:"depth,omitempty" jsonschema:"nesting depth of returned objects (0-5)"`
}

// List*Input types for list_ tools (account-scoped and resource-scoped variants).

type ListIPBlocksInput struct {
	Depth   *int32            `json:"depth,omitempty"   jsonschema:"nesting depth of returned objects (0-5, default 1)"`
	Filters map[string]string `json:"filters,omitempty" jsonschema:"server-side filters as property→value pairs (contains match); e.g. {\"name\":\"prod\",\"location\":\"de/fra\"} If the result is empty, retry without filters — a filter typo or mismatch silently returns nothing."`
}

type ListTargetGroupsInput struct {
	Depth   *int32            `json:"depth,omitempty"   jsonschema:"nesting depth of returned objects (0-5, default 1)"`
	Filters map[string]string `json:"filters,omitempty" jsonschema:"server-side filters as property→value pairs (contains match); e.g. {\"name\":\"prod\"} If the result is empty, retry without filters — a filter typo or mismatch silently returns nothing."`
}

type ListPrivateCrossConnectsInput struct {
	Depth   *int32            `json:"depth,omitempty"   jsonschema:"nesting depth of returned objects (0-5, default 1)"`
	Filters map[string]string `json:"filters,omitempty" jsonschema:"server-side filters as property→value pairs (contains match); e.g. {\"name\":\"prod\"} If the result is empty, retry without filters — a filter typo or mismatch silently returns nothing."`
}

type ListTemplatesInput struct {
	Depth   *int32            `json:"depth,omitempty"   jsonschema:"nesting depth of returned objects (0-5, default 1)"`
	Filters map[string]string `json:"filters,omitempty" jsonschema:"server-side filters as property→value pairs (contains match); e.g. {\"name\":\"prod\",\"cpuFamily\":\"INTEL_SKYLAKE\"} If the result is empty, retry without filters — a filter typo or mismatch silently returns nothing."`
}

type ListImagesInput struct {
	Depth   *int32            `json:"depth,omitempty"   jsonschema:"nesting depth of returned objects (0-5, default 1)"`
	Filters map[string]string `json:"filters,omitempty" jsonschema:"server-side filters as property→value pairs (contains match); e.g. {\"name\":\"ubuntu\",\"imageType\":\"HDD\",\"licenceType\":\"LINUX\",\"location\":\"de/fra\"} If the result is empty, retry without filters — a filter typo or mismatch silently returns nothing."`
}

type ListLocationsInput struct {
	Depth   *int32            `json:"depth,omitempty"   jsonschema:"nesting depth of returned objects (0-5, default 1)"`
	Filters map[string]string `json:"filters,omitempty" jsonschema:"server-side filters as property→value pairs (contains match); e.g. {\"name\":\"Frankfurt\"} If the result is empty, retry without filters — a filter typo or mismatch silently returns nothing."`
}

type ListSnapshotsInput struct {
	Depth   *int32            `json:"depth,omitempty"   jsonschema:"nesting depth of returned objects (0-5, default 1)"`
	Filters map[string]string `json:"filters,omitempty" jsonschema:"server-side filters as property→value pairs (contains match); e.g. {\"name\":\"prod\",\"location\":\"de/fra\"} If the result is empty, retry without filters — a filter typo or mismatch silently returns nothing."`
}

type ListRequestsInput struct {
	Depth   *int32            `json:"depth,omitempty"   jsonschema:"nesting depth of returned objects (0-5, default 1)"`
	Filters map[string]string `json:"filters,omitempty" jsonschema:"server-side filters as property→value pairs (contains match); e.g. {\"method\":\"POST\",\"requestStatus\":\"DONE\",\"createdBy\":\"user@example.com\"} If the result is empty, retry without filters — a filter typo or mismatch silently returns nothing."`
}

type ListK8sClustersInput struct {
	Depth *int32 `json:"depth,omitempty" jsonschema:"nesting depth of returned objects (0-5, default 1)"`
}

type ListInDatacenterInput struct {
	DatacenterID string            `json:"datacenter_id" jsonschema:"the ID of the data center"`
	Depth        *int32            `json:"depth,omitempty"   jsonschema:"nesting depth of returned objects (0-5, default 1)"`
	Filters      map[string]string `json:"filters,omitempty" jsonschema:"server-side filters as property→value pairs (contains match); e.g. {\"name\":\"prod\"} If the result is empty, retry without filters — a filter typo or mismatch silently returns nothing."`
}

type ListInServerInput struct {
	DatacenterID string            `json:"datacenter_id" jsonschema:"the ID of the data center"`
	ServerID     string            `json:"server_id"     jsonschema:"the ID of the server"`
	Depth        *int32            `json:"depth,omitempty"   jsonschema:"nesting depth of returned objects (0-5, default 1)"`
	Filters      map[string]string `json:"filters,omitempty" jsonschema:"server-side filters as property→value pairs (contains match); e.g. {\"name\":\"prod\"} If the result is empty, retry without filters — a filter typo or mismatch silently returns nothing."`
}

type ListNatGatewayRulesInput struct {
	DatacenterID string            `json:"datacenter_id"  jsonschema:"the ID of the data center"`
	NatGatewayID string            `json:"nat_gateway_id" jsonschema:"the ID of the NAT gateway"`
	Depth        *int32            `json:"depth,omitempty"   jsonschema:"nesting depth of returned objects (0-5, default 1)"`
	Filters      map[string]string `json:"filters,omitempty" jsonschema:"server-side filters as property→value pairs (contains match); e.g. {\"name\":\"prod\",\"type\":\"SNAT\"} If the result is empty, retry without filters — a filter typo or mismatch silently returns nothing."`
}

type ListLanNicsInput struct {
	DatacenterID string            `json:"datacenter_id" jsonschema:"the ID of the data center"`
	LanID        string            `json:"lan_id"        jsonschema:"the ID of the LAN"`
	Depth        *int32            `json:"depth,omitempty"   jsonschema:"nesting depth of returned objects (0-5, default 1)"`
	Filters      map[string]string `json:"filters,omitempty" jsonschema:"server-side filters as property→value pairs (contains match); e.g. {\"name\":\"prod\"} If the result is empty, retry without filters — a filter typo or mismatch silently returns nothing."`
}

type ListAlbForwardingRulesInput struct {
	DatacenterID              string            `json:"datacenter_id"               jsonschema:"the ID of the data center"`
	ApplicationLoadBalancerID string            `json:"application_loadbalancer_id" jsonschema:"the ID of the application load balancer"`
	Depth                     *int32            `json:"depth,omitempty"   jsonschema:"nesting depth of returned objects (0-5, default 1)"`
	Filters                   map[string]string `json:"filters,omitempty" jsonschema:"server-side filters as property→value pairs (contains match); e.g. {\"name\":\"prod\"} If the result is empty, retry without filters — a filter typo or mismatch silently returns nothing."`
}

type ListNlbForwardingRulesInput struct {
	DatacenterID          string            `json:"datacenter_id"           jsonschema:"the ID of the data center"`
	NetworkLoadBalancerID string            `json:"network_loadbalancer_id" jsonschema:"the ID of the network load balancer"`
	Depth                 *int32            `json:"depth,omitempty"   jsonschema:"nesting depth of returned objects (0-5, default 1)"`
	Filters               map[string]string `json:"filters,omitempty" jsonschema:"server-side filters as property→value pairs (contains match); e.g. {\"name\":\"prod\"} If the result is empty, retry without filters — a filter typo or mismatch silently returns nothing."`
}

type ListLoadBalancerNicsInput struct {
	DatacenterID   string            `json:"datacenter_id"   jsonschema:"the ID of the data center"`
	LoadBalancerID string            `json:"loadbalancer_id" jsonschema:"the ID of the load balancer"`
	Depth          *int32            `json:"depth,omitempty"   jsonschema:"nesting depth of returned objects (0-5, default 1)"`
	Filters        map[string]string `json:"filters,omitempty" jsonschema:"server-side filters as property→value pairs (contains match); e.g. {\"name\":\"prod\"} If the result is empty, retry without filters — a filter typo or mismatch silently returns nothing."`
}

type ListSecurityGroupRulesInput struct {
	DatacenterID    string            `json:"datacenter_id"     jsonschema:"the ID of the data center"`
	SecurityGroupID string            `json:"security_group_id" jsonschema:"the ID of the security group"`
	Depth           *int32            `json:"depth,omitempty"   jsonschema:"nesting depth of returned objects (0-5, default 1)"`
	Filters         map[string]string `json:"filters,omitempty" jsonschema:"server-side filters as property→value pairs (contains match); e.g. {\"name\":\"prod\",\"type\":\"INGRESS\",\"direction\":\"INBOUND\"} If the result is empty, retry without filters — a filter typo or mismatch silently returns nothing."`
}

type ListFirewallRulesInput struct {
	DatacenterID string            `json:"datacenter_id" jsonschema:"the ID of the data center"`
	ServerID     string            `json:"server_id"     jsonschema:"the ID of the server"`
	NicID        string            `json:"nic_id"        jsonschema:"the ID of the network interface"`
	Depth        *int32            `json:"depth,omitempty"   jsonschema:"nesting depth of returned objects (0-5, default 1)"`
	Filters      map[string]string `json:"filters,omitempty" jsonschema:"server-side filters as property→value pairs (contains match); e.g. {\"name\":\"prod\",\"type\":\"INGRESS\",\"direction\":\"INBOUND\"} If the result is empty, retry without filters — a filter typo or mismatch silently returns nothing."`
}

// DNS input types

type ZoneIDInput struct {
	ZoneID string `json:"zone_id" jsonschema:"the ID of the DNS zone"`
}

type RecordIDInput struct {
	ZoneID   string `json:"zone_id" jsonschema:"the ID of the DNS zone"`
	RecordID string `json:"record_id" jsonschema:"the ID of the DNS record"`
}

type ReverseRecordIDInput struct {
	ReverseRecordID string `json:"reverse_record_id" jsonschema:"the ID of the reverse DNS record"`
}

type SecondaryZoneIDInput struct {
	SecondaryZoneID string `json:"secondary_zone_id" jsonschema:"the ID of the secondary DNS zone"`
}

// DNS write input types. Every DNS update is a PUT, so the identity fields
// (zone_name, a record's name and type, a reverse record's ip) are read from the
// resource and carried forward rather than accepted here — see tools/dns.

type CreateDnsZoneInput struct {
	ZoneName          string  `json:"zone_name" jsonschema:"the zone name, e.g. example.com"`
	Description       *string `json:"description,omitempty" jsonschema:"free-text description of what the zone is for"`
	Enabled           *bool   `json:"enabled,omitempty" jsonschema:"whether the zone answers lookups. Defaults to true."`
	ConfirmationToken *string `json:"confirmation_token,omitempty" jsonschema:"omit on the first call to get a preview and a one-time token; pass that token on the second call to create the zone"`
}

type UpdateDnsZoneInput struct {
	ZoneID      string  `json:"zone_id" jsonschema:"the ID of the zone to update"`
	Description *string `json:"description,omitempty" jsonschema:"a new description. Omit to keep the current one."`
	Enabled     *bool   `json:"enabled,omitempty" jsonschema:"set false to stop the zone answering lookups, true to resume. Omit to keep the current setting."`
}

type DeleteDnsZoneInput struct {
	ZoneID            string  `json:"zone_id" jsonschema:"the ID of the zone to delete"`
	ConfirmationToken *string `json:"confirmation_token,omitempty" jsonschema:"omit on the first call to get a blast-radius preview and a one-time token; pass that token on the second call to delete"`
}

type ImportDnsZoneFileInput struct {
	ZoneID            string  `json:"zone_id" jsonschema:"the ID of the zone to overwrite"`
	ZoneFile          string  `json:"zone_file" jsonschema:"the zone file in BIND format (RFC 1035). Every record currently in the zone is REPLACED by the records in this file. SOA and NS records may be present but are ignored."`
	ConfirmationToken *string `json:"confirmation_token,omitempty" jsonschema:"omit on the first call to see how many existing records would be replaced and get a one-time token; pass that token on the second call to import"`
}

type CreateDnsRecordInput struct {
	ZoneID            string  `json:"zone_id" jsonschema:"the ID of the zone to add the record to"`
	Name              string  `json:"name" jsonschema:"the record name relative to the zone, e.g. www. Use an empty string for a record on the zone apex (example.com itself)."`
	Type              string  `json:"type" jsonschema:"the record type: A, AAAA, CNAME, ALIAS, MX, NS, SRV, TXT, CAA, SSHFP, TLSA, SMIMEA, DS, HTTPS, SVCB, OPENPGPKEY, CERT, URI, RP or LOC"`
	Content           string  `json:"content" jsonschema:"the record value, e.g. 192.0.2.1 for an A record or mail.example.com for an MX record"`
	Ttl               *int32  `json:"ttl,omitempty" jsonschema:"time to live in seconds, between 60 and 604800. Defaults to 3600."`
	Priority          *int32  `json:"priority,omitempty" jsonschema:"priority between 0 and 65535. Required for MX, SRV and URI records; ignored for every other type."`
	Enabled           *bool   `json:"enabled,omitempty" jsonschema:"whether the record is visible for lookup. Defaults to true."`
	ConfirmationToken *string `json:"confirmation_token,omitempty" jsonschema:"omit on the first call to get a preview and a one-time token; pass that token on the second call to create the record"`
}

// UpdateDnsRecordInput omits name and type: the endpoint is a PUT, and changing
// either is a delete-and-recreate that can trip the API's CNAME/SPF conflict rules.
type UpdateDnsRecordInput struct {
	ZoneID   string  `json:"zone_id" jsonschema:"the ID of the zone the record belongs to"`
	RecordID string  `json:"record_id" jsonschema:"the ID of the record to update"`
	Content  *string `json:"content,omitempty" jsonschema:"a new record value. Omit to keep the current one."`
	Ttl      *int32  `json:"ttl,omitempty" jsonschema:"a new time to live in seconds, between 60 and 604800. Omit to keep the current one."`
	Priority *int32  `json:"priority,omitempty" jsonschema:"a new priority between 0 and 65535, for MX, SRV and URI records. Omit to keep the current one."`
	Enabled  *bool   `json:"enabled,omitempty" jsonschema:"set false to hide the record from lookups, true to publish it. Omit to keep the current setting."`
}

type DeleteDnsRecordInput struct {
	ZoneID            string  `json:"zone_id" jsonschema:"the ID of the zone the record belongs to"`
	RecordID          string  `json:"record_id" jsonschema:"the ID of the record to delete"`
	ConfirmationToken *string `json:"confirmation_token,omitempty" jsonschema:"omit on the first call to get a preview and a one-time token; pass that token on the second call to delete"`
}

type CreateDnsSecondaryZoneInput struct {
	ZoneName          string   `json:"zone_name" jsonschema:"the zone name, e.g. example.com"`
	PrimaryIps        []string `json:"primary_ips" jsonschema:"IPv4 or IPv6 addresses of the primary nameservers to transfer the zone from. At least one, no duplicates. Whitelist IONOS's notify sources on your primaries: 212.227.123.25 and 2001:8d8:fe:53::5cd:25."`
	Description       *string  `json:"description,omitempty" jsonschema:"free-text description of what the zone is for"`
	ConfirmationToken *string  `json:"confirmation_token,omitempty" jsonschema:"omit on the first call to get a preview and a one-time token; pass that token on the second call to create the zone"`
}

type UpdateDnsSecondaryZoneInput struct {
	SecondaryZoneID string   `json:"secondary_zone_id" jsonschema:"the ID of the secondary zone to update"`
	PrimaryIps      []string `json:"primary_ips,omitempty" jsonschema:"REPLACE the primary nameserver IPs with this list. Include every IP that should remain — any you leave out stops being used. Omit the field to keep the current list."`
	Description     *string  `json:"description,omitempty" jsonschema:"a new description. Omit to keep the current one."`
}

type DeleteDnsSecondaryZoneInput struct {
	SecondaryZoneID   string  `json:"secondary_zone_id" jsonschema:"the ID of the secondary zone to delete"`
	ConfirmationToken *string `json:"confirmation_token,omitempty" jsonschema:"omit on the first call to get a preview and a one-time token; pass that token on the second call to delete"`
}

type CreateDnsReverseRecordInput struct {
	Name              string  `json:"name" jsonschema:"the hostname the IP should resolve back to, e.g. mail.example.com"`
	Ip                string  `json:"ip" jsonschema:"the IPv4 or IPv6 address to create the reverse record for. It must be an IP your contract owns — an IPv4 from one of your IP blocks or an IPv6 from a VDC."`
	Description       *string `json:"description,omitempty" jsonschema:"free-text description of what the record is for"`
	ConfirmationToken *string `json:"confirmation_token,omitempty" jsonschema:"omit on the first call to get a preview and a one-time token; pass that token on the second call to create the record"`
}

// UpdateDnsReverseRecordInput omits ip: it identifies which address the record
// covers, and pointing the record at a different one is a create, not an update.
type UpdateDnsReverseRecordInput struct {
	ReverseRecordID string  `json:"reverse_record_id" jsonschema:"the ID of the reverse record to update"`
	Name            *string `json:"name,omitempty" jsonschema:"a new hostname for the IP to resolve back to. Omit to keep the current one."`
	Description     *string `json:"description,omitempty" jsonschema:"a new description. Omit to keep the current one."`
}

type DeleteDnsReverseRecordInput struct {
	ReverseRecordID   string  `json:"reverse_record_id" jsonschema:"the ID of the reverse record to delete"`
	ConfirmationToken *string `json:"confirmation_token,omitempty" jsonschema:"omit on the first call to get a preview and a one-time token; pass that token on the second call to delete"`
}

type CreateDnsDnssecKeyInput struct {
	ZoneID          string  `json:"zone_id" jsonschema:"the ID of the zone to sign"`
	Validity        int32   `json:"validity" jsonschema:"signature validity in days, between 90 and 365"`
	Algorithm       *string `json:"algorithm,omitempty" jsonschema:"signing algorithm. RSASHA256 is the only value the API accepts, and is the default."`
	KskBits         *int32  `json:"ksk_bits,omitempty" jsonschema:"key signing key length: 1024, 2048 or 4096. Must be greater than or equal to zsk_bits. Defaults to 4096."`
	ZskBits         *int32  `json:"zsk_bits,omitempty" jsonschema:"zone signing key length: 1024, 2048 or 4096. Defaults to 2048."`
	NsecMode        *string `json:"nsec_mode,omitempty" jsonschema:"proof-of-nonexistence mode, NSEC or NSEC3. Defaults to NSEC3."`
	Nsec3Iterations *int32  `json:"nsec3_iterations,omitempty" jsonschema:"NSEC3 hash iterations, between 0 and 50. Defaults to 0, which is what RFC 9276 recommends. Ignored for NSEC mode but still sent, because the API requires the field."`
	Nsec3SaltBits   *int32  `json:"nsec3_salt_bits,omitempty" jsonschema:"NSEC3 salt length in bits, between 64 and 128 and a multiple of 8. Defaults to 64. Ignored for NSEC mode but still sent, because the API requires the field."`

	ConfirmationToken *string `json:"confirmation_token,omitempty" jsonschema:"omit on the first call to get a preview and a one-time token; pass that token on the second call to enable DNSSEC"`
}

type DeleteDnsDnssecKeyInput struct {
	ZoneID            string  `json:"zone_id" jsonschema:"the ID of the zone to stop signing"`
	ConfirmationToken *string `json:"confirmation_token,omitempty" jsonschema:"omit on the first call to get a preview and a one-time token; pass that token on the second call to disable DNSSEC"`
}

type StartDnsZoneTransferInput struct {
	SecondaryZoneID string `json:"secondary_zone_id" jsonschema:"the ID of the secondary zone to transfer"`
}

// Billing input types
// Most billing tools require a contract number — call get_billing_profile first to get it.
// Exception: list_billing_invoices_by_period is contract-agnostic (the underlying API endpoint does not accept a contract parameter).

type BillingContractInput struct {
	Contract int32 `json:"contract" jsonschema:"contract number from get_billing_profile"`
}

type BillingContractPeriodInput struct {
	Contract int32  `json:"contract" jsonschema:"contract number from get_billing_profile"`
	Period   string `json:"period" jsonschema:"billing period in YYYY-MM format (e.g. 2026-04). Maximum one month per request — for wider ranges call once per month"`
}

type BillingInvoiceIDInput struct {
	Contract  int32  `json:"contract" jsonschema:"contract number from get_billing_profile"`
	InvoiceID string `json:"invoice_id" jsonschema:"the invoice ID (e.g. GY00111111)"`
}

type BillingDatacenterInput struct {
	Contract     int32  `json:"contract" jsonschema:"contract number from get_billing_profile"`
	DatacenterID string `json:"datacenter_id" jsonschema:"the VDC UUID"`
}

// BillingUtilizationInput is the input for list_billing_utilization.
// Compaction flags reduce response size — zero-quantity meters are dropped by default.
type BillingUtilizationInput struct {
	Contract     int32    `json:"contract" jsonschema:"contract number from get_billing_profile"`
	IncludeZero  *bool    `json:"include_zero,omitempty" jsonschema:"include meters with quantity 0 (default false); set true to find existing resources that didn't consume in the window"`
	GroupBy      *string  `json:"group_by,omitempty" jsonschema:"aggregation level: omitted or '' = per-resource (default), 'meter' = sum per SKU per datacenter, 'datacenter' = sum per type per datacenter — coarser groupings shrink output but lose detail"`
	DatacenterID *string  `json:"datacenter_id,omitempty" jsonschema:"scope to a single datacenter (VDC UUID)"`
	MeterTypes   []string `json:"meter_types,omitempty" jsonschema:"filter to these meter type categories only (client-side); e.g. ['DBAAS','DNS','SERVER']"`
	Regions      []string `json:"regions,omitempty" jsonschema:"filter to these regions only (client-side); e.g. ['de/fra','es/vit']"`
	TopN         *int32   `json:"top_n,omitempty" jsonschema:"return only the N largest meters globally, sorted by quantity desc — flat list with dc_id/dc_name on each row, datacenters[] omitted; ideal for cost audits on contracts with many datacenters. When combined with group_by='datacenter', top_meters[] rows have no meter_id (type+unit aggregates); with group_by='meter', meter_id is the SKU"`
}

type BillingUtilizationPeriodInput struct {
	Contract     int32    `json:"contract" jsonschema:"contract number from get_billing_profile"`
	Period       string   `json:"period" jsonschema:"billing period in YYYY-MM format (e.g. 2026-04). Maximum one month per request — for wider ranges call once per month"`
	IncludeZero  *bool    `json:"include_zero,omitempty" jsonschema:"include meters with quantity 0 (default false)"`
	GroupBy      *string  `json:"group_by,omitempty" jsonschema:"aggregation level: omitted or '' = per-resource (default), 'meter' = sum per SKU per datacenter, 'datacenter' = sum per type per datacenter"`
	DatacenterID *string  `json:"datacenter_id,omitempty" jsonschema:"scope to a single datacenter (VDC UUID)"`
	MeterTypes   []string `json:"meter_types,omitempty" jsonschema:"filter to these meter type categories only (client-side); e.g. ['DBAAS','DNS','SERVER']"`
	Regions      []string `json:"regions,omitempty" jsonschema:"filter to these regions only (client-side); e.g. ['de/fra']"`
	TopN         *int32   `json:"top_n,omitempty" jsonschema:"return only the N largest meters globally, sorted by quantity desc (flat list, datacenters[] omitted). When combined with group_by='datacenter', top_meters[] rows have no meter_id"`
}

type BillingUtilizationDateInput struct {
	Contract     int32    `json:"contract" jsonschema:"contract number from get_billing_profile"`
	Date         string   `json:"date" jsonschema:"date in YYYY-MM-DD format (e.g. 2026-04-15)"`
	IncludeZero  *bool    `json:"include_zero,omitempty" jsonschema:"include meters with quantity 0 (default false)"`
	GroupBy      *string  `json:"group_by,omitempty" jsonschema:"aggregation level: omitted or '' = per-resource (default), 'meter' = sum per SKU per datacenter, 'datacenter' = sum per type per datacenter"`
	DatacenterID *string  `json:"datacenter_id,omitempty" jsonschema:"scope to a single datacenter (VDC UUID)"`
	MeterTypes   []string `json:"meter_types,omitempty" jsonschema:"filter to these meter type categories only (client-side); e.g. ['DBAAS']"`
	Regions      []string `json:"regions,omitempty" jsonschema:"filter to these regions only (client-side); e.g. ['de/fra']"`
	TopN         *int32   `json:"top_n,omitempty" jsonschema:"return only the N largest meters globally, sorted by quantity desc (flat list, datacenters[] omitted). When combined with group_by='datacenter', top_meters[] rows have no meter_id"`
}

// BillingUsageInput is the input for list_billing_usage.
// UsageMeter has fewer fields than UtilizationMeter — type/region/resource filters do not apply.
type BillingUsageInput struct {
	Contract     int32   `json:"contract" jsonschema:"contract number from get_billing_profile"`
	IncludeZero  *bool   `json:"include_zero,omitempty" jsonschema:"include meters with quantity 0 (default false); set true to find datacenters with metered SKUs that didn't consume"`
	DatacenterID *string `json:"datacenter_id,omitempty" jsonschema:"scope to a single datacenter (VDC UUID)"`
}

type BillingUsageDatacenterInput struct {
	Contract     int32  `json:"contract" jsonschema:"contract number from get_billing_profile"`
	DatacenterID string `json:"datacenter_id" jsonschema:"the VDC UUID"`
	IncludeZero  *bool  `json:"include_zero,omitempty" jsonschema:"include meters with quantity 0 (default false)"`
}

type BillingDateInput struct {
	Contract int32  `json:"contract" jsonschema:"contract number from get_billing_profile"`
	Date     string `json:"date" jsonschema:"date in YYYY-MM-DD format (e.g. 2026-04-15)"`
}

type BillingPeriodOnlyInput struct {
	Period string `json:"period" jsonschema:"billing period in YYYY-MM format (e.g. 2026-04)"`
}

type BillingProductsInput struct {
	Contract int32  `json:"contract" jsonschema:"contract number from get_billing_profile"`
	Filter   string `json:"filter" jsonschema:"keyword to filter products by description (e.g. 'RAM', 'Kubernetes', 'Postgres', 'storage'). Use broad terms to find relevant pricing"`
}

// Object Storage input types

type ObjectStorageBucketInput struct {
	Bucket string `json:"bucket" jsonschema:"the name of the object storage bucket"`
}

type ObjectStorageObjectInput struct {
	Bucket string `json:"bucket" jsonschema:"the name of the object storage bucket"`
	Key    string `json:"key" jsonschema:"the object key (path within the bucket)"`
}

type ObjectStorageListObjectsInput struct {
	Bucket            string  `json:"bucket" jsonschema:"the name of the object storage bucket"`
	Prefix            *string `json:"prefix,omitempty" jsonschema:"optional key prefix to filter results (e.g. 'images/' to list only objects under that path)"`
	ContinuationToken *string `json:"continuation_token,omitempty" jsonschema:"optional pagination token returned by a previous list operation to continue listing objects"`
	MaxKeys           *int32  `json:"max_keys,omitempty" jsonschema:"optional maximum number of objects to return in a single page"`
}

type AccessKeyIDInput struct {
	AccessKeyID string `json:"access_key_id" jsonschema:"the ID of the object storage access key"`
}

type ObjectStorageRegionInput struct {
	Region string `json:"region" jsonschema:"the region name (e.g. eu-central-3)"`
}

type ObjectStorageListObjectVersionsInput struct {
	Bucket string  `json:"bucket" jsonschema:"the name of the object storage bucket"`
	Prefix *string `json:"prefix,omitempty" jsonschema:"optional key prefix to filter versions"`
}

// Activity Log input types

type ActivityLogQueryInput struct {
	Contract             int32    `json:"contract" jsonschema:"the contract number whose activity log to query; reseller/partner users get IDs from list_activitylog_contracts, single-contract users read it from their JWT"`
	DateStart            *string  `json:"date_start,omitempty" jsonschema:"inclusive start date YYYY-MM-DD; defaults to 7 days ago when omitted"`
	DateEnd              *string  `json:"date_end,omitempty" jsonschema:"inclusive end date YYYY-MM-DD; defaults to today when omitted; maximum range is 90 days"`
	Offset               *int32   `json:"offset,omitempty" jsonschema:"0-based pagination offset"`
	Limit                *int32   `json:"limit,omitempty" jsonschema:"max events to return; defaults to 25; increase only when the user explicitly asks for bulk data"`
	User                 *string  `json:"user,omitempty" jsonschema:"filter by username (client-side); e.g. 'ionosctl-v6@cloud.ionos.com' — drastically reduces output when investigating a specific user"`
	EventTypes           []string `json:"event_types,omitempty" jsonschema:"filter to these event types only (client-side); e.g. ['Error','RequestAccepted'] — omit Provision and RequestStatusUpdate to cut ~65% of typical log volume"`
	IncludeStatusUpdates *bool    `json:"include_status_updates,omitempty" jsonschema:"include RequestStatusUpdate events (default false); these are async provisioning echoes that account for ~55% of log volume and are rarely useful"`
}

// Certificate Manager input types

type CertificateIDInput struct {
	CertificateID string `json:"certificate_id" jsonschema:"the ID of the certificate"`
}

type AutoCertificateIDInput struct {
	AutoCertificateID string `json:"auto_certificate_id" jsonschema:"the ID of the auto-certificate"`
}

type ProviderIDInput struct {
	ProviderID string `json:"provider_id" jsonschema:"the ID of the certificate provider"`
}

// Certificate Manager write input types. All three PATCH endpoints accept only the
// spec's PatchName, so an update renames the resource and changes nothing else.

type CreateCertCertificateInput struct {
	Name              string  `json:"name" jsonschema:"a name for the certificate, used for management purposes only; it does not have to match the common name"`
	Certificate       string  `json:"certificate" jsonschema:"the certificate body in PEM format, including the BEGIN CERTIFICATE and END CERTIFICATE lines"`
	CertificateChain  string  `json:"certificate_chain" jsonschema:"the chain of intermediate CA certificates in PEM format, leaf-issuer first. Required by the API; the root CA does not need to be included."`
	PrivateKey        string  `json:"private_key" jsonschema:"the unencrypted private key matching the certificate, in PEM format. Write-only: the API never returns it and it is never echoed in a preview."`
	ConfirmationToken *string `json:"confirmation_token,omitempty" jsonschema:"omit on the first call to get a preview and a one-time token; pass that token on the second call to upload the certificate"`
}

type UpdateCertCertificateInput struct {
	CertificateID string `json:"certificate_id" jsonschema:"the ID of the certificate to rename"`
	Name          string `json:"name" jsonschema:"the new certificate name"`
}

type DeleteCertCertificateInput struct {
	CertificateID     string  `json:"certificate_id" jsonschema:"the ID of the certificate to delete"`
	ConfirmationToken *string `json:"confirmation_token,omitempty" jsonschema:"omit on the first call to get a preview and a one-time token; pass that token on the second call to delete"`
}

type CreateCertAutoCertificateInput struct {
	ProviderID              string   `json:"provider_id" jsonschema:"the ID of the certificate provider that will issue the certificates; list them with list_cert_providers"`
	CommonName              string   `json:"common_name" jsonschema:"the common name (DNS) to issue the certificate for, e.g. www.example.com. It must belong to a zone hosted in IONOS Cloud DNS, or issuing fails."`
	KeyAlgorithm            string   `json:"key_algorithm" jsonschema:"the key algorithm: rsa2048, rsa3072 or rsa4096"`
	Name                    string   `json:"name" jsonschema:"a name for the auto-certificate, used for management purposes only"`
	SubjectAlternativeNames []string `json:"subject_alternative_names,omitempty" jsonschema:"additional DNS names to add to the issued certificate. Each one must also belong to a zone hosted in IONOS Cloud DNS."`
	ConfirmationToken       *string  `json:"confirmation_token,omitempty" jsonschema:"omit on the first call to get a preview and a one-time token; pass that token on the second call to create the auto-certificate"`
}

type UpdateCertAutoCertificateInput struct {
	AutoCertificateID string `json:"auto_certificate_id" jsonschema:"the ID of the auto-certificate to rename"`
	Name              string `json:"name" jsonschema:"the new auto-certificate name"`
}

type DeleteCertAutoCertificateInput struct {
	AutoCertificateID string  `json:"auto_certificate_id" jsonschema:"the ID of the auto-certificate to delete"`
	ConfirmationToken *string `json:"confirmation_token,omitempty" jsonschema:"omit on the first call to get a blast-radius preview and a one-time token; pass that token on the second call to delete"`
}

// CertExternalAccountBindingInput carries both EAB halves as required fields, so a
// key ID can never be sent without its secret.
type CertExternalAccountBindingInput struct {
	KeyID     string `json:"key_id" jsonschema:"the external account binding key ID issued by the ACME provider"`
	KeySecret string `json:"key_secret" jsonschema:"the external account binding secret (usually base64url). Write-only: the API never returns it and it is never echoed in a preview."`
}

type CreateCertProviderInput struct {
	Name                   string                           `json:"name" jsonschema:"a name for the provider, used for management purposes only, e.g. Let's Encrypt"`
	Email                  string                           `json:"email" jsonschema:"the email address registered with the ACME provider as the certificate requester; it receives expiry notices"`
	Server                 string                           `json:"server" jsonschema:"the ACME directory URL, e.g. https://acme-v02.api.letsencrypt.org/directory"`
	ExternalAccountBinding *CertExternalAccountBindingInput `json:"external_account_binding,omitempty" jsonschema:"external account binding credentials, for ACME providers that require an account to be pre-registered (e.g. ZeroSSL, Google Trust Services). Omit for Let's Encrypt."`
	ConfirmationToken      *string                          `json:"confirmation_token,omitempty" jsonschema:"omit on the first call to get a preview and a one-time token; pass that token on the second call to create the provider"`
}

type UpdateCertProviderInput struct {
	ProviderID string `json:"provider_id" jsonschema:"the ID of the certificate provider to rename"`
	Name       string `json:"name" jsonschema:"the new provider name"`
}

type DeleteCertProviderInput struct {
	ProviderID        string  `json:"provider_id" jsonschema:"the ID of the certificate provider to delete"`
	ConfirmationToken *string `json:"confirmation_token,omitempty" jsonschema:"omit on the first call to get a blast-radius preview and a one-time token; pass that token on the second call to delete"`
}

// Kubernetes input types

type K8sClusterIDInput struct {
	K8sClusterID string `json:"k8s_cluster_id" jsonschema:"the ID of the Kubernetes cluster"`
	Depth        *int32 `json:"depth,omitempty" jsonschema:"nesting depth of returned objects (0-5, default 1 for list operations)"`
}

type K8sNodepoolIDInput struct {
	K8sClusterID string `json:"k8s_cluster_id" jsonschema:"the ID of the Kubernetes cluster"`
	NodepoolID   string `json:"nodepool_id" jsonschema:"the ID of the Kubernetes node pool"`
	Depth        *int32 `json:"depth,omitempty" jsonschema:"nesting depth of returned objects (0-5, default 1 for list operations)"`
}

type K8sNodeIDInput struct {
	K8sClusterID string `json:"k8s_cluster_id" jsonschema:"the ID of the Kubernetes cluster"`
	NodepoolID   string `json:"nodepool_id" jsonschema:"the ID of the Kubernetes node pool"`
	NodeID       string `json:"node_id" jsonschema:"the ID of the Kubernetes node"`
	Depth        *int32 `json:"depth,omitempty" jsonschema:"nesting depth of returned objects (0-5)"`
}

// Kubernetes write input types. Both update endpoints are PUT, not PATCH, so the
// update tools read the resource first and carry forward whatever is not supplied.

// K8sMaintenanceWindowInput is the weekly maintenance window.
type K8sMaintenanceWindowInput struct {
	DayOfTheWeek string `json:"day_of_the_week" jsonschema:"the weekday the window opens: Monday, Tuesday, Wednesday, Thursday, Friday, Saturday or Sunday"`
	Time         string `json:"time" jsonschema:"the time the window opens, as HH:mm:ss (e.g. 03:30:00), optionally suffixed with Z. The actual start may vary by up to 15 minutes."`
}

// K8sAutoScalingInput turns the cluster autoscaler on. Once set it owns the node count.
type K8sAutoScalingInput struct {
	MinNodeCount int32 `json:"min_node_count" jsonschema:"the fewest worker nodes the autoscaler may scale down to. Must be at least 1."`
	MaxNodeCount int32 `json:"max_node_count" jsonschema:"the most worker nodes the autoscaler may scale up to. Must be greater than or equal to min_node_count."`
}

// K8sNodePoolLanRouteInput is one static route. Both fields are optional per the spec.
type K8sNodePoolLanRouteInput struct {
	Network   *string `json:"network,omitempty" jsonschema:"the IPv4 or IPv6 CIDR to route via this interface (e.g. 10.0.0.0/24)"`
	GatewayIp *string `json:"gateway_ip,omitempty" jsonschema:"the IPv4 or IPv6 gateway IP for the route"`
}

// K8sNodePoolLanInput attaches an existing private LAN to the worker nodes.
type K8sNodePoolLanInput struct {
	ID     int32                      `json:"id" jsonschema:"the numeric LAN ID of an existing LAN in the node pool's data center"`
	Dhcp   *bool                      `json:"dhcp,omitempty" jsonschema:"whether the worker nodes reserve an IP on this LAN via DHCP"`
	Routes []K8sNodePoolLanRouteInput `json:"routes,omitempty" jsonschema:"static routes to add on this LAN interface"`
}

// There is deliberately no taint input: the spec marks node pool `taints` x-internal,
// like `vnet` and `placementGroupId`, which this repo never exposes.

// CreateK8sClusterInput is the input for create_k8s_cluster. Two-phase confirmed.
type CreateK8sClusterInput struct {
	Name               string                     `json:"name" jsonschema:"the name of the new cluster: 63 characters or fewer, beginning and ending with an alphanumeric character, with dashes, underscores, dots and alphanumerics between"`
	K8sVersion         *string                    `json:"k8s_version,omitempty" jsonschema:"the Kubernetes version the control plane runs, e.g. 1.31.2. Omit to take the account default (get_k8s_default_version); list the choices with list_k8s_versions. This caps which versions the cluster's node pools may run."`
	MaintenanceWindow  *K8sMaintenanceWindowInput `json:"maintenance_window,omitempty" jsonschema:"the weekly window in which IONOS may apply control-plane maintenance. Omit to let IONOS choose one."`
	Public             *bool                      `json:"public,omitempty" jsonschema:"whether the Kubernetes API server is reachable from the internet. Defaults to true. Setting it to false creates a private cluster, which also requires location and nat_gateway_ip. PRERELEASE at IONOS, along with the three fields below — expect it to be unavailable on some contracts."`
	Location           *string                    `json:"location,omitempty" jsonschema:"the location of a private cluster, e.g. de/fra. Mandatory when public is false, optional otherwise, and immutable either way. The location must be enabled for your contract or already hold a data center of yours. Prerelease."`
	NatGatewayIp       *string                    `json:"nat_gateway_ip,omitempty" jsonschema:"the NAT gateway IP of a private cluster. Mandatory when public is false and immutable. Must be an IP you have already reserved (see list_ip_blocks) in the same location as the cluster. Prerelease."`
	NodeSubnet         *string                    `json:"node_subnet,omitempty" jsonschema:"the node subnet of a private cluster in CIDR notation with a 16-bit prefix, e.g. 10.0.0.0/16. Optional and immutable. Prerelease."`
	ApiSubnetAllowList []string                   `json:"api_subnet_allow_list,omitempty" jsonschema:"restrict access to the Kubernetes API server to these IPs or CIDRs. Traffic inside the cluster is unaffected. A bare IP is treated as /32 (IPv4) or /128 (IPv6). Omit to leave API server access unrestricted."`
	S3Buckets          []string                   `json:"s3_buckets,omitempty" jsonschema:"names of already-existing Object Storage buckets for Kubernetes use. At most one, which receives the Kubernetes API audit logs. Only the name is sent: IONOS writes the logs itself, so no Object Storage credentials are involved here."`
	ConfirmationToken  *string                    `json:"confirmation_token,omitempty" jsonschema:"leave empty on the FIRST call to receive a preview plus a one-time token; pass that token on the SECOND call (with the same name and location) to actually create the cluster. The token expires after a few minutes."`
}

// UpdateK8sClusterInput is the input for update_k8s_cluster. Single call. Omitted
// fields are carried forward; dropping api_subnet_allow_list would expose the API
// server. location, nat_gateway_ip, node_subnet and public are immutable and absent.
type UpdateK8sClusterInput struct {
	K8sClusterID       string                     `json:"k8s_cluster_id" jsonschema:"the ID of the cluster to update"`
	Name               *string                    `json:"name,omitempty" jsonschema:"a new cluster name. Omit to keep the current one."`
	K8sVersion         *string                    `json:"k8s_version,omitempty" jsonschema:"a Kubernetes version to upgrade the control plane to. Omit to keep the current one. Only the versions in the cluster's availableUpgradeVersions (see get_k8s_cluster) are accepted, and upgrading is not reversible."`
	MaintenanceWindow  *K8sMaintenanceWindowInput `json:"maintenance_window,omitempty" jsonschema:"a new maintenance window. Omit to keep the current one."`
	ApiSubnetAllowList []string                   `json:"api_subnet_allow_list,omitempty" jsonschema:"REPLACE the Kubernetes API server allow list with these IPs or CIDRs. Include every entry that should remain — any you leave out loses access. Omit the field to keep the current list; pass an empty array to remove the restriction entirely and expose the API server to any source."`
	S3Buckets          []string                   `json:"s3_buckets,omitempty" jsonschema:"REPLACE the Object Storage buckets configured for Kubernetes use, by name; they must already exist. Omit the field to keep the current ones; pass an empty array to detach them and stop audit-log delivery."`
}

// DeleteK8sClusterInput is the input for delete_k8s_cluster. Two-phase confirmed.
type DeleteK8sClusterInput struct {
	K8sClusterID      string  `json:"k8s_cluster_id" jsonschema:"the ID of the cluster to delete"`
	ConfirmationToken *string `json:"confirmation_token,omitempty" jsonschema:"leave empty on the FIRST call to receive a preview of the cluster and its node pools plus a one-time token; pass that token on the SECOND call to actually delete it. The token expires after a few minutes."`
}

// CreateK8sNodepoolInput is the input for create_k8s_nodepool. Two-phase confirmed.
// The node hardware fields are fixed at creation, so none appear on the update input.
type CreateK8sNodepoolInput struct {
	K8sClusterID      string                     `json:"k8s_cluster_id" jsonschema:"the ID of the cluster the node pool belongs to"`
	Name              string                     `json:"name" jsonschema:"the name of the new node pool: 63 characters or fewer, beginning and ending with an alphanumeric character, with dashes, underscores, dots and alphanumerics between"`
	DatacenterID      string                     `json:"datacenter_id" jsonschema:"the ID of the data center that hosts the worker nodes. It must be in the same location as the cluster, or one of that location's associated locations. For a private cluster every node pool's data center must be in the same location. Immutable."`
	NodeCount         int32                      `json:"node_count" jsonschema:"how many worker nodes to provision. With auto_scaling set, this is the starting count and must fall between min_node_count and max_node_count."`
	CoresCount        int32                      `json:"cores_count" jsonschema:"CPU cores per worker node. Immutable after creation."`
	RamSize           int32                      `json:"ram_size" jsonschema:"RAM per worker node in MB. Must be a multiple of 1024 and at least 2048. Immutable after creation."`
	AvailabilityZone  string                     `json:"availability_zone" jsonschema:"the availability zone for the worker nodes: AUTO, ZONE_1 or ZONE_2. Immutable after creation."`
	StorageType       string                     `json:"storage_type" jsonschema:"the storage type for the worker nodes: HDD or SSD. Immutable after creation."`
	StorageSize       int32                      `json:"storage_size" jsonschema:"the volume size per worker node in GB. More than 100 GB is recommended for SSD. Immutable after creation."`
	CpuFamily         *string                    `json:"cpu_family,omitempty" jsonschema:"DEPRECATED by IONOS — use server_type instead. The CPU family for the worker nodes, e.g. INTEL_ICELAKE. Omit to let IONOS pick one available at the location. An empty string is not accepted. Immutable after creation."`
	ServerType        *string                    `json:"server_type,omitempty" jsonschema:"whether the nodes get dedicated or shared CPU cores: DedicatedCore or VCPU. Defaults to DedicatedCore. Prefer this over the deprecated cpu_family."`
	K8sVersion        *string                    `json:"k8s_version,omitempty" jsonschema:"the Kubernetes version the worker nodes run, e.g. 1.31.2. Omit to take the cluster's version. Must be one of the cluster's viableNodePoolVersions (see get_k8s_cluster)."`
	MaintenanceWindow *K8sMaintenanceWindowInput `json:"maintenance_window,omitempty" jsonschema:"the weekly window in which IONOS may apply node maintenance, which replaces nodes one at a time. Omit to let IONOS choose one."`
	AutoScaling       *K8sAutoScalingInput       `json:"auto_scaling,omitempty" jsonschema:"turn on the cluster autoscaler and bound the node count. Omit for a fixed-size pool of node_count nodes."`
	Lans              []K8sNodePoolLanInput      `json:"lans,omitempty" jsonschema:"existing private LANs to attach to the worker nodes"`
	Labels            map[string]string          `json:"labels,omitempty" jsonschema:"Kubernetes labels to set on every node in the pool, as key-value pairs"`
	Annotations       map[string]string          `json:"annotations,omitempty" jsonschema:"Kubernetes annotations to set on every node in the pool, as key-value pairs"`
	PublicIps         []string                   `json:"public_ips,omitempty" jsonschema:"reserved public IPs for the worker nodes (see list_ip_blocks), all from the node pool's data center location. One more IP is needed than the maximum node count — node_count+1, or max_node_count+1 with auto_scaling — because the spare is used while a node is rebuilt."`
	ConfirmationToken *string                    `json:"confirmation_token,omitempty" jsonschema:"leave empty on the FIRST call to receive a preview plus a one-time token; pass that token on the SECOND call (with the same k8s_cluster_id, name and datacenter_id) to actually create the node pool. The token expires after a few minutes."`
}

// UpdateK8sNodepoolInput is the input for update_k8s_nodepool. Single call. Omitted
// fields are carried forward, which is what stops node_count going to 0 and draining
// the pool. name and the node hardware are immutable and absent.
type UpdateK8sNodepoolInput struct {
	K8sClusterID      string                     `json:"k8s_cluster_id" jsonschema:"the ID of the cluster the node pool belongs to"`
	NodepoolID        string                     `json:"nodepool_id" jsonschema:"the ID of the node pool to update"`
	NodeCount         *int32                     `json:"node_count,omitempty" jsonschema:"a new worker node count. Omit to keep the current one. Scaling down removes nodes and evicts whatever runs on them. Ignored while auto_scaling is active, since the autoscaler owns the count."`
	ServerType        *string                    `json:"server_type,omitempty" jsonschema:"a new server type: DedicatedCore or VCPU. Omit to keep the current one."`
	K8sVersion        *string                    `json:"k8s_version,omitempty" jsonschema:"a Kubernetes version to upgrade the worker nodes to. Omit to keep the current one. Only the versions in the pool's availableUpgradeVersions (see get_k8s_nodepool) are accepted; upgrading replaces every node and is not reversible."`
	MaintenanceWindow *K8sMaintenanceWindowInput `json:"maintenance_window,omitempty" jsonschema:"a new maintenance window. Omit to keep the current one."`
	AutoScaling       *K8sAutoScalingInput       `json:"auto_scaling,omitempty" jsonschema:"new autoscaler bounds, both at least 1. Omit to keep the current setting. An existing autoscaler CANNOT be turned off through this endpoint — the API rejects zero bounds and ignores an omitted field — so recreate the node pool without auto_scaling if you need it gone."`
	Lans              []K8sNodePoolLanInput      `json:"lans,omitempty" jsonschema:"REPLACE the attached private LANs with this list. Include every LAN the pool should keep — any you leave out is detached from the worker nodes. Omit the field to keep the current LANs; pass an empty array to detach them all."`
	Labels            map[string]string          `json:"labels,omitempty" jsonschema:"REPLACE the node labels with these key-value pairs. Omit the field to keep the current labels; pass an empty object to remove them all."`
	Annotations       map[string]string          `json:"annotations,omitempty" jsonschema:"REPLACE the node annotations with these key-value pairs. Omit the field to keep the current annotations; pass an empty object to remove them all."`
	PublicIps         []string                   `json:"public_ips,omitempty" jsonschema:"REPLACE the reserved public IPs for the worker nodes. One more IP is needed than the maximum node count. Omit the field to keep the current IPs; pass an empty array to remove them all."`
}

// DeleteK8sNodepoolInput is the input for delete_k8s_nodepool. Two-phase confirmed.
type DeleteK8sNodepoolInput struct {
	K8sClusterID      string  `json:"k8s_cluster_id" jsonschema:"the ID of the cluster the node pool belongs to"`
	NodepoolID        string  `json:"nodepool_id" jsonschema:"the ID of the node pool to delete"`
	ConfirmationToken *string `json:"confirmation_token,omitempty" jsonschema:"leave empty on the FIRST call to receive a preview of the node pool and the worker nodes it would destroy plus a one-time token; pass that token on the SECOND call to actually delete it. The token expires after a few minutes."`
}

// K8sNodeActionInput is the input for recreate_k8s_node and delete_k8s_node.
type K8sNodeActionInput struct {
	K8sClusterID      string  `json:"k8s_cluster_id" jsonschema:"the ID of the cluster the node belongs to"`
	NodepoolID        string  `json:"nodepool_id" jsonschema:"the ID of the node pool the node belongs to"`
	NodeID            string  `json:"node_id" jsonschema:"the ID of the node to act on"`
	ConfirmationToken *string `json:"confirmation_token,omitempty" jsonschema:"leave empty on the FIRST call to receive a preview of the node plus a one-time token; pass that token on the SECOND call to actually perform the operation. The token expires after a few minutes."`
}

// Dynamic load-mode meta-tool input types. Used only when the server runs in
// 'dynamic' mode, where these are the only tools the client sees and the full
// product catalog is browsed/invoked through them.

type SearchToolsInput struct {
	Query string  `json:"query" jsonschema:"keywords to match against tool names and descriptions; leave empty to browse, optionally with group"`
	Group *string `json:"group,omitempty" jsonschema:"restrict results to a single product group (e.g. compute, dns, billing, cert, activitylog, objectstorage, k8s)"`
	Limit *int    `json:"limit,omitempty" jsonschema:"maximum number of results to return; omit for the default of 10, or pass 0 for no limit. Must not be negative."`
}

type DescribeToolsInput struct {
	Names []string `json:"names" jsonschema:"the exact names of the tools to describe; returns each tool's description and full JSON input schema"`
}

type CallToolInput struct {
	Name      string         `json:"name" jsonschema:"the exact name of the tool to invoke (from ionos_search_tools)"`
	Arguments map[string]any `json:"arguments,omitempty" jsonschema:"the tool's arguments as a JSON object; see ionos_describe_tools for the schema"`
}
