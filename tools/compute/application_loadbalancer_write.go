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

// Write tools for the application load balancer and its forwarding rules. A rule is
// what makes the balancer carry traffic. Shared LAN and listener validation lives in
// network_loadbalancer_write.go.

// RegisterApplicationLoadBalancerWriteTools registers create/update/delete for the
// application load balancer and for its forwarding rules.
func RegisterApplicationLoadBalancerWriteTools(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	registerAlbCreate(server, client, scope, confirm)
	registerAlbUpdate(server, client, scope)
	registerAlbDelete(server, client, scope, confirm)
	registerAlbForwardingRuleTools(server, client, scope, confirm)
}

// albState is the subset of the load balancer's state the write tools need:
// enough for update to carry values forward and for delete to size its blast radius.
type albState struct {
	Name        string
	ListenerLan int32
	TargetLan   int32
	Ips         []string
	RuleCount   int
}

// readAlb fetches the current state at depth 2, which is what makes the forwarding
// rule count available without a second call.
func readAlb(ctx context.Context, client *ionos.APIClient, dcID, id string) (albState, error) {
	lb, _, err := client.ApplicationLoadBalancersApi.DatacentersApplicationloadbalancersFindByApplicationLoadBalancerId(ctx, dcID, id).Depth(2).Execute()
	if err != nil {
		return albState{}, err
	}
	p := lb.GetProperties()
	st := albState{
		Name: p.GetName(), ListenerLan: p.GetListenerLan(), TargetLan: p.GetTargetLan(),
		Ips: p.GetIps(),
	}
	if e := lb.Entities; e != nil && e.Forwardingrules != nil {
		st.RuleCount = len(e.Forwardingrules.Items)
	}
	return st, nil
}

// buildAlbProperties builds the properties body. name, listenerLan and targetLan are
// serialized unconditionally, so every caller must pass the values it wants kept —
// an update that omitted them would move the balancer off both its networks.
func buildAlbProperties(name string, listenerLan, targetLan int32, f tools.ManagedLoadBalancerFields) *ionos.ApplicationLoadBalancerProperties {
	props := &ionos.ApplicationLoadBalancerProperties{Name: name, ListenerLan: listenerLan, TargetLan: targetLan}
	if len(f.Ips) > 0 {
		props.SetIps(f.Ips)
	}
	if len(f.LbPrivateIps) > 0 {
		props.SetLbPrivateIps(f.LbPrivateIps)
	}
	if f.CentralLogging != nil {
		props.SetCentralLogging(*f.CentralLogging)
	}
	if f.LoggingFormat != nil {
		props.SetLoggingFormat(*f.LoggingFormat)
	}
	return props
}

