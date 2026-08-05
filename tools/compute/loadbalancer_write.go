package compute

import (
	"context"
	"fmt"
	"strings"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

// The classic load balancer balances traffic across NICs attached directly to it,
// rather than through forwarding rules. Unlike the managed balancers, all of its
// properties are optional, so update_loadbalancer is a plain partial PATCH.

// RegisterLoadBalancerWriteTools registers the create/update/delete classic load
// balancer tools.
func RegisterLoadBalancerWriteTools(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	registerCreateLoadBalancer(server, client, scope, confirm)
	registerUpdateLoadBalancer(server, client, scope)
	registerDeleteLoadBalancer(server, client, scope, confirm)
}

func registerCreateLoadBalancer(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterTool(server, scope, tools.MethodPost, &mcp.Tool{
		Name: "create_loadbalancer",
		Description: "Create one classic load balancer. Two-phase: call first WITHOUT confirmation_token to get a preview and a one-time token, then call again WITH the token (and the same datacenter_id and name) to create it. " +
			"The classic load balancer balances traffic across NICs attached directly to it, which is a different model from the network and application load balancers — those forward to IP targets or target groups defined in forwarding rules. " +
			"For new work prefer create_network_loadbalancer (TCP/UDP) or create_application_loadbalancer (HTTP), which offer health checks and finer routing. Creates exactly one per call.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.CreateLoadBalancerInput) (*mcp.CallToolResult, any, error) {
		dcID := strings.TrimSpace(input.DatacenterID)
		name := strings.TrimSpace(input.Name)
		if dcID == "" {
			return tools.ErrorText("datacenter_id is required"), nil, nil
		}
		if name == "" {
			return tools.ErrorText("name is required"), nil, nil
		}
		target := tools.Target(req, dcID, name)

		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "create_loadbalancer", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("create_loadbalancer", "datacenter_id and name", err)), nil, nil
			}
			// LoadbalancerProperties is all-pointer and its ToMap guards every field,
			// so a zero literal sends only what the caller supplied.
			props := &ionos.LoadbalancerProperties{}
			props.SetName(name)
			if input.Ip != nil {
				props.SetIp(*input.Ip)
			}
			if input.Dhcp != nil {
				props.SetDhcp(*input.Dhcp)
			}
			body := ionos.NewLoadbalancer(*props)
			created, _, err := client.LoadBalancersApi.DatacentersLoadbalancersPost(ctx, dcID).Loadbalancer(*body).Execute()
			return tools.ToResult(created, err)
		}

		token, err := confirm.Mint("create_loadbalancer", target)
		if err != nil {
			return nil, nil, err
		}
		return tools.TextResult(tools.Preview{
			Headline: "About to CREATE one classic load balancer:",
			Fields: tools.Fields(
				"datacenter_id", dcID,
				"name", name,
				"ip", tools.OptStr(input.Ip),
				"dhcp", tools.OptBool(input.Dhcp),
			),
			Tool:      "create_loadbalancer",
			Replay:    tools.Fields("datacenter_id", dcID, "name", name),
			TokenNote: "It balances nothing until NICs are attached to it. The token authorizes creating only this load balancer in this data center",
		}.Render(token)), nil, nil
	})
}

