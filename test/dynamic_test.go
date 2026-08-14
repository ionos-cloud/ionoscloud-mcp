package test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	activitylogSDK "github.com/ionos-cloud/sdk-go-bundle/products/activitylog/v2"
	billSDK "github.com/ionos-cloud/sdk-go-bundle/products/billing/v2"
	certSDK "github.com/ionos-cloud/sdk-go-bundle/products/cert/v2"
	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	dnsSDK "github.com/ionos-cloud/sdk-go-bundle/products/dns/v2"
	objstSDK "github.com/ionos-cloud/sdk-go-bundle/products/objectstorage/v2"
	objmgmtSDK "github.com/ionos-cloud/sdk-go-bundle/products/objectstoragemanagement/v2"
	"github.com/ionos-cloud/sdk-go-bundle/shared"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
	"github.com/ionos-cloud/ionoscloud-mcp/tools/activitylog"
	"github.com/ionos-cloud/ionoscloud-mcp/tools/billing"
	"github.com/ionos-cloud/ionoscloud-mcp/tools/cert"
	"github.com/ionos-cloud/ionoscloud-mcp/tools/compute"
	"github.com/ionos-cloud/ionoscloud-mcp/tools/dns"
	"github.com/ionos-cloud/ionoscloud-mcp/tools/dynamic"
	k8s "github.com/ionos-cloud/ionoscloud-mcp/tools/k8s"
	"github.com/ionos-cloud/ionoscloud-mcp/tools/objectstorage"
)

// setupDynamic wires an MCP server in dynamic load mode (only the three
// meta-tools are public; the full catalog lives on a private in-memory server),
// backed by the same mock HTTP API the eager setup uses.
func setupDynamic(t *testing.T) *testSetup {
	return setupDynamicWithScope(t, tools.Scope{})
}

func setupDynamicWithScope(t *testing.T, scope tools.Scope) *testSetup {
	t.Helper()

	confirm := tools.NewConfirmationStore()

	log := &requestLog{}
	resp := newResponder()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.record(recordedRequest{Method: r.Method, Path: r.URL.Path, Query: r.URL.Query()})
		status, body := resp.get(r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		w.Write([]byte(body))
	}))
	t.Cleanup(ts.Close)

	testCfg := func() *shared.Configuration {
		return &shared.Configuration{
			Token:              "test-token",
			DefaultHeader:      map[string]string{},
			DefaultQueryParams: make(map[string][]string),
			UserAgent:          "ionos-cloud-mcp-test",
			Servers:            shared.ServerConfigurations{{URL: ts.URL, Description: "Test"}},
			OperationServers:   map[string]shared.ServerConfigurations{},
		}
	}

	computeClient := ionos.NewAPIClient(testCfg())
	dnsClient := dnsSDK.NewAPIClient(testCfg())
	billingClient := billSDK.NewAPIClient(testCfg())
	certClient := certSDK.NewAPIClient(testCfg())
	objstClient := objstSDK.NewAPIClient(testCfg())
	objmgmtClient := objmgmtSDK.NewAPIClient(testCfg())
	activitylogClient := activitylogSDK.NewAPIClient(testCfg())

	server := mcp.NewServer(&mcp.Implementation{Name: "ionos-cloud-mcp", Version: "1.0.0-test"}, nil)

	products := []dynamic.Product{
		{Name: "compute", Summary: "Compute Engine.", Register: func(s *mcp.Server) { compute.RegisterAll(s, computeClient, scope, confirm) }},
		{Name: "k8s", Summary: "Managed Kubernetes.", Register: func(s *mcp.Server) { k8s.RegisterAll(s, computeClient, scope, confirm) }},
		{Name: "objectstorage", Summary: "Object Storage.", Register: func(s *mcp.Server) { objectstorage.RegisterAll(s, objstClient, objmgmtClient, testCfg()) }},
		{Name: "dns", Summary: "DNS.", Register: func(s *mcp.Server) { dns.RegisterAll(s, dnsClient) }},
		{Name: "billing", Summary: "Billing.", Register: func(s *mcp.Server) { billing.RegisterAll(s, billingClient, "") }},
		{Name: "cert", Summary: "Certificate Manager.", Register: func(s *mcp.Server) { cert.RegisterAll(s, certClient) }},
		{Name: "activitylog", Summary: "Activity Log.", Register: func(s *mcp.Server) { activitylog.RegisterAll(s, activitylogClient) }},
	}

	ctx := context.Background()
	closer, err := dynamic.Register(ctx, server, products, scope)
	if err != nil {
		t.Fatalf("dynamic.Register failed: %v", err)
	}
	t.Cleanup(func() { closer.Close() })

	ct, st := mcp.NewInMemoryTransports()
	ss, err := server.Connect(ctx, st, nil)
	if err != nil {
		t.Fatalf("server.Connect failed: %v", err)
	}
	t.Cleanup(func() { ss.Close() })

	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "1.0.0"}, nil)
	session, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatalf("client.Connect failed: %v", err)
	}
	t.Cleanup(func() { session.Close() })

	return &testSetup{session: session, log: log, resp: resp}
}

