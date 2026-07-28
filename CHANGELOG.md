# Changelog

## Unreleased

### Added

- **Opt-in server power control, volume snapshot actions and volume attach/detach** — 12 non-CRUD tools registered through the new action-verb gate. The mutation class comes from the verb rather than the HTTP method, because the two disagree in both directions: `stop_server` is a `POST` that carries `destructiveHint: true` and needs `destructive` scope, while `detach_server_volume` is a `DELETE` that is not a resource deletion.
  - Non-disruptive, `write` scope, single call: `start_server`, `resume_server`, `attach_server_volume`, `assign_server_security_groups`, `assign_nic_security_groups`, `create_volume_snapshot`.
  - Disruptive, `destructive` scope, two-phase confirmed: `stop_server`, `reboot_server`, `suspend_server`, `upgrade_server`, `restore_volume_snapshot`, `detach_server_volume`.
  - Power-action previews name the server and report its current `vmState`, so a "wrong server" mistake is caught before the action lands. The token is bound to the operation as well as the target, so one minted to stop a server cannot reboot it.
  - `restore_volume_snapshot`'s token is bound to the volume **and** the snapshot, so it cannot be replayed to restore a different snapshot than the one previewed. Its preview tells you to `stop_server` first when the volume is attached.
  - `detach_server_volume` spells out that the volume survives as an unattached, still-billed volume; `assign_*_security_groups` spell out that the `PUT` replaces the entire set rather than adding to it.

- **Opt-in write operations for servers, volumes, NICs and LANs** — 12 new tools (`create_`/`update_`/`delete_` for each), gated by `IONOS_MCP_TOOL_SCOPE` exactly like the data-center tools, with two-phase confirmation on create and delete. Highlights:
  - `create_server` rejects a missing or contradictory size (`cores`+`ram` vs `template_uuid`) before the request reaches the API, and its preview states that the new server has no volume and no NIC.
  - `delete_server` takes `delete_volumes` (default `false`). The preview spells out what happens to the attached volumes — kept as unattached volumes that keep incurring cost, or destroyed irrecoverably — and the confirmation token is **bound to that choice**, so a token previewed as "keep" cannot execute as "destroy".
  - `create_volume` requires `image`/`image_alias` or `licence_type`, and reports `image_password`/`user_data` as `(set, not shown)` instead of echoing secrets into a preview that clients log.
  - `delete_volume`'s preview warns when the volume is still attached to a server and may be its boot disk.
  - `update_nic` reads the NIC's current `lan` and sends it back unchanged when you omit it. The SDK serializes that field unconditionally, so a naive partial update would send `"lan": 0` and move the NIC off its LAN as a side effect of, say, a rename.
  - `create_nic` warns that switching `firewall_active` on before any rules exist blocks all incoming traffic.
  - `delete_lan`'s preview counts the NICs that will lose their network connection; `delete_nic`'s counts the firewall rules and flow logs that go with it.

- **Opt-in update and delete for snapshots and images** — 4 tools. Neither resource has a create: a snapshot is produced from a volume by `create_volume_snapshot`, and the API exposes no way to create an image at all, so inventing either would be a tool that always fails (a test asserts neither exists). Both expose a wider capability-flag set than volumes, adding `cpu_hot_unplug`, `ram_hot_unplug`, `disc_scsi_hot_plug` and `disc_scsi_hot_unplug`. `delete_snapshot`'s preview points out that a snapshot is often the only copy of a volume's earlier state and flags one marked `sec_auth_protection`; `delete_image`'s lists the image's aliases, since those are how scripts refer to it, and says up front when an image is public and therefore undeletable rather than letting the caller discover that after a token round trip.

