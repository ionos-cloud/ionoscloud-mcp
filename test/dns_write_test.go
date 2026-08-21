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

// DNS write-tool tests. Unlike compute, DNS paths carry no /cloudapi/v6 prefix.
// Every DNS update is a PUT, so the carry-forward assertions are the core of this
// file: the SDK serializes each resource's identity fields unconditionally, and a
// PUT that omitted them would blank them out.

const (
	dnsZoneID      = "z-1"
	dnsRecordID    = "r-1"
	dnsSecZoneID   = "sz-1"
	dnsRevRecordID = "rr-1"

	dnsZonesPath      = "/zones"
	dnsZonePath       = dnsZonesPath + "/" + dnsZoneID
	dnsRecordsPath    = dnsZonePath + "/records"
	dnsRecordPath     = dnsRecordsPath + "/" + dnsRecordID
	dnsZonefilePath   = dnsZonePath + "/zonefile"
	dnsKeysPath       = dnsZonePath + "/keys"
	dnsSecZonesPath   = "/secondaryzones"
	dnsSecZonePath    = dnsSecZonesPath + "/" + dnsSecZoneID
	dnsAxfrPath       = dnsSecZonePath + "/axfr"
	dnsRevRecordsPath = "/reverserecords"
	dnsRevRecordPath  = dnsRevRecordsPath + "/" + dnsRevRecordID

	zoneFixture = `{"id":"z-1","properties":{"zoneName":"example.com","description":"prod zone","enabled":true},"metadata":{"state":"AVAILABLE","nameservers":["ns-ic.ui-dns.com"]}}`
	// A full record so a partial update has something to lose.
	recordFixture    = `{"id":"r-1","properties":{"name":"mail","type":"MX","content":"mail.example.com","ttl":7200,"priority":10,"enabled":true},"metadata":{"state":"AVAILABLE","fqdn":"mail.example.com","zoneId":"z-1"}}`
	secZoneFixture   = `{"id":"sz-1","properties":{"zoneName":"example.org","description":"mirror","primaryIps":["1.2.3.4","5.6.7.8"]},"metadata":{"state":"AVAILABLE","nameservers":["nscs.ui-dns.com"]}}`
	revRecordFixture = `{"id":"rr-1","properties":{"name":"mail.example.com","ip":"192.0.2.10","description":"mail server"},"metadata":{}}`
	keysFixture      = `{"id":"k-1","type":"dnsseckeys","metadata":{"zoneId":"z-1","items":[{"keyTag":49057,"digest":"CF58"}]},"properties":{"keyParameters":{"algorithm":"RSASHA256"},"nsecParameters":{"nsecMode":"NSEC3"}}}`
)

// dnsPutProperties returns the decoded properties object of the one PUT in the log.
// Every DNS update reads the resource first, so the log holds a GET as well.
func dnsPutProperties(t *testing.T, h *testSetup) map[string]any {
	t.Helper()
	var puts []recordedRequest
	for _, r := range h.log.allRequests() {
		if r.Method == http.MethodPut {
			puts = append(puts, r)
		}
	}
	if len(puts) != 1 {
		t.Fatalf("expected exactly 1 PUT, got %d: %+v", len(puts), h.log.allRequests())
	}
	var body struct {
		Properties map[string]any `json:"properties"`
	}
	if err := json.Unmarshal([]byte(puts[0].Body), &body); err != nil {
		t.Fatalf("decoding PUT body %q: %v", puts[0].Body, err)
	}
	return body.Properties
}

// assertKeys asserts a body object has exactly the given keys.
func assertDnsKeys(t *testing.T, got map[string]any, want ...string) {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("body has %d keys %v, want %d %v", len(got), keysOf(got), len(want), want)
	}
	for _, k := range want {
		if _, ok := got[k]; !ok {
			t.Errorf("body missing key %q; got %v", k, keysOf(got))
		}
	}
}

func keysOf(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// ---------- create_dns_zone ----------

func TestCreateDnsZoneTwoPhase(t *testing.T) {
	h := destructiveSetup(t)
	preview, res := previewThenExecute(t, h, "create_dns_zone", map[string]any{
		"zone_name": "example.com", "description": "prod zone",
	})

	for _, want := range []string{"example.com", "prod zone", "CREATE one primary DNS zone"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing %q:\n%s", want, preview)
		}
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodPost)
	if req.Path != dnsZonesPath {
		t.Errorf("POST path = %s, want %s", req.Path, dnsZonesPath)
	}
	for _, want := range []string{`"zoneName":"example.com"`, `"description":"prod zone"`} {
		if !strings.Contains(req.Body, want) {
			t.Errorf("POST body missing %s:\n%s", want, req.Body)
		}
	}
}

// TestCreateDnsZoneOmitsUnsetEnabled covers the constructor hazard: dnsSDK.NewZone
// injects enabled=true, so a body built from it would claim the caller asked for it.
func TestCreateDnsZoneOmitsUnsetEnabled(t *testing.T) {
	h := destructiveSetup(t)
	_, res := previewThenExecute(t, h, "create_dns_zone", map[string]any{"zone_name": "example.com"})
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodPost)
	if strings.Contains(req.Body, "enabled") {
		t.Errorf("POST body must not mention enabled when the caller omitted it:\n%s", req.Body)
	}
}

func TestCreateDnsZoneRequiresName(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "create_dns_zone", map[string]any{"zone_name": "  "})
	if !res.IsError {
		t.Fatal("a blank zone_name should be rejected")
	}
	assertNoMutation(t, h, "create_dns_zone validation")
}

func TestCreateDnsZoneBadTokenDoesNotCreate(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "create_dns_zone", map[string]any{
		"zone_name": "example.com", "confirmation_token": "deadbeef",
	})
	if !res.IsError {
		t.Fatal("a bogus confirmation_token must be rejected")
	}
	if n := len(h.log.allRequests()); n != 0 {
		t.Errorf("a rejected token must issue no requests, got %d", n)
	}
}

// ---------- update_dns_zone ----------

