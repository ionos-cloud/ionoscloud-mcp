---
subcategory: "Managed Kubernetes"
page_title: "Clusters"
description: |-
  Tools for listing, inspecting, creating, updating and deleting Managed Kubernetes clusters in IONOS CLOUD.
---

# Clusters

A cluster is the managed Kubernetes control plane. It runs no workloads on its own —
create at least one [node pool](nodepool.md) for that.

The write tools (`create_k8s_cluster`, `update_k8s_cluster`, `delete_k8s_cluster`)
require `IONOS_MCP_TOOL_SCOPE` to include `write` (`destructive` for the delete).
Create and delete are two-phase: the first call previews and returns a one-time
`confirmation_token`, and only a second call carrying that token mutates anything.

## Asynchronous operations

**Every mutating Kubernetes endpoint answers `202 Accepted`**, so all three write tools
return before the change has taken effect. Poll `get_k8s_cluster` and read
`metadata.state`:

| State | Meaning |
|---|---|
| `ACTIVE`, `AVAILABLE` | done |
| `DEPLOYING`, `UPDATING`, `BUSY`, `MAINTENANCE` | still working |
| `DESTROYING` | delete in progress; the cluster then stops resolving |
| `FAILED`, `FAILED_UPDATING`, `FAILED_DESTROYING`, `FAILED_SUSPENDED`, `FAILED_HIBERNATING`, `FAILED_MAINTENANCE` | did not succeed |
| `SUSPENDED`, `HIBERNATING`, `INACTIVE`, `TERMINATED`, `UNKNOWN` | not running |

Leave at least 30 seconds between polls. Creating a cluster takes several minutes; a
`k8s_version` upgrade takes considerably longer.

While a resource is `BUSY` the API **queues** further modifications instead of
rejecting them, so do not resend a change just because nothing looks different yet —
you will stack up duplicate modifications.

## list_k8s_clusters

Lists all Managed Kubernetes clusters provisioned in your IONOS CLOUD account.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `depth` | integer | No | Nesting depth of returned objects (0–5, default `1`). Depth 1 includes names and basic properties. |

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
| `depth` | integer | No | Nesting depth of returned objects (0–5). |

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

## create_k8s_cluster

Creates one Managed Kubernetes cluster (control plane only). Requires `IONOS_MCP_TOOL_SCOPE` to include `write`.

Two-phase: call once without `confirmation_token` for a preview plus a one-time token, then again with the token and the same `name` and `location`.

Provisioning is asynchronous and takes several minutes. `location`, `nat_gateway_ip`, `node_subnet` and `public` are immutable, so decide public vs private up front.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `name` | string | Yes | Cluster name: ≤63 characters, beginning and ending with an alphanumeric character |
| `k8s_version` | string | No | Control-plane Kubernetes version (e.g. `1.31.2`). Defaults to the account default — see `get_k8s_default_version` |
| `maintenance_window` | object | No | `{ "day_of_the_week": "Sunday", "time": "02:00:00" }`. IONOS picks one if omitted |
| `public` | boolean | No | Whether the Kubernetes API server is reachable from the internet. Default `true`. `false` (private) also requires `location` and `nat_gateway_ip`. **Prerelease at IONOS**, along with the three fields below — expect it to be unavailable on some contracts |
| `location` | string | No | Location of a private cluster (e.g. `de/fra`). Mandatory when `public` is `false`. Immutable |
| `nat_gateway_ip` | string | No | NAT gateway IP of a private cluster; must already be reserved in the same location (`list_ip_blocks`). Mandatory when `public` is `false`. Immutable |
| `node_subnet` | string | No | Node subnet of a private cluster, a 16-bit IPv4 prefix (e.g. `10.0.0.0/16`). Immutable |
| `api_subnet_allow_list` | string[] | No | Restrict Kubernetes API server access to these IPs or CIDRs. **Omitting this leaves the API server open to any source address**, and the preview says so |
| `s3_buckets` | string[] | No | Name of an existing Object Storage bucket to receive Kubernetes API audit logs (at most one). Only the name is sent — IONOS writes the logs, so no Object Storage credentials are involved |
| `confirmation_token` | string | No | Omit on the first call; pass the token returned by the preview on the second |

**Example:**

```json
{
  "name": "create_k8s_cluster",
  "arguments": {
    "name": "prod",
    "k8s_version": "1.31.2",
    "api_subnet_allow_list": ["203.0.113.0/24"],
    "maintenance_window": { "day_of_the_week": "Sunday", "time": "02:00:00" }
  }
}
```

**API Reference:** [k8sPost](https://api.ionos.com/docs/cloud/v6/#tag/Kubernetes/operation/k8sPost)

---

## update_k8s_cluster

Updates a cluster's name, version, maintenance window, API server allow list or audit-log buckets. Requires `IONOS_MCP_TOOL_SCOPE` to include `write`. Single call — no confirmation token.

`location`, `nat_gateway_ip`, `node_subnet` and `public` are immutable and therefore not accepted.

> **PUT semantics.** The underlying endpoint replaces the cluster's properties rather than patching them, so this tool reads the cluster first and sends every omitted field back unchanged. Omitting a field keeps it; it does not clear it. This is what stops a rename from dropping `api_subnet_allow_list` and silently exposing the Kubernetes API server.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `k8s_cluster_id` | string | Yes | The ID of the cluster to update |
| `name` | string | No | New cluster name. Omit to keep the current one |
| `k8s_version` | string | No | Version to upgrade the control plane to. Only values from the cluster's `availableUpgradeVersions` are accepted, and the upgrade is not reversible |
| `maintenance_window` | object | No | New window, same shape as on create. Omit to keep the current one |
| `api_subnet_allow_list` | string[] | No | **Replaces** the allow list — include every entry that should remain. Omit to keep the current list; pass `[]` to remove the restriction entirely |
| `s3_buckets` | string[] | No | **Replaces** the audit-log buckets, by name. Omit to keep the current ones; pass `[]` to detach them |

**Example:**

```json
{
  "name": "update_k8s_cluster",
  "arguments": {
    "k8s_cluster_id": "3f2e4b1c-8d9a-4f7e-b3c2-1a5d6e9f0b4c",
    "k8s_version": "1.32.0"
  }
}
```

**API Reference:** [k8sPut](https://api.ionos.com/docs/cloud/v6/#tag/Kubernetes/operation/k8sPut)

---

## delete_k8s_cluster

Deletes a Managed Kubernetes cluster. Requires `IONOS_MCP_TOOL_SCOPE` to include `destructive`. **Irreversible.**

Two-phase: the first call returns a preview with a blast-radius summary (node pools and the worker nodes they are sized for) plus a one-time token; the second call carries the token and deletes.

The preview gives a count of the node pools and worker nodes the cluster would take with it — a summary, not a per-resource list. Delete the pools first with [`delete_k8s_nodepool`](nodepool.md#delete_k8s_nodepool) if you want to go step by step.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `k8s_cluster_id` | string | Yes | The ID of the cluster to delete |
| `confirmation_token` | string | No | Omit on the first call; pass the token returned by the preview on the second |

**Example:**

```json
{
  "name": "delete_k8s_cluster",
  "arguments": {
    "k8s_cluster_id": "3f2e4b1c-8d9a-4f7e-b3c2-1a5d6e9f0b4c"
  }
}
```

**API Reference:** [k8sDelete](https://api.ionos.com/docs/cloud/v6/#tag/Kubernetes/operation/k8sDelete)
