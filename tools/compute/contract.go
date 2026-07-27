package compute

import (
	"context"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterContractTools(server *mcp.Server, client *ionos.APIClient, scope tools.Scope) {
	tools.RegisterTool(server, scope, tools.MethodGet, &mcp.Tool{
		Name:        "get_contract",
		Description: "Get contract and resource limit information for your IONOS CLOUD account",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.GetContractInput) (*mcp.CallToolResult, any, error) {
		apiReq := client.ContractResourcesApi.ContractsGet(ctx)
		if input.Depth != nil {
			apiReq = apiReq.Depth(*input.Depth)
		}
		contract, _, err := apiReq.Execute()
		return tools.ToResult(contract, err)
	})
}
