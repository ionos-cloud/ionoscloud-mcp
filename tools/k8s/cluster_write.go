package k8s

import (
	"context"
	"fmt"
	"net"
	"strings"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

// A Kubernetes cluster is the managed control plane. It runs no workloads of its
// own: those live on the worker nodes of its node pools (nodepool_write.go).

// RegisterClusterWriteTools registers the create/update/delete cluster tools.
func RegisterClusterWriteTools(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	registerCreateCluster(server, client, scope, confirm)
	registerUpdateCluster(server, client, scope)
	registerDeleteCluster(server, client, scope, confirm)
}

func registerCreateCluster(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterTool(server, scope, tools.MethodPost, &mcp.Tool{
		Name: "create_k8s_cluster",
		Description: "Create one Managed Kubernetes cluster (the control plane). Two-phase: call first WITHOUT confirmation_token to get a preview and a one-time token, then call again WITH the token (and the same name and location) to create it. Creates exactly one cluster per call. " +
			"It runs no workloads until it has a node pool — add one with create_k8s_nodepool once it reports ACTIVE. location, nat_gateway_ip, node_subnet and public are immutable." + asyncResourceNote,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.CreateK8sClusterInput) (*mcp.CallToolResult, any, error) {
		name := strings.TrimSpace(input.Name)
		if name == "" {
			return tools.ErrorText("name is required to create a Kubernetes cluster"), nil, nil
		}
		location := strings.TrimSpace(tools.OptStr(input.Location))

		// A private cluster needs a location and a reserved NAT gateway IP. The API
		// rejects it otherwise, without saying which of the two is missing.
		isPrivate := input.Public != nil && !*input.Public
		if isPrivate {
			if location == "" {
				return tools.ErrorText("location is required for a private cluster (public=false), e.g. de/fra"), nil, nil
			}
			if strings.TrimSpace(tools.OptStr(input.NatGatewayIp)) == "" {
				return tools.ErrorText("nat_gateway_ip is required for a private cluster (public=false); it must be an IP you have already reserved in the cluster's location (see list_ip_blocks)"), nil, nil
			}
		}
		if input.NatGatewayIp != nil {
			if net.ParseIP(strings.TrimSpace(*input.NatGatewayIp)) == nil {
				return tools.ErrorText(fmt.Sprintf("nat_gateway_ip %q is not an IP address", *input.NatGatewayIp)), nil, nil
			}
		}
		if input.NodeSubnet != nil {
			if _, _, err := net.ParseCIDR(strings.TrimSpace(*input.NodeSubnet)); err != nil {
				return tools.ErrorText(fmt.Sprintf("node_subnet %q is not a CIDR; use a 16-bit IPv4 prefix, e.g. 10.0.0.0/16", *input.NodeSubnet)), nil, nil
			}
		}
		if msg := validateIPsOrCIDRs("api_subnet_allow_list", input.ApiSubnetAllowList); msg != "" {
			return tools.ErrorText(msg), nil, nil
		}
		window, msg := buildMaintenanceWindow(input.MaintenanceWindow)
		if msg != "" {
			return tools.ErrorText(msg), nil, nil
		}
		buckets, msg := buildS3Buckets(input.S3Buckets)
		if msg != "" {
			return tools.ErrorText(msg), nil, nil
		}
		target := tools.Target(req, name, location)

		// Phase 2: token present -> validate and execute.
		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "create_k8s_cluster", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("create_k8s_cluster", "name and location", err)), nil, nil
			}
			props := ionos.NewKubernetesClusterPropertiesForPost(name)
			if input.K8sVersion != nil {
				props.SetK8sVersion(strings.TrimSpace(*input.K8sVersion))
			}
			if window != nil {
				props.SetMaintenanceWindow(*window)
			}
			if input.Public != nil {
				props.SetPublic(*input.Public)
			}
			if location != "" {
				props.SetLocation(location)
			}
			if input.NatGatewayIp != nil {
				props.SetNatGatewayIp(strings.TrimSpace(*input.NatGatewayIp))
			}
			if input.NodeSubnet != nil {
				props.SetNodeSubnet(strings.TrimSpace(*input.NodeSubnet))
			}
			if len(input.ApiSubnetAllowList) > 0 {
				props.SetApiSubnetAllowList(input.ApiSubnetAllowList)
			}
			if len(buckets) > 0 {
				props.SetS3Buckets(buckets)
			}
			body := ionos.NewKubernetesClusterForPost(*props)
			created, _, err := client.KubernetesApi.K8sPost(ctx).KubernetesCluster(*body).Execute()
			return tools.ToResult(created, err)
		}

		// Phase 1: no token -> preview and mint a one-time token.
		token, err := confirm.Mint("create_k8s_cluster", target)
		if err != nil {
			return nil, nil, err
		}
		headline := "About to CREATE one Kubernetes cluster (control plane only):"
		if len(input.ApiSubnetAllowList) == 0 {
			headline += "\nNOTE: no api_subnet_allow_list was given, so the Kubernetes API server accepts connections from any source address."
		}
		return tools.TextResult(tools.Preview{
			Headline: headline,
			Fields: tools.Fields(
				"name", name,
				"k8s_version", strings.TrimSpace(tools.OptStr(input.K8sVersion)),
				"public", tools.OptBool(input.Public),
				"location", location,
				"nat_gateway_ip", strings.TrimSpace(tools.OptStr(input.NatGatewayIp)),
				"node_subnet", strings.TrimSpace(tools.OptStr(input.NodeSubnet)),
				"api_subnet_allow_list", strings.Join(input.ApiSubnetAllowList, ", "),
				"s3_buckets", s3BucketNames(buckets),
				"maintenance_window", maintenanceWindowText(window),
			),
			Tool:      "create_k8s_cluster",
			Replay:    tools.Fields("name", name, "location", location),
			TokenNote: "This creates exactly one cluster, with no node pools yet. The token authorizes creating only this name+location",
		}.Render(token)), nil, nil
	})
}

