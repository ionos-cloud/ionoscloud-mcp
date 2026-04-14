package billing

import (
	"context"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
	sdk "github.com/ionos-cloud/sdk-go-bundle/products/billing/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterProfileTools(server *mcp.Server, client *sdk.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "billing_profile",
		Description: "Get the billing profile for your IONOS Cloud account. Call this first before any other billing tool — the contract number in the response is required by all other billing tools.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		profile, _, err := client.ProfileApi.ProfilesGet(ctx).Execute()
		return tools.ToResult(profile, err)
	})
}
