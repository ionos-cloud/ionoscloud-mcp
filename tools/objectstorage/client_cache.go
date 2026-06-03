package objectstorage

import (
	"context"
	"fmt"
	"strings"
	"sync"

	sdk "github.com/ionos-cloud/sdk-go-bundle/products/objectstorage/v2"
	mgmtSDK "github.com/ionos-cloud/sdk-go-bundle/products/objectstoragemanagement/v2"
	"github.com/ionos-cloud/sdk-go-bundle/shared"
)

// clientCache maps buckets to regional API clients.
// S3-compatible APIs require requests to target the bucket's regional endpoint;
// the wrong endpoint causes a redirect loop because the signed Authorization
// header is bound to the original host and cannot follow cross-host redirects.
//
// cfg is kept as a live pointer (not a value snapshot) so any mutation made
// during the process lifetime — UA updates after the MCP handshake, future
// token rotation, etc. — flows through to regional clients constructed
// later. Regional clients are still independent SDK clients with their own
// Server URL override; only the User-Agent injection (handled by the
// RoundTripper wired on cfg.HTTPClient.Transport) is shared.
type clientCache struct {
	mu           sync.Mutex
	cfg          *shared.Configuration
	base         *sdk.APIClient
	mgmt         *mgmtSDK.APIClient
	byRegion     map[string]*sdk.APIClient
	bucketRegion map[string]string
}

func newClientCache(base *sdk.APIClient, mgmt *mgmtSDK.APIClient, cfg *shared.Configuration) *clientCache {
	return &clientCache{
		cfg:          cfg,
		base:         base,
		mgmt:         mgmt,
		byRegion:     map[string]*sdk.APIClient{"": base},
		bucketRegion: make(map[string]string),
	}
}

// forBucket returns the regional API client for bucket. On the first access for
// a given bucket it calls GetBucketLocation (always answered by the default
// endpoint without redirecting) to determine the region, resolves the endpoint
// via the management API, then caches both mappings so subsequent calls are
// served from memory.
func (c *clientCache) forBucket(ctx context.Context, bucket string) (*sdk.APIClient, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if region, ok := c.bucketRegion[bucket]; ok {
		return c.byRegion[region], nil
	}

	loc, _, err := c.base.BucketsApi.GetBucketLocation(ctx, bucket).Execute()
	if err != nil {
		return nil, err
	}

	region := ""
	if loc.LocationConstraint != nil && *loc.LocationConstraint != "" {
		region = *loc.LocationConstraint
	}

	if client, ok := c.byRegion[region]; ok {
		c.bucketRegion[bucket] = region // cache only after client is confirmed
		return client, nil
	}

	regionInfo, _, err := c.mgmt.RegionsApi.RegionsFindByRegion(ctx, region).Execute()
	if err != nil {
		return nil, fmt.Errorf("resolving endpoint for region %q: %w", region, err)
	}

	endpoint := regionInfo.Properties.Endpoint
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		endpoint = "https://" + endpoint
	}
	client := sdk.NewAPIClient(c.regionalConfig(endpoint))
	c.byRegion[region] = client
	c.bucketRegion[bucket] = region // cache only after client is confirmed
	return client, nil
}

// regionalConfig derives a configuration for a single regional endpoint by
// dereferencing the live cfg pointer at call time. The result is a fresh
// value the SDK is free to mutate; the shared HTTPClient — and therefore
// the User-Agent RoundTripper — is preserved through the shallow copy.
func (c *clientCache) regionalConfig(endpoint string) *shared.Configuration {
	derived := *c.cfg
	derived.Servers = shared.ServerConfigurations{{URL: endpoint}}
	return &derived
}
