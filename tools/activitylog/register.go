package activitylog

import (
	sdk "github.com/ionos-cloud/sdk-go-bundle/products/activitylog/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterAll registers all Activity Log tools with the MCP server.
func RegisterAll(server *mcp.Server, client *sdk.APIClient) {
	RegisterContractTools(server, client)
	RegisterEventTools(server, client)
}
