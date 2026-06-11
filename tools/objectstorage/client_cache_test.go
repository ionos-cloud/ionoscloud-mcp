package objectstorage

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	sdk "github.com/ionos-cloud/sdk-go-bundle/products/objectstorage/v2"
	mgmtSDK "github.com/ionos-cloud/sdk-go-bundle/products/objectstoragemanagement/v2"
	"github.com/ionos-cloud/sdk-go-bundle/shared"
)

// cacheBackend is a fake S3 + management API. It answers GetBucketLocation
// (GET /{bucket}) with a configurable region and RegionsFindByRegion
// (GET /regions/{region}) with a configurable endpoint, counting hits.
type cacheBackend struct {
	mu           sync.Mutex
	region       string // LocationConstraint returned for any bucket
	endpoint     string // endpoint returned by the management API
	locationHits int32
	regionHits   int32
}

func (b *cacheBackend) handler(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/regions/") {
		atomic.AddInt32(&b.regionHits, 1)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		b.mu.Lock()
		ep := b.endpoint
		b.mu.Unlock()
		w.Write([]byte(`{"id":"r","type":"region","href":"","metadata":{},"properties":{"version":1,"endpoint":"` + ep + `","website":"","capability":{},"location":"x"}}`))
		return
	}
	// Everything else is treated as a GetBucketLocation request.
	atomic.AddInt32(&b.locationHits, 1)
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	b.mu.Lock()
	region := b.region
	b.mu.Unlock()
	w.Write([]byte("<LocationConstraint>" + region + "</LocationConstraint>"))
}

func newTestCache(t *testing.T, b *cacheBackend) *clientCache {
	t.Helper()
	ts := httptest.NewServer(http.HandlerFunc(b.handler))
	t.Cleanup(ts.Close)

	cfg := &shared.Configuration{
		Servers:    shared.ServerConfigurations{{URL: ts.URL}},
		HTTPClient: http.DefaultClient,
	}
	base := sdk.NewAPIClient(cfg)
	mgmt := mgmtSDK.NewAPIClient(cfg)
	return newClientCache(base, mgmt, cfg)
}

func TestForBucketResolvesAndCaches(t *testing.T) {
	b := &cacheBackend{region: "eu-central-3", endpoint: "https://s3.eu-central-3.ionoscloud.com"}
	c := newTestCache(t, b)
	ctx := context.Background()

	first, err := c.forBucket(ctx, "bucket-a")
	if err != nil {
		t.Fatalf("forBucket: %v", err)
	}
	if first == nil {
		t.Fatal("forBucket returned nil client")
	}
	if got := atomic.LoadInt32(&b.locationHits); got != 1 {
		t.Errorf("location hits = %d, want 1", got)
	}
	if got := atomic.LoadInt32(&b.regionHits); got != 1 {
		t.Errorf("region hits = %d, want 1", got)
	}

	// Second access for the same bucket is served from cache — no new hits.
	second, err := c.forBucket(ctx, "bucket-a")
	if err != nil {
		t.Fatalf("forBucket (cached): %v", err)
	}
	if second != first {
		t.Error("cached call returned a different client instance")
	}
	if got := atomic.LoadInt32(&b.locationHits); got != 1 {
		t.Errorf("location hits after cache = %d, want 1", got)
	}
	if got := atomic.LoadInt32(&b.regionHits); got != 1 {
		t.Errorf("region hits after cache = %d, want 1", got)
	}
}

func TestForBucketEmptyLocationUsesBase(t *testing.T) {
	b := &cacheBackend{region: ""} // empty LocationConstraint
	c := newTestCache(t, b)

	client, err := c.forBucket(context.Background(), "bucket-b")
	if err != nil {
		t.Fatalf("forBucket: %v", err)
	}
	if client != c.base {
		t.Error("empty location should resolve to the base client")
	}
	if got := atomic.LoadInt32(&b.regionHits); got != 0 {
		t.Errorf("region hits = %d, want 0 (no management lookup for empty region)", got)
	}
}

