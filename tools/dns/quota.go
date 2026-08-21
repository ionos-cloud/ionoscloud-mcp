package dns

import (
	"context"

	dnsSDK "github.com/ionos-cloud/sdk-go-bundle/products/dns/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterQuotaTools(server *mcp.Server, client *dnsSDK.APIClient, scope tools.Scope) {
	tools.RegisterTool(server, scope, tools.MethodGet, &mcp.Tool{
		Name:        "get_dns_quota",
		Description: "Get DNS quota usage and limits for your IONOS CLOUD account",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		quota, _, err := client.QuotaApi.QuotaGet(ctx).Execute()
		return tools.ToResult(quota, err)
	})
}
