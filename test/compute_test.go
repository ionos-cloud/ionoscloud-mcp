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

	d1 := url.Values{"depth": []string{"1"}}

	f1 := func(k, v string) url.Values {
		return url.Values{"depth": []string{"1"}, "filter." + k: []string{v}}
	}

	tests := []toolTest{
		// Datacenters
		{name: "list_datacenters", args: map[string]any{}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters"}, wantQuery: []url.Values{d1}},
		{name: "list_datacenters", args: map[string]any{"filters": map[string]any{"name": "prod", "location": "de/fra"}}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters"}, wantQuery: []url.Values{{"depth": []string{"1"}, "filter.name": []string{"prod"}, "filter.location": []string{"de/fra"}}}},
		{name: "get_datacenter", args: map[string]any{"datacenter_id": dc}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc}},

		// Servers
		{name: "list_servers", args: map[string]any{"datacenter_id": dc}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/servers"}, wantQuery: []url.Values{d1}},
		{name: "list_servers", args: map[string]any{"datacenter_id": dc, "filters": map[string]any{"vmState": "RUNNING"}}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/servers"}, wantQuery: []url.Values{f1("vmState", "RUNNING")}},
		{name: "get_server", args: map[string]any{"datacenter_id": dc, "server_id": srv}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/servers/" + srv}},
		{name: "list_server_volumes", args: map[string]any{"datacenter_id": dc, "server_id": srv}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/servers/" + srv + "/volumes"}, wantQuery: []url.Values{d1}},
		{name: "list_server_volumes", args: map[string]any{"datacenter_id": dc, "server_id": srv, "filters": map[string]any{"name": "boot"}}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/servers/" + srv + "/volumes"}, wantQuery: []url.Values{f1("name", "boot")}},
		{name: "list_server_cdroms", args: map[string]any{"datacenter_id": dc, "server_id": srv}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/servers/" + srv + "/cdroms"}, wantQuery: []url.Values{d1}},
		{name: "list_server_cdroms", args: map[string]any{"datacenter_id": dc, "server_id": srv, "filters": map[string]any{"name": "iso"}}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/servers/" + srv + "/cdroms"}, wantQuery: []url.Values{f1("name", "iso")}},
		{name: "list_server_gpus", args: map[string]any{"datacenter_id": dc, "server_id": srv}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/servers/" + srv + "/gpus"}, wantQuery: []url.Values{d1}},
		{name: "list_server_gpus", args: map[string]any{"datacenter_id": dc, "server_id": srv, "filters": map[string]any{"name": "gpu"}}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/servers/" + srv + "/gpus"}, wantQuery: []url.Values{f1("name", "gpu")}},
		{name: "get_server_gpu", args: map[string]any{"datacenter_id": dc, "server_id": srv, "gpu_id": "gpu-1"}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/servers/" + srv + "/gpus/gpu-1"}},
		{name: "get_server_remote_console", args: map[string]any{"datacenter_id": dc, "server_id": srv}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/servers/" + srv + "/remoteconsole"}},

		// Volumes
		{name: "list_volumes", args: map[string]any{"datacenter_id": dc}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/volumes"}, wantQuery: []url.Values{d1}},
		{name: "list_volumes", args: map[string]any{"datacenter_id": dc, "filters": map[string]any{"name": "data"}}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/volumes"}, wantQuery: []url.Values{f1("name", "data")}},
		{name: "get_volume", args: map[string]any{"datacenter_id": dc, "volume_id": "vol-1"}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/volumes/vol-1"}},

		// NICs
		{name: "list_nics", args: map[string]any{"datacenter_id": dc, "server_id": srv}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/servers/" + srv + "/nics"}, wantQuery: []url.Values{d1}},
		{name: "list_nics", args: map[string]any{"datacenter_id": dc, "server_id": srv, "filters": map[string]any{"name": "eth"}}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/servers/" + srv + "/nics"}, wantQuery: []url.Values{f1("name", "eth")}},
		{name: "get_nic", args: map[string]any{"datacenter_id": dc, "server_id": srv, "nic_id": nic}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/servers/" + srv + "/nics/" + nic}},

		// LANs
		{name: "list_lans", args: map[string]any{"datacenter_id": dc}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/lans"}, wantQuery: []url.Values{d1}},
		{name: "list_lans", args: map[string]any{"datacenter_id": dc, "filters": map[string]any{"name": "public"}}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/lans"}, wantQuery: []url.Values{f1("name", "public")}},
		{name: "get_lan", args: map[string]any{"datacenter_id": dc, "lan_id": "lan-1"}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/lans/lan-1"}},
		{name: "list_lan_nics", args: map[string]any{"datacenter_id": dc, "lan_id": "lan-1"}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/lans/lan-1/nics"}, wantQuery: []url.Values{d1}},
		{name: "list_lan_nics", args: map[string]any{"datacenter_id": dc, "lan_id": "lan-1", "filters": map[string]any{"name": "eth"}}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/lans/lan-1/nics"}, wantQuery: []url.Values{f1("name", "eth")}},

		// Firewall Rules
		{name: "list_firewall_rules", args: map[string]any{"datacenter_id": dc, "server_id": srv, "nic_id": nic}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/servers/" + srv + "/nics/" + nic + "/firewallrules"}, wantQuery: []url.Values{d1}},
		{name: "list_firewall_rules", args: map[string]any{"datacenter_id": dc, "server_id": srv, "nic_id": nic, "filters": map[string]any{"name": "allow-ssh"}}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/servers/" + srv + "/nics/" + nic + "/firewallrules"}, wantQuery: []url.Values{f1("name", "allow-ssh")}},
		{name: "get_firewall_rule", args: map[string]any{"datacenter_id": dc, "server_id": srv, "nic_id": nic, "firewallrule_id": "fw-1"}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/servers/" + srv + "/nics/" + nic + "/firewallrules/fw-1"}},

		// IP Blocks
		{name: "list_ip_blocks", args: map[string]any{}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/ipblocks"}, wantQuery: []url.Values{d1}},
		{name: "list_ip_blocks", args: map[string]any{"filters": map[string]any{"location": "de/fra"}}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/ipblocks"}, wantQuery: []url.Values{f1("location", "de/fra")}},
		{name: "get_ip_block", args: map[string]any{"ipblock_id": "ip-1"}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/ipblocks/ip-1"}},

		// Load Balancers
		{name: "list_loadbalancers", args: map[string]any{"datacenter_id": dc}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/loadbalancers"}, wantQuery: []url.Values{d1}},
		{name: "list_loadbalancers", args: map[string]any{"datacenter_id": dc, "filters": map[string]any{"name": "lb"}}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/loadbalancers"}, wantQuery: []url.Values{f1("name", "lb")}},
		{name: "get_loadbalancer", args: map[string]any{"datacenter_id": dc, "loadbalancer_id": "lb-1"}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/loadbalancers/lb-1"}},
		{name: "list_loadbalancer_nics", args: map[string]any{"datacenter_id": dc, "loadbalancer_id": "lb-1"}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/loadbalancers/lb-1/balancednics"}, wantQuery: []url.Values{d1}},
		{name: "list_loadbalancer_nics", args: map[string]any{"datacenter_id": dc, "loadbalancer_id": "lb-1", "filters": map[string]any{"name": "eth"}}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/loadbalancers/lb-1/balancednics"}, wantQuery: []url.Values{f1("name", "eth")}},

		// Network Load Balancers
		{name: "list_network_loadbalancers", args: map[string]any{"datacenter_id": dc}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/networkloadbalancers"}, wantQuery: []url.Values{d1}},
		{name: "list_network_loadbalancers", args: map[string]any{"datacenter_id": dc, "filters": map[string]any{"name": "nlb"}}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/networkloadbalancers"}, wantQuery: []url.Values{f1("name", "nlb")}},
		{name: "get_network_loadbalancer", args: map[string]any{"datacenter_id": dc, "network_loadbalancer_id": "nlb-1"}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/networkloadbalancers/nlb-1"}},
		{name: "list_nlb_forwarding_rules", args: map[string]any{"datacenter_id": dc, "network_loadbalancer_id": "nlb-1"}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/networkloadbalancers/nlb-1/forwardingrules"}, wantQuery: []url.Values{d1}},
		{name: "list_nlb_forwarding_rules", args: map[string]any{"datacenter_id": dc, "network_loadbalancer_id": "nlb-1", "filters": map[string]any{"name": "rule"}}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/networkloadbalancers/nlb-1/forwardingrules"}, wantQuery: []url.Values{f1("name", "rule")}},

		// Application Load Balancers
		{name: "list_application_loadbalancers", args: map[string]any{"datacenter_id": dc}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/applicationloadbalancers"}, wantQuery: []url.Values{d1}},
		{name: "list_application_loadbalancers", args: map[string]any{"datacenter_id": dc, "filters": map[string]any{"name": "alb"}}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/applicationloadbalancers"}, wantQuery: []url.Values{f1("name", "alb")}},
		{name: "get_application_loadbalancer", args: map[string]any{"datacenter_id": dc, "application_loadbalancer_id": "alb-1"}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/applicationloadbalancers/alb-1"}},
		{name: "list_alb_forwarding_rules", args: map[string]any{"datacenter_id": dc, "application_loadbalancer_id": "alb-1"}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/applicationloadbalancers/alb-1/forwardingrules"}, wantQuery: []url.Values{d1}},
		{name: "list_alb_forwarding_rules", args: map[string]any{"datacenter_id": dc, "application_loadbalancer_id": "alb-1", "filters": map[string]any{"name": "rule"}}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/applicationloadbalancers/alb-1/forwardingrules"}, wantQuery: []url.Values{f1("name", "rule")}},

		// Target Groups
		{name: "list_target_groups", args: map[string]any{}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/targetgroups"}, wantQuery: []url.Values{d1}},
		{name: "list_target_groups", args: map[string]any{"filters": map[string]any{"name": "tg"}}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/targetgroups"}, wantQuery: []url.Values{f1("name", "tg")}},
		{name: "get_target_group", args: map[string]any{"target_group_id": "tg-1"}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/targetgroups/tg-1"}},

		// NAT Gateways
		{name: "list_nat_gateways", args: map[string]any{"datacenter_id": dc}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/natgateways"}, wantQuery: []url.Values{d1}},
		{name: "list_nat_gateways", args: map[string]any{"datacenter_id": dc, "filters": map[string]any{"name": "nat"}}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/natgateways"}, wantQuery: []url.Values{f1("name", "nat")}},
		{name: "get_nat_gateway", args: map[string]any{"datacenter_id": dc, "nat_gateway_id": "nat-1"}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/natgateways/nat-1"}},
		{name: "list_nat_gateway_rules", args: map[string]any{"datacenter_id": dc, "nat_gateway_id": "nat-1"}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/natgateways/nat-1/rules"}, wantQuery: []url.Values{d1}},
		{name: "list_nat_gateway_rules", args: map[string]any{"datacenter_id": dc, "nat_gateway_id": "nat-1", "filters": map[string]any{"name": "rule"}}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/natgateways/nat-1/rules"}, wantQuery: []url.Values{f1("name", "rule")}},

		// Private Cross-Connects
		{name: "list_private_cross_connects", args: map[string]any{}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/pccs"}, wantQuery: []url.Values{d1}},
		{name: "list_private_cross_connects", args: map[string]any{"filters": map[string]any{"name": "pcc"}}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/pccs"}, wantQuery: []url.Values{f1("name", "pcc")}},
		{name: "get_private_cross_connect", args: map[string]any{"pcc_id": "pcc-1"}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/pccs/pcc-1"}},

		// Security Groups
		{name: "list_security_groups", args: map[string]any{"datacenter_id": dc}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/securitygroups"}, wantQuery: []url.Values{d1}},
		{name: "list_security_groups", args: map[string]any{"datacenter_id": dc, "filters": map[string]any{"name": "web"}}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/securitygroups"}, wantQuery: []url.Values{f1("name", "web")}},
		{name: "get_security_group", args: map[string]any{"datacenter_id": dc, "security_group_id": "sg-1"}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/securitygroups/sg-1"}},
		{name: "list_security_group_rules", args: map[string]any{"datacenter_id": dc, "security_group_id": "sg-1"}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/securitygroups/sg-1/rules"}, wantQuery: []url.Values{d1}},
		{name: "list_security_group_rules", args: map[string]any{"datacenter_id": dc, "security_group_id": "sg-1", "filters": map[string]any{"name": "rule"}}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/securitygroups/sg-1/rules"}, wantQuery: []url.Values{f1("name", "rule")}},
		{name: "get_security_group_rule", args: map[string]any{"datacenter_id": dc, "security_group_id": "sg-1", "rule_id": "rule-1"}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/datacenters/" + dc + "/securitygroups/sg-1/rules/rule-1"}},

		// Contract
		{name: "get_contract", args: map[string]any{}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/contracts"}},

		// Requests
		{name: "list_requests", args: map[string]any{}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/requests"}, wantQuery: []url.Values{d1}},
		{name: "list_requests", args: map[string]any{"filters": map[string]any{"requestStatus": "DONE"}}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/requests"}, wantQuery: []url.Values{f1("requestStatus", "DONE")}},
		{name: "get_request", args: map[string]any{"request_id": "req-1"}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/requests/req-1"}},
		{name: "get_request_status", args: map[string]any{"request_id": "req-1"}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/requests/req-1/status"}},

		// Templates
		{name: "list_templates", args: map[string]any{}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/templates"}, wantQuery: []url.Values{d1}},
		{name: "list_templates", args: map[string]any{"filters": map[string]any{"name": "cubeXS"}}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/templates"}, wantQuery: []url.Values{f1("name", "cubeXS")}},
		{name: "get_template", args: map[string]any{"template_id": "tpl-1"}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/templates/tpl-1"}},

		// Images
		{name: "list_images", args: map[string]any{}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/images"}, wantQuery: []url.Values{d1}},
		{name: "list_images", args: map[string]any{"filters": map[string]any{"imageType": "HDD"}}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/images"}, wantQuery: []url.Values{f1("imageType", "HDD")}},

		// Locations
		{name: "list_locations", args: map[string]any{}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/locations"}, wantQuery: []url.Values{d1}},
		{name: "list_locations", args: map[string]any{"filters": map[string]any{"name": "Frankfurt"}}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/locations"}, wantQuery: []url.Values{f1("name", "Frankfurt")}},

		// Snapshots
		{name: "list_snapshots", args: map[string]any{}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/snapshots"}, wantQuery: []url.Values{d1}},
		{name: "list_snapshots", args: map[string]any{"filters": map[string]any{"location": "de/fra"}}, wantMethods: []string{"GET"}, wantPaths: []string{"/cloudapi/v6/snapshots"}, wantQuery: []url.Values{f1("location", "de/fra")}},
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
