package compute

import (
	"context"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterImageTools(server *mcp.Server, client *ionos.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_images",
		Annotations: tools.ReadOnly,
		Description: "List all images available to your account: public OS images, CD-ROM/ISO images, and private images, with licence type and location. Use it to identify the image behind a volume or CD-ROM.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		images, _, err := client.ImagesApi.ImagesGet(ctx).Execute()
		return tools.ToResult(images, err)
	})
}
