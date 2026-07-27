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
	DatacenterID      string   `json:"datacenter_id" jsonschema:"the ID of the data center to create the volume in"`
	Name              string   `json:"name" jsonschema:"the name of the new volume"`
	Size              float32  `json:"size" jsonschema:"the size of the volume in GB"`
	Type              string   `json:"type" jsonschema:"storage type: HDD, SSD, SSD Standard, SSD Premium, or DAS. DAS (Direct Attached Storage) works only inline with a CUBE server and ignores size."`
	Image             *string  `json:"image,omitempty" jsonschema:"ID of an image or snapshot to use as the template for this volume. Provide exactly one of image, image_alias or licence_type; without one of the first two the volume is created empty and has no operating system. Find IDs with list_images or list_snapshots."`
	ImageAlias        *string  `json:"image_alias,omitempty" jsonschema:"alias of an image to use as the template, e.g. ubuntu:latest. An alternative to image."`
	ImagePassword     *string  `json:"image_password,omitempty" jsonschema:"initial root/administrator password for the installed OS; works with public images only. Allowed characters are a-z, A-Z and 0-9, minimum 8 characters. Cannot be changed later. Prefer ssh_keys for Linux images."`
	SshKeys           []string `json:"ssh_keys,omitempty" jsonschema:"public SSH keys to authorize for login. Supported only when creating from a public Linux image. Can only be set at creation; reads always return null."`
	LicenceType       *string  `json:"licence_type,omitempty" jsonschema:"OS type for the volume: LINUX, WINDOWS, WINDOWS2016, WINDOWS2022, WINDOWS2025, UNKNOWN or OTHER. Required when neither image nor image_alias is given, since the licence type cannot then be inferred."`
	AvailabilityZone  *string  `json:"availability_zone,omitempty" jsonschema:"availability zone to provision in: AUTO (default), ZONE_1, ZONE_2 or ZONE_3. Not available for DAS."`
	Bus               *string  `json:"bus,omitempty" jsonschema:"bus type: VIRTIO (default, faster) or IDE. Use IDE only for images without VirtIO drivers."`
	UserData          *string  `json:"user_data,omitempty" jsonschema:"cloud-init configuration as a base64-encoded string. Requires a cloud-init-capable image or image_alias. Can only be set at creation."`
	BackupunitId      *string  `json:"backupunit_id,omitempty" jsonschema:"ID of a backup unit to associate. Requires image or image_alias. Can only be set at creation."`
	ExposeSerial      *bool    `json:"expose_serial,omitempty" jsonschema:"expose the disk serial ID to the server; some operating systems and licensed software require it"`
	ConfirmationToken *string  `json:"confirmation_token,omitempty" jsonschema:"leave empty on the FIRST call to receive a preview of what will be created plus a one-time token; pass that token on the SECOND call (with the same datacenter_id and name) to actually create the volume. The token expires after a few minutes."`
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
	BootOrder    *string  `json:"boot_order,omitempty" jsonschema:"whether this volume is used as a boot volume: PRIMARY, NONE or AUTO. PRIMARY makes it the boot volume, and requires EVERY other volume on the same server to be set to NONE first, so set the others before this one. AUTO is the legacy behaviour and requires all volumes on the server to be AUTO. To point a server at a different disk, prefer update_server with boot_volume_id — it takes one call and needs no coordination between volumes."`
}

// DeleteVolumeInput is the input for delete_volume. Two-phase confirmed; the data
// on the volume is destroyed and cannot be recovered without a snapshot.
type DeleteVolumeInput struct {
	DatacenterID      string  `json:"datacenter_id" jsonschema:"the ID of the data center the volume is in"`
	VolumeID          string  `json:"volume_id" jsonschema:"the ID of the volume to delete"`
	ConfirmationToken *string `json:"confirmation_token,omitempty" jsonschema:"leave empty on the FIRST call to receive a preview (including whether the volume is attached to a server) plus a one-time token; pass that token on the SECOND call to actually delete. All data on the volume is lost. The token authorizes deleting only the volume it was issued for and expires after a few minutes."`
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
	Pcc           *string `json:"pcc,omitempty" jsonschema:"attach the LAN to this private cross connect ID, or pass an empty string to detach it"`
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
