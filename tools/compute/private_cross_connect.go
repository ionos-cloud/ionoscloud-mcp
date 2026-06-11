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
		Annotations: tools.ReadOnly,
		Description: "List all private cross-connects in your IONOS CLOUD account — the links that connect LANs across datacenters within the same location. Shows connectable datacenters and current peers.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		pccs, _, err := client.PrivateCrossConnectsApi.PccsGet(ctx).Execute()
		return tools.ToResult(pccs, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_private_cross_connect",
		Annotations: tools.ReadOnly,
		Description: "Get details of a single private cross-connect, including its connected LAN peers and the datacenters that can join it. Use list_private_cross_connects to find IDs.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.PccIDInput) (*mcp.CallToolResult, any, error) {
		pcc, _, err := client.PrivateCrossConnectsApi.PccsFindById(ctx, input.PccID).Execute()
		return tools.ToResult(pcc, err)
	})
}
