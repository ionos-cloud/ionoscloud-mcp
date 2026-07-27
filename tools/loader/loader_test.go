package loader

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	computeSDK "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	objstSDK "github.com/ionos-cloud/sdk-go-bundle/products/objectstorage/v2"
	objmgmtSDK "github.com/ionos-cloud/sdk-go-bundle/products/objectstoragemanagement/v2"
	"github.com/ionos-cloud/sdk-go-bundle/shared"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

// loaderHarness wires an in-memory MCP server+client. notify receives a value
// each time the server emits notifications/tools/list_changed.
type loaderHarness struct {
	server  *mcp.Server
	session *mcp.ClientSession
	notify  chan struct{}
}

func testConfig() *shared.Configuration {
	return &shared.Configuration{
		Servers: shared.ServerConfigurations{{URL: "http://127.0.0.1:0"}},
	}
}

func newHarness(t *testing.T, register func(server *mcp.Server)) *loaderHarness {
	t.Helper()

	server := mcp.NewServer(&mcp.Implementation{Name: "test-server", Version: "0"}, nil)
	register(server)

	notify := make(chan struct{}, 16)
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0"}, &mcp.ClientOptions{
		ToolListChangedHandler: func(_ context.Context, _ *mcp.ToolListChangedRequest) {
			select {
			case notify <- struct{}{}:
			default:
			}
		},
	})

	ct, st := mcp.NewInMemoryTransports()
	ctx := context.Background()
	ss, err := server.Connect(ctx, st, nil)
	if err != nil {
		t.Fatalf("server.Connect: %v", err)
	}
	t.Cleanup(func() { ss.Close() })

	session, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatalf("client.Connect: %v", err)
	}
	t.Cleanup(func() { session.Close() })

	return &loaderHarness{server: server, session: session, notify: notify}
}

func (h *loaderHarness) toolNames(t *testing.T) map[string]bool {
	t.Helper()
	res, err := h.session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}
	names := make(map[string]bool, len(res.Tools))
	for _, tool := range res.Tools {
		names[tool.Name] = true
	}
	return names
}

func (h *loaderHarness) call(t *testing.T, name string) string {
	t.Helper()
	res, err := h.session.CallTool(context.Background(), &mcp.CallToolParams{Name: name})
	if err != nil {
		t.Fatalf("CallTool(%q): %v", name, err)
	}
	var out strings.Builder
	for _, c := range res.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			out.WriteString(tc.Text)
		}
	}
	return out.String()
}

func (h *loaderHarness) waitNotify(t *testing.T) {
	t.Helper()
	select {
	case <-h.notify:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for tools/list_changed notification")
	}
}

func TestComputeLoaderRegistersOnFirstCall(t *testing.T) {
	client := computeSDK.NewAPIClient(testConfig())
	h := newHarness(t, func(server *mcp.Server) {
		RegisterComputeLoader(server, client, tools.Scope{}, tools.NewConfirmationStore())
	})

	before := h.toolNames(t)
	if !before["ionos_load_compute_tools"] {
		t.Fatal("sentinel tool ionos_load_compute_tools not registered")
	}
	if before["list_servers"] {
		t.Fatal("compute tools should not be registered before the sentinel is called")
	}

	out := h.call(t, "ionos_load_compute_tools")
	if out == "" {
		t.Fatal("loader returned empty result")
	}
	h.waitNotify(t)

	after := h.toolNames(t)
	for _, want := range []string{"list_servers", "list_datacenters"} {
		if !after[want] {
			t.Errorf("compute tool %q not registered after loader call", want)
		}
	}
}

func TestComputeLoaderIdempotent(t *testing.T) {
	client := computeSDK.NewAPIClient(testConfig())
	h := newHarness(t, func(server *mcp.Server) {
		RegisterComputeLoader(server, client, tools.Scope{}, tools.NewConfirmationStore())
	})

	first := h.call(t, "ionos_load_compute_tools")
	second := h.call(t, "ionos_load_compute_tools")

	if first == second {
		t.Errorf("expected distinct messages for first vs repeat load, both: %q", first)
	}
	if want := "already loaded"; !strings.Contains(second, want) {
		t.Errorf("second call = %q, want it to contain %q", second, want)
	}
}

func TestObjectStorageLoaderRegistersOnFirstCall(t *testing.T) {
	cfg := testConfig()
	base := objstSDK.NewAPIClient(cfg)
	mgmt := objmgmtSDK.NewAPIClient(cfg)
	h := newHarness(t, func(server *mcp.Server) {
		RegisterObjectStorageLoader(server, base, mgmt, cfg)
	})

	before := h.toolNames(t)
	if !before["ionos_load_objectstorage_tools"] {
		t.Fatal("sentinel tool ionos_load_objectstorage_tools not registered")
	}
	if before["list_object_storage_buckets"] {
		t.Fatal("object storage tools should not be registered before the sentinel is called")
	}

	h.call(t, "ionos_load_objectstorage_tools")
	h.waitNotify(t)

	after := h.toolNames(t)
	if !after["list_object_storage_buckets"] {
		t.Error("list_object_storage_buckets not registered after loader call")
	}
}

func TestComputeLoaderConcurrent(t *testing.T) {
	client := computeSDK.NewAPIClient(testConfig())
	h := newHarness(t, func(server *mcp.Server) {
		RegisterComputeLoader(server, client, tools.Scope{}, tools.NewConfirmationStore())
	})

	const n = 10
	results := make([]string, n)
	var wg sync.WaitGroup
	for i := range n {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx] = h.call(t, "ionos_load_compute_tools")
		}(i)
	}
	wg.Wait()

	loaded := 0
	for _, r := range results {
		if strings.Contains(r, "already loaded") {
			continue
		}
		loaded++
	}
	if loaded != 1 {
		t.Errorf("expected exactly 1 registration across %d concurrent calls, got %d", n, loaded)
	}
}
