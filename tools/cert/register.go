package cert

import (
	certSDK "github.com/ionos-cloud/sdk-go-bundle/products/cert/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterAll registers all Certificate Manager tools with the MCP server.
func RegisterAll(server *mcp.Server, client *certSDK.APIClient) {
	RegisterCertificateTools(server, client)
	RegisterAutoCertificateTools(server, client)
	RegisterProviderTools(server, client)
}
