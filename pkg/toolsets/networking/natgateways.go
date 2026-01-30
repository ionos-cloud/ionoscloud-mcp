package networking

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ionos-cloud/ionoscloud-mcp/pkg/api"
	"github.com/ionos-cloud/ionoscloud-mcp/pkg/ionos"
	ionoscloud "github.com/ionos-cloud/sdk-go/v6"
)

// Valid NAT gateway rule protocols
var validNatProtocols = map[string]bool{
	"TCP": true, "UDP": true, "ICMP": true, "ALL": true,
}

// Valid NAT gateway rule types
var validNatRuleTypes = map[string]bool{
	"SNAT": true,
}

func initNatgateways() []api.ServerTool {
	return []api.ServerTool{
		{
			Tool: api.Tool{
				Name:        "list_nat_gateways",
				Description: "List all NAT gateways in a data center",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {
							"type": "string",
							"description": "The ID of the data center"
						}
					},
					"required": ["datacenter_id"]
				}`),
				Annotations: api.ReadOnly("List NAT Gateways"),
			},
			Handler: listNatGateways,
		},
		{
			Tool: api.Tool{
				Name:        "get_nat_gateway",
				Description: "Get details of a specific NAT gateway",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {
							"type": "string",
							"description": "The ID of the data center"
						},
						"nat_gateway_id": {
							"type": "string",
							"description": "The ID of the NAT gateway"
						}
					},
					"required": ["datacenter_id", "nat_gateway_id"]
				}`),
				Annotations: api.ReadOnly("Get NAT Gateway"),
			},
			Handler: getNatGateway,
		},
		{
			Tool: api.Tool{
				Name:        "create_nat_gateway",
				Description: "Create a NAT gateway in a data center",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {
							"type": "string",
							"description": "The ID of the data center"
						},
						"name": {
							"type": "string",
							"description": "The name of the NAT gateway"
						},
						"public_ips": {
							"type": "array",
							"items": {"type": "string"},
							"description": "Collection of public IP addresses of the NAT gateway"
						},
						"lans": {
							"type": "array",
							"items": {
								"type": "object",
								"properties": {
									"id": {"type": "integer", "description": "LAN ID"},
									"gateway_ips": {"type": "array", "items": {"type": "string"}, "description": "Gateway IPs for this LAN"}
								}
							},
							"description": "Collection of LANs connected to the NAT gateway"
						}
					},
					"required": ["datacenter_id", "name", "public_ips"]
				}`),
				Annotations: api.NonIdempotent("Create NAT Gateway"),
			},
			Handler: createNatGateway,
		},
		{
			Tool: api.Tool{
				Name:        "update_nat_gateway",
				Description: "Update a NAT gateway",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {
							"type": "string",
							"description": "The ID of the data center"
						},
						"nat_gateway_id": {
							"type": "string",
							"description": "The ID of the NAT gateway"
						},
						"name": {
							"type": "string",
							"description": "The new name for the NAT gateway"
						},
						"public_ips": {
							"type": "array",
							"items": {"type": "string"},
							"description": "Updated collection of public IP addresses"
						},
						"lans": {
							"type": "array",
							"items": {
								"type": "object",
								"properties": {
									"id": {"type": "integer", "description": "LAN ID"},
									"gateway_ips": {"type": "array", "items": {"type": "string"}, "description": "Gateway IPs for this LAN"}
								}
							},
							"description": "Updated collection of LANs"
						}
					},
					"required": ["datacenter_id", "nat_gateway_id"]
				}`),
				Annotations: api.Idempotent("Update NAT Gateway"),
			},
			Handler: updateNatGateway,
		},
		{
			Tool: api.Tool{
				Name:        "delete_nat_gateway",
				Description: "Delete a NAT gateway",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {
							"type": "string",
							"description": "The ID of the data center"
						},
						"nat_gateway_id": {
							"type": "string",
							"description": "The ID of the NAT gateway to delete"
						}
					},
					"required": ["datacenter_id", "nat_gateway_id"]
				}`),
				Annotations: api.Destructive("Delete NAT Gateway"),
			},
			Handler: deleteNatGateway,
		},
		// NAT Gateway Rules
		{
			Tool: api.Tool{
				Name:        "list_nat_gateway_rules",
				Description: "List all rules for a NAT gateway",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {
							"type": "string",
							"description": "The ID of the data center"
						},
						"nat_gateway_id": {
							"type": "string",
							"description": "The ID of the NAT gateway"
						}
					},
					"required": ["datacenter_id", "nat_gateway_id"]
				}`),
				Annotations: api.ReadOnly("List NAT Gateway Rules"),
			},
			Handler: listNatGatewayRules,
		},
		{
			Tool: api.Tool{
				Name:        "get_nat_gateway_rule",
				Description: "Get details of a specific NAT gateway rule",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {
							"type": "string",
							"description": "The ID of the data center"
						},
						"nat_gateway_id": {
							"type": "string",
							"description": "The ID of the NAT gateway"
						},
						"rule_id": {
							"type": "string",
							"description": "The ID of the NAT gateway rule"
						}
					},
					"required": ["datacenter_id", "nat_gateway_id", "rule_id"]
				}`),
				Annotations: api.ReadOnly("Get NAT Gateway Rule"),
			},
			Handler: getNatGatewayRule,
		},
		{
			Tool: api.Tool{
				Name:        "create_nat_gateway_rule",
				Description: "Create a rule for a NAT gateway",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {
							"type": "string",
							"description": "The ID of the data center"
						},
						"nat_gateway_id": {
							"type": "string",
							"description": "The ID of the NAT gateway"
						},
						"name": {
							"type": "string",
							"description": "The name of the rule"
						},
						"type": {
							"type": "string",
							"description": "Type of the rule (SNAT)",
							"enum": ["SNAT"]
						},
						"protocol": {
							"type": "string",
							"description": "Protocol (TCP, UDP, ICMP, ALL)",
							"enum": ["TCP", "UDP", "ICMP", "ALL"]
						},
						"source_subnet": {
							"type": "string",
							"description": "Source subnet of the NAT gateway rule (CIDR notation)"
						},
						"public_ip": {
							"type": "string",
							"description": "Public IP address of the NAT gateway rule"
						},
						"target_subnet": {
							"type": "string",
							"description": "Target subnet (for DNAT only)"
						},
						"target_port_range_start": {
							"type": "integer",
							"description": "Target port range start"
						},
						"target_port_range_end": {
							"type": "integer",
							"description": "Target port range end"
						}
					},
					"required": ["datacenter_id", "nat_gateway_id", "name", "source_subnet", "public_ip"]
				}`),
				Annotations: api.NonIdempotent("Create NAT Gateway Rule"),
			},
			Handler: createNatGatewayRule,
		},
		{
			Tool: api.Tool{
				Name:        "update_nat_gateway_rule",
				Description: "Update a NAT gateway rule",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {
							"type": "string",
							"description": "The ID of the data center"
						},
						"nat_gateway_id": {
							"type": "string",
							"description": "The ID of the NAT gateway"
						},
						"rule_id": {
							"type": "string",
							"description": "The ID of the rule to update"
						},
						"name": {
							"type": "string",
							"description": "The new name for the rule"
						},
						"protocol": {
							"type": "string",
							"description": "Protocol (TCP, UDP, ICMP, ALL)",
							"enum": ["TCP", "UDP", "ICMP", "ALL"]
						},
						"source_subnet": {
							"type": "string",
							"description": "Source subnet (CIDR notation)"
						},
						"public_ip": {
							"type": "string",
							"description": "Public IP address"
						},
						"target_subnet": {
							"type": "string",
							"description": "Target subnet"
						},
						"target_port_range_start": {
							"type": "integer",
							"description": "Target port range start"
						},
						"target_port_range_end": {
							"type": "integer",
							"description": "Target port range end"
						}
					},
					"required": ["datacenter_id", "nat_gateway_id", "rule_id"]
				}`),
				Annotations: api.Idempotent("Update NAT Gateway Rule"),
			},
			Handler: updateNatGatewayRule,
		},
		{
			Tool: api.Tool{
				Name:        "delete_nat_gateway_rule",
				Description: "Delete a NAT gateway rule",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"datacenter_id": {
							"type": "string",
							"description": "The ID of the data center"
						},
						"nat_gateway_id": {
							"type": "string",
							"description": "The ID of the NAT gateway"
						},
						"rule_id": {
							"type": "string",
							"description": "The ID of the rule to delete"
						}
					},
					"required": ["datacenter_id", "nat_gateway_id", "rule_id"]
				}`),
				Annotations: api.Destructive("Delete NAT Gateway Rule"),
			},
			Handler: deleteNatGatewayRule,
		},
	}
}

