package test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
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
	k8s "github.com/ionos-cloud/ionoscloud-mcp/tools/k8s"
	"github.com/ionos-cloud/ionoscloud-mcp/tools/objectstorage"
)

// recordedRequest stores the HTTP method, path and query of a request.
type recordedRequest struct {
	Method string
	Path   string
	Query  url.Values
	Body   string
}

// requestLog records HTTP requests hitting the test server
type requestLog struct {
	mu       sync.Mutex
	requests []recordedRequest
}

// responder controls the HTTP response the test backend returns. By default it
// answers every request with 200 and an empty JSON object. Tests can override
// the default (setStatus) or pin a specific body/status for one path
// (serve / serveStatus) to exercise output transforms and error handling.
type responder struct {
	mu     sync.Mutex
	status int
	body   string
	byPath map[string]pathResponse
}

type pathResponse struct {
	status int
	body   string
}

func newResponder() *responder {
	return &responder{status: http.StatusOK, body: "{}", byPath: map[string]pathResponse{}}
}

func (rs *responder) get(path string) (int, string) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	if p, ok := rs.byPath[path]; ok {
		return p.status, p.body
	}
	return rs.status, rs.body
}

// setStatus overrides the default status/body returned for any path without a
// specific override.
func (rs *responder) setStatus(status int, body string) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.status = status
	rs.body = body
}

// serve pins a 200 response body for a single path.
func (rs *responder) serve(path, body string) {
	rs.serveStatus(path, http.StatusOK, body)
}

// serveStatus pins a status + body for a single path.
func (rs *responder) serveStatus(path string, status int, body string) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.byPath[path] = pathResponse{status: status, body: body}
}

// reset restores the default 200/{} behaviour and clears per-path overrides.
func (rs *responder) reset() {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.status = http.StatusOK
	rs.body = "{}"
	rs.byPath = map[string]pathResponse{}
}

// testSetup wires an MCP server with a local HTTP backend for testing
type testSetup struct {
	session *mcp.ClientSession
	log     *requestLog
	resp    *responder
}

// toolTest maps a tool name and args to the expected HTTP methods and paths (in order).
// wantQuery and wantContain are optional: nil/empty skips the corresponding check.
type toolTest struct {
	name        string
	args        map[string]any
	wantMethods []string
	wantPaths   []string
	// wantQuery, when non-nil, asserts query params on the request at the same
	// index. Only the keys present in each url.Values are checked.
	wantQuery []url.Values
	// fixture, when set, is served as the 200 body for wantPaths[0].
	fixture string
	// wantContain asserts the tool's text output contains each substring.
	wantContain []string
	// wantBody, when non-empty, asserts the concatenation of all request bodies
	// contains each substring (useful for POST/PUT/PATCH payloads).
	wantBody []string
}

func (r *requestLog) record(req recordedRequest) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.requests = append(r.requests, req)
}

func (r *requestLog) allRequests() []recordedRequest {
	r.mu.Lock()
	defer r.mu.Unlock()

	cp := make([]recordedRequest, len(r.requests))
	copy(cp, r.requests)
	return cp
}

func (r *requestLog) clear() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.requests = r.requests[:0]
}

// resultText concatenates the text content of an MCP tool result.
func resultText(res *mcp.CallToolResult) string {
	if res == nil {
		return ""
	}
	var b strings.Builder
	for _, c := range res.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			b.WriteString(tc.Text)
		}
	}
	return b.String()
}

// run executes a table of tool calls, asserting method/path (always) plus
// query params and output substrings (when specified on the case).
func (h *testSetup) run(t *testing.T, tests []toolTest) {
	t.Helper()
	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h.log.clear()
			h.resp.reset()
			if tt.fixture != "" && len(tt.wantPaths) > 0 {
				h.resp.serve(tt.wantPaths[0], tt.fixture)
			}

			res, err := h.session.CallTool(ctx, &mcp.CallToolParams{
				Name:      tt.name,
				Arguments: tt.args,
			})
			if err != nil {
				t.Fatalf("CallTool(%q) returned protocol error: %v", tt.name, err)
			}
			// Routing-only cases use the default {} backend body, which does not
			// unmarshal into array-returning endpoints; only assert success when
			// the case actually inspects output (fixture / wantContain).
			if (tt.fixture != "" || len(tt.wantContain) > 0) && res != nil && res.IsError {
				t.Fatalf("CallTool(%q) returned tool error: %s", tt.name, resultText(res))
			}

			reqs := h.log.allRequests()
			if len(tt.wantMethods) != len(tt.wantPaths) {
				t.Fatalf("test %q: wantMethods has %d entries, wantPaths has %d", tt.name, len(tt.wantMethods), len(tt.wantPaths))
			}
			if len(reqs) != len(tt.wantPaths) {
				t.Fatalf("CallTool(%q) made %d requests, want %d", tt.name, len(reqs), len(tt.wantPaths))
			}
			for i, req := range reqs {
				if req.Method != tt.wantMethods[i] {
					t.Errorf("CallTool(%q) request[%d] method = %q, want %q", tt.name, i, req.Method, tt.wantMethods[i])
				}
				if req.Path != tt.wantPaths[i] {
					t.Errorf("CallTool(%q) request[%d] path = %q, want %q", tt.name, i, req.Path, tt.wantPaths[i])
				}
				if i < len(tt.wantQuery) && tt.wantQuery[i] != nil {
					for k, want := range tt.wantQuery[i] {
						got := req.Query[k]
						if !equalStrings(got, want) {
							t.Errorf("CallTool(%q) request[%d] query[%q] = %v, want %v", tt.name, i, k, got, want)
						}
					}
				}
			}

			if len(tt.wantContain) > 0 {
				text := resultText(res)
				for _, want := range tt.wantContain {
					if !strings.Contains(text, want) {
						t.Errorf("CallTool(%q) output missing %q\ngot: %s", tt.name, want, text)
					}
				}
			}

			if len(tt.wantBody) > 0 {
				var bodies strings.Builder
				for _, req := range reqs {
					bodies.WriteString(req.Body)
				}
				all := bodies.String()
				for _, want := range tt.wantBody {
					if !strings.Contains(all, want) {
						t.Errorf("CallTool(%q) request body missing %q\ngot: %s", tt.name, want, all)
					}
				}
			}
		})
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func setup(t *testing.T) *testSetup {
	return setupWithScope(t, tools.Scope{})
}

