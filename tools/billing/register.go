package billing

import (
	sdk "github.com/ionos-cloud/sdk-go-bundle/products/billing/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterAll registers all billing tools with the MCP server.
func RegisterAll(server *mcp.Server, client *sdk.APIClient) {
	RegisterProfileTools(server, client)
	RegisterEvnTools(server, client)
	RegisterInvoiceTools(server, client)
	RegisterProductTools(server, client)
	RegisterTrafficTools(server, client)
	RegisterUsageTools(server, client)
	RegisterUtilizationTools(server, client)
}
