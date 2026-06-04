package ionosclient_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	computeSDK "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/ionos-cloud/sdk-go-bundle/shared"

	"github.com/ionos-cloud/ionoscloud-mcp/tools/ionosclient"
)

// TestUserAgentReachesSDKOutboundRequests locks the invariant that powers
// this whole package: when the UA RoundTripper is wired onto cfg.HTTPClient
// and the IONOS SDK constructs an APIClient from that cfg, every outbound
// HTTP request the SDK issues carries the current UA — including segments
// added after construction via SetClient. A future SDK refactor that broke
// the HTTPClient pointer through its shallow-copy path would fail here.
func TestUserAgentReachesSDKOutboundRequests(t *testing.T) {
	var (
		mu       sync.Mutex
		captured []string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		captured = append(captured, r.Header.Get("User-Agent"))
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"items":[]}`))
	}))
	defer srv.Close()

	ua := ionosclient.New(ionosclient.Options{
		Product:          "ionos-cloud-mcp",
		Version:          "1.0.0",
		SDKBundleVersion: "test",
		Transport:        "stdio",
		Mode:             "lazy",
		GOOS:             "linux",
		GOARCH:           "amd64",
	})

	cfg := shared.NewConfiguration("", "", "fake-token", srv.URL)
	cfg.HTTPClient = &http.Client{Transport: ua.Transport(nil)}
	cfg.UserAgent = ua.String()

	client := computeSDK.NewAPIClient(cfg)

	// First outbound call is pre-handshake: only static segments present.
	_, _, _ = client.DataCentersApi.DatacentersGet(context.Background()).Execute()

	ua.SetClient("Claude Code", "1.0.42", "2024-11-05")

	// Second outbound call must reflect the post-handshake mutation.
	_, _, _ = client.DataCentersApi.DatacentersGet(context.Background()).Execute()

	mu.Lock()
	defer mu.Unlock()
	if len(captured) < 2 {
		t.Fatalf("expected at least 2 captured requests, got %d: %v", len(captured), captured)
	}
	if strings.Contains(captured[0], "_host/") {
		t.Errorf("pre-handshake request should not carry host segment: %q", captured[0])
	}
	if !strings.HasPrefix(captured[0], "ionos-cloud-mcp/1.0.0") {
		t.Errorf("pre-handshake request missing static UA prefix: %q", captured[0])
	}
	if !strings.Contains(captured[1], "_host/claude-code") {
		t.Errorf("post-handshake request must carry sanitised host: %q", captured[1])
	}
	if !strings.Contains(captured[1], "_host-version/1.0.42") {
		t.Errorf("post-handshake request must carry host-version: %q", captured[1])
	}
	if !strings.Contains(captured[1], "_protocol/2024-11-05") {
		t.Errorf("post-handshake request must carry protocol: %q", captured[1])
	}
}
