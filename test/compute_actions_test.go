package test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

// Tests for the non-CRUD action tools: server power control, volume snapshot
// actions, attach/detach relations and security-group assignment. These are the
// first tools to register through tools.RegisterActionTool, where the mutation
// class comes from the name's verb rather than the HTTP method.

const (
	srvAPI     = serversAPI + "/srv-1"
	srvVolsAPI = srvAPI + "/volumes"
)

// ---------- non-disruptive power actions ----------

func TestPowerUpActionsAreSingleCall(t *testing.T) {
	for _, tc := range []struct{ tool, path string }{
		{"start_server", srvAPI + "/start"},
		{"resume_server", srvAPI + "/resume"},
	} {
		t.Run(tc.tool, func(t *testing.T) {
			h := destructiveSetup(t)
			res := callTool(t, h, tc.tool, map[string]any{"datacenter_id": dcID, "server_id": srvID})
			if res.IsError {
				t.Fatalf("%s failed: %s", tc.tool, resultText(res))
			}
			// No confirmation step: one call, one POST.
			req := singleRequest(t, h, http.MethodPost)
			if req.Path != tc.path {
				t.Errorf("POST path = %s, want %s", req.Path, tc.path)
			}
			// These endpoints return no body, so the tool must report acceptance
			// in text rather than marshalling an empty struct.
			if txt := resultText(res); !strings.Contains(txt, "asynchronous") {
				t.Errorf("success text should say the action is asynchronous: %s", txt)
			}
		})
	}
}

func TestPowerUpActionRequiresIDs(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "start_server", map[string]any{"datacenter_id": dcID, "server_id": "  "})
	if !res.IsError {
		t.Fatal("a blank server_id should be rejected")
	}
	if len(h.log.allRequests()) != 0 {
		t.Error("validation failure must not reach the API")
	}
}

// ---------- disruptive power actions ----------

func TestDisruptiveActionsAreTwoPhase(t *testing.T) {
	// serverType matters: suspend_server accepts only CUBE and stop_server
	// rejects it, so each case needs a type its endpoint actually allows.
	for _, tc := range []struct{ tool, path, serverType, wantInPreview string }{
		{"stop_server", srvAPI + "/stop", "ENTERPRISE", "pulling its power"},
		{"reboot_server", srvAPI + "/reboot", "ENTERPRISE", "hard reset"},
		{"suspend_server", srvAPI + "/suspend", "CUBE", "SUSPEND"},
		{"upgrade_server", srvAPI + "/upgrade", "ENTERPRISE", "UPGRADE"},
	} {
		t.Run(tc.tool, func(t *testing.T) {
			h := destructiveSetup(t)
			h.resp.serve(srvAPI, `{"id":"srv-1","properties":{"name":"web-1","type":"`+tc.serverType+`","vmState":"RUNNING"}}`)

			preview, res := previewThenExecute(t, h, tc.tool, map[string]any{
				"datacenter_id": dcID, "server_id": srvID,
			})

			// The preview must name the server and report its current state —
			// that is what lets a caller catch "wrong server" before it lands.
			for _, want := range []string{"web-1", "RUNNING", tc.wantInPreview} {
				if !strings.Contains(preview, want) {
					t.Errorf("preview missing %q:\n%s", want, preview)
				}
			}
			if res.IsError {
				t.Fatalf("execute failed: %s", resultText(res))
			}
			req := singleRequest(t, h, http.MethodPost)
			if req.Path != tc.path {
				t.Errorf("POST path = %s, want %s", req.Path, tc.path)
			}
		})
	}
}

func TestDisruptiveActionBadToken(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "stop_server", map[string]any{
		"datacenter_id": dcID, "server_id": srvID, "confirmation_token": "bogus",
	})
	if !res.IsError {
		t.Fatal("a bogus confirmation_token must be rejected")
	}
	if len(h.log.allRequests()) != 0 {
		t.Error("a rejected token must not reach the API")
	}
}

// TestDisruptiveActionTokenIsPerAction proves the token is bound to the operation
// name as well as the target: a token minted to stop a server must not reboot it.
func TestDisruptiveActionTokenIsPerAction(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(srvAPI, `{"id":"srv-1","properties":{"name":"web-1","vmState":"RUNNING"}}`)

	res := callTool(t, h, "stop_server", map[string]any{"datacenter_id": dcID, "server_id": srvID})
	token := extractToken(t, resultText(res))
	h.log.clear()

	res = callTool(t, h, "reboot_server", map[string]any{
		"datacenter_id": dcID, "server_id": srvID, "confirmation_token": token,
	})
	if !res.IsError {
		t.Fatal("a token minted for stop_server must not authorize reboot_server")
	}
	for _, r := range h.log.allRequests() {
		if r.Method == http.MethodPost {
			t.Fatal("a mismatched token must not POST")
		}
	}
}