// setupWithScope wires the MCP server with the given tool scope, so write-tool
// tests can enable create/update/delete while the default setup stays read-only.
func setupWithScope(t *testing.T, scope tools.Scope) *testSetup {
	t.Helper()

	log := &requestLog{}
	resp := newResponder()

	// local server replaces the real api.ionos.com
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqBody, _ := io.ReadAll(r.Body)
		log.record(recordedRequest{Method: r.Method, Path: r.URL.Path, Query: r.URL.Query(), Body: string(reqBody)})
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
			Servers: shared.ServerConfigurations{
				{
					URL:         ts.URL,
					Description: "Test",
				},
			},
			OperationServers: map[string]shared.ServerConfigurations{},
		}
	}

	computeClient := ionos.NewAPIClient(testCfg())
	dnsClient := dnsSDK.NewAPIClient(testCfg())
	billingClient := billSDK.NewAPIClient(testCfg())
	certClient := certSDK.NewAPIClient(testCfg())
	objstClient := objstSDK.NewAPIClient(testCfg())
	objmgmtClient := objmgmtSDK.NewAPIClient(testCfg())
	activitylogClient := activitylogSDK.NewAPIClient(testCfg())

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "ionos-cloud-mcp",
		Version: "1.0.0-test",
	}, nil)

	// One store for every product, as main.go does, so a token minted by one
	// product's tool is validated against the same state the others see.
	confirm := tools.NewConfirmationStore()

	activitylog.RegisterAll(server, activitylogClient)
	compute.RegisterAll(server, computeClient, scope, confirm)
	dns.RegisterAll(server, dnsClient)
	billing.RegisterAll(server, billingClient, "")
	cert.RegisterAll(server, certClient)
	k8s.RegisterAll(server, computeClient, scope, confirm)
	objectstorage.RegisterAll(server, objstClient, objmgmtClient, testCfg())

	// in-memory pipe between MCP client and server (replaces stdio)
	ct, st := mcp.NewInMemoryTransports()

	ctx := context.Background()
	ss, err := server.Connect(ctx, st, nil)
	if err != nil {
		t.Fatalf("server.Connect failed: %v", err)
	}
	t.Cleanup(func() { ss.Close() })

	// MCP client used by tests to call tools
	client := mcp.NewClient(&mcp.Implementation{
		Name:    "test-client",
		Version: "1.0.0",
	}, nil)

	session, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatalf("client.Connect failed: %v", err)
	}
	t.Cleanup(func() { session.Close() })

	return &testSetup{
		session: session,
		log:     log,
		resp:    resp,
	}
}

// computeOnlyTools registers ONLY the compute product on a throwaway server and
// returns exactly the tools it registered, at the given scope. Unlike
// setupWithScope (which registers every product), this isolates one product so a
// test can make an exhaustive per-product assertion without name heuristics.
// Registration never calls the API, so no backend is needed.
func computeOnlyTools(t *testing.T, ctx context.Context, scope tools.Scope) []*mcp.Tool {
	t.Helper()

	client := ionos.NewAPIClient(&shared.Configuration{
		Token:              "test-token",
		DefaultHeader:      map[string]string{},
		DefaultQueryParams: make(map[string][]string),
		Servers:            shared.ServerConfigurations{{URL: "http://127.0.0.1:0", Description: "unused"}},
		OperationServers:   map[string]shared.ServerConfigurations{},
	})

	srv := mcp.NewServer(&mcp.Implementation{Name: "compute-only", Version: "test"}, nil)
	compute.RegisterAll(srv, client, scope, tools.NewConfirmationStore())

	ct, st := mcp.NewInMemoryTransports()
	ss, err := srv.Connect(ctx, st, nil)
	if err != nil {
		t.Fatalf("compute-only server.Connect failed: %v", err)
	}
	t.Cleanup(func() { ss.Close() })

	cs, err := mcp.NewClient(&mcp.Implementation{Name: "compute-only-reader", Version: "test"}, nil).Connect(ctx, ct, nil)
	if err != nil {
		t.Fatalf("compute-only client.Connect failed: %v", err)
	}
	t.Cleanup(func() { cs.Close() })

	var out []*mcp.Tool
	for tool, err := range cs.Tools(ctx, nil) {
		if err != nil {
			t.Fatalf("listing compute-only tools: %v", err)
		}
		out = append(out, tool)
	}
	return out
}
