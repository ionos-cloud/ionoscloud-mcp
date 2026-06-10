---
subcategory: "Managed Kubernetes"
page_title: "Kubernetes"
description: |-
  Tools for listing and inspecting Kubernetes clusters, node pools, nodes, and available versions in IONOS CLOUD.
---

# Kubernetes

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

---

## list_k8s_nodepools

Lists all node pools belonging to a specific Kubernetes cluster.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `k8s_cluster_id` | string | Yes | The ID of the Kubernetes cluster |

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

---

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

---

## list_k8s_versions

Lists all Kubernetes versions currently available for cluster and node pool creation in IONOS CLOUD.

**Parameters:** None

**Example:**

```json
{
  "name": "list_k8s_versions",
  "arguments": {}
}
```

**API Reference:** [k8sVersionsGet](https://api.ionos.com/docs/cloud/v6/#tag/Kubernetes/operation/k8sVersionsGet)

---

## get_k8s_default_version

Returns the current default Kubernetes version that IONOS CLOUD applies to new clusters and node pools when no version is explicitly specified.

**Parameters:** None

**Example:**

```json
{
  "name": "get_k8s_default_version",
  "arguments": {}
}
```

**API Reference:** [k8sVersionsDefaultGet](https://api.ionos.com/docs/cloud/v6/#tag/Kubernetes/operation/k8sVersionsDefaultGet)
