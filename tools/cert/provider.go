package cert

import (
	"context"

	certSDK "github.com/ionos-cloud/sdk-go-bundle/products/cert/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterProviderTools(server *mcp.Server, client *certSDK.APIClient, scope tools.Scope) {
	tools.RegisterTool(server, scope, tools.MethodGet, &mcp.Tool{
		Name:        "list_cert_providers",
		Description: "List all certificate providers in your IONOS Cloud Certificate Manager account. Returns provider configuration but not the external account binding secret.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		result, _, err := client.ProviderApi.ProvidersGet(ctx).Execute()
		if err != nil {
			return tools.ToResult(nil, err)
		}
		return tools.ToResult(redactProviderList(result), nil)
	})

	tools.RegisterTool(server, scope, tools.MethodGet, &mcp.Tool{
		Name:        "get_cert_provider",
		Description: "Get details of a specific certificate provider by ID. Returns provider configuration but not the external account binding secret.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ProviderIDInput) (*mcp.CallToolResult, any, error) {
		result, _, err := client.ProviderApi.ProvidersFindById(ctx, input.ProviderID).Execute()
		if err != nil {
			return tools.ToResult(nil, err)
		}
		return tools.ToResult(redactProvider(result), nil)
	})
}
