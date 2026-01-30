package networking

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ionos-cloud/ionoscloud-mcp/pkg/api"
	"github.com/ionos-cloud/ionoscloud-mcp/pkg/ionos"
	ionoscloud "github.com/ionos-cloud/sdk-go/v6"
)

func initFirewall() []api.ServerTool {
	return []api.ServerTool{
		{
			Tool: api.Tool{
				Name:        "list_firewall_rules",
				Description: "List all firewall rules on a NIC",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {
							"type": "string",
							"description": "The ID of the data center"
						},
						"server_id": {
							"type": "string",
							"description": "The ID of the server"
						},
						"nic_id": {
							"type": "string",
							"description": "The ID of the NIC"
						}
					},
					"required": ["datacenter_id", "server_id", "nic_id"]
				}`),
				Annotations: api.ReadOnly("List Firewall Rules"),
			},
			Handler: listFirewallRules,
		},
		{
			Tool: api.Tool{
				Name:        "get_firewall_rule",
				Description: "Get details of a specific firewall rule",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {
							"type": "string",
							"description": "The ID of the data center"
						},
						"server_id": {
							"type": "string",
							"description": "The ID of the server"
						},
						"nic_id": {
							"type": "string",
							"description": "The ID of the NIC"
						},
						"firewallrule_id": {
							"type": "string",
							"description": "The ID of the firewall rule"
						}
					},
					"required": ["datacenter_id", "server_id", "nic_id", "firewallrule_id"]
				}`),
				Annotations: api.ReadOnly("Get Firewall Rule"),
			},
			Handler: getFirewallRule,
		},
		{
			Tool: api.Tool{
				Name:        "create_firewall_rule",
				Description: "Create a new firewall rule on a NIC",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {
							"type": "string",
							"description": "The ID of the data center"
						},
						"server_id": {
							"type": "string",
							"description": "The ID of the server"
						},
						"nic_id": {
							"type": "string",
							"description": "The ID of the NIC"
						},
						"name": {
							"type": "string",
							"description": "The name of the firewall rule"
						},
						"protocol": {
							"type": "string",
							"description": "The protocol (TCP, UDP, ICMP, ICMPv6, GRE, ESP, AH, ANY)"
						},
						"source_mac": {
							"type": "string",
							"description": "Only traffic from this MAC address is allowed"
						},
						"source_ip": {
							"type": "string",
							"description": "Only traffic from this IP address/CIDR is allowed"
						},
						"target_ip": {
							"type": "string",
							"description": "Only traffic to this IP address/CIDR is allowed"
						},
						"port_range_start": {
							"type": "integer",
							"description": "Start port of the allowed port range (1-65534)"
						},
						"port_range_end": {
							"type": "integer",
							"description": "End port of the allowed port range (1-65534)"
						},
						"icmp_type": {
							"type": "integer",
							"description": "ICMP type (for ICMP protocol)"
						},
						"icmp_code": {
							"type": "integer",
							"description": "ICMP code (for ICMP protocol)"
						},
						"type": {
							"type": "string",
							"description": "The type of firewall rule (INGRESS or EGRESS)"
						}
					},
					"required": ["datacenter_id", "server_id", "nic_id", "protocol"]
				}`),
				Annotations: api.NonIdempotent("Create Firewall Rule"),
			},
			Handler: createFirewallRule,
		},
		{
			Tool: api.Tool{
				Name:        "update_firewall_rule",
				Description: "Update an existing firewall rule",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {
							"type": "string",
							"description": "The ID of the data center"
						},
						"server_id": {
							"type": "string",
							"description": "The ID of the server"
						},
						"nic_id": {
							"type": "string",
							"description": "The ID of the NIC"
						},
						"firewallrule_id": {
							"type": "string",
							"description": "The ID of the firewall rule to update"
						},
						"name": {
							"type": "string",
							"description": "The new name for the firewall rule"
						},
						"protocol": {
							"type": "string",
							"description": "The new protocol"
						},
						"source_mac": {
							"type": "string",
							"description": "The new source MAC address"
						},
						"source_ip": {
							"type": "string",
							"description": "The new source IP address/CIDR"
						},
						"target_ip": {
							"type": "string",
							"description": "The new target IP address/CIDR"
						},
						"port_range_start": {
							"type": "integer",
							"description": "The new start port"
						},
						"port_range_end": {
							"type": "integer",
							"description": "The new end port"
						},
						"icmp_type": {
							"type": "integer",
							"description": "The new ICMP type"
						},
						"icmp_code": {
							"type": "integer",
							"description": "The new ICMP code"
						},
						"type": {
							"type": "string",
							"description": "The new firewall rule type (INGRESS or EGRESS)"
						}
					},
					"required": ["datacenter_id", "server_id", "nic_id", "firewallrule_id"]
				}`),
				Annotations: api.Idempotent("Update Firewall Rule"),
			},
			Handler: updateFirewallRule,
		},
		{
			Tool: api.Tool{
				Name:        "delete_firewall_rule",
				Description: "Delete a firewall rule",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {
							"type": "string",
							"description": "The ID of the data center"
						},
						"server_id": {
							"type": "string",
							"description": "The ID of the server"
						},
						"nic_id": {
							"type": "string",
							"description": "The ID of the NIC"
						},
						"firewallrule_id": {
							"type": "string",
							"description": "The ID of the firewall rule to delete"
						}
					},
					"required": ["datacenter_id", "server_id", "nic_id", "firewallrule_id"]
				}`),
				Annotations: api.Destructive("Delete Firewall Rule"),
			},
			Handler: deleteFirewallRule,
		},
	}
}

func listFirewallRules(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	serverID, ok := api.GetRequiredString(params.Arguments, "server_id")
	if !ok {
		return nil, fmt.Errorf("server_id is required")
	}
	nicID, ok := api.GetRequiredString(params.Arguments, "nic_id")
	if !ok {
		return nil, fmt.Errorf("nic_id is required")
	}

	rules, _, err := params.Client.FirewallRulesApi.DatacentersServersNicsFirewallrulesGet(ctx, datacenterID, serverID, nicID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to list firewall rules: %w", err)
	}
	return api.MarshalResult(rules, "firewall rules")
}

func getFirewallRule(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	serverID, ok := api.GetRequiredString(params.Arguments, "server_id")
	if !ok {
		return nil, fmt.Errorf("server_id is required")
	}
	nicID, ok := api.GetRequiredString(params.Arguments, "nic_id")
	if !ok {
		return nil, fmt.Errorf("nic_id is required")
	}
	firewallRuleID, ok := api.GetRequiredString(params.Arguments, "firewallrule_id")
	if !ok {
		return nil, fmt.Errorf("firewallrule_id is required")
	}

	rule, _, err := params.Client.FirewallRulesApi.DatacentersServersNicsFirewallrulesFindById(ctx, datacenterID, serverID, nicID, firewallRuleID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get firewall rule: %w", err)
	}
	return api.MarshalResult(rule, "firewall rule")
}

func createFirewallRule(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	serverID, ok := api.GetRequiredString(params.Arguments, "server_id")
	if !ok {
		return nil, fmt.Errorf("server_id is required")
	}
	nicID, ok := api.GetRequiredString(params.Arguments, "nic_id")
	if !ok {
		return nil, fmt.Errorf("nic_id is required")
	}
	protocol, ok := api.GetRequiredString(params.Arguments, "protocol")
	if !ok {
		return nil, fmt.Errorf("protocol is required")
	}

	// Validate protocol
	if err := ionos.ValidateProtocol(protocol); err != nil {
		return nil, err
	}

	name := api.GetOptionalString(params.Arguments, "name")
	sourceMac := api.GetOptionalString(params.Arguments, "source_mac")
	sourceIP := api.GetOptionalString(params.Arguments, "source_ip")
	targetIP := api.GetOptionalString(params.Arguments, "target_ip")
	ruleType := api.GetOptionalString(params.Arguments, "type")
	portRangeStart, portRangeStartSet := api.GetOptionalInt32(params.Arguments, "port_range_start")
	portRangeEnd, portRangeEndSet := api.GetOptionalInt32(params.Arguments, "port_range_end")
	icmpType, icmpTypeSet := api.GetOptionalInt32(params.Arguments, "icmp_type")
	icmpCode, icmpCodeSet := api.GetOptionalInt32(params.Arguments, "icmp_code")

	// Validate MAC address
	if err := ionos.ValidateMAC(sourceMac); err != nil {
		return nil, err
	}
	// Validate IP addresses
	if err := ionos.ValidateIP(sourceIP); err != nil {
		return nil, fmt.Errorf("invalid source_ip: %w", err)
	}
	if err := ionos.ValidateIP(targetIP); err != nil {
		return nil, fmt.Errorf("invalid target_ip: %w", err)
	}
	// Validate port range
	if err := ionos.ValidatePortRange(portRangeStart, portRangeEnd, "port_range"); err != nil {
		return nil, err
	}
	// Validate ICMP parameters
	if err := ionos.ValidateICMP(icmpType, icmpCode, icmpTypeSet, icmpCodeSet); err != nil {
		return nil, err
	}

	properties := ionoscloud.FirewallruleProperties{
		Protocol: &protocol,
	}
	if name != "" {
		properties.Name = &name
	}
	if sourceMac != "" {
		properties.SourceMac = &sourceMac
	}
	if sourceIP != "" {
		properties.SourceIp = &sourceIP
	}
	if targetIP != "" {
		properties.TargetIp = &targetIP
	}
	if portRangeStartSet && portRangeStart != 0 {
		properties.PortRangeStart = &portRangeStart
	}
	if portRangeEndSet && portRangeEnd != 0 {
		properties.PortRangeEnd = &portRangeEnd
	}
	if icmpTypeSet {
		properties.IcmpType = &icmpType
	}
	if icmpCodeSet {
		properties.IcmpCode = &icmpCode
	}
	if ruleType != "" {
		properties.Type = &ruleType
	}

	rule := ionoscloud.FirewallRule{
		Properties: &properties,
	}

	result, _, err := params.Client.FirewallRulesApi.DatacentersServersNicsFirewallrulesPost(ctx, datacenterID, serverID, nicID).Firewallrule(rule).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to create firewall rule: %w", err)
	}
	return api.MarshalResult(result, "firewall rule")
}

func updateFirewallRule(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	serverID, ok := api.GetRequiredString(params.Arguments, "server_id")
	if !ok {
		return nil, fmt.Errorf("server_id is required")
	}
	nicID, ok := api.GetRequiredString(params.Arguments, "nic_id")
	if !ok {
		return nil, fmt.Errorf("nic_id is required")
	}
	firewallRuleID, ok := api.GetRequiredString(params.Arguments, "firewallrule_id")
	if !ok {
		return nil, fmt.Errorf("firewallrule_id is required")
	}

	name := api.GetOptionalString(params.Arguments, "name")
	protocol := api.GetOptionalString(params.Arguments, "protocol")
	sourceMac := api.GetOptionalString(params.Arguments, "source_mac")
	sourceIP := api.GetOptionalString(params.Arguments, "source_ip")
	targetIP := api.GetOptionalString(params.Arguments, "target_ip")
	ruleType := api.GetOptionalString(params.Arguments, "type")
	portRangeStart, portRangeStartSet := api.GetOptionalInt32(params.Arguments, "port_range_start")
	portRangeEnd, portRangeEndSet := api.GetOptionalInt32(params.Arguments, "port_range_end")
	icmpType, icmpTypeSet := api.GetOptionalInt32(params.Arguments, "icmp_type")
	icmpCode, icmpCodeSet := api.GetOptionalInt32(params.Arguments, "icmp_code")

	// Check if at least one field is provided
	if name == "" && protocol == "" && sourceMac == "" && sourceIP == "" && targetIP == "" && !portRangeStartSet && !portRangeEndSet && !icmpTypeSet && !icmpCodeSet && ruleType == "" {
		return nil, fmt.Errorf("at least one field must be provided for update")
	}

	// Validate protocol if provided
	if protocol != "" {
		if err := ionos.ValidateProtocol(protocol); err != nil {
			return nil, err
		}
	}
	// Validate MAC address if provided
	if err := ionos.ValidateMAC(sourceMac); err != nil {
		return nil, err
	}
	// Validate IP addresses if provided
	if err := ionos.ValidateIP(sourceIP); err != nil {
		return nil, fmt.Errorf("invalid source_ip: %w", err)
	}
	if err := ionos.ValidateIP(targetIP); err != nil {
		return nil, fmt.Errorf("invalid target_ip: %w", err)
	}
	// Validate port range
	if err := ionos.ValidatePortRange(portRangeStart, portRangeEnd, "port_range"); err != nil {
		return nil, err
	}
	// Validate ICMP parameters
	if err := ionos.ValidateICMP(icmpType, icmpCode, icmpTypeSet, icmpCodeSet); err != nil {
		return nil, err
	}

	properties := ionoscloud.FirewallruleProperties{}
	if name != "" {
		properties.Name = &name
	}
	if protocol != "" {
		properties.Protocol = &protocol
	}
	if sourceMac != "" {
		properties.SourceMac = &sourceMac
	}
	if sourceIP != "" {
		properties.SourceIp = &sourceIP
	}
	if targetIP != "" {
		properties.TargetIp = &targetIP
	}
	if portRangeStartSet {
		properties.PortRangeStart = &portRangeStart
	}
	if portRangeEndSet {
		properties.PortRangeEnd = &portRangeEnd
	}
	if icmpTypeSet {
		properties.IcmpType = &icmpType
	}
	if icmpCodeSet {
		properties.IcmpCode = &icmpCode
	}
	if ruleType != "" {
		properties.Type = &ruleType
	}

	result, _, err := params.Client.FirewallRulesApi.DatacentersServersNicsFirewallrulesPatch(ctx, datacenterID, serverID, nicID, firewallRuleID).Firewallrule(properties).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to update firewall rule: %w", err)
	}
	return api.MarshalResult(result, "firewall rule")
}

func deleteFirewallRule(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	serverID, ok := api.GetRequiredString(params.Arguments, "server_id")
	if !ok {
		return nil, fmt.Errorf("server_id is required")
	}
	nicID, ok := api.GetRequiredString(params.Arguments, "nic_id")
	if !ok {
		return nil, fmt.Errorf("nic_id is required")
	}
	firewallRuleID, ok := api.GetRequiredString(params.Arguments, "firewallrule_id")
	if !ok {
		return nil, fmt.Errorf("firewallrule_id is required")
	}

	_, err := params.Client.FirewallRulesApi.DatacentersServersNicsFirewallrulesDelete(ctx, datacenterID, serverID, nicID, firewallRuleID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to delete firewall rule: %w", err)
	}
	return api.StatusResult(map[string]string{"status": "deleted", "firewallrule_id": firewallRuleID})
}
