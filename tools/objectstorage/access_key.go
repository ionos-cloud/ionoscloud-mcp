package objectstorage

import (
	"context"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
	mgmtSDK "github.com/ionos-cloud/sdk-go-bundle/products/objectstoragemanagement/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterAccessKeyTools(server *mcp.Server, client *mgmtSDK.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_object_storage_access_keys",
		Description: "List all Object Storage access keys for the contract. Returns key IDs and metadata but not the secret keys.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		result, _, err := client.AccesskeysApi.AccesskeysGet(ctx).Execute()
		if err != nil {
			return tools.ToResult(nil, err)
		}
		for i := range result.Items {
			result.Items[i].Properties.SecretKey = ""
		}
		return tools.ToResult(result, nil)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_object_storage_access_key",
		Description: "Get details of a specific Object Storage access key by its ID. Returns key metadata but not the secret key.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.AccessKeyIDInput) (*mcp.CallToolResult, any, error) {
		result, _, err := client.AccesskeysApi.AccesskeysFindById(ctx, input.AccessKeyID).Execute()
		if err != nil {
			return tools.ToResult(nil, err)
		}
		result.Properties.SecretKey = ""
		return tools.ToResult(result, nil)
	})
}
