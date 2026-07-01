package compute

import (
	"context"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterRequestTools(server *mcp.Server, client *ionos.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_requests",
		Description: "List all API requests in your IONOS CLOUD account",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ListRequestsInput) (*mcp.CallToolResult, any, error) {
		depth := int32(1)
		if input.Depth != nil {
			depth = *input.Depth
		}
		r := client.RequestsApi.RequestsGet(ctx).Depth(depth)
		for k, v := range input.Filters {
			r = r.Filter(k, v)
		}
		requests, _, err := r.Execute()
		return tools.ToResult(requests, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_request",
		Description: "Get details of a specific API request",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.RequestIDInput) (*mcp.CallToolResult, any, error) {
		apiReq := client.RequestsApi.RequestsFindById(ctx, input.RequestID)
		if input.Depth != nil {
			apiReq = apiReq.Depth(*input.Depth)
		}
		request, _, err := apiReq.Execute()
		return tools.ToResult(request, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_request_status",
		Description: "Get the status of a specific API request",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.RequestIDInput) (*mcp.CallToolResult, any, error) {
		apiReq := client.RequestsApi.RequestsStatusGet(ctx, input.RequestID)
		if input.Depth != nil {
			apiReq = apiReq.Depth(*input.Depth)
		}
		status, _, err := apiReq.Execute()
		return tools.ToResult(status, err)
	})
}
