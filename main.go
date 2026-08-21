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

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
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
	ctx := context.Background()

	// Tolerant manual arg parsing: this binary is spawned by third-party MCP
	// clients that may pass args we don't recognise, so unknown flags are
	// ignored rather than fatal (stdlib flag.Parse would exit the process).
	var loadModeFlag, transportFlag, httpAddrFlag string
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
		case arg == "--transport":
			if i+1 < len(args) {
				transportFlag = args[i+1]
				i++
			} else {
				log.Println("--transport given without a value; ignoring")
			}
		case strings.HasPrefix(arg, "--transport="):
			transportFlag = strings.TrimPrefix(arg, "--transport=")
		case arg == "--http-addr":
			if i+1 < len(args) {
				httpAddrFlag = args[i+1]
				i++
			} else {
				log.Println("--http-addr given without a value; ignoring")
			}
		case strings.HasPrefix(arg, "--http-addr="):
			httpAddrFlag = strings.TrimPrefix(arg, "--http-addr=")
		default:
			// Unknown arg: ignore for resilience.
		}
	}

	// Resolve the load mode once (flag > env > default) and report the effective
	// mode and its source to stderr, so client-config issues are diagnosable.
	mode, modeSrc := resolveLoadMode(loadModeFlag, os.Getenv("IONOS_MCP_LOAD_MODE"))
	log.Printf("load mode: %s (source: %s)", mode, modeSrc)

	// Resolve the wire transport once (flag > env > default stdio).
	transport, transportSrc := resolveTransport(transportFlag, os.Getenv("IONOS_MCP_TRANSPORT"))
	log.Printf("transport: %s (source: %s)", transport, transportSrc)

	// Resolve the HTTP listen address (flag > env > default). Only used when
	// transport is http.
	httpAddr := strings.TrimSpace(httpAddrFlag)
	if httpAddr == "" {
		httpAddr = strings.TrimSpace(os.Getenv("IONOS_MCP_HTTP_ADDR"))
	}
	if httpAddr == "" {
		httpAddr = "127.0.0.1:8080"
	}

	// Resolve the tool scope once. Read-only by default; write and destructive
	// tools register only when IONOS_MCP_TOOL_SCOPE opts in (hierarchical).
	scope := tools.ParseScope(os.Getenv("IONOS_MCP_TOOL_SCOPE"))
	log.Printf("tool scope: %s", scope)

	// Shared two-phase confirmation store for the create/delete write tools.
	confirm := tools.NewConfirmationStore()

	cfg := shared.NewConfigurationFromEnv()

	ua := ionosclient.New(ionosclient.Options{
		Product:          serverName,
		Version:          serverVersion,
		SDKBundleVersion: sdkBundleVersion(),
		Transport:        string(transport),
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
		{Name: "compute", Summary: "Compute Engine: servers, datacenters, volumes, NICs, LANs, firewall rules, IP blocks, load balancers, NAT gateways, security groups, snapshots.", Register: func(s *mcp.Server) { compute.RegisterAll(s, client, scope, confirm) }},
		{Name: "k8s", Summary: "Managed Kubernetes: clusters, node pools, nodes, versions.", Register: func(s *mcp.Server) { k8s.RegisterAll(s, client, scope, confirm) }},
		{Name: "objectstorage", Summary: "S3-compatible Object Storage: buckets, bucket config (CORS, encryption, lifecycle, policy, replication, versioning), objects, access keys, regions.", Register: func(s *mcp.Server) { objectstorage.RegisterAll(s, objstClient, objmgmtClient, cfg) }},
		{Name: "dns", Summary: "DNS: zones, records, reverse records, secondary zones, DNSSEC, quota, zone-file import.", Register: func(s *mcp.Server) { dns.RegisterAll(s, dnsClient, scope, confirm) }},
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
		d, err := dynamic.Register(ctx, server, products, scope)
		if err != nil {
			log.Fatalf("dynamic load mode: %v", err)
		}
		defer d.Close()
	case LoadModeLazy:
		// Small products register eagerly; Compute and Object Storage defer
		// behind ionos_load_*_tools sentinel tools. Requires MCP client support
		// for notifications/tools/list_changed.
		eagerRegister(server, products, "dns", "billing", "cert", "activitylog", "k8s")
		loader.RegisterComputeLoader(server, client, scope, confirm)
		loader.RegisterObjectStorageLoader(server, objstClient, objmgmtClient, cfg)
	default:
		// Eager: register every tool at startup.
		for _, p := range products {
			p.Register(server)
		}
	}

	switch transport {
	case TransportHTTP:
		handler := mcp.NewStreamableHTTPHandler(
			func(*http.Request) *mcp.Server { return server },
			&mcp.StreamableHTTPOptions{SessionTimeout: httpSessionTimeout},
		)
		// Reject non-safe cross-origin browser requests (CSRF)
		protected := http.NewCrossOriginProtection().Handler(handler)
		srv := &http.Server{
			Addr:              httpAddr,
			Handler:           protected,
			ReadHeaderTimeout: httpReadHeaderTimeout,
		}
		log.Printf("listening on %s (streamable HTTP transport)", httpAddr)
		if err := srv.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	default:
		if err := server.Run(ctx, &mcp.StdioTransport{}); err != nil {
			log.Fatal(err)
		}
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

A Model Context Protocol (MCP) server for IONOS Cloud, served over stdio by
default, or over streamable HTTP with --transport http.

Usage:
  %s [--load-mode <mode>] [--transport <stdio|http>] [--http-addr <addr>]
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

  --transport <mode>   wire transport (overrides IONOS_MCP_TRANSPORT). One of:
                         stdio     (default) serve over stdin/stdout. Required
                                   for clients that spawn the server as a
                                   subprocess (Claude Desktop, Claude Code,
                                   Cursor, Windsurf, etc.).
                         http      serve the Streamable HTTP transport on
                                   --http-addr, for remote/networked clients.

  --http-addr <addr>   listen address for --transport http (overrides
                       IONOS_MCP_HTTP_ADDR). Default "127.0.0.1:8080".

Environment:
  IONOS_MCP_LOAD_MODE          same values as --load-mode (the flag wins if both set).
  IONOS_MCP_TRANSPORT          same values as --transport (the flag wins if both set).
  IONOS_MCP_HTTP_ADDR          same as --http-addr (the flag wins if both set).
  IONOS_MCP_TOOL_SCOPE         write access, off by default. Comma-separated: read (default),
                               write (enables create_/update_ tools), destructive (also enables
                               delete_ tools; implies write). Unrecognised values stay read-only.
  IONOS_TOKEN                  IONOS Cloud API token (required for API calls; all products).
  IONOS_S3_ACCESS_KEY          Object Storage access key (Object Storage tools only).
  IONOS_S3_SECRET_KEY          Object Storage secret key (Object Storage tools only).

Secrets are read from the environment only — never pass tokens as flags.
`, serverName, serverVersion, serverName, serverName, serverName)
}
