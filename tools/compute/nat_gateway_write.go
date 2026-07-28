package compute

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

// A NAT gateway gives servers on private LANs outbound internet access by
// translating their private addresses to public ones. Both its own model and its
// rules' model serialize their required fields unconditionally, so each update
// reads the current resource and carries those values forward.

// RegisterNatGatewayWriteTools registers create/update/delete for NAT gateways and
// their rules.
func RegisterNatGatewayWriteTools(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	registerNatGatewayCRUD(server, client, scope, confirm)
	registerNatGatewayRuleCRUD(server, client, scope, confirm)
}

func registerNatGatewayCRUD(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	api := client.NATGatewaysApi

	tools.RegisterTool(server, scope, tools.MethodPost, &mcp.Tool{
		Name: "create_nat_gateway",
		Description: "Create one NAT gateway, which gives servers on private LANs outbound internet access by translating their private addresses to public ones. Two-phase: call first WITHOUT confirmation_token to get a preview and a one-time token, then call again WITH the token (and the same datacenter_id and name) to create it. " +
			"public_ips must come from a reserved IP block in the same location (see create_ip_block). The gateway translates nothing until you add a rule with create_nat_gateway_rule, and reaches nothing until it serves at least one LAN. Creates exactly one gateway per call.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.CreateNatGatewayInput) (*mcp.CallToolResult, any, error) {
		dcID := strings.TrimSpace(input.DatacenterID)
		name := strings.TrimSpace(input.Name)
		if dcID == "" {
			return tools.ErrorText("datacenter_id is required"), nil, nil
		}
		if name == "" {
			return tools.ErrorText("name is required"), nil, nil
		}
		if len(input.PublicIps) == 0 {
			return tools.ErrorText("public_ips is required and must contain at least one address: a NAT gateway has nothing to translate to without one. Reserve addresses with create_ip_block."), nil, nil
		}
		target := tools.Target(dcID, name)

		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "create_nat_gateway", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("create_nat_gateway", "datacenter_id and name", err)), nil, nil
			}
			props := ionos.NewNatGatewayProperties(name, input.PublicIps)
			if lans := buildNatGatewayLans(input.Lans); len(lans) > 0 {
				props.SetLans(lans)
			}
			body := ionos.NewNatGateway(*props)
			created, _, err := api.DatacentersNatgatewaysPost(ctx, dcID).NatGateway(*body).Execute()
			return tools.ToResult(created, err)
		}

		token, err := confirm.Mint("create_nat_gateway", target)
		if err != nil {
			return nil, nil, err
		}
		headline := "About to CREATE one NAT gateway:"
		if len(input.Lans) == 0 {
			headline += "\nNOTE: no lans were given, so nothing routes through the gateway yet — add the private LANs it should serve."
		}
		return tools.TextResult(tools.Preview{
			Headline: headline,
			Fields: tools.Fields(
				"datacenter_id", dcID,
				"name", name,
				"public_ips", ipSummary(input.PublicIps),
				"lans served", natGatewayLanSummary(input.Lans),
			),
			Tool:      "create_nat_gateway",
			Replay:    tools.Fields("datacenter_id", dcID, "name", name),
			TokenNote: "It translates nothing until you add a rule with create_nat_gateway_rule. The token authorizes creating only this gateway in this data center",
		}.Render(token)), nil, nil
	})

	tools.RegisterTool(server, scope, tools.MethodPatch, &mcp.Tool{
		Name: "update_nat_gateway",
		Description: "Update a NAT gateway's name, public addresses or served LANs. Applies a partial update (only the fields you provide). " +
			"Omit name or public_ips to keep the current value — both are read and sent back unchanged, because the API always receives them and an empty public_ips would leave the gateway with nothing to translate to. " +
			"Supplying public_ips or lans REPLACES that list, so include every entry the gateway should keep.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.UpdateNatGatewayInput) (*mcp.CallToolResult, any, error) {
		dcID := strings.TrimSpace(input.DatacenterID)
		id := strings.TrimSpace(input.NatGatewayID)
		if dcID == "" {
			return tools.ErrorText("datacenter_id is required"), nil, nil
		}
		if id == "" {
			return tools.ErrorText("natgateway_id is required"), nil, nil
		}
		if input.Name == nil && input.PublicIps == nil && input.Lans == nil {
			return tools.ErrorText("nothing to update: provide at least one of name, public_ips, lans"), nil, nil
		}
		if input.PublicIps != nil && len(input.PublicIps) == 0 {
			return tools.ErrorText("public_ips must contain at least one address; omit the field entirely to leave the current addresses untouched"), nil, nil
		}

		current, _, err := api.DatacentersNatgatewaysFindByNatGatewayId(ctx, dcID, id).Depth(1).Execute()
		if err != nil {
			if tools.IsNotFound(err) {
				return tools.ErrorText(fmt.Sprintf("NAT gateway %s does not exist in data center %s; nothing to update", id, dcID)), nil, nil
			}
			return tools.ToResult(nil, err)
		}
		cp := current.GetProperties()

		// name and publicIps are serialized unconditionally, so carry them forward.
		name, publicIps := cp.GetName(), cp.GetPublicIps()
		if input.Name != nil {
			if strings.TrimSpace(*input.Name) == "" {
				return tools.ErrorText("name must not be empty; omit it entirely to keep the current name"), nil, nil
			}
			name = *input.Name
		}
		if input.PublicIps != nil {
			publicIps = input.PublicIps
		}

		props := ionos.NewNatGatewayProperties(name, publicIps)
		if input.Lans != nil {
			props.SetLans(buildNatGatewayLans(input.Lans))
		} else if len(cp.GetLans()) > 0 {
			// lans is optional and guarded, but carrying it forward keeps an
			// unrelated change from disconnecting the gateway from its LANs.
			props.SetLans(cp.GetLans())
		}
		updated, _, err := api.DatacentersNatgatewaysPatch(ctx, dcID, id).NatGatewayProperties(*props).Execute()
		return tools.ToResult(updated, err)
	})

	tools.RegisterTool(server, scope, tools.MethodDelete, &mcp.Tool{
		Name: "delete_nat_gateway",
		Description: "Delete a NAT gateway and all of its rules. Two-phase: call first WITHOUT confirmation_token to get a blast-radius preview and a one-time token, then call again WITH the token to delete. " +
			"Servers on the LANs it served lose their outbound internet access, which usually breaks package updates, outbound API calls and anything else leaving those LANs. This is irreversible.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.DeleteNatGatewayInput) (*mcp.CallToolResult, any, error) {
		dcID := strings.TrimSpace(input.DatacenterID)
		id := strings.TrimSpace(input.NatGatewayID)
		if dcID == "" {
			return tools.ErrorText("datacenter_id is required"), nil, nil
		}
		if id == "" {
			return tools.ErrorText("natgateway_id is required"), nil, nil
		}
		target := tools.Target(dcID, id)

		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "delete_nat_gateway", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("delete_nat_gateway", "datacenter_id and natgateway_id", err)), nil, nil
			}
			_, err := api.DatacentersNatgatewaysDelete(ctx, dcID, id).Execute()
			if err != nil {
				return tools.ToResult(nil, err)
			}
			return tools.TextResult(tools.DeletedAsync("NAT gateway", id)), nil, nil
		}

		gw, _, err := api.DatacentersNatgatewaysFindByNatGatewayId(ctx, dcID, id).Depth(2).Execute()
		if err != nil {
			if tools.IsNotFound(err) {
				return tools.ErrorText(fmt.Sprintf("NAT gateway %s does not exist in data center %s; nothing to delete", id, dcID)), nil, nil
			}
			return tools.ToResult(nil, err)
		}
		cp := gw.GetProperties()
		radius := &tools.BlastRadius{}
		if e := gw.Entities; e != nil && e.Rules != nil {
			radius.Add("translation rules deleted with it", len(e.Rules.Items))
		}
		radius.Add("LANs that lose outbound internet access", len(cp.GetLans()))
		token, mErr := confirm.Mint("delete_nat_gateway", target)
		if mErr != nil {
			return nil, nil, mErr
		}
		return tools.TextResult(tools.Preview{
			Headline: "About to DELETE a NAT gateway and all of its rules. This is IRREVERSIBLE.\n" +
				"Servers on the LANs below lose their outbound internet access, which usually breaks package updates and any outbound calls they make.",
			Fields: tools.Fields(
				"datacenter_id", dcID,
				"natgateway_id", id,
				"name", cp.GetName(),
				"public_ips", ipSummary(cp.GetPublicIps()),
			),
			Radius:    radius,
			EmptyNote: "It has no rules and serves no LANs, so deleting it affects nothing.",
			Tool:      "delete_nat_gateway",
			Replay:    tools.Fields("datacenter_id", dcID, "natgateway_id", id),
			TokenNote: "This token authorizes deleting ONLY this NAT gateway",
		}.Render(token)), nil, nil
	})
}