func listNatGateways(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}

	natgateways, _, err := params.Client.NATGatewaysApi.DatacentersNatgatewaysGet(ctx, datacenterID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to list NAT gateways: %w", err)
	}
	return api.MarshalResult(natgateways, "NAT gateways")
}

func getNatGateway(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	natGatewayID, ok := api.GetRequiredString(params.Arguments, "nat_gateway_id")
	if !ok {
		return nil, fmt.Errorf("nat_gateway_id is required")
	}

	natgateway, _, err := params.Client.NATGatewaysApi.DatacentersNatgatewaysFindByNatGatewayId(ctx, datacenterID, natGatewayID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get NAT gateway: %w", err)
	}
	return api.MarshalResult(natgateway, "NAT gateway")
}

func createNatGateway(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	name, ok := api.GetRequiredString(params.Arguments, "name")
	if !ok {
		return nil, fmt.Errorf("name is required")
	}

	publicIPs := api.GetStringSlice(params.Arguments, "public_ips")
	if len(publicIPs) == 0 {
		return nil, fmt.Errorf("public_ips is required")
	}

	// Validate public IPs
	for i, ip := range publicIPs {
		if err := ionos.ValidateIP(ip); err != nil {
			return nil, fmt.Errorf("public_ips[%d] invalid: %w", i, err)
		}
	}

	properties := ionoscloud.NatGatewayProperties{
		Name:      &name,
		PublicIps: &publicIPs,
	}

	// Parse LAN configurations if provided
	if lansRaw, ok := params.Arguments["lans"].([]interface{}); ok && len(lansRaw) > 0 {
		natGatewayLans, err := parseLanConfigurations(lansRaw)
		if err != nil {
			return nil, err
		}
		properties.Lans = &natGatewayLans
	}

	natGateway := ionoscloud.NatGateway{
		Properties: &properties,
	}

	result, _, err := params.Client.NATGatewaysApi.DatacentersNatgatewaysPost(ctx, datacenterID).NatGateway(natGateway).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to create NAT gateway: %w", err)
	}
	return api.MarshalResult(result, "NAT gateway")
}

