package billing

import (
	"context"
	"fmt"

	sdk "github.com/ionos-cloud/sdk-go-bundle/products/billing/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

const utilizationScopingHint = "For contracts with many datacenters, scope with regions, datacenter_id, or meter_types — or set top_n=10 for a flat global top-N list — before group_by=datacenter to keep the response under 25 KB. "

func RegisterUtilizationTools(server *mcp.Server, client *sdk.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_billing_utilization",
		Description: "Get per-resource utilization for the current billing period, grouped by datacenter. Defaults exclude zero-quantity meters (set include_zero=true to find idle resources). Use group_by='meter' or 'datacenter' to aggregate further, or top_n=N for a flat global ranking of the largest meters. Filter by datacenter_id, meter_types, or regions to narrow scope. " + utilizationScopingHint + "For FOCUS v1.3 compliant output, read resource ionos://billing/focus-v1.3.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.BillingUtilizationInput) (*mcp.CallToolResult, any, error) {
		opts, err := buildUtilOptions(in.IncludeZero, in.GroupBy, in.DatacenterID, in.MeterTypes, in.Regions, in.TopN)
		if err != nil {
			return tools.ToResult(nil, err)
		}
		raw, _, err := client.UtilizationApi.UtilizationGet(ctx, in.Contract).Execute()
		if err != nil {
			return tools.ToResult(nil, err)
		}
		return tools.ToResult(CompactUtilizationGet(raw, opts), nil)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_billing_utilization_by_period",
		Description: "Get per-resource utilization for a specific billing period (YYYY-MM). One month per call. If the user requests a range longer than one month, calculate the number of monthly calls required, inform the user, and ask for permission before proceeding. Same compaction flags as list_billing_utilization (include_zero, group_by, datacenter_id, meter_types, regions, top_n). " + utilizationScopingHint + "For FOCUS v1.3 compliant output, read resource ionos://billing/focus-v1.3.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.BillingUtilizationPeriodInput) (*mcp.CallToolResult, any, error) {
		if err := tools.ValidatePeriod(in.Period); err != nil {
			return tools.ToResult(nil, err)
		}
		opts, err := buildUtilOptions(in.IncludeZero, in.GroupBy, in.DatacenterID, in.MeterTypes, in.Regions, in.TopN)
		if err != nil {
			return tools.ToResult(nil, err)
		}
		raw, _, err := client.UtilizationApi.UtilizationFindByPeriod(ctx, in.Contract, in.Period).Execute()
		if err != nil {
			return tools.ToResult(nil, err)
		}
		return tools.ToResult(CompactUtilizationGet(raw, opts), nil)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_billing_utilization_daily",
		Description: "Get per-resource utilization for a specific date (YYYY-MM-DD). Use this for day-level analysis within a month. Same compaction flags as list_billing_utilization. " + utilizationScopingHint + "For FOCUS v1.3 compliant output, read resource ionos://billing/focus-v1.3.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in tools.BillingUtilizationDateInput) (*mcp.CallToolResult, any, error) {
		if err := tools.ValidateDate(in.Date); err != nil {
			return tools.ToResult(nil, err)
		}
		opts, err := buildUtilOptions(in.IncludeZero, in.GroupBy, in.DatacenterID, in.MeterTypes, in.Regions, in.TopN)
		if err != nil {
			return tools.ToResult(nil, err)
		}
		raw, _, err := client.UtilizationApi.UtilizationDailyFindByDate(ctx, in.Contract, in.Date).Execute()
		if err != nil {
			return tools.ToResult(nil, err)
		}
		return tools.ToResult(CompactUtilizationDaily(raw, opts), nil)
	})
}

// buildUtilOptions validates and assembles a CompactOptions struct for utilization handlers.
// Returns an error before any HTTP call if group_by or top_n is invalid.
func buildUtilOptions(includeZero *bool, groupBy *string, dcID *string, meterTypes, regions []string, topN *int32) (CompactOptions, error) {
	opts := CompactOptions{
		IncludeZero:  includeZero != nil && *includeZero,
		DatacenterID: dcID,
		MeterTypes:   meterTypes,
		Regions:      regions,
		TopN:         topN,
	}
	if groupBy != nil {
		switch *groupBy {
		case "", "meter", "datacenter":
			opts.GroupBy = *groupBy
		default:
			return opts, fmt.Errorf("invalid group_by %q: must be '', 'meter', or 'datacenter'", *groupBy)
		}
	}
	if topN != nil && *topN <= 0 {
		return opts, fmt.Errorf("top_n must be > 0, got %d", *topN)
	}
	return opts, nil
}
