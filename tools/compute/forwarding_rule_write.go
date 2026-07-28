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

// Forwarding rules are what make a load balancer carry traffic: the balancer itself
// only defines which LANs it sits between.
//
// Both rule models serialize their required fields unconditionally, and the network
// flavour includes `targets` among them. That makes the carry-forward read here more
// than tidiness: a partial update built without it would send an empty targets list
// and remove every backend from the load balancer — an outage caused by renaming a
// rule. Each update therefore reads the current rule and overrides only the fields
// the caller supplied.

// RegisterForwardingRuleWriteTools registers create/update/delete for the network
// and application load balancer forwarding rules.
func RegisterForwardingRuleWriteTools(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	registerNlbForwardingRuleTools(server, client, scope, confirm)
	registerAlbForwardingRuleTools(server, client, scope, confirm)
}

// ---------- network load balancer forwarding rules ----------

func registerNlbForwardingRuleTools(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	api := client.NetworkLoadBalancersApi

	tools.RegisterTool(server, scope, tools.MethodPost, &mcp.Tool{
		Name: "create_nlb_forwarding_rule",
		Description: "Add one forwarding rule to a network load balancer, which is what makes it actually carry traffic. Two-phase: call first WITHOUT confirmation_token to get a preview and a one-time token, then call again WITH the token (and the same parent IDs and name) to create it. " +
			"The rule listens on one of the load balancer's own addresses and forwards TCP or UDP connections to the targets you list; at least one target is required. " +
			"Targets are usually the private IPs of backend servers on the load balancer's target LAN. Creates exactly one rule per call.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.CreateNlbForwardingRuleInput) (*mcp.CallToolResult, any, error) {
		dcID := strings.TrimSpace(input.DatacenterID)
		lbID := strings.TrimSpace(input.LoadBalancerID)
		name := strings.TrimSpace(input.Name)
		if dcID == "" || lbID == "" {
			return tools.ErrorText("datacenter_id and loadbalancer_id are both required"), nil, nil
		}
		if name == "" {
			return tools.ErrorText("name is required"), nil, nil
		}
		if msg := validateListener(input.ListenerIp, input.ListenerPort); msg != "" {
			return tools.ErrorText(msg), nil, nil
		}
		if strings.TrimSpace(input.Algorithm) == "" {
			return tools.ErrorText("algorithm is required: ROUND_ROBIN, LEAST_CONNECTION, RANDOM or SOURCE_IP"), nil, nil
		}
		if strings.TrimSpace(input.Protocol) == "" {
			return tools.ErrorText("protocol is required: TCP or UDP"), nil, nil
		}
		if len(input.Targets) == 0 {
			return tools.ErrorText("targets is required and must contain at least one backend: a rule with no targets accepts connections and has nowhere to send them"), nil, nil
		}
		if msg := validateNlbTargets(input.Targets); msg != "" {
			return tools.ErrorText(msg), nil, nil
		}
		target := tools.Target(dcID, lbID, name)

		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "create_nlb_forwarding_rule", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("create_nlb_forwarding_rule", "datacenter_id, loadbalancer_id and name", err)), nil, nil
			}
			props := ionos.NewNetworkLoadBalancerForwardingRuleProperties(
				name, strings.ToUpper(input.Algorithm), strings.ToUpper(input.Protocol),
				input.ListenerIp, input.ListenerPort, buildNlbTargets(input.Targets))
			if hc := buildNlbHealthCheck(input.HealthCheck); hc != nil {
				props.SetHealthCheck(*hc)
			}
			body := ionos.NewNetworkLoadBalancerForwardingRule(*props)
			created, _, err := api.DatacentersNetworkloadbalancersForwardingrulesPost(ctx, dcID, lbID).NetworkLoadBalancerForwardingRule(*body).Execute()
			return tools.ToResult(created, err)
		}

		token, err := confirm.Mint("create_nlb_forwarding_rule", target)
		if err != nil {
			return nil, nil, err
		}
		return tools.TextResult(tools.Preview{
			Headline: "About to CREATE one forwarding rule on a network load balancer:",
			Fields: append(tools.Fields(
				"datacenter_id", dcID,
				"loadbalancer_id", lbID,
				"name", name,
				"listener", fmt.Sprintf("%s:%d", input.ListenerIp, input.ListenerPort),
				"protocol", strings.ToUpper(input.Protocol),
				"algorithm", strings.ToUpper(input.Algorithm),
			), nlbTargetPreview(input.Targets)...),
			Tool:      "create_nlb_forwarding_rule",
			Replay:    tools.Fields("datacenter_id", dcID, "loadbalancer_id", lbID, "name", name),
			TokenNote: "Clients reaching the listener above start being forwarded to these backends. The token authorizes creating only this rule on this load balancer",
		}.Render(token)), nil, nil
	})

	tools.RegisterTool(server, scope, tools.MethodPatch, &mcp.Tool{
		Name: "update_nlb_forwarding_rule",
		Description: "Update a network load balancer forwarding rule. Applies a partial update (only the fields you provide). " +
			"Omit any of name, algorithm, protocol, listener_ip, listener_port or targets to keep the current value — each is read and sent back unchanged, because the API always receives all of them. " +
			"Supplying targets REPLACES the backend list, so include every backend the rule should keep: any you leave out stops receiving traffic.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.UpdateNlbForwardingRuleInput) (*mcp.CallToolResult, any, error) {
		dcID := strings.TrimSpace(input.DatacenterID)
		lbID := strings.TrimSpace(input.LoadBalancerID)
		id := strings.TrimSpace(input.RuleID)
		if dcID == "" || lbID == "" || id == "" {
			return tools.ErrorText("datacenter_id, loadbalancer_id and rule_id are all required"), nil, nil
		}
		if input.Name == nil && input.Algorithm == nil && input.Protocol == nil &&
			input.ListenerIp == nil && input.ListenerPort == nil &&
			input.Targets == nil && input.HealthCheck == nil {
			return tools.ErrorText("nothing to update: provide at least one of name, algorithm, protocol, listener_ip, listener_port, targets, health_check"), nil, nil
		}
		if input.Targets != nil {
			if len(input.Targets) == 0 {
				return tools.ErrorText("targets must contain at least one backend; omit the field entirely to leave the current backends untouched"), nil, nil
			}
			if msg := validateNlbTargets(input.Targets); msg != "" {
				return tools.ErrorText(msg), nil, nil
			}
		}

		// Every required field — including targets — is serialized unconditionally,
		// so read the rule and override only what the caller supplied. Skipping this
		// would send an empty targets list and drop every backend.
		current, _, err := api.DatacentersNetworkloadbalancersForwardingrulesFindByForwardingRuleId(ctx, dcID, lbID, id).Depth(1).Execute()
		if err != nil {
			if tools.IsNotFound(err) {
				return tools.ErrorText(fmt.Sprintf("forwarding rule %s does not exist on network load balancer %s; nothing to update", id, lbID)), nil, nil
			}
			return tools.ToResult(nil, err)
		}
		cp := current.GetProperties()

		name, algorithm, protocol := cp.GetName(), cp.GetAlgorithm(), cp.GetProtocol()
		listenerIP, listenerPort := cp.GetListenerIp(), cp.GetListenerPort()
		targets := cp.GetTargets()

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
		if input.ListenerIp != nil {
			listenerIP = *input.ListenerIp
		}
		if input.ListenerPort != nil {
			listenerPort = *input.ListenerPort
		}
		if input.Targets != nil {
			targets = buildNlbTargets(input.Targets)
		}
		if msg := validateListener(listenerIP, listenerPort); msg != "" {
			return tools.ErrorText(msg), nil, nil
		}

		props := ionos.NewNetworkLoadBalancerForwardingRuleProperties(name, algorithm, protocol, listenerIP, listenerPort, targets)
		if hc := buildNlbHealthCheck(input.HealthCheck); hc != nil {
			props.SetHealthCheck(*hc)
		} else if cp.HasHealthCheck() {
			// The health check is optional and guarded, but carrying it forward keeps
			// an unrelated update from silently resetting the rule's timeouts.
			props.SetHealthCheck(cp.GetHealthCheck())
		}
		updated, _, err := api.DatacentersNetworkloadbalancersForwardingrulesPatch(ctx, dcID, lbID, id).NetworkLoadBalancerForwardingRuleProperties(*props).Execute()
		return tools.ToResult(updated, err)
	})

	tools.RegisterTool(server, scope, tools.MethodDelete, &mcp.Tool{
		Name: "delete_nlb_forwarding_rule",
		Description: "Delete a forwarding rule from a network load balancer. Two-phase: call first WITHOUT confirmation_token to get a preview of the listener and its backends plus a one-time token, then call again WITH the token to delete. " +
			"Clients connecting to that listener address and port stop being served. If it is the load balancer's only rule, the load balancer stops carrying traffic entirely. This is irreversible.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.DeleteNlbForwardingRuleInput) (*mcp.CallToolResult, any, error) {
		dcID := strings.TrimSpace(input.DatacenterID)
		lbID := strings.TrimSpace(input.LoadBalancerID)
		id := strings.TrimSpace(input.RuleID)
		if dcID == "" || lbID == "" || id == "" {
			return tools.ErrorText("datacenter_id, loadbalancer_id and rule_id are all required"), nil, nil
		}
		target := tools.Target(dcID, lbID, id)

		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "delete_nlb_forwarding_rule", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("delete_nlb_forwarding_rule", "datacenter_id, loadbalancer_id and rule_id", err)), nil, nil
			}
			_, err := api.DatacentersNetworkloadbalancersForwardingrulesDelete(ctx, dcID, lbID, id).Execute()
			if err != nil {
				return tools.ToResult(nil, err)
			}
			return tools.TextResult(tools.DeletedAsync("forwarding rule", id)), nil, nil
		}

		rule, _, err := api.DatacentersNetworkloadbalancersForwardingrulesFindByForwardingRuleId(ctx, dcID, lbID, id).Depth(1).Execute()
		if err != nil {
			if tools.IsNotFound(err) {
				return tools.ErrorText(fmt.Sprintf("forwarding rule %s does not exist on network load balancer %s; nothing to delete", id, lbID)), nil, nil
			}
			return tools.ToResult(nil, err)
		}
		cp := rule.GetProperties()
		radius := tools.AffectedRadius()
		radius.Add("backends that stop receiving traffic through this rule", len(cp.GetTargets()))
		token, mErr := confirm.Mint("delete_nlb_forwarding_rule", target)
		if mErr != nil {
			return nil, nil, mErr
		}
		return tools.TextResult(tools.Preview{
			Headline: "About to DELETE a forwarding rule from a network load balancer. This is IRREVERSIBLE.\n" +
				"Clients connecting to the listener below stop being served. If this is the load balancer's only rule, it stops carrying traffic entirely.",
			Fields: tools.Fields(
				"datacenter_id", dcID,
				"loadbalancer_id", lbID,
				"rule_id", id,
				"name", cp.GetName(),
				"listener", fmt.Sprintf("%s:%d", cp.GetListenerIp(), cp.GetListenerPort()),
				"protocol", cp.GetProtocol(),
				"algorithm", cp.GetAlgorithm(),
			),
			Radius:    radius,
			EmptyNote: "This rule has no backends, so it is not currently serving anything.",
			Tool:      "delete_nlb_forwarding_rule",
			Replay:    tools.Fields("datacenter_id", dcID, "loadbalancer_id", lbID, "rule_id", id),
			TokenNote: "This token authorizes deleting ONLY this rule from ONLY this load balancer",
		}.Render(token)), nil, nil
	})
}

