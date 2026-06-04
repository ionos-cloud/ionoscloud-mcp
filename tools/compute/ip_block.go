package compute

import (
	"context"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterIpBlockTools(server *mcp.Server, client *ionos.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_ip_blocks",
		Description: "List all reserved IP blocks in your IONOS CLOUD account",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		ipblocks, _, err := client.IPBlocksApi.IpblocksGet(ctx).Execute()
		return tools.ToResult(ipblocks, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_ip_block",
		Description: "Get details of a specific reserved IP block",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.IpBlockIDInput) (*mcp.CallToolResult, any, error) {
		ipblock, _, err := client.IPBlocksApi.IpblocksFindById(ctx, input.IpBlockID).Execute()
		return tools.ToResult(ipblock, err)
	})
}