func registerNatGatewayRuleCRUD(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	api := client.NATGatewaysApi

	tools.RegisterTool(server, scope, tools.MethodPost, &mcp.Tool{
		Name: "create_nat_gateway_rule",
		Description: "Add one translation rule to a NAT gateway, which is what actually gives a private range outbound internet access. Two-phase: call first WITHOUT confirmation_token to get a preview and a one-time token, then call again WITH the token (and the same parent IDs and name) to create it. " +
			"source_subnet is the private range whose outbound traffic is translated; public_ip must be one of the gateway's own public addresses. Only SNAT is supported. Creates exactly one rule per call.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.CreateNatGatewayRuleInput) (*mcp.CallToolResult, any, error) {
		dcID := strings.TrimSpace(input.DatacenterID)
		gwID := strings.TrimSpace(input.NatGatewayID)
		name := strings.TrimSpace(input.Name)
		sourceSubnet := strings.TrimSpace(input.SourceSubnet)
		publicIP := strings.TrimSpace(input.PublicIp)
		if dcID == "" || gwID == "" {
			return tools.ErrorText("datacenter_id and natgateway_id are both required"), nil, nil
		}
		if name == "" {
			return tools.ErrorText("name is required"), nil, nil
		}
		if sourceSubnet == "" {
			return tools.ErrorText("source_subnet is required: the private range whose outbound traffic is translated, in CIDR form, e.g. 10.0.1.0/24"), nil, nil
		}
		if publicIP == "" {
			return tools.ErrorText("public_ip is required and must be one of the gateway's own public_ips (see get_nat_gateway)"), nil, nil
		}
		if msg := validateNatRulePorts(input.Protocol, input.TargetPortRangeStart, input.TargetPortRangeEnd); msg != "" {
			return tools.ErrorText(msg), nil, nil
		}
		target := tools.Target(dcID, gwID, name)

		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "create_nat_gateway_rule", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("create_nat_gateway_rule", "datacenter_id, natgateway_id and name", err)), nil, nil
			}
			props := ionos.NewNatGatewayRuleProperties(name, sourceSubnet, publicIP)
			applyNatRuleOptionalFields(props, input.Type, input.Protocol, input.TargetSubnet, input.TargetPortRangeStart, input.TargetPortRangeEnd)
			body := ionos.NewNatGatewayRule(*props)
			created, _, err := api.DatacentersNatgatewaysRulesPost(ctx, dcID, gwID).NatGatewayRule(*body).Execute()
			return tools.ToResult(created, err)
		}

		token, err := confirm.Mint("create_nat_gateway_rule", target)
		if err != nil {
			return nil, nil, err
		}
		return tools.TextResult(tools.Preview{
			Headline: "About to CREATE one NAT translation rule:",
			Fields: tools.Fields(
				"datacenter_id", dcID,
				"natgateway_id", gwID,
				"name", name,
				"source_subnet (translated from)", sourceSubnet,
				"public_ip (translated to)", publicIP,
				"type", tools.OptStr(input.Type),
				"protocol", tools.OptStr(input.Protocol),
				"target_subnet", tools.OptStr(input.TargetSubnet),
				"target_port_range", natPortRangeSummary(input.TargetPortRangeStart, input.TargetPortRangeEnd),
			),
			Tool:      "create_nat_gateway_rule",
			Replay:    tools.Fields("datacenter_id", dcID, "natgateway_id", gwID, "name", name),
			TokenNote: "Traffic leaving the source range starts being translated to the public address above. The token authorizes creating only this rule on this gateway",
		}.Render(token)), nil, nil
	})

	tools.RegisterTool(server, scope, tools.MethodPatch, &mcp.Tool{
		Name: "update_nat_gateway_rule",
		Description: "Update a NAT gateway translation rule. Applies a partial update (only the fields you provide). " +
			"Omit name, source_subnet or public_ip to keep the current value — each is read and sent back unchanged, because the API always receives all three. " +
			"Changing source_subnet changes which servers get translated, and changing public_ip changes the address they appear to come from.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.UpdateNatGatewayRuleInput) (*mcp.CallToolResult, any, error) {
		dcID := strings.TrimSpace(input.DatacenterID)
		gwID := strings.TrimSpace(input.NatGatewayID)
		id := strings.TrimSpace(input.RuleID)
		if dcID == "" || gwID == "" || id == "" {
			return tools.ErrorText("datacenter_id, natgateway_id and rule_id are all required"), nil, nil
		}
		if input.Name == nil && input.SourceSubnet == nil && input.PublicIp == nil &&
			input.Type == nil && input.Protocol == nil && input.TargetSubnet == nil &&
			input.TargetPortRangeStart == nil && input.TargetPortRangeEnd == nil {
			return tools.ErrorText("nothing to update: provide at least one of name, source_subnet, public_ip, type, protocol, target_subnet, target_port_range_start, target_port_range_end"), nil, nil
		}
		if msg := validateNatRulePorts(input.Protocol, input.TargetPortRangeStart, input.TargetPortRangeEnd); msg != "" {
			return tools.ErrorText(msg), nil, nil
		}

		current, _, err := api.DatacentersNatgatewaysRulesFindByNatGatewayRuleId(ctx, dcID, gwID, id).Depth(1).Execute()
		if err != nil {
			if tools.IsNotFound(err) {
				return tools.ErrorText(fmt.Sprintf("rule %s does not exist on NAT gateway %s; nothing to update", id, gwID)), nil, nil
			}
			return tools.ToResult(nil, err)
		}
		cp := current.GetProperties()
		name, sourceSubnet, publicIP := cp.GetName(), cp.GetSourceSubnet(), cp.GetPublicIp()
		if input.Name != nil {
			if strings.TrimSpace(*input.Name) == "" {
				return tools.ErrorText("name must not be empty; omit it entirely to keep the current name"), nil, nil
			}
			name = *input.Name
		}
		if input.SourceSubnet != nil {
			sourceSubnet = *input.SourceSubnet
		}
		if input.PublicIp != nil {
			publicIP = *input.PublicIp
		}

		props := ionos.NewNatGatewayRuleProperties(name, sourceSubnet, publicIP)
		applyNatRuleOptionalFields(props, input.Type, input.Protocol, input.TargetSubnet, input.TargetPortRangeStart, input.TargetPortRangeEnd)
		updated, _, err := api.DatacentersNatgatewaysRulesPatch(ctx, dcID, gwID, id).NatGatewayRuleProperties(*props).Execute()
		return tools.ToResult(updated, err)
	})

	tools.RegisterTool(server, scope, tools.MethodDelete, &mcp.Tool{
		Name: "delete_nat_gateway_rule",
		Description: "Delete a translation rule from a NAT gateway. Two-phase: call first WITHOUT confirmation_token to get a preview of what the rule translates plus a one-time token, then call again WITH the token to delete. " +
			"Servers in the rule's source range lose the outbound translation it provided, which usually means losing internet access. This is irreversible.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.DeleteNatGatewayRuleInput) (*mcp.CallToolResult, any, error) {
		dcID := strings.TrimSpace(input.DatacenterID)
		gwID := strings.TrimSpace(input.NatGatewayID)
		id := strings.TrimSpace(input.RuleID)
		if dcID == "" || gwID == "" || id == "" {
			return tools.ErrorText("datacenter_id, natgateway_id and rule_id are all required"), nil, nil
		}
		target := tools.Target(dcID, gwID, id)

		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "delete_nat_gateway_rule", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("delete_nat_gateway_rule", "datacenter_id, natgateway_id and rule_id", err)), nil, nil
			}
			_, err := api.DatacentersNatgatewaysRulesDelete(ctx, dcID, gwID, id).Execute()
			if err != nil {
				return tools.ToResult(nil, err)
			}
			return tools.TextResult(tools.DeletedAsync("NAT gateway rule", id)), nil, nil
		}

		rule, _, err := api.DatacentersNatgatewaysRulesFindByNatGatewayRuleId(ctx, dcID, gwID, id).Depth(1).Execute()
		if err != nil {
			if tools.IsNotFound(err) {
				return tools.ErrorText(fmt.Sprintf("rule %s does not exist on NAT gateway %s; nothing to delete", id, gwID)), nil, nil
			}
			return tools.ToResult(nil, err)
		}
		cp := rule.GetProperties()
		token, mErr := confirm.Mint("delete_nat_gateway_rule", target)
		if mErr != nil {
			return nil, nil, mErr
		}
		return tools.TextResult(tools.Preview{
			Headline: "About to DELETE a NAT translation rule. This is IRREVERSIBLE.\n" +
				"Servers in the source range below lose the outbound translation this rule provided, which usually means losing internet access.",
			Fields: tools.Fields(
				"datacenter_id", dcID,
				"natgateway_id", gwID,
				"rule_id", id,
				"name", cp.GetName(),
				"source_subnet (translated from)", cp.GetSourceSubnet(),
				"public_ip (translated to)", cp.GetPublicIp(),
				"protocol", string(cp.GetProtocol()),
				"target_subnet", cp.GetTargetSubnet(),
			),
			Tool:      "delete_nat_gateway_rule",
			Replay:    tools.Fields("datacenter_id", dcID, "natgateway_id", gwID, "rule_id", id),
			TokenNote: "This token authorizes deleting ONLY this rule from ONLY this gateway",
		}.Render(token)), nil, nil
	})
}

