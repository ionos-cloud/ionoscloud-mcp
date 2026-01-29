package main

import (
	"context"
	"encoding/json"
	"fmt"

	dns "github.com/ionos-cloud/sdk-go-bundle/products/dns/v2"
	ionoscloud "github.com/ionos-cloud/sdk-go/v6"
)

func (s *Server) executeTool(name string, arguments map[string]interface{}) (string, error) {
	switch name {
	case "list_datacenters":
		return s.listDatacenters(s.client, s.ctx)
	case "get_datacenter":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		return s.getDatacenter(s.client, s.ctx, datacenterID)
	case "create_datacenter":
		name, ok := arguments["name"].(string)
		if !ok {
			return "", fmt.Errorf("name is required")
		}
		location, ok := arguments["location"].(string)
		if !ok {
			return "", fmt.Errorf("location is required")
		}
		description, _ := arguments["description"].(string)
		return s.createDatacenter(s.client, s.ctx, name, location, description)
	case "update_datacenter":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		name, _ := arguments["name"].(string)
		description, _ := arguments["description"].(string)
		return s.updateDatacenter(s.client, s.ctx, datacenterID, name, description)
	case "delete_datacenter":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		return s.deleteDatacenter(s.client, s.ctx, datacenterID)
	case "list_servers":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		return s.listServers(s.client, s.ctx, datacenterID)
	case "get_server":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		serverID, ok := arguments["server_id"].(string)
		if !ok {
			return "", fmt.Errorf("server_id is required")
		}
		return s.getServer(s.client, s.ctx, datacenterID, serverID)
	case "create_server":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		name, ok := arguments["name"].(string)
		if !ok {
			return "", fmt.Errorf("name is required")
		}
		cores, ok := arguments["cores"].(float64)
		if !ok {
			return "", fmt.Errorf("cores is required")
		}
		ram, ok := arguments["ram"].(float64)
		if !ok {
			return "", fmt.Errorf("ram is required")
		}
		cpuFamily, _ := arguments["cpu_family"].(string)
		availabilityZone, _ := arguments["availability_zone"].(string)
		return s.createServer(s.client, s.ctx, datacenterID, name, int32(cores), int32(ram), cpuFamily, availabilityZone)
	case "update_server":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		serverID, ok := arguments["server_id"].(string)
		if !ok {
			return "", fmt.Errorf("server_id is required")
		}
		name, _ := arguments["name"].(string)
		var cores, ram int32
		if c, ok := arguments["cores"].(float64); ok {
			cores = int32(c)
		}
		if r, ok := arguments["ram"].(float64); ok {
			ram = int32(r)
		}
		return s.updateServer(s.client, s.ctx, datacenterID, serverID, name, cores, ram)
	case "delete_server":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		serverID, ok := arguments["server_id"].(string)
		if !ok {
			return "", fmt.Errorf("server_id is required")
		}
		return s.deleteServer(s.client, s.ctx, datacenterID, serverID)
	case "start_server":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		serverID, ok := arguments["server_id"].(string)
		if !ok {
			return "", fmt.Errorf("server_id is required")
		}
		return s.startServer(s.client, s.ctx, datacenterID, serverID)
	case "stop_server":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		serverID, ok := arguments["server_id"].(string)
		if !ok {
			return "", fmt.Errorf("server_id is required")
		}
		return s.stopServer(s.client, s.ctx, datacenterID, serverID)
	case "reboot_server":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		serverID, ok := arguments["server_id"].(string)
		if !ok {
			return "", fmt.Errorf("server_id is required")
		}
		return s.rebootServer(s.client, s.ctx, datacenterID, serverID)
	case "list_volumes":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		return s.listVolumes(s.client, s.ctx, datacenterID)
	case "get_volume":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		volumeID, ok := arguments["volume_id"].(string)
		if !ok {
			return "", fmt.Errorf("volume_id is required")
		}
		return s.getVolume(s.client, s.ctx, datacenterID, volumeID)
	case "create_volume":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		name, ok := arguments["name"].(string)
		if !ok {
			return "", fmt.Errorf("name is required")
		}
		size, ok := arguments["size"].(float64)
		if !ok {
			return "", fmt.Errorf("size is required")
		}
		volumeType, _ := arguments["type"].(string)
		bus, _ := arguments["bus"].(string)
		availabilityZone, _ := arguments["availability_zone"].(string)
		image, _ := arguments["image"].(string)
		imagePassword, _ := arguments["image_password"].(string)
		licenceType, _ := arguments["licence_type"].(string)
		return s.createVolume(s.client, s.ctx, datacenterID, name, float32(size), volumeType, bus, availabilityZone, image, imagePassword, licenceType)
	case "update_volume":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		volumeID, ok := arguments["volume_id"].(string)
		if !ok {
			return "", fmt.Errorf("volume_id is required")
		}
		name, _ := arguments["name"].(string)
		var size float32
		if s, ok := arguments["size"].(float64); ok {
			size = float32(s)
		}
		return s.updateVolume(s.client, s.ctx, datacenterID, volumeID, name, size)
	case "delete_volume":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		volumeID, ok := arguments["volume_id"].(string)
		if !ok {
			return "", fmt.Errorf("volume_id is required")
		}
		return s.deleteVolume(s.client, s.ctx, datacenterID, volumeID)
	case "attach_volume":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		serverID, ok := arguments["server_id"].(string)
		if !ok {
			return "", fmt.Errorf("server_id is required")
		}
		volumeID, ok := arguments["volume_id"].(string)
		if !ok {
			return "", fmt.Errorf("volume_id is required")
		}
		return s.attachVolume(s.client, s.ctx, datacenterID, serverID, volumeID)
	case "detach_volume":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		serverID, ok := arguments["server_id"].(string)
		if !ok {
			return "", fmt.Errorf("server_id is required")
		}
		volumeID, ok := arguments["volume_id"].(string)
		if !ok {
			return "", fmt.Errorf("volume_id is required")
		}
		return s.detachVolume(s.client, s.ctx, datacenterID, serverID, volumeID)
	case "list_images":
		return s.listImages(s.client, s.ctx)
	case "list_locations":
		return s.listLocations(s.client, s.ctx)
	case "list_snapshots":
		return s.listSnapshots(s.client, s.ctx)
	case "get_snapshot":
		snapshotID, ok := arguments["snapshot_id"].(string)
		if !ok {
			return "", fmt.Errorf("snapshot_id is required")
		}
		return s.getSnapshot(s.client, s.ctx, snapshotID)
	case "create_snapshot":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		volumeID, ok := arguments["volume_id"].(string)
		if !ok {
			return "", fmt.Errorf("volume_id is required")
		}
		name, _ := arguments["name"].(string)
		description, _ := arguments["description"].(string)
		return s.createSnapshot(s.client, s.ctx, datacenterID, volumeID, name, description)
	case "update_snapshot":
		snapshotID, ok := arguments["snapshot_id"].(string)
		if !ok {
			return "", fmt.Errorf("snapshot_id is required")
		}
		name, _ := arguments["name"].(string)
		description, _ := arguments["description"].(string)
		return s.updateSnapshot(s.client, s.ctx, snapshotID, name, description)
	case "delete_snapshot":
		snapshotID, ok := arguments["snapshot_id"].(string)
		if !ok {
			return "", fmt.Errorf("snapshot_id is required")
		}
		return s.deleteSnapshot(s.client, s.ctx, snapshotID)
	case "restore_snapshot":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		volumeID, ok := arguments["volume_id"].(string)
		if !ok {
			return "", fmt.Errorf("volume_id is required")
		}
		snapshotID, ok := arguments["snapshot_id"].(string)
		if !ok {
			return "", fmt.Errorf("snapshot_id is required")
		}
		return s.restoreSnapshot(s.client, s.ctx, datacenterID, volumeID, snapshotID)
	// Networking - LANs
	case "list_lans":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		return s.listLans(s.client, s.ctx, datacenterID)
	case "get_lan":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		lanID, ok := arguments["lan_id"].(string)
		if !ok {
			return "", fmt.Errorf("lan_id is required")
		}
		return s.getLan(s.client, s.ctx, datacenterID, lanID)
	// Networking - NICs
	case "list_nics":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		serverID, ok := arguments["server_id"].(string)
		if !ok {
			return "", fmt.Errorf("server_id is required")
		}
		return s.listNics(s.client, s.ctx, datacenterID, serverID)
	case "get_nic":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		serverID, ok := arguments["server_id"].(string)
		if !ok {
			return "", fmt.Errorf("server_id is required")
		}
		nicID, ok := arguments["nic_id"].(string)
		if !ok {
			return "", fmt.Errorf("nic_id is required")
		}
		return s.getNic(s.client, s.ctx, datacenterID, serverID, nicID)
	// Networking - IP Blocks
	case "list_ipblocks":
		return s.listIpBlocks(s.client, s.ctx)
	case "get_ipblock":
		ipblockID, ok := arguments["ipblock_id"].(string)
		if !ok {
			return "", fmt.Errorf("ipblock_id is required")
		}
		return s.getIpBlock(s.client, s.ctx, ipblockID)
	// Networking - Firewall Rules
	case "list_firewall_rules":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		serverID, ok := arguments["server_id"].(string)
		if !ok {
			return "", fmt.Errorf("server_id is required")
		}
		nicID, ok := arguments["nic_id"].(string)
		if !ok {
			return "", fmt.Errorf("nic_id is required")
		}
		return s.listFirewallRules(s.client, s.ctx, datacenterID, serverID, nicID)
	case "get_firewall_rule":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		serverID, ok := arguments["server_id"].(string)
		if !ok {
			return "", fmt.Errorf("server_id is required")
		}
		nicID, ok := arguments["nic_id"].(string)
		if !ok {
			return "", fmt.Errorf("nic_id is required")
		}
		firewallRuleID, ok := arguments["firewallrule_id"].(string)
		if !ok {
			return "", fmt.Errorf("firewallrule_id is required")
		}
		return s.getFirewallRule(s.client, s.ctx, datacenterID, serverID, nicID, firewallRuleID)
	// Networking - NAT Gateways
	case "list_nat_gateways":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		return s.listNatGateways(s.client, s.ctx, datacenterID)
	case "get_nat_gateway":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		natGatewayID, ok := arguments["nat_gateway_id"].(string)
		if !ok {
			return "", fmt.Errorf("nat_gateway_id is required")
		}
		return s.getNatGateway(s.client, s.ctx, datacenterID, natGatewayID)
	// Networking - Private Cross Connects
	case "list_pccs":
		return s.listPccs(s.client, s.ctx)
	case "get_pcc":
		pccID, ok := arguments["pcc_id"].(string)
		if !ok {
			return "", fmt.Errorf("pcc_id is required")
		}
		return s.getPcc(s.client, s.ctx, pccID)
	// Load Balancers - Application Load Balancers
	case "list_application_load_balancers":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		return s.listApplicationLoadBalancers(s.client, s.ctx, datacenterID)
	case "get_application_load_balancer":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		albID, ok := arguments["alb_id"].(string)
		if !ok {
			return "", fmt.Errorf("alb_id is required")
		}
		return s.getApplicationLoadBalancer(s.client, s.ctx, datacenterID, albID)
	case "list_alb_forwarding_rules":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		albID, ok := arguments["alb_id"].(string)
		if !ok {
			return "", fmt.Errorf("alb_id is required")
		}
		return s.listAlbForwardingRules(s.client, s.ctx, datacenterID, albID)
	case "get_alb_forwarding_rule":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		albID, ok := arguments["alb_id"].(string)
		if !ok {
			return "", fmt.Errorf("alb_id is required")
		}
		ruleID, ok := arguments["rule_id"].(string)
		if !ok {
			return "", fmt.Errorf("rule_id is required")
		}
		return s.getAlbForwardingRule(s.client, s.ctx, datacenterID, albID, ruleID)
	// Load Balancers - Network Load Balancers
	case "list_network_load_balancers":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		return s.listNetworkLoadBalancers(s.client, s.ctx, datacenterID)
	case "get_network_load_balancer":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		nlbID, ok := arguments["nlb_id"].(string)
		if !ok {
			return "", fmt.Errorf("nlb_id is required")
		}
		return s.getNetworkLoadBalancer(s.client, s.ctx, datacenterID, nlbID)
	// Load Balancers - Target Groups
	case "list_target_groups":
		return s.listTargetGroups(s.client, s.ctx)
	case "get_target_group":
		targetGroupID, ok := arguments["target_group_id"].(string)
		if !ok {
			return "", fmt.Errorf("target_group_id is required")
		}
		return s.getTargetGroup(s.client, s.ctx, targetGroupID)
	// Kubernetes
	case "list_k8s_clusters":
		return s.listK8sClusters(s.client, s.ctx)
	case "get_k8s_cluster":
		k8sClusterID, ok := arguments["k8s_cluster_id"].(string)
		if !ok {
			return "", fmt.Errorf("k8s_cluster_id is required")
		}
		return s.getK8sCluster(s.client, s.ctx, k8sClusterID)
	case "get_k8s_kubeconfig":
		k8sClusterID, ok := arguments["k8s_cluster_id"].(string)
		if !ok {
			return "", fmt.Errorf("k8s_cluster_id is required")
		}
		return s.getK8sKubeconfig(s.client, s.ctx, k8sClusterID)
	case "list_k8s_nodepools":
		k8sClusterID, ok := arguments["k8s_cluster_id"].(string)
		if !ok {
			return "", fmt.Errorf("k8s_cluster_id is required")
		}
		return s.listK8sNodepools(s.client, s.ctx, k8sClusterID)
	case "get_k8s_nodepool":
		k8sClusterID, ok := arguments["k8s_cluster_id"].(string)
		if !ok {
			return "", fmt.Errorf("k8s_cluster_id is required")
		}
		nodepoolID, ok := arguments["nodepool_id"].(string)
		if !ok {
			return "", fmt.Errorf("nodepool_id is required")
		}
		return s.getK8sNodepool(s.client, s.ctx, k8sClusterID, nodepoolID)
	case "list_k8s_nodes":
		k8sClusterID, ok := arguments["k8s_cluster_id"].(string)
		if !ok {
			return "", fmt.Errorf("k8s_cluster_id is required")
		}
		nodepoolID, ok := arguments["nodepool_id"].(string)
		if !ok {
			return "", fmt.Errorf("nodepool_id is required")
		}
		return s.listK8sNodes(s.client, s.ctx, k8sClusterID, nodepoolID)
	case "get_k8s_node":
		k8sClusterID, ok := arguments["k8s_cluster_id"].(string)
		if !ok {
			return "", fmt.Errorf("k8s_cluster_id is required")
		}
		nodepoolID, ok := arguments["nodepool_id"].(string)
		if !ok {
			return "", fmt.Errorf("nodepool_id is required")
		}
		nodeID, ok := arguments["node_id"].(string)
		if !ok {
			return "", fmt.Errorf("node_id is required")
		}
		return s.getK8sNode(s.client, s.ctx, k8sClusterID, nodepoolID, nodeID)
	case "list_k8s_versions":
		return s.listK8sVersions(s.client, s.ctx)
	// User Management - Users
	case "list_users":
		return s.listUsers(s.client, s.ctx)
	case "get_user":
		userID, ok := arguments["user_id"].(string)
		if !ok {
			return "", fmt.Errorf("user_id is required")
		}
		return s.getUser(s.client, s.ctx, userID)
	// User Management - Groups
	case "list_groups":
		return s.listGroups(s.client, s.ctx)
	case "get_group":
		groupID, ok := arguments["group_id"].(string)
		if !ok {
			return "", fmt.Errorf("group_id is required")
		}
		return s.getGroup(s.client, s.ctx, groupID)
	case "list_group_members":
		groupID, ok := arguments["group_id"].(string)
		if !ok {
			return "", fmt.Errorf("group_id is required")
		}
		return s.listGroupMembers(s.client, s.ctx, groupID)
	case "list_user_groups":
		userID, ok := arguments["user_id"].(string)
		if !ok {
			return "", fmt.Errorf("user_id is required")
		}
		return s.listUserGroups(s.client, s.ctx, userID)
	// User Management - S3 Keys
	case "list_s3_keys":
		userID, ok := arguments["user_id"].(string)
		if !ok {
			return "", fmt.Errorf("user_id is required")
		}
		return s.listS3Keys(s.client, s.ctx, userID)
	case "get_s3_key":
		userID, ok := arguments["user_id"].(string)
		if !ok {
			return "", fmt.Errorf("user_id is required")
		}
		keyID, ok := arguments["key_id"].(string)
		if !ok {
			return "", fmt.Errorf("key_id is required")
		}
		return s.getS3Key(s.client, s.ctx, userID, keyID)
	// User Management - Contract
	case "get_contract":
		return s.getContract(s.client, s.ctx)
	case "list_resources":
		resourceType, _ := arguments["resource_type"].(string)
		return s.listResources(s.client, s.ctx, resourceType)
	// DNS
	case "list_dns_zones":
		return s.listDnsZones(s.dnsClient, s.ctx)
	case "get_dns_zone":
		zoneID, ok := arguments["zone_id"].(string)
		if !ok {
			return "", fmt.Errorf("zone_id is required")
		}
		return s.getDnsZone(s.dnsClient, s.ctx, zoneID)
	case "list_dns_records":
		zoneID, ok := arguments["zone_id"].(string)
		if !ok {
			return "", fmt.Errorf("zone_id is required")
		}
		return s.listDnsRecords(s.dnsClient, s.ctx, zoneID)
	case "get_dns_record":
		zoneID, ok := arguments["zone_id"].(string)
		if !ok {
			return "", fmt.Errorf("zone_id is required")
		}
		recordID, ok := arguments["record_id"].(string)
		if !ok {
			return "", fmt.Errorf("record_id is required")
		}
		return s.getDnsRecord(s.dnsClient, s.ctx, zoneID, recordID)
	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}

func (s *Server) listDatacenters(client *ionoscloud.APIClient, ctx context.Context) (string, error) {
	datacenters, _, err := client.DataCentersApi.DatacentersGet(ctx).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list datacenters: %w", err)
	}

	data, err := json.MarshalIndent(datacenters, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal datacenters: %w", err)
	}

	return string(data), nil
}

