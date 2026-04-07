package test

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// All compute tool names for registration test.
var computeToolNames = []string{
	"list_datacenters", "get_datacenter",
	"list_servers", "get_server", "list_server_volumes", "list_server_cdroms",
	"list_server_gpus", "get_server_gpu", "get_server_remote_console",
	"list_volumes", "get_volume",
	"list_nics", "get_nic",
	"list_lans", "get_lan", "list_lan_nics",
	"list_firewall_rules", "get_firewall_rule",
	"list_ip_blocks", "get_ip_block",
	"list_loadbalancers", "get_loadbalancer", "list_loadbalancer_nics",
	"list_network_loadbalancers", "get_network_loadbalancer", "list_nlb_forwarding_rules",
	"list_application_loadbalancers", "get_application_loadbalancer", "list_alb_forwarding_rules",
	"list_target_groups", "get_target_group",
	"list_nat_gateways", "get_nat_gateway", "list_nat_gateway_rules",
	"list_private_cross_connects", "get_private_cross_connect",
	"list_security_groups", "get_security_group", "list_security_group_rules", "get_security_group_rule",
	"get_contract",
	"list_requests", "get_request", "get_request_status",
	"list_templates", "get_template",
	"list_images",
	"list_locations",
	"list_snapshots", "get_snapshot",
}

func TestComputeToolRegistration(t *testing.T) {
	h := setup(t)

	ctx := context.Background()
	result, err := h.session.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}

	registered := make(map[string]bool)
	for _, tool := range result.Tools {
		registered[tool.Name] = true
	}

	for _, name := range computeToolNames {
		if !registered[name] {
			t.Errorf("compute tool %q not registered", name)
		}
	}
}

