package main

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/ionos-cloud/ionoscloud-mcp/tools/loader"
	billSDK "github.com/ionos-cloud/sdk-go-bundle/products/billing/v2"
	computeSDK "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	certSDK "github.com/ionos-cloud/sdk-go-bundle/products/cert/v2"
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
	serverVersion = "1.0.0"
)

func buildUserAgent() string {
	bundleVersion := "unknown"
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, dep := range info.Deps {
			if dep.Path == "github.com/ionos-cloud/sdk-go-bundle/shared" {
				bundleVersion = strings.TrimPrefix(dep.Version, "v")
				break
			}
		}
	}

	return fmt.Sprintf("ionoscloud-mcp/%s_ionos-cloud-sdk-go-bundle/%s_os/%s_arch/%s",
		serverVersion, bundleVersion, runtime.GOOS, runtime.GOARCH)
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

	eagerDomains := loader.ParseDomains(os.Getenv("IONOS_DOMAINS"), loader.EagerDomains)

	dl := loader.NewDomainLoader(server, loader.DomainClients{
		Compute: client,
		DNS:     dnsClient,
		Billing: billingClient,
		Cert:    certClient,
		ObjSt:   objstClient,
		ObjMgmt: objmgmtClient,
	})

	for _, d := range eagerDomains {
		if _, err := dl.Load(d); err != nil {
			log.Fatalf("failed to load domain %q: %v", d, err)
		}
	}

	loader.RegisterLoaderTools(server, dl)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
