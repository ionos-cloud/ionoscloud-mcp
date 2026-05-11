package dns

import (
	"context"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
	dnsSDK "github.com/ionos-cloud/sdk-go-bundle/products/dns/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterReverseRecordTools(server *mcp.Server, client *dnsSDK.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_dns_reverse_records",
		Description: "List all reverse DNS records in your IONOS CLOUD account",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		records, _, err := client.ReverseRecordsApi.ReverserecordsGet(ctx).Execute()
		return tools.ToResult(records, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_dns_reverse_record",
		Description: "Get details of a specific reverse DNS record",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ReverseRecordIDInput) (*mcp.CallToolResult, any, error) {
		record, _, err := client.ReverseRecordsApi.ReverserecordsFindById(ctx, input.ReverseRecordID).Execute()
		return tools.ToResult(record, err)
	})
}
