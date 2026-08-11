package k8s

import (
	"context"
	"fmt"
	"sort"
	"strings"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

// A node pool is a group of identically sized worker nodes in one data center. The
// cluster's workloads run here, so a cluster without a node pool runs nothing.

// RegisterNodepoolWriteTools registers the create/update/delete node pool tools.
func RegisterNodepoolWriteTools(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	registerCreateNodepool(server, client, scope, confirm)
	registerUpdateNodepool(server, client, scope)
	registerDeleteNodepool(server, client, scope, confirm)
}

func registerCreateNodepool(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterTool(server, scope, tools.MethodPost, &mcp.Tool{
		Name: "create_k8s_nodepool",
		Description: "Create one node pool of worker nodes. Two-phase: call first WITHOUT confirmation_token to get a preview and a one-time token, then call again WITH the token (and the same k8s_cluster_id, name and datacenter_id) to create it. Creates exactly one node pool per call. " +
			"The cluster must already be ACTIVE, and datacenter_id must be in its location. The per-node hardware and datacenter_id are immutable — recreate the pool to change them." + asyncResourceNote,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.CreateK8sNodepoolInput) (*mcp.CallToolResult, any, error) {
		clusterID := strings.TrimSpace(input.K8sClusterID)
		name := strings.TrimSpace(input.Name)
		dcID := strings.TrimSpace(input.DatacenterID)
		if clusterID == "" {
			return tools.ErrorText("k8s_cluster_id is required"), nil, nil
		}
		if name == "" {
			return tools.ErrorText("name is required to create a node pool"), nil, nil
		}
		if dcID == "" {
			return tools.ErrorText("datacenter_id is required: the node pool's worker nodes live in a data center in the same location as the cluster"), nil, nil
		}
		if input.NodeCount < 1 {
			return tools.ErrorText(fmt.Sprintf("node_count must be at least 1, got %d", input.NodeCount)), nil, nil
		}
		if msg := validateNodeHardware(input.CoresCount, input.RamSize, input.StorageSize); msg != "" {
			return tools.ErrorText(msg), nil, nil
		}
		storageType, msg := normalizeStorageType(input.StorageType)
		if msg != "" {
			return tools.ErrorText(msg), nil, nil
		}
		zone, msg := normalizeAvailabilityZone(input.AvailabilityZone)
		if msg != "" {
			return tools.ErrorText(msg), nil, nil
		}
		var serverType ionos.KubernetesNodePoolServerType
		if input.ServerType != nil {
			if serverType, msg = normalizeServerType(*input.ServerType); msg != "" {
				return tools.ErrorText(msg), nil, nil
			}
		}
		window, msg := buildMaintenanceWindow(input.MaintenanceWindow)
		if msg != "" {
			return tools.ErrorText(msg), nil, nil
		}
		auto, msg := buildAutoScaling(input.AutoScaling)
		if msg != "" {
			return tools.ErrorText(msg), nil, nil
		}
		if msg := validateNodeCountWithinAutoScaling(auto, input.NodeCount); msg != "" {
			return tools.ErrorText(msg), nil, nil
		}
		lans, msg := buildLans(input.Lans)
		if msg != "" {
			return tools.ErrorText(msg), nil, nil
		}
		if msg := validateCpuFamily(input.CpuFamily); msg != "" {
			return tools.ErrorText(msg), nil, nil
		}
		if msg := validatePublicIps(input.PublicIps, input.NodeCount, auto); msg != "" {
			return tools.ErrorText(msg), nil, nil
		}
		target := tools.Target(req, clusterID, name, dcID)

		// Phase 2: token present -> validate and execute.
		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "create_k8s_nodepool", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("create_k8s_nodepool", "k8s_cluster_id, name and datacenter_id", err)), nil, nil
			}
			props := ionos.NewKubernetesNodePoolPropertiesForPost(
				name, dcID, input.NodeCount, input.CoresCount, input.RamSize, zone, storageType, input.StorageSize,
			)
			if input.CpuFamily != nil {
				props.SetCpuFamily(strings.TrimSpace(*input.CpuFamily))
			}
			if serverType != "" {
				props.SetServerType(serverType)
			}
			if input.K8sVersion != nil {
				props.SetK8sVersion(strings.TrimSpace(*input.K8sVersion))
			}
			if window != nil {
				props.SetMaintenanceWindow(*window)
			}
			if auto != nil {
				props.SetAutoScaling(*auto)
			}
			if len(lans) > 0 {
				props.SetLans(lans)
			}
			if len(input.Labels) > 0 {
				props.SetLabels(input.Labels)
			}
			if len(input.Annotations) > 0 {
				props.SetAnnotations(input.Annotations)
			}
			if len(input.PublicIps) > 0 {
				props.SetPublicIps(input.PublicIps)
			}
			body := ionos.NewKubernetesNodePoolForPost(*props)
			created, _, err := client.KubernetesApi.K8sNodepoolsPost(ctx, clusterID).KubernetesNodePool(*body).Execute()
			return tools.ToResult(created, err)
		}

		// Phase 1: no token -> preview and mint a one-time token.
		token, err := confirm.Mint("create_k8s_nodepool", target)
		if err != nil {
			return nil, nil, err
		}
		return tools.TextResult(tools.Preview{
			Headline: "About to CREATE one Kubernetes node pool. The per-node hardware below is immutable afterwards:",
			Fields: tools.Fields(
				"k8s_cluster_id", clusterID,
				"name", name,
				"datacenter_id", dcID,
				"node_count", fmt.Sprintf("%d", input.NodeCount),
				"cores_count", fmt.Sprintf("%d", input.CoresCount),
				"ram_size", fmt.Sprintf("%d MB", input.RamSize),
				"storage", fmt.Sprintf("%d GB %s", input.StorageSize, storageType),
				"availability_zone", zone,
				"server_type", string(serverType),
				"cpu_family", strings.TrimSpace(tools.OptStr(input.CpuFamily)),
				"k8s_version", strings.TrimSpace(tools.OptStr(input.K8sVersion)),
				"maintenance_window", maintenanceWindowText(window),
				"auto_scaling", autoScalingText(auto),
				"lans", lansText(lans),
				"labels", mapText(input.Labels),
				"annotations", mapText(input.Annotations),
				"public_ips", strings.Join(input.PublicIps, ", "),
			),
			Tool:      "create_k8s_nodepool",
			Replay:    tools.Fields("k8s_cluster_id", clusterID, "name", name, "datacenter_id", dcID),
			TokenNote: "This creates exactly one node pool. The token authorizes creating only this cluster+name+datacenter",
		}.Render(token)), nil, nil
	})
}

