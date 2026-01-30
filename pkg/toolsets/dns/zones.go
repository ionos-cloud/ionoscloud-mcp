package dns

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ionos-cloud/ionoscloud-mcp/pkg/api"
)

func initZones() []api.ServerTool {
	return []api.ServerTool{
		{
			Tool: api.Tool{
				Name:        "list_dns_zones",
				Description: "List all DNS zones in your IONOS Cloud account",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {},
					"required": []
				}`),
				Annotations: api.ReadOnly("List DNS Zones"),
			},
			Handler: listDnsZones,
		},
		{
			Tool: api.Tool{
				Name:        "get_dns_zone",
				Description: "Get details of a specific DNS zone",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"zone_id": {"type": "string", "description": "The ID of the DNS zone"}
					},
					"required": ["zone_id"]
				}`),
				Annotations: api.ReadOnly("Get DNS Zone"),
			},
			Handler: getDnsZone,
		},
	}
}

func listDnsZones(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	if params.DNSClient == nil {
		return nil, fmt.Errorf("DNS client not configured")
	}

	zones, _, err := params.DNSClient.ZonesApi.ZonesGet(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to list DNS zones: %w", err)
	}
	return api.MarshalResult(zones, "DNS zones")
}

func getDnsZone(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	if params.DNSClient == nil {
		return nil, fmt.Errorf("DNS client not configured")
	}

	zoneID, ok := api.GetRequiredString(params.Arguments, "zone_id")
	if !ok {
		return nil, fmt.Errorf("zone_id is required")
	}

	zone, _, err := params.DNSClient.ZonesApi.ZonesFindById(ctx, zoneID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get DNS zone: %w", err)
	}
	return api.MarshalResult(zone, "DNS zone")
}
