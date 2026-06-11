package dns

import (
	"context"

	dnsSDK "github.com/ionos-cloud/sdk-go-bundle/products/dns/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterQuotaTools(server *mcp.Server, client *dnsSDK.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_dns_quota",
		Annotations: tools.ReadOnly,
		Description: "Get DNS quota usage versus limits for your account: zones, secondary zones, records, and reverse records consumed vs allowed.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		quota, _, err := client.QuotaApi.QuotaGet(ctx).Execute()
		return tools.ToResult(quota, err)
	})
}
