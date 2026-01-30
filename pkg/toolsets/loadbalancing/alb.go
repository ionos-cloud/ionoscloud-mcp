package loadbalancing

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ionos-cloud/ionoscloud-mcp/pkg/api"
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