func registerAlbCreate(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterTool(server, scope, tools.MethodPost, &mcp.Tool{
		Name: "create_application_loadbalancer",
		Description: "Create one application load balancer. Two-phase: call first WITHOUT confirmation_token to get a preview and a one-time token, then call again WITH the token (and the same datacenter_id and name) to create it. It forwards traffic at the HTTP layer, and its rules route to target groups (see create_target_group) rather than to raw IP targets. " +
			"listener_lan is the LAN clients connect on (usually public) and target_lan the LAN holding the backends (usually private); they must differ. " +
			"The new load balancer carries no traffic until you add a forwarding rule to it. Creates exactly one per call.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.CreateManagedLoadBalancerInput) (*mcp.CallToolResult, any, error) {
		dcID := strings.TrimSpace(input.DatacenterID)
		name := strings.TrimSpace(input.Name)
		if dcID == "" {
			return tools.ErrorText("datacenter_id is required"), nil, nil
		}
		if name == "" {
			return tools.ErrorText("name is required"), nil, nil
		}
		if msg := validateListenerAndTargetLan(input.ListenerLan, input.TargetLan); msg != "" {
			return tools.ErrorText(msg), nil, nil
		}
		target := tools.Target(dcID, name)

		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "create_application_loadbalancer", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("create_application_loadbalancer", "datacenter_id and name", err)), nil, nil
			}
			props := buildAlbProperties(name, input.ListenerLan, input.TargetLan, input.ManagedLoadBalancerFields)
			body := ionos.NewApplicationLoadBalancer(*props)
			created, _, err := client.ApplicationLoadBalancersApi.DatacentersApplicationloadbalancersPost(ctx, dcID).ApplicationLoadBalancer(*body).Execute()
			return tools.ToResult(created, err)
		}

		token, err := confirm.Mint("create_application_loadbalancer", target)
		if err != nil {
			return nil, nil, err
		}
		return tools.TextResult(tools.Preview{
			Headline: "About to CREATE one application load balancer:",
			Fields: tools.Fields(
				"datacenter_id", dcID,
				"name", name,
				"listener_lan (client side)", strconv.FormatInt(int64(input.ListenerLan), 10),
				"target_lan (backend side)", strconv.FormatInt(int64(input.TargetLan), 10),
				"ips", ipSummary(input.Ips),
				"lb_private_ips", ipSummary(input.LbPrivateIps),
				"central_logging", tools.OptBool(input.CentralLogging),
				"logging_format", tools.OptStr(input.LoggingFormat),
			),
			Tool:      "create_application_loadbalancer",
			Replay:    tools.Fields("datacenter_id", dcID, "name", name),
			TokenNote: "It carries no traffic until a forwarding rule is added. The token authorizes creating only this load balancer in this data center",
		}.Render(token)), nil, nil
	})
}

func registerAlbUpdate(server *mcp.Server, client *ionos.APIClient, scope tools.Scope) {
	tools.RegisterTool(server, scope, tools.MethodPatch, &mcp.Tool{
		Name: "update_application_loadbalancer",
		Description: "Update a application load balancer's name, listener LAN, target LAN, addresses or logging. Applies a partial update (only the fields you provide). " +
			"Omit name, listener_lan or target_lan to keep the current value — each is read and sent back unchanged, because the API always receives all three and empty values would move the load balancer off both of its networks. " +
			"Changing listener_lan moves where clients connect; changing target_lan repoints it at a different backend network. Its forwarding rules are managed separately.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.UpdateManagedLoadBalancerInput) (*mcp.CallToolResult, any, error) {
		dcID := strings.TrimSpace(input.DatacenterID)
		id := strings.TrimSpace(input.LoadBalancerID)
		if dcID == "" {
			return tools.ErrorText("datacenter_id is required"), nil, nil
		}
		if id == "" {
			return tools.ErrorText("loadbalancer_id is required"), nil, nil
		}
		if input.Name == nil && input.ListenerLan == nil && input.TargetLan == nil &&
			len(input.Ips) == 0 && len(input.LbPrivateIps) == 0 &&
			input.CentralLogging == nil && input.LoggingFormat == nil {
			return tools.ErrorText("nothing to update: provide at least one of name, listener_lan, target_lan, ips, lb_private_ips, central_logging, logging_format"), nil, nil
		}

		// name, listenerLan and targetLan are serialized unconditionally, so read
		// the current values and let the caller's input override only what it set.
		current, err := readAlb(ctx, client, dcID, id)
		if err != nil {
			if tools.IsNotFound(err) {
				return tools.ErrorText(fmt.Sprintf("application load balancer %s does not exist in data center %s; nothing to update", id, dcID)), nil, nil
			}
			return tools.ToResult(nil, err)
		}
		name, listenerLan, targetLan := current.Name, current.ListenerLan, current.TargetLan
		if input.Name != nil {
			if strings.TrimSpace(*input.Name) == "" {
				return tools.ErrorText("name must not be empty; omit it entirely to keep the current name"), nil, nil
			}
			name = *input.Name
		}
		if input.ListenerLan != nil {
			listenerLan = *input.ListenerLan
		}
		if input.TargetLan != nil {
			targetLan = *input.TargetLan
		}
		if msg := validateListenerAndTargetLan(listenerLan, targetLan); msg != "" {
			return tools.ErrorText(msg), nil, nil
		}

		props := buildAlbProperties(name, listenerLan, targetLan, input.ManagedLoadBalancerFields)
		updated, _, err := client.ApplicationLoadBalancersApi.DatacentersApplicationloadbalancersPatch(ctx, dcID, id).ApplicationLoadBalancerProperties(*props).Execute()
		return tools.ToResult(updated, err)
	})
}