// applyNatRuleOptionalFields sets the optional rule fields the caller supplied.
func applyNatRuleOptionalFields(props *ionos.NatGatewayRuleProperties, ruleType, protocol, targetSubnet *string, portStart, portEnd *int32) {
	if ruleType != nil {
		props.SetType(ionos.NatGatewayRuleType(strings.ToUpper(strings.TrimSpace(*ruleType))))
	}
	if protocol != nil {
		props.SetProtocol(ionos.NatGatewayRuleProtocol(strings.ToUpper(strings.TrimSpace(*protocol))))
	}
	if targetSubnet != nil {
		props.SetTargetSubnet(*targetSubnet)
	}
	if portStart != nil || portEnd != nil {
		r := ionos.NewTargetPortRangeWithDefaults()
		if portStart != nil {
			r.SetStart(*portStart)
		}
		if portEnd != nil {
			r.SetEnd(*portEnd)
		}
		props.SetTargetPortRange(*r)
	}
}

// validateNatRulePorts enforces the API's coupling between protocol and ports: a
// port range only means anything for TCP and UDP, and both ends must be given.
func validateNatRulePorts(protocol *string, start, end *int32) string {
	hasPorts := start != nil || end != nil
	if !hasPorts {
		return ""
	}
	proto := strings.ToUpper(strings.TrimSpace(tools.OptStr(protocol)))
	switch proto {
	case "TCP", "UDP":
		// Ports are meaningful here.
	case "":
		return "target_port_range_start and target_port_range_end require protocol to be set to TCP or UDP"
	default:
		return fmt.Sprintf("target_port_range_start and target_port_range_end are not valid with protocol %s; a port range applies only to TCP and UDP", proto)
	}
	if (start == nil) != (end == nil) {
		return "target_port_range_start and target_port_range_end must be given together"
	}
	if *start > *end {
		return fmt.Sprintf("target_port_range_start (%d) must not be greater than target_port_range_end (%d)", *start, *end)
	}
	for label, v := range map[string]*int32{"target_port_range_start": start, "target_port_range_end": end} {
		if *v < 1 || *v > 65535 {
			return fmt.Sprintf("%s must be between 1 and 65535, got %d", label, *v)
		}
	}
	return ""
}

func buildNatGatewayLans(in []tools.NatGatewayLanInput) []ionos.NatGatewayLanProperties {
	out := make([]ionos.NatGatewayLanProperties, 0, len(in))
	for _, l := range in {
		lan := ionos.NewNatGatewayLanProperties(l.ID)
		if len(l.GatewayIps) > 0 {
			lan.SetGatewayIps(l.GatewayIps)
		}
		out = append(out, *lan)
	}
	return out
}

// natGatewayLanSummary lists the served LAN IDs, which is what identifies them.
func natGatewayLanSummary(lans []tools.NatGatewayLanInput) string {
	if len(lans) == 0 {
		return ""
	}
	parts := make([]string, 0, len(lans))
	for _, l := range lans {
		parts = append(parts, strconv.FormatInt(int64(l.ID), 10))
	}
	return "LAN " + strings.Join(parts, ", LAN ")
}

// natPortRangeSummary renders the destination port range, saying "all ports" when
// unset since that is its effective meaning.
func natPortRangeSummary(start, end *int32) string {
	if start == nil && end == nil {
		return ""
	}
	return fmt.Sprintf("%s-%s", tools.OptInt32(start), tools.OptInt32(end))
}
