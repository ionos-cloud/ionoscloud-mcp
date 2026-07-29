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

// Write tools for the network load balancer and its forwarding rules. A rule is what
// makes the balancer carry traffic. validateListenerAndTargetLan and validateListener
// are shared with the application load balancer and live here.

// RegisterNetworkLoadBalancerWriteTools registers create/update/delete for the
// network load balancer and for its forwarding rules.
func RegisterNetworkLoadBalancerWriteTools(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	registerNlbCreate(server, client, scope, confirm)
	registerNlbUpdate(server, client, scope)
	registerNlbDelete(server, client, scope, confirm)
	registerNlbForwardingRuleTools(server, client, scope, confirm)
}

// nlbState is the subset of the load balancer's state the write tools need:
// enough for update to carry values forward and for delete to size its blast radius.
type nlbState struct {
	Name        string
	ListenerLan int32
	TargetLan   int32
	Ips         []string
	RuleCount   int
}

// readNlb fetches the current state at depth 2, which is what makes the forwarding
// rule count available without a second call.
func readNlb(ctx context.Context, client *ionos.APIClient, dcID, id string) (nlbState, error) {
	lb, _, err := client.NetworkLoadBalancersApi.DatacentersNetworkloadbalancersFindByNetworkLoadBalancerId(ctx, dcID, id).Depth(2).Execute()
	if err != nil {
		return nlbState{}, err
	}
	p := lb.GetProperties()
	st := nlbState{
		Name: p.GetName(), ListenerLan: p.GetListenerLan(), TargetLan: p.GetTargetLan(),
		Ips: p.GetIps(),
	}
	if e := lb.Entities; e != nil && e.Forwardingrules != nil {
		st.RuleCount = len(e.Forwardingrules.Items)
	}
	return st, nil
}

// buildNlbProperties builds the properties body. name, listenerLan and targetLan are
// serialized unconditionally, so every caller must pass the values it wants kept —
// an update that omitted them would move the balancer off both its networks.
func buildNlbProperties(name string, listenerLan, targetLan int32, f tools.ManagedLoadBalancerFields) *ionos.NetworkLoadBalancerProperties {
	props := &ionos.NetworkLoadBalancerProperties{Name: name, ListenerLan: listenerLan, TargetLan: targetLan}
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

func registerNlbCreate(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterTool(server, scope, tools.MethodPost, &mcp.Tool{
		Name: "create_network_loadbalancer",
		Description: "Create one network load balancer. Two-phase: call first WITHOUT confirmation_token to get a preview and a one-time token, then call again WITH the token (and the same datacenter_id and name) to create it. It forwards traffic at the transport layer (TCP/UDP) to a list of IP targets defined in its forwarding rules. " +
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
			if err := confirm.Consume(*input.ConfirmationToken, "create_network_loadbalancer", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("create_network_loadbalancer", "datacenter_id and name", err)), nil, nil
			}
			props := buildNlbProperties(name, input.ListenerLan, input.TargetLan, input.ManagedLoadBalancerFields)
			body := ionos.NewNetworkLoadBalancer(*props)
			created, _, err := client.NetworkLoadBalancersApi.DatacentersNetworkloadbalancersPost(ctx, dcID).NetworkLoadBalancer(*body).Execute()
			return tools.ToResult(created, err)
		}

		token, err := confirm.Mint("create_network_loadbalancer", target)
		if err != nil {
			return nil, nil, err
		}
		return tools.TextResult(tools.Preview{
			Headline: "About to CREATE one network load balancer:",
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
			Tool:      "create_network_loadbalancer",
			Replay:    tools.Fields("datacenter_id", dcID, "name", name),
			TokenNote: "It carries no traffic until a forwarding rule is added. The token authorizes creating only this load balancer in this data center",
		}.Render(token)), nil, nil
	})
}

func registerNlbUpdate(server *mcp.Server, client *ionos.APIClient, scope tools.Scope) {
	tools.RegisterTool(server, scope, tools.MethodPatch, &mcp.Tool{
		Name: "update_network_loadbalancer",
		Description: "Update a network load balancer's name, listener LAN, target LAN, addresses or logging. Applies a partial update (only the fields you provide). " +
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
		current, err := readNlb(ctx, client, dcID, id)
		if err != nil {
			if tools.IsNotFound(err) {
				return tools.ErrorText(fmt.Sprintf("network load balancer %s does not exist in data center %s; nothing to update", id, dcID)), nil, nil
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

		props := buildNlbProperties(name, listenerLan, targetLan, input.ManagedLoadBalancerFields)
		updated, _, err := client.NetworkLoadBalancersApi.DatacentersNetworkloadbalancersPatch(ctx, dcID, id).NetworkLoadBalancerProperties(*props).Execute()
		return tools.ToResult(updated, err)
	})
}

func registerNlbDelete(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterTool(server, scope, tools.MethodDelete, &mcp.Tool{
		Name: "delete_network_loadbalancer",
		Description: "Delete a network load balancer and all of its forwarding rules. Two-phase: call first WITHOUT confirmation_token to get a blast-radius preview and a one-time token, then call again WITH the token to delete. " +
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
			if err := confirm.Consume(*input.ConfirmationToken, "delete_network_loadbalancer", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("delete_network_loadbalancer", "datacenter_id and loadbalancer_id", err)), nil, nil
			}
			if _, err := client.NetworkLoadBalancersApi.DatacentersNetworkloadbalancersDelete(ctx, dcID, id).Execute(); err != nil {
				return tools.ToResult(nil, err)
			}
			return tools.TextResult(tools.DeletedAsync("network load balancer", id)), nil, nil
		}

		current, err := readNlb(ctx, client, dcID, id)
		if err != nil {
			if tools.IsNotFound(err) {
				return tools.ErrorText(fmt.Sprintf("network load balancer %s does not exist in data center %s; nothing to delete", id, dcID)), nil, nil
			}
			return tools.ToResult(nil, err)
		}
		radius := tools.DestroyedRadius()
		radius.Add("forwarding rules deleted with it", current.RuleCount)
		token, mErr := confirm.Mint("delete_network_loadbalancer", target)
		if mErr != nil {
			return nil, nil, mErr
		}
		return tools.TextResult(tools.Preview{
			Headline: "About to DELETE a network load balancer. This is IRREVERSIBLE.\n" +
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
			Tool:      "delete_network_loadbalancer",
			Replay:    tools.Fields("datacenter_id", dcID, "loadbalancer_id", id),
			TokenNote: "This token authorizes deleting ONLY this network load balancer",
		}.Render(token)), nil, nil
	})
}

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

		// targets is serialized unconditionally, so a rename without this read would
		// empty the backend pool. Read the rule and override only what was supplied.
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

// validateListenerAndTargetLan rejects a non-positive LAN ID, or a balancer whose
// client and backend sides are the same LAN. The API refuses both, vaguely.
func validateListenerAndTargetLan(listenerLan, targetLan int32) string {
	if listenerLan <= 0 {
		return "listener_lan is required and must be a positive LAN ID — the numeric lan value from list_lans, not a UUID"
	}
	if targetLan <= 0 {
		return "target_lan is required and must be a positive LAN ID — the numeric lan value from list_lans, not a UUID"
	}
	if listenerLan == targetLan {
		return fmt.Sprintf("listener_lan and target_lan must be different LANs, both are %d: the listener side is where clients connect (usually a public LAN) and the target side is where the backends live (usually a private LAN)", listenerLan)
	}
	return ""
}

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