func registerAlbDelete(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterTool(server, scope, tools.MethodDelete, &mcp.Tool{
		Name: "delete_application_loadbalancer",
		Description: "Delete a application load balancer and all of its forwarding rules. Two-phase: call first WITHOUT confirmation_token to get a blast-radius preview and a one-time token, then call again WITH the token to delete. " +
			"Clients connecting to its listener addresses stop reaching the backends, so this takes the service behind it offline. The backend servers themselves are not deleted. This is irreversible.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.DeleteManagedLoadBalancerInput) (*mcp.CallToolResult, any, error) {
		dcID := strings.TrimSpace(input.DatacenterID)
		id := strings.TrimSpace(input.LoadBalancerID)
		if dcID == "" {
			return tools.ErrorText("datacenter_id is required"), nil, nil
		}
		if id == "" {
			return tools.ErrorText("loadbalancer_id is required"), nil, nil
		}
		target := tools.Target(dcID, id)

		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "delete_application_loadbalancer", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("delete_application_loadbalancer", "datacenter_id and loadbalancer_id", err)), nil, nil
			}
			if _, err := client.ApplicationLoadBalancersApi.DatacentersApplicationloadbalancersDelete(ctx, dcID, id).Execute(); err != nil {
				return tools.ToResult(nil, err)
			}
			return tools.TextResult(tools.DeletedAsync("application load balancer", id)), nil, nil
		}

		current, err := readAlb(ctx, client, dcID, id)
		if err != nil {
			if tools.IsNotFound(err) {
				return tools.ErrorText(fmt.Sprintf("application load balancer %s does not exist in data center %s; nothing to delete", id, dcID)), nil, nil
			}
			return tools.ToResult(nil, err)
		}
		radius := tools.DestroyedRadius()
		radius.Add("forwarding rules deleted with it", current.RuleCount)
		token, mErr := confirm.Mint("delete_application_loadbalancer", target)
		if mErr != nil {
			return nil, nil, mErr
		}
		return tools.TextResult(tools.Preview{
			Headline: "About to DELETE a application load balancer. This is IRREVERSIBLE.\n" +
				"Clients connecting to the listener addresses below stop reaching the backends, so the service behind it goes offline. The backend servers themselves are not deleted.",
			Fields: tools.Fields(
				"datacenter_id", dcID,
				"loadbalancer_id", id,
				"name", current.Name,
				"listener_lan", strconv.FormatInt(int64(current.ListenerLan), 10),
				"target_lan", strconv.FormatInt(int64(current.TargetLan), 10),
				"listener addresses", ipSummary(current.Ips),
			),
			Radius:    radius,
			EmptyNote: "It has no forwarding rules, so it is not currently carrying traffic.",
			Tool:      "delete_application_loadbalancer",
			Replay:    tools.Fields("datacenter_id", dcID, "loadbalancer_id", id),
			TokenNote: "This token authorizes deleting ONLY this application load balancer",
		}.Render(token)), nil, nil
	})
}

