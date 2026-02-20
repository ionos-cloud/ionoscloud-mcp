package loadbalancing

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ionos-cloud/ionoscloud-mcp/pkg/api"
	ionoscloud "github.com/ionos-cloud/sdk-go/v6"
)

func initAlb() []api.ServerTool {
	return []api.ServerTool{
		{
			Tool: api.Tool{
				Name:        "list_application_load_balancers",
				Description: "List all Application Load Balancers in a data center",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {"type": "string", "description": "The ID of the data center"}
					},
					"required": ["datacenter_id"]
				}`),
				Annotations: api.ReadOnly("List ALBs"),
			},
			Handler: listApplicationLoadBalancers,
		},
		{
			Tool: api.Tool{
				Name:        "get_application_load_balancer",
				Description: "Get details of a specific Application Load Balancer",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {"type": "string", "description": "The ID of the data center"},
						"alb_id": {"type": "string", "description": "The ID of the Application Load Balancer"}
					},
					"required": ["datacenter_id", "alb_id"]
				}`),
				Annotations: api.ReadOnly("Get ALB"),
			},
			Handler: getApplicationLoadBalancer,
		},
		{
			Tool: api.Tool{
				Name:        "create_application_load_balancer",
				Description: "Create an Application Load Balancer in a data center",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {
							"type": "string",
							"description": "The ID of the data center"
						},
						"name": {
							"type": "string",
							"description": "The name of the Application Load Balancer"
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
							"description": "Collection of the ALB IP addresses (inbound and outbound IPs of the listenerLan)"
						},
						"lb_private_ips": {
							"type": "array",
							"items": {"type": "string"},
							"description": "Collection of private IP addresses with subnet mask of the ALB"
						}
					},
					"required": ["datacenter_id", "name", "listener_lan", "target_lan"]
				}`),
				Annotations: api.NonIdempotent("Create ALB"),
			},
			Handler: createApplicationLoadBalancer,
		},
		{
			Tool: api.Tool{
				Name:        "update_application_load_balancer",
				Description: "Update an Application Load Balancer",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {
							"type": "string",
							"description": "The ID of the data center"
						},
						"alb_id": {
							"type": "string",
							"description": "The ID of the Application Load Balancer"
						},
						"name": {
							"type": "string",
							"description": "The new name for the ALB"
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
							"description": "Updated collection of ALB IP addresses"
						},
						"lb_private_ips": {
							"type": "array",
							"items": {"type": "string"},
							"description": "Updated collection of private IP addresses with subnet mask"
						}
					},
					"required": ["datacenter_id", "alb_id"]
				}`),
				Annotations: api.Idempotent("Update ALB"),
			},
			Handler: updateApplicationLoadBalancer,
		},
		{
			Tool: api.Tool{
				Name:        "delete_application_load_balancer",
				Description: "Delete an Application Load Balancer",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {
							"type": "string",
							"description": "The ID of the data center"
						},
						"alb_id": {
							"type": "string",
							"description": "The ID of the Application Load Balancer to delete"
						}
					},
					"required": ["datacenter_id", "alb_id"]
				}`),
				Annotations: api.Destructive("Delete ALB"),
			},
			Handler: deleteApplicationLoadBalancer,
		},
		// ALB Forwarding Rules
		{
			Tool: api.Tool{
				Name:        "list_alb_forwarding_rules",
				Description: "List all forwarding rules of an Application Load Balancer",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {"type": "string", "description": "The ID of the data center"},
						"alb_id": {"type": "string", "description": "The ID of the Application Load Balancer"}
					},
					"required": ["datacenter_id", "alb_id"]
				}`),
				Annotations: api.ReadOnly("List ALB Rules"),
			},
			Handler: listAlbForwardingRules,
		},
		{
			Tool: api.Tool{
				Name:        "get_alb_forwarding_rule",
				Description: "Get details of a specific ALB forwarding rule",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {"type": "string", "description": "The ID of the data center"},
						"alb_id": {"type": "string", "description": "The ID of the Application Load Balancer"},
						"rule_id": {"type": "string", "description": "The ID of the forwarding rule"}
					},
					"required": ["datacenter_id", "alb_id", "rule_id"]
				}`),
				Annotations: api.ReadOnly("Get ALB Rule"),
			},
			Handler: getAlbForwardingRule,
		},
		{
			Tool: api.Tool{
				Name:        "create_alb_forwarding_rule",
				Description: "Create a forwarding rule for an Application Load Balancer",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {
							"type": "string",
							"description": "The ID of the data center"
						},
						"alb_id": {
							"type": "string",
							"description": "The ID of the Application Load Balancer"
						},
						"name": {
							"type": "string",
							"description": "The name of the forwarding rule"
						},
						"protocol": {
							"type": "string",
							"description": "The balancing protocol (HTTP)",
							"enum": ["HTTP"]
						},
						"listener_ip": {
							"type": "string",
							"description": "The listening (inbound) IP address"
						},
						"listener_port": {
							"type": "integer",
							"description": "The listening (inbound) port number (1-65535)"
						},
						"client_timeout": {
							"type": "integer",
							"description": "Maximum time in milliseconds to wait for the client (default: 50000)"
						},
						"server_certificates": {
							"type": "array",
							"items": {"type": "string"},
							"description": "Array of server certificate IDs"
						}
					},
					"required": ["datacenter_id", "alb_id", "name", "protocol", "listener_ip", "listener_port"]
				}`),
				Annotations: api.NonIdempotent("Create ALB Rule"),
			},
			Handler: createAlbForwardingRule,
		},
		{
			Tool: api.Tool{
				Name:        "update_alb_forwarding_rule",
				Description: "Update an ALB forwarding rule",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {
							"type": "string",
							"description": "The ID of the data center"
						},
						"alb_id": {
							"type": "string",
							"description": "The ID of the Application Load Balancer"
						},
						"rule_id": {
							"type": "string",
							"description": "The ID of the forwarding rule to update"
						},
						"name": {
							"type": "string",
							"description": "The new name for the forwarding rule"
						},
						"protocol": {
							"type": "string",
							"description": "The balancing protocol (HTTP)",
							"enum": ["HTTP"]
						},
						"listener_ip": {
							"type": "string",
							"description": "The new listening IP address"
						},
						"listener_port": {
							"type": "integer",
							"description": "The new listening port number (1-65535)"
						},
						"client_timeout": {
							"type": "integer",
							"description": "Maximum time in milliseconds to wait for the client"
						},
						"server_certificates": {
							"type": "array",
							"items": {"type": "string"},
							"description": "Updated array of server certificate IDs"
						}
					},
					"required": ["datacenter_id", "alb_id", "rule_id"]
				}`),
				Annotations: api.Idempotent("Update ALB Rule"),
			},
			Handler: updateAlbForwardingRule,
		},
		{
			Tool: api.Tool{
				Name:        "delete_alb_forwarding_rule",
				Description: "Delete an ALB forwarding rule",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {
							"type": "string",
							"description": "The ID of the data center"
						},
						"alb_id": {
							"type": "string",
							"description": "The ID of the Application Load Balancer"
						},
						"rule_id": {
							"type": "string",
							"description": "The ID of the forwarding rule to delete"
						}
					},
					"required": ["datacenter_id", "alb_id", "rule_id"]
				}`),
				Annotations: api.Destructive("Delete ALB Rule"),
			},
			Handler: deleteAlbForwardingRule,
		},
	}
}

