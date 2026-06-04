package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"

	activitylogSDK "github.com/ionos-cloud/sdk-go-bundle/products/activitylog/v2"
	billSDK "github.com/ionos-cloud/sdk-go-bundle/products/billing/v2"
	certSDK "github.com/ionos-cloud/sdk-go-bundle/products/cert/v2"
	computeSDK "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	dnsSDK "github.com/ionos-cloud/sdk-go-bundle/products/dns/v2"
	objstSDK "github.com/ionos-cloud/sdk-go-bundle/products/objectstorage/v2"
	objmgmtSDK "github.com/ionos-cloud/sdk-go-bundle/products/objectstoragemanagement/v2"
	"github.com/ionos-cloud/sdk-go-bundle/shared"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools/activitylog"
	"github.com/ionos-cloud/ionoscloud-mcp/tools/billing"
	"github.com/ionos-cloud/ionoscloud-mcp/tools/cert"
	"github.com/ionos-cloud/ionoscloud-mcp/tools/compute"
	"github.com/ionos-cloud/ionoscloud-mcp/tools/dns"
	"github.com/ionos-cloud/ionoscloud-mcp/tools/ionosclient"
	"github.com/ionos-cloud/ionoscloud-mcp/tools/loader"
	"github.com/ionos-cloud/ionoscloud-mcp/tools/objectstorage"
)

func main() {
	const transport = "stdio"

	for _, arg := range os.Args[1:] {
		switch arg {
		case "--version", "-v", "version":
			fmt.Printf("%s %s\n", serverName, serverVersion)
			return
		case "--help", "-h", "help":
			fmt.Printf("%s %s\nUsage: %s        run MCP server over stdio (requires IONOS_TOKEN)\n       %s --version\n",
				serverName, serverVersion, serverName, serverName)
			return
		}
	}

	cfg := shared.NewConfigurationFromEnv()
	if cfg.Token == "" {
		fmt.Fprintf(os.Stderr, "Error: IONOS_TOKEN environment variable is required.\n")
		os.Exit(1)
	}

	ua := ionosclient.New(ionosclient.Options{
		Product:          serverName,
		Version:          serverVersion,
		SDKBundleVersion: sdkBundleVersion(),
		Transport:        transport,
		Mode:             loadModeLabel(),
		GOOS:             runtime.GOOS,
		GOARCH:           runtime.GOARCH,
	})
	// Install a single RoundTripper on a *fresh* http.Client. Every SDK
	// client (compute, DNS, billing, cert, object storage base + regional)
	// shares this HTTPClient pointer through shallow cfg copies, so one
	// wrap covers every outbound request without chasing cfg snapshots.
	//
	// shared.NewConfigurationFromEnv hands back cfg.HTTPClient =
	// http.DefaultClient with a custom Transport already installed; if we
	// mutated that Transport in place we would poison http.DefaultClient
	// for the rest of the process. Instead we keep the SDK's transport as
	// our base and wrap it inside a dedicated *http.Client we own.
	var base http.RoundTripper
	if cfg.HTTPClient != nil {
		base = cfg.HTTPClient.Transport
	}
	cfg.HTTPClient = &http.Client{Transport: ua.Transport(base)}
	cfg.UserAgent = ua.String()

	client := computeSDK.NewAPIClient(cfg)
	dnsClient := dnsSDK.NewAPIClient(cfg)
	billingClient := billSDK.NewAPIClient(cfg)
	certClient := certSDK.NewAPIClient(cfg)
	objmgmtClient := objmgmtSDK.NewAPIClient(cfg)
	objstClient := objstSDK.NewAPIClient(cfg)
	activitylogClient := activitylogSDK.NewAPIClient(cfg)

	server := mcp.NewServer(&mcp.Implementation{
		Name:    serverName,
		Version: serverVersion,
	}, &mcp.ServerOptions{
		InitializedHandler: func(_ context.Context, req *mcp.InitializedRequest) {
			if req == nil || req.Session == nil {
				return
			}
			params := req.Session.InitializeParams()
			if params == nil {
				return
			}
			var name, version string
			if params.ClientInfo != nil {
				name = params.ClientInfo.Name
				version = params.ClientInfo.Version
			}
			ua.SetClient(name, version, params.ProtocolVersion)
		},
	})

	registerResources(server)

	activitylog.RegisterAll(server, activitylogClient)
	dns.RegisterAll(server, dnsClient)
	billing.RegisterAll(server, billingClient)
	cert.RegisterAll(server, certClient)

	if eagerLoad() {
		// Register every tool up front so it appears in the initial
		// tools/list response. Required for MCP clients that do not
		// refresh their tool catalog on notifications/tools/list_changed.
		compute.RegisterAll(server, client)
		objectstorage.RegisterAll(server, objstClient, objmgmtClient, cfg)
	} else {
		loader.RegisterComputeLoader(server, client)
		loader.RegisterObjectStorageLoader(server, objstClient, objmgmtClient, cfg)
	}

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
