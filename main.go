package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/ionos-cloud/ionoscloud-mcp/tools/compute"
	"github.com/ionos-cloud/ionoscloud-mcp/tools/dns"
	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	dnsSDK "github.com/ionos-cloud/sdk-go-bundle/products/dns/v2"
	"github.com/ionos-cloud/sdk-go-bundle/shared"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

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

	server := mcp.NewServer(&mcp.Implementation{
		Name:    serverName,
		Version: serverVersion,
	}, nil)

	compute.RegisterAll(server, client)
	dns.RegisterAll(server, dnsClient)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
