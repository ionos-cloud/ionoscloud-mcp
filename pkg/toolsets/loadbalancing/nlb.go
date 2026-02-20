package loadbalancing

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ionos-cloud/ionoscloud-mcp/pkg/api"
	"github.com/ionos-cloud/ionoscloud-mcp/pkg/ionos"
	ionoscloud "github.com/ionos-cloud/sdk-go/v6"
)

// Valid NLB forwarding rule algorithms
var validNlbAlgorithms = map[string]bool{
	"ROUND_ROBIN": true, "LEAST_CONNECTION": true, "RANDOM": true,
	"SOURCE_IP": true,
}

// Valid NLB forwarding rule protocols
var validNlbProtocols = map[string]bool{
	"TCP": true, "HTTP": true,
}

func initNlb() []api.ServerTool {
	return []api.ServerTool{
		{
			Tool: api.Tool{
				Name:        "list_network_load_balancers",
				Description: "List all Network Load Balancers in a data center",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {"type": "string", "description": "The ID of the data center"}
					},
					"required": ["datacenter_id"]
				}`),
				Annotations: api.ReadOnly("List NLBs"),
			},
			Handler: listNetworkLoadBalancers,
		},
		{
			Tool: api.Tool{
				Name:        "get_network_load_balancer",
				Description: "Get details of a specific Network Load Balancer",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {"type": "string", "description": "The ID of the data center"},
						"nlb_id": {"type": "string", "description": "The ID of the Network Load Balancer"}
					},
					"required": ["datacenter_id", "nlb_id"]
				}`),
				Annotations: api.ReadOnly("Get NLB"),
			},
			Handler: getNetworkLoadBalancer,
		},
		{
			Tool: api.Tool{
				Name:        "create_network_load_balancer",
				Description: "Create a Network Load Balancer in a data center",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {
							"type": "string",
							"description": "The ID of the data center"
						},
						"name": {
							"type": "string",
							"description": "The name of the Network Load Balancer"
						},
						"listener_lan": {
							"type": "integer",
							"description": "The ID of the listening (inbound) LAN"
						},
						"target_lan": {
							"type": "integer",
							"description": "The ID of the balanced private target LAN (outbound)"
						},
						"ips": {
							"type": "array",
							"items": {"type": "string"},
							"description": "Collection of the NLB IP addresses (inbound and outbound IPs of the listenerLan)"
						},
						"lb_private_ips": {
							"type": "array",
							"items": {"type": "string"},
							"description": "Collection of private IP addresses with subnet mask of the NLB"
						}
					},
					"required": ["datacenter_id", "name", "listener_lan", "target_lan"]
				}`),
				Annotations: api.NonIdempotent("Create NLB"),
			},
			Handler: createNetworkLoadBalancer,
		},
		{
			Tool: api.Tool{
				Name:        "update_network_load_balancer",
				Description: "Update a Network Load Balancer",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {
							"type": "string",
							"description": "The ID of the data center"
						},
						"nlb_id": {
							"type": "string",
							"description": "The ID of the Network Load Balancer"
						},
						"name": {
							"type": "string",
							"description": "The new name for the NLB"
						},
						"listener_lan": {
							"type": "integer",
							"description": "The new listening (inbound) LAN ID"
						},
						"target_lan": {
							"type": "integer",
							"description": "The new target LAN ID (outbound)"
						},
						"ips": {
							"type": "array",
							"items": {"type": "string"},
							"description": "Updated collection of NLB IP addresses"
						},
						"lb_private_ips": {
							"type": "array",
							"items": {"type": "string"},
							"description": "Updated collection of private IP addresses with subnet mask"
						}
					},
					"required": ["datacenter_id", "nlb_id"]
				}`),
				Annotations: api.Idempotent("Update NLB"),
			},
			Handler: updateNetworkLoadBalancer,
		},
		{
			Tool: api.Tool{
				Name:        "delete_network_load_balancer",
				Description: "Delete a Network Load Balancer",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {
							"type": "string",
							"description": "The ID of the data center"
						},
						"nlb_id": {
							"type": "string",
							"description": "The ID of the Network Load Balancer to delete"
						}
					},
					"required": ["datacenter_id", "nlb_id"]
				}`),
				Annotations: api.Destructive("Delete NLB"),
			},
			Handler: deleteNetworkLoadBalancer,
		},
		// NLB Forwarding Rules
		{
			Tool: api.Tool{
				Name:        "list_nlb_forwarding_rules",
				Description: "List all forwarding rules of a Network Load Balancer",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {"type": "string", "description": "The ID of the data center"},
						"nlb_id": {"type": "string", "description": "The ID of the Network Load Balancer"}
					},
					"required": ["datacenter_id", "nlb_id"]
				}`),
				Annotations: api.ReadOnly("List NLB Rules"),
			},
			Handler: listNlbForwardingRules,
		},
		{
			Tool: api.Tool{
				Name:        "get_nlb_forwarding_rule",
				Description: "Get details of a specific NLB forwarding rule",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {"type": "string", "description": "The ID of the data center"},
						"nlb_id": {"type": "string", "description": "The ID of the Network Load Balancer"},
						"rule_id": {"type": "string", "description": "The ID of the forwarding rule"}
					},
					"required": ["datacenter_id", "nlb_id", "rule_id"]
				}`),
				Annotations: api.ReadOnly("Get NLB Rule"),
			},
			Handler: getNlbForwardingRule,
		},
		{
			Tool: api.Tool{
				Name:        "create_nlb_forwarding_rule",
				Description: "Create a forwarding rule for a Network Load Balancer",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {
							"type": "string",
							"description": "The ID of the data center"
						},
						"nlb_id": {
							"type": "string",
							"description": "The ID of the Network Load Balancer"
						},
						"name": {
							"type": "string",
							"description": "The name of the forwarding rule"
						},
						"algorithm": {
							"type": "string",
							"description": "Balancing algorithm",
							"enum": ["ROUND_ROBIN", "LEAST_CONNECTION", "RANDOM", "SOURCE_IP"]
						},
						"protocol": {
							"type": "string",
							"description": "Balancing protocol",
							"enum": ["TCP", "HTTP"]
						},
						"listener_ip": {
							"type": "string",
							"description": "The listening (inbound) IP address"
						},
						"listener_port": {
							"type": "integer",
							"description": "The listening (inbound) port number (1-65535)"
						},
						"targets": {
							"type": "array",
							"items": {
								"type": "object",
								"properties": {
									"ip": {"type": "string", "description": "Target IP address"},
									"port": {"type": "integer", "description": "Target port (1-65535)"},
									"weight": {"type": "integer", "description": "Traffic weight (0-256, default: 1)"}
								},
								"required": ["ip", "port", "weight"]
							},
							"description": "Array of balanced targets"
						}
					},
					"required": ["datacenter_id", "nlb_id", "name", "algorithm", "protocol", "listener_ip", "listener_port", "targets"]
				}`),
				Annotations: api.NonIdempotent("Create NLB Rule"),
			},
			Handler: createNlbForwardingRule,
		},
		{
			Tool: api.Tool{
				Name:        "update_nlb_forwarding_rule",
				Description: "Update an NLB forwarding rule",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {
							"type": "string",
							"description": "The ID of the data center"
						},
						"nlb_id": {
							"type": "string",
							"description": "The ID of the Network Load Balancer"
						},
						"rule_id": {
							"type": "string",
							"description": "The ID of the forwarding rule to update"
						},
						"name": {
							"type": "string",
							"description": "The new name for the forwarding rule"
						},
						"algorithm": {
							"type": "string",
							"description": "Balancing algorithm",
							"enum": ["ROUND_ROBIN", "LEAST_CONNECTION", "RANDOM", "SOURCE_IP"]
						},
						"protocol": {
							"type": "string",
							"description": "Balancing protocol",
							"enum": ["TCP", "HTTP"]
						},
						"listener_ip": {
							"type": "string",
							"description": "The new listening IP address"
						},
						"listener_port": {
							"type": "integer",
							"description": "The new listening port number (1-65535)"
						},
						"targets": {
							"type": "array",
							"items": {
								"type": "object",
								"properties": {
									"ip": {"type": "string", "description": "Target IP address"},
									"port": {"type": "integer", "description": "Target port (1-65535)"},
									"weight": {"type": "integer", "description": "Traffic weight (0-256, default: 1)"}
								},
								"required": ["ip", "port", "weight"]
							},
							"description": "Updated array of balanced targets"
						}
					},
					"required": ["datacenter_id", "nlb_id", "rule_id"]
				}`),
				Annotations: api.Idempotent("Update NLB Rule"),
			},
			Handler: updateNlbForwardingRule,
		},
		{
			Tool: api.Tool{
				Name:        "delete_nlb_forwarding_rule",
				Description: "Delete an NLB forwarding rule",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {
							"type": "string",
							"description": "The ID of the data center"
						},
						"nlb_id": {
							"type": "string",
							"description": "The ID of the Network Load Balancer"
						},
						"rule_id": {
							"type": "string",
							"description": "The ID of the forwarding rule to delete"
						}
					},
					"required": ["datacenter_id", "nlb_id", "rule_id"]
				}`),
				Annotations: api.Destructive("Delete NLB Rule"),
			},
			Handler: deleteNlbForwardingRule,
		},
	}
}

