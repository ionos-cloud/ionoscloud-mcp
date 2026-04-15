package billing

import (
	"context"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
	sdk "github.com/ionos-cloud/sdk-go-bundle/products/billing/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterUtilizationTools(server *mcp.Server, client *sdk.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "billing_utilization",
		Description: "Get high-granularity resource utilization for your contract for the current billing period. Shows per-resource metrics (CPU, RAM, storage, DNS) grouped by datacenter. For FOCUS v1.3 compliant output, read resource ionos://billing/focus-v1.3.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input tools.BillingContractInput) (*mcp.CallToolResult, any, error) {
		utilization, _, err := client.UtilizationApi.UtilizationGet(ctx, input.Contract).Execute()
		return tools.ToResult(utilization, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "billing_utilization_by_period",
		Description: "Get high-granularity resource utilization for a specific billing period (YYYY-MM). One month per call. If the user requests a range longer than one month, calculate the number of monthly calls required, inform the user, and ask for permission before proceeding. For FOCUS v1.3 compliant output, read resource ionos://billing/focus-v1.3.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input tools.BillingContractPeriodInput) (*mcp.CallToolResult, any, error) {
		if err := tools.ValidatePeriod(input.Period); err != nil {
			return tools.ToResult(nil, err)
		}
		utilization, _, err := client.UtilizationApi.UtilizationFindByPeriod(ctx, input.Contract, input.Period).Execute()
		return tools.ToResult(utilization, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "billing_utilization_daily",
		Description: "Get high-granularity resource utilization for a specific date (YYYY-MM-DD). Use this for day-level analysis within a month. For FOCUS v1.3 compliant output, read resource ionos://billing/focus-v1.3.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input tools.BillingDateInput) (*mcp.CallToolResult, any, error) {
		if err := tools.ValidateDate(input.Date); err != nil {
			return tools.ToResult(nil, err)
		}
		utilization, _, err := client.UtilizationApi.UtilizationDailyFindByDate(ctx, input.Contract, input.Date).Execute()
		return tools.ToResult(utilization, err)
	})
}
