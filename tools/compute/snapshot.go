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
		Annotations: tools.ReadOnly,
		Description: "List all volume snapshots in your IONOS CLOUD account, across all datacenters, with size, location, and source-volume properties.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		snapshots, _, err := client.SnapshotsApi.SnapshotsGet(ctx).Execute()
		return tools.ToResult(snapshots, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_snapshot",
		Annotations: tools.ReadOnly,
		Description: "Get details of a single volume snapshot: size, location, licence type, and the disk features captured from the source volume. Use list_snapshots to find snapshot IDs.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.SnapshotIDInput) (*mcp.CallToolResult, any, error) {
		snapshot, _, err := client.SnapshotsApi.SnapshotsFindById(ctx, input.SnapshotID).Execute()
		return tools.ToResult(snapshot, err)
	})
}
