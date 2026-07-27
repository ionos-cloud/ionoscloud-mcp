package test

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

const dcPath = "/cloudapi/v6/datacenters"

// extractToken pulls the confirmation_token value out of a two-phase preview.
func extractToken(t *testing.T, preview string) string {
	t.Helper()
	for _, line := range strings.Split(preview, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "confirmation_token:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "confirmation_token:"))
		}
	}
	t.Fatalf("no confirmation_token found in preview:\n%s", preview)
	return ""
}

// toolNames lists the tool names currently exposed on a session.
func toolNames(t *testing.T, ctx context.Context, h *testSetup) map[string]bool {
	t.Helper()
	names := map[string]bool{}
	for tool, err := range h.session.Tools(ctx, nil) {
		if err != nil {
			t.Fatalf("listing tools: %v", err)
		}
		names[tool.Name] = true
	}
	return names
}

func TestCreateDatacenterTwoPhase(t *testing.T) {
	h := setupWithScope(t, tools.Scope{Write: true, Destructive: true})
	ctx := context.Background()

	// Phase 1: no token -> preview + token, and NO POST.
	res, err := h.session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "create_datacenter",
		Arguments: map[string]any{"name": "my-dc", "location": "de/fra"},
	})
	if err != nil {
		t.Fatalf("preview call: %v", err)
	}
	if res.IsError {
		t.Fatalf("preview must not be an error: %s", resultText(res))
	}
	preview := resultText(res)
	if !strings.Contains(preview, "my-dc") || !strings.Contains(preview, "de/fra") {
		t.Errorf("preview missing name/location:\n%s", preview)
	}
	for _, r := range h.log.allRequests() {
		if r.Method == http.MethodPost {
			t.Fatal("preview phase must not POST")
		}
	}
	token := extractToken(t, preview)

	// Phase 2: token -> exactly one POST carrying name+location.
	h.log.clear()
	h.resp.serve(dcPath, `{"id":"dc-123"}`)
	res, err = h.session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "create_datacenter",
		Arguments: map[string]any{"name": "my-dc", "location": "de/fra", "confirmation_token": token},
	})
	if err != nil {
		t.Fatalf("execute call: %v", err)
	}
	if res.IsError {
		t.Fatalf("execute must succeed: %s", resultText(res))
	}
	reqs := h.log.allRequests()
	if len(reqs) != 1 || reqs[0].Method != http.MethodPost || reqs[0].Path != dcPath {
		t.Fatalf("expected one POST %s, got %+v", dcPath, reqs)
	}
	if !strings.Contains(reqs[0].Body, `"location":"de/fra"`) || !strings.Contains(reqs[0].Body, `"name":"my-dc"`) {
		t.Errorf("POST body missing name/location: %s", reqs[0].Body)
	}
}

func TestCreateDatacenterBadToken(t *testing.T) {
	h := setupWithScope(t, tools.Scope{Write: true})
	ctx := context.Background()
	res, err := h.session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "create_datacenter",
		Arguments: map[string]any{"name": "x", "location": "de/fra", "confirmation_token": "bogus"},
	})
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if !res.IsError {
		t.Fatal("create with an unknown token must be an error")
	}
	if len(h.log.allRequests()) != 0 {
		t.Fatalf("create with a bad token must not hit the API, got %+v", h.log.allRequests())
	}
}

func TestUpdateDatacenter(t *testing.T) {
	h := setupWithScope(t, tools.Scope{Write: true})
	ctx := context.Background()
	h.resp.serve(dcPath+"/dc-1", `{"id":"dc-1"}`)

	res, err := h.session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "update_datacenter",
		Arguments: map[string]any{"datacenter_id": "dc-1", "name": "renamed"},
	})
	if err != nil {
		t.Fatalf("update call: %v", err)
	}
	if res.IsError {
		t.Fatalf("update must succeed: %s", resultText(res))
	}
	reqs := h.log.allRequests()
	if len(reqs) != 1 || reqs[0].Method != http.MethodPatch || reqs[0].Path != dcPath+"/dc-1" {
		t.Fatalf("expected one PATCH %s/dc-1, got %+v", dcPath, reqs)
	}
	if !strings.Contains(reqs[0].Body, `"name":"renamed"`) {
		t.Errorf("PATCH body missing name: %s", reqs[0].Body)
	}
	if strings.Contains(reqs[0].Body, "location") {
		t.Errorf("PATCH body must not carry the immutable location: %s", reqs[0].Body)
	}
}