func (s *Server) getDatacenter(client *ionoscloud.APIClient, ctx context.Context, datacenterID string) (string, error) {
	datacenter, _, err := client.DataCentersApi.DatacentersFindById(ctx, datacenterID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get datacenter: %w", err)
	}

	data, err := json.MarshalIndent(datacenter, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal datacenter: %w", err)
	}

	return string(data), nil
}

func (s *Server) createDatacenter(client *ionoscloud.APIClient, ctx context.Context, name, location, description string) (string, error) {
	properties := ionoscloud.DatacenterPropertiesPost{
		Name:     &name,
		Location: &location,
	}
	if description != "" {
		properties.Description = &description
	}

	datacenter := ionoscloud.DatacenterPost{
		Properties: &properties,
	}

	result, _, err := client.DataCentersApi.DatacentersPost(ctx).Datacenter(datacenter).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to create datacenter: %w", err)
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal datacenter: %w", err)
	}

	return string(data), nil
}

func (s *Server) updateDatacenter(client *ionoscloud.APIClient, ctx context.Context, datacenterID, name, description string) (string, error) {
	properties := ionoscloud.DatacenterPropertiesPut{}
	if name != "" {
		properties.Name = &name
	}
	if description != "" {
		properties.Description = &description
	}

	result, _, err := client.DataCentersApi.DatacentersPatch(ctx, datacenterID).Datacenter(properties).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to update datacenter: %w", err)
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal datacenter: %w", err)
	}

	return string(data), nil
}

