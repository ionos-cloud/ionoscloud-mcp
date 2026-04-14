package billing

import (
	"context"
	"strings"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
	sdk "github.com/ionos-cloud/sdk-go-bundle/products/billing/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type cleanProducts struct {
	Metadata      *sdk.Metadata `json:"metadata,omitempty"`
	AppliedFilter string        `json:"appliedFilter"`
	MatchCount    int           `json:"matchCount"`
	Products      []sdk.Product `json:"products"`
}

func RegisterProductTools(server *mcp.Server, client *sdk.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name: "billing_products",
		Description: "Search the IONOS Cloud product/pricing catalog by keyword. " +
			"Returns non-deprecated products whose description matches the filter. " +
			"IMPORTANT: Only call this tool when the user has explicitly specified a product or category they want to see pricing for. " +
			"If the user asks a broad question like 'what are the prices' or 'show me all products', do NOT guess keywords or call this tool multiple times — instead ask the user which specific product or category they are interested in. " +
			"Examples of valid filters: 'RAM', 'core', 'storage', 'Kubernetes', 'Postgres', 'network', 'Windows'.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input tools.BillingProductsInput) (*mcp.CallToolResult, any, error) {
		products, _, err := client.ProductsApi.ProductsGet(ctx, input.Contract).Execute()
		if err != nil {
			return tools.ToResult(nil, err)
		}

		filterLower := strings.ToLower(input.Filter)
		var filtered []sdk.Product
		for _, p := range products.GetProducts() {
			if p.GetDeprecated() {
				continue
			}
			if !strings.Contains(strings.ToLower(p.GetMeterDesc()), filterLower) {
				continue
			}
			filtered = append(filtered, p)
		}

		result := cleanProducts{
			Metadata:      products.Metadata,
			AppliedFilter: input.Filter,
			MatchCount:    len(filtered),
			Products:      filtered,
		}
		return tools.ToResult(result, nil)
	})
}