func TestDeleteDatacenterTwoPhase(t *testing.T) {
	h := setupWithScope(t, tools.Scope{Write: true, Destructive: true})
	ctx := context.Background()
	h.resp.serve(dcPath+"/dc-1", `{"id":"dc-1","properties":{"name":"doomed","location":"de/txl"},"entities":{"servers":{"items":[{"id":"s1"},{"id":"s2"}]},"volumes":{"items":[{"id":"v1"}]}}}`)

	// Phase 1: preview -> one GET, no DELETE, blast-radius counts + token.
	res, err := h.session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "delete_datacenter",
		Arguments: map[string]any{"datacenter_id": "dc-1"},
	})
	if err != nil {
		t.Fatalf("preview call: %v", err)
	}
	if res.IsError {
		t.Fatalf("preview must not be an error: %s", resultText(res))
	}
	preview := resultText(res)
	if !strings.Contains(preview, "2 servers") || !strings.Contains(preview, "1 volumes") {
		t.Errorf("preview missing blast-radius counts:\n%s", preview)
	}
	reqs := h.log.allRequests()
	if len(reqs) != 1 || reqs[0].Method != http.MethodGet {
		t.Fatalf("preview should make exactly one GET, got %+v", reqs)
	}
	token := extractToken(t, preview)

	// Phase 2: token -> DELETE.
	h.log.clear()
	res, err = h.session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "delete_datacenter",
		Arguments: map[string]any{"datacenter_id": "dc-1", "confirmation_token": token},
	})
	if err != nil {
		t.Fatalf("execute call: %v", err)
	}
	if res.IsError {
		t.Fatalf("execute must succeed: %s", resultText(res))
	}
	reqs = h.log.allRequests()
	if len(reqs) != 1 || reqs[0].Method != http.MethodDelete || reqs[0].Path != dcPath+"/dc-1" {
		t.Fatalf("expected one DELETE %s/dc-1, got %+v", dcPath, reqs)
	}
}

func TestDeleteDatacenterWrongTargetToken(t *testing.T) {
	h := setupWithScope(t, tools.Scope{Write: true, Destructive: true})
	ctx := context.Background()
	h.resp.serve(dcPath+"/dc-A", `{"id":"dc-A"}`)

	res, _ := h.session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "delete_datacenter",
		Arguments: map[string]any{"datacenter_id": "dc-A"},
	})
	token := extractToken(t, resultText(res))

	// Using DC-A's token against DC-B must be rejected with no DELETE.
	h.log.clear()
	h.resp.serve(dcPath+"/dc-B", `{"id":"dc-B"}`)
	res, err := h.session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "delete_datacenter",
		Arguments: map[string]any{"datacenter_id": "dc-B", "confirmation_token": token},
	})
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if !res.IsError {
		t.Fatal("a token minted for DC-A must not delete DC-B")
	}
	for _, r := range h.log.allRequests() {
		if r.Method == http.MethodDelete {
			t.Fatal("mismatched token must not issue a DELETE")
		}
	}
}

func TestDeleteDatacenterNotFound(t *testing.T) {
	h := setupWithScope(t, tools.Scope{Write: true, Destructive: true})
	ctx := context.Background()
	h.resp.serveStatus(dcPath+"/ghost", http.StatusNotFound, `{"messages":[{"errorCode":"404"}]}`)

	res, err := h.session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "delete_datacenter",
		Arguments: map[string]any{"datacenter_id": "ghost"},
	})
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if !res.IsError {
		t.Fatalf("deleting a missing datacenter should be an error: %s", resultText(res))
	}
	if !strings.Contains(resultText(res), "does not exist") {
		t.Errorf("want a 'does not exist' message, got: %s", resultText(res))
	}
	for _, r := range h.log.allRequests() {
		if r.Method == http.MethodDelete {
			t.Fatal("a 404 preview must not issue a DELETE")
		}
	}
}

