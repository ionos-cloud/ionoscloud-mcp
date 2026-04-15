package billing

import (
	"context"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
	sdk "github.com/ionos-cloud/sdk-go-bundle/products/billing/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterUsageTools(server *mcp.Server, client *sdk.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "billing_usage",
		Description: "Get aggregated resource usage for your contract for the current billing period. Shows metered quantities (CPU hours, GB-hours, etc.) grouped by datacenter. For FOCUS v1.3 compliant output, read resource ionos://billing/focus-v1.3.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input tools.BillingContractInput) (*mcp.CallToolResult, any, error) {
		usage, _, err := client.UsageApi.UsageGet(ctx, input.Contract).Execute()
		return tools.ToResult(usage, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "billing_usage_by_datacenter",
		Description: "Get aggregated resource usage for a specific datacenter (VDC UUID) in the current billing period. Use billing_usage first to find datacenter IDs. For FOCUS v1.3 compliant output, read resource ionos://billing/focus-v1.3.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input tools.BillingDatacenterInput) (*mcp.CallToolResult, any, error) {
		usage, _, err := client.UsageApi.UsageFindByDatacenter(ctx, input.Contract, input.DatacenterID).Execute()
		return tools.ToResult(usage, err)
	})
}
