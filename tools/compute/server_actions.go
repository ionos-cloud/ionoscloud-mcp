package compute

import (
	"context"
	"fmt"
	"strings"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

// Server power control. These are POSTs that are not creates, so they register
// through tools.RegisterActionTool with a domain verb: the mutation class comes
// from the verb, not the HTTP method, because stop_server is a POST that is
// destructive while create_server is a POST that is not.
//
// start_ and resume_ bring a server up and are single-call. stop_, reboot_,
// suspend_ and upgrade_ interrupt a running workload, so they are classified
// destructive and use the same two-phase preview flow as delete_*: the preview
// names the server and reports its current state, which is what lets a caller
// catch "wrong server" before the action lands.

// Not registered here, deliberately: CD-ROM attach/detach and LAN-NIC attach.
//
// Attaching an existing resource is expressed as a body carrying only its id.
// That works for volumes — Volume.Properties is a pointer and Volume.ToMap
// guards it, so the body marshals to {"id":"..."}. It does NOT work for images
// or NICs: Image.Properties and Nic.Properties are non-pointer fields whose
// ToMap serializes unconditionally, so the smallest body the SDK can produce is
//
//	attach CD-ROM: {"id":"...","properties":{"licenceType":""}}
//	attach LAN NIC: {"id":"...","properties":{"lan":0}}
//
// Both send property values the caller never asked for, and "lan":0 is exactly
// the corruption update_nic goes out of its way to avoid. The request builders
// accept only the typed struct, so there is no way to send a correct body
// without hand-rolling the HTTP call and duplicating auth.
//
// attach_lan_nic is redundant regardless: update_nic with an explicit lan moves a
// NIC onto a LAN using a body we control. CD-ROM attach has no substitute and is
// deferred pending an upstream fix in the SDK templates (attach-by-reference
// should model the body as an id-only object).

// RegisterServerActionTools registers server power control plus the volume
// attach/detach relations.
func RegisterServerActionTools(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	registerServerPowerUpActions(server, client, scope)
	registerServerDisruptiveActions(server, client, scope, confirm)
	registerServerVolumeRelations(server, client, scope, confirm)
}

// powerUpAction describes one of the non-disruptive power verbs.
type powerUpAction struct {
	tool        string
	verb        string
	description string
	call        func(ctx context.Context, client *ionos.APIClient, dcID, serverID string) error
}

func registerServerPowerUpActions(server *mcp.Server, client *ionos.APIClient, scope tools.Scope) {
	actions := []powerUpAction{
		{
			tool: "start_server",
			verb: "start_",
			description: "Start a stopped ENTERPRISE, VCPU or GPU server. Requires IONOS_MCP_TOOL_SCOPE to include write. Single call — starting a server does not interrupt anything, so there is no confirmation step. " +
				"Does NOT work on a CUBE server: the API rejects this method for CUBE, which is suspended and resumed instead — use resume_server. " +
				"Starting is asynchronous: the API accepts the request and the server boots afterwards. Starting an already-running server has no effect. For an ENTERPRISE server this re-allocates its cores and RAM and resumes charging for them.",
			call: func(ctx context.Context, c *ionos.APIClient, dcID, serverID string) error {
				_, err := c.ServersApi.DatacentersServersStartPost(ctx, dcID, serverID).Execute()
				return err
			},
		},
		{
			tool: "resume_server",
			verb: "resume_",
			description: "Resume a suspended CUBE server. Requires IONOS_MCP_TOOL_SCOPE to include write. Single call. " +
				"This is the counterpart to suspend_server and applies to CUBE servers only; use start_server for a stopped ENTERPRISE or VCPU server.",
			call: func(ctx context.Context, c *ionos.APIClient, dcID, serverID string) error {
				_, err := c.ServersApi.DatacentersServersResumePost(ctx, dcID, serverID).Execute()
				return err
			},
		},
	}

	for _, a := range actions {
		a := a
		tools.RegisterActionTool(server, scope,
			tools.Action{Verb: a.verb, Method: tools.MethodPost, Idempotent: true},
			&mcp.Tool{Name: a.tool, Description: a.description},
			func(ctx context.Context, req *mcp.CallToolRequest, input tools.ServerPowerInput) (*mcp.CallToolResult, any, error) {
				dcID := strings.TrimSpace(input.DatacenterID)
				id := strings.TrimSpace(input.ServerID)
				if dcID == "" {
					return tools.ErrorText("datacenter_id is required"), nil, nil
				}
				if id == "" {
					return tools.ErrorText("server_id is required"), nil, nil
				}
				if err := a.call(ctx, client, dcID, id); err != nil {
					return tools.ToResult(nil, err)
				}
				// These endpoints return no body, so report acceptance in text.
				return tools.TextResult(fmt.Sprintf("Requested %s for server %s. The action is asynchronous; the API has accepted the request. Check progress with get_server (see vmState).", strings.TrimSuffix(a.verb, "_"), id)), nil, nil
			})
	}
}

// disruptiveAction describes one of the power verbs that interrupts a server.
type disruptiveAction struct {
	tool string
	verb string
	// idempotent is false for the verbs where repeating changes the outcome
	// (a second reboot is a second reboot).
	idempotent  bool
	description string
	// headline and consequence drive the two-phase preview.
	headline    string
	consequence string
	// wantCube constrains which server types the API accepts for this action:
	// false rejects CUBE, true accepts only CUBE, nil accepts any. The preview
	// already fetches the server, so this is checked for free before a token is
	// minted rather than left to a late, generic API rejection.
	wantCube *bool
	call     func(ctx context.Context, client *ionos.APIClient, dcID, serverID string) error
}

// cubeOnly and notCube express a disruptiveAction's server-type constraint.
func cubeOnly() *bool { b := true; return &b }
func notCube() *bool  { b := false; return &b }

// checkServerType reports an error message when the server's type is one the
// action's endpoint refuses, naming the tool to use instead. CUBE servers are
// suspended and resumed; every other type is stopped and started.
func checkServerType(wantCube *bool, serverType, tool string) string {
	if wantCube == nil {
		return ""
	}
	isCube := strings.EqualFold(strings.TrimSpace(serverType), "CUBE")
	switch {
	case *wantCube && !isCube:
		return fmt.Sprintf("%s works only on CUBE servers, but this server is of type %q. Use stop_server to stop it and start_server to start it.",
			tool, firstNonEmpty(serverType, "unknown"))
	case !*wantCube && isCube:
		return fmt.Sprintf("%s does not work on CUBE servers — the API rejects this method for them. Use suspend_server to suspend this server and resume_server to bring it back.", tool)
	}
	return ""
}

func registerServerDisruptiveActions(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	actions := []disruptiveAction{
		{
			tool:       "stop_server",
			verb:       "stop_",
			idempotent: true,
			description: "Stop an ENTERPRISE, VCPU or GPU server, equivalent to pulling its power rather than a clean shutdown — anything not written to disk is lost. Requires IONOS_MCP_TOOL_SCOPE to include destructive. " +
				"Does NOT work on a CUBE server: the API rejects this method for CUBE, which is suspended rather than stopped — use suspend_server. " +
				"Two-phase: call first WITHOUT confirmation_token to see which server and what state it is in, plus a one-time token, then call again WITH the token to stop it. " +
				"For an ENTERPRISE server this also frees its cores and RAM and stops charging for them; the volumes are kept and still billed. Prefer a shutdown from inside the guest OS when the workload needs to flush state.",
			headline:    "About to STOP a server. This is like pulling its power: unwritten data is lost.",
			consequence: "For an ENTERPRISE server the cores and RAM are freed and no longer charged; attached volumes are kept and still billed. Restart it later with start_server.",
			wantCube:    notCube(),
			call: func(ctx context.Context, c *ionos.APIClient, dcID, serverID string) error {
				_, err := c.ServersApi.DatacentersServersStopPost(ctx, dcID, serverID).Execute()
				return err
			},
		},
		{
			tool:       "reboot_server",
			verb:       "reboot_",
			idempotent: false,
			description: "Reboot a server. Requires IONOS_MCP_TOOL_SCOPE to include destructive. " +
				"Two-phase: call first WITHOUT confirmation_token to see which server and what state it is in, plus a one-time token, then call again WITH the token to reboot it. " +
				"This is a hard reset, not a graceful restart, so unwritten data is lost and any workload on the server is interrupted.",
			headline:    "About to REBOOT a server. This is a hard reset: unwritten data is lost.",
			consequence: "Every workload on the server is interrupted while it restarts.",
			call: func(ctx context.Context, c *ionos.APIClient, dcID, serverID string) error {
				_, err := c.ServersApi.DatacentersServersRebootPost(ctx, dcID, serverID).Execute()
				return err
			},
		},
		{
			tool:       "suspend_server",
			verb:       "suspend_",
			idempotent: true,
			description: "Suspend a CUBE server. Requires IONOS_MCP_TOOL_SCOPE to include destructive. " +
				"Two-phase: call first WITHOUT confirmation_token to see which server and what state it is in, plus a one-time token, then call again WITH the token to suspend it. " +
				"CUBE servers only — the API rejects it for other types. Resume it later with resume_server.",
			headline:    "About to SUSPEND a CUBE server.",
			consequence: "The server stops running until you call resume_server. Its storage is retained.",
			wantCube:    cubeOnly(),
			call: func(ctx context.Context, c *ionos.APIClient, dcID, serverID string) error {
				_, err := c.ServersApi.DatacentersServersSuspendPost(ctx, dcID, serverID).Execute()
				return err
			},
		},
		{
			tool:       "upgrade_server",
			verb:       "upgrade_",
			idempotent: false,
			description: "Upgrade a server to the latest available hardware generation. Requires IONOS_MCP_TOOL_SCOPE to include destructive. " +
				"Two-phase: call first WITHOUT confirmation_token to see which server and what state it is in, plus a one-time token, then call again WITH the token to upgrade it. " +
				"The server is restarted as part of the upgrade, so its workload is interrupted.",
			headline:    "About to UPGRADE a server to the latest hardware generation.",
			consequence: "The server is restarted as part of the upgrade, interrupting everything running on it.",
			call: func(ctx context.Context, c *ionos.APIClient, dcID, serverID string) error {
				_, err := c.ServersApi.DatacentersServersUpgradePost(ctx, dcID, serverID).Execute()
				return err
			},
		},
	}

	for _, a := range actions {
		a := a
		tools.RegisterActionTool(server, scope,
			tools.Action{Verb: a.verb, Method: tools.MethodPost, Idempotent: a.idempotent},
			&mcp.Tool{Name: a.tool, Description: a.description},
			func(ctx context.Context, req *mcp.CallToolRequest, input tools.ServerDisruptiveActionInput) (*mcp.CallToolResult, any, error) {
				dcID := strings.TrimSpace(input.DatacenterID)
				id := strings.TrimSpace(input.ServerID)
				if dcID == "" {
					return tools.ErrorText("datacenter_id is required"), nil, nil
				}
				if id == "" {
					return tools.ErrorText("server_id is required"), nil, nil
				}
				target := tools.Target(dcID, id)

				// Phase 2: token present -> validate and execute.
				if tools.HasToken(input.ConfirmationToken) {
					if err := confirm.Consume(*input.ConfirmationToken, a.tool, target); err != nil {
						return tools.ErrorText(tools.ConfirmErrorText(a.tool, "datacenter_id and server_id", err)), nil, nil
					}
					if err := a.call(ctx, client, dcID, id); err != nil {
						return tools.ToResult(nil, err)
					}
					return tools.TextResult(fmt.Sprintf("Requested %s for server %s. The action is asynchronous; the API has accepted the request. Check progress with get_server (see vmState).", strings.TrimSuffix(a.verb, "_"), id)), nil, nil
				}

				// Phase 1: no token -> show which server and its state, mint a token.
				// The state matters: stopping an already-stopped server is a no-op,
				// while stopping the wrong running server is an outage.
				srv, _, err := client.ServersApi.DatacentersServersFindById(ctx, dcID, id).Depth(1).Execute()
				if err != nil {
					if tools.IsNotFound(err) {
						return tools.ErrorText(fmt.Sprintf("server %s does not exist in data center %s", id, dcID)), nil, nil
					}
					return tools.ToResult(nil, err)
				}
				props := srv.GetProperties()
				// Catch a type mismatch here, before minting a token: the API's
				// own rejection comes late and does not say which tool to use.
				if msg := checkServerType(a.wantCube, props.GetType(), a.tool); msg != "" {
					return tools.ErrorText(msg), nil, nil
				}
				token, mErr := confirm.Mint(a.tool, target)
				if mErr != nil {
					return nil, nil, mErr
				}
				return tools.TextResult(tools.Preview{
					Headline: a.headline + "\n" + a.consequence,
					Fields: tools.Fields(
						"datacenter_id", dcID,
						"server_id", id,
						"name", props.GetName(),
						"type", props.GetType(),
						"current vm_state", props.GetVmState(),
					),
					Tool:      a.tool,
					Replay:    tools.Fields("datacenter_id", dcID, "server_id", id),
					TokenNote: fmt.Sprintf("This token authorizes only %s on only this server", a.tool),
				}.Render(token)), nil, nil
			})
	}
}

func registerServerVolumeRelations(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	// attach_ is additive and trivially reversible, so it is a single call.
	tools.RegisterActionTool(server, scope,
		tools.Action{Verb: "attach_", Method: tools.MethodPost, Idempotent: true},
		&mcp.Tool{
			Name: "attach_server_volume",
			Description: "Attach an existing volume to a server. Requires IONOS_MCP_TOOL_SCOPE to include write. Single call. " +
				"The volume must already exist (create_volume) and be in the same data center as the server. " +
				"Attaching does NOT make the volume the boot device: the server keeps booting from whatever it booted from before, and a server with no boot setting still will not boot. To boot from this volume, follow up with update_server boot_volume_id, then reboot.",
		},
		func(ctx context.Context, req *mcp.CallToolRequest, input tools.AttachServerVolumeInput) (*mcp.CallToolResult, any, error) {
			dcID := strings.TrimSpace(input.DatacenterID)
			serverID := strings.TrimSpace(input.ServerID)
			volumeID := strings.TrimSpace(input.VolumeID)
			if dcID == "" || serverID == "" || volumeID == "" {
				return tools.ErrorText("datacenter_id, server_id and volume_id are all required"), nil, nil
			}
			// Attaching an existing resource is expressed as a body carrying only
			// its id — no properties, or the API would try to create a new volume.
			body := ionos.NewVolume()
			body.SetId(volumeID)
			attached, _, err := client.ServersApi.DatacentersServersVolumesPost(ctx, dcID, serverID).Volume(*body).Execute()
			return tools.ToResult(attached, err)
		})

	tools.RegisterActionTool(server, scope,
		tools.Action{Verb: "detach_", Method: tools.MethodDelete, Idempotent: true},
		&mcp.Tool{
			Name: "detach_server_volume",
			Description: "Detach a volume from a server. Requires IONOS_MCP_TOOL_SCOPE to include destructive. " +
				"Two-phase: call first WITHOUT confirmation_token to get a preview and a one-time token, then call again WITH the token to detach. " +
				"The volume is NOT deleted — it survives as an unattached volume and keeps incurring cost until removed with delete_volume. " +
				"Detaching the server's boot volume also CLEARS its boot setting, so the server cannot boot until you point it at another attached volume with update_server boot_volume_id. " +
				"To swap a server's disk, attach the replacement with attach_server_volume, set it with update_server boot_volume_id, and only then detach the old one — that order never leaves the server unbootable.",
		},
		func(ctx context.Context, req *mcp.CallToolRequest, input tools.DetachServerVolumeInput) (*mcp.CallToolResult, any, error) {
			dcID := strings.TrimSpace(input.DatacenterID)
			serverID := strings.TrimSpace(input.ServerID)
			volumeID := strings.TrimSpace(input.VolumeID)
			if dcID == "" || serverID == "" || volumeID == "" {
				return tools.ErrorText("datacenter_id, server_id and volume_id are all required"), nil, nil
			}
			target := tools.Target(dcID, serverID, volumeID)

			if tools.HasToken(input.ConfirmationToken) {
				if err := confirm.Consume(*input.ConfirmationToken, "detach_server_volume", target); err != nil {
					return tools.ErrorText(tools.ConfirmErrorText("detach_server_volume", "datacenter_id, server_id and volume_id", err)), nil, nil
				}
				_, err := client.ServersApi.DatacentersServersVolumesDelete(ctx, dcID, serverID, volumeID).Execute()
				if err != nil {
					return tools.ToResult(nil, err)
				}
				return tools.TextResult(fmt.Sprintf("Detached volume %s from server %s. The volume still exists as an unattached volume and continues to incur cost; delete it with delete_volume if it is no longer needed.", volumeID, serverID)), nil, nil
			}

			vol, _, err := client.VolumesApi.DatacentersVolumesFindById(ctx, dcID, volumeID).Depth(1).Execute()
			if err != nil {
				if tools.IsNotFound(err) {
					return tools.ErrorText(fmt.Sprintf("volume %s does not exist in data center %s", volumeID, dcID)), nil, nil
				}
				return tools.ToResult(nil, err)
			}
			props := vol.GetProperties()
			token, mErr := confirm.Mint("detach_server_volume", target)
			if mErr != nil {
				return nil, nil, mErr
			}
			return tools.TextResult(tools.Preview{
				Headline: "About to DETACH a volume from a server.\n" +
					"The volume is NOT deleted: it becomes an unattached volume and keeps incurring cost.\n" +
					"If this is the server's boot volume, detaching also clears the server's boot setting — recover with update_server boot_volume_id pointing at another ATTACHED volume. " +
					"When swapping disks, attach the replacement and set boot_volume_id BEFORE detaching this one.",
				Fields: tools.Fields(
					"datacenter_id", dcID,
					"server_id", serverID,
					"volume_id", volumeID,
					"volume name", props.GetName(),
					"size (GB)", fmt.Sprintf("%g", props.GetSize()),
					"currently attached to", props.GetBootServer(),
				),
				Tool:      "detach_server_volume",
				Replay:    tools.Fields("datacenter_id", dcID, "server_id", serverID, "volume_id", volumeID),
				TokenNote: "This token authorizes detaching ONLY this volume from ONLY this server",
			}.Render(token)), nil, nil
		})
}
