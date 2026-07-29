package compute

import (
	"context"
	"fmt"
	"strings"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

// A target group is a named pool of backends that application load balancer HTTP
// rules forward to. It is account-level, not tied to a data center.

// RegisterTargetGroupWriteTools registers the create/update/delete target group
// tools.
func RegisterTargetGroupWriteTools(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	registerCreateTargetGroup(server, client, scope, confirm)
	registerUpdateTargetGroup(server, client, scope)
	registerDeleteTargetGroup(server, client, scope, confirm)
}

func registerCreateTargetGroup(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterTool(server, scope, tools.MethodPost, &mcp.Tool{
		Name: "create_target_group",
		Description: "Create one target group: a named pool of backends that application load balancer HTTP rules forward to. Two-phase: call first WITHOUT confirmation_token to get a preview and a one-time token, then call again WITH the token (and the same name) to create it. " +
			"A target group belongs to the account rather than a data center, so one group can serve rules on several load balancers. A group with no targets accepts no traffic. " +
			"Reference it from an application load balancer with create_alb_forwarding_rule's http_rules, using a FORWARD rule whose target_group is this group's ID. Creates exactly one group per call.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.CreateTargetGroupInput) (*mcp.CallToolResult, any, error) {
		name := strings.TrimSpace(input.Name)
		algorithm := strings.ToUpper(strings.TrimSpace(input.Algorithm))
		protocol := strings.ToUpper(strings.TrimSpace(input.Protocol))
		if name == "" {
			return tools.ErrorText("name is required"), nil, nil
		}
		if algorithm == "" {
			return tools.ErrorText("algorithm is required: ROUND_ROBIN, LEAST_CONNECTION, RANDOM or SOURCE_IP"), nil, nil
		}
		if protocol == "" {
			return tools.ErrorText("protocol is required: HTTP or TCP"), nil, nil
		}
		if msg := validateTargetGroupTargets(input.Targets); msg != "" {
			return tools.ErrorText(msg), nil, nil
		}
		target := tools.Target(name)

		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "create_target_group", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("create_target_group", "name", err)), nil, nil
			}
			props := ionos.NewTargetGroupProperties(name, algorithm, protocol)
			if input.ProtocolVersion != nil {
				props.SetProtocolVersion(*input.ProtocolVersion)
			}
			if len(input.Targets) > 0 {
				props.SetTargets(buildTargetGroupTargets(input.Targets))
			}
			body := ionos.NewTargetGroup(*props)
			created, _, err := client.TargetGroupsApi.TargetgroupsPost(ctx).TargetGroup(*body).Execute()
			return tools.ToResult(created, err)
		}

		token, err := confirm.Mint("create_target_group", target)
		if err != nil {
			return nil, nil, err
		}
		headline := "About to CREATE one target group:"
		if len(input.Targets) == 0 {
			headline += "\nNOTE: no targets were given, so the group accepts no traffic until you add some with update_target_group."
		}
		return tools.TextResult(tools.Preview{
			Headline: headline,
			Fields: append(tools.Fields(
				"name", name,
				"algorithm", algorithm,
				"protocol", protocol,
				"protocol_version", tools.OptStr(input.ProtocolVersion),
			), targetGroupTargetPreview(input.Targets)...),
			Tool:      "create_target_group",
			Replay:    tools.Fields("name", name),
			TokenNote: "The token authorizes creating only this target group",
		}.Render(token)), nil, nil
	})
}

func registerUpdateTargetGroup(server *mcp.Server, client *ionos.APIClient, scope tools.Scope) {
	tools.RegisterTool(server, scope, tools.MethodPatch, &mcp.Tool{
		Name: "update_target_group",
		Description: "Update a target group's name, algorithm, protocol or backends. Applies a partial update (only the fields you provide). " +
			"Omit name, algorithm or protocol to keep the current value — each is read and sent back unchanged, because the API always receives all three. " +
			"Supplying targets REPLACES the backend list, so include every backend the group should keep: any you leave out stops receiving traffic from every load balancer rule that forwards here.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.UpdateTargetGroupInput) (*mcp.CallToolResult, any, error) {
		id := strings.TrimSpace(input.TargetGroupID)
		if id == "" {
			return tools.ErrorText("target_group_id is required"), nil, nil
		}
		if input.Name == nil && input.Algorithm == nil && input.Protocol == nil &&
			input.ProtocolVersion == nil && input.Targets == nil {
			return tools.ErrorText("nothing to update: provide at least one of name, algorithm, protocol, protocol_version, targets"), nil, nil
		}
		if input.Targets != nil {
			if len(input.Targets) == 0 {
				return tools.ErrorText("targets must contain at least one backend; omit the field entirely to leave the current backends untouched"), nil, nil
			}
			if msg := validateTargetGroupTargets(input.Targets); msg != "" {
				return tools.ErrorText(msg), nil, nil
			}
		}

		// name, algorithm and protocol are serialized unconditionally, so read the
		// current group and override only what the caller supplied.
		current, _, err := client.TargetGroupsApi.TargetgroupsFindByTargetGroupId(ctx, id).Depth(1).Execute()
		if err != nil {
			if tools.IsNotFound(err) {
				return tools.ErrorText(fmt.Sprintf("target group %s does not exist; nothing to update", id)), nil, nil
			}
			return tools.ToResult(nil, err)
		}
		cp := current.GetProperties()
		name, algorithm, protocol := cp.GetName(), cp.GetAlgorithm(), cp.GetProtocol()
		if input.Name != nil {
			if strings.TrimSpace(*input.Name) == "" {
				return tools.ErrorText("name must not be empty; omit it entirely to keep the current name"), nil, nil
			}
			name = *input.Name
		}
		if input.Algorithm != nil {
			algorithm = strings.ToUpper(*input.Algorithm)
		}
		if input.Protocol != nil {
			protocol = strings.ToUpper(*input.Protocol)
		}

		props := ionos.NewTargetGroupProperties(name, algorithm, protocol)
		if input.ProtocolVersion != nil {
			props.SetProtocolVersion(*input.ProtocolVersion)
		} else if cp.HasProtocolVersion() {
			props.SetProtocolVersion(cp.GetProtocolVersion())
		}
		if input.Targets != nil {
			props.SetTargets(buildTargetGroupTargets(input.Targets))
		} else if len(cp.GetTargets()) > 0 {
			// Carrying the backends forward keeps an unrelated change from emptying
			// the pool every load balancer rule forwards to.
			props.SetTargets(cp.GetTargets())
		}
		updated, _, err := client.TargetGroupsApi.TargetgroupsPatch(ctx, id).TargetGroupProperties(*props).Execute()
		return tools.ToResult(updated, err)
	})
}

