package compute

import (
	"context"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterContractTools(server *mcp.Server, client *ionos.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_contract",
		Annotations: tools.ReadOnly,
		Description: "Get the contract(s) of the authenticated account: contract number, owner, status, and resource limits (max cores, RAM, IPs, etc.). Use it to check quota headroom; for billing and cost data use get_billing_profile instead.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		contract, _, err := client.ContractResourcesApi.ContractsGet(ctx).Execute()
		return tools.ToResult(contract, err)
	})
}
