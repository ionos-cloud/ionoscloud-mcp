package dns

import (
	"context"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
	dnsSDK "github.com/ionos-cloud/sdk-go-bundle/products/dns/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterZoneTools(server *mcp.Server, client *dnsSDK.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_dns_zones",
		Description: "List all DNS zones in your IONOS Cloud account",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		zones, _, err := client.ZonesApi.ZonesGet(ctx).Execute()
		return tools.ToResult(zones, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_dns_zone",
		Description: "Get details of a specific DNS zone",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ZoneIDInput) (*mcp.CallToolResult, any, error) {
		zone, _, err := client.ZonesApi.ZonesFindById(ctx, input.ZoneID).Execute()
		return tools.ToResult(zone, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_dns_zone_file",
		Description: "Get the zone file (BIND format) for a specific DNS zone",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ZoneIDInput) (*mcp.CallToolResult, any, error) {
		resp, err := client.ZoneFilesApi.ZonesZonefileGet(ctx, input.ZoneID).Execute()
		return tools.ToRawResult(resp, err)
	})
}
