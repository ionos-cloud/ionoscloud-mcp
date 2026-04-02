package dns

import (
	dnsSDK "github.com/ionos-cloud/sdk-go-bundle/products/dns/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterAll registers all DNS tools with the MCP server.
func RegisterAll(server *mcp.Server, client *dnsSDK.APIClient) {
	RegisterZoneTools(server, client)
	RegisterRecordTools(server, client)
	RegisterReverseRecordTools(server, client)
	RegisterSecondaryZoneTools(server, client)
	RegisterDNSSECTools(server, client)
	RegisterQuotaTools(server, client)
}
