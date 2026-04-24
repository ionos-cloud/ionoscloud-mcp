package main

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"os"

	"github.com/ionos-cloud/ionoscloud-mcp/tools/billing"
	"github.com/ionos-cloud/ionoscloud-mcp/tools/compute"
	"github.com/ionos-cloud/ionoscloud-mcp/tools/dns"
	"github.com/ionos-cloud/ionoscloud-mcp/tools/objectstorage"
	billSDK "github.com/ionos-cloud/sdk-go-bundle/products/billing/v2"
	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	dnsSDK "github.com/ionos-cloud/sdk-go-bundle/products/dns/v2"
	objstSDK "github.com/ionos-cloud/sdk-go-bundle/products/objectstorage/v2"
	objmgmtSDK "github.com/ionos-cloud/sdk-go-bundle/products/objectstoragemanagement/v2"
	"github.com/ionos-cloud/sdk-go-bundle/shared"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed docs/billing/focus-v1.3.md
var focusSpec string

const (
	serverName    = "ionoscloud-mcp"
	serverVersion = "0.1.0"
)

func main() {
	cfg := shared.NewConfigurationFromEnv()

	if cfg.Token == "" {
		fmt.Fprintf(os.Stderr, "Error: IONOS_TOKEN environment variable is required.\n")
		os.Exit(1)
	}

	client := ionos.NewAPIClient(cfg)
	dnsClient := dnsSDK.NewAPIClient(cfg)
	billingClient := billSDK.NewAPIClient(cfg)
	objmgmtClient := objmgmtSDK.NewAPIClient(cfg)
	objstClient := objstSDK.NewAPIClient(cfg)

	server := mcp.NewServer(&mcp.Implementation{
		Name:    serverName,
		Version: serverVersion,
	}, nil)

	server.AddResource(&mcp.Resource{
		URI:         "ionos://billing/focus-v1.3", // not a real URI, just an identifier for the resource
		Name:        "focus-v1.3",
		Title:       "FOCUS v1.3 Billing Spec",
		Description: "FOCUS v1.3 column names, allowed values, and IONOS tool → FOCUS field mappings. Read this when producing standards-compliant billing output.",
		MIMEType:    "text/markdown",
	}, func(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{{
				URI:      req.Params.URI,
				Text:     focusSpec,
				MIMEType: "text/markdown",
			}},
		}, nil
	})

	compute.RegisterAll(server, client)
	dns.RegisterAll(server, dnsClient)
	billing.RegisterAll(server, billingClient)
	objectstorage.RegisterAll(server, objstClient, objmgmtClient)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
