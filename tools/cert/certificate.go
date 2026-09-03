package cert

import (
	"context"

	certSDK "github.com/ionos-cloud/sdk-go-bundle/products/cert/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterCertificateTools(server *mcp.Server, client *certSDK.APIClient, scope tools.Scope) {
	tools.RegisterTool(server, scope, tools.MethodGet, &mcp.Tool{
		Name:        "list_cert_certificates",
		Description: "List all SSL/TLS certificates in your IONOS Cloud Certificate Manager account. Returns certificate metadata and public key material but not the private key.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		result, _, err := client.CertificateApi.CertificatesGet(ctx).Execute()
		if err != nil {
			return tools.ToResult(nil, err)
		}
		return tools.ToResult(redactCertificateList(result), nil)
	})

	tools.RegisterTool(server, scope, tools.MethodGet, &mcp.Tool{
		Name:        "get_cert_certificate",
		Description: "Get details of a specific SSL/TLS certificate by ID. Returns certificate metadata and public key material but not the private key.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.CertificateIDInput) (*mcp.CallToolResult, any, error) {
		result, _, err := client.CertificateApi.CertificatesFindById(ctx, input.CertificateID).Execute()
		if err != nil {
			return tools.ToResult(nil, err)
		}
		return tools.ToResult(redactCertificate(result), nil)
	})
}
