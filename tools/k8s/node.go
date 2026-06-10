package k8s

import (
	"context"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterNodeTools(server *mcp.Server, client *ionos.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_k8s_nodepool_nodes",
		Description: "List all nodes in a Kubernetes node pool",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input tools.K8sNodepoolIDInput) (*mcp.CallToolResult, any, error) {
		nodes, _, err := client.KubernetesApi.K8sNodepoolsNodesGet(ctx, input.K8sClusterID, input.NodepoolID).Execute()
		return tools.ToResult(nodes, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_k8s_node",
		Description: "Get details of a specific node in a Kubernetes node pool",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input tools.K8sNodeIDInput) (*mcp.CallToolResult, any, error) {
		node, _, err := client.KubernetesApi.K8sNodepoolsNodesFindById(ctx, input.K8sClusterID, input.NodepoolID, input.NodeID).Execute()
		return tools.ToResult(node, err)
	})
}