func listApplicationLoadBalancers(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}

	albs, _, err := params.Client.ApplicationLoadBalancersApi.DatacentersApplicationloadbalancersGet(ctx, datacenterID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to list Application Load Balancers: %w", err)
	}
	return api.MarshalResult(albs, "Application Load Balancers")
}

func getApplicationLoadBalancer(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	albID, ok := api.GetRequiredString(params.Arguments, "alb_id")
	if !ok {
		return nil, fmt.Errorf("alb_id is required")
	}

	alb, _, err := params.Client.ApplicationLoadBalancersApi.DatacentersApplicationloadbalancersFindByApplicationLoadBalancerId(ctx, datacenterID, albID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get Application Load Balancer: %w", err)
	}
	return api.MarshalResult(alb, "Application Load Balancer")
}

func createApplicationLoadBalancer(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
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

	properties := ionoscloud.ApplicationLoadBalancerProperties{
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

	alb := ionoscloud.ApplicationLoadBalancer{
		Properties: &properties,
	}

	result, _, err := params.Client.ApplicationLoadBalancersApi.DatacentersApplicationloadbalancersPost(ctx, datacenterID).ApplicationLoadBalancer(alb).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to create Application Load Balancer: %w", err)
	}
	return api.MarshalResult(result, "Application Load Balancer")
}

func updateApplicationLoadBalancer(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	albID, ok := api.GetRequiredString(params.Arguments, "alb_id")
	if !ok {
		return nil, fmt.Errorf("alb_id is required")
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

	properties := ionoscloud.ApplicationLoadBalancerProperties{}
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

	result, _, err := params.Client.ApplicationLoadBalancersApi.DatacentersApplicationloadbalancersPatch(ctx, datacenterID, albID).ApplicationLoadBalancerProperties(properties).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to update Application Load Balancer: %w", err)
	}
	return api.MarshalResult(result, "Application Load Balancer")
}

func deleteApplicationLoadBalancer(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	albID, ok := api.GetRequiredString(params.Arguments, "alb_id")
	if !ok {
		return nil, fmt.Errorf("alb_id is required")
	}

	_, err := params.Client.ApplicationLoadBalancersApi.DatacentersApplicationloadbalancersDelete(ctx, datacenterID, albID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to delete Application Load Balancer: %w", err)
	}
	return api.StatusResult(map[string]string{"status": "deleted", "alb_id": albID})
}

func listAlbForwardingRules(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	albID, ok := api.GetRequiredString(params.Arguments, "alb_id")
	if !ok {
		return nil, fmt.Errorf("alb_id is required")
	}

	rules, _, err := params.Client.ApplicationLoadBalancersApi.DatacentersApplicationloadbalancersForwardingrulesGet(ctx, datacenterID, albID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to list ALB forwarding rules: %w", err)
	}
	return api.MarshalResult(rules, "ALB forwarding rules")
}

