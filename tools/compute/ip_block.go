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
		Annotations: tools.ReadOnly,
		Description: "List all reserved public IP blocks in your IONOS CLOUD account, with IPs and location. Use get_ip_block to see which resource consumes each IP of a block.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		ipblocks, _, err := client.IPBlocksApi.IpblocksGet(ctx).Execute()
		return tools.ToResult(ipblocks, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_ip_block",
		Annotations: tools.ReadOnly,
		Description: "Get details of a single reserved IP block, including per-IP consumer info (which NIC or load balancer uses each IP). Use list_ip_blocks to find block IDs.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.IpBlockIDInput) (*mcp.CallToolResult, any, error) {
		ipblock, _, err := client.IPBlocksApi.IpblocksFindById(ctx, input.IpBlockID).Execute()
		return tools.ToResult(ipblock, err)
	})
}
