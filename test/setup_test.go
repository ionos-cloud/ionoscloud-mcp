package test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/ionos-cloud/ionoscloud-mcp/tools/activitylog"
	"github.com/ionos-cloud/ionoscloud-mcp/tools/billing"
	"github.com/ionos-cloud/ionoscloud-mcp/tools/cert"
	"github.com/ionos-cloud/ionoscloud-mcp/tools/compute"
	"github.com/ionos-cloud/ionoscloud-mcp/tools/dns"
	"github.com/ionos-cloud/ionoscloud-mcp/tools/objectstorage"
	activitylogSDK "github.com/ionos-cloud/sdk-go-bundle/products/activitylog/v2"
	billSDK "github.com/ionos-cloud/sdk-go-bundle/products/billing/v2"
	certSDK "github.com/ionos-cloud/sdk-go-bundle/products/cert/v2"
	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	dnsSDK "github.com/ionos-cloud/sdk-go-bundle/products/dns/v2"
	objstSDK "github.com/ionos-cloud/sdk-go-bundle/products/objectstorage/v2"
	objmgmtSDK "github.com/ionos-cloud/sdk-go-bundle/products/objectstoragemanagement/v2"
	"github.com/ionos-cloud/sdk-go-bundle/shared"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// recordedRequest stores the HTTP method and path of a request.
type recordedRequest struct {
	Method string
	Path   string
}

// requestLog records HTTP requests hitting the test server
type requestLog struct {
	mu       sync.Mutex
	requests []recordedRequest
}

// testSetup wires an MCP server with a local HTTP backend for testing
type testSetup struct {
	session *mcp.ClientSession
	log     *requestLog
}

// toolTest maps a tool name and args to the expected HTTP methods and paths (in order)
type toolTest struct {
	name        string
	args        map[string]any
	wantMethods []string
	wantPaths   []string
}

func (r *requestLog) record(method, path string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.requests = append(r.requests, recordedRequest{Method: method, Path: path})
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

func setup(t *testing.T) *testSetup {
	t.Helper()

	log := &requestLog{}

	// local server replaces the real api.ionos.com
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.record(r.Method, r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	t.Cleanup(ts.Close)

	testCfg := func() *shared.Configuration {
		return &shared.Configuration{
			Token:              "test-token",
			DefaultHeader:      map[string]string{},
			DefaultQueryParams: make(map[string][]string),
			UserAgent:          "ionoscloud-mcp-test",
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
		Name:    "ionoscloud-mcp",
		Version: "1.0.0-test",
	}, nil)

	activitylog.RegisterAll(server, activitylogClient)
	compute.RegisterAll(server, computeClient)
	dns.RegisterAll(server, dnsClient)
	billing.RegisterAll(server, billingClient)
	cert.RegisterAll(server, certClient)
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
	}
}
