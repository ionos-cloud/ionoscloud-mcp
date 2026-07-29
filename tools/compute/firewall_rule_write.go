package compute

import (
	"context"
	"fmt"
	"strings"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

// Firewall rules live either on a single NIC or in a security group, whose members
// all inherit them. The API models both with one body, so validation, body building
// and previews are shared here and only the parent chain differs.

// RegisterFirewallRuleWriteTools registers the NIC-scoped firewall rule tools and
// the security-group rule tools.
func RegisterFirewallRuleWriteTools(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	registerNicFirewallRuleTools(server, client, scope, confirm)
	registerSecurityGroupRuleTools(server, client, scope, confirm)
}

// clearableRuleFields are the rule properties the API models as nullable, where
// null means "do not match on this at all". They are the only fields `clear` accepts.
var clearableRuleFields = map[string]bool{
	"source_ip": true, "target_ip": true, "source_mac": true,
	"ip_version": true, "icmp_type": true, "icmp_code": true,
}

// validateRuleAddress rejects an all-addresses CIDR. The API stores "0.0.0.0/0" as
// the bare "0.0.0.0", which matches nothing, so a rule meant to open a port closes
// it. "Any" is a null field: omit on create, list in clear on update.
func validateRuleAddress(field string, v *string, onUpdate bool) string {
	if v == nil {
		return ""
	}
	addr := strings.TrimSpace(*v)
	if !strings.HasSuffix(addr, "/0") {
		return ""
	}
	remedy := fmt.Sprintf("omit %s entirely", field)
	if onUpdate {
		remedy = fmt.Sprintf("list %q in the clear field", field)
	}
	return fmt.Sprintf("%s = %q means every address, but the API does not store it that way: it keeps the bare network address (0.0.0.0), which is non-routable and matches NO traffic — the rule would silently allow nothing. To match any address, %s instead.",
		field, addr, remedy)
}

// validateRuleClear checks the clear list: every name must be a nullable field, and
// a field cannot be both set and cleared in the same call.
func validateRuleClear(clear []string, f tools.RuleFields) string {
	for _, name := range clear {
		key := strings.ToLower(strings.TrimSpace(name))
		if !clearableRuleFields[key] {
			return fmt.Sprintf("clear contains %q, which is not a clearable field. Only these can be reset to 'any': source_ip, target_ip, source_mac, ip_version, icmp_type, icmp_code", name)
		}
		set := map[string]bool{
			"source_ip":  f.SourceIp != nil,
			"target_ip":  f.TargetIp != nil,
			"source_mac": f.SourceMac != nil,
			"ip_version": f.IpVersion != nil,
			"icmp_type":  f.IcmpType != nil,
			"icmp_code":  f.IcmpCode != nil,
		}
		if set[key] {
			return fmt.Sprintf("%s is both given a value and listed in clear; pick one — set it to match on that value, or clear it to stop matching on it", key)
		}
	}
	return ""
}

// validateRuleFields rejects the combinations the API refuses without naming the
// offending field. requireProtocol is set for creates, which need one.
func validateRuleFields(f tools.RuleFields, requireProtocol bool) string {
	protocol := strings.ToUpper(strings.TrimSpace(tools.OptStr(f.Protocol)))
	if requireProtocol && protocol == "" {
		return "protocol is required to create a firewall rule: TCP, UDP, ICMP, ICMPv6, GRE, VRRP, ESP, AH or ANY"
	}
	// onUpdate is the inverse of requireProtocol, which is the only signal available
	// here for which flavour of remedy to suggest.
	for field, v := range map[string]*string{"source_ip": f.SourceIp, "target_ip": f.TargetIp} {
		if msg := validateRuleAddress(field, v, !requireProtocol); msg != "" {
			return msg
		}
	}

	hasPorts := f.PortRangeStart != nil || f.PortRangeEnd != nil
	hasIcmp := f.IcmpType != nil || f.IcmpCode != nil

	// Ports belong to TCP and UDP; ICMP type/code to ICMP and ICMPv6. Mixing them
	// is rejected by the API with a message that does not say which field is wrong.
	switch protocol {
	case "TCP", "UDP":
		if hasIcmp {
			return fmt.Sprintf("icmp_type and icmp_code are not valid with protocol %s; they apply only to ICMP and ICMPv6", protocol)
		}
	case "ICMP", "ICMPv6", "ICMPV6":
		if hasPorts {
			return fmt.Sprintf("port_range_start and port_range_end are not valid with protocol %s; they apply only to TCP and UDP", protocol)
		}
	case "":
		// Update with no protocol change: the existing protocol governs, so leave
		// the combination to the API rather than guessing against stored state.
	default:
		if hasPorts {
			return fmt.Sprintf("port_range_start and port_range_end are not valid with protocol %s; they apply only to TCP and UDP", protocol)
		}
		if hasIcmp {
			return fmt.Sprintf("icmp_type and icmp_code are not valid with protocol %s; they apply only to ICMP and ICMPv6", protocol)
		}
	}

	// A half-open port range is silently unhelpful rather than an error at the API,
	// so catch it here.
	if (f.PortRangeStart == nil) != (f.PortRangeEnd == nil) {
		return "port_range_start and port_range_end must be given together; to allow every port, omit both"
	}
	if f.PortRangeStart != nil && f.PortRangeEnd != nil && *f.PortRangeStart > *f.PortRangeEnd {
		return fmt.Sprintf("port_range_start (%d) must not be greater than port_range_end (%d)", *f.PortRangeStart, *f.PortRangeEnd)
	}
	for label, v := range map[string]*int32{"port_range_start": f.PortRangeStart, "port_range_end": f.PortRangeEnd} {
		if v != nil && (*v < 1 || *v > 65534) {
			return fmt.Sprintf("%s must be between 1 and 65534, got %d", label, *v)
		}
	}
	for label, v := range map[string]*int32{"icmp_type": f.IcmpType, "icmp_code": f.IcmpCode} {
		if v != nil && (*v < 0 || *v > 254) {
			return fmt.Sprintf("%s must be between 0 and 254, got %d", label, *v)
		}
	}
	return ""
}

// buildRuleProperties sets only what the caller supplied. FirewallruleProperties is
// all-pointer with a fully guarded ToMap, so a zero literal sends nothing extra.
func buildRuleProperties(f tools.RuleFields, clear []string) *ionos.FirewallruleProperties {
	props := &ionos.FirewallruleProperties{}
	// An explicit null is how the API expresses "do not match on this field".
	// Without it these fields are set-only and a rule can never be widened again.
	for _, name := range clear {
		switch strings.ToLower(strings.TrimSpace(name)) {
		case "source_ip":
			props.SetSourceIpNil()
		case "target_ip":
			props.SetTargetIpNil()
		case "source_mac":
			props.SetSourceMacNil()
		case "ip_version":
			props.SetIpVersionNil()
		case "icmp_type":
			props.SetIcmpTypeNil()
		case "icmp_code":
			props.SetIcmpCodeNil()
		}
	}
	if f.Name != nil {
		props.SetName(*f.Name)
	}
	if f.Protocol != nil {
		props.SetProtocol(strings.ToUpper(strings.TrimSpace(*f.Protocol)))
	}
	if f.Type != nil {
		props.SetType(strings.ToUpper(strings.TrimSpace(*f.Type)))
	}
	if f.SourceMac != nil {
		props.SetSourceMac(*f.SourceMac)
	}
	if f.SourceIp != nil {
		props.SetSourceIp(*f.SourceIp)
	}
	if f.TargetIp != nil {
		props.SetTargetIp(*f.TargetIp)
	}
	if f.IpVersion != nil {
		props.SetIpVersion(*f.IpVersion)
	}
	if f.PortRangeStart != nil {
		props.SetPortRangeStart(*f.PortRangeStart)
	}
	if f.PortRangeEnd != nil {
		props.SetPortRangeEnd(*f.PortRangeEnd)
	}
	if f.IcmpType != nil {
		props.SetIcmpType(*f.IcmpType)
	}
	if f.IcmpCode != nil {
		props.SetIcmpCode(*f.IcmpCode)
	}
	return props
}

// ruleFieldsPreview renders the rule for a preview. Unset fields render empty and
// are dropped, so the preview shows exactly what the rule will match.
func ruleFieldsPreview(f tools.RuleFields) []tools.KV {
	return tools.Fields(
		"protocol", tools.OptStr(f.Protocol),
		"name", tools.OptStr(f.Name),
		"direction (type)", tools.OptStr(f.Type),
		"source_ip", tools.OptStr(f.SourceIp),
		"target_ip", tools.OptStr(f.TargetIp),
		"source_mac", tools.OptStr(f.SourceMac),
		"ip_version", tools.OptStr(f.IpVersion),
		"port_range", portRangeSummary(f.PortRangeStart, f.PortRangeEnd),
		"icmp_type", tools.OptInt32(f.IcmpType),
		"icmp_code", tools.OptInt32(f.IcmpCode),
	)
}

// portRangeSummary renders the port range as one field, and says "all ports" when
// unset because that is the security-relevant reading of an omitted range.
func portRangeSummary(start, end *int32) string {
	if start == nil && end == nil {
		return "all ports"
	}
	return fmt.Sprintf("%s-%s", tools.OptInt32(start), tools.OptInt32(end))
}

// ruleSummary describes an existing rule for a delete preview, so the caller sees
// what traffic stops being allowed rather than just an opaque ID.
func ruleSummary(props ionos.FirewallruleProperties) []tools.KV {
	var ports string
	if props.HasPortRangeStart() || props.HasPortRangeEnd() {
		ports = fmt.Sprintf("%d-%d", props.GetPortRangeStart(), props.GetPortRangeEnd())
	} else {
		ports = "all ports"
	}
	return tools.Fields(
		"name", props.GetName(),
		"protocol", props.GetProtocol(),
		"direction (type)", props.GetType(),
		"source_ip", props.GetSourceIp(),
		"target_ip", props.GetTargetIp(),
		"port_range", ports,
	)
}

func registerNicFirewallRuleTools(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterTool(server, scope, tools.MethodPost, &mcp.Tool{
		Name: "create_firewall_rule",
		Description: "Add one firewall rule to a NIC. Two-phase: call first WITHOUT confirmation_token to get a preview and a one-time token, then call again WITH the token (and the same parent IDs and protocol) to create it. " +
			"Rules are allow-rules: an active firewall permits only what its rules match, so the first rule you add to a NIC is what makes it reachable at all. Omitting port_range_start and port_range_end allows EVERY port for that protocol. " +
			"For a rule that should apply to several servers or NICs, put it in a security group instead (create_security_group_rule) rather than repeating it per NIC. Creates exactly one rule per call.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.CreateFirewallRuleInput) (*mcp.CallToolResult, any, error) {
		dcID := strings.TrimSpace(input.DatacenterID)
		serverID := strings.TrimSpace(input.ServerID)
		nicID := strings.TrimSpace(input.NicID)
		if dcID == "" || serverID == "" || nicID == "" {
			return tools.ErrorText("datacenter_id, server_id and nic_id are all required"), nil, nil
		}
		if msg := validateRuleFields(input.RuleFields, true); msg != "" {
			return tools.ErrorText(msg), nil, nil
		}
		protocol := strings.ToUpper(strings.TrimSpace(tools.OptStr(input.Protocol)))
		target := tools.Target(dcID, serverID, nicID, protocol, tools.OptStr(input.Name))

		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "create_firewall_rule", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("create_firewall_rule", "the same parent IDs, protocol and name", err)), nil, nil
			}
			body := ionos.NewFirewallRule(*buildRuleProperties(input.RuleFields, nil))
			created, _, err := client.FirewallRulesApi.DatacentersServersNicsFirewallrulesPost(ctx, dcID, serverID, nicID).Firewallrule(*body).Execute()
			return tools.ToResult(created, err)
		}

		token, err := confirm.Mint("create_firewall_rule", target)
		if err != nil {
			return nil, nil, err
		}
		fields := tools.Fields("datacenter_id", dcID, "server_id", serverID, "nic_id", nicID)
		return tools.TextResult(tools.Preview{
			Headline:  "About to CREATE one firewall rule on a NIC, allowing the traffic it matches:",
			Fields:    append(fields, ruleFieldsPreview(input.RuleFields)...),
			Tool:      "create_firewall_rule",
			Replay:    tools.Fields("datacenter_id", dcID, "server_id", serverID, "nic_id", nicID, "protocol", protocol),
			TokenNote: "This affects only this one NIC. The token authorizes creating only this rule on it",
		}.Render(token)), nil, nil
	})

	tools.RegisterTool(server, scope, tools.MethodPatch, &mcp.Tool{
		Name: "update_firewall_rule",
		Description: "Update a firewall rule on a NIC. Applies a partial update (only the fields you provide). " +
			"To WIDEN a rule back to 'any' — for example so it matches every source address — list the field in clear; that is the only way, since omitting a field leaves it unchanged and no value means anywhere. Clearing source_ip on an INGRESS rule opens it to the whole internet, so say so first. Narrowing a rule can cut off a service that depends on it. " +
			"Ports apply only to TCP and UDP; icmp_type and icmp_code only to ICMP and ICMPv6.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.UpdateFirewallRuleInput) (*mcp.CallToolResult, any, error) {
		dcID := strings.TrimSpace(input.DatacenterID)
		serverID := strings.TrimSpace(input.ServerID)
		nicID := strings.TrimSpace(input.NicID)
		id := strings.TrimSpace(input.FirewallRuleID)
		if dcID == "" || serverID == "" || nicID == "" || id == "" {
			return tools.ErrorText("datacenter_id, server_id, nic_id and firewallrule_id are all required"), nil, nil
		}
		if msg := validateRuleFields(input.RuleFields, false); msg != "" {
			return tools.ErrorText(msg), nil, nil
		}
		if msg := validateRuleClear(input.Clear, input.RuleFields); msg != "" {
			return tools.ErrorText(msg), nil, nil
		}
		if !anyRuleFieldSet(input.RuleFields) && len(input.Clear) == 0 {
			return tools.ErrorText("nothing to update: provide at least one of name, protocol, type, source_ip, target_ip, source_mac, ip_version, port_range_start, port_range_end, icmp_type, icmp_code, or list fields in clear"), nil, nil
		}
		props := buildRuleProperties(input.RuleFields, input.Clear)
		updated, _, err := client.FirewallRulesApi.DatacentersServersNicsFirewallrulesPatch(ctx, dcID, serverID, nicID, id).Firewallrule(*props).Execute()
		return tools.ToResult(updated, err)
	})

	tools.RegisterTool(server, scope, tools.MethodDelete, &mcp.Tool{
		Name: "delete_firewall_rule",
		Description: "Delete a firewall rule from a NIC. Two-phase: call first WITHOUT confirmation_token to get a preview of exactly what the rule allows plus a one-time token, then call again WITH the token to delete. " +
			"The traffic this rule was permitting is blocked afterwards. If it is the NIC's last rule and its firewall is active, ALL incoming traffic to that NIC is blocked. This is irreversible.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.DeleteFirewallRuleInput) (*mcp.CallToolResult, any, error) {
		dcID := strings.TrimSpace(input.DatacenterID)
		serverID := strings.TrimSpace(input.ServerID)
		nicID := strings.TrimSpace(input.NicID)
		id := strings.TrimSpace(input.FirewallRuleID)
		if dcID == "" || serverID == "" || nicID == "" || id == "" {
			return tools.ErrorText("datacenter_id, server_id, nic_id and firewallrule_id are all required"), nil, nil
		}
		target := tools.Target(dcID, serverID, nicID, id)

		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "delete_firewall_rule", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("delete_firewall_rule", "the same parent IDs and firewallrule_id", err)), nil, nil
			}
			_, err := client.FirewallRulesApi.DatacentersServersNicsFirewallrulesDelete(ctx, dcID, serverID, nicID, id).Execute()
			if err != nil {
				return tools.ToResult(nil, err)
			}
			return tools.TextResult(tools.DeletedAsync("firewall rule", id)), nil, nil
		}

		rule, _, err := client.FirewallRulesApi.DatacentersServersNicsFirewallrulesFindById(ctx, dcID, serverID, nicID, id).Depth(1).Execute()
		if err != nil {
			if tools.IsNotFound(err) {
				return tools.ErrorText(fmt.Sprintf("firewall rule %s does not exist on NIC %s; nothing to delete", id, nicID)), nil, nil
			}
			return tools.ToResult(nil, err)
		}
		token, mErr := confirm.Mint("delete_firewall_rule", target)
		if mErr != nil {
			return nil, nil, mErr
		}
		fields := tools.Fields("datacenter_id", dcID, "server_id", serverID, "nic_id", nicID, "firewallrule_id", id)
		return tools.TextResult(tools.Preview{
			Headline: "About to DELETE a firewall rule from a NIC. This is IRREVERSIBLE.\n" +
				"The traffic described below stops being allowed. If this is the NIC's last rule and its firewall is active, ALL incoming traffic to it will be blocked.",
			Fields:    append(fields, ruleSummary(rule.GetProperties())...),
			Tool:      "delete_firewall_rule",
			Replay:    tools.Fields("datacenter_id", dcID, "server_id", serverID, "nic_id", nicID, "firewallrule_id", id),
			TokenNote: "This token authorizes deleting ONLY this rule from ONLY this NIC",
		}.Render(token)), nil, nil
	})
}