func listNetworkLoadBalancers(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}

	nlbs, _, err := params.Client.NetworkLoadBalancersApi.DatacentersNetworkloadbalancersGet(ctx, datacenterID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to list Network Load Balancers: %w", err)
	}
	return api.MarshalResult(nlbs, "Network Load Balancers")
}

func getNetworkLoadBalancer(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	nlbID, ok := api.GetRequiredString(params.Arguments, "nlb_id")
	if !ok {
		return nil, fmt.Errorf("nlb_id is required")
	}

	nlb, _, err := params.Client.NetworkLoadBalancersApi.DatacentersNetworkloadbalancersFindByNetworkLoadBalancerId(ctx, datacenterID, nlbID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get Network Load Balancer: %w", err)
	}
	return api.MarshalResult(nlb, "Network Load Balancer")
}

func createNetworkLoadBalancer(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	name, ok := api.GetRequiredString(params.Arguments, "name")
	if !ok {
		return nil, fmt.Errorf("name is required")
	}
	listenerLanFloat, ok := api.GetOptionalFloat(params.Arguments, "listener_lan")
	if !ok {
		return nil, fmt.Errorf("listener_lan is required")
	}
	targetLanFloat, ok := api.GetOptionalFloat(params.Arguments, "target_lan")
	if !ok {
		return nil, fmt.Errorf("target_lan is required")
	}

	listenerLan := int32(listenerLanFloat)
	targetLan := int32(targetLanFloat)

	if listenerLan < 1 {
		return nil, fmt.Errorf("listener_lan must be at least 1, got %d", listenerLan)
	}
	if targetLan < 1 {
		return nil, fmt.Errorf("target_lan must be at least 1, got %d", targetLan)
	}

	properties := ionoscloud.NetworkLoadBalancerProperties{
		Name:        &name,
		ListenerLan: &listenerLan,
		TargetLan:   &targetLan,
	}

	if ips := api.GetStringSlice(params.Arguments, "ips"); len(ips) > 0 {
		properties.Ips = &ips
	}
	if lbPrivateIPs := api.GetStringSlice(params.Arguments, "lb_private_ips"); len(lbPrivateIPs) > 0 {
		properties.LbPrivateIps = &lbPrivateIPs
	}

	nlb := ionoscloud.NetworkLoadBalancer{
		Properties: &properties,
	}

	result, _, err := params.Client.NetworkLoadBalancersApi.DatacentersNetworkloadbalancersPost(ctx, datacenterID).NetworkLoadBalancer(nlb).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to create Network Load Balancer: %w", err)
	}
	return api.MarshalResult(result, "Network Load Balancer")
}