// TestDisruptiveActionServerTypeGuard covers the second CUBE trap. The API
// documents that stop_server cannot be used on a CUBE server and that
// suspend_server works only on one, but its rejection is late and generic. The
// preview already fetches the server, so the type is checked before a token is
// even minted and the message names the tool to use instead.
func TestDisruptiveActionServerTypeGuard(t *testing.T) {
	tests := []struct {
		name       string
		tool       string
		serverType string
		wantMsg    string
	}{
		{"stop rejects CUBE", "stop_server", "CUBE", "suspend_server"},
		{"suspend rejects ENTERPRISE", "suspend_server", "ENTERPRISE", "stop_server"},
		{"suspend rejects VCPU", "suspend_server", "VCPU", "only on CUBE"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := destructiveSetup(t)
			h.resp.serve(srvAPI, `{"id":"srv-1","properties":{"name":"s","type":"`+tt.serverType+`","vmState":"RUNNING"}}`)

			res := callTool(t, h, tt.tool, map[string]any{"datacenter_id": dcID, "server_id": srvID})
			if !res.IsError {
				t.Fatalf("%s on a %s server should be rejected, got: %s", tt.tool, tt.serverType, resultText(res))
			}
			if !strings.Contains(resultText(res), tt.wantMsg) {
				t.Errorf("error should point at the right tool (%q), got: %s", tt.wantMsg, resultText(res))
			}
			// Rejected before any token is minted, so there is nothing to replay.
			if strings.Contains(resultText(res), "confirmation_token") {
				t.Errorf("a type mismatch must not mint a token:\n%s", resultText(res))
			}
			for _, r := range h.log.allRequests() {
				if r.Method == http.MethodPost {
					t.Error("a type mismatch must not POST")
				}
			}
		})
	}
}

// TestDisruptiveActionAllowsMatchingType is the other half: the guard must not
// block the combinations the API does accept.
func TestDisruptiveActionAllowsMatchingType(t *testing.T) {
	for _, tt := range []struct{ tool, serverType string }{
		{"stop_server", "ENTERPRISE"},
		{"stop_server", "VCPU"},
		{"suspend_server", "CUBE"},
		// reboot and upgrade carry no documented type restriction.
		{"reboot_server", "CUBE"},
		{"upgrade_server", "ENTERPRISE"},
	} {
		t.Run(tt.tool+"/"+tt.serverType, func(t *testing.T) {
			h := destructiveSetup(t)
			h.resp.serve(srvAPI, `{"id":"srv-1","properties":{"name":"s","type":"`+tt.serverType+`","vmState":"RUNNING"}}`)
			res := callTool(t, h, tt.tool, map[string]any{"datacenter_id": dcID, "server_id": srvID})
			if res.IsError {
				t.Fatalf("%s on a %s server should be allowed: %s", tt.tool, tt.serverType, resultText(res))
			}
			// A token was minted, so the preview really did proceed.
			extractToken(t, resultText(res))
		})
	}
}

func TestDisruptiveActionNotFound(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serveStatus(srvAPI, http.StatusNotFound, `{"messages":[{"message":"not found"}]}`)

	res := callTool(t, h, "stop_server", map[string]any{"datacenter_id": dcID, "server_id": srvID})
	if !res.IsError {
		t.Fatal("stopping a non-existent server should report an error")
	}
	if !strings.Contains(resultText(res), "does not exist") {
		t.Errorf("want a friendly does-not-exist message, got: %s", resultText(res))
	}
}

// ---------- volume snapshot actions ----------

func TestCreateVolumeSnapshotTwoPhase(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(volumesAPI+"/vol-1", `{"id":"vol-1","properties":{"name":"data-1","size":50}}`)

	preview, res := previewThenExecute(t, h, "create_volume_snapshot", map[string]any{
		"datacenter_id": dcID, "volume_id": "vol-1", "name": "pre-upgrade",
	})

	// The preview names the volume being captured, not just its ID.
	for _, want := range []string{"data-1", "50", "pre-upgrade"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing %q:\n%s", want, preview)
		}
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodPost)
	if req.Path != volumesAPI+"/vol-1/create-snapshot" {
		t.Errorf("POST path = %s, want the create-snapshot action path", req.Path)
	}
	if !strings.Contains(req.Body, `"name":"pre-upgrade"`) {
		t.Errorf("POST body missing the snapshot name:\n%s", req.Body)
	}
}

