package test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

// Write-tool tests for the managed load balancers. Network and application load
// balancers share one implementation because their API models are field-for-field
// identical, so every test runs against both flavours.

const (
	nlbAPI = "/cloudapi/v6/datacenters/dc-1/networkloadbalancers"
	albAPI = "/cloudapi/v6/datacenters/dc-1/applicationloadbalancers"
)

// managedLbFlavours is the table both flavours are exercised through.
var managedLbFlavours = []struct {
	tool string // suffix, e.g. "network_loadbalancer"
	api  string // collection path
}{
	{"network_loadbalancer", nlbAPI},
	{"application_loadbalancer", albAPI},
}

func TestCreateManagedLoadBalancerTwoPhase(t *testing.T) {
	for _, f := range managedLbFlavours {
		t.Run(f.tool, func(t *testing.T) {
			h := destructiveSetup(t)
			preview, res := previewThenExecute(t, h, "create_"+f.tool, map[string]any{
				"datacenter_id": dcID, "name": "edge-lb",
				"listener_lan": 1, "target_lan": 2, "ips": []string{"1.2.3.4"},
			})

			for _, want := range []string{"edge-lb", "1", "2", "1.2.3.4"} {
				if !strings.Contains(preview, want) {
					t.Errorf("preview missing %q:\n%s", want, preview)
				}
			}
			// A load balancer with no forwarding rule carries nothing — the preview
			// must say so, since that is the difference between "created" and "working".
			if !strings.Contains(preview, "carries no traffic until a forwarding rule is added") {
				t.Errorf("preview should say it needs a forwarding rule:\n%s", preview)
			}
			if res.IsError {
				t.Fatalf("execute failed: %s", resultText(res))
			}
			req := singleRequest(t, h, http.MethodPost)
			if req.Path != f.api {
				t.Errorf("POST path = %s, want %s", req.Path, f.api)
			}
			for _, want := range []string{`"name":"edge-lb"`, `"listenerLan":1`, `"targetLan":2`} {
				if !strings.Contains(req.Body, want) {
					t.Errorf("POST body missing %s:\n%s", want, req.Body)
				}
			}
		})
	}
}

