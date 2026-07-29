package test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

// Write-tool tests for the networking resources: IP blocks, security groups and
// their rules, NIC firewall rules, and private cross connects.

const (
	ipBlocksAPI = "/cloudapi/v6/ipblocks"
	sgAPI       = "/cloudapi/v6/datacenters/dc-1/securitygroups"
	fwRulesAPI  = nicsAPI + "/nic-1/firewallrules"
	pccsAPI     = "/cloudapi/v6/pccs"
)

// ---------- IP blocks ----------

func TestCreateIpBlockTwoPhase(t *testing.T) {
	h := destructiveSetup(t)
	preview, res := previewThenExecute(t, h, "create_ip_block", map[string]any{
		"location": "de/fra", "size": 4, "name": "web-ips",
	})

	for _, want := range []string{"de/fra", "4", "web-ips"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing %q:\n%s", want, preview)
		}
	}
	// An IP block bills from creation even when unused, which the preview must say.
	if !strings.Contains(preview, "billed from creation") {
		t.Errorf("preview should warn the block is billed while unused:\n%s", preview)
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodPost)
	if req.Path != ipBlocksAPI {
		t.Errorf("POST path = %s, want %s", req.Path, ipBlocksAPI)
	}
	for _, want := range []string{`"location":"de/fra"`, `"size":4`, `"name":"web-ips"`} {
		if !strings.Contains(req.Body, want) {
			t.Errorf("POST body missing %s:\n%s", want, req.Body)
		}
	}
}

func TestCreateIpBlockValidation(t *testing.T) {
	for _, tt := range []struct {
		name    string
		args    map[string]any
		wantMsg string
	}{
		// location is schema-required, so an omitted one is rejected by the MCP SDK
		// before the handler runs. The handler's own check catches the case the
		// schema cannot: a present but blank value.
		{"blank location", map[string]any{"location": "   ", "size": 1}, "location is required"},
		{"size must be positive", map[string]any{"location": "de/fra", "size": 0}, "size must be at least 1"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			h := destructiveSetup(t)
			res := callTool(t, h, "create_ip_block", tt.args)
			if !res.IsError {
				t.Fatal("expected rejection")
			}
			if !strings.Contains(resultText(res), tt.wantMsg) {
				t.Errorf("want %q, got: %s", tt.wantMsg, resultText(res))
			}
			if len(h.log.allRequests()) != 0 {
				t.Error("validation failure must not reach the API")
			}
		})
	}
}

// TestUpdateIpBlockIsNotRegistered replaces a test that pinned the wrong behaviour.
//
// It used to assert that update_ip_block carried location and size forward, on the
// reading that IpBlockProperties serializes them unconditionally so the PATCH must at
// least send the *correct* values rather than "" and 0. That was the wrong remedy: the
// SDK's own comment on the field says location is "disallowed in update requests", so
// the API rejects a PATCH that carries it at all, correct value or not. The tool
// therefore failed every time against the real API, and the test passed because the
// mock did not enforce the constraint.
//
// The tool is gone until the SDK models Location and Size as pointers, the way
// sdk-go/v6 does — see the note in ip_block_write.go.
func TestUpdateIpBlockIsNotRegistered(t *testing.T) {
	names := toolNames(t, context.Background(), setupWithScope(t, tools.Scope{Write: true, Destructive: true}))
	if names["update_ip_block"] {
		t.Error("update_ip_block is registered, but the smallest PATCH the SDK can build carries location and size, which the API rejects as immutable — see ip_block_write.go")
	}
	// The reserve/release pair must survive: only the update is impossible.
	for _, want := range []string{"create_ip_block", "delete_ip_block"} {
		if !names[want] {
			t.Errorf("%q should still be registered", want)
		}
	}
}

// ---------- security groups ----------

func TestCreateSecurityGroupTwoPhase(t *testing.T) {
	h := destructiveSetup(t)
	preview, res := previewThenExecute(t, h, "create_security_group", map[string]any{
		"datacenter_id": dcID, "name": "web-sg", "description": "web tier",
	})
	// An empty group permits nothing, which is the trap worth stating.
	if !strings.Contains(preview, "no rules") {
		t.Errorf("preview should say the group starts with no rules:\n%s", preview)
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodPost)
	if req.Path != sgAPI {
		t.Errorf("POST path = %s, want %s", req.Path, sgAPI)
	}
	if !strings.Contains(req.Body, `"name":"web-sg"`) {
		t.Errorf("POST body missing the name:\n%s", req.Body)
	}
}