func (s *Server) deleteDatacenter(client *ionoscloud.APIClient, ctx context.Context, datacenterID string) (string, error) {
	_, err := client.DataCentersApi.DatacentersDelete(ctx, datacenterID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to delete datacenter: %w", err)
	}

	return fmt.Sprintf(`{"status": "deleted", "datacenter_id": "%s"}`, datacenterID), nil
}

func (s *Server) listServers(client *ionoscloud.APIClient, ctx context.Context, datacenterID string) (string, error) {
	servers, _, err := client.ServersApi.DatacentersServersGet(ctx, datacenterID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list servers: %w", err)
	}

	data, err := json.MarshalIndent(servers, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal servers: %w", err)
	}

	return string(data), nil
}

func (s *Server) getServer(client *ionoscloud.APIClient, ctx context.Context, datacenterID, serverID string) (string, error) {
	server, _, err := client.ServersApi.DatacentersServersFindById(ctx, datacenterID, serverID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get server: %w", err)
	}

	data, err := json.MarshalIndent(server, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal server: %w", err)
	}

	return string(data), nil
}

func (s *Server) createServer(client *ionoscloud.APIClient, ctx context.Context, datacenterID, name string, cores, ram int32, cpuFamily, availabilityZone string) (string, error) {
	properties := ionoscloud.ServerProperties{
		Name:  &name,
		Cores: &cores,
		Ram:   &ram,
	}
	if cpuFamily != "" {
		properties.CpuFamily = &cpuFamily
	}
	if availabilityZone != "" {
		properties.AvailabilityZone = &availabilityZone
	}

	server := ionoscloud.Server{
		Properties: &properties,
	}

	result, _, err := client.ServersApi.DatacentersServersPost(ctx, datacenterID).Server(server).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to create server: %w", err)
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal server: %w", err)
	}

	return string(data), nil
}