// TestManagedLoadBalancerLanValidation covers the configuration that cannot work:
// the same LAN on both sides, or a non-positive LAN ID. The API rejects both but
// not in terms that name the field.
func TestManagedLoadBalancerLanValidation(t *testing.T) {
	tests := []struct {
		name    string
		args    map[string]any
		wantMsg string
	}{
		{"same lan on both sides", map[string]any{"listener_lan": 1, "target_lan": 1}, "must be different LANs"},
		{"zero listener lan", map[string]any{"listener_lan": 0, "target_lan": 2}, "listener_lan is required"},
		{"zero target lan", map[string]any{"listener_lan": 1, "target_lan": 0}, "target_lan is required"},
	}
	for _, f := range managedLbFlavours {
		for _, tt := range tests {
			t.Run(f.tool+"/"+tt.name, func(t *testing.T) {
				h := destructiveSetup(t)
				args := map[string]any{"datacenter_id": dcID, "name": "lb"}
				for k, v := range tt.args {
					args[k] = v
				}
				res := callTool(t, h, "create_"+f.tool, args)
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
}

// TestUpdateManagedLoadBalancerCarriesRequiredFields is the critical test for these
// resources. Both models serialize name, listenerLan and targetLan unconditionally,
// so a PATCH built without them would send an empty name and LAN 0 on both sides —
// moving the load balancer off its client network AND its backend network as a side
// effect of, say, toggling logging.
func TestUpdateManagedLoadBalancerCarriesRequiredFields(t *testing.T) {
	for _, f := range managedLbFlavours {
		t.Run(f.tool, func(t *testing.T) {
			h := destructiveSetup(t)
			h.resp.serve(f.api+"/lb-1", `{"id":"lb-1","properties":{
				"name":"edge-lb","listenerLan":7,"targetLan":9,"ips":["1.2.3.4"]}}`)

			res := callTool(t, h, "update_"+f.tool, map[string]any{
				"datacenter_id": dcID, "loadbalancer_id": "lb-1", "central_logging": true,
			})
			if res.IsError {
				t.Fatalf("update failed: %s", resultText(res))
			}

			reqs := h.log.allRequests()
			if len(reqs) != 2 || reqs[0].Method != http.MethodGet {
				t.Fatalf("expected a GET (to read the required fields) then a PATCH, got %+v", reqs)
			}
			patch := reqs[1]
			if patch.Method != http.MethodPatch {
				t.Fatalf("second request should be the PATCH, got %s", patch.Method)
			}
			for _, want := range []string{`"name":"edge-lb"`, `"listenerLan":7`, `"targetLan":9`, `"centralLogging":true`} {
				if !strings.Contains(patch.Body, want) {
					t.Errorf("PATCH must carry %s:\n%s", want, patch.Body)
				}
			}
			// The corruption this guards against.
			for _, bad := range []string{`"listenerLan":0`, `"targetLan":0`, `"name":""`} {
				if strings.Contains(patch.Body, bad) {
					t.Errorf("PATCH must never send %s — it would move the load balancer off its networks:\n%s", bad, patch.Body)
				}
			}
		})
	}
}

// TestUpdateManagedLoadBalancerRejectsSameLanAfterMerge checks the validation runs
// against the MERGED state, not just the input: moving only the listener onto the
// LAN the target side already uses is invalid even though the input names one LAN.
func TestUpdateManagedLoadBalancerRejectsSameLanAfterMerge(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(nlbAPI+"/lb-1", `{"id":"lb-1","properties":{"name":"edge","listenerLan":1,"targetLan":2}}`)

	res := callTool(t, h, "update_network_loadbalancer", map[string]any{
		"datacenter_id": dcID, "loadbalancer_id": "lb-1", "listener_lan": 2,
	})
	if !res.IsError {
		t.Fatal("moving the listener onto the existing target LAN must be rejected")
	}
	if !strings.Contains(resultText(res), "must be different LANs") {
		t.Errorf("error should explain the LANs collide: %s", resultText(res))
	}
	for _, r := range h.log.allRequests() {
		if r.Method == http.MethodPatch {
			t.Error("a merged-state validation failure must not PATCH")
		}
	}
}

func TestUpdateManagedLoadBalancerRejectsBlankName(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(nlbAPI+"/lb-1", `{"id":"lb-1","properties":{"name":"edge","listenerLan":1,"targetLan":2}}`)
	res := callTool(t, h, "update_network_loadbalancer", map[string]any{
		"datacenter_id": dcID, "loadbalancer_id": "lb-1", "name": "  ",
	})
	if !res.IsError {
		t.Fatal("a blank name should be rejected rather than clearing the name")
	}
	if !strings.Contains(resultText(res), "omit it entirely") {
		t.Errorf("error should say how to keep the current name: %s", resultText(res))
	}
}

func TestUpdateManagedLoadBalancerRejectsEmptyUpdate(t *testing.T) {
	for _, f := range managedLbFlavours {
		t.Run(f.tool, func(t *testing.T) {
			h := destructiveSetup(t)
			res := callTool(t, h, "update_"+f.tool, map[string]any{
				"datacenter_id": dcID, "loadbalancer_id": "lb-1",
			})
			if !res.IsError {
				t.Fatal("an update with no fields should be rejected")
			}
			// Rejected before the carry-forward read, so nothing is fetched either.
			if len(h.log.allRequests()) != 0 {
				t.Error("an empty update must not reach the API")
			}
		})
	}
}

func TestDeleteManagedLoadBalancerBlastRadius(t *testing.T) {
	for _, f := range managedLbFlavours {
		t.Run(f.tool, func(t *testing.T) {
			h := destructiveSetup(t)
			h.resp.serve(f.api+"/lb-1", `{"id":"lb-1","properties":{
				"name":"edge-lb","listenerLan":1,"targetLan":2,"ips":["1.2.3.4","1.2.3.5"]},
				"entities":{"forwardingrules":{"items":[{"id":"r1"},{"id":"r2"}]}}}`)

			preview, res := previewThenExecute(t, h, "delete_"+f.tool, map[string]any{
				"datacenter_id": dcID, "loadbalancer_id": "lb-1",
			})

			for _, want := range []string{"2 forwarding rules", "1.2.3.4", "goes offline"} {
				if !strings.Contains(preview, want) {
					t.Errorf("preview missing %q:\n%s", want, preview)
				}
			}
			if res.IsError {
				t.Fatalf("execute failed: %s", resultText(res))
			}
			req := singleRequest(t, h, http.MethodDelete)
			if req.Path != f.api+"/lb-1" {
				t.Errorf("DELETE path = %s, want %s", req.Path, f.api+"/lb-1")
			}
		})
	}
}

func TestDeleteManagedLoadBalancerNotFound(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serveStatus(nlbAPI+"/lb-1", http.StatusNotFound, `{"messages":[{"message":"not found"}]}`)
	res := callTool(t, h, "delete_network_loadbalancer", map[string]any{
		"datacenter_id": dcID, "loadbalancer_id": "lb-1",
	})
	if !res.IsError || !strings.Contains(resultText(res), "does not exist") {
		t.Errorf("want a friendly does-not-exist message, got: %s", resultText(res))
	}
	for _, r := range h.log.allRequests() {
		if r.Method == http.MethodDelete {
			t.Fatal("a 404 preview must not issue a DELETE")
		}
	}
}

// TestManagedLoadBalancerToolsAreDistinct guards the shared implementation: the two
// flavours must hit their own endpoints, not share one by accident.
func TestManagedLoadBalancerToolsAreDistinct(t *testing.T) {
	h := destructiveSetup(t)
	_, res := previewThenExecute(t, h, "create_network_loadbalancer", map[string]any{
		"datacenter_id": dcID, "name": "n", "listener_lan": 1, "target_lan": 2,
	})
	if res.IsError {
		t.Fatalf("NLB create failed: %s", resultText(res))
	}
	if got := singleRequest(t, h, http.MethodPost).Path; got != nlbAPI {
		t.Errorf("network flavour hit %s, want %s", got, nlbAPI)
	}

	h2 := destructiveSetup(t)
	_, res = previewThenExecute(t, h2, "create_application_loadbalancer", map[string]any{
		"datacenter_id": dcID, "name": "a", "listener_lan": 1, "target_lan": 2,
	})
	if res.IsError {
		t.Fatalf("ALB create failed: %s", resultText(res))
	}
	if got := singleRequest(t, h2, http.MethodPost).Path; got != albAPI {
		t.Errorf("application flavour hit %s, want %s", got, albAPI)
	}
}

func TestManagedLoadBalancerScopeGating(t *testing.T) {
	ctx := context.Background()
	readOnly := toolNames(t, ctx, setup(t))
	write := toolNames(t, ctx, setupWithScope(t, tools.Scope{Write: true}))
	destructive := toolNames(t, ctx, setupWithScope(t, tools.Scope{Write: true, Destructive: true}))

	for _, f := range managedLbFlavours {
		for _, verb := range []string{"create_", "update_"} {
			name := verb + f.tool
			if readOnly[name] {
				t.Errorf("read-only scope must not expose %q", name)
			}
			if !write[name] {
				t.Errorf("write scope should expose %q", name)
			}
		}
		del := "delete_" + f.tool
		if write[del] {
			t.Errorf("write scope must not expose %q", del)
		}
		if !destructive[del] {
			t.Errorf("destructive scope should expose %q", del)
		}
	}
}

// ---------- forwarding rules ----------

const (
	nlbRulesAPI = nlbAPI + "/lb-1/forwardingrules"
	albRulesAPI = albAPI + "/lb-1/forwardingrules"
)

func TestCreateNlbForwardingRuleTwoPhase(t *testing.T) {
	h := destructiveSetup(t)
	preview, res := previewThenExecute(t, h, "create_nlb_forwarding_rule", map[string]any{
		"datacenter_id": dcID, "loadbalancer_id": "lb-1", "name": "https",
		"algorithm": "ROUND_ROBIN", "protocol": "TCP",
		"listener_ip": "1.2.3.4", "listener_port": 443,
		"targets": []map[string]any{
			{"ip": "10.0.0.1", "port": 8443, "weight": 100},
			{"ip": "10.0.0.2", "port": 8443, "weight": 50, "maintenance": true},
		},
	})

	// The preview must show the actual destinations, not just a count.
	for _, want := range []string{"1.2.3.4:443", "10.0.0.1:8443 weight 100", "10.0.0.2:8443 weight 50", "in maintenance"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing %q:\n%s", want, preview)
		}
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodPost)
	if req.Path != nlbRulesAPI {
		t.Errorf("POST path = %s, want %s", req.Path, nlbRulesAPI)
	}
	for _, want := range []string{`"listenerPort":443`, `"ip":"10.0.0.1"`, `"weight":100`, `"maintenance":true`} {
		if !strings.Contains(req.Body, want) {
			t.Errorf("POST body missing %s:\n%s", want, req.Body)
		}
	}
}

// TestUpdateNlbForwardingRulePreservesTargets is the most important test in this
// batch. NetworkLoadBalancerForwardingRuleProperties serializes `targets`
// unconditionally, so a partial update built without it would send an empty targets
// list and remove EVERY backend from the load balancer — an outage caused by
// renaming a rule.
func TestUpdateNlbForwardingRulePreservesTargets(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(nlbRulesAPI+"/r-1", `{"id":"r-1","properties":{
		"name":"https","algorithm":"ROUND_ROBIN","protocol":"TCP",
		"listenerIp":"1.2.3.4","listenerPort":443,
		"targets":[{"ip":"10.0.0.1","port":8443,"weight":100},{"ip":"10.0.0.2","port":8443,"weight":100}]}}`)

	res := callTool(t, h, "update_nlb_forwarding_rule", map[string]any{
		"datacenter_id": dcID, "loadbalancer_id": "lb-1", "rule_id": "r-1", "name": "https-renamed",
	})
	if res.IsError {
		t.Fatalf("update failed: %s", resultText(res))
	}

	reqs := h.log.allRequests()
	if len(reqs) != 2 || reqs[0].Method != http.MethodGet {
		t.Fatalf("expected a GET (to read the rule) then a PATCH, got %+v", reqs)
	}
	patch := reqs[1]

	var body struct {
		Name         string `json:"name"`
		Algorithm    string `json:"algorithm"`
		Protocol     string `json:"protocol"`
		ListenerIp   string `json:"listenerIp"`
		ListenerPort int32  `json:"listenerPort"`
		Targets      []struct {
			Ip     string `json:"ip"`
			Port   int32  `json:"port"`
			Weight int32  `json:"weight"`
		} `json:"targets"`
	}
	if err := json.Unmarshal([]byte(patch.Body), &body); err != nil {
		t.Fatalf("PATCH body is not JSON (%v): %s", err, patch.Body)
	}
	if len(body.Targets) != 2 {
		t.Fatalf("PATCH must carry BOTH existing backends forward, got %d: %s", len(body.Targets), patch.Body)
	}
	if body.Targets[0].Ip != "10.0.0.1" || body.Targets[1].Ip != "10.0.0.2" {
		t.Errorf("carried-forward backends are wrong: %s", patch.Body)
	}
	// Every other required field must survive the rename too.
	if body.Name != "https-renamed" || body.Algorithm != "ROUND_ROBIN" ||
		body.Protocol != "TCP" || body.ListenerIp != "1.2.3.4" || body.ListenerPort != 443 {
		t.Errorf("required fields not carried forward correctly: %+v (%s)", body, patch.Body)
	}
}

// TestUpdateNlbForwardingRuleReplacesTargetsWhenGiven is the other half: supplying
// targets deliberately replaces the set.
func TestUpdateNlbForwardingRuleReplacesTargetsWhenGiven(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(nlbRulesAPI+"/r-1", `{"id":"r-1","properties":{
		"name":"https","algorithm":"ROUND_ROBIN","protocol":"TCP","listenerIp":"1.2.3.4","listenerPort":443,
		"targets":[{"ip":"10.0.0.1","port":8443,"weight":100}]}}`)

	res := callTool(t, h, "update_nlb_forwarding_rule", map[string]any{
		"datacenter_id": dcID, "loadbalancer_id": "lb-1", "rule_id": "r-1",
		"targets": []map[string]any{{"ip": "10.0.9.9", "port": 9000, "weight": 10}},
	})
	if res.IsError {
		t.Fatalf("update failed: %s", resultText(res))
	}
	reqs := h.log.allRequests()
	patch := reqs[len(reqs)-1]
	if !strings.Contains(patch.Body, `"ip":"10.0.9.9"`) {
		t.Errorf("PATCH should carry the replacement backend:\n%s", patch.Body)
	}
	if strings.Contains(patch.Body, "10.0.0.1") {
		t.Errorf("an explicit targets list must REPLACE the old ones:\n%s", patch.Body)
	}
}

// TestUpdateNlbForwardingRuleRejectsEmptyTargetList stops the caller from emptying
// the backend pool by accident: an explicit empty list is refused, and the way to
// leave backends alone is to omit the field.
func TestUpdateNlbForwardingRuleRejectsEmptyTargetList(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "update_nlb_forwarding_rule", map[string]any{
		"datacenter_id": dcID, "loadbalancer_id": "lb-1", "rule_id": "r-1",
		"targets": []map[string]any{},
	})
	if !res.IsError {
		t.Fatal("an explicit empty targets list should be rejected")
	}
	if !strings.Contains(resultText(res), "omit the field entirely") {
		t.Errorf("error should say how to leave backends untouched: %s", resultText(res))
	}
	if len(h.log.allRequests()) != 0 {
		t.Error("validation failure must not reach the API")
	}
}

func TestNlbForwardingRuleValidation(t *testing.T) {
	base := map[string]any{
		"datacenter_id": dcID, "loadbalancer_id": "lb-1", "name": "r",
		"algorithm": "ROUND_ROBIN", "protocol": "TCP", "listener_ip": "1.2.3.4", "listener_port": 443,
		"targets": []map[string]any{{"ip": "10.0.0.1", "port": 80, "weight": 1}},
	}
	tests := []struct {
		name    string
		mutate  func(map[string]any)
		wantMsg string
	}{
		{"no targets", func(a map[string]any) { a["targets"] = []map[string]any{} }, "at least one backend"},
		{"blank listener ip", func(a map[string]any) { a["listener_ip"] = " " }, "listener_ip is required"},
		{"bad listener port", func(a map[string]any) { a["listener_port"] = 0 }, "listener_port must be between"},
		{"bad target port", func(a map[string]any) {
			a["targets"] = []map[string]any{{"ip": "10.0.0.1", "port": 99999, "weight": 1}}
		}, "targets[0].port must be between"},
		{"bad target weight", func(a map[string]any) {
			a["targets"] = []map[string]any{{"ip": "10.0.0.1", "port": 80, "weight": 999}}
		}, "targets[0].weight must be between"},
		{"blank target ip", func(a map[string]any) {
			a["targets"] = []map[string]any{{"ip": "", "port": 80, "weight": 1}}
		}, "targets[0].ip is required"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := destructiveSetup(t)
			args := map[string]any{}
			for k, v := range base {
				args[k] = v
			}
			tt.mutate(args)
			res := callTool(t, h, "create_nlb_forwarding_rule", args)
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

func TestDeleteNlbForwardingRuleShowsBackends(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(nlbRulesAPI+"/r-1", `{"id":"r-1","properties":{
		"name":"https","protocol":"TCP","algorithm":"ROUND_ROBIN","listenerIp":"1.2.3.4","listenerPort":443,
		"targets":[{"ip":"10.0.0.1","port":8443,"weight":100},{"ip":"10.0.0.2","port":8443,"weight":100}]}}`)

	preview, res := previewThenExecute(t, h, "delete_nlb_forwarding_rule", map[string]any{
		"datacenter_id": dcID, "loadbalancer_id": "lb-1", "rule_id": "r-1",
	})
	for _, want := range []string{"1.2.3.4:443", "2 backends that stop receiving traffic", "only rule"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing %q:\n%s", want, preview)
		}
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	singleRequest(t, h, http.MethodDelete)
}

// ---------- ALB forwarding rules ----------

func TestCreateAlbForwardingRuleTwoPhase(t *testing.T) {
	h := destructiveSetup(t)
	preview, res := previewThenExecute(t, h, "create_alb_forwarding_rule", map[string]any{
		"datacenter_id": dcID, "loadbalancer_id": "lb-1", "name": "web",
		"protocol": "HTTP", "listener_ip": "1.2.3.4", "listener_port": 80,
		"http_rules": []map[string]any{
			{"name": "to-app", "type": "FORWARD", "target_group": "tg-1",
				"conditions": []map[string]any{{"type": "PATH", "condition": "STARTS_WITH", "value": "/api"}}},
		},
	})

	for _, want := range []string{"1.2.3.4:80", "to-app", "FORWARD", "target group tg-1", "1 condition"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing %q:\n%s", want, preview)
		}
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodPost)
	if req.Path != albRulesAPI {
		t.Errorf("POST path = %s, want %s", req.Path, albRulesAPI)
	}
	for _, want := range []string{`"targetGroup":"tg-1"`, `"condition":"STARTS_WITH"`, `"value":"/api"`} {
		if !strings.Contains(req.Body, want) {
			t.Errorf("POST body missing %s:\n%s", want, req.Body)
		}
	}
}

// TestCreateAlbForwardingRuleWarnsOnMissingPieces covers the two configurations that
// are accepted but useless: HTTPS with no certificate, and a listener with no rules.
func TestCreateAlbForwardingRuleWarnsOnMissingPieces(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "create_alb_forwarding_rule", map[string]any{
		"datacenter_id": dcID, "loadbalancer_id": "lb-1", "name": "web",
		"protocol": "HTTPS", "listener_ip": "1.2.3.4", "listener_port": 443,
	})
	preview := resultText(res)
	if !strings.Contains(preview, "no server_certificates") {
		t.Errorf("preview should warn about HTTPS without certificates:\n%s", preview)
	}
	if !strings.Contains(preview, "route nothing") {
		t.Errorf("preview should warn about a listener with no http_rules:\n%s", preview)
	}
}

func TestAlbHttpRuleValidation(t *testing.T) {
	base := map[string]any{
		"datacenter_id": dcID, "loadbalancer_id": "lb-1", "name": "web",
		"protocol": "HTTP", "listener_ip": "1.2.3.4", "listener_port": 80,
	}
	tests := []struct {
		rule    map[string]any
		wantMsg string
	}{
		{map[string]any{"name": "r", "type": "FORWARD"}, "target_group is required"},
		{map[string]any{"name": "r", "type": "REDIRECT"}, "location is required"},
		{map[string]any{"name": "r", "type": "STATIC"}, "status_code is required"},
		{map[string]any{"name": "r", "type": "WAT"}, "must be FORWARD, REDIRECT or STATIC"},
		{map[string]any{"name": "", "type": "FORWARD", "target_group": "tg"}, "name is required"},
	}
	for _, tt := range tests {
		t.Run(tt.wantMsg, func(t *testing.T) {
			h := destructiveSetup(t)
			args := map[string]any{}
			for k, v := range base {
				args[k] = v
			}
			args["http_rules"] = []map[string]any{tt.rule}
			res := callTool(t, h, "create_alb_forwarding_rule", args)
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

// TestUpdateAlbForwardingRuleCarriesRequiredFields checks the same carry-forward
// discipline, plus that omitted optional lists are preserved rather than cleared.
func TestUpdateAlbForwardingRuleCarriesRequiredFields(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(albRulesAPI+"/r-1", `{"id":"r-1","properties":{
		"name":"web","protocol":"HTTPS","listenerIp":"1.2.3.4","listenerPort":443,
		"serverCertificates":["cert-1"],
		"httpRules":[{"name":"to-app","type":"FORWARD","targetGroup":"tg-1"}]}}`)

	res := callTool(t, h, "update_alb_forwarding_rule", map[string]any{
		"datacenter_id": dcID, "loadbalancer_id": "lb-1", "rule_id": "r-1", "client_timeout": 5000,
	})
	if res.IsError {
		t.Fatalf("update failed: %s", resultText(res))
	}
	reqs := h.log.allRequests()
	if len(reqs) != 2 || reqs[0].Method != http.MethodGet {
		t.Fatalf("expected a GET then a PATCH, got %+v", reqs)
	}
	patch := reqs[1]
	for _, want := range []string{`"name":"web"`, `"protocol":"HTTPS"`, `"listenerPort":443`, `"clientTimeout":5000`} {
		if !strings.Contains(patch.Body, want) {
			t.Errorf("PATCH must carry %s:\n%s", want, patch.Body)
		}
	}
	// Omitted lists are preserved, not cleared — losing them would stop the
	// listener routing or serving TLS.
	for _, want := range []string{`"cert-1"`, `"targetGroup":"tg-1"`} {
		if !strings.Contains(patch.Body, want) {
			t.Errorf("PATCH must preserve the omitted list containing %s:\n%s", want, patch.Body)
		}
	}
}

func TestDeleteAlbForwardingRuleShowsRules(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(albRulesAPI+"/r-1", `{"id":"r-1","properties":{
		"name":"web","protocol":"HTTP","listenerIp":"1.2.3.4","listenerPort":80,
		"httpRules":[{"name":"a","type":"FORWARD"},{"name":"b","type":"STATIC"}]}}`)

	preview, res := previewThenExecute(t, h, "delete_alb_forwarding_rule", map[string]any{
		"datacenter_id": dcID, "loadbalancer_id": "lb-1", "rule_id": "r-1",
	})
	for _, want := range []string{"1.2.3.4:80", "2 HTTP rules deleted with it"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing %q:\n%s", want, preview)
		}
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	singleRequest(t, h, http.MethodDelete)
}

// ---------- target groups, NAT gateways, classic load balancer ----------

const (
	tgAPI     = "/cloudapi/v6/targetgroups"
	natAPI    = "/cloudapi/v6/datacenters/dc-1/natgateways"
	natRules  = natAPI + "/gw-1/rules"
	classicLB = "/cloudapi/v6/datacenters/dc-1/loadbalancers"
)

func TestCreateTargetGroupTwoPhase(t *testing.T) {
	h := destructiveSetup(t)
	preview, res := previewThenExecute(t, h, "create_target_group", map[string]any{
		"name": "app-pool", "algorithm": "ROUND_ROBIN", "protocol": "HTTP",
		"targets": []map[string]any{{"ip": "10.0.0.1", "port": 8080, "weight": 100}},
	})
	for _, want := range []string{"app-pool", "ROUND_ROBIN", "10.0.0.1:8080 weight 100"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing %q:\n%s", want, preview)
		}
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodPost)
	if req.Path != tgAPI {
		t.Errorf("POST path = %s, want %s", req.Path, tgAPI)
	}
}

func TestCreateTargetGroupWarnsWithNoTargets(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "create_target_group", map[string]any{
		"name": "empty", "algorithm": "ROUND_ROBIN", "protocol": "HTTP",
	})
	if !strings.Contains(resultText(res), "accepts no traffic") {
		t.Errorf("preview should warn an empty group accepts no traffic:\n%s", resultText(res))
	}
}

// TestUpdateTargetGroupCarriesRequiredFields covers the third carry-forward case:
// name, algorithm and protocol are serialized unconditionally, and the backend list
// must survive an unrelated change or every load balancer rule forwarding here
// loses its pool.
func TestUpdateTargetGroupCarriesRequiredFields(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(tgAPI+"/tg-1", `{"id":"tg-1","properties":{
		"name":"app-pool","algorithm":"ROUND_ROBIN","protocol":"HTTP",
		"targets":[{"ip":"10.0.0.1","port":8080,"weight":100}]}}`)

	res := callTool(t, h, "update_target_group", map[string]any{
		"target_group_id": "tg-1", "name": "renamed",
	})
	if res.IsError {
		t.Fatalf("update failed: %s", resultText(res))
	}
	reqs := h.log.allRequests()
	if len(reqs) != 2 || reqs[0].Method != http.MethodGet {
		t.Fatalf("expected a GET then a PATCH, got %+v", reqs)
	}
	patch := reqs[1]
	for _, want := range []string{`"name":"renamed"`, `"algorithm":"ROUND_ROBIN"`, `"protocol":"HTTP"`, `"ip":"10.0.0.1"`} {
		if !strings.Contains(patch.Body, want) {
			t.Errorf("PATCH must carry %s:\n%s", want, patch.Body)
		}
	}
}

// TestDeleteTargetGroupUsesTheOddlyCasedMethod pins the SDK quirk: every operation on
// this resource is Targetgroups* except the delete, which is TargetGroupsDelete.
func TestDeleteTargetGroupUsesTheOddlyCasedMethod(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(tgAPI+"/tg-1", `{"id":"tg-1","properties":{"name":"app-pool","algorithm":"ROUND_ROBIN","protocol":"HTTP",
		"targets":[{"ip":"10.0.0.1","port":80,"weight":1},{"ip":"10.0.0.2","port":80,"weight":1}]}}`)

	preview, res := previewThenExecute(t, h, "delete_target_group", map[string]any{"target_group_id": "tg-1"})
	for _, want := range []string{"2 backends in this group", "several data centers"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing %q:\n%s", want, preview)
		}
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodDelete)
	if req.Path != tgAPI+"/tg-1" {
		t.Errorf("DELETE path = %s, want %s", req.Path, tgAPI+"/tg-1")
	}
}

func TestCreateNatGatewayTwoPhase(t *testing.T) {
	h := destructiveSetup(t)
	preview, res := previewThenExecute(t, h, "create_nat_gateway", map[string]any{
		"datacenter_id": dcID, "name": "egress", "public_ips": []string{"1.2.3.4"},
		"lans": []map[string]any{{"id": 2}},
	})
	for _, want := range []string{"egress", "1.2.3.4", "LAN 2"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing %q:\n%s", want, preview)
		}
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodPost)
	if req.Path != natAPI {
		t.Errorf("POST path = %s, want %s", req.Path, natAPI)
	}
}

func TestCreateNatGatewayRequiresPublicIps(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "create_nat_gateway", map[string]any{
		"datacenter_id": dcID, "name": "egress", "public_ips": []string{},
	})
	if !res.IsError || !strings.Contains(resultText(res), "at least one address") {
		t.Errorf("want a public_ips requirement error, got: %s", resultText(res))
	}
}

// TestUpdateNatGatewayCarriesRequiredFields is the fourth carry-forward case:
// name and publicIps are serialized unconditionally, so an unrelated change would
// otherwise leave the gateway with nothing to translate to.
func TestUpdateNatGatewayCarriesRequiredFields(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(natAPI+"/gw-1", `{"id":"gw-1","properties":{
		"name":"egress","publicIps":["1.2.3.4","1.2.3.5"],"lans":[{"id":2,"gatewayIps":["10.0.2.1/24"]}]}}`)

	res := callTool(t, h, "update_nat_gateway", map[string]any{
		"datacenter_id": dcID, "natgateway_id": "gw-1", "name": "renamed",
	})
	if res.IsError {
		t.Fatalf("update failed: %s", resultText(res))
	}
	reqs := h.log.allRequests()
	patch := reqs[len(reqs)-1]
	for _, want := range []string{`"name":"renamed"`, `"1.2.3.4"`, `"1.2.3.5"`, `"id":2`} {
		if !strings.Contains(patch.Body, want) {
			t.Errorf("PATCH must carry %s forward:\n%s", want, patch.Body)
		}
	}
}

// TestUpdateNatGatewayEmptyLansIsForwarded pins the asymmetry between the two list
// fields, which is deliberate and matches the spec: public_ips is required, so an
// empty list is refused before any call; lans is optional, so an empty list is a
// legal state and must reach the API rather than being swallowed or rejected.
// Confirmed live (2026-08-05): the gateway went AVAILABLE with lans: [].
func TestUpdateNatGatewayEmptyLansIsForwarded(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(natAPI+"/gw-1", `{"id":"gw-1","properties":{
		"name":"egress","publicIps":["1.2.3.4"],"lans":[{"id":2,"gatewayIps":["10.0.2.1/24"]}]}}`)

	res := callTool(t, h, "update_nat_gateway", map[string]any{
		"datacenter_id": dcID, "natgateway_id": "gw-1", "lans": []map[string]any{},
	})
	if res.IsError {
		t.Fatalf("an empty lans list is a valid state and must not be refused: %s", resultText(res))
	}
	reqs := h.log.allRequests()
	patch := reqs[len(reqs)-1]
	// Sent as [], not omitted: omitting it would carry the current LANs forward and
	// silently ignore what the caller asked for.
	if !strings.Contains(patch.Body, `"lans":[]`) {
		t.Errorf("PATCH must send lans as an empty list, not omit it:\n%s", patch.Body)
	}
	if strings.Contains(patch.Body, `"id":2`) {
		t.Errorf("the current LAN must not be carried forward when lans is explicitly empty:\n%s", patch.Body)
	}
}

// TestUpdateNatGatewayEmptyPublicIpsIsRefused is the other half of that asymmetry.
func TestUpdateNatGatewayEmptyPublicIpsIsRefused(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "update_nat_gateway", map[string]any{
		"datacenter_id": dcID, "natgateway_id": "gw-1", "public_ips": []string{},
	})
	if !res.IsError {
		t.Fatal("a gateway with no public IP has nothing to translate to; want a refusal")
	}
	if out := resultText(res); !strings.Contains(out, "omit the field entirely") {
		t.Errorf("refusal should name the way out:\n%s", out)
	}
	if n := len(h.log.allRequests()); n != 0 {
		t.Errorf("the refusal must land before any API call, got %d requests", n)
	}
}

func TestCreateNatGatewayRuleTwoPhase(t *testing.T) {
	h := destructiveSetup(t)
	preview, res := previewThenExecute(t, h, "create_nat_gateway_rule", map[string]any{
		"datacenter_id": dcID, "natgateway_id": "gw-1", "name": "egress-all",
		"source_subnet": "10.0.2.0/24", "public_ip": "1.2.3.4",
	})
	for _, want := range []string{"10.0.2.0/24", "1.2.3.4", "translated"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing %q:\n%s", want, preview)
		}
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodPost)
	if req.Path != natRules {
		t.Errorf("POST path = %s, want %s", req.Path, natRules)
	}
}