// TestUpdateSecurityGroupPreservesName is the second carry-forward trap:
// SecurityGroupProperties.Name is serialized unconditionally, so changing only the
// description would send an empty name and wipe it.
func TestUpdateSecurityGroupPreservesName(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(sgAPI+"/sg-1", `{"id":"sg-1","properties":{"name":"web-sg","description":"old"}}`)

	res := callTool(t, h, "update_security_group", map[string]any{
		"datacenter_id": dcID, "security_group_id": "sg-1", "description": "new description",
	})
	if res.IsError {
		t.Fatalf("update failed: %s", resultText(res))
	}

	reqs := h.log.allRequests()
	if len(reqs) != 2 || reqs[0].Method != http.MethodGet {
		t.Fatalf("expected a GET (to read the current name) then a PATCH, got %+v", reqs)
	}
	patch := reqs[1]
	if !strings.Contains(patch.Body, `"name":"web-sg"`) {
		t.Errorf("PATCH must carry the existing name forward:\n%s", patch.Body)
	}
	if strings.Contains(patch.Body, `"name":""`) {
		t.Errorf("PATCH must never send an empty name — it would wipe the group's name:\n%s", patch.Body)
	}
	if !strings.Contains(patch.Body, `"description":"new description"`) {
		t.Errorf("PATCH missing the requested description:\n%s", patch.Body)
	}
}

// TestUpdateSecurityGroupExplicitNameSkipsRead is the other half: when the caller
// supplies a name there is nothing to read.
func TestUpdateSecurityGroupExplicitNameSkipsRead(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "update_security_group", map[string]any{
		"datacenter_id": dcID, "security_group_id": "sg-1", "name": "renamed",
	})
	if res.IsError {
		t.Fatalf("update failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodPatch)
	if !strings.Contains(req.Body, `"name":"renamed"`) {
		t.Errorf("PATCH should carry the requested name:\n%s", req.Body)
	}
}

func TestUpdateSecurityGroupRejectsBlankName(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "update_security_group", map[string]any{
		"datacenter_id": dcID, "security_group_id": "sg-1", "name": "   ",
	})
	if !res.IsError {
		t.Fatal("a blank name should be rejected rather than clearing the group's name")
	}
	if !strings.Contains(resultText(res), "omit it entirely") {
		t.Errorf("error should say how to keep the current name: %s", resultText(res))
	}
}

func TestDeleteSecurityGroupBlastRadius(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(sgAPI+"/sg-1", `{"id":"sg-1","properties":{"name":"web-sg"},"entities":{
		"rules":{"items":[{"id":"r1"},{"id":"r2"},{"id":"r3"}]},
		"servers":{"items":[{"id":"s1"},{"id":"s2"}]},
		"nics":{"items":[{"id":"n1"}]}}}`)

	preview, res := previewThenExecute(t, h, "delete_security_group", map[string]any{
		"datacenter_id": dcID, "security_group_id": "sg-1",
	})

	for _, want := range []string{"3 rules", "2 servers that lose these rules", "1 NICs that lose these rules"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing %q:\n%s", want, preview)
		}
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	singleRequest(t, h, http.MethodDelete)
}

// ---------- NIC firewall rules ----------

func TestCreateFirewallRuleTwoPhase(t *testing.T) {
	h := destructiveSetup(t)
	preview, res := previewThenExecute(t, h, "create_firewall_rule", map[string]any{
		"datacenter_id": dcID, "server_id": srvID, "nic_id": "nic-1",
		"protocol": "TCP", "name": "allow-https", "port_range_start": 443, "port_range_end": 443,
	})

	for _, want := range []string{"TCP", "allow-https", "443-443"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing %q:\n%s", want, preview)
		}
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodPost)
	if req.Path != fwRulesAPI {
		t.Errorf("POST path = %s, want %s", req.Path, fwRulesAPI)
	}
	for _, want := range []string{`"protocol":"TCP"`, `"portRangeStart":443`} {
		if !strings.Contains(req.Body, want) {
			t.Errorf("POST body missing %s:\n%s", want, req.Body)
		}
	}
}

