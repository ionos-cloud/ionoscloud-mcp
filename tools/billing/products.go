package billing

import (
	"context"
	"fmt"
	"strings"

	sdk "github.com/ionos-cloud/sdk-go-bundle/products/billing/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

type cleanProducts struct {
	Metadata      *sdk.Metadata `json:"metadata,omitempty"`
	Liability     *string       `json:"liability,omitempty"`
	AppliedFilter string        `json:"appliedFilter"`
	MatchCount    int           `json:"matchCount"`
	Products      []sdk.Product `json:"products"`
}

func RegisterProductTools(server *mcp.Server, client *sdk.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_billing_products",
		Annotations: tools.ReadOnly,
		Description: "Search the IONOS CLOUD product/pricing catalog by keyword. " +
			"The filter is applied client-side as a case-insensitive partial match on each product's description — it is not an API-level parameter. " +
			"Returns non-deprecated products whose description contains the filter string. " +
			"IMPORTANT: Only call this tool when the user has explicitly specified a product or category they want to see pricing for. " +
			"If the user asks a broad question like 'what are the prices' or 'show me all products', do NOT guess keywords or call this tool multiple times — instead ask the user which specific product or category they are interested in. " +
			"Examples of valid filters: 'RAM', 'core', 'storage', 'Kubernetes', 'Postgres', 'network', 'Windows'.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input tools.BillingProductsInput) (*mcp.CallToolResult, any, error) {
		filter := strings.TrimSpace(input.Filter)
		if filter == "" {
			return tools.ToResult(nil, fmt.Errorf("filter is required: provide a keyword to search products by (e.g. 'RAM', 'Kubernetes', 'storage')"))
		}

		products, _, err := client.ProductsApi.ProductsGet(ctx, input.Contract).Execute()
		if err != nil {
			return tools.ToResult(nil, err)
		}

		filterLower := strings.ToLower(filter)
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
			Liability:     products.Liability,
			AppliedFilter: filter,
			MatchCount:    len(filtered),
			Products:      filtered,
		}
		return tools.ToResult(result, nil)
	})
}
