package k8s

import (
	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

// RegisterAll registers all Managed Kubernetes tools with the MCP server. Write
// tools appear only when scope allows; confirm backs the two-phase create, delete
// and recreate flows.
func RegisterAll(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	RegisterClusterTools(server, client, scope)
	RegisterClusterWriteTools(server, client, scope, confirm)
	RegisterNodepoolTools(server, client, scope)
	RegisterNodepoolWriteTools(server, client, scope, confirm)
	RegisterNodeTools(server, client, scope)
	RegisterNodeWriteTools(server, client, scope, confirm)
	RegisterVersionTools(server, client, scope)
}
