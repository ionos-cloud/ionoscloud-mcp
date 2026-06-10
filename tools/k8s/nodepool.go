package k8s

import (
	"context"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterNodepoolTools(server *mcp.Server, client *ionos.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_k8s_nodepools",
		Description: "List all node pools in a Kubernetes cluster",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input tools.K8sClusterIDInput) (*mcp.CallToolResult, any, error) {
		nodepools, _, err := client.KubernetesApi.K8sNodepoolsGet(ctx, input.K8sClusterID).Execute()
		return tools.ToResult(nodepools, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_k8s_nodepool",
		Description: "Get details of a specific Kubernetes node pool",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input tools.K8sNodepoolIDInput) (*mcp.CallToolResult, any, error) {
		nodepool, _, err := client.KubernetesApi.K8sNodepoolsFindById(ctx, input.K8sClusterID, input.NodepoolID).Execute()
		return tools.ToResult(nodepool, err)
	})
}
