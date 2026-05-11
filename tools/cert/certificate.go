package cert

import (
	"context"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
	certSDK "github.com/ionos-cloud/sdk-go-bundle/products/cert/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterCertificateTools(server *mcp.Server, client *certSDK.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_cert_certificates",
		Description: "List all SSL/TLS certificates in your IONOS Cloud Certificate Manager account. Returns certificate metadata and public key material but not the private key.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		result, _, err := client.CertificateApi.CertificatesGet(ctx).Execute()
		if err != nil {
			return tools.ToResult(nil, err)
		}
		for i := range result.Items {
			result.Items[i].Properties.PrivateKey = ""
		}
		return tools.ToResult(result, nil)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_cert_certificate",
		Description: "Get details of a specific SSL/TLS certificate by ID. Returns certificate metadata and public key material but not the private key.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.CertificateIDInput) (*mcp.CallToolResult, any, error) {
		result, _, err := client.CertificateApi.CertificatesFindById(ctx, input.CertificateID).Execute()
		if err != nil {
			return tools.ToResult(nil, err)
		}
		result.Properties.PrivateKey = ""
		return tools.ToResult(result, nil)
	})
}
