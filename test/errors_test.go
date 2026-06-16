package test

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// callTool is a thin wrapper that fails the test on a protocol-level error.
func (h *testSetup) callTool(t *testing.T, name string, args map[string]any) *mcp.CallToolResult {
	t.Helper()
	res, err := h.session.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("CallTool(%q) returned protocol error: %v", name, err)
	}
	return res
}

// TestMissingRequiredArg verifies the MCP SDK rejects a call that omits a
// required input field before any HTTP request is made.
func TestMissingRequiredArg(t *testing.T) {
	h := setup(t)
	h.log.clear()

	// get_dns_zone requires zone_id.
	res, err := h.session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "get_dns_zone",
		Arguments: map[string]any{},
	})

	// The SDK may surface schema violations either as a protocol error or as an
	// error result — both are acceptable; what matters is that no API call was made.
	if err == nil && (res == nil || !res.IsError) {
		t.Errorf("expected schema validation failure for missing zone_id, got success")
	}
	if reqs := h.log.allRequests(); len(reqs) != 0 {
		t.Errorf("missing required arg made %d HTTP requests, want 0", len(reqs))
	}
}

// TestUnauthorized401Enriched verifies a backend 401 is wrapped with actionable
// token-configuration guidance before reaching the LLM.
func TestUnauthorized401Enriched(t *testing.T) {
	h := setup(t)
	h.log.clear()
	h.resp.setStatus(401, `{"httpStatus":401,"messages":[{"errorCode":"401","message":"Unauthorized"}]}`)

	res := h.callTool(t, "list_dns_zones", map[string]any{})
	if res == nil || !res.IsError {
		t.Fatalf("expected error result for 401, got: %+v", res)
	}
	text := resultText(res)
	for _, want := range []string{
		"401",
		"IONOS_TOKEN",
		".mcp.json",
		"restart the MCP client",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("enriched 401 output missing %q\ngot: %s", want, text)
		}
	}
	// Original API body is echoed for debugging.
	if !strings.Contains(text, "Unauthorized") {
		t.Errorf("enriched 401 should echo the original API body, got: %s", text)
	}
}

// TestServerError500PassThrough verifies non-401 errors are surfaced as errors
// but not rewritten with token guidance.
func TestServerError500PassThrough(t *testing.T) {
	h := setup(t)
	h.log.clear()
	h.resp.setStatus(500, `{"httpStatus":500,"messages":[{"message":"boom"}]}`)

	res := h.callTool(t, "list_datacenters", map[string]any{})
	if res == nil || !res.IsError {
		t.Fatalf("expected error result for 500, got: %+v", res)
	}
	text := resultText(res)
	if strings.Contains(text, "IONOS_TOKEN") {
		t.Errorf("500 must not be enriched with token guidance, got: %s", text)
	}
}

// TestZoneFileRawResult verifies the raw passthrough path: get_dns_zone_file
// returns a non-JSON (BIND zone file) body verbatim.
func TestZoneFileRawResult(t *testing.T) {
	h := setup(t)
	h.log.clear()

	zoneFile := "$ORIGIN example.com.\n@ 3600 IN SOA ns.example.com. hostmaster.example.com. 1 7200 3600 1209600 3600\n"
	h.resp.serve("/zones/z-1/zonefile", zoneFile)

	res := h.callTool(t, "get_dns_zone_file", map[string]any{"zone_id": "z-1"})
	if res != nil && res.IsError {
		t.Fatalf("get_dns_zone_file returned error: %s", resultText(res))
	}
	text := resultText(res)
	if !strings.Contains(text, "$ORIGIN example.com.") || !strings.Contains(text, "IN SOA") {
		t.Errorf("raw zone file not returned verbatim, got: %s", text)
	}
}
