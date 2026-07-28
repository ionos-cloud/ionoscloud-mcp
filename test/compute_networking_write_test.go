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

// TestUpdateIpBlockPreservesLocationAndSize is the critical test for this resource.
// IpBlockProperties.Location and .Size are non-pointer fields the SDK serializes
// unconditionally, so a PATCH built without them would send "location":"" and
// "size":0 — asking the API to relocate and resize the block as a side effect of a
// rename. update_ip_block therefore reads the current values and sends them back.
func TestUpdateIpBlockPreservesLocationAndSize(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(ipBlocksAPI+"/ipb-1", `{"id":"ipb-1","properties":{"name":"old","location":"de/txl","size":8}}`)

	res := callTool(t, h, "update_ip_block", map[string]any{
		"ipblock_id": "ipb-1", "name": "renamed",
	})
	if res.IsError {
		t.Fatalf("update failed: %s", resultText(res))
	}

	reqs := h.log.allRequests()
	if len(reqs) != 2 {
		t.Fatalf("expected a GET then a PATCH, got %d: %+v", len(reqs), reqs)
	}
	if reqs[0].Method != http.MethodGet {
		t.Errorf("first request should read the block's current location and size, got %s", reqs[0].Method)
	}
	patch := reqs[1]
	if patch.Method != http.MethodPatch {
		t.Fatalf("second request should be the PATCH, got %s", patch.Method)
	}
	for _, want := range []string{`"name":"renamed"`, `"location":"de/txl"`, `"size":8`} {
		if !strings.Contains(patch.Body, want) {
			t.Errorf("PATCH must carry %s forward:\n%s", want, patch.Body)
		}
	}
	// The corruption this guards against.
	if strings.Contains(patch.Body, `"size":0`) || strings.Contains(patch.Body, `"location":""`) {
		t.Errorf("PATCH must never send an empty location or zero size:\n%s", patch.Body)
	}
}

func TestUpdateIpBlockNotFound(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serveStatus(ipBlocksAPI+"/ipb-1", http.StatusNotFound, `{"messages":[{"message":"not found"}]}`)
	res := callTool(t, h, "update_ip_block", map[string]any{"ipblock_id": "ipb-1", "name": "x"})
	if !res.IsError || !strings.Contains(resultText(res), "does not exist") {
		t.Errorf("want a friendly does-not-exist message, got: %s", resultText(res))
	}
	for _, r := range h.log.allRequests() {
		if r.Method == http.MethodPatch {
			t.Fatal("a 404 on the carry-forward read must not issue a PATCH")
		}
	}
}

// TestDeleteIpBlockListsConsumers checks the preview names what breaks. The API
// reports ipConsumers alongside a block, which is far more useful than a count of
// addresses: it says which servers and NICs lose connectivity.
func TestDeleteIpBlockListsConsumers(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(ipBlocksAPI+"/ipb-1", `{"id":"ipb-1","properties":{
		"name":"web-ips","location":"de/fra","size":2,"ips":["1.2.3.4","1.2.3.5"],
		"ipConsumers":[
			{"ip":"1.2.3.4","serverId":"srv-a","nicId":"nic-a","serverName":"web-1"},
			{"ip":"1.2.3.5","serverId":"srv-b","nicId":"nic-b","serverName":"web-2"}
		]}}`)

	preview, res := previewThenExecute(t, h, "delete_ip_block", map[string]any{"ipblock_id": "ipb-1"})

	for _, want := range []string{"1.2.3.4", "2 addresses currently assigned", "2 servers", "WARNING"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing %q:\n%s", want, preview)
		}
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodDelete)
	if req.Path != ipBlocksAPI+"/ipb-1" {
		t.Errorf("DELETE path = %s, want %s", req.Path, ipBlocksAPI+"/ipb-1")
	}
}

func TestDeleteIpBlockUnusedHasNoWarning(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(ipBlocksAPI+"/ipb-1", `{"id":"ipb-1","properties":{"name":"spare","location":"de/fra","size":1}}`)
	res := callTool(t, h, "delete_ip_block", map[string]any{"ipblock_id": "ipb-1"})
	preview := resultText(res)
	if strings.Contains(preview, "WARNING") {
		t.Errorf("an unused block should not carry an in-use warning:\n%s", preview)
	}
	if !strings.Contains(preview, "None of these addresses") {
		t.Errorf("preview should say the addresses are unassigned:\n%s", preview)
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

// TestDeletePccNamesPeeredLans checks the preview identifies which connections
// break, not just how many.
func TestDeletePccNamesPeeredLans(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(pccsAPI+"/pcc-1", `{"id":"pcc-1","properties":{"name":"dc-link","peers":[
		{"id":"1","name":"prod-lan","datacenterName":"fra-dc"},
		{"id":"2","name":"dr-lan","datacenterName":"txl-dc"}]}}`)

	preview, res := previewThenExecute(t, h, "delete_pcc", map[string]any{"pcc_id": "pcc-1"})

	for _, want := range []string{"prod-lan", "fra-dc", "dr-lan", "2 LANs that lose their private connection"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing %q:\n%s", want, preview)
		}
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	singleRequest(t, h, http.MethodDelete)
}

// ---------- cross-cutting ----------

// TestNetworkingWriteToolsScopeGating pins which scope exposes each new tool.
func TestNetworkingWriteToolsScopeGating(t *testing.T) {
	ctx := context.Background()
	writeClass := []string{
		"create_ip_block", "update_ip_block",
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