func TestRestoreVolumeSnapshotTwoPhase(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(volumesAPI+"/vol-1", `{"id":"vol-1","properties":{"name":"data-1","size":50,"bootServer":"srv-9"}}`)

	preview, res := previewThenExecute(t, h, "restore_volume_snapshot", map[string]any{
		"datacenter_id": dcID, "volume_id": "vol-1", "snapshot_id": "snap-1",
	})

	for _, want := range []string{"OVERWRITING", "data-1", "snap-1", "IRREVERSIBLE"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing %q:\n%s", want, preview)
		}
	}
	// Restoring into a volume that a running server is using needs the server
	// stopped first, or the guest and the disk disagree.
	if !strings.Contains(preview, "stop_server") {
		t.Errorf("preview should tell the caller to stop the attached server first:\n%s", preview)
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodPost)
	if req.Path != volumesAPI+"/vol-1/restore-snapshot" {
		t.Errorf("POST path = %s, want the restore-snapshot action path", req.Path)
	}
	if !strings.Contains(req.Body, `"snapshotId":"snap-1"`) {
		t.Errorf("POST body missing snapshotId:\n%s", req.Body)
	}
}

// TestRestoreVolumeSnapshotTokenBoundToSnapshot is why snapshot_id is part of the
// target: a token must not be replayable to restore a DIFFERENT snapshot, which
// would overwrite the volume with contents the preview never described.
func TestRestoreVolumeSnapshotTokenBoundToSnapshot(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(volumesAPI+"/vol-1", `{"id":"vol-1","properties":{"name":"data-1"}}`)

	res := callTool(t, h, "restore_volume_snapshot", map[string]any{
		"datacenter_id": dcID, "volume_id": "vol-1", "snapshot_id": "snap-1",
	})
	token := extractToken(t, resultText(res))
	h.log.clear()

	res = callTool(t, h, "restore_volume_snapshot", map[string]any{
		"datacenter_id": dcID, "volume_id": "vol-1", "snapshot_id": "snap-OTHER",
		"confirmation_token": token,
	})
	if !res.IsError {
		t.Fatal("a token previewed for snap-1 must not restore snap-OTHER")
	}
	for _, r := range h.log.allRequests() {
		if r.Method == http.MethodPost {
			t.Fatal("a mismatched token must not POST")
		}
	}
}

// ---------- attach / detach volume ----------

// TestAttachServerVolumeSendsIdOnly pins the attach-by-reference body. It must be
// exactly {"id": ...}: any properties object would be property values the caller
// never supplied, which is why the CD-ROM and LAN-NIC attach tools are not
// shipped (their SDK models force one).
func TestAttachServerVolumeSendsIdOnly(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "attach_server_volume", map[string]any{
		"datacenter_id": dcID, "server_id": srvID, "volume_id": "vol-1",
	})
	if res.IsError {
		t.Fatalf("attach failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodPost)
	if req.Path != srvVolsAPI {
		t.Errorf("POST path = %s, want %s", req.Path, srvVolsAPI)
	}

	var body map[string]any
	if err := json.Unmarshal([]byte(req.Body), &body); err != nil {
		t.Fatalf("attach body is not JSON (%v): %s", err, req.Body)
	}
	if got := body["id"]; got != "vol-1" {
		t.Errorf("attach body id = %v, want vol-1", got)
	}
	if len(body) != 1 {
		t.Errorf("attach body must carry ONLY the id, got %s", req.Body)
	}
}

func TestDetachServerVolumeTwoPhase(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(volumesAPI+"/vol-1", `{"id":"vol-1","properties":{"name":"data-1","size":50,"bootServer":"srv-1"}}`)

	preview, res := previewThenExecute(t, h, "detach_server_volume", map[string]any{
		"datacenter_id": dcID, "server_id": srvID, "volume_id": "vol-1",
	})

	// Detaching is not deleting, and the leftover volume keeps costing money —
	// both facts have to be in the preview.
	for _, want := range []string{"NOT deleted", "cost", "data-1"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing %q:\n%s", want, preview)
		}
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodDelete)
	if req.Path != srvVolsAPI+"/vol-1" {
		t.Errorf("DELETE path = %s, want %s", req.Path, srvVolsAPI+"/vol-1")
	}
	if txt := resultText(res); !strings.Contains(txt, "still exists") {
		t.Errorf("success text should say the volume survives: %s", txt)
	}
}