func updateNetworkLoadBalancer(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	nlbID, ok := api.GetRequiredString(params.Arguments, "nlb_id")
	if !ok {
		return nil, fmt.Errorf("nlb_id is required")
	}

	name := api.GetOptionalString(params.Arguments, "name")
	listenerLan, listenerLanSet := api.GetOptionalInt32(params.Arguments, "listener_lan")
	targetLan, targetLanSet := api.GetOptionalInt32(params.Arguments, "target_lan")
	ips := api.GetStringSlice(params.Arguments, "ips")
	lbPrivateIPs := api.GetStringSlice(params.Arguments, "lb_private_ips")

	if name == "" && !listenerLanSet && !targetLanSet && len(ips) == 0 && len(lbPrivateIPs) == 0 {
		return nil, fmt.Errorf("at least one field must be provided for update")
	}

	if listenerLanSet && listenerLan < 1 {
		return nil, fmt.Errorf("listener_lan must be at least 1, got %d", listenerLan)
	}
	if targetLanSet && targetLan < 1 {
		return nil, fmt.Errorf("target_lan must be at least 1, got %d", targetLan)
	}

	properties := ionoscloud.NetworkLoadBalancerProperties{}
	if name != "" {
		properties.Name = &name
	}
	if listenerLanSet {
		properties.ListenerLan = &listenerLan
	}
	if targetLanSet {
		properties.TargetLan = &targetLan
	}
	if len(ips) > 0 {
		properties.Ips = &ips
	}
	if len(lbPrivateIPs) > 0 {
		properties.LbPrivateIps = &lbPrivateIPs
	}

	result, _, err := params.Client.NetworkLoadBalancersApi.DatacentersNetworkloadbalancersPatch(ctx, datacenterID, nlbID).NetworkLoadBalancerProperties(properties).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to update Network Load Balancer: %w", err)
	}
	return api.MarshalResult(result, "Network Load Balancer")
}