func updateNatGateway(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	natGatewayID, ok := api.GetRequiredString(params.Arguments, "nat_gateway_id")
	if !ok {
		return nil, fmt.Errorf("nat_gateway_id is required")
	}

	name := api.GetOptionalString(params.Arguments, "name")
	publicIPs := api.GetStringSlice(params.Arguments, "public_ips")
	lansRaw, lansSet := params.Arguments["lans"].([]interface{})

	// Check if at least one field is provided
	if name == "" && len(publicIPs) == 0 && !lansSet {
		return nil, fmt.Errorf("at least one field must be provided for update")
	}

	// Validate public IPs
	for i, ip := range publicIPs {
		if err := ionos.ValidateIP(ip); err != nil {
			return nil, fmt.Errorf("public_ips[%d] invalid: %w", i, err)
		}
	}

	properties := ionoscloud.NatGatewayProperties{}
	if name != "" {
		properties.Name = &name
	}
	if len(publicIPs) > 0 {
		properties.PublicIps = &publicIPs
	}

	// Parse LAN configurations if provided
	if lansSet && len(lansRaw) > 0 {
		natGatewayLans, err := parseLanConfigurations(lansRaw)
		if err != nil {
			return nil, err
		}
		properties.Lans = &natGatewayLans
	}

	result, _, err := params.Client.NATGatewaysApi.DatacentersNatgatewaysPatch(ctx, datacenterID, natGatewayID).NatGatewayProperties(properties).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to update NAT gateway: %w", err)
	}
	return api.MarshalResult(result, "NAT gateway")
}

