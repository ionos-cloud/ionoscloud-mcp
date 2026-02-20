package loadbalancing

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ionos-cloud/ionoscloud-mcp/pkg/api"
	"github.com/ionos-cloud/ionoscloud-mcp/pkg/ionos"
	ionoscloud "github.com/ionos-cloud/sdk-go/v6"
)

// Valid target group algorithms
var validTargetGroupAlgorithms = map[string]bool{
	"ROUND_ROBIN": true, "LEAST_CONNECTION": true, "RANDOM": true,
	"SOURCE_IP": true,
}

// Valid target group protocols
var validTargetGroupProtocols = map[string]bool{
	"HTTP": true,
}

func initTargetgroups() []api.ServerTool {
	return []api.ServerTool{
		{
			Tool: api.Tool{
				Name:        "list_target_groups",
				Description: "List all target groups in your IONOS Cloud account",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {},
					"required": []
				}`),
				Annotations: api.ReadOnly("List Target Groups"),
			},
			Handler: listTargetGroups,
		},
		{
			Tool: api.Tool{
				Name:        "get_target_group",
				Description: "Get details of a specific target group",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"target_group_id": {"type": "string", "description": "The ID of the target group"}
					},
					"required": ["target_group_id"]
				}`),
				Annotations: api.ReadOnly("Get Target Group"),
			},
			Handler: getTargetGroup,
		},
		{
			Tool: api.Tool{
				Name:        "create_target_group",
				Description: "Create a target group for load balancing",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"name": {
							"type": "string",
							"description": "The name of the target group"
						},
						"algorithm": {
							"type": "string",
							"description": "Balancing algorithm",
							"enum": ["ROUND_ROBIN", "LEAST_CONNECTION", "RANDOM", "SOURCE_IP"]
						},
						"protocol": {
							"type": "string",
							"description": "The forwarding protocol (HTTP)",
							"enum": ["HTTP"]
						},
						"targets": {
							"type": "array",
							"items": {
								"type": "object",
								"properties": {
									"ip": {"type": "string", "description": "Target IP address"},
									"port": {"type": "integer", "description": "Target port (1-65535)"},
									"weight": {"type": "integer", "description": "Traffic weight (0-256, default: 1)"},
									"health_check_enabled": {"type": "boolean", "description": "Enable health check for this target (default: true)"},
									"maintenance_enabled": {"type": "boolean", "description": "Enable maintenance mode (default: false)"}
								},
								"required": ["ip", "port", "weight"]
							},
							"description": "Array of balanced targets"
						},
						"health_check": {
							"type": "object",
							"properties": {
								"check_timeout": {"type": "integer", "description": "Max wait time in ms for target to respond"},
								"check_interval": {"type": "integer", "description": "Interval in ms between health checks (default: 2000)"},
								"retries": {"type": "integer", "description": "Max reconnect attempts (0-65535, default: 3)"}
							},
							"description": "Health check configuration"
						},
						"http_health_check": {
							"type": "object",
							"properties": {
								"path": {"type": "string", "description": "Health check URL path (default: /)"},
								"method": {"type": "string", "description": "HTTP method (GET, HEAD, POST, PUT, PATCH, OPTIONS)"},
								"match_type": {"type": "string", "description": "Response match type (STATUS_CODE, RESPONSE_BODY)"},
								"response": {"type": "string", "description": "Expected response (status code or body)"},
								"regex": {"type": "boolean", "description": "Use regex for response matching (default: false)"},
								"negate": {"type": "boolean", "description": "Negate the match (default: false)"}
							},
							"description": "HTTP health check configuration"
						}
					},
					"required": ["name", "algorithm", "protocol"]
				}`),
				Annotations: api.NonIdempotent("Create Target Group"),
			},
			Handler: createTargetGroup,
		},
		{
			Tool: api.Tool{
				Name:        "update_target_group",
				Description: "Update a target group",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"target_group_id": {
							"type": "string",
							"description": "The ID of the target group"
						},
						"name": {
							"type": "string",
							"description": "The new name for the target group"
						},
						"algorithm": {
							"type": "string",
							"description": "Balancing algorithm",
							"enum": ["ROUND_ROBIN", "LEAST_CONNECTION", "RANDOM", "SOURCE_IP"]
						},
						"protocol": {
							"type": "string",
							"description": "The forwarding protocol (HTTP)",
							"enum": ["HTTP"]
						},
						"targets": {
							"type": "array",
							"items": {
								"type": "object",
								"properties": {
									"ip": {"type": "string", "description": "Target IP address"},
									"port": {"type": "integer", "description": "Target port (1-65535)"},
									"weight": {"type": "integer", "description": "Traffic weight (0-256, default: 1)"},
									"health_check_enabled": {"type": "boolean", "description": "Enable health check for this target"},
									"maintenance_enabled": {"type": "boolean", "description": "Enable maintenance mode"}
								},
								"required": ["ip", "port", "weight"]
							},
							"description": "Updated array of balanced targets"
						},
						"health_check": {
							"type": "object",
							"properties": {
								"check_timeout": {"type": "integer", "description": "Max wait time in ms for target to respond"},
								"check_interval": {"type": "integer", "description": "Interval in ms between health checks"},
								"retries": {"type": "integer", "description": "Max reconnect attempts (0-65535)"}
							},
							"description": "Updated health check configuration"
						},
						"http_health_check": {
							"type": "object",
							"properties": {
								"path": {"type": "string", "description": "Health check URL path"},
								"method": {"type": "string", "description": "HTTP method"},
								"match_type": {"type": "string", "description": "Response match type"},
								"response": {"type": "string", "description": "Expected response"},
								"regex": {"type": "boolean", "description": "Use regex for response matching"},
								"negate": {"type": "boolean", "description": "Negate the match"}
							},
							"description": "Updated HTTP health check configuration"
						}
					},
					"required": ["target_group_id"]
				}`),
				Annotations: api.Idempotent("Update Target Group"),
			},
			Handler: updateTargetGroup,
		},
		{
			Tool: api.Tool{
				Name:        "delete_target_group",
				Description: "Delete a target group",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {
						"target_group_id": {
							"type": "string",
							"description": "The ID of the target group to delete"
						}
					},
					"required": ["target_group_id"]
				}`),
				Annotations: api.Destructive("Delete Target Group"),
			},
			Handler: deleteTargetGroup,
		},
	}
}

func listTargetGroups(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	targetGroups, _, err := params.Client.TargetGroupsApi.TargetgroupsGet(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to list target groups: %w", err)
	}
	return api.MarshalResult(targetGroups, "target groups")
}

func getTargetGroup(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	targetGroupID, ok := api.GetRequiredString(params.Arguments, "target_group_id")
	if !ok {
		return nil, fmt.Errorf("target_group_id is required")
	}

	targetGroup, _, err := params.Client.TargetGroupsApi.TargetgroupsFindByTargetGroupId(ctx, targetGroupID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get target group: %w", err)
	}
	return api.MarshalResult(targetGroup, "target group")
}

func createTargetGroup(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
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

	if !validTargetGroupAlgorithms[algorithm] {
		return nil, fmt.Errorf("invalid algorithm: %s (valid: ROUND_ROBIN, LEAST_CONNECTION, RANDOM, SOURCE_IP)", algorithm)
	}
	if !validTargetGroupProtocols[protocol] {
		return nil, fmt.Errorf("invalid protocol: %s (valid: HTTP)", protocol)
	}

	properties := ionoscloud.TargetGroupProperties{
		Name:      &name,
		Algorithm: &algorithm,
		Protocol:  &protocol,
	}

	// Parse targets if provided
	if _, ok := params.Arguments["targets"].([]interface{}); ok {
		targets, err := parseTargetGroupTargets(params.Arguments)
		if err != nil {
			return nil, err
		}
		properties.Targets = &targets
	}

	// Parse health check if provided
	if hcMap, ok := params.Arguments["health_check"].(map[string]interface{}); ok {
		hc := parseTargetGroupHealthCheck(hcMap)
		properties.HealthCheck = &hc
	}

	// Parse HTTP health check if provided
	if httpHcMap, ok := params.Arguments["http_health_check"].(map[string]interface{}); ok {
		httpHc := parseTargetGroupHttpHealthCheck(httpHcMap)
		properties.HttpHealthCheck = &httpHc
	}

	targetGroup := ionoscloud.TargetGroup{
		Properties: &properties,
	}

	result, _, err := params.Client.TargetGroupsApi.TargetgroupsPost(ctx).TargetGroup(targetGroup).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to create target group: %w", err)
	}
	return api.MarshalResult(result, "target group")
}

func updateTargetGroup(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	targetGroupID, ok := api.GetRequiredString(params.Arguments, "target_group_id")
	if !ok {
		return nil, fmt.Errorf("target_group_id is required")
	}

	name := api.GetOptionalString(params.Arguments, "name")
	algorithm := api.GetOptionalString(params.Arguments, "algorithm")
	protocol := api.GetOptionalString(params.Arguments, "protocol")
	_, targetsSet := params.Arguments["targets"].([]interface{})
	_, hcSet := params.Arguments["health_check"].(map[string]interface{})
	_, httpHcSet := params.Arguments["http_health_check"].(map[string]interface{})

	if name == "" && algorithm == "" && protocol == "" && !targetsSet && !hcSet && !httpHcSet {
		return nil, fmt.Errorf("at least one field must be provided for update")
	}

	if algorithm != "" && !validTargetGroupAlgorithms[algorithm] {
		return nil, fmt.Errorf("invalid algorithm: %s (valid: ROUND_ROBIN, LEAST_CONNECTION, RANDOM, SOURCE_IP)", algorithm)
	}
	if protocol != "" && !validTargetGroupProtocols[protocol] {
		return nil, fmt.Errorf("invalid protocol: %s (valid: HTTP)", protocol)
	}

	properties := ionoscloud.TargetGroupProperties{}
	if name != "" {
		properties.Name = &name
	}
	if algorithm != "" {
		properties.Algorithm = &algorithm
	}
	if protocol != "" {
		properties.Protocol = &protocol
	}
	if targetsSet {
		targets, err := parseTargetGroupTargets(params.Arguments)
		if err != nil {
			return nil, err
		}
		properties.Targets = &targets
	}
	if hcSet {
		hcMap := params.Arguments["health_check"].(map[string]interface{})
		hc := parseTargetGroupHealthCheck(hcMap)
		properties.HealthCheck = &hc
	}
	if httpHcSet {
		httpHcMap := params.Arguments["http_health_check"].(map[string]interface{})
		httpHc := parseTargetGroupHttpHealthCheck(httpHcMap)
		properties.HttpHealthCheck = &httpHc
	}

	result, _, err := params.Client.TargetGroupsApi.TargetgroupsPatch(ctx, targetGroupID).TargetGroupProperties(properties).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to update target group: %w", err)
	}
	return api.MarshalResult(result, "target group")
}

func deleteTargetGroup(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	targetGroupID, ok := api.GetRequiredString(params.Arguments, "target_group_id")
	if !ok {
		return nil, fmt.Errorf("target_group_id is required")
	}

	_, err := params.Client.TargetGroupsApi.TargetGroupsDelete(ctx, targetGroupID).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to delete target group: %w", err)
	}
	return api.StatusResult(map[string]string{"status": "deleted", "target_group_id": targetGroupID})
}

// parseTargetGroupTargets parses and validates target group targets from arguments.
func parseTargetGroupTargets(args map[string]interface{}) ([]ionoscloud.TargetGroupTarget, error) {
	targetsRaw, ok := args["targets"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("targets must be an array")
	}

	targets := make([]ionoscloud.TargetGroupTarget, 0, len(targetsRaw))
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

		target := ionoscloud.TargetGroupTarget{
			Ip:     &ip,
			Port:   &port,
			Weight: &weight,
		}

		if hcEnabled, ok := targetMap["health_check_enabled"].(bool); ok {
			target.HealthCheckEnabled = &hcEnabled
		}
		if maintEnabled, ok := targetMap["maintenance_enabled"].(bool); ok {
			target.MaintenanceEnabled = &maintEnabled
		}

		targets = append(targets, target)
	}

	return targets, nil
}

// parseTargetGroupHealthCheck parses health check configuration from a map.
func parseTargetGroupHealthCheck(hcMap map[string]interface{}) ionoscloud.TargetGroupHealthCheck {
	hc := ionoscloud.TargetGroupHealthCheck{}
	if v, ok := hcMap["check_timeout"].(float64); ok {
		val := int32(v)
		hc.CheckTimeout = &val
	}
	if v, ok := hcMap["check_interval"].(float64); ok {
		val := int32(v)
		hc.CheckInterval = &val
	}
	if v, ok := hcMap["retries"].(float64); ok {
		val := int32(v)
		hc.Retries = &val
	}
	return hc
}

// parseTargetGroupHttpHealthCheck parses HTTP health check configuration from a map.
func parseTargetGroupHttpHealthCheck(httpHcMap map[string]interface{}) ionoscloud.TargetGroupHttpHealthCheck {
	httpHc := ionoscloud.TargetGroupHttpHealthCheck{}
	if v, ok := httpHcMap["path"].(string); ok && v != "" {
		httpHc.Path = &v
	}
	if v, ok := httpHcMap["method"].(string); ok && v != "" {
		httpHc.Method = &v
	}
	if v, ok := httpHcMap["match_type"].(string); ok && v != "" {
		httpHc.MatchType = &v
	}
	if v, ok := httpHcMap["response"].(string); ok && v != "" {
		httpHc.Response = &v
	}
	if v, ok := httpHcMap["regex"].(bool); ok {
		httpHc.Regex = &v
	}
	if v, ok := httpHcMap["negate"].(bool); ok {
		httpHc.Negate = &v
	}
	return httpHc
}
