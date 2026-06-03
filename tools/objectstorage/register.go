package objectstorage

import (
	sdk "github.com/ionos-cloud/sdk-go-bundle/products/objectstorage/v2"
	mgmtSDK "github.com/ionos-cloud/sdk-go-bundle/products/objectstoragemanagement/v2"
	"github.com/ionos-cloud/sdk-go-bundle/shared"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterAll registers all Object Storage tools with the MCP server.
func RegisterAll(server *mcp.Server, client *sdk.APIClient, mgmtClient *mgmtSDK.APIClient, cfg *shared.Configuration) {
	cache := newClientCache(client, mgmtClient, cfg)
	RegisterBucketTools(server, cache)
	RegisterBucketConfigTools(server, cache)
	RegisterObjectTools(server, cache)
	RegisterAccessKeyTools(server, mgmtClient)
	RegisterRegionTools(server, mgmtClient)
}
