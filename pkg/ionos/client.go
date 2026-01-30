// Package ionos provides the IONOS Cloud client wrapper.
package ionos

import (
	"context"
	"fmt"
	"os"

	dns "github.com/ionos-cloud/sdk-go-bundle/products/dns/v2"
	"github.com/ionos-cloud/sdk-go-bundle/shared"
	ionoscloud "github.com/ionos-cloud/sdk-go/v6"
)

// Client wraps both the compute and DNS API clients for IONOS Cloud.
type Client struct {
	// Compute is the IONOS Cloud compute API client
	Compute *ionoscloud.APIClient
	// DNS is the IONOS Cloud DNS API client
	DNS *dns.APIClient
	// Ctx is the default context for API calls
	Ctx context.Context
}

// NewClient creates a new IONOS Cloud client from environment variables.
// It reads credentials from:
// - IONOS_USERNAME + IONOS_PASSWORD for username/password auth
// - IONOS_TOKEN for token-based auth
func NewClient() (*Client, error) {
	username := os.Getenv("IONOS_USERNAME")
	password := os.Getenv("IONOS_PASSWORD")
	token := os.Getenv("IONOS_TOKEN")

	// Validate that at least one authentication method is provided
	if username == "" && token == "" {
		fmt.Fprintf(os.Stderr, "Warning: No IONOS Cloud credentials found. Set IONOS_USERNAME/IONOS_PASSWORD or IONOS_TOKEN environment variables.\n")
	}

	// Initialize Compute API client (uses sdk-go/v6)
	computeConfig := ionoscloud.NewConfiguration(username, password, token, "")
	computeClient := ionoscloud.NewAPIClient(computeConfig)

	// Initialize DNS API client using shared configuration from sdk-go-bundle
	dnsConfig := shared.NewConfiguration(username, password, token, "")
	dnsClient := dns.NewAPIClient(dnsConfig)

	return &Client{
		Compute: computeClient,
		DNS:     dnsClient,
		Ctx:     context.Background(),
	}, nil
}

// NewClientWithConfig creates a new IONOS Cloud client with explicit credentials.
func NewClientWithConfig(username, password, token string) (*Client, error) {
	// Initialize Compute API client
	computeConfig := ionoscloud.NewConfiguration(username, password, token, "")
	computeClient := ionoscloud.NewAPIClient(computeConfig)

	// Initialize DNS API client
	dnsConfig := shared.NewConfiguration(username, password, token, "")
	dnsClient := dns.NewAPIClient(dnsConfig)

	return &Client{
		Compute: computeClient,
		DNS:     dnsClient,
		Ctx:     context.Background(),
	}, nil
}