func registerAlbForwardingRuleTools(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	api := client.ApplicationLoadBalancersApi

	tools.RegisterTool(server, scope, tools.MethodPost, &mcp.Tool{
		Name: "create_alb_forwarding_rule",
		Description: "Add one forwarding rule to an application load balancer, which is what makes it serve traffic. Two-phase: call first WITHOUT confirmation_token to get a preview and a one-time token, then call again WITH the token (and the same parent IDs and name) to create it. " +
			"The rule listens on one of the load balancer's own addresses and applies its http_rules to matching requests, routing them to target groups (see create_target_group), returning redirects, or returning static responses. " +
			"Without any http_rules the listener accepts connections but routes nothing. For HTTPS you also need server_certificates. Creates exactly one rule per call.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.CreateAlbForwardingRuleInput) (*mcp.CallToolResult, any, error) {
		dcID := strings.TrimSpace(input.DatacenterID)
		lbID := strings.TrimSpace(input.LoadBalancerID)
		name := strings.TrimSpace(input.Name)
		if dcID == "" || lbID == "" {
			return tools.ErrorText("datacenter_id and loadbalancer_id are both required"), nil, nil
		}
		if name == "" {
			return tools.ErrorText("name is required"), nil, nil
		}
		protocol := strings.ToUpper(strings.TrimSpace(input.Protocol))
		if protocol == "" {
			return tools.ErrorText("protocol is required: HTTP or HTTPS"), nil, nil
		}
		if msg := validateListener(input.ListenerIp, input.ListenerPort); msg != "" {
			return tools.ErrorText(msg), nil, nil
		}
		if msg := validateAlbHttpRules(input.HttpRules); msg != "" {
			return tools.ErrorText(msg), nil, nil
		}
		target := tools.Target(dcID, lbID, name)

		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "create_alb_forwarding_rule", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("create_alb_forwarding_rule", "datacenter_id, loadbalancer_id and name", err)), nil, nil
			}
			props := ionos.NewApplicationLoadBalancerForwardingRuleProperties(name, protocol, input.ListenerIp, input.ListenerPort)
			applyAlbOptionalFields(props, input.ClientTimeout, input.ServerCertificates, input.HttpRules)
			body := ionos.NewApplicationLoadBalancerForwardingRule(*props)
			created, _, err := api.DatacentersApplicationloadbalancersForwardingrulesPost(ctx, dcID, lbID).ApplicationLoadBalancerForwardingRule(*body).Execute()
			return tools.ToResult(created, err)
		}

		token, err := confirm.Mint("create_alb_forwarding_rule", target)
		if err != nil {
			return nil, nil, err
		}
		headline := "About to CREATE one forwarding rule on an application load balancer:"
		if protocol == "HTTPS" && len(input.ServerCertificates) == 0 {
			headline += "\nNOTE: protocol is HTTPS but no server_certificates were given, so the listener has nothing to present to clients and the API may reject this."
		}
		if len(input.HttpRules) == 0 {
			headline += "\nNOTE: no http_rules were given, so this listener will accept connections but route nothing."
		}
		return tools.TextResult(tools.Preview{
			Headline: headline,
			Fields: tools.Fields(
				"datacenter_id", dcID,
				"loadbalancer_id", lbID,
				"name", name,
				"listener", fmt.Sprintf("%s:%d", input.ListenerIp, input.ListenerPort),
				"protocol", protocol,
				"client_timeout (ms)", tools.OptInt32(input.ClientTimeout),
				"server_certificates", certSummary(input.ServerCertificates),
				"http_rules", albHttpRuleSummary(input.HttpRules),
			),
			Tool:      "create_alb_forwarding_rule",
			Replay:    tools.Fields("datacenter_id", dcID, "loadbalancer_id", lbID, "name", name),
			TokenNote: "The token authorizes creating only this rule on this load balancer",
		}.Render(token)), nil, nil
	})

	tools.RegisterTool(server, scope, tools.MethodPatch, &mcp.Tool{
		Name: "update_alb_forwarding_rule",
		Description: "Update an application load balancer forwarding rule. Applies a partial update (only the fields you provide). " +
			"Omit name, protocol, listener_ip or listener_port to keep the current value — each is read and sent back unchanged, because the API always receives all four. " +
			"Supplying http_rules or server_certificates REPLACES that list, so include every entry the rule should keep.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.UpdateAlbForwardingRuleInput) (*mcp.CallToolResult, any, error) {
		dcID := strings.TrimSpace(input.DatacenterID)
		lbID := strings.TrimSpace(input.LoadBalancerID)
		id := strings.TrimSpace(input.RuleID)
		if dcID == "" || lbID == "" || id == "" {
			return tools.ErrorText("datacenter_id, loadbalancer_id and rule_id are all required"), nil, nil
		}
		if input.Name == nil && input.Protocol == nil && input.ListenerIp == nil &&
			input.ListenerPort == nil && input.ClientTimeout == nil &&
			input.ServerCertificates == nil && input.HttpRules == nil {
			return tools.ErrorText("nothing to update: provide at least one of name, protocol, listener_ip, listener_port, client_timeout, server_certificates, http_rules"), nil, nil
		}
		if msg := validateAlbHttpRules(input.HttpRules); msg != "" {
			return tools.ErrorText(msg), nil, nil
		}

		current, _, err := api.DatacentersApplicationloadbalancersForwardingrulesFindByForwardingRuleId(ctx, dcID, lbID, id).Depth(1).Execute()
		if err != nil {
			if tools.IsNotFound(err) {
				return tools.ErrorText(fmt.Sprintf("forwarding rule %s does not exist on application load balancer %s; nothing to update", id, lbID)), nil, nil
			}
			return tools.ToResult(nil, err)
		}
		cp := current.GetProperties()
		name, protocol := cp.GetName(), cp.GetProtocol()
		listenerIP, listenerPort := cp.GetListenerIp(), cp.GetListenerPort()

		if input.Name != nil {
			if strings.TrimSpace(*input.Name) == "" {
				return tools.ErrorText("name must not be empty; omit it entirely to keep the current name"), nil, nil
			}
			name = *input.Name
		}
		if input.Protocol != nil {
			protocol = strings.ToUpper(*input.Protocol)
		}
		if input.ListenerIp != nil {
			listenerIP = *input.ListenerIp
		}
		if input.ListenerPort != nil {
			listenerPort = *input.ListenerPort
		}
		if msg := validateListener(listenerIP, listenerPort); msg != "" {
			return tools.ErrorText(msg), nil, nil
		}

		props := ionos.NewApplicationLoadBalancerForwardingRuleProperties(name, protocol, listenerIP, listenerPort)
		// Optional lists are only replaced when supplied; otherwise carry the current
		// values so an unrelated change does not clear the rule's routing.
		certs := input.ServerCertificates
		if certs == nil {
			certs = cp.GetServerCertificates()
		}
		rules := input.HttpRules
		var httpRules []ionos.ApplicationLoadBalancerHttpRule
		if rules != nil {
			httpRules = buildAlbHttpRules(rules)
		} else {
			httpRules = cp.GetHttpRules()
		}
		timeout := input.ClientTimeout
		if timeout == nil && cp.HasClientTimeout() {
			t := cp.GetClientTimeout()
			timeout = &t
		}
		if timeout != nil {
			props.SetClientTimeout(*timeout)
		}
		if len(certs) > 0 {
			props.SetServerCertificates(certs)
		}
		if len(httpRules) > 0 {
			props.SetHttpRules(httpRules)
		}
		updated, _, err := api.DatacentersApplicationloadbalancersForwardingrulesPatch(ctx, dcID, lbID, id).ApplicationLoadBalancerForwardingRuleProperties(*props).Execute()
		return tools.ToResult(updated, err)
	})

	tools.RegisterTool(server, scope, tools.MethodDelete, &mcp.Tool{
		Name: "delete_alb_forwarding_rule",
		Description: "Delete a forwarding rule from an application load balancer. Two-phase: call first WITHOUT confirmation_token to get a preview of the listener and its HTTP rules plus a one-time token, then call again WITH the token to delete. " +
			"Clients connecting to that listener address and port stop being served. If it is the load balancer's only rule, the load balancer stops serving traffic entirely. This is irreversible.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.DeleteAlbForwardingRuleInput) (*mcp.CallToolResult, any, error) {
		dcID := strings.TrimSpace(input.DatacenterID)
		lbID := strings.TrimSpace(input.LoadBalancerID)
		id := strings.TrimSpace(input.RuleID)
		if dcID == "" || lbID == "" || id == "" {
			return tools.ErrorText("datacenter_id, loadbalancer_id and rule_id are all required"), nil, nil
		}
		target := tools.Target(dcID, lbID, id)

		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "delete_alb_forwarding_rule", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("delete_alb_forwarding_rule", "datacenter_id, loadbalancer_id and rule_id", err)), nil, nil
			}
			_, err := api.DatacentersApplicationloadbalancersForwardingrulesDelete(ctx, dcID, lbID, id).Execute()
			if err != nil {
				return tools.ToResult(nil, err)
			}
			return tools.TextResult(tools.DeletedAsync("forwarding rule", id)), nil, nil
		}

		rule, _, err := api.DatacentersApplicationloadbalancersForwardingrulesFindByForwardingRuleId(ctx, dcID, lbID, id).Depth(1).Execute()
		if err != nil {
			if tools.IsNotFound(err) {
				return tools.ErrorText(fmt.Sprintf("forwarding rule %s does not exist on application load balancer %s; nothing to delete", id, lbID)), nil, nil
			}
			return tools.ToResult(nil, err)
		}
		cp := rule.GetProperties()
		radius := tools.DestroyedRadius()
		radius.Add("HTTP rules deleted with it", len(cp.GetHttpRules()))
		token, mErr := confirm.Mint("delete_alb_forwarding_rule", target)
		if mErr != nil {
			return nil, nil, mErr
		}
		return tools.TextResult(tools.Preview{
			Headline: "About to DELETE a forwarding rule from an application load balancer. This is IRREVERSIBLE.\n" +
				"Clients connecting to the listener below stop being served. If this is the load balancer's only rule, it stops serving traffic entirely.",
			Fields: tools.Fields(
				"datacenter_id", dcID,
				"loadbalancer_id", lbID,
				"rule_id", id,
				"name", cp.GetName(),
				"listener", fmt.Sprintf("%s:%d", cp.GetListenerIp(), cp.GetListenerPort()),
				"protocol", cp.GetProtocol(),
			),
			Radius:    radius,
			EmptyNote: "This rule has no HTTP rules, so it is not currently routing anything.",
			Tool:      "delete_alb_forwarding_rule",
			Replay:    tools.Fields("datacenter_id", dcID, "loadbalancer_id", lbID, "rule_id", id),
			TokenNote: "This token authorizes deleting ONLY this rule from ONLY this load balancer",
		}.Render(token)), nil, nil
	})
}

