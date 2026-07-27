package compute

import (
	"context"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterIpBlockTools(server *mcp.Server, client *ionos.APIClient, scope tools.Scope) {
	tools.RegisterTool(server, scope, tools.MethodGet, &mcp.Tool{
		Name:        "list_ip_blocks",
		Description: "List all reserved IP blocks in your IONOS CLOUD account",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ListIPBlocksInput) (*mcp.CallToolResult, any, error) {
		depth := int32(1)
		if input.Depth != nil {
			depth = *input.Depth
		}
		r := client.IPBlocksApi.IpblocksGet(ctx).Depth(depth)
		for k, v := range input.Filters {
			r = r.Filter(k, v)
		}
		ipblocks, _, err := r.Execute()
		return tools.ToResult(ipblocks, err)
	})

	tools.RegisterTool(server, scope, tools.MethodGet, &mcp.Tool{
		Name:        "get_ip_block",
		Description: "Get details of a specific reserved IP block",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.IpBlockIDInput) (*mcp.CallToolResult, any, error) {
		apiReq := client.IPBlocksApi.IpblocksFindById(ctx, input.IpBlockID)
		if input.Depth != nil {
			apiReq = apiReq.Depth(*input.Depth)
		}
		ipblock, _, err := apiReq.Execute()
		return tools.ToResult(ipblock, err)
	})
}
