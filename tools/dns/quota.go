package dns

import (
	"context"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
	dnsSDK "github.com/ionos-cloud/sdk-go-bundle/products/dns/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterQuotaTools(server *mcp.Server, client *dnsSDK.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_dns_quota",
		Description: "Get DNS quota usage and limits for your IONOS CLOUD account",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		quota, _, err := client.QuotaApi.QuotaGet(ctx).Execute()
		return tools.ToResult(quota, err)
	})
}
