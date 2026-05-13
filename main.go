package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/ionos-cloud/ionoscloud-mcp/tools/billing"
	"github.com/ionos-cloud/ionoscloud-mcp/tools/cert"
	"github.com/ionos-cloud/ionoscloud-mcp/tools/dns"
	"github.com/ionos-cloud/ionoscloud-mcp/tools/loader"
	billSDK "github.com/ionos-cloud/sdk-go-bundle/products/billing/v2"
	certSDK "github.com/ionos-cloud/sdk-go-bundle/products/cert/v2"
	computeSDK "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	dnsSDK "github.com/ionos-cloud/sdk-go-bundle/products/dns/v2"
	objstSDK "github.com/ionos-cloud/sdk-go-bundle/products/objectstorage/v2"
	objmgmtSDK "github.com/ionos-cloud/sdk-go-bundle/products/objectstoragemanagement/v2"
	"github.com/ionos-cloud/sdk-go-bundle/shared"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

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

	loader.RegisterComputeLoader(server, client)
	loader.RegisterObjectStorageLoader(server, objstClient, objmgmtClient)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
