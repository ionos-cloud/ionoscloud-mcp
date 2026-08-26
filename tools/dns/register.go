package dns

import (
	dnsSDK "github.com/ionos-cloud/sdk-go-bundle/products/dns/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

// RegisterAll registers all DNS tools with the MCP server. Write tools appear only
// when scope allows; confirm backs the two-phase create, delete and import flows.
func RegisterAll(server *mcp.Server, client *dnsSDK.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	RegisterZoneTools(server, client, scope)
	RegisterZoneWriteTools(server, client, scope, confirm)
	RegisterRecordTools(server, client, scope)
	RegisterRecordWriteTools(server, client, scope, confirm)
	RegisterReverseRecordTools(server, client, scope)
	RegisterReverseRecordWriteTools(server, client, scope, confirm)
	RegisterSecondaryZoneTools(server, client, scope)
	RegisterSecondaryZoneWriteTools(server, client, scope, confirm)
	RegisterDNSSECTools(server, client, scope)
	RegisterDNSSECWriteTools(server, client, scope, confirm)
	RegisterQuotaTools(server, client, scope)
}
