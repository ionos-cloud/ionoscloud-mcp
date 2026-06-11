package dns

import (
	"context"

	dnsSDK "github.com/ionos-cloud/sdk-go-bundle/products/dns/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterDNSSECTools(server *mcp.Server, client *dnsSDK.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_dns_zone_dnssec_keys",
		Annotations: tools.ReadOnly,
		Description: "List the DNSSEC keys of a DNS zone, including the DS-record parameters (key tag, algorithm, digest) needed at the domain registrar. Use list_dns_zones first to find the zone ID.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ZoneIDInput) (*mcp.CallToolResult, any, error) {
		keys, _, err := client.DNSSECApi.ZonesKeysGet(ctx, input.ZoneID).Execute()
		return tools.ToResult(keys, err)
	})
}