func (s *Server) updateServer(client *ionoscloud.APIClient, ctx context.Context, datacenterID, serverID, name string, cores, ram int32) (string, error) {
	properties := ionoscloud.ServerProperties{}
	if name != "" {
		properties.Name = &name
	}
	if cores > 0 {
		properties.Cores = &cores
	}
	if ram > 0 {
		properties.Ram = &ram
	}

	result, _, err := client.ServersApi.DatacentersServersPatch(ctx, datacenterID, serverID).Server(properties).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to update server: %w", err)
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal server: %w", err)
	}

	return string(data), nil
}

func (s *Server) deleteServer(client *ionoscloud.APIClient, ctx context.Context, datacenterID, serverID string) (string, error) {
	_, err := client.ServersApi.DatacentersServersDelete(ctx, datacenterID, serverID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to delete server: %w", err)
	}

	return fmt.Sprintf(`{"status": "deleted", "server_id": "%s"}`, serverID), nil
}

func (s *Server) startServer(client *ionoscloud.APIClient, ctx context.Context, datacenterID, serverID string) (string, error) {
	_, err := client.ServersApi.DatacentersServersStartPost(ctx, datacenterID, serverID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to start server: %w", err)
	}

	return fmt.Sprintf(`{"status": "starting", "server_id": "%s"}`, serverID), nil
}

func (s *Server) stopServer(client *ionoscloud.APIClient, ctx context.Context, datacenterID, serverID string) (string, error) {
	_, err := client.ServersApi.DatacentersServersStopPost(ctx, datacenterID, serverID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to stop server: %w", err)
	}

	return fmt.Sprintf(`{"status": "stopping", "server_id": "%s"}`, serverID), nil
}

func (s *Server) rebootServer(client *ionoscloud.APIClient, ctx context.Context, datacenterID, serverID string) (string, error) {
	_, err := client.ServersApi.DatacentersServersRebootPost(ctx, datacenterID, serverID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to reboot server: %w", err)
	}

	return fmt.Sprintf(`{"status": "rebooting", "server_id": "%s"}`, serverID), nil
}

func (s *Server) listVolumes(client *ionoscloud.APIClient, ctx context.Context, datacenterID string) (string, error) {
	volumes, _, err := client.VolumesApi.DatacentersVolumesGet(ctx, datacenterID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list volumes: %w", err)
	}

	data, err := json.MarshalIndent(volumes, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal volumes: %w", err)
	}

	return string(data), nil
}