func deleteNetworkLoadBalancer(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	nlbID, ok := api.GetRequiredString(params.Arguments, "nlb_id")
	if !ok {
		return nil, fmt.Errorf("nlb_id is required")
	}

	_, err := params.Client.NetworkLoadBalancersApi.DatacentersNetworkloadbalancersDelete(ctx, datacenterID, nlbID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to delete Network Load Balancer: %w", err)
	}
	return api.StatusResult(map[string]string{"status": "deleted", "nlb_id": nlbID})
}

// NLB Forwarding Rules handlers

func listNlbForwardingRules(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	nlbID, ok := api.GetRequiredString(params.Arguments, "nlb_id")
	if !ok {
		return nil, fmt.Errorf("nlb_id is required")
	}

	rules, _, err := params.Client.NetworkLoadBalancersApi.DatacentersNetworkloadbalancersForwardingrulesGet(ctx, datacenterID, nlbID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to list NLB forwarding rules: %w", err)
	}
	return api.MarshalResult(rules, "NLB forwarding rules")
}

func getNlbForwardingRule(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	nlbID, ok := api.GetRequiredString(params.Arguments, "nlb_id")
	if !ok {
		return nil, fmt.Errorf("nlb_id is required")
	}
	ruleID, ok := api.GetRequiredString(params.Arguments, "rule_id")
	if !ok {
		return nil, fmt.Errorf("rule_id is required")
	}

	rule, _, err := params.Client.NetworkLoadBalancersApi.DatacentersNetworkloadbalancersForwardingrulesFindByForwardingRuleId(ctx, datacenterID, nlbID, ruleID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get NLB forwarding rule: %w", err)
	}
	return api.MarshalResult(rule, "NLB forwarding rule")
}