func getAlbForwardingRule(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	albID, ok := api.GetRequiredString(params.Arguments, "alb_id")
	if !ok {
		return nil, fmt.Errorf("alb_id is required")
	}
	ruleID, ok := api.GetRequiredString(params.Arguments, "rule_id")
	if !ok {
		return nil, fmt.Errorf("rule_id is required")
	}

	rule, _, err := params.Client.ApplicationLoadBalancersApi.DatacentersApplicationloadbalancersForwardingrulesFindByForwardingRuleId(ctx, datacenterID, albID, ruleID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get ALB forwarding rule: %w", err)
	}
	return api.MarshalResult(rule, "ALB forwarding rule")
}

func createAlbForwardingRule(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	albID, ok := api.GetRequiredString(params.Arguments, "alb_id")
	if !ok {
		return nil, fmt.Errorf("alb_id is required")
	}
	name, ok := api.GetRequiredString(params.Arguments, "name")
	if !ok {
		return nil, fmt.Errorf("name is required")
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
	if listenerPort < 1 || listenerPort > 65535 {
		return nil, fmt.Errorf("listener_port must be between 1-65535, got %d", listenerPort)
	}

	if protocol != "HTTP" {
		return nil, fmt.Errorf("invalid protocol: %s (valid: HTTP)", protocol)
	}

	properties := ionoscloud.ApplicationLoadBalancerForwardingRuleProperties{
		Name:         &name,
		Protocol:     &protocol,
		ListenerIp:   &listenerIP,
		ListenerPort: &listenerPort,
	}

	if clientTimeout, ok := api.GetOptionalInt32(params.Arguments, "client_timeout"); ok {
		properties.ClientTimeout = &clientTimeout
	}
	if certs := api.GetStringSlice(params.Arguments, "server_certificates"); len(certs) > 0 {
		properties.ServerCertificates = &certs
	}

	rule := ionoscloud.ApplicationLoadBalancerForwardingRule{
		Properties: &properties,
	}

	result, _, err := params.Client.ApplicationLoadBalancersApi.DatacentersApplicationloadbalancersForwardingrulesPost(ctx, datacenterID, albID).ApplicationLoadBalancerForwardingRule(rule).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to create ALB forwarding rule: %w", err)
	}
	return api.MarshalResult(result, "ALB forwarding rule")
}

func updateAlbForwardingRule(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	albID, ok := api.GetRequiredString(params.Arguments, "alb_id")
	if !ok {
		return nil, fmt.Errorf("alb_id is required")
	}
	ruleID, ok := api.GetRequiredString(params.Arguments, "rule_id")
	if !ok {
		return nil, fmt.Errorf("rule_id is required")
	}

	name := api.GetOptionalString(params.Arguments, "name")
	protocol := api.GetOptionalString(params.Arguments, "protocol")
	listenerIP := api.GetOptionalString(params.Arguments, "listener_ip")
	listenerPort, listenerPortSet := api.GetOptionalInt32(params.Arguments, "listener_port")
	clientTimeout, clientTimeoutSet := api.GetOptionalInt32(params.Arguments, "client_timeout")
	certs := api.GetStringSlice(params.Arguments, "server_certificates")

	if name == "" && protocol == "" && listenerIP == "" && !listenerPortSet && !clientTimeoutSet && len(certs) == 0 {
		return nil, fmt.Errorf("at least one field must be provided for update")
	}

	if protocol != "" && protocol != "HTTP" {
		return nil, fmt.Errorf("invalid protocol: %s (valid: HTTP)", protocol)
	}
	if listenerPortSet && (listenerPort < 1 || listenerPort > 65535) {
		return nil, fmt.Errorf("listener_port must be between 1-65535, got %d", listenerPort)
	}

	properties := ionoscloud.ApplicationLoadBalancerForwardingRuleProperties{}
	if name != "" {
		properties.Name = &name
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
	if clientTimeoutSet {
		properties.ClientTimeout = &clientTimeout
	}
	if len(certs) > 0 {
		properties.ServerCertificates = &certs
	}

	result, _, err := params.Client.ApplicationLoadBalancersApi.DatacentersApplicationloadbalancersForwardingrulesPatch(ctx, datacenterID, albID, ruleID).ApplicationLoadBalancerForwardingRuleProperties(properties).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to update ALB forwarding rule: %w", err)
	}
	return api.MarshalResult(result, "ALB forwarding rule")
}

func deleteAlbForwardingRule(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	albID, ok := api.GetRequiredString(params.Arguments, "alb_id")
	if !ok {
		return nil, fmt.Errorf("alb_id is required")
	}
	ruleID, ok := api.GetRequiredString(params.Arguments, "rule_id")
	if !ok {
		return nil, fmt.Errorf("rule_id is required")
	}

	_, err := params.Client.ApplicationLoadBalancersApi.DatacentersApplicationloadbalancersForwardingrulesDelete(ctx, datacenterID, albID, ruleID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to delete ALB forwarding rule: %w", err)
	}
	return api.StatusResult(map[string]string{"status": "deleted", "rule_id": ruleID})
}
