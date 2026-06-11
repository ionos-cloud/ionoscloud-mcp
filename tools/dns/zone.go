package dns

import (
	"context"

	dnsSDK "github.com/ionos-cloud/sdk-go-bundle/products/dns/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterZoneTools(server *mcp.Server, client *dnsSDK.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_dns_zones",
		Annotations: tools.ReadOnly,
		Description: "List all DNS zones in your IONOS CLOUD account, with IDs, names, and state. Most DNS tools require a zone ID from this list — call it first. Secondary zones are listed separately by list_dns_secondary_zones.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		zones, _, err := client.ZonesApi.ZonesGet(ctx).Execute()
		return tools.ToResult(zones, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_dns_zone",
		Annotations: tools.ReadOnly,
		Description: "Get details of a single DNS zone: name, description, enabled state, and nameservers. Use list_dns_zones to find zone IDs; use list_dns_zone_records for its records.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ZoneIDInput) (*mcp.CallToolResult, any, error) {
		zone, _, err := client.ZonesApi.ZonesFindById(ctx, input.ZoneID).Execute()
		return tools.ToResult(zone, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_dns_zone_file",
		Annotations: tools.ReadOnly,
		Description: "Export a complete DNS zone as a BIND-format zone file (plain text). Prefer it over list_dns_zone_records when the user wants a portable backup or standard zone-file output. Use list_dns_zones first to find the zone ID.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ZoneIDInput) (*mcp.CallToolResult, any, error) {
		resp, err := client.ZoneFilesApi.ZonesZonefileGet(ctx, input.ZoneID).Execute()
		return tools.ToRawResult(resp, err)
	})
}