// TestUpdateDnsZoneCarriesZoneNameForward is the central PUT guarantee: zoneName is
// a non-pointer field the SDK always serializes, so an update that did not read it
// back first would send an empty zone name.
func TestUpdateDnsZoneCarriesZoneNameForward(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(dnsZonePath, zoneFixture)

	res := callTool(t, h, "update_dns_zone", map[string]any{"zone_id": dnsZoneID, "enabled": false})
	if res.IsError {
		t.Fatalf("update failed: %s", resultText(res))
	}
	reqs := h.log.allRequests()
	if len(reqs) != 2 || reqs[0].Method != http.MethodGet || reqs[1].Method != http.MethodPut {
		t.Fatalf("expected a GET then a PUT, got %+v", reqs)
	}
	var body struct {
		Properties map[string]any `json:"properties"`
	}
	if err := json.Unmarshal([]byte(reqs[1].Body), &body); err != nil {
		t.Fatalf("decoding PUT body: %v", err)
	}
	if body.Properties["zoneName"] != "example.com" {
		t.Errorf("PUT must carry zoneName forward, got %v", body.Properties["zoneName"])
	}
	if body.Properties["enabled"] != false {
		t.Errorf("PUT enabled = %v, want false", body.Properties["enabled"])
	}
	// description was not supplied, so it must be carried, not dropped.
	if body.Properties["description"] != "prod zone" {
		t.Errorf("PUT must carry description forward, got %v", body.Properties["description"])
	}
}

func TestUpdateDnsZoneRejectsEmptyRequest(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "update_dns_zone", map[string]any{"zone_id": dnsZoneID})
	if !res.IsError {
		t.Fatal("an update with no fields should be rejected")
	}
	if !strings.Contains(resultText(res), "nothing to update") {
		t.Errorf("error should say nothing to update: %s", resultText(res))
	}
	assertNoMutation(t, h, "update_dns_zone with no fields")
}

func TestUpdateDnsZoneNotFound(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serveStatus(dnsZonePath, http.StatusNotFound, `{"httpStatus":404}`)
	res := callTool(t, h, "update_dns_zone", map[string]any{"zone_id": dnsZoneID, "enabled": true})
	if !res.IsError {
		t.Fatal("updating a missing zone should be an error")
	}
	assertNoMutation(t, h, "update_dns_zone on a missing zone")
}

// ---------- delete_dns_zone ----------

func TestDeleteDnsZoneTwoPhase(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(dnsZonePath, zoneFixture)
	h.resp.serve("/records", `{"items":[{"id":"r-1"},{"id":"r-2"},{"id":"r-3"}]}`)
	h.resp.serve(dnsKeysPath, keysFixture)

	preview, res := previewThenExecute(t, h, "delete_dns_zone", map[string]any{"zone_id": dnsZoneID})
	for _, want := range []string{"example.com", "IRREVERSIBLE", "3 records", "1 DNSSEC signing keys"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing %q:\n%s", want, preview)
		}
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodDelete)
	if req.Path != dnsZonePath {
		t.Errorf("DELETE path = %s, want %s", req.Path, dnsZonePath)
	}
}

// TestDeleteDnsZoneCountsRecordsViaFilter pins the SDK workaround: the zone-scoped
// records endpoint exposes no limit, so its 100-item default would understate the
// blast radius. The count must come from /records?filter.zoneId with a limit.
func TestDeleteDnsZoneCountsRecordsViaFilter(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(dnsZonePath, zoneFixture)

	res := callTool(t, h, "delete_dns_zone", map[string]any{"zone_id": dnsZoneID})
	if res.IsError {
		t.Fatalf("preview failed: %s", resultText(res))
	}
	var found bool
	for _, r := range h.log.allRequests() {
		if r.Path != "/records" {
			continue
		}
		found = true
		if got := r.Query.Get("filter.zoneId"); got != dnsZoneID {
			t.Errorf("filter.zoneId = %q, want %q", got, dnsZoneID)
		}
		if got := r.Query.Get("limit"); got != "1000" {
			t.Errorf("limit = %q, want 1000", got)
		}
	}
	if !found {
		t.Errorf("the preview must count records via /records, got %+v", h.log.allRequests())
	}
}

func TestDeleteDnsZoneNotFound(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serveStatus(dnsZonePath, http.StatusNotFound, `{"httpStatus":404}`)
	res := callTool(t, h, "delete_dns_zone", map[string]any{"zone_id": dnsZoneID})
	if !res.IsError {
		t.Fatal("deleting a missing zone should be an error")
	}
	if !strings.Contains(resultText(res), "nothing to delete") {
		t.Errorf("error should say nothing to delete: %s", resultText(res))
	}
	assertNoMutation(t, h, "delete_dns_zone on a missing zone")
}

// ---------- create_dns_record ----------

func TestCreateDnsRecordTwoPhase(t *testing.T) {
	h := destructiveSetup(t)
	preview, res := previewThenExecute(t, h, "create_dns_record", map[string]any{
		"zone_id": dnsZoneID, "name": "www", "type": "a", "content": "192.0.2.1", "ttl": 300,
	})
	for _, want := range []string{"www", "A", "192.0.2.1", "300"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing %q:\n%s", want, preview)
		}
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodPost)
	if req.Path != dnsRecordsPath {
		t.Errorf("POST path = %s, want %s", req.Path, dnsRecordsPath)
	}
	// "a" must be normalised to the SDK's enum constant.
	for _, want := range []string{`"name":"www"`, `"type":"A"`, `"content":"192.0.2.1"`, `"ttl":300`} {
		if !strings.Contains(req.Body, want) {
			t.Errorf("POST body missing %s:\n%s", want, req.Body)
		}
	}
}