func registerSecurityGroupRuleTools(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterTool(server, scope, tools.MethodPost, &mcp.Tool{
		Name: "create_security_group_rule",
		Description: "Add one firewall rule to a security group, so every server and NIC assigned to that group inherits it. Two-phase: call first WITHOUT confirmation_token to get a preview (including how many servers and NICs will inherit the rule) and a one-time token, then call again WITH the token to create it. " +
			"This is the tool to use for a rule that should apply to several resources, rather than repeating create_firewall_rule per NIC. Omitting port_range_start and port_range_end allows EVERY port for that protocol. Creates exactly one rule per call.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.CreateSecurityGroupRuleInput) (*mcp.CallToolResult, any, error) {
		dcID := strings.TrimSpace(input.DatacenterID)
		groupID := strings.TrimSpace(input.SecurityGroupID)
		if dcID == "" || groupID == "" {
			return tools.ErrorText("datacenter_id and security_group_id are both required"), nil, nil
		}
		if msg := validateRuleFields(input.RuleFields, true); msg != "" {
			return tools.ErrorText(msg), nil, nil
		}
		protocol := strings.ToUpper(strings.TrimSpace(tools.OptStr(input.Protocol)))
		target := tools.Target(dcID, groupID, protocol, tools.OptStr(input.Name))

		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "create_security_group_rule", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("create_security_group_rule", "datacenter_id, security_group_id, protocol and name", err)), nil, nil
			}
			body := ionos.NewFirewallRule(*buildRuleProperties(input.RuleFields, nil))
			created, _, err := client.SecurityGroupsApi.DatacentersSecuritygroupsFirewallrulesPost(ctx, dcID, groupID).FirewallRule(*body).Execute()
			return tools.ToResult(created, err)
		}

		// Reading the group first lets the preview say how far the rule reaches,
		// which is the difference between this tool and create_firewall_rule.
		members := securityGroupMemberCounts(ctx, client, dcID, groupID)
		token, err := confirm.Mint("create_security_group_rule", target)
		if err != nil {
			return nil, nil, err
		}
		fields := tools.Fields("datacenter_id", dcID, "security_group_id", groupID)
		return tools.TextResult(tools.Preview{
			Headline:  "About to CREATE one firewall rule in a security group, allowing the traffic it matches:",
			Fields:    append(fields, ruleFieldsPreview(input.RuleFields)...),
			Radius:    members,
			EmptyNote: "No servers or NICs are assigned to this group yet, so the rule has no effect until you assign it with assign_server_security_groups or assign_nic_security_groups.",
			Tool:      "create_security_group_rule",
			Replay:    tools.Fields("datacenter_id", dcID, "security_group_id", groupID, "protocol", protocol),
			TokenNote: "Every server and NIC listed above starts allowing this traffic at once. The token authorizes creating only this rule in this group",
		}.Render(token)), nil, nil
	})

	tools.RegisterTool(server, scope, tools.MethodPatch, &mcp.Tool{
		Name: "update_security_group_rule",
		Description: "Update a firewall rule in a security group. Applies a partial update (only the fields you provide). " +
			"The change takes effect at once for EVERY server and NIC assigned to the group, so widening a rule here exposes all of them and narrowing it can cut off a service on any of them. To widen a field back to 'any', list it in clear — omitting it leaves it unchanged. " +
			"Ports apply only to TCP and UDP; icmp_type and icmp_code only to ICMP and ICMPv6.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.UpdateSecurityGroupRuleInput) (*mcp.CallToolResult, any, error) {
		dcID := strings.TrimSpace(input.DatacenterID)
		groupID := strings.TrimSpace(input.SecurityGroupID)
		id := strings.TrimSpace(input.RuleID)
		if dcID == "" || groupID == "" || id == "" {
			return tools.ErrorText("datacenter_id, security_group_id and rule_id are all required"), nil, nil
		}
		if msg := validateRuleFields(input.RuleFields, false); msg != "" {
			return tools.ErrorText(msg), nil, nil
		}
		if msg := validateRuleClear(input.Clear, input.RuleFields); msg != "" {
			return tools.ErrorText(msg), nil, nil
		}
		if !anyRuleFieldSet(input.RuleFields) && len(input.Clear) == 0 {
			return tools.ErrorText("nothing to update: provide at least one of name, protocol, type, source_ip, target_ip, source_mac, ip_version, port_range_start, port_range_end, icmp_type, icmp_code, or list fields in clear"), nil, nil
		}
		props := buildRuleProperties(input.RuleFields, input.Clear)
		updated, _, err := client.SecurityGroupsApi.DatacentersSecuritygroupsRulesPatch(ctx, dcID, groupID, id).Rule(*props).Execute()
		return tools.ToResult(updated, err)
	})

	tools.RegisterTool(server, scope, tools.MethodDelete, &mcp.Tool{
		Name: "delete_security_group_rule",
		Description: "Delete a firewall rule from a security group. Two-phase: call first WITHOUT confirmation_token to get a preview of what the rule allows and how many servers and NICs lose it, plus a one-time token, then call again WITH the token to delete. " +
			"Every member of the group stops allowing this traffic at once, so this can cut off a service on several servers simultaneously. This is irreversible.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.DeleteSecurityGroupRuleInput) (*mcp.CallToolResult, any, error) {
		dcID := strings.TrimSpace(input.DatacenterID)
		groupID := strings.TrimSpace(input.SecurityGroupID)
		id := strings.TrimSpace(input.RuleID)
		if dcID == "" || groupID == "" || id == "" {
			return tools.ErrorText("datacenter_id, security_group_id and rule_id are all required"), nil, nil
		}
		target := tools.Target(dcID, groupID, id)

		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "delete_security_group_rule", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("delete_security_group_rule", "datacenter_id, security_group_id and rule_id", err)), nil, nil
			}
			_, err := client.SecurityGroupsApi.DatacentersSecuritygroupsFirewallrulesDelete(ctx, dcID, groupID, id).Execute()
			if err != nil {
				return tools.ToResult(nil, err)
			}
			return tools.TextResult(tools.DeletedAsync("security group rule", id)), nil, nil
		}

		rule, _, err := client.SecurityGroupsApi.DatacentersSecuritygroupsRulesFindById(ctx, dcID, groupID, id).Depth(1).Execute()
		if err != nil {
			if tools.IsNotFound(err) {
				return tools.ErrorText(fmt.Sprintf("rule %s does not exist in security group %s; nothing to delete", id, groupID)), nil, nil
			}
			return tools.ToResult(nil, err)
		}
		members := securityGroupMemberCounts(ctx, client, dcID, groupID)
		token, mErr := confirm.Mint("delete_security_group_rule", target)
		if mErr != nil {
			return nil, nil, mErr
		}
		fields := tools.Fields("datacenter_id", dcID, "security_group_id", groupID, "rule_id", id)
		return tools.TextResult(tools.Preview{
			Headline: "About to DELETE a firewall rule from a security group. This is IRREVERSIBLE.\n" +
				"Every server and NIC assigned to the group stops allowing the traffic described below, all at once.",
			Fields:    append(fields, ruleSummary(rule.GetProperties())...),
			Radius:    members,
			EmptyNote: "No servers or NICs are assigned to this group, so deleting the rule changes nothing in practice.",
			Tool:      "delete_security_group_rule",
			Replay:    tools.Fields("datacenter_id", dcID, "security_group_id", groupID, "rule_id", id),
			TokenNote: "This token authorizes deleting ONLY this rule from ONLY this group",
		}.Render(token)), nil, nil
	})
}