func createNlbForwardingRule(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	nlbID, ok := api.GetRequiredString(params.Arguments, "nlb_id")
	if !ok {
		return nil, fmt.Errorf("nlb_id is required")
	}
	name, ok := api.GetRequiredString(params.Arguments, "name")
	if !ok {
		return nil, fmt.Errorf("name is required")
	}
	algorithm, ok := api.GetRequiredString(params.Arguments, "algorithm")
	if !ok {
		return nil, fmt.Errorf("algorithm is required")
	}
	protocol, ok := api.GetRequiredString(params.Arguments, "protocol")
	if !ok {
		return nil, fmt.Errorf("protocol is required")
	}
	listenerIP, ok := api.GetRequiredString(params.Arguments, "listener_ip")
	if !ok {
		return nil, fmt.Errorf("listener_ip is required")
	}
	listenerPortFloat, ok := api.GetOptionalFloat(params.Arguments, "listener_port")
	if !ok {
		return nil, fmt.Errorf("listener_port is required")
	}

	listenerPort := int32(listenerPortFloat)

	// Validate inputs
	if !validNlbAlgorithms[algorithm] {
		return nil, fmt.Errorf("invalid algorithm: %s (valid: ROUND_ROBIN, LEAST_CONNECTION, RANDOM, SOURCE_IP)", algorithm)
	}
	if !validNlbProtocols[protocol] {
		return nil, fmt.Errorf("invalid protocol: %s (valid: TCP, HTTP)", protocol)
	}
	if err := ionos.ValidateIP(listenerIP); err != nil {
		return nil, fmt.Errorf("invalid listener_ip: %w", err)
	}
	if listenerPort < 1 || listenerPort > 65535 {
		return nil, fmt.Errorf("listener_port must be between 1-65535, got %d", listenerPort)
	}

	// Parse targets
	targets, err := parseNlbTargets(params.Arguments)
	if err != nil {
		return nil, err
	}
	if len(targets) == 0 {
		return nil, fmt.Errorf("targets is required and must not be empty")
	}

	properties := ionoscloud.NetworkLoadBalancerForwardingRuleProperties{
		Name:         &name,
		Algorithm:    &algorithm,
		Protocol:     &protocol,
		ListenerIp:   &listenerIP,
		ListenerPort: &listenerPort,
		Targets:      &targets,
	}

	rule := ionoscloud.NetworkLoadBalancerForwardingRule{
		Properties: &properties,
	}

	result, _, err := params.Client.NetworkLoadBalancersApi.DatacentersNetworkloadbalancersForwardingrulesPost(ctx, datacenterID, nlbID).NetworkLoadBalancerForwardingRule(rule).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to create NLB forwarding rule: %w", err)
	}
	return api.MarshalResult(result, "NLB forwarding rule")
}

