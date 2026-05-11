package compute

import (
	"context"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterSnapshotTools(server *mcp.Server, client *ionos.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_snapshots",
		Description: "List all snapshots in your IONOS CLOUD account",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		snapshots, _, err := client.SnapshotsApi.SnapshotsGet(ctx).Execute()
		return tools.ToResult(snapshots, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_snapshot",
		Description: "Get details of a specific snapshot",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.SnapshotIDInput) (*mcp.CallToolResult, any, error) {
		snapshot, _, err := client.SnapshotsApi.SnapshotsFindById(ctx, input.SnapshotID).Execute()
		return tools.ToResult(snapshot, err)
	})
}