// validateAlbHttpRules checks the field combinations each rule type needs. The type
// decides which fields apply, and the API's rejection does not name the rule.
func validateAlbHttpRules(rules []tools.AlbHttpRuleInput) string {
	for i, r := range rules {
		if strings.TrimSpace(r.Name) == "" {
			return fmt.Sprintf("http_rules[%d].name is required", i)
		}
		switch strings.ToUpper(strings.TrimSpace(r.Type)) {
		case "FORWARD":
			if r.TargetGroup == nil || strings.TrimSpace(*r.TargetGroup) == "" {
				return fmt.Sprintf("http_rules[%d] is type FORWARD, so target_group is required (the ID of the group to route to; see create_target_group)", i)
			}
		case "REDIRECT":
			if r.Location == nil || strings.TrimSpace(*r.Location) == "" {
				return fmt.Sprintf("http_rules[%d] is type REDIRECT, so location is required (the URL to redirect to)", i)
			}
		case "STATIC":
			if r.StatusCode == nil {
				return fmt.Sprintf("http_rules[%d] is type STATIC, so status_code is required (200, 503 or 599)", i)
			}
		case "":
			return fmt.Sprintf("http_rules[%d].type is required: FORWARD, REDIRECT or STATIC", i)
		default:
			return fmt.Sprintf("http_rules[%d].type must be FORWARD, REDIRECT or STATIC, got %q", i, r.Type)
		}
		for j, c := range r.Conditions {
			if strings.TrimSpace(c.Type) == "" || strings.TrimSpace(c.Condition) == "" {
				return fmt.Sprintf("http_rules[%d].conditions[%d] needs both type and condition", i, j)
			}
		}
	}
	return ""
}