// searchResponse mirrors the ionos_search_tools JSON payload.
type searchResponse struct {
	Count int `json:"count"`
	Tools []struct {
		Name        string `json:"name"`
		Group       string `json:"group"`
		Description string `json:"description"`
	} `json:"tools"`
}

func callSearch(t *testing.T, h *testSetup, args map[string]any) searchResponse {
	t.Helper()
	res, err := h.session.CallTool(context.Background(), &mcp.CallToolParams{Name: "ionos_search_tools", Arguments: args})
	if err != nil {
		t.Fatalf("ionos_search_tools protocol error: %v", err)
	}
	if res.IsError {
		t.Fatalf("ionos_search_tools returned error: %s", resultText(res))
	}
	var out searchResponse
	if err := json.Unmarshal([]byte(resultText(res)), &out); err != nil {
		t.Fatalf("unmarshal search response: %v\ngot: %s", err, resultText(res))
	}
	return out
}

func TestDynamicExposesOnlyMetaTools(t *testing.T) {
	h := setupDynamic(t)

	var names []string
	for tool, err := range h.session.Tools(context.Background(), nil) {
		if err != nil {
			t.Fatalf("listing tools: %v", err)
		}
		names = append(names, tool.Name)
	}
	sort.Strings(names)

	want := []string{"ionos_call_tool", "ionos_describe_tools", "ionos_search_tools"}
	if !equalStrings(names, want) {
		t.Fatalf("public tools = %v, want %v", names, want)
	}
}

func TestDynamicSearchRanksNameMatchFirst(t *testing.T) {
	h := setupDynamic(t)

	// Exact name-token match ranks the listing tool first.
	out := callSearch(t, h, map[string]any{"query": "datacenters"})
	if out.Count == 0 {
		t.Fatal("search for 'datacenters' returned no results")
	}
	if out.Tools[0].Name != "list_datacenters" {
		t.Errorf("top result for 'datacenters' = %q, want list_datacenters", out.Tools[0].Name)
	}

	// A looser keyword still surfaces the relevant compute tools near the top
	// (above tools that merely mention datacenters in their description).
	loose := callSearch(t, h, map[string]any{"query": "datacenter", "limit": 5})
	var found bool
	for _, tool := range loose.Tools {
		if tool.Name == "list_datacenters" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("search for 'datacenter' (limit 5) did not surface list_datacenters; got %+v", loose.Tools)
	}
}

func TestDynamicSearchBrowseByGroup(t *testing.T) {
	h := setupDynamic(t)

	// Empty query + group browses an entire product group.
	dnsOut := callSearch(t, h, map[string]any{"query": "", "group": "dns", "limit": 100})
	if dnsOut.Count == 0 {
		t.Fatal("browsing group 'dns' returned no results")
	}
	for _, tool := range dnsOut.Tools {
		if tool.Group != "dns" {
			t.Errorf("group browse returned %q with group %q, want dns", tool.Name, tool.Group)
		}
	}

	// k8s is behind the catalog too (regression: it was easy to miss).
	k8sOut := callSearch(t, h, map[string]any{"query": "", "group": "k8s", "limit": 100})
	if k8sOut.Count == 0 {
		t.Fatal("browsing group 'k8s' returned no results")
	}
}

func TestDynamicSearchIndexesFullDescription(t *testing.T) {
	h := setupDynamic(t)

	// "idle" appears only in a LATER sentence of list_billing_utilization's
	// description ("set include_zero=true to find idle resources"). Truncating
	// the displayed snippet must not make the tool unfindable — scoring runs
	// over the full description.
	out := callSearch(t, h, map[string]any{"query": "idle", "limit": 20})
	found := false
	var snippet string
	for _, tool := range out.Tools {
		if tool.Name == "list_billing_utilization" {
			found = true
			snippet = tool.Description
		}
	}
	if !found {
		t.Fatalf("search 'idle' did not find list_billing_utilization; got %+v", out.Tools)
	}
	// And the shown snippet should be the matching sentence (contains "idle"),
	// not the generic first sentence.
	if !strings.Contains(strings.ToLower(snippet), "idle") {
		t.Errorf("snippet for 'idle' match should contain the matching sentence; got %q", snippet)
	}
}

