package compute

import (
	"context"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterLocationTools(server *mcp.Server, client *ionos.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_locations",
		Description: "List all available locations (regions) in IONOS Cloud",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		locations, _, err := client.LocationsApi.LocationsGet(ctx).Execute()
		return tools.ToResult(locations, err)
	})
}
