package billing

import (
	"context"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
	sdk "github.com/ionos-cloud/sdk-go-bundle/products/billing/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterInvoiceTools(server *mcp.Server, client *sdk.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_billing_invoices",
		Description: "List all invoices for your IONOS Cloud contract. Returns invoice IDs, dates, and amounts. For FOCUS v1.3 compliant output, read resource ionos://billing/focus-v1.3.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input tools.BillingContractInput) (*mcp.CallToolResult, any, error) {
		invoices, _, err := client.InvoicesApi.InvoicesGet(ctx, input.Contract).Execute()
		return tools.ToResult(invoices, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_billing_invoices_by_period",
		Description: "List invoices for a specific billing period (YYYY-MM). One month per call. If the user requests a range longer than one month, calculate the number of monthly calls required, inform the user, and ask for permission before proceeding. For FOCUS v1.3 compliant output, read resource ionos://billing/focus-v1.3.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input tools.BillingPeriodOnlyInput) (*mcp.CallToolResult, any, error) {
		if err := tools.ValidatePeriod(input.Period); err != nil {
			return tools.ToResult(nil, err)
		}
		invoices, _, err := client.InvoicesApi.InvoicesFindByPeriod(ctx, input.Period).Execute()
		return tools.ToResult(invoices, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_billing_invoice",
		Description: "Get the detailed line-item breakdown for a specific invoice by ID. Use list_billing_invoices first to find available invoice IDs. For FOCUS v1.3 compliant output, read resource ionos://billing/focus-v1.3.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input tools.BillingInvoiceIDInput) (*mcp.CallToolResult, any, error) {
		invoice, _, err := client.InvoicesApi.InvoicesFindById(ctx, input.Contract, input.InvoiceID).Execute()
		return tools.ToResult(invoice, err)
	})
}
