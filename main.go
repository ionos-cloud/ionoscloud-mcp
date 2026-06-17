package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"

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
	"github.com/ionos-cloud/ionoscloud-mcp/tools/dynamic"
	"github.com/ionos-cloud/ionoscloud-mcp/tools/ionosclient"
	k8s "github.com/ionos-cloud/ionoscloud-mcp/tools/k8s"
	"github.com/ionos-cloud/ionoscloud-mcp/tools/loader"
	"github.com/ionos-cloud/ionoscloud-mcp/tools/objectstorage"
)

func main() {
	const transport = "stdio"
	ctx := context.Background()

	// Tolerant manual arg parsing: this binary is spawned by third-party MCP
	// clients that may pass args we don't recognise, so unknown flags are
	// ignored rather than fatal (stdlib flag.Parse would exit the process).
	var loadModeFlag string
	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--version":
			fmt.Printf("%s %s\n", serverName, serverVersion)
			return
		case arg == "--help" || arg == "-h":
			printHelp()
			return
		case arg == "--load-mode":
			if i+1 < len(args) {
				loadModeFlag = args[i+1]
				i++
			} else {
				log.Println("--load-mode given without a value; ignoring")
			}
		case strings.HasPrefix(arg, "--load-mode="):
			loadModeFlag = strings.TrimPrefix(arg, "--load-mode=")
		default:
			// Unknown arg: ignore for resilience.
		}
	}

	// Resolve the load mode once (flag > env > default) and report the effective
	// mode and its source to stderr, so client-config issues are diagnosable.
	mode, modeSrc := resolveLoadMode(loadModeFlag, os.Getenv("IONOS_MCP_LOAD_MODE"))
	log.Printf("load mode: %s (source: %s)", mode, modeSrc)

	cfg := shared.NewConfigurationFromEnv()

	ua := ionosclient.New(ionosclient.Options{
		Product:          serverName,
		Version:          serverVersion,
		SDKBundleVersion: sdkBundleVersion(),
		Transport:        transport,
		Mode:             string(mode),
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

	// Single source of truth for the product tool groups. Each Register closure
	// is the same call the eager path makes; lazy and dynamic modes reuse these
	// so the three modes can never drift. Order here is the catalog listing order
	// in dynamic mode.
	products := []dynamic.Product{
		{Name: "compute", Summary: "Compute Engine: servers, datacenters, volumes, NICs, LANs, firewall rules, IP blocks, load balancers, NAT gateways, security groups, snapshots.", Register: func(s *mcp.Server) { compute.RegisterAll(s, client) }},
		{Name: "k8s", Summary: "Managed Kubernetes: clusters, node pools, nodes, versions.", Register: func(s *mcp.Server) { k8s.RegisterAll(s, client) }},
		{Name: "objectstorage", Summary: "S3-compatible Object Storage: buckets, bucket config (CORS, encryption, lifecycle, policy, replication, versioning), objects, access keys, regions.", Register: func(s *mcp.Server) { objectstorage.RegisterAll(s, objstClient, objmgmtClient, cfg) }},
		{Name: "dns", Summary: "DNS: zones, records, reverse records, secondary zones, DNSSEC, quota.", Register: func(s *mcp.Server) { dns.RegisterAll(s, dnsClient) }},
		{Name: "billing", Summary: "Billing: invoices, usage, utilization, traffic, EVN, FOCUS spec.", Register: func(s *mcp.Server) { billing.RegisterAll(s, billingClient, focusSpec) }},
		{Name: "cert", Summary: "Certificate Manager: certificates, auto-certificates, providers.", Register: func(s *mcp.Server) { cert.RegisterAll(s, certClient) }},
		{Name: "activitylog", Summary: "Activity Log: contracts, events.", Register: func(s *mcp.Server) { activitylog.RegisterAll(s, activitylogClient) }},
	}

	switch mode {
	case LoadModeDynamic:
		// Expose only the search/describe/call meta-tools; the full catalog
		// lives on a private in-memory server. Best for clients with hard tool
		// caps and no tool search of their own (Cursor, Windsurf).
		// The returned closer tears down the private catalog connection on exit.
		// The catalog is a process-lifetime singleton, so this only matters for a
		// clean shutdown (server.Run returning); the OS would reclaim it anyway.
		d, err := dynamic.Register(ctx, server, products)
		if err != nil {
			log.Fatalf("dynamic load mode: %v", err)
		}
		defer d.Close()
	case LoadModeLazy:
		// Small products register eagerly; Compute and Object Storage defer
		// behind ionos_load_*_tools sentinel tools. Requires MCP client support
		// for notifications/tools/list_changed.
		eagerRegister(server, products, "dns", "billing", "cert", "activitylog", "k8s")
		loader.RegisterComputeLoader(server, client)
		loader.RegisterObjectStorageLoader(server, objstClient, objmgmtClient, cfg)
	default:
		// Eager: register every tool at startup.
		for _, p := range products {
			p.Register(server)
		}
	}

	if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}

// eagerRegister runs the Register closures of the named products on server,
// preserving the products slice order. Used by lazy mode to register the
// small product groups while Compute and Object Storage defer behind loaders.
func eagerRegister(server *mcp.Server, products []dynamic.Product, names ...string) {
	want := make(map[string]bool, len(names))
	for _, n := range names {
		want[n] = true
	}
	for _, p := range products {
		if want[p.Name] {
			p.Register(server)
		}
	}
}

// printHelp writes usage, flags, environment variables, and per-mode guidance.
func printHelp() {
	fmt.Printf(`%s %s

A Model Context Protocol (MCP) server for IONOS Cloud, served over stdio.

Usage:
  %s [--load-mode <mode>]   run the MCP server over stdio
  %s --version              print version and exit
  %s --help                 print this help and exit

Flags:
  --load-mode <mode>   tool registration strategy (overrides IONOS_MCP_LOAD_MODE).
                       One of:
                         eager     (default) register all tools at startup.
                                   Best for Claude Code, which defers tool
                                   schemas client-side via its own tool search.
                         lazy      register small products eagerly; defer Compute
                                   and Object Storage behind loader tools (needs
                                   client support for tools/list_changed).
                         dynamic   expose only 3 meta-tools (search / describe /
                                   call) and browse the full catalog through them.
                                   For clients with hard tool caps and no tool
                                   search of their own (e.g. Cursor, Windsurf).
                                   Alias: search.

Environment:
  IONOS_MCP_LOAD_MODE          same values as --load-mode (the flag wins if both set).
  IONOS_TOKEN                  IONOS Cloud API token (required for API calls; all products).
  IONOS_S3_ACCESS_KEY          Object Storage access key (Object Storage tools only).
  IONOS_S3_SECRET_KEY          Object Storage secret key (Object Storage tools only).

Secrets are read from the environment only — never pass tokens as flags.
`, serverName, serverVersion, serverName, serverName, serverName)
}