func registerUpdateLoadBalancer(server *mcp.Server, client *ionos.APIClient, scope tools.Scope) {
	tools.RegisterTool(server, scope, tools.MethodPatch, &mcp.Tool{
		Name: "update_loadbalancer",
		Description: "Update a classic load balancer's name, listening address or DHCP setting. Applies a partial update (only the fields you provide). " +
			"Unlike the network and application load balancers, every property here is optional in the API, so nothing has to be carried forward — the fields you omit are left exactly as they are.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.UpdateLoadBalancerInput) (*mcp.CallToolResult, any, error) {
		dcID := strings.TrimSpace(input.DatacenterID)
		id := strings.TrimSpace(input.LoadBalancerID)
		if dcID == "" {
			return tools.ErrorText("datacenter_id is required"), nil, nil
		}
		if id == "" {
			return tools.ErrorText("loadbalancer_id is required"), nil, nil
		}
		if input.Name == nil && input.Ip == nil && input.Dhcp == nil {
			return tools.ErrorText("nothing to update: provide at least one of name, ip, dhcp"), nil, nil
		}
		props := &ionos.LoadbalancerProperties{}
		if input.Name != nil {
			props.SetName(*input.Name)
		}
		if input.Ip != nil {
			props.SetIp(*input.Ip)
		}
		if input.Dhcp != nil {
			props.SetDhcp(*input.Dhcp)
		}
		updated, _, err := client.LoadBalancersApi.DatacentersLoadbalancersPatch(ctx, dcID, id).Loadbalancer(*props).Execute()
		return tools.ToResult(updated, err)
	})
}

func registerDeleteLoadBalancer(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterTool(server, scope, tools.MethodDelete, &mcp.Tool{
		Name: "delete_loadbalancer",
		Description: "Delete a classic load balancer. Two-phase: call first WITHOUT confirmation_token to get a blast-radius preview (the NICs it balances across) and a one-time token, then call again WITH the token to delete. " +
			"The NICs themselves are not deleted, but traffic stops being balanced to them, so whatever was reaching the service through this load balancer stops. This is irreversible.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.DeleteLoadBalancerInput) (*mcp.CallToolResult, any, error) {
		dcID := strings.TrimSpace(input.DatacenterID)
		id := strings.TrimSpace(input.LoadBalancerID)
		if dcID == "" {
			return tools.ErrorText("datacenter_id is required"), nil, nil
		}
		if id == "" {
			return tools.ErrorText("loadbalancer_id is required"), nil, nil
		}
		target := tools.Target(req, dcID, id)

		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "delete_loadbalancer", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("delete_loadbalancer", "datacenter_id and loadbalancer_id", err)), nil, nil
			}
			_, err := client.LoadBalancersApi.DatacentersLoadbalancersDelete(ctx, dcID, id).Execute()
			if err != nil {
				return tools.ToResult(nil, err)
			}
			return tools.TextResult(tools.DeletedAsync("load balancer", id)), nil, nil
		}

		lb, _, err := client.LoadBalancersApi.DatacentersLoadbalancersFindById(ctx, dcID, id).Depth(2).Execute()
		if err != nil {
			if tools.IsNotFound(err) {
				return tools.ErrorText(fmt.Sprintf("load balancer %s does not exist in data center %s; nothing to delete", id, dcID)), nil, nil
			}
			return tools.ToResult(nil, err)
		}
		props := lb.GetProperties()
		radius := tools.AffectedRadius()
		if e := lb.Entities; e != nil && e.Balancednics != nil {
			radius.Add("NICs that stop having traffic balanced to them", len(e.Balancednics.Items))
		}
		token, mErr := confirm.Mint("delete_loadbalancer", target)
		if mErr != nil {
			return nil, nil, mErr
		}
		return tools.TextResult(tools.Preview{
			Headline: "About to DELETE a classic load balancer. This is IRREVERSIBLE.\n" +
				"The NICs below are not deleted, but traffic stops being balanced to them, so whatever reached the service through this load balancer stops.",
			Fields: tools.Fields(
				"datacenter_id", dcID,
				"loadbalancer_id", id,
				"name", props.GetName(),
				"ip", props.GetIp(),
			),
			Radius:    radius,
			EmptyNote: "It has no balanced NICs, so it is not currently carrying traffic.",
			Tool:      "delete_loadbalancer",
			Replay:    tools.Fields("datacenter_id", dcID, "loadbalancer_id", id),
			TokenNote: "This token authorizes deleting ONLY this load balancer",
		}.Render(token)), nil, nil
	})
}