func deleteNatGateway(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	natGatewayID, ok := api.GetRequiredString(params.Arguments, "nat_gateway_id")
	if !ok {
		return nil, fmt.Errorf("nat_gateway_id is required")
	}

	_, err := params.Client.NATGatewaysApi.DatacentersNatgatewaysDelete(ctx, datacenterID, natGatewayID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to delete NAT gateway: %w", err)
	}
	return api.StatusResult(map[string]string{"status": "deleted", "nat_gateway_id": natGatewayID})
}

// parseLanConfigurations parses and validates LAN configurations for NAT gateways
func parseLanConfigurations(lans []interface{}) ([]ionoscloud.NatGatewayLanProperties, error) {
	natGatewayLans := make([]ionoscloud.NatGatewayLanProperties, len(lans))
	for i, l := range lans {
		lan, ok := l.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("lan[%d] must be an object", i)
		}

		// Validate LAN ID
		idFloat, ok := lan["id"].(float64)
		if !ok {
			return nil, fmt.Errorf("lan[%d].id is required and must be a number", i)
		}
		lanID := int32(idFloat)
		if lanID <= 0 {
			return nil, fmt.Errorf("lan[%d].id must be positive, got %d", i, lanID)
		}

		lanProps := ionoscloud.NatGatewayLanProperties{
			Id: &lanID,
		}

		// Validate gateway IPs if provided
		if gatewayIPs, ok := lan["gateway_ips"].([]interface{}); ok {
			ips := make([]string, len(gatewayIPs))
			for j, ip := range gatewayIPs {
				ipStr, ok := ip.(string)
				if !ok {
					return nil, fmt.Errorf("lan[%d].gateway_ips[%d] must be a string", i, j)
				}
				if err := ionos.ValidateIP(ipStr); err != nil {
					return nil, fmt.Errorf("lan[%d].gateway_ips[%d] invalid: %w", i, j, err)
				}
				ips[j] = ipStr
			}
			lanProps.GatewayIps = &ips
		}

		natGatewayLans[i] = lanProps
	}
	return natGatewayLans, nil
}

func listNatGatewayRules(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	natGatewayID, ok := api.GetRequiredString(params.Arguments, "nat_gateway_id")
	if !ok {
		return nil, fmt.Errorf("nat_gateway_id is required")
	}

	rules, _, err := params.Client.NATGatewaysApi.DatacentersNatgatewaysRulesGet(ctx, datacenterID, natGatewayID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to list NAT gateway rules: %w", err)
	}
	return api.MarshalResult(rules, "NAT gateway rules")
}

func getNatGatewayRule(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	natGatewayID, ok := api.GetRequiredString(params.Arguments, "nat_gateway_id")
	if !ok {
		return nil, fmt.Errorf("nat_gateway_id is required")
	}
	ruleID, ok := api.GetRequiredString(params.Arguments, "rule_id")
	if !ok {
		return nil, fmt.Errorf("rule_id is required")
	}

	rule, _, err := params.Client.NATGatewaysApi.DatacentersNatgatewaysRulesFindByNatGatewayRuleId(ctx, datacenterID, natGatewayID, ruleID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get NAT gateway rule: %w", err)
	}
	return api.MarshalResult(rule, "NAT gateway rule")
}

