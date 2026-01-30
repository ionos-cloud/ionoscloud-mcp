package kubernetes

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ionos-cloud/ionoscloud-mcp/pkg/api"
)

func initNodepools() []api.ServerTool {
	return []api.ServerTool{
		{
			Tool: api.Tool{
				Name:        "list_k8s_nodepools",
				Description: "List all node pools in a Kubernetes cluster",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"k8s_cluster_id": {"type": "string", "description": "The ID of the Kubernetes cluster"}
					},
					"required": ["k8s_cluster_id"]
				}`),
				Annotations: api.ReadOnly("List K8s Node Pools"),
			},
			Handler: listK8sNodepools,
		},
		{
			Tool: api.Tool{
				Name:        "get_k8s_nodepool",
				Description: "Get details of a specific node pool",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"k8s_cluster_id": {"type": "string", "description": "The ID of the Kubernetes cluster"},
						"nodepool_id": {"type": "string", "description": "The ID of the node pool"}
					},
					"required": ["k8s_cluster_id", "nodepool_id"]
				}`),
				Annotations: api.ReadOnly("Get K8s Node Pool"),
			},
			Handler: getK8sNodepool,
		},
		{
			Tool: api.Tool{
				Name:        "list_k8s_nodes",
				Description: "List all nodes in a node pool",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"k8s_cluster_id": {"type": "string", "description": "The ID of the Kubernetes cluster"},
						"nodepool_id": {"type": "string", "description": "The ID of the node pool"}
					},
					"required": ["k8s_cluster_id", "nodepool_id"]
				}`),
				Annotations: api.ReadOnly("List K8s Nodes"),
			},
			Handler: listK8sNodes,
		},
		{
			Tool: api.Tool{
				Name:        "get_k8s_node",
				Description: "Get details of a specific node",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"k8s_cluster_id": {"type": "string", "description": "The ID of the Kubernetes cluster"},
						"nodepool_id": {"type": "string", "description": "The ID of the node pool"},
						"node_id": {"type": "string", "description": "The ID of the node"}
					},
					"required": ["k8s_cluster_id", "nodepool_id", "node_id"]
				}`),
				Annotations: api.ReadOnly("Get K8s Node"),
			},
			Handler: getK8sNode,
		},
	}
}

func listK8sNodepools(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	k8sClusterID, ok := api.GetRequiredString(params.Arguments, "k8s_cluster_id")
	if !ok {
		return nil, fmt.Errorf("k8s_cluster_id is required")
	}

	nodepools, _, err := params.Client.KubernetesApi.K8sNodepoolsGet(ctx, k8sClusterID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to list node pools: %w", err)
	}
	return api.MarshalResult(nodepools, "node pools")
}

func getK8sNodepool(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	k8sClusterID, ok := api.GetRequiredString(params.Arguments, "k8s_cluster_id")
	if !ok {
		return nil, fmt.Errorf("k8s_cluster_id is required")
	}
	nodepoolID, ok := api.GetRequiredString(params.Arguments, "nodepool_id")
	if !ok {
		return nil, fmt.Errorf("nodepool_id is required")
	}

	nodepool, _, err := params.Client.KubernetesApi.K8sNodepoolsFindById(ctx, k8sClusterID, nodepoolID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get node pool: %w", err)
	}
	return api.MarshalResult(nodepool, "node pool")
}

func listK8sNodes(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	k8sClusterID, ok := api.GetRequiredString(params.Arguments, "k8s_cluster_id")
	if !ok {
		return nil, fmt.Errorf("k8s_cluster_id is required")
	}
	nodepoolID, ok := api.GetRequiredString(params.Arguments, "nodepool_id")
	if !ok {
		return nil, fmt.Errorf("nodepool_id is required")
	}

	nodes, _, err := params.Client.KubernetesApi.K8sNodepoolsNodesGet(ctx, k8sClusterID, nodepoolID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}
	return api.MarshalResult(nodes, "nodes")
}

func getK8sNode(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	k8sClusterID, ok := api.GetRequiredString(params.Arguments, "k8s_cluster_id")
	if !ok {
		return nil, fmt.Errorf("k8s_cluster_id is required")
	}
	nodepoolID, ok := api.GetRequiredString(params.Arguments, "nodepool_id")
	if !ok {
		return nil, fmt.Errorf("nodepool_id is required")
	}
	nodeID, ok := api.GetRequiredString(params.Arguments, "node_id")
	if !ok {
		return nil, fmt.Errorf("node_id is required")
	}

	node, _, err := params.Client.KubernetesApi.K8sNodepoolsNodesFindById(ctx, k8sClusterID, nodepoolID, nodeID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}
	return api.MarshalResult(node, "node")
}
