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
		Description: "List all secondary DNS zones in your IONOS CLOUD account",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		zones, _, err := client.SecondaryZonesApi.SecondaryzonesGet(ctx).Execute()
		return tools.ToResult(zones, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_dns_secondary_zone",
		Description: "Get details of a specific secondary DNS zone",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.SecondaryZoneIDInput) (*mcp.CallToolResult, any, error) {
		zone, _, err := client.SecondaryZonesApi.SecondaryzonesFindById(ctx, input.SecondaryZoneID).Execute()
		return tools.ToResult(zone, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_dns_secondary_zone_axfr",
		Description: "Get the zone transfer (AXFR) status of a specific secondary DNS zone",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.SecondaryZoneIDInput) (*mcp.CallToolResult, any, error) {
		status, _, err := client.SecondaryZonesApi.SecondaryzonesAxfrGet(ctx, input.SecondaryZoneID).Execute()
		return tools.ToResult(status, err)
	})
}