func TestNatGatewayRulePortValidation(t *testing.T) {
	base := map[string]any{
		"datacenter_id": dcID, "natgateway_id": "gw-1", "name": "r",
		"source_subnet": "10.0.2.0/24", "public_ip": "1.2.3.4",
	}
	tests := []struct {
		name    string
		extra   map[string]any
		wantMsg string
	}{
		{"ports need a protocol", map[string]any{"target_port_range_start": 80, "target_port_range_end": 90}, "require protocol"},
		{"ports rejected for ICMP", map[string]any{"protocol": "ICMP", "target_port_range_start": 80, "target_port_range_end": 90}, "not valid with protocol ICMP"},
		{"half-open range", map[string]any{"protocol": "TCP", "target_port_range_start": 80}, "must be given together"},
		{"inverted range", map[string]any{"protocol": "TCP", "target_port_range_start": 900, "target_port_range_end": 100}, "must not be greater than"},
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
			res := callTool(t, h, "create_nat_gateway_rule", args)
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

// TestUpdateNatGatewayRuleCarriesRequiredFields is the fifth and last carry-forward
// case in this batch.
func TestUpdateNatGatewayRuleCarriesRequiredFields(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(natRules+"/r-1", `{"id":"r-1","properties":{
		"name":"egress-all","sourceSubnet":"10.0.2.0/24","publicIp":"1.2.3.4","protocol":"ALL"}}`)

	res := callTool(t, h, "update_nat_gateway_rule", map[string]any{
		"datacenter_id": dcID, "natgateway_id": "gw-1", "rule_id": "r-1", "name": "renamed",
	})
	if res.IsError {
		t.Fatalf("update failed: %s", resultText(res))
	}
	patch := h.log.allRequests()[1]
	for _, want := range []string{`"name":"renamed"`, `"sourceSubnet":"10.0.2.0/24"`, `"publicIp":"1.2.3.4"`} {
		if !strings.Contains(patch.Body, want) {
			t.Errorf("PATCH must carry %s forward:\n%s", want, patch.Body)
		}
	}
	if strings.Contains(patch.Body, `"sourceSubnet":""`) {
		t.Errorf("PATCH must never send an empty sourceSubnet:\n%s", patch.Body)
	}
}

func TestDeleteNatGatewayBlastRadius(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(natAPI+"/gw-1", `{"id":"gw-1","properties":{"name":"egress","publicIps":["1.2.3.4"],
		"lans":[{"id":2},{"id":3}]},"entities":{"rules":{"items":[{"id":"r1"}]}}}`)

	preview, res := previewThenExecute(t, h, "delete_nat_gateway", map[string]any{
		"datacenter_id": dcID, "natgateway_id": "gw-1",
	})
	for _, want := range []string{"1 translation rules", "2 LANs that lose outbound internet access"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing %q:\n%s", want, preview)
		}
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	singleRequest(t, h, http.MethodDelete)
}

// TestClassicLoadBalancerUpdateNeedsNoCarryForward documents the one resource in this
// area whose properties are all optional, so its PATCH really is partial.
func TestClassicLoadBalancerUpdateNeedsNoCarryForward(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "update_loadbalancer", map[string]any{
		"datacenter_id": dcID, "loadbalancer_id": "lb-1", "name": "renamed",
	})
	if res.IsError {
		t.Fatalf("update failed: %s", resultText(res))
	}
	// One request only: no read needed, unlike every other load balancer here.
	req := singleRequest(t, h, http.MethodPatch)
	var body map[string]any
	if err := json.Unmarshal([]byte(req.Body), &body); err != nil {
		t.Fatalf("PATCH body is not JSON (%v): %s", err, req.Body)
	}
	if len(body) != 1 || body["name"] != "renamed" {
		t.Errorf("PATCH should contain only the name, got %s", req.Body)
	}
}

func TestDeleteClassicLoadBalancerBlastRadius(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(classicLB+"/lb-1", `{"id":"lb-1","properties":{"name":"legacy","ip":"1.2.3.4"},
		"entities":{"balancednics":{"items":[{"id":"n1"},{"id":"n2"}]}}}`)

	preview, res := previewThenExecute(t, h, "delete_loadbalancer", map[string]any{
		"datacenter_id": dcID, "loadbalancer_id": "lb-1",
	})
	if !strings.Contains(preview, "2 NICs that stop having traffic balanced to them") {
		t.Errorf("preview should count the balanced NICs:\n%s", preview)
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	singleRequest(t, h, http.MethodDelete)
}

// TestBalancedNicToolsAreNotRegistered documents the deliberate omission: attaching a
// NIC to a classic load balancer hits the same SDK defect as attach_lan_nic, where the
// smallest body the typed model can produce carries properties.lan = 0.
func TestBalancedNicToolsAreNotRegistered(t *testing.T) {
	names := toolNames(t, context.Background(), setupWithScope(t, tools.Scope{Write: true, Destructive: true}))
	for _, name := range []string{"attach_loadbalancer_nic", "detach_loadbalancer_nic", "attach_lan_nic", "attach_server_cdrom", "detach_server_cdrom"} {
		if names[name] {
			t.Errorf("%q is registered, but its attach body cannot be expressed correctly with the current SDK models — see the note in loadbalancer_write.go", name)
		}
	}
}