func createNatGatewayRule(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	natGatewayID, ok := api.GetRequiredString(params.Arguments, "nat_gateway_id")
	if !ok {
		return nil, fmt.Errorf("nat_gateway_id is required")
	}
	name, ok := api.GetRequiredString(params.Arguments, "name")
	if !ok {
		return nil, fmt.Errorf("name is required")
	}
	sourceSubnet, ok := api.GetRequiredString(params.Arguments, "source_subnet")
	if !ok {
		return nil, fmt.Errorf("source_subnet is required")
	}
	publicIP, ok := api.GetRequiredString(params.Arguments, "public_ip")
	if !ok {
		return nil, fmt.Errorf("public_ip is required")
	}

	ruleType := api.GetOptionalString(params.Arguments, "type")
	protocol := api.GetOptionalString(params.Arguments, "protocol")
	targetSubnet := api.GetOptionalString(params.Arguments, "target_subnet")
	targetPortRangeStart, targetPortRangeStartSet := api.GetOptionalInt32(params.Arguments, "target_port_range_start")
	targetPortRangeEnd, targetPortRangeEndSet := api.GetOptionalInt32(params.Arguments, "target_port_range_end")

	// Validate source subnet (CIDR)
	if err := ionos.ValidateIP(sourceSubnet); err != nil {
		return nil, fmt.Errorf("invalid source_subnet: %w", err)
	}
	// Validate public IP
	if err := ionos.ValidateIP(publicIP); err != nil {
		return nil, fmt.Errorf("invalid public_ip: %w", err)
	}
	// Validate protocol if provided
	if protocol != "" && !validNatProtocols[protocol] {
		return nil, fmt.Errorf("invalid protocol: %s (valid: TCP, UDP, ICMP, ALL)", protocol)
	}
	// Validate rule type if provided
	if ruleType != "" && !validNatRuleTypes[ruleType] {
		return nil, fmt.Errorf("invalid type: %s (valid: SNAT)", ruleType)
	}
	// Validate port range
	if err := ionos.ValidatePortRange(targetPortRangeStart, targetPortRangeEnd, "target_port_range"); err != nil {
		return nil, err
	}

	properties := ionoscloud.NatGatewayRuleProperties{
		Name:         &name,
		SourceSubnet: &sourceSubnet,
		PublicIp:     &publicIP,
	}

	if ruleType != "" {
		natRuleType := ionoscloud.NatGatewayRuleType(ruleType)
		properties.Type = &natRuleType
	}
	if protocol != "" {
		natProtocol := ionoscloud.NatGatewayRuleProtocol(protocol)
		properties.Protocol = &natProtocol
	}
	if targetSubnet != "" {
		if err := ionos.ValidateIP(targetSubnet); err != nil {
			return nil, fmt.Errorf("invalid target_subnet: %w", err)
		}
		properties.TargetSubnet = &targetSubnet
	}
	if targetPortRangeStartSet || targetPortRangeEndSet {
		portRange := ionoscloud.TargetPortRange{}
		if targetPortRangeStartSet && targetPortRangeStart > 0 {
			portRange.Start = &targetPortRangeStart
		}
		if targetPortRangeEndSet && targetPortRangeEnd > 0 {
			portRange.End = &targetPortRangeEnd
		}
		properties.TargetPortRange = &portRange
	}

	rule := ionoscloud.NatGatewayRule{
		Properties: &properties,
	}

	result, _, err := params.Client.NATGatewaysApi.DatacentersNatgatewaysRulesPost(ctx, datacenterID, natGatewayID).NatGatewayRule(rule).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to create NAT gateway rule: %w", err)
	}
	return api.MarshalResult(result, "NAT gateway rule")
}

