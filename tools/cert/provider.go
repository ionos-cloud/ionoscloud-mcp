package cert

import (
	"context"

	certSDK "github.com/ionos-cloud/sdk-go-bundle/products/cert/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterProviderTools(server *mcp.Server, client *certSDK.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_cert_providers",
		Description: "List all certificate providers in your IONOS Cloud Certificate Manager account. Returns provider configuration but not the external account binding secret.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		result, _, err := client.ProviderApi.ProvidersGet(ctx).Execute()
		if err != nil {
			return tools.ToResult(nil, err)
		}
		for i := range result.Items {
			if eab := result.Items[i].Properties.ExternalAccountBinding; eab != nil {
				eab.KeySecret = nil
			}
		}
		return tools.ToResult(result, nil)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_cert_provider",
		Description: "Get details of a specific certificate provider by ID. Returns provider configuration but not the external account binding secret.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ProviderIDInput) (*mcp.CallToolResult, any, error) {
		result, _, err := client.ProviderApi.ProvidersFindById(ctx, input.ProviderID).Execute()
		if err != nil {
			return tools.ToResult(nil, err)
		}
		if eab := result.Properties.ExternalAccountBinding; eab != nil {
			eab.KeySecret = nil
		}
		return tools.ToResult(result, nil)
	})
}