- **Opt-in write operations for load balancing** — 24 new tools: `create_`/`update_`/`delete_` for classic load balancers, network load balancers, application load balancers, both flavours of forwarding rule, target groups, NAT gateways and NAT gateway rules. Highlights:
  - Network and application load balancers share one implementation, since their API models are field-for-field identical; only the client calls differ. Their previews state that a load balancer carries no traffic until a forwarding rule is added.
  - `create_nlb_forwarding_rule` previews the actual destinations (`10.0.0.1:8443 weight 100`, flagging any backend in maintenance) rather than a count, and requires at least one target — a rule with none accepts connections and has nowhere to send them.
  - ALB previews flag the two configurations that are accepted but useless: HTTPS with no `server_certificates`, and a listener with no `http_rules`. HTTP rule validation is type-aware — `FORWARD` needs `target_group`, `REDIRECT` needs `location`, `STATIC` needs `status_code`.
  - `listener_lan` and `target_lan` must differ, and the check runs against the **merged** state on update, so moving only the listener onto the LAN the target side already uses is caught.
  - `delete_nat_gateway`'s preview counts the LANs that lose outbound internet access; `delete_target_group`'s warns that a group is account-level, so rules forwarding to it may sit on load balancers in several data centers.
  - NAT rule validation enforces the API's coupling between protocol and ports: a port range is only meaningful for TCP and UDP, and both ends must be given.

- **Opt-in write operations for networking** — 15 new tools: `create_`/`update_`/`delete_` for IP blocks, security groups, security group rules, NIC firewall rules and private cross connects. Gated and two-phase confirmed like the rest. Highlights:
  - `delete_ip_block`'s preview lists every resource still using the addresses — servers, NICs and Kubernetes node pools — from the API's `ipConsumers`, because releasing an in-use block breaks connectivity for exactly those resources, and a block is requested by location and size only — there is no way to ask for the same addresses back.
  - `create_ip_block` states that a block is billed from creation whether or not its addresses are assigned, and that `location` and `size` are immutable.
  - Firewall rules are exposed in both places the API supports them, sharing one implementation: on a single NIC (`create_firewall_rule`) and in a security group (`create_security_group_rule`), where the preview reports how many servers and NICs will inherit the rule and flags a group with no members as having no effect yet.
  - Rule validation rejects the combinations the API refuses without naming the offending field: ICMP fields with TCP/UDP, ports with ICMP, a half-open or inverted port range, and out-of-range values. An omitted port range is previewed as "all ports", since that is its security-relevant meaning.
  - `delete_firewall_rule`'s preview describes what the rule allows (protocol, direction, addresses, ports) and warns that deleting a NIC's last rule while its firewall is active blocks all incoming traffic.
  - `delete_security_group`'s preview counts the rules deleted with it plus the servers and NICs that lose their protection; `delete_pcc` refuses while any LAN is still connected, naming them (see Fixed).
  - `create_security_group` states that a new group has no rules and so permits nothing until they are added.

- **Volume hot-plug capability flags** on `create_volume`, `update_volume` and `create_server`'s inline `boot_volume`: `cpu_hot_plug`, `ram_hot_plug`, `nic_hot_plug`, `nic_hot_unplug`, `disc_virtio_hot_plug`, `disc_virtio_hot_unplug`. These decide whether the server a volume is attached to can change CPU, RAM, NICs or disks without a reboot — they live on the volume rather than the server, and without them a running server cannot be resized at all. Exposed on all three paths to match ionosctl (`volume create`/`update`) and the Terraform provider (the volume resource plus the inline volume block of its `server`, `cube_server` and `gpu_server` resources), which all offer the full set. Each is independently settable, so enabling one does not reset the other five.

- **Shared two-phase preview helpers** (`tools/preview.go`): `tools.Preview`, `tools.Fields`, `tools.BlastRadius`, `tools.ConfirmErrorText`, `tools.Target`, `tools.HasToken`, `tools.DeletedAsync` and the `tools.Opt*` formatters. Product-agnostic, so DNS, cert and object storage write tools can reuse one preview format rather than each growing their own.