func (s *Server) getVolume(client *ionoscloud.APIClient, ctx context.Context, datacenterID, volumeID string) (string, error) {
	volume, _, err := client.VolumesApi.DatacentersVolumesFindById(ctx, datacenterID, volumeID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get volume: %w", err)
	}

	data, err := json.MarshalIndent(volume, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal volume: %w", err)
	}

	return string(data), nil
}

func (s *Server) createVolume(client *ionoscloud.APIClient, ctx context.Context, datacenterID, name string, size float32, volumeType, bus, availabilityZone, image, imagePassword, licenceType string) (string, error) {
	properties := ionoscloud.VolumeProperties{
		Name: &name,
		Size: &size,
	}
	if volumeType != "" {
		properties.Type = &volumeType
	}
	if bus != "" {
		properties.Bus = &bus
	}
	if availabilityZone != "" {
		properties.AvailabilityZone = &availabilityZone
	}
	if image != "" {
		properties.Image = &image
	}
	if imagePassword != "" {
		properties.ImagePassword = &imagePassword
	}
	if licenceType != "" {
		properties.LicenceType = &licenceType
	}

	volume := ionoscloud.Volume{
		Properties: &properties,
	}

	result, _, err := client.VolumesApi.DatacentersVolumesPost(ctx, datacenterID).Volume(volume).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to create volume: %w", err)
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal volume: %w", err)
	}

	return string(data), nil
}

func (s *Server) updateVolume(client *ionoscloud.APIClient, ctx context.Context, datacenterID, volumeID, name string, size float32) (string, error) {
	properties := ionoscloud.VolumeProperties{}
	if name != "" {
		properties.Name = &name
	}
	if size > 0 {
		properties.Size = &size
	}

	result, _, err := client.VolumesApi.DatacentersVolumesPatch(ctx, datacenterID, volumeID).Volume(properties).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to update volume: %w", err)
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal volume: %w", err)
	}

	return string(data), nil
}

func (s *Server) deleteVolume(client *ionoscloud.APIClient, ctx context.Context, datacenterID, volumeID string) (string, error) {
	_, err := client.VolumesApi.DatacentersVolumesDelete(ctx, datacenterID, volumeID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to delete volume: %w", err)
	}

	return fmt.Sprintf(`{"status": "deleted", "volume_id": "%s"}`, volumeID), nil
}

func (s *Server) attachVolume(client *ionoscloud.APIClient, ctx context.Context, datacenterID, serverID, volumeID string) (string, error) {
	volume := ionoscloud.Volume{
		Id: &volumeID,
	}

	result, _, err := client.ServersApi.DatacentersServersVolumesPost(ctx, datacenterID, serverID).Volume(volume).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to attach volume: %w", err)
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal volume: %w", err)
	}

	return string(data), nil
}

func (s *Server) detachVolume(client *ionoscloud.APIClient, ctx context.Context, datacenterID, serverID, volumeID string) (string, error) {
	_, err := client.ServersApi.DatacentersServersVolumesDelete(ctx, datacenterID, serverID, volumeID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to detach volume: %w", err)
	}

	return fmt.Sprintf(`{"status": "detached", "volume_id": "%s", "server_id": "%s"}`, volumeID, serverID), nil
}

func (s *Server) listImages(client *ionoscloud.APIClient, ctx context.Context) (string, error) {
	images, _, err := client.ImagesApi.ImagesGet(ctx).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list images: %w", err)
	}

	data, err := json.MarshalIndent(images, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal images: %w", err)
	}

	return string(data), nil
}

func (s *Server) listLocations(client *ionoscloud.APIClient, ctx context.Context) (string, error) {
	locations, _, err := client.LocationsApi.LocationsGet(ctx).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list locations: %w", err)
	}

	data, err := json.MarshalIndent(locations, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal locations: %w", err)
	}

	return string(data), nil
}

func (s *Server) listSnapshots(client *ionoscloud.APIClient, ctx context.Context) (string, error) {
	snapshots, _, err := client.SnapshotsApi.SnapshotsGet(ctx).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list snapshots: %w", err)
	}

	data, err := json.MarshalIndent(snapshots, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal snapshots: %w", err)
	}

	return string(data), nil
}

func (s *Server) getSnapshot(client *ionoscloud.APIClient, ctx context.Context, snapshotID string) (string, error) {
	snapshot, _, err := client.SnapshotsApi.SnapshotsFindById(ctx, snapshotID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get snapshot: %w", err)
	}

	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	return string(data), nil
}

func (s *Server) createSnapshot(client *ionoscloud.APIClient, ctx context.Context, datacenterID, volumeID, name, description string) (string, error) {
	properties := ionoscloud.CreateSnapshotProperties{}
	if name != "" {
		properties.Name = &name
	}
	if description != "" {
		properties.Description = &description
	}

	snapshot := ionoscloud.CreateSnapshot{
		Properties: &properties,
	}

	result, _, err := client.VolumesApi.DatacentersVolumesCreateSnapshotPost(ctx, datacenterID, volumeID).Snapshot(snapshot).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to create snapshot: %w", err)
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	return string(data), nil
}

func (s *Server) updateSnapshot(client *ionoscloud.APIClient, ctx context.Context, snapshotID, name, description string) (string, error) {
	properties := ionoscloud.SnapshotProperties{}
	if name != "" {
		properties.Name = &name
	}
	if description != "" {
		properties.Description = &description
	}

	result, _, err := client.SnapshotsApi.SnapshotsPatch(ctx, snapshotID).Snapshot(properties).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to update snapshot: %w", err)
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	return string(data), nil
}

func (s *Server) deleteSnapshot(client *ionoscloud.APIClient, ctx context.Context, snapshotID string) (string, error) {
	_, err := client.SnapshotsApi.SnapshotsDelete(ctx, snapshotID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to delete snapshot: %w", err)
	}

	return fmt.Sprintf(`{"status": "deleted", "snapshot_id": "%s"}`, snapshotID), nil
}

