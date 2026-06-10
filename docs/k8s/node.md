---
subcategory: "Managed Kubernetes"
page_title: "Nodes"
description: |-
  Tools for listing and inspecting Kubernetes nodes in IONOS CLOUD.
---

# Nodes

## list_k8s_nodepool_nodes

Lists all individual nodes in a Kubernetes node pool, including their state and public/private IP addresses.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `k8s_cluster_id` | string | Yes | The ID of the Kubernetes cluster |
| `nodepool_id` | string | Yes | The ID of the Kubernetes node pool |

**Example:**

```json
{
  "name": "list_k8s_nodepool_nodes",
  "arguments": {
    "k8s_cluster_id": "3f2e4b1c-8d9a-4f7e-b3c2-1a5d6e9f0b4c",
    "nodepool_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  }
}
```

**API Reference:** [k8sNodepoolsNodesGet](https://api.ionos.com/docs/cloud/v6/#tag/Kubernetes/operation/k8sNodepoolsNodesGet)

---

## get_k8s_node

Gets detailed information about a specific node in a Kubernetes node pool.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `k8s_cluster_id` | string | Yes | The ID of the Kubernetes cluster |
| `nodepool_id` | string | Yes | The ID of the Kubernetes node pool |
| `node_id` | string | Yes | The ID of the Kubernetes node |

**Example:**

```json
{
  "name": "get_k8s_node",
  "arguments": {
    "k8s_cluster_id": "3f2e4b1c-8d9a-4f7e-b3c2-1a5d6e9f0b4c",
    "nodepool_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "node_id": "f9e8d7c6-b5a4-3210-fedc-ba9876543210"
  }
}
```

**API Reference:** [k8sNodepoolsNodesFindById](https://api.ionos.com/docs/cloud/v6/#tag/Kubernetes/operation/k8sNodepoolsNodesFindById)