// TestCreateDnsRecordAllowsApexName covers the empty-name case: the spec's own
// A/AAAA/MX/TXT examples use "" for a record on the zone apex, so it must not be
// rejected the way a blank zone_name is.
func TestCreateDnsRecordAllowsApexName(t *testing.T) {
	h := destructiveSetup(t)
	preview, res := previewThenExecute(t, h, "create_dns_record", map[string]any{
		"zone_id": dnsZoneID, "name": "", "type": "A", "content": "192.0.2.1",
	})
	if !strings.Contains(preview, "zone apex") {
		t.Errorf("preview should name the apex explicitly:\n%s", preview)
	}
	if res.IsError {
		t.Fatalf("an apex record must be accepted: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodPost)
	if !strings.Contains(req.Body, `"name":""`) {
		t.Errorf("POST body should carry an empty name:\n%s", req.Body)
	}
}

// TestCreateDnsRecordOmitsUnsetTtl covers the constructor hazard: dnsSDK.NewRecord
// injects ttl=3600 and enabled=true.
func TestCreateDnsRecordOmitsUnsetTtl(t *testing.T) {
	h := destructiveSetup(t)
	_, res := previewThenExecute(t, h, "create_dns_record", map[string]any{
		"zone_id": dnsZoneID, "name": "www", "type": "A", "content": "192.0.2.1",
	})
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodPost)
	for _, unwanted := range []string{"ttl", "enabled", "priority"} {
		if strings.Contains(req.Body, unwanted) {
			t.Errorf("POST body must not mention %s when the caller omitted it:\n%s", unwanted, req.Body)
		}
	}
}

func TestCreateDnsRecordValidation(t *testing.T) {
	tests := []struct {
		name string
		args map[string]any
		want string
	}{
		{
			"unknown type",
			map[string]any{"zone_id": dnsZoneID, "name": "www", "type": "PTR", "content": "x"},
			"not a DNS record type",
		},
		{
			"ttl below the minimum",
			map[string]any{"zone_id": dnsZoneID, "name": "www", "type": "A", "content": "192.0.2.1", "ttl": 30},
			"out of range",
		},
		{
			"ttl above the maximum",
			map[string]any{"zone_id": dnsZoneID, "name": "www", "type": "A", "content": "192.0.2.1", "ttl": 700000},
			"out of range",
		},
		{
			"MX without a priority",
			map[string]any{"zone_id": dnsZoneID, "name": "", "type": "MX", "content": "mail.example.com"},
			"priority is required",
		},
		{
			"priority above the maximum",
			map[string]any{"zone_id": dnsZoneID, "name": "", "type": "MX", "content": "mail.example.com", "priority": 70000},
			"out of range",
		},
		{
			"missing content",
			map[string]any{"zone_id": dnsZoneID, "name": "www", "type": "A", "content": " "},
			"content is required",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := destructiveSetup(t)
			res := callTool(t, h, "create_dns_record", tt.args)
			if !res.IsError {
				t.Fatalf("%s should be rejected", tt.name)
			}
			if !strings.Contains(resultText(res), tt.want) {
				t.Errorf("error should mention %q, got: %s", tt.want, resultText(res))
			}
			assertNoMutation(t, h, "create_dns_record "+tt.name)
		})
	}
}

// TestCreateDnsRecordIgnoresPriorityForOtherTypes covers the spec's "ignored for all
// other types": sending it anyway would be echoing a value the caller cannot mean.
func TestCreateDnsRecordIgnoresPriorityForOtherTypes(t *testing.T) {
	h := destructiveSetup(t)
	_, res := previewThenExecute(t, h, "create_dns_record", map[string]any{
		"zone_id": dnsZoneID, "name": "www", "type": "A", "content": "192.0.2.1", "priority": 10,
	})
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodPost)
	if strings.Contains(req.Body, "priority") {
		t.Errorf("priority must not be sent for an A record:\n%s", req.Body)
	}
}

// ---------- update_dns_record ----------

// TestUpdateDnsRecordCarriesFieldsForward is the sharpest carry-forward case. The
// endpoint is a PUT; name/type/content are serialized unconditionally, and ttl,
// priority and enabled are pointers the SDK drops when nil while the spec gives ttl
// and enabled defaults. Changing only the TTL must preserve everything else.
func TestUpdateDnsRecordCarriesFieldsForward(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(dnsRecordPath, recordFixture)

	res := callTool(t, h, "update_dns_record", map[string]any{
		"zone_id": dnsZoneID, "record_id": dnsRecordID, "ttl": 600,
	})
	if res.IsError {
		t.Fatalf("update failed: %s", resultText(res))
	}
	reqs := h.log.allRequests()
	if len(reqs) != 2 || reqs[0].Method != http.MethodGet || reqs[1].Method != http.MethodPut {
		t.Fatalf("expected a GET then a PUT, got %+v", reqs)
	}
	if reqs[1].Path != dnsRecordPath {
		t.Errorf("PUT path = %s, want %s", reqs[1].Path, dnsRecordPath)
	}
	var body struct {
		Properties map[string]any `json:"properties"`
	}
	if err := json.Unmarshal([]byte(reqs[1].Body), &body); err != nil {
		t.Fatalf("decoding PUT body: %v", err)
	}
	want := map[string]any{
		"name": "mail", "type": "MX", "content": "mail.example.com",
		"ttl": float64(600), "priority": float64(10), "enabled": true,
	}
	for k, v := range want {
		if body.Properties[k] != v {
			t.Errorf("PUT %s = %v, want %v (carry-forward lost)", k, body.Properties[k], v)
		}
	}
	assertDnsKeys(t, body.Properties, "name", "type", "content", "ttl", "priority", "enabled")
}

// TestUpdateDnsToolsDoNotExposeIdentityFields pins that identity fields are not tool
// inputs. They are read and carried forward instead, so accepting them would be a
// rename the API treats as a delete-and-recreate.
func TestUpdateDnsToolsDoNotExposeIdentityFields(t *testing.T) {
	h := destructiveSetup(t)
	ctx := context.Background()

	forbidden := map[string][]string{
		"update_dns_zone":           {"zone_name"},
		"update_dns_record":         {"name", "type"},
		"update_dns_secondary_zone": {"zone_name"},
		"update_dns_reverse_record": {"ip"},
	}
	seen := map[string]bool{}
	for tool, err := range h.session.Tools(ctx, nil) {
		if err != nil {
			t.Fatalf("listing tools: %v", err)
		}
		names, ok := forbidden[tool.Name]
		if !ok {
			continue
		}
		seen[tool.Name] = true
		if tool.InputSchema == nil {
			t.Errorf("%s has no input schema", tool.Name)
			continue
		}
		// InputSchema is an untyped any, so the property set is read back through JSON.
		raw, err := json.Marshal(tool.InputSchema)
		if err != nil {
			t.Fatalf("marshalling %s input schema: %v", tool.Name, err)
		}
		var schema struct {
			Properties map[string]any `json:"properties"`
		}
		if err := json.Unmarshal(raw, &schema); err != nil {
			t.Fatalf("decoding %s input schema: %v", tool.Name, err)
		}
		if len(schema.Properties) == 0 {
			t.Fatalf("%s input schema has no properties: %s", tool.Name, raw)
		}
		for _, n := range names {
			if _, present := schema.Properties[n]; present {
				t.Errorf("%s must not accept %q: it is carried forward from the resource", tool.Name, n)
			}
		}
	}
	for name := range forbidden {
		if !seen[name] {
			t.Errorf("%s was not registered", name)
		}
	}
}