func (s *Server) restoreSnapshot(client *ionoscloud.APIClient, ctx context.Context, datacenterID, volumeID, snapshotID string) (string, error) {
	properties := ionoscloud.RestoreSnapshotProperties{
		SnapshotId: &snapshotID,
	}
	restoreSnapshot := ionoscloud.RestoreSnapshot{
		Properties: &properties,
	}

	_, err := client.VolumesApi.DatacentersVolumesRestoreSnapshotPost(ctx, datacenterID, volumeID).RestoreSnapshot(restoreSnapshot).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to restore snapshot: %w", err)
	}

	return fmt.Sprintf(`{"status": "restoring", "volume_id": "%s", "snapshot_id": "%s"}`, volumeID, snapshotID), nil
}

// =============================================================================
// Networking - LANs
// =============================================================================

func (s *Server) listLans(client *ionoscloud.APIClient, ctx context.Context, datacenterID string) (string, error) {
	lans, _, err := client.LANsApi.DatacentersLansGet(ctx, datacenterID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list LANs: %w", err)
	}

	data, err := json.MarshalIndent(lans, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal LANs: %w", err)
	}

	return string(data), nil
}

func (s *Server) getLan(client *ionoscloud.APIClient, ctx context.Context, datacenterID, lanID string) (string, error) {
	lan, _, err := client.LANsApi.DatacentersLansFindById(ctx, datacenterID, lanID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get LAN: %w", err)
	}

	data, err := json.MarshalIndent(lan, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal LAN: %w", err)
	}

	return string(data), nil
}

// =============================================================================
// Networking - NICs
// =============================================================================

func (s *Server) listNics(client *ionoscloud.APIClient, ctx context.Context, datacenterID, serverID string) (string, error) {
	nics, _, err := client.NetworkInterfacesApi.DatacentersServersNicsGet(ctx, datacenterID, serverID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list NICs: %w", err)
	}

	data, err := json.MarshalIndent(nics, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal NICs: %w", err)
	}

	return string(data), nil
}

func (s *Server) getNic(client *ionoscloud.APIClient, ctx context.Context, datacenterID, serverID, nicID string) (string, error) {
	nic, _, err := client.NetworkInterfacesApi.DatacentersServersNicsFindById(ctx, datacenterID, serverID, nicID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get NIC: %w", err)
	}

	data, err := json.MarshalIndent(nic, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal NIC: %w", err)
	}

	return string(data), nil
}

// =============================================================================
// Networking - IP Blocks
// =============================================================================

func (s *Server) listIpBlocks(client *ionoscloud.APIClient, ctx context.Context) (string, error) {
	ipblocks, _, err := client.IPBlocksApi.IpblocksGet(ctx).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list IP blocks: %w", err)
	}

	data, err := json.MarshalIndent(ipblocks, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal IP blocks: %w", err)
	}

	return string(data), nil
}

func (s *Server) getIpBlock(client *ionoscloud.APIClient, ctx context.Context, ipblockID string) (string, error) {
	ipblock, _, err := client.IPBlocksApi.IpblocksFindById(ctx, ipblockID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get IP block: %w", err)
	}

	data, err := json.MarshalIndent(ipblock, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal IP block: %w", err)
	}

	return string(data), nil
}

// =============================================================================
// Networking - Firewall Rules
// =============================================================================

func (s *Server) listFirewallRules(client *ionoscloud.APIClient, ctx context.Context, datacenterID, serverID, nicID string) (string, error) {
	rules, _, err := client.FirewallRulesApi.DatacentersServersNicsFirewallrulesGet(ctx, datacenterID, serverID, nicID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list firewall rules: %w", err)
	}

	data, err := json.MarshalIndent(rules, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal firewall rules: %w", err)
	}

	return string(data), nil
}

func (s *Server) getFirewallRule(client *ionoscloud.APIClient, ctx context.Context, datacenterID, serverID, nicID, firewallRuleID string) (string, error) {
	rule, _, err := client.FirewallRulesApi.DatacentersServersNicsFirewallrulesFindById(ctx, datacenterID, serverID, nicID, firewallRuleID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get firewall rule: %w", err)
	}

	data, err := json.MarshalIndent(rule, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal firewall rule: %w", err)
	}

	return string(data), nil
}

// =============================================================================
// Networking - NAT Gateways
// =============================================================================

func (s *Server) listNatGateways(client *ionoscloud.APIClient, ctx context.Context, datacenterID string) (string, error) {
	natgateways, _, err := client.NATGatewaysApi.DatacentersNatgatewaysGet(ctx, datacenterID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list NAT gateways: %w", err)
	}

	data, err := json.MarshalIndent(natgateways, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal NAT gateways: %w", err)
	}

	return string(data), nil
}

func (s *Server) getNatGateway(client *ionoscloud.APIClient, ctx context.Context, datacenterID, natGatewayID string) (string, error) {
	natgateway, _, err := client.NATGatewaysApi.DatacentersNatgatewaysFindByNatGatewayId(ctx, datacenterID, natGatewayID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get NAT gateway: %w", err)
	}

	data, err := json.MarshalIndent(natgateway, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal NAT gateway: %w", err)
	}

	return string(data), nil
}

// =============================================================================
// Networking - Private Cross Connects
// =============================================================================

func (s *Server) listPccs(client *ionoscloud.APIClient, ctx context.Context) (string, error) {
	pccs, _, err := client.PrivateCrossConnectsApi.PccsGet(ctx).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list Private Cross Connects: %w", err)
	}

	data, err := json.MarshalIndent(pccs, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal Private Cross Connects: %w", err)
	}

	return string(data), nil
}

func (s *Server) getPcc(client *ionoscloud.APIClient, ctx context.Context, pccID string) (string, error) {
	pcc, _, err := client.PrivateCrossConnectsApi.PccsFindById(ctx, pccID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get Private Cross Connect: %w", err)
	}

	data, err := json.MarshalIndent(pcc, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal Private Cross Connect: %w", err)
	}

	return string(data), nil
}

// =============================================================================
// Load Balancers - Application Load Balancers
// =============================================================================

func (s *Server) listApplicationLoadBalancers(client *ionoscloud.APIClient, ctx context.Context, datacenterID string) (string, error) {
	albs, _, err := client.ApplicationLoadBalancersApi.DatacentersApplicationloadbalancersGet(ctx, datacenterID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list Application Load Balancers: %w", err)
	}

	data, err := json.MarshalIndent(albs, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal Application Load Balancers: %w", err)
	}

	return string(data), nil
}