func updateNlbForwardingRule(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	nlbID, ok := api.GetRequiredString(params.Arguments, "nlb_id")
	if !ok {
		return nil, fmt.Errorf("nlb_id is required")
	}
	ruleID, ok := api.GetRequiredString(params.Arguments, "rule_id")
	if !ok {
		return nil, fmt.Errorf("rule_id is required")
	}

	name := api.GetOptionalString(params.Arguments, "name")
	algorithm := api.GetOptionalString(params.Arguments, "algorithm")
	protocol := api.GetOptionalString(params.Arguments, "protocol")
	listenerIP := api.GetOptionalString(params.Arguments, "listener_ip")
	listenerPort, listenerPortSet := api.GetOptionalInt32(params.Arguments, "listener_port")
	targetsRaw, targetsSet := params.Arguments["targets"].([]interface{})

	if name == "" && algorithm == "" && protocol == "" && listenerIP == "" && !listenerPortSet && !targetsSet {
		return nil, fmt.Errorf("at least one field must be provided for update")
	}

	// Validate inputs
	if algorithm != "" && !validNlbAlgorithms[algorithm] {
		return nil, fmt.Errorf("invalid algorithm: %s (valid: ROUND_ROBIN, LEAST_CONNECTION, RANDOM, SOURCE_IP)", algorithm)
	}
	if protocol != "" && !validNlbProtocols[protocol] {
		return nil, fmt.Errorf("invalid protocol: %s (valid: TCP, HTTP)", protocol)
	}
	if listenerIP != "" {
		if err := ionos.ValidateIP(listenerIP); err != nil {
			return nil, fmt.Errorf("invalid listener_ip: %w", err)
		}
	}
	if listenerPortSet && (listenerPort < 1 || listenerPort > 65535) {
		return nil, fmt.Errorf("listener_port must be between 1-65535, got %d", listenerPort)
	}

	properties := ionoscloud.NetworkLoadBalancerForwardingRuleProperties{}
	if name != "" {
		properties.Name = &name
	}
	if algorithm != "" {
		properties.Algorithm = &algorithm
	}
	if protocol != "" {
		properties.Protocol = &protocol
	}
	if listenerIP != "" {
		properties.ListenerIp = &listenerIP
	}
	if listenerPortSet {
		properties.ListenerPort = &listenerPort
	}
	if targetsSet && len(targetsRaw) > 0 {
		targets, err := parseNlbTargets(params.Arguments)
		if err != nil {
			return nil, err
		}
		properties.Targets = &targets
	}

	result, _, err := params.Client.NetworkLoadBalancersApi.DatacentersNetworkloadbalancersForwardingrulesPatch(ctx, datacenterID, nlbID, ruleID).NetworkLoadBalancerForwardingRuleProperties(properties).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to update NLB forwarding rule: %w", err)
	}
	return api.MarshalResult(result, "NLB forwarding rule")
}

func deleteNlbForwardingRule(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	nlbID, ok := api.GetRequiredString(params.Arguments, "nlb_id")
	if !ok {
		return nil, fmt.Errorf("nlb_id is required")
	}
	ruleID, ok := api.GetRequiredString(params.Arguments, "rule_id")
	if !ok {
		return nil, fmt.Errorf("rule_id is required")
	}

	_, err := params.Client.NetworkLoadBalancersApi.DatacentersNetworkloadbalancersForwardingrulesDelete(ctx, datacenterID, nlbID, ruleID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to delete NLB forwarding rule: %w", err)
	}
	return api.StatusResult(map[string]string{"status": "deleted", "rule_id": ruleID})
}

// parseNlbTargets parses and validates NLB forwarding rule targets from arguments.
func parseNlbTargets(args map[string]interface{}) ([]ionoscloud.NetworkLoadBalancerForwardingRuleTarget, error) {
	targetsRaw, ok := args["targets"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("targets must be an array")
	}

	targets := make([]ionoscloud.NetworkLoadBalancerForwardingRuleTarget, 0, len(targetsRaw))
	for i, t := range targetsRaw {
		targetMap, ok := t.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("targets[%d] must be an object", i)
		}

		ip, ok := targetMap["ip"].(string)
		if !ok || ip == "" {
			return nil, fmt.Errorf("targets[%d].ip is required", i)
		}
		if err := ionos.ValidateIP(ip); err != nil {
			return nil, fmt.Errorf("targets[%d].ip invalid: %w", i, err)
		}

		portFloat, ok := targetMap["port"].(float64)
		if !ok {
			return nil, fmt.Errorf("targets[%d].port is required", i)
		}
		port := int32(portFloat)
		if port < 1 || port > 65535 {
			return nil, fmt.Errorf("targets[%d].port must be between 1-65535, got %d", i, port)
		}

		weightFloat, ok := targetMap["weight"].(float64)
		if !ok {
			return nil, fmt.Errorf("targets[%d].weight is required", i)
		}
		weight := int32(weightFloat)
		if weight < 0 || weight > 256 {
			return nil, fmt.Errorf("targets[%d].weight must be between 0-256, got %d", i, weight)
		}

		targets = append(targets, ionoscloud.NetworkLoadBalancerForwardingRuleTarget{
			Ip:     &ip,
			Port:   &port,
			Weight: &weight,
		})
	}

	return targets, nil
}