func TestDynamicSearchBrowseSnippetIsSingleSentence(t *testing.T) {
	h := setupDynamic(t)

	// Browse billing (no query) — snippets should be first sentences only, i.e.
	// no second sentence in the verbose billing descriptions.
	out := callSearch(t, h, map[string]any{"query": "", "group": "billing", "limit": 100})
	for _, tool := range out.Tools {
		if strings.Count(tool.Description, ". ") > 0 {
			t.Errorf("browse snippet for %q still multi-sentence: %q", tool.Name, tool.Description)
		}
	}
}

func TestDynamicSearchDefaultLimit(t *testing.T) {
	h := setupDynamic(t)

	// A broad query with no explicit limit must cap at the default (10), even
	// though the catalog has 100+ tools.
	out := callSearch(t, h, map[string]any{"query": "list"})
	if out.Count > 10 {
		t.Errorf("default search limit not applied: got %d results, want <= 10", out.Count)
	}
	if out.Count == 0 {
		t.Fatal("broad query 'list' returned nothing")
	}

	// limit=0 means "no limit": the same broad query should now return more than
	// the default cap.
	unlimited := callSearch(t, h, map[string]any{"query": "list", "limit": 0})
	if unlimited.Count <= 10 {
		t.Errorf("limit=0 should be uncapped, got %d results (expected > 10)", unlimited.Count)
	}

	// Negative limit is rejected with an error result.
	res, err := h.session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "ionos_search_tools",
		Arguments: map[string]any{"query": "list", "limit": -5},
	})
	if err != nil {
		t.Fatalf("ionos_search_tools protocol error: %v", err)
	}
	if !res.IsError || !strings.Contains(resultText(res), "must not be negative") {
		t.Errorf("negative limit should error with a clear message; got IsError=%v text=%q", res.IsError, resultText(res))
	}
}

func TestDynamicConcurrentCalls(t *testing.T) {
	h := setupDynamic(t)

	// The catalog client session is shared across meta-tool handlers; fire a
	// burst of concurrent calls to shake out races (run under -race).
	const n = 20
	errs := make(chan error, n)
	for range n {
		go func() {
			res, err := h.session.CallTool(context.Background(), &mcp.CallToolParams{
				Name:      "ionos_call_tool",
				Arguments: map[string]any{"name": "list_datacenters"},
			})
			if err != nil {
				errs <- err
				return
			}
			if res.IsError {
				errs <- errors.New(resultText(res))
				return
			}
			errs <- nil
		}()
	}
	for i := range n {
		if err := <-errs; err != nil {
			t.Errorf("concurrent call %d failed: %v", i, err)
		}
	}
}

func TestDynamicDescribeReturnsSchema(t *testing.T) {
	h := setupDynamic(t)

	res, err := h.session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "ionos_describe_tools",
		Arguments: map[string]any{"names": []string{"get_dns_zone"}},
	})
	if err != nil {
		t.Fatalf("ionos_describe_tools protocol error: %v", err)
	}
	if res.IsError {
		t.Fatalf("ionos_describe_tools returned error: %s", resultText(res))
	}
	text := resultText(res)
	if !strings.Contains(text, "zone_id") {
		t.Errorf("describe output missing 'zone_id'\ngot: %s", text)
	}
}

func TestDynamicCallToolForwardsToBackend(t *testing.T) {
	h := setupDynamic(t)
	h.log.clear()
	h.resp.reset()

	res, err := h.session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "ionos_call_tool",
		Arguments: map[string]any{"name": "list_datacenters"},
	})
	if err != nil {
		t.Fatalf("ionos_call_tool protocol error: %v", err)
	}
	if res.IsError {
		t.Fatalf("ionos_call_tool returned error: %s", resultText(res))
	}

	reqs := h.log.allRequests()
	if len(reqs) != 1 {
		t.Fatalf("made %d backend requests, want 1", len(reqs))
	}
	if reqs[0].Method != "GET" || reqs[0].Path != "/cloudapi/v6/datacenters" {
		t.Errorf("backend request = %s %s, want GET /cloudapi/v6/datacenters", reqs[0].Method, reqs[0].Path)
	}
}

