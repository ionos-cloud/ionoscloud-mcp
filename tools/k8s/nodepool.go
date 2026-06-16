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
		depth := int32(1)
		if input.Depth != nil {
			depth = *input.Depth
		}
		nodepools, _, err := client.KubernetesApi.K8sNodepoolsGet(ctx, input.K8sClusterID).Depth(depth).Execute()
		return tools.ToResult(nodepools, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_k8s_nodepool",
		Description: "Get details of a specific Kubernetes node pool",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input tools.K8sNodepoolIDInput) (*mcp.CallToolResult, any, error) {
		apiReq := client.KubernetesApi.K8sNodepoolsFindById(ctx, input.K8sClusterID, input.NodepoolID)
		if input.Depth != nil {
			apiReq = apiReq.Depth(*input.Depth)
		}
		nodepool, _, err := apiReq.Execute()
		return tools.ToResult(nodepool, err)
	})
}