// TestCreateFirewallRuleOpenPortsArePreviewed checks the security-relevant reading
// of an omitted port range: it allows every port, which the preview must state
// rather than leaving the field blank.
func TestCreateFirewallRuleOpenPortsArePreviewed(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "create_firewall_rule", map[string]any{
		"datacenter_id": dcID, "server_id": srvID, "nic_id": "nic-1", "protocol": "TCP",
	})
	if !strings.Contains(resultText(res), "all ports") {
		t.Errorf("preview should say an omitted port range allows all ports:\n%s", resultText(res))
	}
}

func TestFirewallRuleValidation(t *testing.T) {
	base := map[string]any{"datacenter_id": dcID, "server_id": srvID, "nic_id": "nic-1"}
	tests := []struct {
		name    string
		extra   map[string]any
		wantMsg string
	}{
		{"protocol required on create", map[string]any{}, "protocol is required"},
		{"icmp fields rejected for TCP", map[string]any{"protocol": "TCP", "icmp_type": 8}, "not valid with protocol TCP"},
		{"ports rejected for ICMP", map[string]any{"protocol": "ICMP", "port_range_start": 1, "port_range_end": 2}, "not valid with protocol ICMP"},
		{"half-open range", map[string]any{"protocol": "TCP", "port_range_start": 443}, "must be given together"},
		{"inverted range", map[string]any{"protocol": "TCP", "port_range_start": 500, "port_range_end": 100}, "must not be greater than"},
		{"port out of range", map[string]any{"protocol": "TCP", "port_range_start": 0, "port_range_end": 70000}, "between 1 and 65534"},
		{"icmp out of range", map[string]any{"protocol": "ICMP", "icmp_type": 300}, "between 0 and 254"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := destructiveSetup(t)
			args := map[string]any{}
			for k, v := range base {
				args[k] = v
			}
			for k, v := range tt.extra {
				args[k] = v
			}
			res := callTool(t, h, "create_firewall_rule", args)
			if !res.IsError {
				t.Fatalf("expected rejection, got: %s", resultText(res))
			}
			if !strings.Contains(resultText(res), tt.wantMsg) {
				t.Errorf("want %q, got: %s", tt.wantMsg, resultText(res))
			}
			if len(h.log.allRequests()) != 0 {
				t.Error("validation failure must not reach the API")
			}
		})
	}
}

func TestUpdateFirewallRuleRejectsEmptyUpdate(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "update_firewall_rule", map[string]any{
		"datacenter_id": dcID, "server_id": srvID, "nic_id": "nic-1", "firewallrule_id": "fw-1",
	})
	if !res.IsError {
		t.Fatal("an update with no fields should be rejected")
	}
	if len(h.log.allRequests()) != 0 {
		t.Error("an empty update must not reach the API")
	}
}

// TestUpdateFirewallRuleClearsNullableFields covers widening a rule back to "any".
// The six nullable rule fields are the only way to express "do not match on this",
// and a set-only implementation cannot reach it: an omitted field means "leave
// unchanged", so a rule that once had a source_ip could never be reopened without
// deleting and recreating it (losing its ID). Clearing must emit an explicit null.
func TestUpdateFirewallRuleClearsNullableFields(t *testing.T) {
	for _, tool := range []string{"update_firewall_rule", "update_security_group_rule"} {
		t.Run(tool, func(t *testing.T) {
			h := destructiveSetup(t)
			args := map[string]any{
				"datacenter_id": dcID,
				"clear":         []any{"source_ip", "icmp_code"},
			}
			if tool == "update_firewall_rule" {
				args["server_id"], args["nic_id"], args["firewallrule_id"] = srvID, "nic-1", "fw-1"
			} else {
				args["security_group_id"], args["rule_id"] = "sg-1", "r-1"
			}
			res := callTool(t, h, tool, args)
			if res.IsError {
				t.Fatalf("clear-only update failed: %s", resultText(res))
			}
			req := singleRequest(t, h, http.MethodPatch)
			for _, want := range []string{`"sourceIp":null`, `"icmpCode":null`} {
				if !strings.Contains(req.Body, want) {
					t.Errorf("PATCH should carry %s to reset the field:\n%s", want, req.Body)
				}
			}
			// Only the cleared fields; the rest still follow the PATCH-purity rule.
			for _, unwanted := range []string{"targetIp", "sourceMac", "ipVersion", "icmpType", "name", "protocol"} {
				if strings.Contains(req.Body, unwanted) {
					t.Errorf("PATCH carries %s, which was neither set nor cleared:\n%s", unwanted, req.Body)
				}
			}
		})
	}
}

