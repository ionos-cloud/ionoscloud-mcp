package dns

import (
	"context"

	dnsSDK "github.com/ionos-cloud/sdk-go-bundle/products/dns/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterSecondaryZoneTools(server *mcp.Server, client *dnsSDK.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_dns_secondary_zones",
		Annotations: tools.ReadOnly,
		Description: "List all secondary DNS zones — zones whose records are transferred from external primary nameservers via AXFR. For zones managed directly in IONOS use list_dns_zones instead.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		zones, _, err := client.SecondaryZonesApi.SecondaryzonesGet(ctx).Execute()
		return tools.ToResult(zones, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_dns_secondary_zone",
		Annotations: tools.ReadOnly,
		Description: "Get details of a single secondary DNS zone: primary nameserver IPs and zone state. Use list_dns_secondary_zones to find zone IDs; use get_dns_secondary_zone_axfr for its transfer status.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.SecondaryZoneIDInput) (*mcp.CallToolResult, any, error) {
		zone, _, err := client.SecondaryZonesApi.SecondaryzonesFindById(ctx, input.SecondaryZoneID).Execute()
		return tools.ToResult(zone, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_dns_secondary_zone_axfr",
		Annotations: tools.ReadOnly,
		Description: "Get the zone-transfer (AXFR) status of a secondary DNS zone per primary nameserver IP — when the last transfer ran and whether it succeeded. Use it to diagnose stale secondary-zone data.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.SecondaryZoneIDInput) (*mcp.CallToolResult, any, error) {
		status, _, err := client.SecondaryZonesApi.SecondaryzonesAxfrGet(ctx, input.SecondaryZoneID).Execute()
		return tools.ToResult(status, err)
	})
}
