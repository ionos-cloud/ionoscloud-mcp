package activitylog

import (
	"context"

	sdk "github.com/ionos-cloud/sdk-go-bundle/products/activitylog/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterContractTools(server *mcp.Server, client *sdk.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_activitylog_contracts",
		Annotations: tools.ReadOnly,
		Description: "List contracts accessible for IONOS CLOUD activity log queries. Primarily useful for reseller and partner users with multiple contracts. Single-contract users can skip this — their contract number is embedded in the JWT token returned by get_billing_profile or visible in the IONOS DCD console.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		contracts, _, err := client.ContractsApi.GetAvailableContracts(ctx).Execute()
		return tools.ToResult(contracts, err)
	})
}
