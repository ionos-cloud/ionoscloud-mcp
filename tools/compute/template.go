package compute

import (
	"context"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterTemplateTools(server *mcp.Server, client *ionos.APIClient, scope tools.Scope) {
	tools.RegisterTool(server, scope, tools.MethodGet, &mcp.Tool{
		Name:        "list_templates",
		Description: "List all available server templates in IONOS CLOUD",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ListTemplatesInput) (*mcp.CallToolResult, any, error) {
		depth := int32(1)
		if input.Depth != nil {
			depth = *input.Depth
		}
		r := client.TemplatesApi.TemplatesGet(ctx).Depth(depth)
		for k, v := range input.Filters {
			r = r.Filter(k, v)
		}
		templates, _, err := r.Execute()
		return tools.ToResult(templates, err)
	})

	tools.RegisterTool(server, scope, tools.MethodGet, &mcp.Tool{
		Name:        "get_template",
		Description: "Get details of a specific server template",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.TemplateIDInput) (*mcp.CallToolResult, any, error) {
		apiReq := client.TemplatesApi.TemplatesFindById(ctx, input.TemplateID)
		if input.Depth != nil {
			apiReq = apiReq.Depth(*input.Depth)
		}
		template, _, err := apiReq.Execute()
		return tools.ToResult(template, err)
	})
}