// ---------- delete_dns_record ----------

func TestDeleteDnsRecordTwoPhase(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(dnsRecordPath, recordFixture)

	preview, res := previewThenExecute(t, h, "delete_dns_record", map[string]any{
		"zone_id": dnsZoneID, "record_id": dnsRecordID,
	})
	for _, want := range []string{"mail.example.com", "MX", "IRREVERSIBLE"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing %q:\n%s", want, preview)
		}
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	if req := singleRequest(t, h, http.MethodDelete); req.Path != dnsRecordPath {
		t.Errorf("DELETE path = %s, want %s", req.Path, dnsRecordPath)
	}
}

// ---------- secondary zones ----------

func TestCreateDnsSecondaryZoneTwoPhase(t *testing.T) {
	h := destructiveSetup(t)
	preview, res := previewThenExecute(t, h, "create_dns_secondary_zone", map[string]any{
		"zone_name": "example.org", "primary_ips": []any{"1.2.3.4", "5.6.7.8"},
	})
	for _, want := range []string{"example.org", "1.2.3.4", "5.6.7.8"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing %q:\n%s", want, preview)
		}
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodPost)
	if req.Path != dnsSecZonesPath {
		t.Errorf("POST path = %s, want %s", req.Path, dnsSecZonesPath)
	}
	if !strings.Contains(req.Body, `"primaryIps":["1.2.3.4","5.6.7.8"]`) {
		t.Errorf("POST body missing primaryIps:\n%s", req.Body)
	}
}

func TestCreateDnsSecondaryZoneValidation(t *testing.T) {
	tests := []struct {
		name string
		args map[string]any
		want string
	}{
		{"no primary IPs", map[string]any{"zone_name": "example.org", "primary_ips": []any{}}, "at least one"},
		{"not an IP", map[string]any{"zone_name": "example.org", "primary_ips": []any{"nope"}}, "not an IPv4 or IPv6"},
		{"duplicate IPs", map[string]any{"zone_name": "example.org", "primary_ips": []any{"1.2.3.4", "1.2.3.4"}}, "twice"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := destructiveSetup(t)
			res := callTool(t, h, "create_dns_secondary_zone", tt.args)
			if !res.IsError {
				t.Fatalf("%s should be rejected", tt.name)
			}
			if !strings.Contains(resultText(res), tt.want) {
				t.Errorf("error should mention %q, got: %s", tt.want, resultText(res))
			}
			assertNoMutation(t, h, "create_dns_secondary_zone "+tt.name)
		})
	}
}

// TestUpdateDnsSecondaryZoneNeverSendsNullPrimaryIps covers the specific 400 the SDK
// invites: primaryIps is a non-pointer slice, so a nil value marshals to null.
func TestUpdateDnsSecondaryZoneNeverSendsNullPrimaryIps(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(dnsSecZonePath, secZoneFixture)

	res := callTool(t, h, "update_dns_secondary_zone", map[string]any{
		"secondary_zone_id": dnsSecZoneID, "description": "renamed",
	})
	if res.IsError {
		t.Fatalf("update failed: %s", resultText(res))
	}
	props := dnsPutProperties(t, h)
	if props["zoneName"] != "example.org" {
		t.Errorf("PUT must carry zoneName forward, got %v", props["zoneName"])
	}
	ips, ok := props["primaryIps"].([]any)
	if !ok || len(ips) != 2 {
		t.Fatalf("PUT must carry the two primaryIps forward, got %v", props["primaryIps"])
	}
	if props["description"] != "renamed" {
		t.Errorf("PUT description = %v, want renamed", props["description"])
	}
}

func TestUpdateDnsSecondaryZoneReplacesPrimaryIps(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(dnsSecZonePath, secZoneFixture)

	res := callTool(t, h, "update_dns_secondary_zone", map[string]any{
		"secondary_zone_id": dnsSecZoneID, "primary_ips": []any{"9.9.9.9"},
	})
	if res.IsError {
		t.Fatalf("update failed: %s", resultText(res))
	}
	props := dnsPutProperties(t, h)
	ips, ok := props["primaryIps"].([]any)
	if !ok || len(ips) != 1 || ips[0] != "9.9.9.9" {
		t.Errorf("primaryIps should be replaced, got %v", props["primaryIps"])
	}
}

func TestDeleteDnsSecondaryZoneTwoPhase(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(dnsSecZonePath, secZoneFixture)

	preview, res := previewThenExecute(t, h, "delete_dns_secondary_zone", map[string]any{
		"secondary_zone_id": dnsSecZoneID,
	})
	if !strings.Contains(preview, "example.org") {
		t.Errorf("preview missing the zone name:\n%s", preview)
	}
	// A secondary zone's records come from the primaries, so no record count is
	// fetched: implying a loss that is not one would be misleading.
	if strings.Contains(preview, "records") {
		t.Errorf("preview must not claim records are destroyed:\n%s", preview)
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	if req := singleRequest(t, h, http.MethodDelete); req.Path != dnsSecZonePath {
		t.Errorf("DELETE path = %s, want %s", req.Path, dnsSecZonePath)
	}
}

// TestStartDnsZoneTransferIsSingleCall pins that a transfer needs no confirmation:
// it only refreshes the IONOS-side copy.
func TestStartDnsZoneTransferIsSingleCall(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "start_dns_zone_transfer", map[string]any{"secondary_zone_id": dnsSecZoneID})
	if res.IsError {
		t.Fatalf("start_dns_zone_transfer failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodPut)
	if req.Path != dnsAxfrPath {
		t.Errorf("PUT path = %s, want %s", req.Path, dnsAxfrPath)
	}
	if !strings.Contains(resultText(res), "get_dns_secondary_zone_axfr") {
		t.Errorf("result should point at the status tool: %s", resultText(res))
	}
}

// ---------- reverse records ----------

func TestCreateDnsReverseRecordTwoPhase(t *testing.T) {
	h := destructiveSetup(t)
	preview, res := previewThenExecute(t, h, "create_dns_reverse_record", map[string]any{
		"name": "mail.example.com", "ip": "192.0.2.10",
	})
	for _, want := range []string{"mail.example.com", "192.0.2.10"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing %q:\n%s", want, preview)
		}
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodPost)
	if req.Path != dnsRevRecordsPath {
		t.Errorf("POST path = %s, want %s", req.Path, dnsRevRecordsPath)
	}
	if !strings.Contains(req.Body, `"ip":"192.0.2.10"`) {
		t.Errorf("POST body missing the ip:\n%s", req.Body)
	}
}

func TestCreateDnsReverseRecordRejectsBadIP(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "create_dns_reverse_record", map[string]any{"name": "mail.example.com", "ip": "nope"})
	if !res.IsError {
		t.Fatal("a non-IP should be rejected")
	}
	assertNoMutation(t, h, "create_dns_reverse_record with a bad IP")
}

