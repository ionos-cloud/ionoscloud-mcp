---
subcategory: "Managed Kubernetes"
page_title: "Versions"
description: |-
  Tools for listing available Kubernetes versions in IONOS CLOUD.
---

# Versions

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
