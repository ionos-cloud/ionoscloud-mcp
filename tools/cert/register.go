package cert

import (
	certSDK "github.com/ionos-cloud/sdk-go-bundle/products/cert/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

// RegisterAll registers all Certificate Manager tools with the MCP server. Write
// tools appear only when scope allows; confirm backs the two-phase create/delete.
func RegisterAll(server *mcp.Server, client *certSDK.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	RegisterCertificateTools(server, client, scope)
	RegisterCertificateWriteTools(server, client, scope, confirm)
	RegisterAutoCertificateTools(server, client, scope)
	RegisterAutoCertificateWriteTools(server, client, scope, confirm)
	RegisterProviderTools(server, client, scope)
	RegisterProviderWriteTools(server, client, scope, confirm)
}
