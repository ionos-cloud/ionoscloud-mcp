package test

import (
	"net/url"
	"testing"
)

func TestComputeToolEndpoints(t *testing.T) {
	h := setup(t)

	dc := "dc-1"
	srv := "srv-1"
	nic := "nic-1"

	tests := []toolTest{
		// Datacenters
		{name: "list_datacenters", args: map[string]any{}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters"}, wantQuery: []url.Values{{"depth": []string{"1"}}}},
		{name: "list_datacenters", args: map[string]any{"name": "my-dc"}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters"}, wantQuery: []url.Values{{"depth": []string{"1"}, "filter.properties.name": []string{"my-dc"}}}},
		{name: "get_datacenter", args: map[string]any{"datacenter_id": dc}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc}},

		// Servers
		{name: "list_servers", args: map[string]any{"datacenter_id": dc}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/servers"}},
		{name: "get_server", args: map[string]any{"datacenter_id": dc, "server_id": srv}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/servers/" + srv}},
		{name: "list_server_volumes", args: map[string]any{"datacenter_id": dc, "server_id": srv}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/servers/" + srv + "/volumes"}},
		{name: "list_server_cdroms", args: map[string]any{"datacenter_id": dc, "server_id": srv}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/servers/" + srv + "/cdroms"}},
		{name: "list_server_gpus", args: map[string]any{"datacenter_id": dc, "server_id": srv}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/servers/" + srv + "/gpus"}},
		{name: "get_server_gpu", args: map[string]any{"datacenter_id": dc, "server_id": srv, "gpu_id": "gpu-1"}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/servers/" + srv + "/gpus/gpu-1"}},
		{name: "get_server_remote_console", args: map[string]any{"datacenter_id": dc, "server_id": srv}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/servers/" + srv + "/remoteconsole"}},

		// Volumes
		{name: "list_volumes", args: map[string]any{"datacenter_id": dc}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/volumes"}},
		{name: "get_volume", args: map[string]any{"datacenter_id": dc, "volume_id": "vol-1"}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/volumes/vol-1"}},

		// NICs
		{name: "list_nics", args: map[string]any{"datacenter_id": dc, "server_id": srv}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/servers/" + srv + "/nics"}},
		{name: "get_nic", args: map[string]any{"datacenter_id": dc, "server_id": srv, "nic_id": nic}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/servers/" + srv + "/nics/" + nic}},

		// LANs
		{name: "list_lans", args: map[string]any{"datacenter_id": dc}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/lans"}},
		{name: "get_lan", args: map[string]any{"datacenter_id": dc, "lan_id": "lan-1"}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/lans/lan-1"}},
		{name: "list_lan_nics", args: map[string]any{"datacenter_id": dc, "lan_id": "lan-1"}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/lans/lan-1/nics"}},

		// Firewall Rules
		{name: "list_firewall_rules", args: map[string]any{"datacenter_id": dc, "server_id": srv, "nic_id": nic}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/servers/" + srv + "/nics/" + nic + "/firewallrules"}},
		{name: "get_firewall_rule", args: map[string]any{"datacenter_id": dc, "server_id": srv, "nic_id": nic, "firewallrule_id": "fw-1"}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/servers/" + srv + "/nics/" + nic + "/firewallrules/fw-1"}},

		// IP Blocks
		{name: "list_ip_blocks", args: map[string]any{}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/ipblocks"}},
		{name: "get_ip_block", args: map[string]any{"ipblock_id": "ip-1"}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/ipblocks/ip-1"}},

		// Load Balancers
		{name: "list_loadbalancers", args: map[string]any{"datacenter_id": dc}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/loadbalancers"}},
		{name: "get_loadbalancer", args: map[string]any{"datacenter_id": dc, "loadbalancer_id": "lb-1"}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/loadbalancers/lb-1"}},
		{name: "list_loadbalancer_nics", args: map[string]any{"datacenter_id": dc, "loadbalancer_id": "lb-1"}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/loadbalancers/lb-1/balancednics"}},

		// Network Load Balancers
		{name: "list_network_loadbalancers", args: map[string]any{"datacenter_id": dc}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/networkloadbalancers"}},
		{name: "get_network_loadbalancer", args: map[string]any{"datacenter_id": dc, "network_loadbalancer_id": "nlb-1"}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/networkloadbalancers/nlb-1"}},
		{name: "list_nlb_forwarding_rules", args: map[string]any{"datacenter_id": dc, "network_loadbalancer_id": "nlb-1"}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/networkloadbalancers/nlb-1/forwardingrules"}},

		// Application Load Balancers
		{name: "list_application_loadbalancers", args: map[string]any{"datacenter_id": dc}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/applicationloadbalancers"}},
		{name: "get_application_loadbalancer", args: map[string]any{"datacenter_id": dc, "application_loadbalancer_id": "alb-1"}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/applicationloadbalancers/alb-1"}},
		{name: "list_alb_forwarding_rules", args: map[string]any{"datacenter_id": dc, "application_loadbalancer_id": "alb-1"}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/applicationloadbalancers/alb-1/forwardingrules"}},

		// Target Groups
		{name: "list_target_groups", args: map[string]any{}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/targetgroups"}},
		{name: "get_target_group", args: map[string]any{"target_group_id": "tg-1"}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/targetgroups/tg-1"}},

		// NAT Gateways
		{name: "list_nat_gateways", args: map[string]any{"datacenter_id": dc}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/natgateways"}},
		{name: "get_nat_gateway", args: map[string]any{"datacenter_id": dc, "nat_gateway_id": "nat-1"}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/natgateways/nat-1"}},
		{name: "list_nat_gateway_rules", args: map[string]any{"datacenter_id": dc, "nat_gateway_id": "nat-1"}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/natgateways/nat-1/rules"}},

		// Private Cross-Connects
		{name: "list_private_cross_connects", args: map[string]any{}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/pccs"}},
		{name: "get_private_cross_connect", args: map[string]any{"pcc_id": "pcc-1"}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/pccs/pcc-1"}},

		// Security Groups
		{name: "list_security_groups", args: map[string]any{"datacenter_id": dc}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/securitygroups"}},
		{name: "get_security_group", args: map[string]any{"datacenter_id": dc, "security_group_id": "sg-1"}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/securitygroups/sg-1"}},
		{name: "list_security_group_rules", args: map[string]any{"datacenter_id": dc, "security_group_id": "sg-1"}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/securitygroups/sg-1/rules"}},
		{name: "get_security_group_rule", args: map[string]any{"datacenter_id": dc, "security_group_id": "sg-1", "rule_id": "rule-1"}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/securitygroups/sg-1/rules/rule-1"}},

		// Contract
		{name: "get_contract", args: map[string]any{}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/contracts"}},

		// Requests
		{name: "list_requests", args: map[string]any{}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/requests"}},
		{name: "get_request", args: map[string]any{"request_id": "req-1"}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/requests/req-1"}},
		{name: "get_request_status", args: map[string]any{"request_id": "req-1"}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/requests/req-1/status"}},

		// Templates
		{name: "list_templates", args: map[string]any{}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/templates"}},
		{name: "get_template", args: map[string]any{"template_id": "tpl-1"}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/templates/tpl-1"}},

		// Images
		{name: "list_images", args: map[string]any{}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/images"}},

		// Locations
		{name: "list_locations", args: map[string]any{}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/locations"}},

		// Snapshots
		{name: "list_snapshots", args: map[string]any{}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/snapshots"}},
		{name: "get_snapshot", args: map[string]any{"snapshot_id": "snap-1"}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/snapshots/snap-1"}},
	}

	h.run(t, tests)
}

// TestComputeOutput asserts the tool returns the API payload to the caller.
func TestComputeOutput(t *testing.T) {
	h := setup(t)

	tests := []toolTest{
		{
			name:        "get_datacenter",
			args:        map[string]any{"datacenter_id": "dc-1"},
			wantMethods: []string{"GET"},
			wantPaths:   []string{"/cloudapi/v6/datacenters/dc-1"},
			fixture:     `{"id":"dc-1","properties":{"name":"my-dc","location":"de/fra"}}`,
			wantContain: []string{`"id":"dc-1"`, "my-dc", "de/fra"},
		},
	}

	h.run(t, tests)
}