// TestRuleClearValidation rejects the two ways a clear list can be wrong before any
// API call: a name that is not nullable, and a field both set and cleared at once.
func TestRuleClearValidation(t *testing.T) {
	tests := []struct {
		name    string
		extra   map[string]any
		wantMsg string
	}{
		{"unknown field", map[string]any{"clear": []any{"protocol"}}, "not a clearable field"},
		{"set and cleared", map[string]any{"source_ip": "10.0.0.1", "clear": []any{"source_ip"}}, "both given a value and listed in clear"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := destructiveSetup(t)
			args := map[string]any{
				"datacenter_id": dcID, "server_id": srvID, "nic_id": "nic-1", "firewallrule_id": "fw-1",
			}
			for k, v := range tt.extra {
				args[k] = v
			}
			res := callTool(t, h, "update_firewall_rule", args)
			if !res.IsError {
				t.Fatalf("expected rejection, got: %s", resultText(res))
			}
			if !strings.Contains(resultText(res), tt.wantMsg) {
				t.Errorf("want %q, got: %s", tt.wantMsg, resultText(res))
			}
			if len(h.log.allRequests()) != 0 {
				t.Error("validation failure must not reach the API")
			}
		})
	}
}

// TestRuleRejectsAllAddressesCidr guards a trap in the API: "0.0.0.0/0" is accepted
// and echoed back, then stored as the bare "0.0.0.0" once the request settles. That
// address matches no traffic, so a rule written to open a port to the world silently
// closes it. The remedy differs by operation, so the message must too.
func TestRuleRejectsAllAddressesCidr(t *testing.T) {
	tests := []struct {
		tool    string
		args    map[string]any
		wantMsg string
	}{
		{"create_firewall_rule", map[string]any{
			"datacenter_id": dcID, "server_id": srvID, "nic_id": "nic-1",
			"protocol": "TCP", "source_ip": "0.0.0.0/0",
		}, "omit source_ip entirely"},
		{"update_firewall_rule", map[string]any{
			"datacenter_id": dcID, "server_id": srvID, "nic_id": "nic-1", "firewallrule_id": "fw-1",
			"source_ip": "0.0.0.0/0",
		}, `list "source_ip" in the clear field`},
		{"update_security_group_rule", map[string]any{
			"datacenter_id": dcID, "security_group_id": "sg-1", "rule_id": "r-1",
			"target_ip": "::/0",
		}, `list "target_ip" in the clear field`},
	}
	for _, tt := range tests {
		t.Run(tt.tool, func(t *testing.T) {
			h := destructiveSetup(t)
			res := callTool(t, h, tt.tool, tt.args)
			if !res.IsError {
				t.Fatalf("expected rejection, got: %s", resultText(res))
			}
			if !strings.Contains(resultText(res), tt.wantMsg) {
				t.Errorf("want %q, got: %s", tt.wantMsg, resultText(res))
			}
			if len(h.log.allRequests()) != 0 {
				t.Error("validation failure must not reach the API")
			}
		})
	}
}

// TestDeletePccRefusesWhileLansAreConnected pins a documented API constraint that
// the preview previously contradicted: PccsDelete states a cross connect "can be
// deleted only if it is not connected to any LANs". The old preview described the
// peered LANs as merely losing their private connection, minted a token, and left
// the caller to discover the rejection. Worse, there is no detach anywhere in the
// tooling, so the error has to name the only real escape.
func TestDeletePccRefusesWhileLansAreConnected(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(pccsAPI+"/pcc-1", `{"id":"pcc-1","properties":{"name":"dc-link","peers":[
		{"id":"1","name":"prod-lan","datacenterName":"fra-dc"},
		{"id":"2","name":"dr-lan","datacenterName":"txl-dc"}]}}`)

	res := callTool(t, h, "delete_pcc", map[string]any{"pcc_id": "pcc-1"})
	if !res.IsError {
		t.Fatalf("a cross connect with peers cannot be deleted; want a refusal, got:\n%s", resultText(res))
	}
	out := resultText(res)
	// The blocking LANs are named, not just counted, since the caller has to go and
	// deal with each one.
	for _, want := range []string{"still connects 2 LAN", "prod-lan", "fra-dc", "dr-lan", "txl-dc", "no way to detach", "delete_lan"} {
		if !strings.Contains(out, want) {
			t.Errorf("refusal should explain the constraint and the escape (%q):\n%s", want, out)
		}
	}
	// The refusal must land before a token exists, or the model will keep the token
	// and retry a call that can never succeed.
	if strings.Contains(out, "confirmation_token") {
		t.Errorf("no token should be minted for an impossible delete:\n%s", out)
	}
}

