package compute

import (
	"context"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterLocationTools(server *mcp.Server, client *ionos.APIClient, scope tools.Scope) {
	tools.RegisterTool(server, scope, tools.MethodGet, &mcp.Tool{
		Name:        "list_locations",
		Description: "List all available locations (regions) in IONOS CLOUD",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ListLocationsInput) (*mcp.CallToolResult, any, error) {
		depth := int32(1)
		if input.Depth != nil {
			depth = *input.Depth
		}
		r := client.LocationsApi.LocationsGet(ctx).Depth(depth)
		for k, v := range input.Filters {
			r = r.Filter(k, v)
		}
		locations, _, err := r.Execute()
		return tools.ToResult(locations, err)
	})
}
