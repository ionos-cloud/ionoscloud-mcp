package objectstorage

import (
	sdk "github.com/ionos-cloud/sdk-go-bundle/products/objectstorage/v2"
	mgmtSDK "github.com/ionos-cloud/sdk-go-bundle/products/objectstoragemanagement/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterAll registers all Object Storage tools with the MCP server.
func RegisterAll(server *mcp.Server, client *sdk.APIClient, mgmtClient *mgmtSDK.APIClient) {
	RegisterBucketTools(server, client)
	RegisterBucketConfigTools(server, client)
	RegisterObjectTools(server, client)
	RegisterAccessKeyTools(server, mgmtClient)
	RegisterRegionTools(server, mgmtClient)
}
