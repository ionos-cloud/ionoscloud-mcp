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
		Description: "List all DNS records across all zones",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		records, _, err := client.RecordsApi.RecordsGet(ctx).Execute()
		return tools.ToResult(records, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_dns_zone_records",
		Description: "List all DNS records in a specific zone",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ZoneIDInput) (*mcp.CallToolResult, any, error) {
		records, _, err := client.RecordsApi.ZonesRecordsGet(ctx, input.ZoneID).Execute()
		return tools.ToResult(records, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_dns_record",
		Description: "Get details of a specific DNS record in a zone",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.RecordIDInput) (*mcp.CallToolResult, any, error) {
		record, _, err := client.RecordsApi.ZonesRecordsFindById(ctx, input.ZoneID, input.RecordID).Execute()
		return tools.ToResult(record, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_dns_secondary_zone_records",
		Description: "List all DNS records in a specific secondary zone",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.SecondaryZoneIDInput) (*mcp.CallToolResult, any, error) {
		records, _, err := client.RecordsApi.SecondaryzonesRecordsGet(ctx, input.SecondaryZoneID).Execute()
		return tools.ToResult(records, err)
	})
}
