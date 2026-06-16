package compute

import (
	"context"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterPrivateCrossConnectTools(server *mcp.Server, client *ionos.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_private_cross_connects",
		Description: "List all private cross-connects in your IONOS CLOUD account",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ListPrivateCrossConnectsInput) (*mcp.CallToolResult, any, error) {
		depth := int32(1)
		if input.Depth != nil {
			depth = *input.Depth
		}
		pccs, _, err := client.PrivateCrossConnectsApi.PccsGet(ctx).Depth(depth).Execute()
		return tools.ToResult(pccs, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_private_cross_connect",
		Description: "Get details of a specific private cross-connect",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.PccIDInput) (*mcp.CallToolResult, any, error) {
		apiReq := client.PrivateCrossConnectsApi.PccsFindById(ctx, input.PccID)
		if input.Depth != nil {
			apiReq = apiReq.Depth(*input.Depth)
		}
		pcc, _, err := apiReq.Execute()
		return tools.ToResult(pcc, err)
	})
}