- **Read-only annotations on every compute tool**: all compute tools now register through the scope gate (`tools.RegisterTool`) instead of `mcp.AddTool`, so the ~51 compute read tools that previously carried no MCP annotations at all now advertise `readOnlyHint: true`. Clients can distinguish a read from a mutation without parsing the tool name. Behaviour is otherwise unchanged — reads were, and remain, always available.

- **Gate support for non-CRUD action verbs**, in preparation for server power control, volume snapshot actions and attach/detach: `tools.RegisterActionTool` registers tools named with a domain verb (`start_`, `stop_`, `reboot_`, `suspend_`, `resume_`, `upgrade_`, `restore_`, `attach_`, `detach_`, `assign_`) rather than `create_`/`update_`/`delete_`. The mutation class comes from the verb, not the HTTP method, because the two disagree in both directions: `stop_server` is a POST that is destructive, `detach_server_volume` is a DELETE that is not a resource deletion. One `actionVerbs` table is the single source of truth, read by both the registration gate and the dynamic dispatcher's classifier, so a destructive action can never be misclassified as a read.

### Fixed

- **CUBE and GPU servers can now be created.** `create_server` gained a `boot_volume` object that creates the server's disk in the *same* API request. This is required for both template-sized types: the API accepts their storage only as part of a composite server-creation call, so such a server created without it was rejected and attaching a volume afterwards did not work. There was no recovery path, which made CUBE and GPU servers uncreatable through the toolset.

  The per-type rules are now validated before the request is sent, with messages naming the field to fix:

  | Server type | `boot_volume` | `boot_volume.type` | `boot_volume.size` |
  |---|---|---|---|
  | CUBE | required | required, must be `DAS` | must be omitted (fixed by the template) |
  | GPU | required | optional — omitted lets the API choose, `SSD Premium` also works | must be omitted (fixed by the template) |
  | ENTERPRISE, VCPU | optional | usually `HDD`/`SSD`/… (advisory only) | usually required (advisory only) |

  Only the CUBE and GPU rules are enforced. They are documented, and such a server created without its inline volume cannot be repaired by attaching one afterwards. The ENTERPRISE/VCPU storage-type and size rules are inferred rather than documented — the Terraform provider marks the inline volume's size `Optional+Computed`, and every `DAS` example in the IONOS docs happens to be a CUBE server without the docs ever stating that other types reject it — so those surface as a `NOTE:` in the preview instead of blocking. A wrong storage type is trivially recoverable by retrying, whereas a mistaken block would break a valid request with no way around it.

  `type` is now also required alongside `template_uuid`, since it is what distinguishes a CUBE template from a GPU one and the two have different storage rules. `boot_volume` is additionally required for a Confidential Computing server, because the API derives its cores and CPU family from the confidential image on that volume. The confirmation token is bound to the boot volume's storage type, so a preview shown "with a disk" cannot be executed as "with no disk".

  The GPU rules are not documented in the Go SDK; they were established from the team's own tooling, which is exercised against the real API: the Terraform provider marks the inline volume `Required` in `resource_gpu_server.go` and `resource_cube_server.go` but `Optional` in the ENTERPRISE/VCPU resource, neither template-sized resource exposes a volume size field, and its GPU acceptance test uses `SSD Premium` with no size. ionosctl likewise builds the volume into `entities` for GPU as well as CUBE, and sets an explicit `DAS` type only for CUBE — for GPU it sends no storage type at all. (Note that ionosctl's `server create` help text says GPU servers get "a Direct Attached Storage"; its code does not set that type, so the text is misleading.)

- **`delete_pcc` no longer offers a delete the API will refuse.** `PccsDelete` documents that "Cross connect can be deleted only if it is not connected to any LANs" — the only delete constraint documented anywhere in the compute API. The tool ignored it: the preview described the peered LANs as merely losing their private connection, minted a token, and left the caller to discover the rejection. It now refuses in phase one, names the LANs standing in the way, and mints nothing.

  The refusal has to name the escape, because there isn't an obvious one: **nothing in the tooling can detach a LAN from a cross connect.** `LanProperties.Pcc` is an optional string with no null setter, so the SDK cannot express the detach; the Terraform provider's LAN update assigns `pcc` only when the new value is non-empty, so clearing it from a config sends nothing and the LAN stays attached; ionosctl exposes `--pcc-id` for attaching only. The LANs must therefore be deleted before the cross connect can be, and the message says so along with the consequence (`delete_lan` disconnects every NIC on the LAN).