func registerDeleteTargetGroup(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterTool(server, scope, tools.MethodDelete, &mcp.Tool{
		Name: "delete_target_group",
		Description: "Delete a target group. Two-phase: call first WITHOUT confirmation_token to get a preview of the group and its backends plus a one-time token, then call again WITH the token to delete. " +
			"Any application load balancer HTTP rule that forwards to this group stops working, and because a group is account-level those rules may be on load balancers in several data centers. This is irreversible.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.DeleteTargetGroupInput) (*mcp.CallToolResult, any, error) {
		id := strings.TrimSpace(input.TargetGroupID)
		if id == "" {
			return tools.ErrorText("target_group_id is required"), nil, nil
		}
		target := tools.Target(id)

		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "delete_target_group", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("delete_target_group", "target_group_id", err)), nil, nil
			}
			// Note the casing: the SDK names this method TargetGroupsDelete while
			// every other operation on the resource is Targetgroups*.
			_, err := client.TargetGroupsApi.TargetGroupsDelete(ctx, id).Execute()
			if err != nil {
				return tools.ToResult(nil, err)
			}
			return tools.TextResult(tools.DeletedAsync("target group", id)), nil, nil
		}

		group, _, err := client.TargetGroupsApi.TargetgroupsFindByTargetGroupId(ctx, id).Depth(1).Execute()
		if err != nil {
			if tools.IsNotFound(err) {
				return tools.ErrorText(fmt.Sprintf("target group %s does not exist; nothing to delete", id)), nil, nil
			}
			return tools.ToResult(nil, err)
		}
		cp := group.GetProperties()
		radius := tools.AffectedRadius()
		radius.Add("backends in this group", len(cp.GetTargets()))
		token, mErr := confirm.Mint("delete_target_group", target)
		if mErr != nil {
			return nil, nil, mErr
		}
		return tools.TextResult(tools.Preview{
			Headline: "About to DELETE a target group. This is IRREVERSIBLE.\n" +
				"Any application load balancer HTTP rule forwarding to this group stops working. Target groups are account-level, so those rules may sit on load balancers in several data centers — check before deleting.",
			Fields: tools.Fields(
				"target_group_id", id,
				"name", cp.GetName(),
				"algorithm", cp.GetAlgorithm(),
				"protocol", cp.GetProtocol(),
			),
			Radius:    radius,
			EmptyNote: "This group has no backends.",
			Tool:      "delete_target_group",
			Replay:    tools.Fields("target_group_id", id),
			TokenNote: "This token authorizes deleting ONLY this target group",
		}.Render(token)), nil, nil
	})
}

// validateTargetGroupTargets checks each backend, since the API rejects a malformed
// entry as a whole-request failure that does not say which one was wrong.
func validateTargetGroupTargets(targets []tools.TargetGroupTargetInput) string {
	for i, t := range targets {
		if strings.TrimSpace(t.Ip) == "" {
			return fmt.Sprintf("targets[%d].ip is required", i)
		}
		if t.Port < 1 || t.Port > 65535 {
			return fmt.Sprintf("targets[%d].port must be between 1 and 65535, got %d", i, t.Port)
		}
		if t.Weight < 0 || t.Weight > 256 {
			return fmt.Sprintf("targets[%d].weight must be between 0 and 256, got %d", i, t.Weight)
		}
	}
	return ""
}

func buildTargetGroupTargets(in []tools.TargetGroupTargetInput) []ionos.TargetGroupTarget {
	out := make([]ionos.TargetGroupTarget, 0, len(in))
	for _, t := range in {
		target := ionos.NewTargetGroupTarget(t.Ip, t.Port, t.Weight)
		if t.ProxyProtocol != nil {
			target.SetProxyProtocol(*t.ProxyProtocol)
		}
		if t.HealthCheckEnabled != nil {
			target.SetHealthCheckEnabled(*t.HealthCheckEnabled)
		}
		if t.MaintenanceEnabled != nil {
			target.SetMaintenanceEnabled(*t.MaintenanceEnabled)
		}
		out = append(out, *target)
	}
	return out
}

// targetGroupTargetPreview lists the backends so the caller sees the destinations
// rather than a count.
func targetGroupTargetPreview(targets []tools.TargetGroupTargetInput) []tools.KV {
	out := make([]tools.KV, 0, len(targets))
	for i, t := range targets {
		desc := fmt.Sprintf("%s:%d weight %d", t.Ip, t.Port, t.Weight)
		if t.MaintenanceEnabled != nil && *t.MaintenanceEnabled {
			desc += " (in maintenance, receives no traffic)"
		}
		out = append(out, tools.Fields(fmt.Sprintf("target[%d]", i), desc)...)
	}
	return out
}
