package test

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestComputeToolEndpoints(t *testing.T) {
	h := setup(t)

	dc := "dc-1"
	srv := "srv-1"
	nic := "nic-1"

	tests := []toolTest{
		// Datacenters
		{"list_datacenters", map[string]any{}, []string{"GET"}, []string{"/cloudapi/v6/datacenters"}},
		{"get_datacenter", map[string]any{"datacenter_id": dc}, []string{"GET"}, []string{"/cloudapi/v6/datacenters/" + dc}},

		// Servers
		{"list_servers", map[string]any{"datacenter_id": dc}, []string{"GET"}, []string{"/cloudapi/v6/datacenters/" + dc + "/servers"}},
		{"get_server", map[string]any{"datacenter_id": dc, "server_id": srv}, []string{"GET"}, []string{"/cloudapi/v6/datacenters/" + dc + "/servers/" + srv}},
		{"list_server_volumes", map[string]any{"datacenter_id": dc, "server_id": srv}, []string{"GET"}, []string{"/cloudapi/v6/datacenters/" + dc + "/servers/" + srv + "/volumes"}},
		{"list_server_cdroms", map[string]any{"datacenter_id": dc, "server_id": srv}, []string{"GET"}, []string{"/cloudapi/v6/datacenters/" + dc + "/servers/" + srv + "/cdroms"}},
		{"list_server_gpus", map[string]any{"datacenter_id": dc, "server_id": srv}, []string{"GET"}, []string{"/cloudapi/v6/datacenters/" + dc + "/servers/" + srv + "/gpus"}},
		{"get_server_gpu", map[string]any{"datacenter_id": dc, "server_id": srv, "gpu_id": "gpu-1"}, []string{"GET"}, []string{"/cloudapi/v6/datacenters/" + dc + "/servers/" + srv + "/gpus/gpu-1"}},
		{"get_server_remote_console", map[string]any{"datacenter_id": dc, "server_id": srv}, []string{"GET"}, []string{"/cloudapi/v6/datacenters/" + dc + "/servers/" + srv + "/remoteconsole"}},

		// Volumes
		{"list_volumes", map[string]any{"datacenter_id": dc}, []string{"GET"}, []string{"/cloudapi/v6/datacenters/" + dc + "/volumes"}},
		{"get_volume", map[string]any{"datacenter_id": dc, "volume_id": "vol-1"}, []string{"GET"}, []string{"/cloudapi/v6/datacenters/" + dc + "/volumes/vol-1"}},

		// NICs
		{"list_nics", map[string]any{"datacenter_id": dc, "server_id": srv}, []string{"GET"}, []string{"/cloudapi/v6/datacenters/" + dc + "/servers/" + srv + "/nics"}},
		{"get_nic", map[string]any{"datacenter_id": dc, "server_id": srv, "nic_id": nic}, []string{"GET"}, []string{"/cloudapi/v6/datacenters/" + dc + "/servers/" + srv + "/nics/" + nic}},

		// LANs
		{"list_lans", map[string]any{"datacenter_id": dc}, []string{"GET"}, []string{"/cloudapi/v6/datacenters/" + dc + "/lans"}},
		{"get_lan", map[string]any{"datacenter_id": dc, "lan_id": "lan-1"}, []string{"GET"}, []string{"/cloudapi/v6/datacenters/" + dc + "/lans/lan-1"}},
		{"list_lan_nics", map[string]any{"datacenter_id": dc, "lan_id": "lan-1"}, []string{"GET"}, []string{"/cloudapi/v6/datacenters/" + dc + "/lans/lan-1/nics"}},

		// Firewall Rules
		{"list_firewall_rules", map[string]any{"datacenter_id": dc, "server_id": srv, "nic_id": nic}, []string{"GET"}, []string{"/cloudapi/v6/datacenters/" + dc + "/servers/" + srv + "/nics/" + nic + "/firewallrules"}},
		{"get_firewall_rule", map[string]any{"datacenter_id": dc, "server_id": srv, "nic_id": nic, "firewallrule_id": "fw-1"}, []string{"GET"}, []string{"/cloudapi/v6/datacenters/" + dc + "/servers/" + srv + "/nics/" + nic + "/firewallrules/fw-1"}},

		// IP Blocks
		{"list_ip_blocks", map[string]any{}, []string{"GET"}, []string{"/cloudapi/v6/ipblocks"}},
		{"get_ip_block", map[string]any{"ipblock_id": "ip-1"}, []string{"GET"}, []string{"/cloudapi/v6/ipblocks/ip-1"}},

		// Load Balancers
		{"list_loadbalancers", map[string]any{"datacenter_id": dc}, []string{"GET"}, []string{"/cloudapi/v6/datacenters/" + dc + "/loadbalancers"}},
		{"get_loadbalancer", map[string]any{"datacenter_id": dc, "loadbalancer_id": "lb-1"}, []string{"GET"}, []string{"/cloudapi/v6/datacenters/" + dc + "/loadbalancers/lb-1"}},
		{"list_loadbalancer_nics", map[string]any{"datacenter_id": dc, "loadbalancer_id": "lb-1"}, []string{"GET"}, []string{"/cloudapi/v6/datacenters/" + dc + "/loadbalancers/lb-1/balancednics"}},

		// Network Load Balancers
		{"list_network_loadbalancers", map[string]any{"datacenter_id": dc}, []string{"GET"}, []string{"/cloudapi/v6/datacenters/" + dc + "/networkloadbalancers"}},
		{"get_network_loadbalancer", map[string]any{"datacenter_id": dc, "network_loadbalancer_id": "nlb-1"}, []string{"GET"}, []string{"/cloudapi/v6/datacenters/" + dc + "/networkloadbalancers/nlb-1"}},
		{"list_nlb_forwarding_rules", map[string]any{"datacenter_id": dc, "network_loadbalancer_id": "nlb-1"}, []string{"GET"}, []string{"/cloudapi/v6/datacenters/" + dc + "/networkloadbalancers/nlb-1/forwardingrules"}},

		// Application Load Balancers
		{"list_application_loadbalancers", map[string]any{"datacenter_id": dc}, []string{"GET"}, []string{"/cloudapi/v6/datacenters/" + dc + "/applicationloadbalancers"}},
		{"get_application_loadbalancer", map[string]any{"datacenter_id": dc, "application_loadbalancer_id": "alb-1"}, []string{"GET"}, []string{"/cloudapi/v6/datacenters/" + dc + "/applicationloadbalancers/alb-1"}},
		{"list_alb_forwarding_rules", map[string]any{"datacenter_id": dc, "application_loadbalancer_id": "alb-1"}, []string{"GET"}, []string{"/cloudapi/v6/datacenters/" + dc + "/applicationloadbalancers/alb-1/forwardingrules"}},

		// Target Groups
		{"list_target_groups", map[string]any{}, []string{"GET"}, []string{"/cloudapi/v6/targetgroups"}},
		{"get_target_group", map[string]any{"target_group_id": "tg-1"}, []string{"GET"}, []string{"/cloudapi/v6/targetgroups/tg-1"}},

		// NAT Gateways
		{"list_nat_gateways", map[string]any{"datacenter_id": dc}, []string{"GET"}, []string{"/cloudapi/v6/datacenters/" + dc + "/natgateways"}},
		{"get_nat_gateway", map[string]any{"datacenter_id": dc, "nat_gateway_id": "nat-1"}, []string{"GET"}, []string{"/cloudapi/v6/datacenters/" + dc + "/natgateways/nat-1"}},
		{"list_nat_gateway_rules", map[string]any{"datacenter_id": dc, "nat_gateway_id": "nat-1"}, []string{"GET"}, []string{"/cloudapi/v6/datacenters/" + dc + "/natgateways/nat-1/rules"}},

		// Private Cross-Connects
		{"list_private_cross_connects", map[string]any{}, []string{"GET"}, []string{"/cloudapi/v6/pccs"}},
		{"get_private_cross_connect", map[string]any{"pcc_id": "pcc-1"}, []string{"GET"}, []string{"/cloudapi/v6/pccs/pcc-1"}},

		// Security Groups
		{"list_security_groups", map[string]any{"datacenter_id": dc}, []string{"GET"}, []string{"/cloudapi/v6/datacenters/" + dc + "/securitygroups"}},
		{"get_security_group", map[string]any{"datacenter_id": dc, "security_group_id": "sg-1"}, []string{"GET"}, []string{"/cloudapi/v6/datacenters/" + dc + "/securitygroups/sg-1"}},
		{"list_security_group_rules", map[string]any{"datacenter_id": dc, "security_group_id": "sg-1"}, []string{"GET"}, []string{"/cloudapi/v6/datacenters/" + dc + "/securitygroups/sg-1/rules"}},
		{"get_security_group_rule", map[string]any{"datacenter_id": dc, "security_group_id": "sg-1", "rule_id": "rule-1"}, []string{"GET"}, []string{"/cloudapi/v6/datacenters/" + dc + "/securitygroups/sg-1/rules/rule-1"}},

		// Contract
		{"get_contract", map[string]any{}, []string{"GET"}, []string{"/cloudapi/v6/contracts"}},

		// Requests
		{"list_requests", map[string]any{}, []string{"GET"}, []string{"/cloudapi/v6/requests"}},
		{"get_request", map[string]any{"request_id": "req-1"}, []string{"GET"}, []string{"/cloudapi/v6/requests/req-1"}},
		{"get_request_status", map[string]any{"request_id": "req-1"}, []string{"GET"}, []string{"/cloudapi/v6/requests/req-1/status"}},

		// Templates
		{"list_templates", map[string]any{}, []string{"GET"}, []string{"/cloudapi/v6/templates"}},
		{"get_template", map[string]any{"template_id": "tpl-1"}, []string{"GET"}, []string{"/cloudapi/v6/templates/tpl-1"}},

		// Images
		{"list_images", map[string]any{}, []string{"GET"}, []string{"/cloudapi/v6/images"}},

		// Locations
		{"list_locations", map[string]any{}, []string{"GET"}, []string{"/cloudapi/v6/locations"}},

		// Snapshots
		{"list_snapshots", map[string]any{}, []string{"GET"}, []string{"/cloudapi/v6/snapshots"}},
		{"get_snapshot", map[string]any{"snapshot_id": "snap-1"}, []string{"GET"}, []string{"/cloudapi/v6/snapshots/snap-1"}},

		// Kubernetes
		{"list_k8s_clusters", map[string]any{}, []string{"GET"}, []string{"/cloudapi/v6/k8s"}},
		{"get_k8s_cluster", map[string]any{"k8s_cluster_id": "k8s-1"}, []string{"GET"}, []string{"/cloudapi/v6/k8s/k8s-1"}},
		{"list_k8s_nodepools", map[string]any{"k8s_cluster_id": "k8s-1"}, []string{"GET"}, []string{"/cloudapi/v6/k8s/k8s-1/nodepools"}},
		{"get_k8s_nodepool", map[string]any{"k8s_cluster_id": "k8s-1", "nodepool_id": "np-1"}, []string{"GET"}, []string{"/cloudapi/v6/k8s/k8s-1/nodepools/np-1"}},
		{"list_k8s_nodepool_nodes", map[string]any{"k8s_cluster_id": "k8s-1", "nodepool_id": "np-1"}, []string{"GET"}, []string{"/cloudapi/v6/k8s/k8s-1/nodepools/np-1/nodes"}},
		{"get_k8s_node", map[string]any{"k8s_cluster_id": "k8s-1", "nodepool_id": "np-1", "node_id": "node-1"}, []string{"GET"}, []string{"/cloudapi/v6/k8s/k8s-1/nodepools/np-1/nodes/node-1"}},
		{"list_k8s_versions", map[string]any{}, []string{"GET"}, []string{"/cloudapi/v6/k8s/versions"}},
		{"get_k8s_default_version", map[string]any{}, []string{"GET"}, []string{"/cloudapi/v6/k8s/versions/default"}},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h.log.clear()

			_, err := h.session.CallTool(ctx, &mcp.CallToolParams{
				Name:      tt.name,
				Arguments: tt.args,
			})
			if err != nil {
				t.Fatalf("CallTool(%q) returned protocol error: %v", tt.name, err)
			}

			reqs := h.log.allRequests()
			if len(tt.wantMethods) != len(tt.wantPaths) {
				t.Fatalf("test %q: wantMethods has %d entries, wantPaths has %d", tt.name, len(tt.wantMethods), len(tt.wantPaths))
			}
			if len(reqs) != len(tt.wantPaths) {
				t.Fatalf("CallTool(%q) made %d requests, want %d", tt.name, len(reqs), len(tt.wantPaths))
			}
			for i, req := range reqs {
				if req.Method != tt.wantMethods[i] {
					t.Errorf("CallTool(%q) request[%d] method = %q, want %q", tt.name, i, req.Method, tt.wantMethods[i])
				}
				if req.Path != tt.wantPaths[i] {
					t.Errorf("CallTool(%q) request[%d] path = %q, want %q", tt.name, i, req.Path, tt.wantPaths[i])
				}
			}
		})
	}
}