// ---------- application load balancer forwarding rules ----------

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

// ---------- shared helpers ----------

// validateListener catches an unusable listener before the request. The API rejects
// both cases but without naming the field.
func validateListener(ip string, port int32) string {
	if strings.TrimSpace(ip) == "" {
		return "listener_ip is required: it must be one of the load balancer's own listener addresses (see get_network_loadbalancer or get_application_loadbalancer)"
	}
	if port < 1 || port > 65535 {
		return fmt.Sprintf("listener_port must be between 1 and 65535, got %d", port)
	}
	return ""
}

// validateNlbTargets checks each backend, since a malformed one is rejected by the
// API as a whole-request failure that does not say which entry was wrong.
func validateNlbTargets(targets []tools.NlbTargetInput) string {
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

func buildNlbTargets(in []tools.NlbTargetInput) []ionos.NetworkLoadBalancerForwardingRuleTarget {
	out := make([]ionos.NetworkLoadBalancerForwardingRuleTarget, 0, len(in))
	for _, t := range in {
		target := ionos.NewNetworkLoadBalancerForwardingRuleTarget(t.Ip, t.Port, t.Weight)
		if t.ProxyProtocol != nil {
			target.SetProxyProtocol(*t.ProxyProtocol)
		}
		if t.HealthCheck != nil || t.HealthCheckInterval != nil || t.Maintenance != nil {
			hc := ionos.NewNetworkLoadBalancerForwardingRuleTargetHealthCheckWithDefaults()
			if t.HealthCheck != nil {
				hc.SetCheck(*t.HealthCheck)
			}
			if t.HealthCheckInterval != nil {
				hc.SetCheckInterval(*t.HealthCheckInterval)
			}
			if t.Maintenance != nil {
				hc.SetMaintenance(*t.Maintenance)
			}
			target.SetHealthCheck(*hc)
		}
		out = append(out, *target)
	}
	return out
}

func buildNlbHealthCheck(in *tools.NlbHealthCheckInput) *ionos.NetworkLoadBalancerForwardingRuleHealthCheck {
	if in == nil {
		return nil
	}
	hc := ionos.NewNetworkLoadBalancerForwardingRuleHealthCheckWithDefaults()
	if in.ClientTimeout != nil {
		hc.SetClientTimeout(*in.ClientTimeout)
	}
	if in.ConnectTimeout != nil {
		hc.SetConnectTimeout(*in.ConnectTimeout)
	}
	if in.TargetTimeout != nil {
		hc.SetTargetTimeout(*in.TargetTimeout)
	}
	if in.Retries != nil {
		hc.SetRetries(*in.Retries)
	}
	return hc
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

// nlbTargetPreview lists the backends a rule will forward to, one line each, so the
// caller can see the actual destinations rather than a count.
func nlbTargetPreview(targets []tools.NlbTargetInput) []tools.KV {
	out := make([]tools.KV, 0, len(targets))
	for i, t := range targets {
		desc := fmt.Sprintf("%s:%d weight %d", t.Ip, t.Port, t.Weight)
		if t.Maintenance != nil && *t.Maintenance {
			desc += " (in maintenance, receives no traffic)"
		}
		out = append(out, tools.Fields(fmt.Sprintf("target[%d]", i), desc)...)
	}
	return out
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
