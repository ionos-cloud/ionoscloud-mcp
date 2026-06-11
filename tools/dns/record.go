package dns

import (
	"context"

	dnsSDK "github.com/ionos-cloud/sdk-go-bundle/products/dns/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterRecordTools(server *mcp.Server, client *dnsSDK.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_dns_records",
		Annotations: tools.ReadOnly,
		Description: "List DNS records across all zones in the account. Use list_dns_zone_records instead when you already know the zone.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		records, _, err := client.RecordsApi.RecordsGet(ctx).Execute()
		return tools.ToResult(records, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_dns_zone_records",
		Annotations: tools.ReadOnly,
		Description: "List all DNS records in a specific zone, with type, content, and TTL. Use list_dns_zones first to find the zone ID; for BIND-format export use get_dns_zone_file.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ZoneIDInput) (*mcp.CallToolResult, any, error) {
		records, _, err := client.RecordsApi.ZonesRecordsGet(ctx, input.ZoneID).Execute()
		return tools.ToResult(records, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_dns_record",
		Annotations: tools.ReadOnly,
		Description: "Get details of a single DNS record: type, content, TTL, and enabled state. Use list_dns_zone_records first to find the record ID.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.RecordIDInput) (*mcp.CallToolResult, any, error) {
		record, _, err := client.RecordsApi.ZonesRecordsFindById(ctx, input.ZoneID, input.RecordID).Execute()
		return tools.ToResult(record, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_dns_secondary_zone_records",
		Annotations: tools.ReadOnly,
		Description: "List all DNS records of a secondary zone — read-only records transferred from the external primary nameserver. Use list_dns_secondary_zones first to find the zone ID.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.SecondaryZoneIDInput) (*mcp.CallToolResult, any, error) {
		records, _, err := client.RecordsApi.SecondaryzonesRecordsGet(ctx, input.SecondaryZoneID).Execute()
		return tools.ToResult(records, err)
	})
}