func registerUpdateCluster(server *mcp.Server, client *ionos.APIClient, scope tools.Scope) {
	tools.RegisterTool(server, scope, tools.MethodPut, &mcp.Tool{
		Name: "update_k8s_cluster",
		Description: "Update a cluster's name, Kubernetes version, maintenance window, API server allow list or audit-log buckets. location, nat_gateway_ip, node_subnet and public are immutable. " +
			"This endpoint replaces the cluster's properties, so fields you omit are read and sent back unchanged. api_subnet_allow_list and s3_buckets replace the current list when supplied. " +
			"BE CAREFUL with k8s_version: it upgrades the control plane and cannot be undone. Confirm the version with the user before sending it, and check the cluster's availableUpgradeVersions first — nothing else is accepted." + asyncResourceNote,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.UpdateK8sClusterInput) (*mcp.CallToolResult, any, error) {
		id := strings.TrimSpace(input.K8sClusterID)
		if id == "" {
			return tools.ErrorText("k8s_cluster_id is required"), nil, nil
		}
		if input.Name == nil && input.K8sVersion == nil && input.MaintenanceWindow == nil &&
			input.ApiSubnetAllowList == nil && input.S3Buckets == nil {
			return tools.ErrorText("nothing to update: provide at least one of name, k8s_version, maintenance_window, api_subnet_allow_list, s3_buckets"), nil, nil
		}
		if input.Name != nil && strings.TrimSpace(*input.Name) == "" {
			return tools.ErrorText("name must not be empty; omit it entirely to keep the current name"), nil, nil
		}
		if msg := validateIPsOrCIDRs("api_subnet_allow_list", input.ApiSubnetAllowList); msg != "" {
			return tools.ErrorText(msg), nil, nil
		}
		window, msg := buildMaintenanceWindow(input.MaintenanceWindow)
		if msg != "" {
			return tools.ErrorText(msg), nil, nil
		}
		buckets, msg := buildS3Buckets(input.S3Buckets)
		if msg != "" {
			return tools.ErrorText(msg), nil, nil
		}

		// The endpoint replaces the cluster's properties and the SDK serializes name
		// unconditionally, so read the current cluster and override only what the
		// caller supplied. Without this, changing the version alone would send an
		// empty name and drop the API server allow list.
		current, _, err := client.KubernetesApi.K8sFindByClusterId(ctx, id).Depth(1).Execute()
		if err != nil {
			if tools.IsNotFound(err) {
				return tools.ErrorText(fmt.Sprintf("Kubernetes cluster %s does not exist; nothing to update", id)), nil, nil
			}
			return tools.ToResult(nil, err)
		}
		cp := current.GetProperties()

		name := cp.GetName()
		if input.Name != nil {
			name = strings.TrimSpace(*input.Name)
		}
		// A zero-valued literal rather than NewKubernetesClusterPropertiesForPut():
		// see the "PATCH bodies" note in CLAUDE.md — the same hazard applies to a PUT
		// body assembled field by field.
		props := &ionos.KubernetesClusterPropertiesForPut{Name: name}

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
		// Carried forward so an unrelated change cannot silently expose the
		// Kubernetes API server to every source address.
		switch {
		case input.ApiSubnetAllowList != nil:
			props.SetApiSubnetAllowList(input.ApiSubnetAllowList)
		case len(cp.GetApiSubnetAllowList()) > 0:
			props.SetApiSubnetAllowList(cp.GetApiSubnetAllowList())
		}
		switch {
		case buckets != nil:
			props.SetS3Buckets(buckets)
		case len(cp.GetS3Buckets()) > 0:
			props.SetS3Buckets(cp.GetS3Buckets())
		}

		body := ionos.NewKubernetesClusterForPut(*props)
		updated, _, err := client.KubernetesApi.K8sPut(ctx, id).KubernetesCluster(*body).Execute()
		return tools.ToResult(updated, err)
	})
}

