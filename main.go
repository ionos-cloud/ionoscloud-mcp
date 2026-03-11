package main

import (
	"context"
	"fmt"
	"log"
	"os"

	ionoscloud "github.com/ionos-cloud/sdk-go/v6"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	serverName    = "ionoscloud-mcp"
	serverVersion = "0.1.0"
)

func main() {
	username := os.Getenv("IONOS_USERNAME")
	password := os.Getenv("IONOS_PASSWORD")
	token := os.Getenv("IONOS_TOKEN")

	if username == "" && token == "" {
		fmt.Fprintf(os.Stderr, "Warning: No IONOS Cloud credentials found. Set IONOS_USERNAME/IONOS_PASSWORD or IONOS_TOKEN environment variables.\n")
	}

	configuration := ionoscloud.NewConfiguration(username, password, token, "")
	client := ionoscloud.NewAPIClient(configuration)

	server := mcp.NewServer(&mcp.Implementation{
		Name:    serverName,
		Version: serverVersion,
	}, nil)

	registerTools(server, client)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