func TestUpdateDnsReverseRecordCarriesIpForward(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(dnsRevRecordPath, revRecordFixture)

	res := callTool(t, h, "update_dns_reverse_record", map[string]any{
		"reverse_record_id": dnsRevRecordID, "name": "smtp.example.com",
	})
	if res.IsError {
		t.Fatalf("update failed: %s", resultText(res))
	}
	props := dnsPutProperties(t, h)
	if props["ip"] != "192.0.2.10" {
		t.Errorf("PUT must carry ip forward, got %v", props["ip"])
	}
	if props["name"] != "smtp.example.com" {
		t.Errorf("PUT name = %v, want smtp.example.com", props["name"])
	}
	if props["description"] != "mail server" {
		t.Errorf("PUT must carry description forward, got %v", props["description"])
	}
}

func TestDeleteDnsReverseRecordTwoPhase(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(dnsRevRecordPath, revRecordFixture)

	preview, res := previewThenExecute(t, h, "delete_dns_reverse_record", map[string]any{
		"reverse_record_id": dnsRevRecordID,
	})
	if !strings.Contains(preview, "192.0.2.10") {
		t.Errorf("preview missing the ip:\n%s", preview)
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	if req := singleRequest(t, h, http.MethodDelete); req.Path != dnsRevRecordPath {
		t.Errorf("DELETE path = %s, want %s", req.Path, dnsRevRecordPath)
	}
}

// ---------- DNSSEC ----------

func TestCreateDnsDnssecKeyTwoPhase(t *testing.T) {
	h := destructiveSetup(t)
	preview, res := previewThenExecute(t, h, "create_dns_zone_dnssec_key", map[string]any{
		"zone_id": dnsZoneID, "validity": 120,
	})
	// The defaults must be visible in the preview, not silently applied.
	for _, want := range []string{"RSASHA256", "4096", "2048", "NSEC3", "120"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing %q:\n%s", want, preview)
		}
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodPost)
	if req.Path != dnsKeysPath {
		t.Errorf("POST path = %s, want %s", req.Path, dnsKeysPath)
	}
	// All three nsecParameters fields are non-pointer, so all must be present even
	// though only validity was supplied.
	for _, want := range []string{`"algorithm":"RSASHA256"`, `"kskBits":4096`, `"zskBits":2048`,
		`"nsecMode":"NSEC3"`, `"nsec3Iterations":0`, `"nsec3SaltBits":64`, `"validity":120`} {
		if !strings.Contains(req.Body, want) {
			t.Errorf("POST body missing %s:\n%s", want, req.Body)
		}
	}
}

func TestCreateDnsDnssecKeyValidation(t *testing.T) {
	tests := []struct {
		name string
		args map[string]any
		want string
	}{
		{"validity too low", map[string]any{"zone_id": dnsZoneID, "validity": 30}, "out of range"},
		{"validity too high", map[string]any{"zone_id": dnsZoneID, "validity": 400}, "out of range"},
		{"bad algorithm", map[string]any{"zone_id": dnsZoneID, "validity": 120, "algorithm": "ECDSAP256SHA256"}, "RSASHA256 is the only"},
		{"bad key length", map[string]any{"zone_id": dnsZoneID, "validity": 120, "ksk_bits": 3072}, "1024, 2048 or 4096"},
		// The spec states this invariant in prose only; the SDK does not check it.
		{"ksk smaller than zsk", map[string]any{"zone_id": dnsZoneID, "validity": 120, "ksk_bits": 1024, "zsk_bits": 4096}, "greater than or equal"},
		{"bad nsec mode", map[string]any{"zone_id": dnsZoneID, "validity": 120, "nsec_mode": "NSEC5"}, "use NSEC or NSEC3"},
		{"too many iterations", map[string]any{"zone_id": dnsZoneID, "validity": 120, "nsec3_iterations": 99}, "out of range"},
		{"salt bits out of range", map[string]any{"zone_id": dnsZoneID, "validity": 120, "nsec3_salt_bits": 256}, "out of range"},
		{"salt bits not a multiple of 8", map[string]any{"zone_id": dnsZoneID, "validity": 120, "nsec3_salt_bits": 100}, "multiple of 8"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := destructiveSetup(t)
			res := callTool(t, h, "create_dns_zone_dnssec_key", tt.args)
			if !res.IsError {
				t.Fatalf("%s should be rejected", tt.name)
			}
			if !strings.Contains(resultText(res), tt.want) {
				t.Errorf("error should mention %q, got: %s", tt.want, resultText(res))
			}
			assertNoMutation(t, h, "create_dns_zone_dnssec_key "+tt.name)
		})
	}
}

