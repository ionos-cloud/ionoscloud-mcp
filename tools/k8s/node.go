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
		depth := int32(1)
		if input.Depth != nil {
			depth = *input.Depth
		}
		nodes, _, err := client.KubernetesApi.K8sNodepoolsNodesGet(ctx, input.K8sClusterID, input.NodepoolID).Depth(depth).Execute()
		return tools.ToResult(nodes, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_k8s_node",
		Description: "Get details of a specific node in a Kubernetes node pool",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input tools.K8sNodeIDInput) (*mcp.CallToolResult, any, error) {
		apiReq := client.KubernetesApi.K8sNodepoolsNodesFindById(ctx, input.K8sClusterID, input.NodepoolID, input.NodeID)
		if input.Depth != nil {
			apiReq = apiReq.Depth(*input.Depth)
		}
		node, _, err := apiReq.Execute()
		return tools.ToResult(node, err)
	})
}