func (s *Server) getApplicationLoadBalancer(client *ionoscloud.APIClient, ctx context.Context, datacenterID, albID string) (string, error) {
	alb, _, err := client.ApplicationLoadBalancersApi.DatacentersApplicationloadbalancersFindByApplicationLoadBalancerId(ctx, datacenterID, albID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get Application Load Balancer: %w", err)
	}

	data, err := json.MarshalIndent(alb, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal Application Load Balancer: %w", err)
	}

	return string(data), nil
}

func (s *Server) listAlbForwardingRules(client *ionoscloud.APIClient, ctx context.Context, datacenterID, albID string) (string, error) {
	rules, _, err := client.ApplicationLoadBalancersApi.DatacentersApplicationloadbalancersForwardingrulesGet(ctx, datacenterID, albID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list ALB forwarding rules: %w", err)
	}

	data, err := json.MarshalIndent(rules, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal ALB forwarding rules: %w", err)
	}

	return string(data), nil
}

func (s *Server) getAlbForwardingRule(client *ionoscloud.APIClient, ctx context.Context, datacenterID, albID, ruleID string) (string, error) {
	rule, _, err := client.ApplicationLoadBalancersApi.DatacentersApplicationloadbalancersForwardingrulesFindByForwardingRuleId(ctx, datacenterID, albID, ruleID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get ALB forwarding rule: %w", err)
	}

	data, err := json.MarshalIndent(rule, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal ALB forwarding rule: %w", err)
	}

	return string(data), nil
}

// =============================================================================
// Load Balancers - Network Load Balancers
// =============================================================================

func (s *Server) listNetworkLoadBalancers(client *ionoscloud.APIClient, ctx context.Context, datacenterID string) (string, error) {
	nlbs, _, err := client.NetworkLoadBalancersApi.DatacentersNetworkloadbalancersGet(ctx, datacenterID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list Network Load Balancers: %w", err)
	}

	data, err := json.MarshalIndent(nlbs, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal Network Load Balancers: %w", err)
	}

	return string(data), nil
}

func (s *Server) getNetworkLoadBalancer(client *ionoscloud.APIClient, ctx context.Context, datacenterID, nlbID string) (string, error) {
	nlb, _, err := client.NetworkLoadBalancersApi.DatacentersNetworkloadbalancersFindByNetworkLoadBalancerId(ctx, datacenterID, nlbID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get Network Load Balancer: %w", err)
	}

	data, err := json.MarshalIndent(nlb, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal Network Load Balancer: %w", err)
	}

	return string(data), nil
}

// =============================================================================
// Load Balancers - Target Groups
// =============================================================================

func (s *Server) listTargetGroups(client *ionoscloud.APIClient, ctx context.Context) (string, error) {
	targetGroups, _, err := client.TargetGroupsApi.TargetgroupsGet(ctx).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list target groups: %w", err)
	}

	data, err := json.MarshalIndent(targetGroups, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal target groups: %w", err)
	}

	return string(data), nil
}

func (s *Server) getTargetGroup(client *ionoscloud.APIClient, ctx context.Context, targetGroupID string) (string, error) {
	targetGroup, _, err := client.TargetGroupsApi.TargetgroupsFindByTargetGroupId(ctx, targetGroupID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get target group: %w", err)
	}

	data, err := json.MarshalIndent(targetGroup, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal target group: %w", err)
	}

	return string(data), nil
}

// =============================================================================
// Kubernetes
// =============================================================================

func (s *Server) listK8sClusters(client *ionoscloud.APIClient, ctx context.Context) (string, error) {
	clusters, _, err := client.KubernetesApi.K8sGet(ctx).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list Kubernetes clusters: %w", err)
	}

	data, err := json.MarshalIndent(clusters, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal Kubernetes clusters: %w", err)
	}

	return string(data), nil
}

func (s *Server) getK8sCluster(client *ionoscloud.APIClient, ctx context.Context, k8sClusterID string) (string, error) {
	cluster, _, err := client.KubernetesApi.K8sFindByClusterId(ctx, k8sClusterID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get Kubernetes cluster: %w", err)
	}

	data, err := json.MarshalIndent(cluster, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal Kubernetes cluster: %w", err)
	}

	return string(data), nil
}

func (s *Server) getK8sKubeconfig(client *ionoscloud.APIClient, ctx context.Context, k8sClusterID string) (string, error) {
	kubeconfig, _, err := client.KubernetesApi.K8sKubeconfigGet(ctx, k8sClusterID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get kubeconfig: %w", err)
	}

	data, err := json.MarshalIndent(kubeconfig, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal kubeconfig: %w", err)
	}

	return string(data), nil
}

func (s *Server) listK8sNodepools(client *ionoscloud.APIClient, ctx context.Context, k8sClusterID string) (string, error) {
	nodepools, _, err := client.KubernetesApi.K8sNodepoolsGet(ctx, k8sClusterID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list node pools: %w", err)
	}

	data, err := json.MarshalIndent(nodepools, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal node pools: %w", err)
	}

	return string(data), nil
}

func (s *Server) getK8sNodepool(client *ionoscloud.APIClient, ctx context.Context, k8sClusterID, nodepoolID string) (string, error) {
	nodepool, _, err := client.KubernetesApi.K8sNodepoolsFindById(ctx, k8sClusterID, nodepoolID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get node pool: %w", err)
	}

	data, err := json.MarshalIndent(nodepool, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal node pool: %w", err)
	}

	return string(data), nil
}

func (s *Server) listK8sNodes(client *ionoscloud.APIClient, ctx context.Context, k8sClusterID, nodepoolID string) (string, error) {
	nodes, _, err := client.KubernetesApi.K8sNodepoolsNodesGet(ctx, k8sClusterID, nodepoolID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list nodes: %w", err)
	}

	data, err := json.MarshalIndent(nodes, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal nodes: %w", err)
	}

	return string(data), nil
}

