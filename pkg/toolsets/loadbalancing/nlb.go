package loadbalancing

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ionos-cloud/ionoscloud-mcp/pkg/api"
)

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
