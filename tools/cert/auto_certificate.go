package cert

import (
	"context"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
	certSDK "github.com/ionos-cloud/sdk-go-bundle/products/cert/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterAutoCertificateTools(server *mcp.Server, client *certSDK.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_auto_certificates",
		Description: "List all auto-certificates in your IONOS Cloud Certificate Manager account",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		result, _, err := client.AutoCertificateApi.AutoCertificatesGet(ctx).Execute()
		return tools.ToResult(result, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_auto_certificate",
		Description: "Get details of a specific auto-certificate by ID",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.AutoCertificateIDInput) (*mcp.CallToolResult, any, error) {
		result, _, err := client.AutoCertificateApi.AutoCertificatesFindById(ctx, input.AutoCertificateID).Execute()
		return tools.ToResult(result, err)
	})
}
