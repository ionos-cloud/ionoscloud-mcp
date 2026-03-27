package compute

import (
	"context"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterImageTools(server *mcp.Server, client *ionos.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_images",
		Description: "List all available images (OS templates) in IONOS Cloud",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		images, _, err := client.ImagesApi.ImagesGet(ctx).Execute()
		return tools.ToResult(images, err)
	})
}