func registerUpdateNodepool(server *mcp.Server, client *ionos.APIClient, scope tools.Scope) {
	tools.RegisterTool(server, scope, tools.MethodPut, &mcp.Tool{
		Name: "update_k8s_nodepool",
		Description: "Update a node pool: scale it, upgrade it, or change its maintenance window, autoscaling, LANs, labels, annotations or public IPs. The pool name and the per-node hardware are immutable. " +
			"This endpoint replaces the pool's properties, so fields you omit are read and sent back unchanged. lans, labels, annotations and public_ips replace the current value when supplied — read get_k8s_nodepool first. " +
			"An autoscaler's bounds can be changed but it cannot be removed. Scaling down evicts what runs on the removed nodes, and a k8s_version change replaces every node one at a time." + asyncResourceNote,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.UpdateK8sNodepoolInput) (*mcp.CallToolResult, any, error) {
		clusterID := strings.TrimSpace(input.K8sClusterID)
		poolID := strings.TrimSpace(input.NodepoolID)
		if clusterID == "" {
			return tools.ErrorText("k8s_cluster_id is required"), nil, nil
		}
		if poolID == "" {
			return tools.ErrorText("nodepool_id is required"), nil, nil
		}
		if input.NodeCount == nil && input.ServerType == nil && input.K8sVersion == nil &&
			input.MaintenanceWindow == nil && input.AutoScaling == nil && input.Lans == nil &&
			input.Labels == nil && input.Annotations == nil && input.PublicIps == nil {
			return tools.ErrorText("nothing to update: provide at least one of node_count, server_type, k8s_version, maintenance_window, auto_scaling, lans, labels, annotations, public_ips"), nil, nil
		}
		if input.NodeCount != nil && *input.NodeCount < 1 {
			return tools.ErrorText(fmt.Sprintf("node_count must be at least 1, got %d", *input.NodeCount)), nil, nil
		}
		var serverType ionos.KubernetesNodePoolServerType
		if input.ServerType != nil {
			var msg string
			if serverType, msg = normalizeServerType(*input.ServerType); msg != "" {
				return tools.ErrorText(msg), nil, nil
			}
		}
		window, msg := buildMaintenanceWindow(input.MaintenanceWindow)
		if msg != "" {
			return tools.ErrorText(msg), nil, nil
		}
		auto, msg := buildAutoScaling(input.AutoScaling)
		if msg != "" {
			return tools.ErrorText(msg), nil, nil
		}
		lans, msg := buildLans(input.Lans)
		if msg != "" {
			return tools.ErrorText(msg), nil, nil
		}
		if input.PublicIps != nil {
			if msg := validateIPs("public_ips", input.PublicIps); msg != "" {
				return tools.ErrorText(msg), nil, nil
			}
		}

		// Replacing PUT + nodeCount always serialized: read first, override only what
		// the caller supplied, or an unrelated change sends nodeCount=0 and drains the pool.
		current, _, err := client.KubernetesApi.K8sNodepoolsFindById(ctx, clusterID, poolID).Depth(1).Execute()
		if err != nil {
			if tools.IsNotFound(err) {
				return tools.ErrorText(fmt.Sprintf("node pool %s does not exist in cluster %s; nothing to update", poolID, clusterID)), nil, nil
			}
			return tools.ToResult(nil, err)
		}
		cp := current.GetProperties()

		nodeCount := cp.GetNodeCount()
		if input.NodeCount != nil {
			nodeCount = *input.NodeCount
		}
		// Literal, not the constructor: it injects serverType=DedicatedCore and would
		// rebuild a VCPU pool's nodes. See CLAUDE.md.
		props := &ionos.KubernetesNodePoolPropertiesForPut{NodeCount: nodeCount}

		// name is never sent: the API rejects it as immutable, and it is optional here.
		switch {
		case serverType != "":
			props.SetServerType(serverType)
		case cp.HasServerType():
			props.SetServerType(cp.GetServerType())
		}
		switch {
		case input.K8sVersion != nil:
			props.SetK8sVersion(strings.TrimSpace(*input.K8sVersion))
		case cp.HasK8sVersion():
			props.SetK8sVersion(cp.GetK8sVersion())
		}
		switch {
		case window != nil:
			props.SetMaintenanceWindow(*window)
		case cp.HasMaintenanceWindow():
			props.SetMaintenanceWindow(cp.GetMaintenanceWindow())
		}
		// Only an ACTIVE autoscaler is carried forward: a pool without one reads back as
		// {0,0}, which the API rejects on write.
		switch {
		case auto != nil:
			props.SetAutoScaling(*auto)
		case autoScalingActive(cp.AutoScaling):
			props.SetAutoScaling(cp.GetAutoScaling())
		}
		switch {
		case lans != nil:
			props.SetLans(lans)
		case len(cp.GetLans()) > 0:
			props.SetLans(cp.GetLans())
		}
		switch {
		case input.Labels != nil:
			props.SetLabels(input.Labels)
		case len(cp.GetLabels()) > 0:
			props.SetLabels(cp.GetLabels())
		}
		switch {
		case input.Annotations != nil:
			props.SetAnnotations(input.Annotations)
		case len(cp.GetAnnotations()) > 0:
			props.SetAnnotations(cp.GetAnnotations())
		}
		// taints has no input (x-internal in the spec) but is still carried forward, so a
		// replacing PUT cannot drop taints applied out of band.
		if len(cp.GetTaints()) > 0 {
			props.SetTaints(cp.GetTaints())
		}
		switch {
		case input.PublicIps != nil:
			props.SetPublicIps(input.PublicIps)
		case len(cp.GetPublicIps()) > 0:
			props.SetPublicIps(cp.GetPublicIps())
		}

		// Resolved values: a supplied node count may meet carried-forward bounds/IPs.
		effectiveAuto := props.AutoScaling
		if msg := validateNodeCountWithinAutoScaling(effectiveAuto, nodeCount); msg != "" {
			return tools.ErrorText(msg), nil, nil
		}
		if msg := validatePublicIps(props.PublicIps, nodeCount, effectiveAuto); msg != "" {
			return tools.ErrorText(msg), nil, nil
		}

		body := ionos.NewKubernetesNodePoolForPut(*props)
		updated, _, err := client.KubernetesApi.K8sNodepoolsPut(ctx, clusterID, poolID).KubernetesNodePool(*body).Execute()
		return tools.ToResult(updated, err)
	})
}

