package billing

import (
	"context"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
	sdk "github.com/ionos-cloud/sdk-go-bundle/products/billing/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// cleanTraffic holds only the structured trafficObj field, dropping the CSV and array duplicates.
type cleanTraffic struct {
	Metadata   *sdk.TrafficMetadata   `json:"metadata,omitempty"`
	TrafficObj *sdk.TrafficTrafficObj `json:"trafficObj,omitempty"`
}

func RegisterTrafficTools(server *mcp.Server, client *sdk.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "billing_traffic",
		Description: "Get network traffic data for your contract for the current billing month. Returns per-datacenter and per-NIC inbound/outbound traffic in bytes.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input tools.BillingContractInput) (*mcp.CallToolResult, any, error) {
		traffic, _, err := client.TrafficApi.TrafficGet(ctx, input.Contract).Execute()
		if err != nil {
			return tools.ToResult(nil, err)
		}
		result := cleanTraffic{
			Metadata:   traffic.Metadata,
			TrafficObj: traffic.TrafficObj,
		}
		return tools.ToResult(result, nil)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "billing_traffic_by_period",
		Description: "Get network traffic data for a specific billing period (YYYY-MM). One month per call. If the user requests a range longer than one month, calculate the number of monthly calls required, inform the user, and ask for permission before proceeding.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input tools.BillingContractPeriodInput) (*mcp.CallToolResult, any, error) {
		if err := tools.ValidatePeriod(input.Period); err != nil {
			return tools.ToResult(nil, err)
		}
		traffic, _, err := client.TrafficApi.TrafficFindByPeriod(ctx, input.Contract, input.Period).Execute()
		if err != nil {
			return tools.ToResult(nil, err)
		}
		result := cleanTraffic{
			Metadata:   traffic.Metadata,
			TrafficObj: traffic.TrafficObj,
		}
		return tools.ToResult(result, nil)
	})
}
