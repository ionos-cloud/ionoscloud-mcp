//go:build e2e

package e2e

import (
	"context"
	"net/http"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestVersionFlag(t *testing.T) {
	out, err := exec.Command(binPath, "--version").Output()
	if err != nil {
		t.Fatalf("--version: %v", err)
	}
	want := "ionos-cloud-mcp " + serverVersion
	if strings.TrimSpace(string(out)) != want {
		t.Errorf("--version = %q, want %q", strings.TrimSpace(string(out)), want)
	}
}

func TestInitialize(t *testing.T) {
	session, _ := spawn(t, nil, nil)
	init := session.InitializeResult()
	if init == nil || init.ServerInfo == nil {
		t.Fatal("no server info from initialize")
	}
	if init.ServerInfo.Name != "ionos-cloud-mcp" {
		t.Errorf("server name = %q, want ionos-cloud-mcp", init.ServerInfo.Name)
	}
	// Proves the ldflags version path (server_config.go) end to end.
	if init.ServerInfo.Version != serverVersion {
		t.Errorf("server version = %q, want %q", init.ServerInfo.Version, serverVersion)
	}
}

func TestEagerToolList(t *testing.T) {
	session, _ := spawn(t, nil, nil)
	names := toolNameSet(t, session)
	for _, want := range []string{"list_servers", "list_dns_zones", "list_object_storage_buckets", "list_k8s_clusters"} {
		if !names[want] {
			t.Errorf("eager tool list missing %q", want)
		}
	}
	for _, sentinel := range []string{"ionos_load_compute_tools", "ionos_load_objectstorage_tools"} {
		if names[sentinel] {
			t.Errorf("eager mode should not expose sentinel %q", sentinel)
		}
	}
}

func TestLazyFlow(t *testing.T) {
	session, notify := spawn(t, map[string]string{"IONOS_MCP_LOAD_MODE": "lazy"}, nil)

	before := toolNameSet(t, session)
	if !before["ionos_load_compute_tools"] {
		t.Fatal("lazy mode should expose ionos_load_compute_tools")
	}
	if before["list_servers"] {
		t.Fatal("lazy mode should not expose list_servers before loading")
	}

	if _, err := session.CallTool(context.Background(), &mcp.CallToolParams{Name: "ionos_load_compute_tools"}); err != nil {
		t.Fatalf("call sentinel: %v", err)
	}

	select {
	case <-notify:
	case <-time.After(5 * time.Second):
		t.Fatal("no tools/list_changed after loading compute tools")
	}

	after := toolNameSet(t, session)
	if !after["list_servers"] {
		t.Error("list_servers not available after lazy load")
	}
}

func TestRouterFallsBackToEager(t *testing.T) {
	var stderr syncBuffer
	session, _ := spawn(t, map[string]string{"IONOS_MCP_LOAD_MODE": "router"}, &stderr)

	names := toolNameSet(t, session)
	if !names["list_servers"] {
		t.Error("router mode should fall back to eager (list_servers expected)")
	}
	if !waitStderrContains(&stderr, "not yet implemented", 5*time.Second) {
		t.Errorf("expected router fallback warning on stderr, got: %q", stderr.String())
	}
}

func TestResourceRead(t *testing.T) {
	session, _ := spawn(t, nil, nil)
	res, err := session.ReadResource(context.Background(), &mcp.ReadResourceParams{
		URI: "ionos://billing/focus-v1.3",
	})
	if err != nil {
		t.Fatalf("ReadResource: %v", err)
	}
	if len(res.Contents) == 0 || !strings.Contains(res.Contents[0].Text, "FOCUS") {
		t.Errorf("FOCUS spec resource not served as expected")
	}
}

func TestToolCallRoundtrip(t *testing.T) {
	clearStatus()
	session, _ := spawn(t, nil, nil)
	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "list_datacenters",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if res.IsError {
		t.Errorf("list_datacenters returned error: %s", textOf(res))
	}
}

func TestUserAgentEndToEnd(t *testing.T) {
	clearStatus()
	session, _ := spawn(t, nil, nil)
	if _, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "list_datacenters",
		Arguments: map[string]any{},
	}); err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	ua := userAgent()
	for _, want := range []string{"ionos-cloud-mcp/" + serverVersion} {
		if !strings.Contains(ua, want) {
			t.Errorf("User-Agent %q missing %q", ua, want)
		}
	}
}

func TestUnauthorized401EndToEnd(t *testing.T) {
	setStatus("/", http.StatusUnauthorized)
	t.Cleanup(clearStatus)

	session, _ := spawn(t, nil, nil)
	res, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "list_dns_zones",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if !res.IsError {
		t.Fatal("expected error result for 401")
	}
	text := textOf(res)
	if !strings.Contains(text, "IONOS_TOKEN") {
		t.Errorf("401 not enriched end-to-end, got: %s", text)
	}
}
