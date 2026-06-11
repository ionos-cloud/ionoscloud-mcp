package compute

import (
	"context"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterTemplateTools(server *mcp.Server, client *ionos.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_templates",
		Annotations: tools.ReadOnly,
		Description: "List all predefined Cubes server templates — fixed vCPU/RAM/storage bundles. Templates apply only to CUBE servers, not ENTERPRISE servers.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		templates, _, err := client.TemplatesApi.TemplatesGet(ctx).Execute()
		return tools.ToResult(templates, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_template",
		Annotations: tools.ReadOnly,
		Description: "Get the vCPU, RAM, and storage size of a single Cubes template. Use list_templates to find template IDs.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.TemplateIDInput) (*mcp.CallToolResult, any, error) {
		template, _, err := client.TemplatesApi.TemplatesFindById(ctx, input.TemplateID).Execute()
		return tools.ToResult(template, err)
	})
}
