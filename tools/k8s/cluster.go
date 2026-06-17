package k8s

import (
	"context"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterClusterTools(server *mcp.Server, client *ionos.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_k8s_clusters",
		Description: "List all Kubernetes clusters in your IONOS CLOUD account",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input tools.ListK8sClustersInput) (*mcp.CallToolResult, any, error) {
		depth := int32(1)
		if input.Depth != nil {
			depth = *input.Depth
		}
		clusters, _, err := client.KubernetesApi.K8sGet(ctx).Depth(depth).Execute()
		return tools.ToResult(clusters, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_k8s_cluster",
		Description: "Get details of a specific Kubernetes cluster",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input tools.K8sClusterIDInput) (*mcp.CallToolResult, any, error) {
		apiReq := client.KubernetesApi.K8sFindByClusterId(ctx, input.K8sClusterID)
		if input.Depth != nil {
			apiReq = apiReq.Depth(*input.Depth)
		}
		cluster, _, err := apiReq.Execute()
		return tools.ToResult(cluster, err)
	})
}
