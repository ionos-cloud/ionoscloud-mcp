package dns

import (
	"context"

	dnsSDK "github.com/ionos-cloud/sdk-go-bundle/products/dns/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterZoneTools(server *mcp.Server, client *dnsSDK.APIClient, scope tools.Scope) {
	tools.RegisterTool(server, scope, tools.MethodGet, &mcp.Tool{
		Name:        "list_dns_zones",
		Description: "List all DNS zones in your IONOS CLOUD account",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		zones, _, err := client.ZonesApi.ZonesGet(ctx).Execute()
		return tools.ToResult(zones, err)
	})

	tools.RegisterTool(server, scope, tools.MethodGet, &mcp.Tool{
		Name:        "get_dns_zone",
		Description: "Get details of a specific DNS zone",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ZoneIDInput) (*mcp.CallToolResult, any, error) {
		zone, _, err := client.ZonesApi.ZonesFindById(ctx, input.ZoneID).Execute()
		return tools.ToResult(zone, err)
	})

	tools.RegisterTool(server, scope, tools.MethodGet, &mcp.Tool{
		Name:        "get_dns_zone_file",
		Description: "Get the zone file (BIND format) for a specific DNS zone",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ZoneIDInput) (*mcp.CallToolResult, any, error) {
		resp, err := client.ZoneFilesApi.ZonesZonefileGet(ctx, input.ZoneID).Execute()
		return tools.ToRawResult(resp, err)
	})
}