- **Previews no longer claim a create will destroy things.** `create_security_group_rule` told the caller that confirming it would destroy the NIC assigned to the group — the exact opposite of what the call does. The counts were right; the framing was invented by the renderer, which wrapped *every* blast-radius list in "Contained resources that will be destroyed". That was correct for a data center delete and false everywhere the entries survive: a LAN delete listed "NICs that will lose their network connection" under a destruction heading, as did the IP-block, cross-connect and load-balancer deletes.

  `tools.BlastRadius` now carries a `Destroys` flag, set by `tools.DestroyedRadius()` for entries that cease to exist and left false by `tools.AffectedRadius()` for entries that survive. The zero value is the safe one, so a radius built as a bare literal — what every call site did before this — reads as merely affected rather than announcing a destruction that is not happening. The summed total prints only for the destroy case; adding "1 address" to "1 server" produced a number that meant nothing.

  It is a flag rather than a caller-supplied heading because free text invited the same overreach it was meant to prevent. The first attempt gave each of the nine affected call sites its own sentence, and on review four merely restated their own labels ("…lose the addresses:" above three labels each ending "lose these addresses"), two repeated what the headline and token note already said, and two made claims that were not true — one inventing servers behind target-group backends that are bare `{ip, port, weight}` entries, one describing consequences of a cross-connect delete the API refuses outright. What each entry loses belongs in its `Add` label, and what the operation does belongs in the headline; the section heading only has to establish which of the two kinds this is.

  Worth recording *why* this matters beyond wording: the agent that met the false warning investigated, correctly concluded it was bogus, and proceeded. That is the right call on the evidence and the wrong habit to teach, because the next warning gets the same treatment. A preview is only load-bearing while it is accurate, so `TestAffectedRadiusNeverClaimsDestruction` and `TestBareRadiusFailsSafe` pin both directions.

- **A firewall rule can now be widened back to "any".** `update_firewall_rule` and `update_security_group_rule` gained a `clear` list, which sends an explicit `null` for the six nullable rule fields (`source_ip`, `target_ip`, `source_mac`, `ip_version`, `icmp_type`, `icmp_code`). They were set-only before: since an omitted field means "leave unchanged" and `null` is the only value that means "do not match on this", a rule that had ever been given a `source_ip` could not be reopened to every source. The only way out was to delete the rule and recreate it, losing its ID and briefly dropping the traffic it permitted.

  The descriptions were part of the same defect. Each nullable field said "omit to match any source", which is true when creating a rule and false when updating one — reading it as advice for an update sends you looking for a *value* that means anywhere, and `0.0.0.0/0` is the obvious guess. It is now rejected: the API accepts it, echoes it back unchanged, and then stores the bare network address `0.0.0.0` once the request settles. That address is non-routable, so a rule written to open a port to the world ends up matching no traffic at all, with a stored value that looks deliberate. Both tools name the remedy that fits the operation — omit the field on create, list it in `clear` on update. `update_firewall_rule`'s description also states that clearing `source_ip` on an `INGRESS` rule opens the port to the whole internet and should be said out loud first.

- **`update_lan` no longer claims an empty `pcc` detaches a LAN.** The field is a plain optional string with no null setter in the SDK, and neither the Terraform provider nor ionosctl offers a detach, so the documented behaviour was unverified. The description now says only what the field does.

- **`update_volume` boot_order documents the Confidential Computing rules**: on such a server the confidential volume is the only one permitted to be `PRIMARY` and must never be `NONE`, so `PRIMARY` on any other volume fails.

