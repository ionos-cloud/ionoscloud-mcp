package dns

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ionos-cloud/ionoscloud-mcp/pkg/api"
)

func initRecords() []api.ServerTool {
	return []api.ServerTool{
		{
			Tool: api.Tool{
				Name:        "list_dns_records",
				Description: "List all DNS records in a zone",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"zone_id": {"type": "string", "description": "The ID of the DNS zone"}
					},
					"required": ["zone_id"]
				}`),
				Annotations: api.ReadOnly("List DNS Records"),
			},
			Handler: listDnsRecords,
		},
		{
			Tool: api.Tool{
				Name:        "get_dns_record",
				Description: "Get details of a specific DNS record",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"zone_id": {"type": "string", "description": "The ID of the DNS zone"},
						"record_id": {"type": "string", "description": "The ID of the DNS record"}
					},
					"required": ["zone_id", "record_id"]
				}`),
				Annotations: api.ReadOnly("Get DNS Record"),
			},
			Handler: getDnsRecord,
		},
	}
}

func listDnsRecords(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	if params.DNSClient == nil {
		return nil, fmt.Errorf("DNS client not configured")
	}

	zoneID, ok := api.GetRequiredString(params.Arguments, "zone_id")
	if !ok {
		return nil, fmt.Errorf("zone_id is required")
	}

	records, _, err := params.DNSClient.RecordsApi.ZonesRecordsGet(ctx, zoneID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to list DNS records: %w", err)
	}
	return api.MarshalResult(records, "DNS records")
}

func getDnsRecord(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	if params.DNSClient == nil {
		return nil, fmt.Errorf("DNS client not configured")
	}

	zoneID, ok := api.GetRequiredString(params.Arguments, "zone_id")
	if !ok {
		return nil, fmt.Errorf("zone_id is required")
	}
	recordID, ok := api.GetRequiredString(params.Arguments, "record_id")
	if !ok {
		return nil, fmt.Errorf("record_id is required")
	}

	record, _, err := params.DNSClient.RecordsApi.ZonesRecordsFindById(ctx, zoneID, recordID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get DNS record: %w", err)
	}
	return api.MarshalResult(record, "DNS record")
}
