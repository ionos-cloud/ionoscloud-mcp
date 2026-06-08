package compute

import (
	"context"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterKubernetesTools(server *mcp.Server, client *ionos.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_kubernetes_clusters",
		Description: "List all Kubernetes clusters in your IONOS Cloud account",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		clusters, _, err := client.KubernetesApi.K8sGet(ctx).Execute()
		return tools.ToResult(clusters, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_kubernetes_cluster",
		Description: "Get details of a specific Kubernetes cluster",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input tools.K8sClusterIDInput) (*mcp.CallToolResult, any, error) {
		cluster, _, err := client.KubernetesApi.K8sFindByClusterId(ctx, input.K8sClusterID).Execute()
		return tools.ToResult(cluster, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_kubernetes_kubeconfig",
		Description: "Get the kubeconfig file for a Kubernetes cluster. WARNING: output contains bearer tokens and TLS credentials — treat as a secret.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input tools.K8sClusterIDInput) (*mcp.CallToolResult, any, error) {
		kubeconfig, _, err := client.KubernetesApi.K8sKubeconfigGet(ctx, input.K8sClusterID).Execute()
		if err != nil {
			return tools.ToResult(nil, err)
		}
		return tools.TextResult(kubeconfig), nil, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_kubernetes_nodepools",
		Description: "List all node pools in a Kubernetes cluster",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input tools.K8sClusterIDInput) (*mcp.CallToolResult, any, error) {
		nodepools, _, err := client.KubernetesApi.K8sNodepoolsGet(ctx, input.K8sClusterID).Execute()
		return tools.ToResult(nodepools, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_kubernetes_nodepool",
		Description: "Get details of a specific Kubernetes node pool",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input tools.K8sNodepoolIDInput) (*mcp.CallToolResult, any, error) {
		nodepool, _, err := client.KubernetesApi.K8sNodepoolsFindById(ctx, input.K8sClusterID, input.NodepoolID).Execute()
		return tools.ToResult(nodepool, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_kubernetes_nodepool_nodes",
		Description: "List all nodes in a Kubernetes node pool",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input tools.K8sNodepoolIDInput) (*mcp.CallToolResult, any, error) {
		nodes, _, err := client.KubernetesApi.K8sNodepoolsNodesGet(ctx, input.K8sClusterID, input.NodepoolID).Execute()
		return tools.ToResult(nodes, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_kubernetes_node",
		Description: "Get details of a specific node in a Kubernetes node pool",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, input tools.K8sNodeIDInput) (*mcp.CallToolResult, any, error) {
		node, _, err := client.KubernetesApi.K8sNodepoolsNodesFindById(ctx, input.K8sClusterID, input.NodepoolID, input.NodeID).Execute()
		return tools.ToResult(node, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_kubernetes_versions",
		Description: "List all available Kubernetes versions in IONOS Cloud",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		versions, _, err := client.KubernetesApi.K8sVersionsGet(ctx).Execute()
		return tools.ToResult(versions, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_kubernetes_default_version",
		Description: "Get the current default Kubernetes version used by new clusters and node pools in IONOS Cloud",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		version, _, err := client.KubernetesApi.K8sVersionsDefaultGet(ctx).Execute()
		if err != nil {
			return tools.ToResult(nil, err)
		}
		return tools.TextResult(version), nil, nil
	})
}