- **A server's boot device can now be changed.** `update_server` gained `boot_volume_id` and `update_volume` gained `boot_order`. Neither mechanism was exposed before, which left a real dead end: `attach_server_volume` does not make a volume bootable, and detaching the previous boot volume clears the server's boot setting entirely. A disk swap therefore produced a server that could not boot and that the toolset had no way to repair — the only remaining option was to delete and recreate it. `boot_volume_id` maps to the API's `properties.bootVolume` reference on a server `PATCH`, the same mechanism the Terraform provider's `UpdateBootDevice` uses; `boot_order` (`PRIMARY`/`NONE`/`AUTO`) is the per-volume alternative, though `PRIMARY` requires every other volume on the server to be `NONE`, so `boot_volume_id` is the one to reach for.

  The related descriptions now say so explicitly, since the gap was as much discoverability as capability: `attach_server_volume` states that attaching does not set the boot device, and `detach_server_volume` — in both its description and its preview — states that detaching the boot volume clears the setting, and gives the safe ordering for a swap (attach the replacement, set `boot_volume_id`, then detach the old one, which never leaves the server unbootable).

- **`stop_server` and `suspend_server` now check the server type before acting.** The API documents that `stop_server` cannot be used on a CUBE server and that `suspend_server` works only on one, but its rejection is late and generic. Both tools already fetch the server for their preview, so the type is validated there — before a token is minted — and the error names the tool to use instead (CUBE servers are suspended and resumed; every other type is stopped and started). The restriction is also stated in the `start_server` and `stop_server` descriptions, which previously did not mention it at all.

- **`update_image` no longer forces the legacy BIOS on.** The hazard turned out to be broader than documented: it is not only the SDK's `...WithDefaults()` constructors that inject values — `NewImageProperties(licenceType)` takes a required argument *and* sets `exposeSerial=false` and `requireLegacyBios=true`. Building an image PATCH from it would have applied both to every volume subsequently created from that image, as a side effect of editing a description. Caught by the update-body purity test while writing it. The guidance in CLAUDE.md now covers any generated constructor rather than just the `WithDefaults` variants, and lists the four known injectors.

- **Five more always-serialized fields caught in the load-balancing models, one of them outage-grade.** Seven of the eight models in this area send their required fields whether or not they were set, so a partial update built the obvious way would have written empty values over live configuration. The worst case: `NetworkLoadBalancerForwardingRuleProperties` always sends `targets`, so **renaming a forwarding rule would have removed every backend from the load balancer** — verified by reverting the fix, which produces `{"…","name":"https-renamed","targets":null}`. The others would have moved a load balancer off both its LANs (`listenerLan`/`targetLan`), reset a target group's algorithm and protocol, or left a NAT gateway with no address to translate to. Every affected update now reads the current resource and carries those values forward, and each is pinned by a test that fails when the carry-forward is removed. Lists (`targets`, `http_rules`, `server_certificates`, `public_ips`, `lans`) replace rather than merge, and an explicit empty list is rejected so a backend pool cannot be emptied by accident. The classic load balancer is the one resource here whose properties are all optional, so its update stays a genuine single-request partial `PATCH`.

- **Two more always-serialized fields caught before they could corrupt data.** The same hazard as `NicProperties.Lan`: `IpBlockProperties` serializes `Location` and `Size` unconditionally, and `SecurityGroupProperties` serializes `Name`. A partial update built the obvious way would have sent `"location": ""`/`"size": 0` — asking the API to relocate and resize an IP block as a side effect of a rename — and an empty security group name, wiping it. Both tools now read the current values and send them back unchanged, and `update_security_group` rejects a blank `name` rather than applying it. `TestUpdateBodiesContainOnlyRequestedFields` covers every update tool, so the next instance is caught by construction.

