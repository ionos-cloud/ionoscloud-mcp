package dns

import (
	"context"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
	dnsSDK "github.com/ionos-cloud/sdk-go-bundle/products/dns/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterDNSSECTools(server *mcp.Server, client *dnsSDK.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_dns_zone_dnssec_keys",
		Description: "List DNSSEC keys for a specific DNS zone",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ZoneIDInput) (*mcp.CallToolResult, any, error) {
		keys, _, err := client.DNSSECApi.ZonesKeysGet(ctx, input.ZoneID).Execute()
		return tools.ToResult(keys, err)
	})
}
