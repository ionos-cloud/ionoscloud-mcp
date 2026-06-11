package compute

import (
	"context"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterLocationTools(server *mcp.Server, client *ionos.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_locations",
		Annotations: tools.ReadOnly,
		Description: "List all IONOS CLOUD locations (e.g. de/fra, us/las) where datacenters can be provisioned, with the CPU architectures available in each.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		locations, _, err := client.LocationsApi.LocationsGet(ctx).Execute()
		return tools.ToResult(locations, err)
	})
}