// TestDeleteDnsDnssecKeyReturnsTheApiResponse pins that the execute phase hands back
// what the API said rather than a summary of it. The response is the only thing that
// separates a queued request from a rejected one — a hand-written success message hid
// a 409 ("the zone has too many operations in progress") behind prose once already.
func TestDeleteDnsDnssecKeyReturnsTheApiResponse(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(dnsKeysPath, keysFixture)

	preview, res := previewThenExecute(t, h, "delete_dns_zone_dnssec_key", map[string]any{"zone_id": dnsZoneID})
	for _, want := range []string{"DS RECORD", "REGISTRAR", "SERVFAIL"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview must warn about %q:\n%s", want, preview)
		}
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	if req := singleRequest(t, h, http.MethodDelete); req.Path != dnsKeysPath {
		t.Errorf("DELETE path = %s, want %s", req.Path, dnsKeysPath)
	}
	// The mock echoes the keys fixture, so the API's own JSON must come back through.
	if !strings.Contains(resultText(res), "dnsseckeys") {
		t.Errorf("result must carry the API response verbatim, got: %s", resultText(res))
	}
}

// TestDeleteDnsDnssecKeySurfacesConflict is the case that motivated the above: a 409
// with an errorCode must reach the caller intact, not be flattened into "failed".
func TestDeleteDnsDnssecKeySurfacesConflict(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(dnsKeysPath, keysFixture)

	res := callTool(t, h, "delete_dns_zone_dnssec_key", map[string]any{"zone_id": dnsZoneID})
	token := extractToken(t, resultText(res))
	h.resp.serveStatus(dnsKeysPath, http.StatusConflict,
		`{"httpStatus":409,"messages":[{"errorCode":"paas-dns-rest-0513","message":"the zone has too many operations in progress, please retry later"}]}`)
	res = callTool(t, h, "delete_dns_zone_dnssec_key",
		map[string]any{"zone_id": dnsZoneID, "confirmation_token": token})

	if !res.IsError {
		t.Fatal("a 409 must surface as an error result")
	}
	for _, want := range []string{"409", "paas-dns-rest-0513", "too many operations in progress"} {
		if !strings.Contains(resultText(res), want) {
			t.Errorf("result must preserve %q from the API body: %s", want, resultText(res))
		}
	}
}

// ---------- zone file import ----------

func TestImportDnsZoneFileTwoPhase(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(dnsZonePath, zoneFixture)
	h.resp.serve("/records", `{"items":[{"id":"r-1"},{"id":"r-2"}]}`)
	zoneFile := "$ORIGIN example.com.\n$TTL 3600\nwww IN A 192.0.2.1\n"

	preview, res := previewThenExecute(t, h, "import_dns_zone_file", map[string]any{
		"zone_id": dnsZoneID, "zone_file": zoneFile,
	})
	for _, want := range []string{"REPLACE every record", "2 existing records", "example.com"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing %q:\n%s", want, preview)
		}
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodPut)
	if req.Path != dnsZonefilePath {
		t.Errorf("PUT path = %s, want %s", req.Path, dnsZonefilePath)
	}
	// The body is the raw zone file, not a JSON envelope — the endpoint is text/plain.
	if req.Body != zoneFile {
		t.Errorf("PUT body should be the zone file verbatim, got %q", req.Body)
	}
}

func TestImportDnsZoneFileRequiresAFile(t *testing.T) {
	h := destructiveSetup(t)
	res := callTool(t, h, "import_dns_zone_file", map[string]any{"zone_id": dnsZoneID, "zone_file": "  "})
	if !res.IsError {
		t.Fatal("an empty zone_file should be rejected")
	}
	assertNoMutation(t, h, "import_dns_zone_file with no file")
}

// ---------- cross-cutting ----------

// TestDnsReadToolsAreAnnotatedReadOnly covers the annotation backfill: before the
// migration the DNS reads used bare mcp.AddTool and carried no annotations, leaving
// clients unable to tell a read from a mutation without parsing the name.
func TestDnsReadToolsAreAnnotatedReadOnly(t *testing.T) {
	h := setup(t) // read-only scope: only the read tools register
	ctx := context.Background()

	want := map[string]bool{
		"list_dns_zones": true, "get_dns_zone": true, "get_dns_zone_file": true,
		"list_dns_records": true, "list_dns_zone_records": true, "get_dns_record": true,
		"list_dns_secondary_zone_records": true,
		"list_dns_reverse_records":        true, "get_dns_reverse_record": true,
		"list_dns_secondary_zones": true, "get_dns_secondary_zone": true,
		"get_dns_secondary_zone_axfr": true,
		"list_dns_zone_dnssec_keys":   true, "get_dns_quota": true,
	}
	seen := map[string]bool{}
	for tool, err := range h.session.Tools(ctx, nil) {
		if err != nil {
			t.Fatalf("listing tools: %v", err)
		}
		if !want[tool.Name] {
			continue
		}
		seen[tool.Name] = true
		if tool.Annotations == nil {
			t.Errorf("%s has no annotations; it must carry ReadOnlyHint", tool.Name)
			continue
		}
		if !tool.Annotations.ReadOnlyHint {
			t.Errorf("%s ReadOnlyHint = false, want true", tool.Name)
		}
	}
	for name := range want {
		if !seen[name] {
			t.Errorf("read tool %s was not registered", name)
		}
	}
}

func dnsWriteToolNames() []string {
	return []string{
		"create_dns_zone", "update_dns_zone",
		"create_dns_record", "update_dns_record",
		"create_dns_secondary_zone", "update_dns_secondary_zone", "start_dns_zone_transfer",
		"create_dns_reverse_record", "update_dns_reverse_record",
		"create_dns_zone_dnssec_key",
	}
}

func dnsDestructiveToolNames() []string {
	return []string{
		"delete_dns_zone", "delete_dns_record", "delete_dns_secondary_zone",
		"delete_dns_reverse_record", "delete_dns_zone_dnssec_key", "import_dns_zone_file",
	}
}