func TestWriteToolScopeGating(t *testing.T) {
	ctx := context.Background()

	ro := toolNames(t, ctx, setup(t)) // read-only default
	if ro["create_datacenter"] || ro["update_datacenter"] || ro["delete_datacenter"] {
		t.Error("read-only scope must not expose any write tools")
	}

	w := toolNames(t, ctx, setupWithScope(t, tools.Scope{Write: true}))
	if !w["create_datacenter"] || !w["update_datacenter"] {
		t.Error("write scope must expose create_/update_datacenter")
	}
	if w["delete_datacenter"] {
		t.Error("write scope must NOT expose delete_datacenter (that needs destructive)")
	}

	d := toolNames(t, ctx, setupWithScope(t, tools.Scope{Write: true, Destructive: true}))
	if !d["delete_datacenter"] {
		t.Error("destructive scope must expose delete_datacenter")
	}
}

func TestDeleteDatacenterDynamicParity(t *testing.T) {
	h := setupDynamicWithScope(t, tools.Scope{Write: true, Destructive: true})
	ctx := context.Background()
	h.resp.serve(dcPath+"/dc-1", `{"id":"dc-1","properties":{"name":"dyn","location":"de/txl"}}`)

	// Phase 1: preview through ionos_call_tool.
	res, err := h.session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "ionos_call_tool",
		Arguments: map[string]any{"name": "delete_datacenter", "arguments": map[string]any{"datacenter_id": "dc-1"}},
	})
	if err != nil {
		t.Fatalf("dynamic preview: %v", err)
	}
	if res.IsError {
		t.Fatalf("dynamic preview must not be an error: %s", resultText(res))
	}
	token := extractToken(t, resultText(res))

	// Phase 2: execute through ionos_call_tool, reusing the shared store.
	h.log.clear()
	res, err = h.session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "ionos_call_tool",
		Arguments: map[string]any{"name": "delete_datacenter", "arguments": map[string]any{"datacenter_id": "dc-1", "confirmation_token": token}},
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

func TestDynamicCallToolAnnotationReflectsScope(t *testing.T) {
	ctx := context.Background()
	find := func(h *testSetup) *mcp.Tool {
		for tool, err := range h.session.Tools(ctx, nil) {
			if err != nil {
				t.Fatalf("listing tools: %v", err)
			}
			if tool.Name == "ionos_call_tool" {
				return tool
			}
		}
		t.Fatal("ionos_call_tool not found")
		return nil
	}

	// Read-only scope: the dispatcher is genuinely read-only.
	ro := find(setupDynamic(t))
	if ro.Annotations == nil || !ro.Annotations.ReadOnlyHint {
		t.Errorf("read-only scope: ionos_call_tool ReadOnlyHint = %v, want true", ro.Annotations)
	}

	// Write scope: not read-only, and not destructive (create/update only).
	w := find(setupDynamicWithScope(t, tools.Scope{Write: true}))
	if w.Annotations == nil || w.Annotations.ReadOnlyHint {
		t.Error("write scope: ionos_call_tool must not advertise ReadOnlyHint:true")
	}
	if w.Annotations.DestructiveHint == nil || *w.Annotations.DestructiveHint {
		t.Error("write scope: ionos_call_tool DestructiveHint should be false")
	}

	// Destructive scope: not read-only, and destructive-capable.
	d := find(setupDynamicWithScope(t, tools.Scope{Write: true, Destructive: true}))
	if d.Annotations == nil || d.Annotations.ReadOnlyHint {
		t.Error("destructive scope: ionos_call_tool must not advertise ReadOnlyHint:true")
	}
	if d.Annotations.DestructiveHint == nil || !*d.Annotations.DestructiveHint {
		t.Error("destructive scope: ionos_call_tool DestructiveHint should be true")
	}
}

func TestWriteToolsHiddenInDynamicReadOnly(t *testing.T) {
	// In dynamic read-only mode, write tools must not be callable through the
	// dispatcher (they are neither catalogued nor permitted).
	h := setupDynamic(t)
	ctx := context.Background()
	res, err := h.session.CallTool(ctx, &mcp.CallToolParams{
		Name:      "ionos_call_tool",
		Arguments: map[string]any{"name": "delete_datacenter", "arguments": map[string]any{"datacenter_id": "dc-1"}},
	})
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if !res.IsError {
		t.Fatal("delete_datacenter must not be callable in dynamic read-only mode")
	}
	for _, r := range h.log.allRequests() {
		if r.Method == http.MethodDelete {
			t.Fatal("no DELETE should reach the API in read-only mode")
		}
	}
}
