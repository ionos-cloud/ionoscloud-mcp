package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ionos-cloud/ionoscloud-mcp/pkg/api"
)

func initClusters() []api.ServerTool {
	return []api.ServerTool{
		{
			Tool: api.Tool{
				Name:        "list_k8s_clusters",
				Description: "List all Kubernetes clusters in your IONOS Cloud account",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {},
					"required": []
				}`),
				Annotations: api.ReadOnly("List K8s Clusters"),
			},
			Handler: listK8sClusters,
		},
		{
			Tool: api.Tool{
				Name:        "get_k8s_cluster",
				Description: "Get details of a specific Kubernetes cluster",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"k8s_cluster_id": {"type": "string", "description": "The ID of the Kubernetes cluster"}
					},
					"required": ["k8s_cluster_id"]
				}`),
				Annotations: api.ReadOnly("Get K8s Cluster"),
			},
			Handler: getK8sCluster,
		},
		{
			Tool: api.Tool{
				Name:        "get_k8s_kubeconfig",
				Description: "Get the kubeconfig for a Kubernetes cluster",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"k8s_cluster_id": {"type": "string", "description": "The ID of the Kubernetes cluster"}
					},
					"required": ["k8s_cluster_id"]
				}`),
				Annotations: api.ReadOnly("Get Kubeconfig"),
			},
			Handler: getK8sKubeconfig,
		},
		{
			Tool: api.Tool{
				Name:        "list_k8s_versions",
				Description: "List all available Kubernetes versions",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {},
					"required": []
				}`),
				Annotations: api.ReadOnly("List K8s Versions"),
			},
			Handler: listK8sVersions,
		},
	}
}

func listK8sClusters(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	clusters, _, err := params.Client.KubernetesApi.K8sGet(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to list Kubernetes clusters: %w", err)
	}
	return api.MarshalResult(clusters, "Kubernetes clusters")
}

func getK8sCluster(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	k8sClusterID, ok := api.GetRequiredString(params.Arguments, "k8s_cluster_id")
	if !ok {
		return nil, fmt.Errorf("k8s_cluster_id is required")
	}

	cluster, _, err := params.Client.KubernetesApi.K8sFindByClusterId(ctx, k8sClusterID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get Kubernetes cluster: %w", err)
	}
	return api.MarshalResult(cluster, "Kubernetes cluster")
}

func getK8sKubeconfig(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	k8sClusterID, ok := api.GetRequiredString(params.Arguments, "k8s_cluster_id")
	if !ok {
		return nil, fmt.Errorf("k8s_cluster_id is required")
	}

	kubeconfig, _, err := params.Client.KubernetesApi.K8sKubeconfigGet(ctx, k8sClusterID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get kubeconfig: %w", err)
	}
	return api.MarshalResult(kubeconfig, "kubeconfig")
}

func listK8sVersions(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	versions, _, err := params.Client.KubernetesApi.K8sVersionsGet(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to list Kubernetes versions: %w", err)
	}
	return api.MarshalResult(versions, "Kubernetes versions")
}