// TestDnsWriteToolsAreScopeGated pins which DNS tools each scope exposes. The
// notable entry is import_dns_zone_file: it is a PUT, but it deletes every record in
// the zone, so it must sit behind destructive rather than write.
func TestDnsWriteToolsAreScopeGated(t *testing.T) {
	reads := []string{
		"list_dns_zones", "get_dns_zone", "get_dns_zone_file", "list_dns_records",
		"list_dns_zone_records", "get_dns_record", "list_dns_secondary_zone_records",
		"list_dns_reverse_records", "get_dns_reverse_record", "list_dns_secondary_zones",
		"get_dns_secondary_zone", "get_dns_secondary_zone_axfr", "list_dns_zone_dnssec_keys",
		"get_dns_quota",
	}
	writes := dnsWriteToolNames()
	destructives := dnsDestructiveToolNames()

	tests := []struct {
		name    string
		scope   tools.Scope
		present []string
		absent  []string
	}{
		{"read only", tools.Scope{}, reads, append(append([]string{}, writes...), destructives...)},
		{"write", tools.Scope{Write: true}, append(append([]string{}, reads...), writes...), destructives},
		{
			"destructive",
			tools.Scope{Write: true, Destructive: true},
			append(append(append([]string{}, reads...), writes...), destructives...),
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := setupWithScope(t, tt.scope)
			names := toolNames(t, context.Background(), h)
			for _, n := range tt.present {
				if !names[n] {
					t.Errorf("scope %s: %s should be registered", tt.scope, n)
				}
			}
			for _, n := range tt.absent {
				if names[n] {
					t.Errorf("scope %s: %s must NOT be registered", tt.scope, n)
				}
			}
		})
	}
}

// TestDnsWriteToolAnnotations pins the annotations a client uses to decide whether
// to prompt before a call.
func TestDnsWriteToolAnnotations(t *testing.T) {
	h := destructiveSetup(t)
	ctx := context.Background()

	want := map[string]struct{ destructive, idempotent bool }{
		"create_dns_zone":            {false, false},
		"update_dns_zone":            {false, true},
		"delete_dns_zone":            {true, true},
		"create_dns_record":          {false, false},
		"update_dns_record":          {false, true},
		"delete_dns_record":          {true, true},
		"create_dns_secondary_zone":  {false, false},
		"update_dns_secondary_zone":  {false, true},
		"delete_dns_secondary_zone":  {true, true},
		"create_dns_reverse_record":  {false, false},
		"update_dns_reverse_record":  {false, true},
		"delete_dns_reverse_record":  {true, true},
		"create_dns_zone_dnssec_key": {false, false},
		"delete_dns_zone_dnssec_key": {true, true},
		"start_dns_zone_transfer":    {false, true},
		"import_dns_zone_file":       {true, true},
	}
	seen := map[string]bool{}
	for tool, err := range h.session.Tools(ctx, nil) {
		if err != nil {
			t.Fatalf("listing tools: %v", err)
		}
		w, ok := want[tool.Name]
		if !ok {
			continue
		}
		seen[tool.Name] = true
		if tool.Annotations == nil {
			t.Errorf("%s has no annotations", tool.Name)
			continue
		}
		if tool.Annotations.ReadOnlyHint {
			t.Errorf("%s ReadOnlyHint = true, want false", tool.Name)
		}
		if tool.Annotations.DestructiveHint == nil || *tool.Annotations.DestructiveHint != w.destructive {
			t.Errorf("%s DestructiveHint = %v, want %v", tool.Name, tool.Annotations.DestructiveHint, w.destructive)
		}
		if tool.Annotations.IdempotentHint != w.idempotent {
			t.Errorf("%s IdempotentHint = %v, want %v", tool.Name, tool.Annotations.IdempotentHint, w.idempotent)
		}
	}
	for name := range want {
		if !seen[name] {
			t.Errorf("write tool %s was not registered", name)
		}
	}
}

// TestDnsWriteToolsDeclareCompletionSemantics covers the per-operation split: DNS
// mutations are mostly 202, but a reverse record's create and update answer with the
// finished record and expose no state to poll, so promising a polling step would
// send the model looking for a field that does not exist.
func TestDnsWriteToolsDeclareCompletionSemantics(t *testing.T) {
	h := destructiveSetup(t)
	ctx := context.Background()

	async := map[string]bool{
		"create_dns_zone": true, "update_dns_zone": true, "delete_dns_zone": true,
		"create_dns_record": true, "update_dns_record": true, "delete_dns_record": true,
		"create_dns_secondary_zone": true, "update_dns_secondary_zone": true,
		"delete_dns_secondary_zone": true, "create_dns_zone_dnssec_key": true,
		"delete_dns_zone_dnssec_key": true, "start_dns_zone_transfer": true,
		"delete_dns_reverse_record": true,
	}
	sync := map[string]bool{
		"create_dns_reverse_record": true, "update_dns_reverse_record": true,
		"import_dns_zone_file": true,
	}
	seen := map[string]bool{}
	for tool, err := range h.session.Tools(ctx, nil) {
		if err != nil {
			t.Fatalf("listing tools: %v", err)
		}
		switch {
		case async[tool.Name]:
			seen[tool.Name] = true
			if !strings.Contains(tool.Description, "Asynchronous (202)") {
				t.Errorf("%s must declare that it is asynchronous: %s", tool.Name, tool.Description)
			}
		case sync[tool.Name]:
			seen[tool.Name] = true
			if !strings.Contains(tool.Description, "Synchronous") {
				t.Errorf("%s must say it is synchronous, so no polling step is invented: %s", tool.Name, tool.Description)
			}
			if strings.Contains(tool.Description, "Asynchronous (202)") {
				t.Errorf("%s must not claim to be asynchronous: %s", tool.Name, tool.Description)
			}
		}
	}
	for name := range async {
		if !seen[name] {
			t.Errorf("%s was not registered", name)
		}
	}
	for name := range sync {
		if !seen[name] {
			t.Errorf("%s was not registered", name)
		}
	}
}

// TestDnsTokensAreBoundToTheirTarget covers the two ways a token must fail: pointing
// at a different resource, and being replayed after use.
func TestDnsTokensAreBoundToTheirTarget(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(dnsZonePath, zoneFixture)

	res := callTool(t, h, "delete_dns_zone", map[string]any{"zone_id": dnsZoneID})
	token := extractToken(t, resultText(res))

	// Wrong target.
	h.resp.serve(dnsZonesPath+"/z-2", strings.Replace(zoneFixture, `"z-1"`, `"z-2"`, 1))
	h.log.clear()
	res = callTool(t, h, "delete_dns_zone", map[string]any{"zone_id": "z-2", "confirmation_token": token})
	if !res.IsError {
		t.Fatal("a token minted for z-1 must not delete z-2")
	}
	assertNoMutation(t, h, "delete_dns_zone with a mismatched token")

	// Right target: spends the token.
	h.log.clear()
	res = callTool(t, h, "delete_dns_zone", map[string]any{"zone_id": dnsZoneID, "confirmation_token": token})
	if res.IsError {
		t.Fatalf("the matching target should succeed: %s", resultText(res))
	}
	singleRequest(t, h, http.MethodDelete)

	// Replay.
	h.log.clear()
	res = callTool(t, h, "delete_dns_zone", map[string]any{"zone_id": dnsZoneID, "confirmation_token": token})
	if !res.IsError {
		t.Fatal("a spent token must not be reusable")
	}
	assertNoMutation(t, h, "delete_dns_zone replaying a spent token")
}