// anyRuleFieldSet reports whether the caller supplied any rule property, so an
// update with nothing in it is rejected rather than sent as an empty PATCH.
func anyRuleFieldSet(f tools.RuleFields) bool {
	return f.Name != nil || f.Protocol != nil || f.Type != nil || f.SourceMac != nil ||
		f.SourceIp != nil || f.TargetIp != nil || f.IpVersion != nil ||
		f.PortRangeStart != nil || f.PortRangeEnd != nil || f.IcmpType != nil || f.IcmpCode != nil
}

// securityGroupMemberCounts reports how far a group rule reaches. A failed read is
// not fatal: the preview just shows no counts.
func securityGroupMemberCounts(ctx context.Context, client *ionos.APIClient, dcID, groupID string) *tools.BlastRadius {
	r := tools.AffectedRadius()
	group, _, err := client.SecurityGroupsApi.DatacentersSecuritygroupsFindById(ctx, dcID, groupID).Depth(2).Execute()
	if err != nil {
		return r
	}
	if e := group.Entities; e != nil {
		if e.Servers != nil {
			r.Add("servers assigned to this group", len(e.Servers.Items))
		}
		if e.Nics != nil {
			r.Add("NICs assigned to this group", len(e.Nics.Items))
		}
	}
	return r
}
