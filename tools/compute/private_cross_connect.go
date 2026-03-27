package compute

import (
	"context"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterPrivateCrossConnectTools(server *mcp.Server, client *ionos.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_private_cross_connects",
		Description: "List all private cross-connects in your IONOS Cloud account",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		pccs, _, err := client.PrivateCrossConnectsApi.PccsGet(ctx).Execute()
		return tools.ToResult(pccs, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_private_cross_connect",
		Description: "Get details of a specific private cross-connect",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.PccIDInput) (*mcp.CallToolResult, any, error) {
		pcc, _, err := client.PrivateCrossConnectsApi.PccsFindById(ctx, input.PccID).Execute()
		return tools.ToResult(pcc, err)
	})
}