- **Partial updates no longer send SDK-injected defaults.** Update handlers built their request body with the SDK's generated `New*Properties[WithDefaults]()` constructors, which pre-set documented API defaults. A `PATCH` built that way sent them as though the caller had asked: `update_volume` would have forced `requireLegacyBios` on and reset `bootOrder` to `AUTO` (which can stop a server booting) while merely renaming a disk, and `update_nic` would have re-enabled DHCP on a NIC where it was deliberately off. Update bodies are now built from zero-valued literals and carry only the fields the caller supplied, asserted per resource by comparing the exact JSON key set. Only the data-center tools shipped previously, and `DatacenterPropertiesPut` injects no defaults, so no released behaviour was affected.

### Known limitations

- **Balanced-NIC attach/detach for the classic load balancer is not exposed**, for the same reason as CD-ROM and LAN-NIC attach below: the smallest body the typed `Nic` model can produce is `{"id":"…","properties":{"lan":0}}`, which asks the API to move the NIC to LAN 0. A test asserts none of these tools are registered, so they cannot reappear without the SDK fix.

- **CD-ROM attach and LAN-NIC attach are not exposed**, pending a fix in the Go SDK templates. Attaching an existing resource should send a body carrying only its id. That works for volumes, but `Image.Properties` and `Nic.Properties` are non-pointer fields whose serializer runs unconditionally, so the smallest body the SDK can produce is `{"id":"…","properties":{"licenceType":""}}` for a CD-ROM and `{"id":"…","properties":{"lan":0}}` for a NIC — property values the caller never supplied, and `lan: 0` is exactly the corruption `update_nic` goes out of its way to avoid. The request builders accept only the typed struct, so a correct body would require hand-rolling the HTTP call and duplicating auth. `attach_lan_nic` is redundant regardless: `update_nic` with an explicit `lan` moves a NIC onto a LAN using a body we control.

### Changed

- `ionos_call_tool`'s description now explains which name prefixes need `write` versus `destructive` scope, and warns that a destructive tool is not always a `delete_` — stopping, rebooting, upgrading, detaching or restoring interrupts a running workload or discards data.

- **Opt-in write operations for data centers** (`create_datacenter`, `update_datacenter`, `delete_datacenter`): the server is read-only by default and its tool list is unchanged unless `IONOS_MCP_TOOL_SCOPE` opts in (`write` enables create/update; `destructive`, which implies write, also enables delete). Tools are classified by HTTP method and gated at registration through a single `tools.RegisterTool` helper, so blocked tools never appear in `tools/list` — applied in all three load modes and re-checked in the dynamic dispatcher. `create_datacenter` and `delete_datacenter` use a two-phase confirmation (preview → single-use, target-bound, 5-minute token → execute); delete's preview shows the blast radius (contained servers, volumes, LANs, and more). `create_datacenter` makes exactly one data center per call. Write tools set MCP annotations (`readOnlyHint`/`destructiveHint`/`idempotentHint`). This is the blueprint for extending writes to other resources.

- **`filters` parameter on all compute `list_*` tools**: every compute list tool now accepts an optional `filters` object (`{"property": "value"}` pairs) that is forwarded to the API as server-side query params, so only matching resources are returned. Useful for scoping large result sets by name, location, state, image type, and more without client-side post-processing. If a filter returns an empty result, retry without it — a typo or value mismatch silently returns nothing. Filterable properties vary by resource (e.g. `name`, `location`, `vmState`, `cpuFamily`, `imageType`).

## v1.0.1 — June 2026

### Added

- **optional `depth` parameter on compute and Kubernetes tools**: every `list_*` tool now defaults to `depth=1`, returning resource names and properties in a single API call instead of only UUIDs. Every `get_*` tool accepts an optional `depth` to control nesting. Eliminates the agent loop of N follow-up `get_*` calls when resolving names.

- **`dynamic` load mode** (alias `search`): exposes only three meta-tools — `ionos_search_tools`, `ionos_describe_tools`, `ionos_call_tool` — through which the model discovers and invokes the full catalogue at runtime. The public tool list never changes (no `notifications/tools/list_changed` needed), so it fits clients with hard tool caps and no tool search of their own (Cursor ~40, Windsurf 100). Select with `--load-mode dynamic` or `IONOS_MCP_LOAD_MODE=dynamic`.

