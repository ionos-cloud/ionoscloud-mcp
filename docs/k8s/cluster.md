---
subcategory: "Managed Kubernetes"
page_title: "Clusters"
description: |-
  Tools for listing and inspecting Managed Kubernetes clusters in IONOS CLOUD.
---

# Clusters

## list_k8s_clusters

Lists all Managed Kubernetes clusters provisioned in your IONOS CLOUD account.

**Parameters:** None

**Example:**

```json
{
  "name": "list_k8s_clusters",
  "arguments": {}
}
```

**API Reference:** [k8sGet](https://api.ionos.com/docs/cloud/v6/#tag/Kubernetes/operation/k8sGet)

---

## get_k8s_cluster

Gets detailed information about a specific Kubernetes cluster, including its state, version, maintenance window, and API server endpoint.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `k8s_cluster_id` | string | Yes | The ID of the Kubernetes cluster |

**Example:**

```json
{
  "name": "get_k8s_cluster",
  "arguments": {
    "k8s_cluster_id": "3f2e4b1c-8d9a-4f7e-b3c2-1a5d6e9f0b4c"
  }
}
```

**API Reference:** [k8sFindByClusterId](https://api.ionos.com/docs/cloud/v6/#tag/Kubernetes/operation/k8sFindByClusterId)
