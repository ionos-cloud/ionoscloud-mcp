package k8s

import (
	"context"
	"fmt"
	"strings"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

// RegisterNodeWriteTools registers the recreate/delete node tools.
func RegisterNodeWriteTools(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	registerRecreateNode(server, client, scope, confirm)
	registerDeleteNode(server, client, scope, confirm)
}

func registerRecreateNode(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterActionTool(server, scope,
		// A POST, but destructive by verb: the old node is thrown away. Not idempotent.
		tools.Action{Verb: "recreate_", Method: tools.MethodPost, Idempotent: false},
		&mcp.Tool{
			Name: "recreate_k8s_node",
			Description: "Recreate one worker node. Two-phase: call first WITHOUT confirmation_token for a preview and a one-time token, then again WITH the token. " +
				"The replacement is provisioned and joined to the cluster before the old node drains, so capacity never dips, at the cost of one extra billable node meanwhile. " +
				"Prefer this over delete_k8s_node to replace a node." + asyncNodeNote,
		},
		func(ctx context.Context, req *mcp.CallToolRequest, input tools.K8sNodeActionInput) (*mcp.CallToolResult, any, error) {
			clusterID, poolID, nodeID, errMsg := nodeIDs(input)
			if errMsg != "" {
				return tools.ErrorText(errMsg), nil, nil
			}
			target := tools.Target(req, clusterID, poolID, nodeID)

			if tools.HasToken(input.ConfirmationToken) {
				if err := confirm.Consume(*input.ConfirmationToken, "recreate_k8s_node", target); err != nil {
					return tools.ErrorText(tools.ConfirmErrorText("recreate_k8s_node", "k8s_cluster_id, nodepool_id and node_id", err)), nil, nil
				}
				_, err := client.KubernetesApi.K8sNodepoolsNodesReplacePost(ctx, clusterID, poolID, nodeID).Execute()
				if err != nil {
					return tools.ToResult(nil, err)
				}
				return tools.TextResult(fmt.Sprintf("Requested recreation of node %s. Follow progress with list_k8s_nodepool_nodes.", nodeID)), nil, nil
			}

			node, errRes := findNode(ctx, client, clusterID, poolID, nodeID)
			if errRes != nil {
				return errRes, nil, nil
			}
			token, mErr := confirm.Mint("recreate_k8s_node", target)
			if mErr != nil {
				return nil, nil, mErr
			}
			return tools.TextResult(tools.Preview{
				Headline: "About to RECREATE a Kubernetes worker node.\n" +
					"The replacement joins the cluster first, then this node drains: its pods are evicted and local storage is lost. The pool size does not change.",
				Fields:    nodeFields(clusterID, poolID, nodeID, *node),
				Tool:      "recreate_k8s_node",
				Replay:    tools.Fields("k8s_cluster_id", clusterID, "nodepool_id", poolID, "node_id", nodeID),
				TokenNote: "This token authorizes recreating ONLY this node",
			}.Render(token)), nil, nil
		})
}

func registerDeleteNode(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterTool(server, scope, tools.MethodDelete, &mcp.Tool{
		Name: "delete_k8s_node",
		Description: "Delete one worker node. Two-phase: call first WITHOUT confirmation_token for a preview and a one-time token, then again WITH the token. This is irreversible — pods are evicted and local storage is lost. " +
			"The node goes first, leaving the pool short, and an active autoscaler may keep it that way (observed 2 nodes to 1). Use recreate_k8s_node to replace a node, or update_k8s_nodepool to resize. " +
			"The API also refuses this when it would take the pool below its autoscaler minimum." + asyncNodeNote,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.K8sNodeActionInput) (*mcp.CallToolResult, any, error) {
		clusterID, poolID, nodeID, errMsg := nodeIDs(input)
		if errMsg != "" {
			return tools.ErrorText(errMsg), nil, nil
		}
		target := tools.Target(req, clusterID, poolID, nodeID)

		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "delete_k8s_node", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("delete_k8s_node", "k8s_cluster_id, nodepool_id and node_id", err)), nil, nil
			}
			_, err := client.KubernetesApi.K8sNodepoolsNodesDelete(ctx, clusterID, poolID, nodeID).Execute()
			if err != nil {
				return tools.ToResult(nil, err)
			}
			return tools.TextResult(tools.DeletedAsync("Kubernetes node", nodeID) +
				" The pool is now one node short; watch list_k8s_nodepool_nodes to see whether it backfills."), nil, nil
		}

		node, errRes := findNode(ctx, client, clusterID, poolID, nodeID)
		if errRes != nil {
			return errRes, nil, nil
		}
		autoscalerActive, blockMsg := inspectNodeDelete(ctx, client, clusterID, poolID)
		if blockMsg != "" {
			return tools.ErrorText(blockMsg), nil, nil
		}
		token, mErr := confirm.Mint("delete_k8s_node", target)
		if mErr != nil {
			return nil, nil, mErr
		}
		headline := "About to DELETE a Kubernetes worker node. This is IRREVERSIBLE.\n" +
			"Its pods are evicted and local storage is lost. The node goes first, so the pool runs one node short; recreate_k8s_node avoids that gap."
		if autoscalerActive {
			headline += "\nWARNING: an ACTIVE autoscaler owns this pool's node count, so the pool may STAY at the reduced size. Resize with update_k8s_nodepool instead."
		}
		return tools.TextResult(tools.Preview{
			Headline:  headline,
			Fields:    nodeFields(clusterID, poolID, nodeID, *node),
			Tool:      "delete_k8s_node",
			Replay:    tools.Fields("k8s_cluster_id", clusterID, "nodepool_id", poolID, "node_id", nodeID),
			TokenNote: "This token authorizes deleting ONLY this node",
		}.Render(token)), nil, nil
	})
}

