package billing

import (
	"context"

	sdk "github.com/ionos-cloud/sdk-go-bundle/products/billing/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterProfileTools(server *mcp.Server, client *sdk.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_billing_profile",
		Annotations: tools.ReadOnly,
		Description: "Get the billing profile for your IONOS CLOUD account. Call this first before any other billing tool — the contract number in the response is required by all other billing tools.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		profile, _, err := client.ProfileApi.ProfilesGet(ctx).Execute()
		return tools.ToResult(profile, err)
	})
}