func TestForBucketSameRegionReusesClient(t *testing.T) {
	b := &cacheBackend{region: "de-fra", endpoint: "https://s3.de-fra.ionoscloud.com"}
	c := newTestCache(t, b)
	ctx := context.Background()

	ca, err := c.forBucket(ctx, "bucket-1")
	if err != nil {
		t.Fatalf("forBucket bucket-1: %v", err)
	}
	cb, err := c.forBucket(ctx, "bucket-2")
	if err != nil {
		t.Fatalf("forBucket bucket-2: %v", err)
	}

	if ca != cb {
		t.Error("two buckets in the same region should share one regional client")
	}
	// Two buckets → two location lookups, but only one region resolution.
	if got := atomic.LoadInt32(&b.locationHits); got != 2 {
		t.Errorf("location hits = %d, want 2", got)
	}
	if got := atomic.LoadInt32(&b.regionHits); got != 1 {
		t.Errorf("region hits = %d, want 1 (region resolved once)", got)
	}
}

func TestForBucketLocationErrorNotCached(t *testing.T) {
	// Backend that fails the first location request, then succeeds.
	var calls int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&calls, 1) == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte("<LocationConstraint></LocationConstraint>"))
	}))
	t.Cleanup(ts.Close)

	cfg := &shared.Configuration{Servers: shared.ServerConfigurations{{URL: ts.URL}}, HTTPClient: http.DefaultClient}
	c := newClientCache(sdk.NewAPIClient(cfg), mgmtSDK.NewAPIClient(cfg), cfg)
	ctx := context.Background()

	if _, err := c.forBucket(ctx, "bucket-x"); err == nil {
		t.Fatal("expected error on failed location lookup")
	}
	// The failed bucket must not be cached: a retry hits the backend again.
	if _, err := c.forBucket(ctx, "bucket-x"); err != nil {
		t.Fatalf("retry after error: %v", err)
	}
	if got := atomic.LoadInt32(&calls); got != 2 {
		t.Errorf("backend calls = %d, want 2 (error result not cached)", got)
	}
}

func TestRegionalConfigDerivesEndpoint(t *testing.T) {
	cfg := &shared.Configuration{
		Servers:    shared.ServerConfigurations{{URL: "https://base.example"}},
		HTTPClient: http.DefaultClient,
	}
	c := newClientCache(sdk.NewAPIClient(cfg), mgmtSDK.NewAPIClient(cfg), cfg)

	// regionalConfig sets the endpoint verbatim (scheme normalisation happens in
	// forBucket before this is called) and must preserve the shared HTTPClient —
	// and therefore the User-Agent RoundTripper — through the shallow copy.
	got := c.regionalConfig("https://s3.eu-central-3.ionoscloud.com")
	if len(got.Servers) != 1 || got.Servers[0].URL != "https://s3.eu-central-3.ionoscloud.com" {
		t.Errorf("regionalConfig Servers = %v, want single https endpoint", got.Servers)
	}
	if got.HTTPClient != cfg.HTTPClient {
		t.Error("regionalConfig did not preserve the shared HTTPClient")
	}
	// The derived config must be a distinct value, not the same pointer.
	if got == c.cfg {
		t.Error("regionalConfig returned the shared cfg pointer instead of a copy")
	}
}

// TestForBucketSchemelessEndpoint exercises the scheme-normalisation branch:
// when the management API returns a bare host, forBucket prefixes https://.
func TestForBucketSchemelessEndpoint(t *testing.T) {
	b := &cacheBackend{region: "eu-central-3", endpoint: "s3.eu-central-3.ionoscloud.com"}
	c := newTestCache(t, b)

	client, err := c.forBucket(context.Background(), "bucket-schemeless")
	if err != nil {
		t.Fatalf("forBucket: %v", err)
	}
	if client == nil || client == c.base {
		t.Error("schemeless endpoint should yield a distinct regional client")
	}
}

func TestForBucketConcurrent(t *testing.T) {
	b := &cacheBackend{region: "eu-central-3", endpoint: "https://s3.eu-central-3.ionoscloud.com"}
	c := newTestCache(t, b)
	ctx := context.Background()

	const n = 20
	var wg sync.WaitGroup
	clients := make([]*sdk.APIClient, n)
	for i := range n {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			cl, err := c.forBucket(ctx, "shared-bucket")
			if err != nil {
				t.Errorf("forBucket: %v", err)
				return
			}
			clients[idx] = cl
		}(i)
	}
	wg.Wait()

	for i := 1; i < n; i++ {
		if clients[i] != clients[0] {
			t.Fatalf("concurrent calls returned different clients at %d", i)
		}
	}
	// forBucket holds the lock across the whole resolve, so the same bucket is
	// looked up exactly once even under contention.
	if got := atomic.LoadInt32(&b.locationHits); got != 1 {
		t.Errorf("location hits = %d, want 1 under concurrency", got)
	}
}