func updateNatGatewayRule(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	natGatewayID, ok := api.GetRequiredString(params.Arguments, "nat_gateway_id")
	if !ok {
		return nil, fmt.Errorf("nat_gateway_id is required")
	}
	ruleID, ok := api.GetRequiredString(params.Arguments, "rule_id")
	if !ok {
		return nil, fmt.Errorf("rule_id is required")
	}

	name := api.GetOptionalString(params.Arguments, "name")
	protocol := api.GetOptionalString(params.Arguments, "protocol")
	sourceSubnet := api.GetOptionalString(params.Arguments, "source_subnet")
	publicIP := api.GetOptionalString(params.Arguments, "public_ip")
	targetSubnet := api.GetOptionalString(params.Arguments, "target_subnet")
	targetPortRangeStart, targetPortRangeStartSet := api.GetOptionalInt32(params.Arguments, "target_port_range_start")
	targetPortRangeEnd, targetPortRangeEndSet := api.GetOptionalInt32(params.Arguments, "target_port_range_end")

	// Check if at least one field is provided
	if name == "" && protocol == "" && sourceSubnet == "" && publicIP == "" && targetSubnet == "" && !targetPortRangeStartSet && !targetPortRangeEndSet {
		return nil, fmt.Errorf("at least one field must be provided for update")
	}

	// Validate inputs
	if sourceSubnet != "" {
		if err := ionos.ValidateIP(sourceSubnet); err != nil {
			return nil, fmt.Errorf("invalid source_subnet: %w", err)
		}
	}
	if publicIP != "" {
		if err := ionos.ValidateIP(publicIP); err != nil {
			return nil, fmt.Errorf("invalid public_ip: %w", err)
		}
	}
	if targetSubnet != "" {
		if err := ionos.ValidateIP(targetSubnet); err != nil {
			return nil, fmt.Errorf("invalid target_subnet: %w", err)
		}
	}
	if protocol != "" && !validNatProtocols[protocol] {
		return nil, fmt.Errorf("invalid protocol: %s (valid: TCP, UDP, ICMP, ALL)", protocol)
	}
	// Validate port range if provided
	if targetPortRangeStartSet || targetPortRangeEndSet {
		if err := ionos.ValidatePortRange(targetPortRangeStart, targetPortRangeEnd, "target_port_range"); err != nil {
			return nil, err
		}
	}

	properties := ionoscloud.NatGatewayRuleProperties{}
	if name != "" {
		properties.Name = &name
	}
	if protocol != "" {
		natProtocol := ionoscloud.NatGatewayRuleProtocol(protocol)
		properties.Protocol = &natProtocol
	}
	if sourceSubnet != "" {
		properties.SourceSubnet = &sourceSubnet
	}
	if publicIP != "" {
		properties.PublicIp = &publicIP
	}
	if targetSubnet != "" {
		properties.TargetSubnet = &targetSubnet
	}
	if targetPortRangeStartSet || targetPortRangeEndSet {
		portRange := ionoscloud.TargetPortRange{}
		if targetPortRangeStartSet {
			portRange.Start = &targetPortRangeStart
		}
		if targetPortRangeEndSet {
			portRange.End = &targetPortRangeEnd
		}
		properties.TargetPortRange = &portRange
	}

	result, _, err := params.Client.NATGatewaysApi.DatacentersNatgatewaysRulesPatch(ctx, datacenterID, natGatewayID, ruleID).NatGatewayRuleProperties(properties).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to update NAT gateway rule: %w", err)
	}
	return api.MarshalResult(result, "NAT gateway rule")
}

func deleteNatGatewayRule(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	datacenterID, ok := api.GetRequiredString(params.Arguments, "datacenter_id")
	if !ok {
		return nil, fmt.Errorf("datacenter_id is required")
	}
	natGatewayID, ok := api.GetRequiredString(params.Arguments, "nat_gateway_id")
	if !ok {
		return nil, fmt.Errorf("nat_gateway_id is required")
	}
	ruleID, ok := api.GetRequiredString(params.Arguments, "rule_id")
	if !ok {
		return nil, fmt.Errorf("rule_id is required")
	}

	_, err := params.Client.NATGatewaysApi.DatacentersNatgatewaysRulesDelete(ctx, datacenterID, natGatewayID, ruleID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to delete NAT gateway rule: %w", err)
	}
	return api.StatusResult(map[string]string{"status": "deleted", "rule_id": ruleID})
}