func registerDeleteNodepool(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterTool(server, scope, tools.MethodDelete, &mcp.Tool{
		Name: "delete_k8s_nodepool",
		Description: "Delete a node pool and all of its worker nodes. Two-phase: call first WITHOUT confirmation_token to get a preview and a one-time token, then call again WITH the token to delete. This is irreversible — pods are evicted and node storage is lost. " +
			"If it is the cluster's last node pool, the cluster keeps running with nowhere to schedule workloads." + asyncResourceNote,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.DeleteK8sNodepoolInput) (*mcp.CallToolResult, any, error) {
		clusterID := strings.TrimSpace(input.K8sClusterID)
		poolID := strings.TrimSpace(input.NodepoolID)
		if clusterID == "" {
			return tools.ErrorText("k8s_cluster_id is required"), nil, nil
		}
		if poolID == "" {
			return tools.ErrorText("nodepool_id is required"), nil, nil
		}
		target := tools.Target(req, clusterID, poolID)

		// Phase 2: token present -> validate and execute.
		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "delete_k8s_nodepool", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("delete_k8s_nodepool", "k8s_cluster_id and nodepool_id", err)), nil, nil
			}
			_, err := client.KubernetesApi.K8sNodepoolsDelete(ctx, clusterID, poolID).Execute()
			if err != nil {
				return tools.ToResult(nil, err)
			}
			return tools.TextResult(tools.DeletedAsync("Kubernetes node pool", poolID) +
				" Poll list_k8s_nodepools to follow it: the pool reports DESTROYING and then disappears; FAILED_DESTROYING means it did not complete."), nil, nil
		}

		// Phase 1: no token -> read the pool, compute the blast radius, mint a token.
		pool, _, err := client.KubernetesApi.K8sNodepoolsFindById(ctx, clusterID, poolID).Depth(1).Execute()
		if err != nil {
			if tools.IsNotFound(err) {
				return tools.ErrorText(fmt.Sprintf("node pool %s does not exist in cluster %s; nothing to delete", poolID, clusterID)), nil, nil
			}
			return tools.ToResult(nil, err)
		}
		cp := pool.GetProperties()
		radius := tools.DestroyedRadius()
		radius.Add("worker nodes (and their volumes)", int(cp.GetNodeCount()))

		token, mErr := confirm.Mint("delete_k8s_nodepool", target)
		if mErr != nil {
			return nil, nil, mErr
		}
		return tools.TextResult(tools.Preview{
			Headline: "About to DELETE a Kubernetes node pool and every worker node in it. This is IRREVERSIBLE.\n" +
				"All pods on those nodes are evicted, and data held on the nodes' own volumes is lost.",
			Fields: tools.Fields(
				"k8s_cluster_id", clusterID,
				"nodepool_id", poolID,
				"name", cp.GetName(),
				"datacenter_id", cp.GetDatacenterId(),
				"node_count", fmt.Sprintf("%d", cp.GetNodeCount()),
				"k8s_version", cp.GetK8sVersion(),
				"auto_scaling", autoScalingText(cp.AutoScaling),
			),
			Radius:    radius,
			EmptyNote: "This node pool currently has no worker nodes.",
			Tool:      "delete_k8s_nodepool",
			Replay:    tools.Fields("k8s_cluster_id", clusterID, "nodepool_id", poolID),
			TokenNote: "This token authorizes deleting ONLY this node pool",
		}.Render(token)), nil, nil
	})
}

// mapText renders a label or annotation map for a preview line, sorted so the
// preview a caller authorizes does not reshuffle between the two phases.
func mapText(m map[string]string) string {
	if len(m) == 0 {
		return ""
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", k, m[k]))
	}
	return strings.Join(parts, ", ")
}