// ---------- security group assignment ----------

func TestAssignSecurityGroupsIsFullSetPut(t *testing.T) {
	for _, tc := range []struct {
		tool string
		args map[string]any
		path string
	}{
		{
			"assign_server_security_groups",
			map[string]any{"datacenter_id": dcID, "server_id": srvID, "security_group_ids": []string{"sg-1", "sg-2"}},
			srvAPI + "/securitygroups",
		},
		{
			"assign_nic_security_groups",
			map[string]any{"datacenter_id": dcID, "server_id": srvID, "nic_id": "nic-1", "security_group_ids": []string{"sg-1", "sg-2"}},
			nicsAPI + "/nic-1/securitygroups",
		},
	} {
		t.Run(tc.tool, func(t *testing.T) {
			h := destructiveSetup(t)
			res := callTool(t, h, tc.tool, tc.args)
			if res.IsError {
				t.Fatalf("%s failed: %s", tc.tool, resultText(res))
			}
			req := singleRequest(t, h, http.MethodPut)
			if req.Path != tc.path {
				t.Errorf("PUT path = %s, want %s", req.Path, tc.path)
			}
			if !strings.Contains(req.Body, `"ids":["sg-1","sg-2"]`) {
				t.Errorf("PUT body should carry the full id list:\n%s", req.Body)
			}
		})
	}
}

// TestAssignSecurityGroupsEmptyListUnassignsAll checks the documented way to
// unassign everything, and that it serialises as [] rather than null.
func TestAssignSecurityGroupsEmptyListUnassignsAll(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "assign_server_security_groups", map[string]any{
		"datacenter_id": dcID, "server_id": srvID, "security_group_ids": []string{},
	})
	if res.IsError {
		t.Fatalf("an empty list should be accepted as unassign-all: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodPut)
	if !strings.Contains(req.Body, `"ids":[]`) {
		t.Errorf("empty assignment must serialise as [] not null:\n%s", req.Body)
	}
}

func TestAssignSecurityGroupsRejectsBlankID(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "assign_server_security_groups", map[string]any{
		"datacenter_id": dcID, "server_id": srvID, "security_group_ids": []string{"sg-1", "  "},
	})
	if !res.IsError {
		t.Fatal("a blank security group ID should be rejected rather than sent")
	}
	if len(h.log.allRequests()) != 0 {
		t.Error("validation failure must not reach the API")
	}
}

// ---------- scope gating ----------

// TestActionScopeGating pins which actions each scope exposes. The point is the
// asymmetry the verb table encodes: stop_server is a POST that needs
// "destructive", while attach_server_volume needs only "write", and
// detach_server_volume is a DELETE that is not a resource deletion.
func TestActionScopeGating(t *testing.T) {
	ctx := context.Background()
	writeClass := []string{"start_server", "resume_server", "attach_server_volume",
		"assign_server_security_groups", "assign_nic_security_groups", "create_volume_snapshot"}
	destructiveClass := []string{"stop_server", "reboot_server", "suspend_server",
		"upgrade_server", "restore_volume_snapshot", "detach_server_volume"}

	readOnly := toolNames(t, ctx, setup(t))
	for _, name := range append(append([]string{}, writeClass...), destructiveClass...) {
		if readOnly[name] {
			t.Errorf("read-only scope must not expose action %q", name)
		}
	}

	write := toolNames(t, ctx, setupWithScope(t, tools.Scope{Write: true}))
	for _, name := range writeClass {
		if !write[name] {
			t.Errorf("write scope should expose %q", name)
		}
	}
	for _, name := range destructiveClass {
		if write[name] {
			t.Errorf("write scope must not expose destructive action %q — it is destructive despite not being a delete_", name)
		}
	}

	destructive := toolNames(t, ctx, setupWithScope(t, tools.Scope{Write: true, Destructive: true}))
	for _, name := range destructiveClass {
		if !destructive[name] {
			t.Errorf("destructive scope should expose %q", name)
		}
	}
}

