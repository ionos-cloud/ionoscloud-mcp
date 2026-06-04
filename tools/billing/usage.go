package billing

import (
	"context"

	sdk "github.com/ionos-cloud/sdk-go-bundle/products/billing/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterUsageTools(server *mcp.Server, client *sdk.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_billing_usage",
		Description: "Get aggregated resource usage for your contract for the current billing period. Shows metered quantities (CPU hours, GB-hours, etc.) grouped by datacenter. Defaults exclude zero-quantity meters (set include_zero=true to keep them). Filter by datacenter_id to narrow scope. For FOCUS v1.3 compliant output, read resource ionos://billing/focus-v1.3.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.BillingUsageInput) (*mcp.CallToolResult, any, error) {
		opts := CompactOptions{
			IncludeZero:  in.IncludeZero != nil && *in.IncludeZero,
			DatacenterID: in.DatacenterID,
		}
		raw, _, err := client.UsageApi.UsageGet(ctx, in.Contract).Execute()
		if err != nil {
			return tools.ToResult(nil, err)
		}
		return tools.ToResult(CompactUsageGet(raw, opts), nil)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_billing_usage_by_datacenter",
		Description: "Get aggregated resource usage for a specific datacenter (VDC UUID) in the current billing period. Use list_billing_usage first to find datacenter IDs. Defaults exclude zero-quantity meters (set include_zero=true to keep them). For FOCUS v1.3 compliant output, read resource ionos://billing/focus-v1.3.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.BillingUsageDatacenterInput) (*mcp.CallToolResult, any, error) {
		opts := CompactOptions{
			IncludeZero: in.IncludeZero != nil && *in.IncludeZero,
		}
		raw, _, err := client.UsageApi.UsageFindByDatacenter(ctx, in.Contract, in.DatacenterID).Execute()
		if err != nil {
			return tools.ToResult(nil, err)
		}
		return tools.ToResult(CompactUsageGet(raw, opts), nil)
	})
}