func buildAlbHttpRules(in []tools.AlbHttpRuleInput) []ionos.ApplicationLoadBalancerHttpRule {
	out := make([]ionos.ApplicationLoadBalancerHttpRule, 0, len(in))
	for _, r := range in {
		rule := ionos.NewApplicationLoadBalancerHttpRule(r.Name, strings.ToUpper(strings.TrimSpace(r.Type)))
		if r.TargetGroup != nil {
			rule.SetTargetGroup(*r.TargetGroup)
		}
		if r.DropQuery != nil {
			rule.SetDropQuery(*r.DropQuery)
		}
		if r.Location != nil {
			rule.SetLocation(*r.Location)
		}
		if r.StatusCode != nil {
			rule.SetStatusCode(*r.StatusCode)
		}
		if r.ResponseMessage != nil {
			rule.SetResponseMessage(*r.ResponseMessage)
		}
		if r.ContentType != nil {
			rule.SetContentType(*r.ContentType)
		}
		if len(r.Conditions) > 0 {
			conds := make([]ionos.ApplicationLoadBalancerHttpRuleCondition, 0, len(r.Conditions))
			for _, c := range r.Conditions {
				cond := ionos.NewApplicationLoadBalancerHttpRuleCondition(
					strings.ToUpper(strings.TrimSpace(c.Type)), strings.ToUpper(strings.TrimSpace(c.Condition)))
				if c.Negate != nil {
					cond.SetNegate(*c.Negate)
				}
				if c.Key != nil {
					cond.SetKey(*c.Key)
				}
				if c.Value != nil {
					cond.SetValue(*c.Value)
				}
				conds = append(conds, *cond)
			}
			rule.SetConditions(conds)
		}
		out = append(out, *rule)
	}
	return out
}

