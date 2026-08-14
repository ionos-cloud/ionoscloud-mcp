---
subcategory: "Managed Kubernetes"
page_title: "Nodes"
description: |-
  Tools for listing, inspecting, recreating and deleting individual Kubernetes worker nodes in IONOS CLOUD.
---

# Nodes

These tools act on one worker node at a time. They differ in **ordering**, and that
difference is the whole reason to prefer one over the other:

| | Sequence | Capacity | Pool size afterwards | Cost |
|---|---|---|---|---|
| `recreate_k8s_node` | replacement joins the cluster, **then** the old node drains | never dips | unchanged | one extra billable node while both run |
| `delete_k8s_node` | node goes **first**, backfill may follow | **runs one node short** | **not guaranteed to recover** — see below | no extra node |

> **`delete_k8s_node` is not a reliable way to replace a node.** Whether the pool
> backfills depends on the pool. Without an autoscaler, `node_count` is the desired size
> and the pool should return to it. With an **active autoscaler**, the autoscaler owns
> the count — `node_count` is no longer authoritative — and it can legitimately hold the
> pool at the smaller size. Observed live: a 2-node pool with `min 1` went to 1 node and
> stayed there.
>
> Use `recreate_k8s_node` to replace a node, and
> [`update_k8s_nodepool`](nodepool.md#update_k8s_nodepool) to change the size on
> purpose. The `delete_k8s_node` preview warns when an autoscaler is active.

Both `recreate_k8s_node` and `delete_k8s_node` require `IONOS_MCP_TOOL_SCOPE` to
include `destructive`, and both are two-phase: the first call previews the node and
returns a one-time `confirmation_token`, and only a second call carrying that token
mutates anything.

## Asynchronous operations

Both endpoints answer `202 Accepted`, so the tool returns before the node has actually
gone. Follow progress with `list_k8s_nodepool_nodes` and read each node's
`metadata.state`. **Node state is its own, smaller enum** — not the cluster/node pool
one:

| State | Meaning |
|---|---|
| `READY` | node is in service |
| `PROVISIONING`, `PROVISIONED` | replacement coming up |
| `REBUILDING` | being recreated |
| `TERMINATING` | being removed |
| `BUSY` | a modification is pending |

Leave at least 30 seconds between polls. A recreate takes several minutes, because the
replacement must be provisioned and join the cluster before the old node is drained.

## list_k8s_nodepool_nodes

Lists all individual nodes in a Kubernetes node pool, including their state and public/private IP addresses.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `k8s_cluster_id` | string | Yes | The ID of the Kubernetes cluster |
| `nodepool_id` | string | Yes | The ID of the Kubernetes node pool |
| `depth` | integer | No | Nesting depth of returned objects (0–5, default `1`). |

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
| `depth` | integer | No | Nesting depth of returned objects (0–5). |

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

## recreate_k8s_node

Recreates one worker node. Requires `IONOS_MCP_TOOL_SCOPE` to include `destructive`.

IONOS provisions a replacement node, waits for it to register with the cluster, and only then drains and destroys the old one — so capacity is preserved throughout. Pods on the old node are evicted and anything on its local storage is lost.

Use this for a node that is unhealthy or stuck. Prefer it over `delete_k8s_node`, which removes the node before the replacement exists.

> **Cost:** while the replacement runs alongside the old node, the node pool carries an extra billable active node.

The tool is a `POST` but is classified **destructive**, because the existing node is thrown away. It is not idempotent: a second call recreates the replacement in turn.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `k8s_cluster_id` | string | Yes | The ID of the cluster the node belongs to |
| `nodepool_id` | string | Yes | The ID of the node pool the node belongs to |
| `node_id` | string | Yes | The ID of the node to recreate |
| `confirmation_token` | string | No | Omit on the first call; pass the token returned by the preview on the second |

**Example:**

```json
{
  "name": "recreate_k8s_node",
  "arguments": {
    "k8s_cluster_id": "3f2e4b1c-8d9a-4f7e-b3c2-1a5d6e9f0b4c",
    "nodepool_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "node_id": "f9e8d7c6-b5a4-3210-fedc-ba9876543210"
  }
}
```

**API Reference:** [k8sNodepoolsNodesReplacePost](https://api.ionos.com/docs/cloud/v6/#tag/Kubernetes/operation/k8sNodepoolsNodesReplacePost)

---

## delete_k8s_node

Deletes one worker node from a node pool. Requires `IONOS_MCP_TOOL_SCOPE` to include `destructive`. **Irreversible.**

The node is drained and destroyed, its pods are evicted, and data on its local storage is lost.

> **The pool is left a node short, and may stay that way.** The node is removed first, and whether a replacement follows depends on whether an autoscaler is active — see the table at the top of this page. To change a pool's size on purpose, use [`update_k8s_nodepool`](nodepool.md#update_k8s_nodepool); to replace a node, use `recreate_k8s_node`.

> **Blocked by the autoscaler minimum.** The API refuses to remove a node when doing so would take the pool below `auto_scaling.min_node_count`, and reports it as `[VDC-14-1826] Operation failed, as last node can not be deleted from nodepool` — even when several nodes remain. A pool of 2 pinned at `min 2` cannot have a node deleted at all. The tool reads the pool during the preview and says this plainly rather than letting you spend a confirmation token on a request the API will refuse. Lower `min_node_count` first, or use `recreate_k8s_node`, which is never blocked because it adds the replacement before removing the old node.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `k8s_cluster_id` | string | Yes | The ID of the cluster the node belongs to |
| `nodepool_id` | string | Yes | The ID of the node pool the node belongs to |
| `node_id` | string | Yes | The ID of the node to delete |
| `confirmation_token` | string | No | Omit on the first call; pass the token returned by the preview on the second |

**Example:**

```json
{
  "name": "delete_k8s_node",
  "arguments": {
    "k8s_cluster_id": "3f2e4b1c-8d9a-4f7e-b3c2-1a5d6e9f0b4c",
    "nodepool_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "node_id": "f9e8d7c6-b5a4-3210-fedc-ba9876543210"
  }
}
```

**API Reference:** [k8sNodepoolsNodesDelete](https://api.ionos.com/docs/cloud/v6/#tag/Kubernetes/operation/k8sNodepoolsNodesDelete)