// TestDeleteDnsZoneDynamicParity checks both phases survive the dynamic dispatcher,
// which forwards calls through a shared session and so must carry the caller id that
// binds the token.
func TestDeleteDnsZoneDynamicParity(t *testing.T) {
	h := setupDynamicWithScope(t, tools.Scope{Write: true, Destructive: true})
	ctx := context.Background()
	h.resp.serve(dnsZonePath, zoneFixture)

	res, err := h.session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "ionos_call_tool",
		Arguments: map[string]any{"name": "delete_dns_zone", "arguments": map[string]any{"zone_id": dnsZoneID}},
	})
	if err != nil {
		t.Fatalf("dynamic preview: %v", err)
	}
	if res.IsError {
		t.Fatalf("dynamic preview must not be an error: %s", resultText(res))
	}
	token := extractToken(t, resultText(res))

	h.log.clear()
	res, err = h.session.CallTool(ctx, &mcp.CallToolParams{
		Name: "ionos_call_tool",
		Arguments: map[string]any{
			"name":      "delete_dns_zone",
			"arguments": map[string]any{"zone_id": dnsZoneID, "confirmation_token": token},
		},
	})
	if err != nil {
		t.Fatalf("dynamic execute: %v", err)
	}
	if res.IsError {
		t.Fatalf("dynamic execute must succeed: %s", resultText(res))
	}
	reqs := h.log.allRequests()
	if len(reqs) != 1 || reqs[0].Method != http.MethodDelete {
		t.Fatalf("expected one DELETE, got %+v", reqs)
	}
}

// TestImportDnsZoneFileIsGatedAsDestructive is the point of the new import_ action
// verb. Named update_*, the tool would classify by HTTP method as a mere write, and a
// write-scoped server could wipe every record in a zone.
func TestImportDnsZoneFileIsGatedAsDestructive(t *testing.T) {
	w := toolNames(t, context.Background(), setupWithScope(t, tools.Scope{Write: true}))
	if w["import_dns_zone_file"] {
		t.Error("a write-only scope must NOT expose import_dns_zone_file: it deletes every record in the zone")
	}
	d := toolNames(t, context.Background(), setupWithScope(t, tools.Scope{Write: true, Destructive: true}))
	if !d["import_dns_zone_file"] {
		t.Error("a destructive scope must expose import_dns_zone_file")
	}
}

// TestDnsUpdateBodiesContainOnlyExpectedFields is the DNS counterpart of
// TestUpdateBodiesContainOnlyRequestedFields, which covers PATCH bodies only. DNS
// updates are PUT, so the expected key set is the caller's fields plus every field
// carried forward — a PUT replaces the resource's properties, so anything missing
// here is a field the update would silently blank.
func TestDnsUpdateBodiesContainOnlyExpectedFields(t *testing.T) {
	tests := []struct {
		tool      string
		args      map[string]any
		wantKeys  []string
		primePath string
		primeBody string
	}{
		{
			tool:      "update_dns_zone",
			args:      map[string]any{"zone_id": dnsZoneID, "enabled": false},
			wantKeys:  []string{"zoneName", "description", "enabled"},
			primePath: dnsZonePath,
			primeBody: zoneFixture,
		},
		{
			tool:      "update_dns_record",
			args:      map[string]any{"zone_id": dnsZoneID, "record_id": dnsRecordID, "content": "mx2.example.com"},
			wantKeys:  []string{"name", "type", "content", "ttl", "priority", "enabled"},
			primePath: dnsRecordPath,
			primeBody: recordFixture,
		},
		{
			tool:      "update_dns_secondary_zone",
			args:      map[string]any{"secondary_zone_id": dnsSecZoneID, "description": "new"},
			wantKeys:  []string{"zoneName", "description", "primaryIps"},
			primePath: dnsSecZonePath,
			primeBody: secZoneFixture,
		},
		{
			tool:      "update_dns_reverse_record",
			args:      map[string]any{"reverse_record_id": dnsRevRecordID, "name": "smtp.example.com"},
			wantKeys:  []string{"name", "ip", "description"},
			primePath: dnsRevRecordPath,
			primeBody: revRecordFixture,
		},
	}

	for _, tt := range tests {
		t.Run(tt.tool, func(t *testing.T) {
			h := destructiveSetup(t)
			h.resp.serve(tt.primePath, tt.primeBody)
			res := callTool(t, h, tt.tool, tt.args)
			if res.IsError {
				t.Fatalf("%s failed: %s", tt.tool, resultText(res))
			}
			assertDnsKeys(t, dnsPutProperties(t, h), tt.wantKeys...)
		})
	}
}

// TestUpdateDnsRecordOnAnARecordSendsNoPriority covers the other half of the
// carry-forward rule: a record with no priority must not acquire one, since the SDK
// would send whatever the struct holds.
func TestUpdateDnsRecordOnAnARecordSendsNoPriority(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(dnsRecordPath, `{"id":"r-1","properties":{"name":"www","type":"A","content":"192.0.2.1","ttl":3600,"enabled":true},"metadata":{"state":"AVAILABLE","fqdn":"www.example.com","zoneId":"z-1"}}`)

	res := callTool(t, h, "update_dns_record", map[string]any{
		"zone_id": dnsZoneID, "record_id": dnsRecordID, "content": "192.0.2.2",
	})
	if res.IsError {
		t.Fatalf("update failed: %s", resultText(res))
	}
	assertDnsKeys(t, dnsPutProperties(t, h), "name", "type", "content", "ttl", "enabled")
}