func registerDeleteCluster(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterTool(server, scope, tools.MethodDelete, &mcp.Tool{
		Name: "delete_k8s_cluster",
		Description: "Delete a Managed Kubernetes cluster. Two-phase: call first WITHOUT confirmation_token to get a blast-radius preview and a one-time token, then call again WITH the token to delete. This is irreversible. " +
			"The preview summarises how many node pools and worker nodes it would take with it." + asyncResourceNote,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.DeleteK8sClusterInput) (*mcp.CallToolResult, any, error) {
		id := strings.TrimSpace(input.K8sClusterID)
		if id == "" {
			return tools.ErrorText("k8s_cluster_id is required"), nil, nil
		}
		target := tools.Target(req, id)

		// Phase 2: token present -> validate and execute.
		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "delete_k8s_cluster", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("delete_k8s_cluster", "k8s_cluster_id", err)), nil, nil
			}
			_, err := client.KubernetesApi.K8sDelete(ctx, id).Execute()
			if err != nil {
				return tools.ToResult(nil, err)
			}
			return tools.TextResult(tools.DeletedAsync("Kubernetes cluster", id) +
				" Poll get_k8s_cluster to follow it: metadata.state goes DESTROYING and then the cluster stops resolving; FAILED_DESTROYING means it did not complete."), nil, nil
		}

		// Phase 1: no token -> read the cluster and its node pools, preview, mint.
		cluster, _, err := client.KubernetesApi.K8sFindByClusterId(ctx, id).Depth(2).Execute()
		if err != nil {
			if tools.IsNotFound(err) {
				return tools.ErrorText(fmt.Sprintf("Kubernetes cluster %s does not exist; nothing to delete", id)), nil, nil
			}
			return tools.ToResult(nil, err)
		}
		pools, workers := clusterNodepoolCounts(cluster)
		radius := tools.DestroyedRadius()
		radius.Add("node pools", pools)
		radius.Add("worker nodes", workers)

		token, mErr := confirm.Mint("delete_k8s_cluster", target)
		if mErr != nil {
			return nil, nil, mErr
		}
		cp := cluster.GetProperties()
		return tools.TextResult(tools.Preview{
			Headline: "About to DELETE a Kubernetes cluster. This is IRREVERSIBLE.",
			Fields: tools.Fields(
				"k8s_cluster_id", id,
				"name", cp.GetName(),
				"k8s_version", cp.GetK8sVersion(),
				"state", clusterState(cluster),
				"location", cp.GetLocation(),
				"public", tools.OptBool(cp.Public),
			),
			Radius:    radius,
			EmptyNote: "This cluster has no node pools; deleting removes only the control plane.",
			Tool:      "delete_k8s_cluster",
			Replay:    tools.Fields("k8s_cluster_id", id),
			TokenNote: "This token authorizes deleting ONLY this cluster",
		}.Render(token)), nil, nil
	})
}

// clusterNodepoolCounts counts the node pools of a cluster read at depth 2, and the
// worker nodes those pools are sized for. Depth 2 populates the nodepool collection
// with each pool's properties, which is where the node count lives.
func clusterNodepoolCounts(cluster ionos.KubernetesCluster) (pools, workers int) {
	e := cluster.Entities
	if e == nil || e.Nodepools == nil {
		return 0, 0
	}
	items := e.Nodepools.Items
	for _, np := range items {
		props := np.GetProperties()
		workers += int(props.GetNodeCount())
	}
	return len(items), workers
}

// clusterState reads the provisioning state out of the cluster metadata, which is
// where the API reports ACTIVE / BUSY / DESTROYING rather than in properties.
func clusterState(cluster ionos.KubernetesCluster) string {
	if cluster.Metadata == nil {
		return ""
	}
	return cluster.Metadata.GetState()
}