func nodeIDs(input tools.K8sNodeActionInput) (clusterID, poolID, nodeID, errMsg string) {
	clusterID = strings.TrimSpace(input.K8sClusterID)
	poolID = strings.TrimSpace(input.NodepoolID)
	nodeID = strings.TrimSpace(input.NodeID)
	switch {
	case clusterID == "":
		return "", "", "", "k8s_cluster_id is required"
	case poolID == "":
		return "", "", "", "nodepool_id is required"
	case nodeID == "":
		return "", "", "", "node_id is required"
	}
	return clusterID, poolID, nodeID, ""
}

// inspectNodeDelete reports whether an autoscaler owns the pool's count, and why the
// API would refuse the delete. The API's own refusal arrives only after a token is
// spent and reads "last node can not be deleted from nodepool" even with nodes left;
// the real rule is that the pool may not drop below its autoscaler minimum. A failed
// read is not an error — the API stays the authority.
func inspectNodeDelete(ctx context.Context, client *ionos.APIClient, clusterID, poolID string) (autoscalerActive bool, blockMsg string) {
	pool, _, err := client.KubernetesApi.K8sNodepoolsFindById(ctx, clusterID, poolID).Depth(1).Execute()
	if err != nil {
		return false, ""
	}
	cp := pool.GetProperties()
	count := cp.GetNodeCount()
	if count <= 1 {
		return false, fmt.Sprintf("node pool %s has only %d node and the API refuses to delete a pool's last node; use delete_k8s_nodepool or recreate_k8s_node", poolID, count)
	}
	if !autoScalingActive(cp.AutoScaling) {
		return false, ""
	}
	if min := cp.AutoScaling.GetMinNodeCount(); count-1 < min {
		return true, fmt.Sprintf("deleting a node would leave %d in node pool %s, below its autoscaler minimum of %d, which the API refuses (reported as \"last node can not be deleted\"). Lower auto_scaling.min_node_count first, or use recreate_k8s_node.", count-1, poolID, min)
	}
	return true, ""
}

func findNode(ctx context.Context, client *ionos.APIClient, clusterID, poolID, nodeID string) (*ionos.KubernetesNode, *mcp.CallToolResult) {
	node, _, err := client.KubernetesApi.K8sNodepoolsNodesFindById(ctx, clusterID, poolID, nodeID).Depth(1).Execute()
	if err != nil {
		if tools.IsNotFound(err) {
			return nil, tools.ErrorText(fmt.Sprintf("node %s does not exist in node pool %s; list them with list_k8s_nodepool_nodes", nodeID, poolID))
		}
		res, _, _ := tools.ToResult(nil, err)
		return nil, res
	}
	return &node, nil
}

func nodeFields(clusterID, poolID, nodeID string, node ionos.KubernetesNode) []tools.KV {
	props := node.GetProperties()
	state := ""
	if node.Metadata != nil {
		state = node.Metadata.GetState()
	}
	return tools.Fields(
		"k8s_cluster_id", clusterID,
		"nodepool_id", poolID,
		"node_id", nodeID,
		"name", props.GetName(),
		"state", state,
		"k8s_version", props.GetK8sVersion(),
		"private_ip", props.GetPrivateIP(),
		"public_ip", props.GetPublicIP(),
	)
}
