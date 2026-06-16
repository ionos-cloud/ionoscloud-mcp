package compute

import (
	"context"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterSnapshotTools(server *mcp.Server, client *ionos.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_snapshots",
		Description: "List all snapshots in your IONOS CLOUD account",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ListSnapshotsInput) (*mcp.CallToolResult, any, error) {
		depth := int32(1)
		if input.Depth != nil {
			depth = *input.Depth
		}
		snapshots, _, err := client.SnapshotsApi.SnapshotsGet(ctx).Depth(depth).Execute()
		return tools.ToResult(snapshots, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_snapshot",
		Description: "Get details of a specific snapshot",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.SnapshotIDInput) (*mcp.CallToolResult, any, error) {
		apiReq := client.SnapshotsApi.SnapshotsFindById(ctx, input.SnapshotID)
		if input.Depth != nil {
			apiReq = apiReq.Depth(*input.Depth)
		}
		snapshot, _, err := apiReq.Execute()
		return tools.ToResult(snapshot, err)
	})
}