// TestDeletePccProceedsWithNoPeers is the other half: the constraint must not block
// the case it does not apply to.
func TestDeletePccProceedsWithNoPeers(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(pccsAPI+"/pcc-1", `{"id":"pcc-1","properties":{"name":"unused"}}`)

	preview, res := previewThenExecute(t, h, "delete_pcc", map[string]any{"pcc_id": "pcc-1"})
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	if !strings.Contains(preview, "No LANs are connected") {
		t.Errorf("preview should say nothing else is affected:\n%s", preview)
	}
	singleRequest(t, h, http.MethodDelete)
}

// TestDeleteFirewallRuleShowsWhatItAllows checks the preview describes the rule's
// effect, so the caller can see which traffic stops being permitted.
func TestDeleteFirewallRuleShowsWhatItAllows(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(fwRulesAPI+"/fw-1", `{"id":"fw-1","properties":{
		"name":"allow-https","protocol":"TCP","type":"INGRESS",
		"sourceIp":"0.0.0.0/0","portRangeStart":443,"portRangeEnd":443}}`)

	preview, res := previewThenExecute(t, h, "delete_firewall_rule", map[string]any{
		"datacenter_id": dcID, "server_id": srvID, "nic_id": "nic-1", "firewallrule_id": "fw-1",
	})

	for _, want := range []string{"allow-https", "TCP", "INGRESS", "443-443", "0.0.0.0/0"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview should describe the rule (%q):\n%s", want, preview)
		}
	}
	// Losing the last rule on an active firewall blocks everything.
	if !strings.Contains(preview, "last rule") {
		t.Errorf("preview should warn about deleting the NIC's last rule:\n%s", preview)
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	singleRequest(t, h, http.MethodDelete)
}

// ---------- security group rules ----------

// TestCreateSecurityGroupRuleShowsReach is what distinguishes a group rule from a
// NIC rule: it applies to every member at once, so the preview counts them.
func TestCreateSecurityGroupRuleShowsReach(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(sgAPI+"/sg-1", `{"id":"sg-1","properties":{"name":"web-sg"},"entities":{
		"servers":{"items":[{"id":"s1"},{"id":"s2"},{"id":"s3"}]},
		"nics":{"items":[{"id":"n1"}]}}}`)

	preview, res := previewThenExecute(t, h, "create_security_group_rule", map[string]any{
		"datacenter_id": dcID, "security_group_id": "sg-1", "protocol": "TCP",
		"name": "allow-ssh", "port_range_start": 22, "port_range_end": 22,
	})

	for _, want := range []string{"3 servers assigned to this group", "1 NICs assigned to this group", "22-22"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing %q:\n%s", want, preview)
		}
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodPost)
	if req.Path != sgAPI+"/sg-1/rules" {
		t.Errorf("POST path = %s, want the group's rules collection", req.Path)
	}
}

// TestCreateSecurityGroupRuleEmptyGroupIsFlagged tells the caller the rule will do
// nothing yet, which is easy to miss when building a group before assigning it.
func TestCreateSecurityGroupRuleEmptyGroupIsFlagged(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(sgAPI+"/sg-1", `{"id":"sg-1","properties":{"name":"web-sg"}}`)
	res := callTool(t, h, "create_security_group_rule", map[string]any{
		"datacenter_id": dcID, "security_group_id": "sg-1", "protocol": "TCP",
	})
	if !strings.Contains(resultText(res), "no effect until you assign it") {
		t.Errorf("preview should say an unassigned group's rule has no effect yet:\n%s", resultText(res))
	}
}