func applyAlbOptionalFields(props *ionos.ApplicationLoadBalancerForwardingRuleProperties, clientTimeout *int32, certs []string, rules []tools.AlbHttpRuleInput) {
	if clientTimeout != nil {
		props.SetClientTimeout(*clientTimeout)
	}
	if len(certs) > 0 {
		props.SetServerCertificates(certs)
	}
	if len(rules) > 0 {
		props.SetHttpRules(buildAlbHttpRules(rules))
	}
}

// albHttpRuleSummary describes the HTTP rules compactly, since their full form is
// too verbose for a preview but their types and destinations are what matter.
func albHttpRuleSummary(rules []tools.AlbHttpRuleInput) string {
	if len(rules) == 0 {
		return ""
	}
	parts := make([]string, 0, len(rules))
	for _, r := range rules {
		desc := r.Name + " (" + strings.ToUpper(strings.TrimSpace(r.Type))
		switch {
		case r.TargetGroup != nil:
			desc += " -> target group " + *r.TargetGroup
		case r.Location != nil:
			desc += " -> " + *r.Location
		case r.StatusCode != nil:
			desc += " -> " + strconv.FormatInt(int64(*r.StatusCode), 10)
		}
		if len(r.Conditions) == 0 {
			desc += ", matches every request"
		} else {
			desc += fmt.Sprintf(", %d condition(s)", len(r.Conditions))
		}
		parts = append(parts, desc+")")
	}
	return strings.Join(parts, "; ")
}

// certSummary counts certificates rather than listing IDs, which are opaque.
func certSummary(certs []string) string {
	if len(certs) == 0 {
		return ""
	}
	return fmt.Sprintf("%d certificate(s)", len(certs))
}