- **`--load-mode` flag**: selects the tool-registration strategy (`eager` | `lazy` | `dynamic`) from the command line. Takes precedence over `IONOS_MCP_LOAD_MODE`; the effective mode and its source are logged to stderr at startup.

- **Managed Kubernetes**: 8 new read-only tools as a dedicated product (eagerly loaded) — `list_k8s_clusters`, `get_k8s_cluster`, `list_k8s_nodepools`, `get_k8s_nodepool`, `list_k8s_nodepool_nodes`, `get_k8s_node`, `list_k8s_versions`, `get_k8s_default_version`. Covers the full cluster/node-pool/node hierarchy plus available version discovery.

## v1.0.0 — June 2026

First release. An MCP server that lets LLM clients explore an IONOS CLOUD account read-only.

### Added

- **Homebrew install**: `brew install ionos-cloud/ionos-cloud/ionoscloud-mcp` (macOS + Linux, Intel + Apple Silicon).
- **Docker image**: `ghcr.io/ionos-cloud/ionoscloud-mcp` (multi-arch: linux/amd64, linux/arm64). Pull and run with `docker run -i --rm -e IONOS_TOKEN=… ghcr.io/ionos-cloud/ionoscloud-mcp`.
- **Pre-built binaries** on every GitHub release for linux, macOS, and Windows (amd64 + arm64).
- **`--version` flag**: prints the server version. Useful for bug reports and version pinning.
- **Compute Engine**: Browse datacenters, servers, volumes, NICs, LANs, IP blocks, load balancers (ALB/NLB), NAT gateways, security groups, firewall rules, target groups, private cross-connects, snapshots, images, templates, locations, server GPUs and CD-ROMs, remote console URLs, and provisioning request status.
- **DNS**: Inspect zones, zone files, records, reverse records, secondary zones (incl. AXFR transfer status), DNSSEC keys, and account quota.
- **Billing**: View your contract and billing profile, invoices, product catalog, traffic and usage breakdowns (per datacenter, per period), and daily utilization.
- **Object Storage**: List buckets and objects across regions, read every bucket configuration surface (CORS, encryption, lifecycle, locking, policy, public access, replication, tagging, versioning), object metadata, retention and legal hold state, access keys, and regions.
- **Certificate Manager**: List certificates and auto-certificates, look up issuance providers.
- **Activity Log**: Look up accessible contracts, query the audit trail of API requests (who did what, when, on which resource). Defaults to the last 7 days and excludes noisy async-provisioning echoes — both overridable.

### Try

- **Cost audit**: "Audit my IONOS CLOUD account, find the top 5 cost-inducing resources this month, and suggest cost-efficiency tips."
- **Security sweep**: "List every bucket whose public access block is off or whose policy is public — flag anything that looks unintentional."
- **Audit trail**: "Show me every failed API request on my contract in the last 30 days, grouped by user."
- **Forgotten resources**: "Find unattached volumes, unused IP blocks, and stopped servers across all my datacenters."
- **DNS sanity check**: "List all zones on my account and flag any without DNSSEC enabled or with records pointing to IPs I no longer own."
- **Certificate expiry**: "Which certificates on my account expire in the next 60 days?"
- **Traffic spike investigation**: "My last invoice was higher than usual — show me daily traffic and utilization for the previous billing period and tell me what changed."
- **Onboarding tour**: "Walk me through what I have running on IONOS CLOUD — datacenters, servers, storage, DNS — like you're explaining it to a new teammate."

### Notes

- Auth via the `IONOS_TOKEN` environment variable. Server refuses to start without it.
- All responses are compacted before being returned to the LLM (no pretty-print, redundant audit envelopes stripped) so a single tool call uses far fewer tokens.
- Brand: tool descriptions and docs consistently use "IONOS CLOUD".