// TestDestructiveActionsAbsentFromDynamicCatalog checks that the registration
// gate keeps destructive actions out of the dynamic catalog under a write-only
// scope, so they are not merely refused but genuinely not there.
//
// Note what this does NOT prove. The refusal here is "no such tool" from the
// catalog lookup, not the class check in callHandler — under a write-only scope
// the destructive actions never reach the catalog, so that second gate is
// unreachable from this direction. The class check is the defence against a tool
// registered with bare mcp.AddTool, and it is covered where it can actually be
// exercised: TestCatalogClassifiesActionVerbs in tools/dynamic. Asserting only
// res.IsError here would pass even with ClassFromName's verb lookup removed,
// which is why the assertion below is on the specific message.
func TestDestructiveActionsAbsentFromDynamicCatalog(t *testing.T) {
	ctx := context.Background()
	h := setupDynamicWithScope(t, tools.Scope{Write: true})

	call := func(name string, args map[string]any) *mcp.CallToolResult {
		res, err := h.session.CallTool(ctx, &mcp.CallToolParams{
			Name:      "ionos_call_tool",
			Arguments: map[string]any{"name": name, "arguments": args},
		})
		if err != nil {
			t.Fatalf("ionos_call_tool(%s): %v", name, err)
		}
		return res
	}

	for _, name := range []string{"stop_server", "reboot_server", "upgrade_server", "detach_server_volume", "restore_volume_snapshot"} {
		h.log.clear()
		res := call(name, map[string]any{"datacenter_id": dcID, "server_id": srvID, "volume_id": "vol-1", "snapshot_id": "snap-1"})
		if !res.IsError {
			t.Errorf("%q must not be callable through ionos_call_tool under a write-only scope", name)
		}
		// Assert the specific reason: the tool is not in the catalog at all.
		// A bare IsError check would also be satisfied by a validation failure or
		// an API error, and would not distinguish the gate working from the gate
		// being bypassed.
		if txt := resultText(res); !strings.Contains(txt, "no such tool") {
			t.Errorf("%q should be absent from the catalog under a write-only scope, got: %s", name, txt)
		}
		for _, r := range h.log.allRequests() {
			switch r.Method {
			case http.MethodPost, http.MethodDelete:
				t.Errorf("%q reached the API (%s %s) under a write-only scope", name, r.Method, r.Path)
			}
		}
	}

	// The write-class actions must still work through the dispatcher, so the
	// gate is discriminating rather than simply blocking every verb.
	h.log.clear()
	if res := call("start_server", map[string]any{"datacenter_id": dcID, "server_id": srvID}); res.IsError {
		t.Errorf("start_server should be callable under a write scope: %s", resultText(res))
	}
	singleRequest(t, h, http.MethodPost)
}

// TestActionToolsCarryVerbDerivedAnnotations checks the annotations clients use to
// decide whether to prompt. The interesting cases are the ones where the verb and
// the HTTP method disagree.
func TestActionToolsCarryVerbDerivedAnnotations(t *testing.T) {
	ctx := context.Background()
	byName := map[string]*mcp.Tool{}
	for _, tool := range computeOnlyTools(t, ctx, tools.Scope{Write: true, Destructive: true}) {
		byName[tool.Name] = tool
	}

	tests := []struct {
		name            string
		wantDestructive bool
		wantIdempotent  bool
	}{
		// A POST that is destructive — method-derived hints would say otherwise.
		{"stop_server", true, true},
		{"reboot_server", true, false},
		{"upgrade_server", true, false},
		{"restore_volume_snapshot", true, false},
		// A DELETE that is not a resource deletion, but is still destructive.
		{"detach_server_volume", true, true},
		// Plain write-class actions.
		{"start_server", false, true},
		{"attach_server_volume", false, true},
		{"assign_server_security_groups", false, true},
	}
	for _, tt := range tests {
		tool, ok := byName[tt.name]
		if !ok {
			t.Errorf("action tool %q not registered", tt.name)
			continue
		}
		if tool.Annotations == nil {
			t.Errorf("%s has no annotations", tt.name)
			continue
		}
		if tool.Annotations.ReadOnlyHint {
			t.Errorf("%s must not advertise ReadOnlyHint", tt.name)
		}
		got := tool.Annotations.DestructiveHint
		if got == nil || *got != tt.wantDestructive {
			t.Errorf("%s DestructiveHint = %v, want %v", tt.name, got, tt.wantDestructive)
		}
		if tool.Annotations.IdempotentHint != tt.wantIdempotent {
			t.Errorf("%s IdempotentHint = %v, want %v", tt.name, tool.Annotations.IdempotentHint, tt.wantIdempotent)
		}
	}
}
