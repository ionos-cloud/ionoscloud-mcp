---
subcategory: "Managed Kubernetes"
page_title: "Node Pools"
description: |-
  Tools for listing, inspecting, creating, updating and deleting Kubernetes node pools in IONOS CLOUD.
---

# Node Pools

A node pool is a group of identically sized worker nodes in one data center. This is
where the cluster's workloads actually run.

The write tools (`create_k8s_nodepool`, `update_k8s_nodepool`, `delete_k8s_nodepool`)
require `IONOS_MCP_TOOL_SCOPE` to include `write` (`destructive` for the delete).
Create and delete are two-phase: the first call previews and returns a one-time
`confirmation_token`, and only a second call carrying that token mutates anything.

## Asynchronous operations

**Every mutating Kubernetes endpoint answers `202 Accepted`**, so all three write tools
return before the change has taken effect. Poll `get_k8s_nodepool` and read
`metadata.state` — the values are the same set as for clusters (see
[cluster.md](cluster.md#asynchronous-operations)); `ACTIVE` means done, `DEPLOYING` /
`UPDATING` / `BUSY` mean still working, any `FAILED_*` means it did not.

Leave at least 30 seconds between polls, and note that the work scales with the number
of nodes, because they are handled one at a time:

| Change | Roughly how long |
|---|---|
| Maintenance window, labels, annotations | quick, but still asynchronous |
| Create the pool | several minutes, proportional to `node_count` |
| Scale up or down | one node provisioned or drained at a time |
| `k8s_version` upgrade | **every node replaced, one at a time** — far longer than anything else here on a large pool |
| Delete the pool | several minutes |

While a resource is `BUSY` the API **queues** further modifications instead of
rejecting them, so do not resend a change just because nothing looks different yet.

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

---

## create_k8s_nodepool

Creates one node pool of worker nodes in a cluster. Requires `IONOS_MCP_TOOL_SCOPE` to include `write`.

Two-phase: call once without `confirmation_token` for a preview plus a one-time token, then again with the token and the same `k8s_cluster_id`, `name` and `datacenter_id`.

The cluster must already be `ACTIVE`, and `datacenter_id` must name a data center in the same location as the cluster. Provisioning is asynchronous and takes several minutes.

> **The per-node hardware is immutable.** `cores_count`, `ram_size`, `storage_type`, `storage_size`, `availability_zone`, `cpu_family` and `datacenter_id` cannot be changed afterwards — resizing means deleting the pool and creating a new one. Size it deliberately; the preview labels these fields as immutable.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `k8s_cluster_id` | string | Yes | The ID of the cluster the node pool belongs to |
| `name` | string | Yes | Node pool name: ≤63 characters, beginning and ending with an alphanumeric character |
| `datacenter_id` | string | Yes | Data center hosting the worker nodes; same location as the cluster. Immutable |
| `node_count` | integer | Yes | Number of worker nodes. With `auto_scaling`, the starting count, which must fall within the bounds |
| `cores_count` | integer | Yes | CPU cores per node. Immutable |
| `ram_size` | integer | Yes | RAM per node in MB: a multiple of 1024, at least 2048. Immutable |
| `availability_zone` | string | Yes | `AUTO`, `ZONE_1` or `ZONE_2`. Immutable |
| `storage_type` | string | Yes | `HDD` or `SSD`. Immutable |
| `storage_size` | integer | Yes | Volume size per node in GB (>100 GB recommended for SSD). Immutable |
| `cpu_family` | string | No | **Deprecated by IONOS — use `server_type` instead.** e.g. `INTEL_ICELAKE`; IONOS picks one available at the location if omitted. An empty string is rejected. Immutable |
| `server_type` | string | No | `DedicatedCore` (default) or `VCPU` |
| `k8s_version` | string | No | Worker node version; defaults to the cluster's. Must be one of the cluster's `viableNodePoolVersions` |
| `maintenance_window` | object | No | `{ "day_of_the_week": "Saturday", "time": "03:00:00" }`. Node maintenance replaces nodes one at a time |
| `auto_scaling` | object | No | `{ "min_node_count": 1, "max_node_count": 5 }`. Omit for a fixed-size pool |
| `lans` | object[] | No | Existing private LANs to attach: `{ "id": 3, "dhcp": true, "routes": [{ "network": "10.0.0.0/24", "gateway_ip": "10.0.0.1" }] }`. Inside a route, `network` and `gateway_ip` are each optional, but an entry with neither is rejected |
| `labels` | object | No | Kubernetes labels on every node, as key-value pairs |
| `annotations` | object | No | Kubernetes annotations on every node, as key-value pairs |
| `public_ips` | string[] | No | Reserved public IPs (`list_ip_blocks`), all from the pool's data center location. Needs **one more** than the maximum node count (`node_count`+1, or `max_node_count`+1 with autoscaling) — the spare covers a node being rebuilt |
| `confirmation_token` | string | No | Omit on the first call; pass the token returned by the preview on the second |

**Example:**

```json
{
  "name": "create_k8s_nodepool",
  "arguments": {
    "k8s_cluster_id": "3f2e4b1c-8d9a-4f7e-b3c2-1a5d6e9f0b4c",
    "name": "workers",
    "datacenter_id": "8e1f9c2d-3b4a-5c6d-7e8f-9a0b1c2d3e4f",
    "node_count": 2,
    "cores_count": 4,
    "ram_size": 4096,
    "availability_zone": "AUTO",
    "storage_type": "SSD",
    "storage_size": 100,
    "auto_scaling": { "min_node_count": 2, "max_node_count": 5 }
  }
}
```

**API Reference:** [k8sNodepoolsPost](https://api.ionos.com/docs/cloud/v6/#tag/Kubernetes/operation/k8sNodepoolsPost)

---

## update_k8s_nodepool

Scales, upgrades or reconfigures a node pool. Requires `IONOS_MCP_TOOL_SCOPE` to include `write`. Single call — no confirmation token.

The per-node hardware and `datacenter_id` are immutable and therefore not accepted; delete and recreate the pool to change them.

> **PUT semantics.** The underlying endpoint replaces the node pool's properties rather than patching them, so this tool reads the pool first and sends every omitted field back unchanged. This matters most for `node_count`: the SDK always serializes it, so without carry-forward a small unrelated change would send `nodeCount: 0` and drain the pool. Carried-forward LANs, labels and annotations likewise keep an unrelated change from detaching worker networking or losing scheduling metadata.
>
> Node **taints** are carried forward too, even though there is no `taints` parameter: IONOS marks the field internal in the API spec, so it is not exposed, but a pool may carry taints applied out of band and a replacing PUT would otherwise drop them.
>
> Two fields are deliberately **never** sent back, because a GET response is not always a legal PUT body: the pool **`name`** (the API rejects it as immutable) and an **inactive autoscaler** (a pool without one reads back as `{minNodeCount: 0, maxNodeCount: 0}`, which the API rejects on a write with `autoScaling.minNodeCount must be > 0`).

> **The node pool name cannot be changed.** There is no `name` parameter here — the API rejects it as immutable. Note the asymmetry: a *cluster* can be renamed with [`update_k8s_cluster`](cluster.md#update_k8s_cluster), a node pool cannot. Recreate the pool under a new name, or leave it and rely on `nodepool_id`.

> **An autoscaler cannot be switched off.** Its bounds can be changed, but there is no request that removes one from an existing pool — the API answers `422` for zero bounds and silently ignores an omitted `autoScaling`, both verified against a live account. Passing zero bounds is therefore rejected with an explanation rather than accepted as a no-op. To end up with a fixed-size pool, delete this one and create a new pool without `auto_scaling`. This is an IONOS API limitation; `ionosctl` has the same gap.

Call `get_k8s_nodepool` first if you intend to supply any of the list-valued fields — each **replaces** the current value rather than adding to it.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `k8s_cluster_id` | string | Yes | The ID of the cluster the node pool belongs to |
| `nodepool_id` | string | Yes | The ID of the node pool to update |
| `node_count` | integer | No | New worker node count. Scaling down evicts whatever runs on the removed nodes. Ignored while autoscaling is active |
| `server_type` | string | No | `DedicatedCore` or `VCPU`. Omit to keep the current one |
| `k8s_version` | string | No | Version to upgrade the nodes to. Only values from the pool's `availableUpgradeVersions` are accepted; the upgrade replaces every node and is not reversible |
| `maintenance_window` | object | No | New window. Omit to keep the current one |
| `auto_scaling` | object | No | New bounds, both at least 1. Omit to keep the current setting. **An existing autoscaler cannot be removed** — see below |
| `lans` | object[] | No | **Replaces** the attached LANs. Omit to keep them; pass `[]` to detach them all |
| `labels` | object | No | **Replaces** the node labels. Omit to keep them; pass `{}` to remove them all |
| `annotations` | object | No | **Replaces** the node annotations. Omit to keep them; pass `{}` to remove them all |
| `public_ips` | string[] | No | **Replaces** the reserved public IPs; still needs one more than the maximum node count. Omit to keep them; pass `[]` to remove them all |

**Example:**

```json
{
  "name": "update_k8s_nodepool",
  "arguments": {
    "k8s_cluster_id": "3f2e4b1c-8d9a-4f7e-b3c2-1a5d6e9f0b4c",
    "nodepool_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "node_count": 5
  }
}
```

**API Reference:** [k8sNodepoolsPut](https://api.ionos.com/docs/cloud/v6/#tag/Kubernetes/operation/k8sNodepoolsPut)

---

## delete_k8s_nodepool

Deletes a node pool and all of its worker nodes. Requires `IONOS_MCP_TOOL_SCOPE` to include `destructive`. **Irreversible.**

Two-phase: the first call returns a preview with the number of worker nodes it would destroy plus a one-time token; the second call carries the token and deletes.

Every pod on those nodes is evicted, and anything on the nodes' own volumes is lost. If this is the cluster's last node pool, the cluster keeps running but has nowhere to schedule workloads.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `k8s_cluster_id` | string | Yes | The ID of the cluster the node pool belongs to |
| `nodepool_id` | string | Yes | The ID of the node pool to delete |
| `confirmation_token` | string | No | Omit on the first call; pass the token returned by the preview on the second |

**Example:**

```json
{
  "name": "delete_k8s_nodepool",
  "arguments": {
    "k8s_cluster_id": "3f2e4b1c-8d9a-4f7e-b3c2-1a5d6e9f0b4c",
    "nodepool_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  }
}
```

**API Reference:** [k8sNodepoolsDelete](https://api.ionos.com/docs/cloud/v6/#tag/Kubernetes/operation/k8sNodepoolsDelete)
