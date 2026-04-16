package billing

import (
	"context"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
	sdk "github.com/ionos-cloud/sdk-go-bundle/products/billing/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// cleanEvn holds only the structured fields from the Evn response, dropping the evnCSV duplicate.
type cleanEvn struct {
	Metadata    *sdk.EvnMetadata     `json:"metadata,omitempty"`
	Datacenters []sdk.EvnDatacenters `json:"datacenters,omitempty"`
}

func RegisterEvnTools(server *mcp.Server, client *sdk.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_billing_evn",
		Description: "Get provisioning itemized data (EVN) for your contract for the current billing month. Shows per-resource usage intervals grouped by datacenter. For FOCUS v1.3 compliant output, read resource ionos://billing/focus-v1.3.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input tools.BillingContractInput) (*mcp.CallToolResult, any, error) {
		evn, _, err := client.EvnApi.EvnGet(ctx, input.Contract).Execute()
		if err != nil {
			return tools.ToResult(nil, err)
		}
		result := cleanEvn{
			Metadata:    evn.Metadata,
			Datacenters: evn.Datacenters,
		}
		return tools.ToResult(result, nil)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_billing_evn_by_period",
		Description: "Get provisioning itemized data (EVN) for a specific billing period (YYYY-MM). One month per call. If the user requests a range longer than one month, calculate the number of monthly calls required, inform the user, and ask for permission before proceeding. For FOCUS v1.3 compliant output, read resource ionos://billing/focus-v1.3.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input tools.BillingContractPeriodInput) (*mcp.CallToolResult, any, error) {
		if err := tools.ValidatePeriod(input.Period); err != nil {
			return tools.ToResult(nil, err)
		}
		evn, _, err := client.EvnApi.EvnFindByPeriod(ctx, input.Contract, input.Period).Execute()
		if err != nil {
			return tools.ToResult(nil, err)
		}
		result := cleanEvn{
			Metadata:    evn.Metadata,
			Datacenters: evn.Datacenters,
		}
		return tools.ToResult(result, nil)
	})
}
