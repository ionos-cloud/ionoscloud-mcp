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
		Annotations: tools.ReadOnly,
		Description: "List recent CloudAPI provisioning requests for the account — the asynchronous operations behind create/update/delete actions made by any tool or user. Useful for auditing recent changes; use get_request_status to track a single request.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		requests, _, err := client.RequestsApi.RequestsGet(ctx).Execute()
		return tools.ToResult(requests, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_request",
		Annotations: tools.ReadOnly,
		Description: "Get a single CloudAPI request, including its method, body, and target resources. Use list_requests to find request IDs; use get_request_status if you only need its current state.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.RequestIDInput) (*mcp.CallToolResult, any, error) {
		request, _, err := client.RequestsApi.RequestsFindById(ctx, input.RequestID).Execute()
		return tools.ToResult(request, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_request_status",
		Annotations: tools.ReadOnly,
		Description: "Get the current state (QUEUED, RUNNING, DONE, FAILED) and message of a single CloudAPI request. Lighter than get_request — prefer it when checking whether an operation finished.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.RequestIDInput) (*mcp.CallToolResult, any, error) {
		status, _, err := client.RequestsApi.RequestsStatusGet(ctx, input.RequestID).Execute()
		return tools.ToResult(status, err)
	})
}
