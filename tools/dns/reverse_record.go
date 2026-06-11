package dns

import (
	"context"

	dnsSDK "github.com/ionos-cloud/sdk-go-bundle/products/dns/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterReverseRecordTools(server *mcp.Server, client *dnsSDK.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_dns_reverse_records",
		Annotations: tools.ReadOnly,
		Description: "List all reverse DNS (PTR) records in your IONOS CLOUD account. Reverse records map reserved IONOS IP addresses to hostnames.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		records, _, err := client.ReverseRecordsApi.ReverserecordsGet(ctx).Execute()
		return tools.ToResult(records, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_dns_reverse_record",
		Annotations: tools.ReadOnly,
		Description: "Get details of a single reverse DNS (PTR) record: IP address, hostname, and description. Use list_dns_reverse_records first to find the record ID.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ReverseRecordIDInput) (*mcp.CallToolResult, any, error) {
		record, _, err := client.ReverseRecordsApi.ReverserecordsFindById(ctx, input.ReverseRecordID).Execute()
		return tools.ToResult(record, err)
	})
}
