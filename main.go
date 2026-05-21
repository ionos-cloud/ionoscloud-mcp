package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ionos-cloud/ionoscloud-mcp/tools/billing"
	"github.com/ionos-cloud/ionoscloud-mcp/tools/cert"
	"github.com/ionos-cloud/ionoscloud-mcp/tools/compute"
	"github.com/ionos-cloud/ionoscloud-mcp/tools/dns"
	"github.com/ionos-cloud/ionoscloud-mcp/tools/loader"
	"github.com/ionos-cloud/ionoscloud-mcp/tools/objectstorage"
	billSDK "github.com/ionos-cloud/sdk-go-bundle/products/billing/v2"
	certSDK "github.com/ionos-cloud/sdk-go-bundle/products/cert/v2"
	computeSDK "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	dnsSDK "github.com/ionos-cloud/sdk-go-bundle/products/dns/v2"
	objstSDK "github.com/ionos-cloud/sdk-go-bundle/products/objectstorage/v2"
	objmgmtSDK "github.com/ionos-cloud/sdk-go-bundle/products/objectstoragemanagement/v2"
	"github.com/ionos-cloud/sdk-go-bundle/shared"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// eagerLoad reports whether the server should register every tool at startup
// instead of gating Compute and Object Storage behind ionos_load_*_tools.
// Set IONOS_MCP_EAGER_LOAD=true for MCP clients that do not honour
// notifications/tools/list_changed (some Claude Desktop tool-search caches,
// claude.ai connectors, Claude in Chrome, custom agents, etc.). Default off
// preserves the lazy behaviour introduced in PR #17 for clients that do.
func eagerLoad() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("IONOS_MCP_EAGER_LOAD"))) {
	case "1", "true", "yes", "on":
		return true
	}
	return false
}

func main() {
	cfg := shared.NewConfigurationFromEnv()
	cfg.UserAgent = buildUserAgent()

	if cfg.Token == "" {
		fmt.Fprintf(os.Stderr, "Error: IONOS_TOKEN environment variable is required.\n")
		os.Exit(1)
	}

	client := computeSDK.NewAPIClient(cfg)
	dnsClient := dnsSDK.NewAPIClient(cfg)
	billingClient := billSDK.NewAPIClient(cfg)
	certClient := certSDK.NewAPIClient(cfg)
	objmgmtClient := objmgmtSDK.NewAPIClient(cfg)
	objstClient := objstSDK.NewAPIClient(cfg)

	server := mcp.NewServer(&mcp.Implementation{
		Name:    serverName,
		Version: serverVersion,
	}, nil)

	registerResources(server)

	dns.RegisterAll(server, dnsClient)
	billing.RegisterAll(server, billingClient)
	cert.RegisterAll(server, certClient)

	if eagerLoad() {
		// Register every tool up front so it appears in the initial
		// tools/list response. Required for MCP clients that do not
		// refresh their tool catalog on notifications/tools/list_changed.
		compute.RegisterAll(server, client)
		objectstorage.RegisterAll(server, objstClient, objmgmtClient)
	} else {
		loader.RegisterComputeLoader(server, client)
		loader.RegisterObjectStorageLoader(server, objstClient, objmgmtClient)
	}

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
