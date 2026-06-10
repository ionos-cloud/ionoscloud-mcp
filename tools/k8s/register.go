package k8s

import (
	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterAll registers all Managed Kubernetes tools with the MCP server.
func RegisterAll(server *mcp.Server, client *ionos.APIClient) {
	RegisterClusterTools(server, client)
	RegisterNodepoolTools(server, client)
	RegisterNodeTools(server, client)
	RegisterVersionTools(server, client)
}
