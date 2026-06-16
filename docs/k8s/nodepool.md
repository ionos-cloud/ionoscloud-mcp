---
subcategory: "Managed Kubernetes"
page_title: "Node Pools"
description: |-
  Tools for listing and inspecting Kubernetes node pools in IONOS CLOUD.
---

# Node Pools

## list_k8s_nodepools

Lists all node pools belonging to a specific Kubernetes cluster.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `k8s_cluster_id` | string | Yes | The ID of the Kubernetes cluster |
| `depth` | integer | No | Nesting depth of returned objects (0–5, default `1`). |

**Example:**

```json
{
  "name": "list_k8s_nodepools",
  "arguments": {
    "k8s_cluster_id": "3f2e4b1c-8d9a-4f7e-b3c2-1a5d6e9f0b4c"
  }
}
```

**API Reference:** [k8sNodepoolsGet](https://api.ionos.com/docs/cloud/v6/#tag/Kubernetes/operation/k8sNodepoolsGet)

---

## get_k8s_nodepool

Gets detailed information about a specific Kubernetes node pool, including node count, hardware configuration, autoscaling settings, and labels.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `k8s_cluster_id` | string | Yes | The ID of the Kubernetes cluster |
| `nodepool_id` | string | Yes | The ID of the Kubernetes node pool |
| `depth` | integer | No | Nesting depth of returned objects (0–5). |

**Example:**

```json
{
  "name": "get_k8s_nodepool",
  "arguments": {
    "k8s_cluster_id": "3f2e4b1c-8d9a-4f7e-b3c2-1a5d6e9f0b4c",
    "nodepool_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  }
}
```

**API Reference:** [k8sNodepoolsFindById](https://api.ionos.com/docs/cloud/v6/#tag/Kubernetes/operation/k8sNodepoolsFindById)