func TestComputeToolEndpoints(t *testing.T) {
	h := setup(t)

	dc := "dc-1"
	srv := "srv-1"
	nic := "nic-1"

	tests := []toolTest{
		// Datacenters
		{"list_datacenters", map[string]any{}, "/cloudapi/v6/datacenters"},
		{"get_datacenter", map[string]any{"datacenter_id": dc}, "/cloudapi/v6/datacenters/" + dc},

		// Servers
		{"list_servers", map[string]any{"datacenter_id": dc}, "/cloudapi/v6/datacenters/" + dc + "/servers"},
		{"get_server", map[string]any{"datacenter_id": dc, "server_id": srv}, "/cloudapi/v6/datacenters/" + dc + "/servers/" + srv},
		{"list_server_volumes", map[string]any{"datacenter_id": dc, "server_id": srv}, "/cloudapi/v6/datacenters/" + dc + "/servers/" + srv + "/volumes"},
		{"list_server_cdroms", map[string]any{"datacenter_id": dc, "server_id": srv}, "/cloudapi/v6/datacenters/" + dc + "/servers/" + srv + "/cdroms"},
		{"list_server_gpus", map[string]any{"datacenter_id": dc, "server_id": srv}, "/cloudapi/v6/datacenters/" + dc + "/servers/" + srv + "/gpus"},
		{"get_server_gpu", map[string]any{"datacenter_id": dc, "server_id": srv, "gpu_id": "gpu-1"}, "/cloudapi/v6/datacenters/" + dc + "/servers/" + srv + "/gpus/gpu-1"},
		{"get_server_remote_console", map[string]any{"datacenter_id": dc, "server_id": srv}, "/cloudapi/v6/datacenters/" + dc + "/servers/" + srv + "/remoteconsole"},

		// Volumes
		{"list_volumes", map[string]any{"datacenter_id": dc}, "/cloudapi/v6/datacenters/" + dc + "/volumes"},
		{"get_volume", map[string]any{"datacenter_id": dc, "volume_id": "vol-1"}, "/cloudapi/v6/datacenters/" + dc + "/volumes/vol-1"},

		// NICs
		{"list_nics", map[string]any{"datacenter_id": dc, "server_id": srv}, "/cloudapi/v6/datacenters/" + dc + "/servers/" + srv + "/nics"},
		{"get_nic", map[string]any{"datacenter_id": dc, "server_id": srv, "nic_id": nic}, "/cloudapi/v6/datacenters/" + dc + "/servers/" + srv + "/nics/" + nic},

		// LANs
		{"list_lans", map[string]any{"datacenter_id": dc}, "/cloudapi/v6/datacenters/" + dc + "/lans"},
		{"get_lan", map[string]any{"datacenter_id": dc, "lan_id": "lan-1"}, "/cloudapi/v6/datacenters/" + dc + "/lans/lan-1"},
		{"list_lan_nics", map[string]any{"datacenter_id": dc, "lan_id": "lan-1"}, "/cloudapi/v6/datacenters/" + dc + "/lans/lan-1/nics"},

		// Firewall Rules
		{"list_firewall_rules", map[string]any{"datacenter_id": dc, "server_id": srv, "nic_id": nic}, "/cloudapi/v6/datacenters/" + dc + "/servers/" + srv + "/nics/" + nic + "/firewallrules"},
		{"get_firewall_rule", map[string]any{"datacenter_id": dc, "server_id": srv, "nic_id": nic, "firewallrule_id": "fw-1"}, "/cloudapi/v6/datacenters/" + dc + "/servers/" + srv + "/nics/" + nic + "/firewallrules/fw-1"},

		// IP Blocks
		{"list_ip_blocks", map[string]any{}, "/cloudapi/v6/ipblocks"},
		{"get_ip_block", map[string]any{"ipblock_id": "ip-1"}, "/cloudapi/v6/ipblocks/ip-1"},

		// Load Balancers
		{"list_loadbalancers", map[string]any{"datacenter_id": dc}, "/cloudapi/v6/datacenters/" + dc + "/loadbalancers"},
		{"get_loadbalancer", map[string]any{"datacenter_id": dc, "loadbalancer_id": "lb-1"}, "/cloudapi/v6/datacenters/" + dc + "/loadbalancers/lb-1"},
		{"list_loadbalancer_nics", map[string]any{"datacenter_id": dc, "loadbalancer_id": "lb-1"}, "/cloudapi/v6/datacenters/" + dc + "/loadbalancers/lb-1/balancednics"},

		// Network Load Balancers
		{"list_network_loadbalancers", map[string]any{"datacenter_id": dc}, "/cloudapi/v6/datacenters/" + dc + "/networkloadbalancers"},
		{"get_network_loadbalancer", map[string]any{"datacenter_id": dc, "network_loadbalancer_id": "nlb-1"}, "/cloudapi/v6/datacenters/" + dc + "/networkloadbalancers/nlb-1"},
		{"list_nlb_forwarding_rules", map[string]any{"datacenter_id": dc, "network_loadbalancer_id": "nlb-1"}, "/cloudapi/v6/datacenters/" + dc + "/networkloadbalancers/nlb-1/forwardingrules"},

		// Application Load Balancers
		{"list_application_loadbalancers", map[string]any{"datacenter_id": dc}, "/cloudapi/v6/datacenters/" + dc + "/applicationloadbalancers"},
		{"get_application_loadbalancer", map[string]any{"datacenter_id": dc, "application_loadbalancer_id": "alb-1"}, "/cloudapi/v6/datacenters/" + dc + "/applicationloadbalancers/alb-1"},
		{"list_alb_forwarding_rules", map[string]any{"datacenter_id": dc, "application_loadbalancer_id": "alb-1"}, "/cloudapi/v6/datacenters/" + dc + "/applicationloadbalancers/alb-1/forwardingrules"},

		// Target Groups
		{"list_target_groups", map[string]any{}, "/cloudapi/v6/targetgroups"},
		{"get_target_group", map[string]any{"target_group_id": "tg-1"}, "/cloudapi/v6/targetgroups/tg-1"},

		// NAT Gateways
		{"list_nat_gateways", map[string]any{"datacenter_id": dc}, "/cloudapi/v6/datacenters/" + dc + "/natgateways"},
		{"get_nat_gateway", map[string]any{"datacenter_id": dc, "nat_gateway_id": "nat-1"}, "/cloudapi/v6/datacenters/" + dc + "/natgateways/nat-1"},
		{"list_nat_gateway_rules", map[string]any{"datacenter_id": dc, "nat_gateway_id": "nat-1"}, "/cloudapi/v6/datacenters/" + dc + "/natgateways/nat-1/rules"},

		// Private Cross-Connects
		{"list_private_cross_connects", map[string]any{}, "/cloudapi/v6/pccs"},
		{"get_private_cross_connect", map[string]any{"pcc_id": "pcc-1"}, "/cloudapi/v6/pccs/pcc-1"},

		// Security Groups
		{"list_security_groups", map[string]any{"datacenter_id": dc}, "/cloudapi/v6/datacenters/" + dc + "/securitygroups"},
		{"get_security_group", map[string]any{"datacenter_id": dc, "security_group_id": "sg-1"}, "/cloudapi/v6/datacenters/" + dc + "/securitygroups/sg-1"},
		{"list_security_group_rules", map[string]any{"datacenter_id": dc, "security_group_id": "sg-1"}, "/cloudapi/v6/datacenters/" + dc + "/securitygroups/sg-1/rules"},
		{"get_security_group_rule", map[string]any{"datacenter_id": dc, "security_group_id": "sg-1", "rule_id": "rule-1"}, "/cloudapi/v6/datacenters/" + dc + "/securitygroups/sg-1/rules/rule-1"},

		// Contract
		{"get_contract", map[string]any{}, "/cloudapi/v6/contracts"},

		// Requests
		{"list_requests", map[string]any{}, "/cloudapi/v6/requests"},
		{"get_request", map[string]any{"request_id": "req-1"}, "/cloudapi/v6/requests/req-1"},
		{"get_request_status", map[string]any{"request_id": "req-1"}, "/cloudapi/v6/requests/req-1/status"},

		// Templates
		{"list_templates", map[string]any{}, "/cloudapi/v6/templates"},
		{"get_template", map[string]any{"template_id": "tpl-1"}, "/cloudapi/v6/templates/tpl-1"},

		// Images
		{"list_images", map[string]any{}, "/cloudapi/v6/images"},

		// Locations
		{"list_locations", map[string]any{}, "/cloudapi/v6/locations"},

		// Snapshots
		{"list_snapshots", map[string]any{}, "/cloudapi/v6/snapshots"},
		{"get_snapshot", map[string]any{"snapshot_id": "snap-1"}, "/cloudapi/v6/snapshots/snap-1"},
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

			req, ok := h.log.lastRequest()
			if !ok {
				t.Fatalf("CallTool(%q) made no HTTP request", tt.name)
			}
			if req.Path != tt.wantPath {
				t.Errorf("CallTool(%q) path = %q, want %q", tt.name, req.Path, tt.wantPath)
			}
		})
	}
}