func (s *Server) getK8sNode(client *ionoscloud.APIClient, ctx context.Context, k8sClusterID, nodepoolID, nodeID string) (string, error) {
	node, _, err := client.KubernetesApi.K8sNodepoolsNodesFindById(ctx, k8sClusterID, nodepoolID, nodeID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get node: %w", err)
	}

	data, err := json.MarshalIndent(node, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal node: %w", err)
	}

	return string(data), nil
}

func (s *Server) listK8sVersions(client *ionoscloud.APIClient, ctx context.Context) (string, error) {
	versions, _, err := client.KubernetesApi.K8sVersionsGet(ctx).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list Kubernetes versions: %w", err)
	}

	data, err := json.MarshalIndent(versions, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal Kubernetes versions: %w", err)
	}

	return string(data), nil
}

// =============================================================================
// User Management - Users
// =============================================================================

func (s *Server) listUsers(client *ionoscloud.APIClient, ctx context.Context) (string, error) {
	users, _, err := client.UserManagementApi.UmUsersGet(ctx).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list users: %w", err)
	}

	data, err := json.MarshalIndent(users, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal users: %w", err)
	}

	return string(data), nil
}

func (s *Server) getUser(client *ionoscloud.APIClient, ctx context.Context, userID string) (string, error) {
	user, _, err := client.UserManagementApi.UmUsersFindById(ctx, userID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get user: %w", err)
	}

	data, err := json.MarshalIndent(user, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal user: %w", err)
	}

	return string(data), nil
}

// =============================================================================
// User Management - Groups
// =============================================================================

func (s *Server) listGroups(client *ionoscloud.APIClient, ctx context.Context) (string, error) {
	groups, _, err := client.UserManagementApi.UmGroupsGet(ctx).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list groups: %w", err)
	}

	data, err := json.MarshalIndent(groups, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal groups: %w", err)
	}

	return string(data), nil
}

func (s *Server) getGroup(client *ionoscloud.APIClient, ctx context.Context, groupID string) (string, error) {
	group, _, err := client.UserManagementApi.UmGroupsFindById(ctx, groupID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get group: %w", err)
	}

	data, err := json.MarshalIndent(group, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal group: %w", err)
	}

	return string(data), nil
}

func (s *Server) listGroupMembers(client *ionoscloud.APIClient, ctx context.Context, groupID string) (string, error) {
	members, _, err := client.UserManagementApi.UmGroupsUsersGet(ctx, groupID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list group members: %w", err)
	}

	data, err := json.MarshalIndent(members, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal group members: %w", err)
	}

	return string(data), nil
}

func (s *Server) listUserGroups(client *ionoscloud.APIClient, ctx context.Context, userID string) (string, error) {
	groups, _, err := client.UserManagementApi.UmUsersGroupsGet(ctx, userID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list user groups: %w", err)
	}

	data, err := json.MarshalIndent(groups, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal user groups: %w", err)
	}

	return string(data), nil
}

// =============================================================================
// User Management - S3 Keys
// =============================================================================

func (s *Server) listS3Keys(client *ionoscloud.APIClient, ctx context.Context, userID string) (string, error) {
	keys, _, err := client.UserS3KeysApi.UmUsersS3keysGet(ctx, userID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list S3 keys: %w", err)
	}

	data, err := json.MarshalIndent(keys, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal S3 keys: %w", err)
	}

	return string(data), nil
}

func (s *Server) getS3Key(client *ionoscloud.APIClient, ctx context.Context, userID, keyID string) (string, error) {
	key, _, err := client.UserS3KeysApi.UmUsersS3keysFindByKeyId(ctx, userID, keyID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get S3 key: %w", err)
	}

	data, err := json.MarshalIndent(key, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal S3 key: %w", err)
	}

	return string(data), nil
}

// =============================================================================
// User Management - Contract
// =============================================================================

func (s *Server) getContract(client *ionoscloud.APIClient, ctx context.Context) (string, error) {
	contract, _, err := client.ContractResourcesApi.ContractsGet(ctx).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get contract: %w", err)
	}

	data, err := json.MarshalIndent(contract, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal contract: %w", err)
	}

	return string(data), nil
}

func (s *Server) listResources(client *ionoscloud.APIClient, ctx context.Context, resourceType string) (string, error) {
	var resources ionoscloud.Resources
	var err error

	if resourceType != "" {
		resources, _, err = client.UserManagementApi.UmResourcesFindByType(ctx, resourceType).Execute()
	} else {
		resources, _, err = client.UserManagementApi.UmResourcesGet(ctx).Execute()
	}

	if err != nil {
		return "", fmt.Errorf("failed to list resources: %w", err)
	}

	data, err := json.MarshalIndent(resources, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal resources: %w", err)
	}

	return string(data), nil
}

// =============================================================================
// DNS
// =============================================================================

func (s *Server) listDnsZones(client *dns.APIClient, ctx context.Context) (string, error) {
	zones, _, err := client.ZonesApi.ZonesGet(ctx).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list DNS zones: %w", err)
	}

	data, err := json.MarshalIndent(zones, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal DNS zones: %w", err)
	}

	return string(data), nil
}

func (s *Server) getDnsZone(client *dns.APIClient, ctx context.Context, zoneID string) (string, error) {
	zone, _, err := client.ZonesApi.ZonesFindById(ctx, zoneID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get DNS zone: %w", err)
	}

	data, err := json.MarshalIndent(zone, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal DNS zone: %w", err)
	}

	return string(data), nil
}

func (s *Server) listDnsRecords(client *dns.APIClient, ctx context.Context, zoneID string) (string, error) {
	records, _, err := client.RecordsApi.ZonesRecordsGet(ctx, zoneID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list DNS records: %w", err)
	}

	data, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal DNS records: %w", err)
	}

	return string(data), nil
}

func (s *Server) getDnsRecord(client *dns.APIClient, ctx context.Context, zoneID, recordID string) (string, error) {
	record, _, err := client.RecordsApi.ZonesRecordsFindById(ctx, zoneID, recordID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get DNS record: %w", err)
	}

	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal DNS record: %w", err)
	}

	return string(data), nil
}
