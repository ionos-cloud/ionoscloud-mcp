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
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		clusters, _, err := client.KubernetesApi.K8sGet(ctx).Execute()
		return tools.ToResult(clusters, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_k8s_cluster",
		Description: "Get details of a specific Kubernetes cluster",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input tools.K8sClusterIDInput) (*mcp.CallToolResult, any, error) {
		cluster, _, err := client.KubernetesApi.K8sFindByClusterId(ctx, input.K8sClusterID).Execute()
		return tools.ToResult(cluster, err)
	})
}