// TestSecurityGroupRuleUsesTheRulesPath pins the SDK's asymmetric naming: create and
// delete go through methods named "...Firewallrules...", update through
// "...Rules...", but all of them hit /securitygroups/{id}/rules.
func TestSecurityGroupRuleUsesTheRulesPath(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "update_security_group_rule", map[string]any{
		"datacenter_id": dcID, "security_group_id": "sg-1", "rule_id": "r-1", "name": "renamed",
	})
	if res.IsError {
		t.Fatalf("update failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodPatch)
	if req.Path != sgAPI+"/sg-1/rules/r-1" {
		t.Errorf("PATCH path = %s, want %s", req.Path, sgAPI+"/sg-1/rules/r-1")
	}
}

func TestDeleteSecurityGroupRuleShowsReach(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(sgAPI+"/sg-1/rules/r-1", `{"id":"r-1","properties":{"name":"allow-ssh","protocol":"TCP","portRangeStart":22,"portRangeEnd":22}}`)
	h.resp.serve(sgAPI+"/sg-1", `{"id":"sg-1","properties":{"name":"web-sg"},"entities":{"servers":{"items":[{"id":"s1"},{"id":"s2"}]}}}`)

	preview, res := previewThenExecute(t, h, "delete_security_group_rule", map[string]any{
		"datacenter_id": dcID, "security_group_id": "sg-1", "rule_id": "r-1",
	})
	for _, want := range []string{"allow-ssh", "2 servers assigned to this group"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing %q:\n%s", want, preview)
		}
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodDelete)
	if req.Path != sgAPI+"/sg-1/rules/r-1" {
		t.Errorf("DELETE path = %s, want the rule path", req.Path)
	}
}

// ---------- private cross connects ----------

func TestCreatePccTwoPhase(t *testing.T) {
	h := destructiveSetup(t)
	preview, res := previewThenExecute(t, h, "create_pcc", map[string]any{
		"name": "dc-link", "description": "fra to txl",
	})
	if !strings.Contains(preview, "connects nothing until you attach LANs") {
		t.Errorf("preview should say LANs must be attached separately:\n%s", preview)
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodPost)
	if req.Path != pccsAPI {
		t.Errorf("POST path = %s, want %s", req.Path, pccsAPI)
	}
	if !strings.Contains(req.Body, `"name":"dc-link"`) {
		t.Errorf("POST body missing the name:\n%s", req.Body)
	}
}

func TestUpdatePcc(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "update_pcc", map[string]any{"pcc_id": "pcc-1", "description": "new"})
	if res.IsError {
		t.Fatalf("update failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodPatch)
	if req.Path != pccsAPI+"/pcc-1" {
		t.Errorf("PATCH path = %s, want %s", req.Path, pccsAPI+"/pcc-1")
	}
	var body map[string]any
	if err := json.Unmarshal([]byte(req.Body), &body); err != nil {
		t.Fatalf("PATCH body is not JSON (%v): %s", err, req.Body)
	}
	// Partial update: only the field asked for.
	if len(body) != 1 || body["description"] != "new" {
		t.Errorf("PATCH should contain only the description, got %s", req.Body)
	}
}

// ---------- cross-cutting ----------

// TestNetworkingWriteToolsScopeGating pins which scope exposes each new tool.
func TestNetworkingWriteToolsScopeGating(t *testing.T) {
	ctx := context.Background()
	writeClass := []string{
		"create_ip_block",
		"create_security_group", "update_security_group",
		"create_firewall_rule", "update_firewall_rule",
		"create_security_group_rule", "update_security_group_rule",
		"create_pcc", "update_pcc",
	}
	destructiveClass := []string{
		"delete_ip_block", "delete_security_group", "delete_firewall_rule",
		"delete_security_group_rule", "delete_pcc",
	}

	readOnly := toolNames(t, ctx, setup(t))
	for _, name := range append(append([]string{}, writeClass...), destructiveClass...) {
		if readOnly[name] {
			t.Errorf("read-only scope must not expose %q", name)
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
			t.Errorf("write scope must not expose destructive %q", name)
		}
	}
	destructive := toolNames(t, ctx, setupWithScope(t, tools.Scope{Write: true, Destructive: true}))
	for _, name := range destructiveClass {
		if !destructive[name] {
			t.Errorf("destructive scope should expose %q", name)
		}
	}
}
