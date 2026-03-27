package compute

import (
	"context"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterRequestTools(server *mcp.Server, client *ionos.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_requests",
		Description: "List all API requests in your IONOS Cloud account",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		requests, _, err := client.RequestsApi.RequestsGet(ctx).Execute()
		return tools.ToResult(requests, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_request",
		Description: "Get details of a specific API request",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.RequestIDInput) (*mcp.CallToolResult, any, error) {
		request, _, err := client.RequestsApi.RequestsFindById(ctx, input.RequestID).Execute()
		return tools.ToResult(request, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_request_status",
		Description: "Get the status of a specific API request",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.RequestIDInput) (*mcp.CallToolResult, any, error) {
		status, _, err := client.RequestsApi.RequestsStatusGet(ctx, input.RequestID).Execute()
		return tools.ToResult(status, err)
	})
}