func TestDynamicCallToolForwardsArguments(t *testing.T) {
	h := setupDynamic(t)
	h.log.clear()
	h.resp.reset()

	// Arguments passed through ionos_call_tool must reach the underlying tool:
	// the datacenter_id must appear in the backend path.
	const dc = "dc-abc-123"
	res, err := h.session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "ionos_call_tool",
		Arguments: map[string]any{"name": "get_datacenter", "arguments": map[string]any{"datacenter_id": dc}},
	})
	if err != nil {
		t.Fatalf("ionos_call_tool protocol error: %v", err)
	}
	if res.IsError {
		t.Fatalf("ionos_call_tool returned error: %s", resultText(res))
	}
	reqs := h.log.allRequests()
	if len(reqs) != 1 {
		t.Fatalf("made %d backend requests, want 1", len(reqs))
	}
	if reqs[0].Path != "/cloudapi/v6/datacenters/"+dc {
		t.Errorf("backend path = %q, want /cloudapi/v6/datacenters/%s", reqs[0].Path, dc)
	}
}

func TestDynamicDescribeUnknownSuggests(t *testing.T) {
	h := setupDynamic(t)

	res, err := h.session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "ionos_describe_tools",
		Arguments: map[string]any{"names": []string{"get_datacenter", "list_datacenter"}}, // 2nd is a typo
	})
	if err != nil {
		t.Fatalf("ionos_describe_tools protocol error: %v", err)
	}
	text := resultText(res)
	// Known tool still described, unknown one flagged with a suggestion.
	for _, want := range []string{"datacenter_id", "no such tool", "list_datacenters"} {
		if !strings.Contains(text, want) {
			t.Errorf("describe output missing %q\ngot: %s", want, text)
		}
	}
}

func TestDynamicDescribeEmptyNames(t *testing.T) {
	h := setupDynamic(t)

	// Empty names list is valid input (required field present, just empty) and
	// must return an empty tools array, not an error.
	res, err := h.session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "ionos_describe_tools",
		Arguments: map[string]any{"names": []string{}},
	})
	if err != nil {
		t.Fatalf("ionos_describe_tools protocol error: %v", err)
	}
	if res.IsError {
		t.Errorf("describe with empty names should not error: %s", resultText(res))
	}
}

// TestDynamicCatalogCoversEveryProductTool guards against forgetting a product
// in the dynamic products slice (which would silently make its tools
// unreachable). It builds an eager server with the same products, lists every
// tool, and asserts ionos_describe_tools resolves all of them in dynamic mode.
func TestDynamicCatalogCoversEveryProductTool(t *testing.T) {
	eager := setup(t)      // registers every product eagerly
	dyn := setupDynamic(t) // same products, behind the catalog

	var eagerNames []string
	for tool, err := range eager.session.Tools(context.Background(), nil) {
		if err != nil {
			t.Fatalf("listing eager tools: %v", err)
		}
		eagerNames = append(eagerNames, tool.Name)
	}
	if len(eagerNames) == 0 {
		t.Fatal("eager server exposed no tools")
	}

	res, err := dyn.session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "ionos_describe_tools",
		Arguments: map[string]any{"names": eagerNames},
	})
	if err != nil {
		t.Fatalf("ionos_describe_tools protocol error: %v", err)
	}
	if strings.Contains(resultText(res), "no such tool") {
		t.Errorf("dynamic catalog is missing tools that eager mode exposes:\n%s", resultText(res))
	}
}

func TestDynamicCallToolValidatesArguments(t *testing.T) {
	h := setupDynamic(t)
	h.log.clear()

	// get_datacenter requires datacenter_id; omitting it must fail at the
	// catalog server's schema validation, before any backend call.
	res, err := h.session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "ionos_call_tool",
		Arguments: map[string]any{"name": "get_datacenter", "arguments": map[string]any{}},
	})
	if err != nil {
		t.Fatalf("ionos_call_tool protocol error: %v", err)
	}
	if !res.IsError {
		t.Errorf("calling get_datacenter without datacenter_id should be an error; got: %s", resultText(res))
	}
	if n := len(h.log.allRequests()); n != 0 {
		t.Errorf("invalid call made %d backend requests, want 0", n)
	}
}

func TestDynamicCallToolUnknownNameSuggests(t *testing.T) {
	h := setupDynamic(t)

	res, err := h.session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "ionos_call_tool",
		Arguments: map[string]any{"name": "list_datacenter"}, // typo: missing trailing 's'
	})
	if err != nil {
		t.Fatalf("ionos_call_tool protocol error: %v", err)
	}
	if !res.IsError {
		t.Fatalf("unknown tool name should be an error; got: %s", resultText(res))
	}
	text := resultText(res)
	if !strings.Contains(text, "list_datacenters") {
		t.Errorf("unknown-name error should suggest list_datacenters\ngot: %s", text)
	}
}
