package main

import (
	"context"
	"fmt"
	"log"
	"os"

	compute "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/ionos-cloud/sdk-go-bundle/shared"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	serverName    = "ionoscloud-mcp"
	serverVersion = "0.1.0"
)

func main() {
	cfg := shared.NewConfigurationFromEnv()

	if cfg.Username == "" && cfg.Token == "" {
		fmt.Fprintf(os.Stderr, "Warning: No IONOS Cloud credentials found. Set IONOS_USERNAME/IONOS_PASSWORD or IONOS_TOKEN environment variables.\n")
	}

	client := compute.NewAPIClient(cfg)

	server := mcp.NewServer(&mcp.Implementation{
		Name:    serverName,
		Version: serverVersion,
	}, nil)

	registerTools(server, client)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
