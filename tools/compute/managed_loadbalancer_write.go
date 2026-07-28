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

// Network and application load balancers have field-for-field identical property
// models, so the two tool sets differ only in which client methods they call. That
// difference is captured in managedLbAPI, letting one implementation serve both.
//
// Both models serialize name, listenerLan and targetLan unconditionally, so every
// update reads the current resource and carries those values forward. Omitting them
// would send an empty name and listener/target LAN 0, moving the load balancer off
// both of its networks as a side effect of an unrelated change.

// managedLbAPI adapts one load balancer flavour to the shared implementation.
type managedLbAPI struct {
	kind     string // "network" or "application", used in tool names and prose
	toolName string // e.g. "network_loadbalancer"
	label    string // human label for previews, e.g. "network load balancer"
	// docOp names the API operation family for the description's reference.
	create func(ctx context.Context, client *ionos.APIClient, dcID string, name string, listenerLan, targetLan int32, f tools.ManagedLoadBalancerFields) (any, error)
	// read returns the current name, listenerLan and targetLan plus the rule count,
	// so update can carry values forward and delete can size its blast radius.
	read   func(ctx context.Context, client *ionos.APIClient, dcID, id string) (managedLbState, error)
	update func(ctx context.Context, client *ionos.APIClient, dcID, id string, name string, listenerLan, targetLan int32, f tools.ManagedLoadBalancerFields) (any, error)
	del    func(ctx context.Context, client *ionos.APIClient, dcID, id string) error
}

// managedLbState is the subset of a load balancer's state the write tools need.
type managedLbState struct {
	Name        string
	ListenerLan int32
	TargetLan   int32
	Ips         []string
	RuleCount   int
	Found       bool
}

// RegisterManagedLoadBalancerWriteTools registers create/update/delete for both the
// network and the application load balancer.
func RegisterManagedLoadBalancerWriteTools(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	for _, api := range []managedLbAPI{networkLbAPI(), applicationLbAPI()} {
		registerManagedLbCreate(server, client, scope, confirm, api)
		registerManagedLbUpdate(server, client, scope, api)
		registerManagedLbDelete(server, client, scope, confirm, api)
	}
}

func networkLbAPI() managedLbAPI {
	return managedLbAPI{
		kind: "network", toolName: "network_loadbalancer", label: "network load balancer",
		create: func(ctx context.Context, c *ionos.APIClient, dcID, name string, listenerLan, targetLan int32, f tools.ManagedLoadBalancerFields) (any, error) {
			props := ionos.NewNetworkLoadBalancerProperties(name, listenerLan, targetLan)
			applyManagedLbFields(props.SetIps, props.SetLbPrivateIps, props.SetCentralLogging, props.SetLoggingFormat, f)
			body := ionos.NewNetworkLoadBalancer(*props)
			created, _, err := c.NetworkLoadBalancersApi.DatacentersNetworkloadbalancersPost(ctx, dcID).NetworkLoadBalancer(*body).Execute()
			return created, err
		},
		read: func(ctx context.Context, c *ionos.APIClient, dcID, id string) (managedLbState, error) {
			lb, _, err := c.NetworkLoadBalancersApi.DatacentersNetworkloadbalancersFindByNetworkLoadBalancerId(ctx, dcID, id).Depth(2).Execute()
			if err != nil {
				return managedLbState{}, err
			}
			p := lb.GetProperties()
			st := managedLbState{
				Name: p.GetName(), ListenerLan: p.GetListenerLan(), TargetLan: p.GetTargetLan(),
				Ips: p.GetIps(), Found: true,
			}
			if e := lb.Entities; e != nil && e.Forwardingrules != nil {
				st.RuleCount = len(e.Forwardingrules.Items)
			}
			return st, nil
		},
		update: func(ctx context.Context, c *ionos.APIClient, dcID, id, name string, listenerLan, targetLan int32, f tools.ManagedLoadBalancerFields) (any, error) {
			props := ionos.NewNetworkLoadBalancerProperties(name, listenerLan, targetLan)
			applyManagedLbFields(props.SetIps, props.SetLbPrivateIps, props.SetCentralLogging, props.SetLoggingFormat, f)
			updated, _, err := c.NetworkLoadBalancersApi.DatacentersNetworkloadbalancersPatch(ctx, dcID, id).NetworkLoadBalancerProperties(*props).Execute()
			return updated, err
		},
		del: func(ctx context.Context, c *ionos.APIClient, dcID, id string) error {
			_, err := c.NetworkLoadBalancersApi.DatacentersNetworkloadbalancersDelete(ctx, dcID, id).Execute()
			return err
		},
	}
}

func applicationLbAPI() managedLbAPI {
	return managedLbAPI{
		kind: "application", toolName: "application_loadbalancer", label: "application load balancer",
		create: func(ctx context.Context, c *ionos.APIClient, dcID, name string, listenerLan, targetLan int32, f tools.ManagedLoadBalancerFields) (any, error) {
			props := ionos.NewApplicationLoadBalancerProperties(name, listenerLan, targetLan)
			applyManagedLbFields(props.SetIps, props.SetLbPrivateIps, props.SetCentralLogging, props.SetLoggingFormat, f)
			body := ionos.NewApplicationLoadBalancer(*props)
			created, _, err := c.ApplicationLoadBalancersApi.DatacentersApplicationloadbalancersPost(ctx, dcID).ApplicationLoadBalancer(*body).Execute()
			return created, err
		},
		read: func(ctx context.Context, c *ionos.APIClient, dcID, id string) (managedLbState, error) {
			lb, _, err := c.ApplicationLoadBalancersApi.DatacentersApplicationloadbalancersFindByApplicationLoadBalancerId(ctx, dcID, id).Depth(2).Execute()
			if err != nil {
				return managedLbState{}, err
			}
			p := lb.GetProperties()
			st := managedLbState{
				Name: p.GetName(), ListenerLan: p.GetListenerLan(), TargetLan: p.GetTargetLan(),
				Ips: p.GetIps(), Found: true,
			}
			if e := lb.Entities; e != nil && e.Forwardingrules != nil {
				st.RuleCount = len(e.Forwardingrules.Items)
			}
			return st, nil
		},
		update: func(ctx context.Context, c *ionos.APIClient, dcID, id, name string, listenerLan, targetLan int32, f tools.ManagedLoadBalancerFields) (any, error) {
			props := ionos.NewApplicationLoadBalancerProperties(name, listenerLan, targetLan)
			applyManagedLbFields(props.SetIps, props.SetLbPrivateIps, props.SetCentralLogging, props.SetLoggingFormat, f)
			updated, _, err := c.ApplicationLoadBalancersApi.DatacentersApplicationloadbalancersPatch(ctx, dcID, id).ApplicationLoadBalancerProperties(*props).Execute()
			return updated, err
		},
		del: func(ctx context.Context, c *ionos.APIClient, dcID, id string) error {
			_, err := c.ApplicationLoadBalancersApi.DatacentersApplicationloadbalancersDelete(ctx, dcID, id).Execute()
			return err
		},
	}
}

// applyManagedLbFields sets the optional shared fields through the flavour's own
// setters, which is the only way to share this code across two generated types that
// have identical fields but no common interface.
func applyManagedLbFields(setIps func([]string), setPrivateIps func([]string), setCentralLogging func(bool), setLoggingFormat func(string), f tools.ManagedLoadBalancerFields) {
	if len(f.Ips) > 0 {
		setIps(f.Ips)
	}
	if len(f.LbPrivateIps) > 0 {
		setPrivateIps(f.LbPrivateIps)
	}
	if f.CentralLogging != nil {
		setCentralLogging(*f.CentralLogging)
	}
	if f.LoggingFormat != nil {
		setLoggingFormat(*f.LoggingFormat)
	}
}

func registerManagedLbCreate(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore, api managedLbAPI) {
	extra := "It forwards traffic at the transport layer (TCP/UDP) to a list of IP targets defined in its forwarding rules."
	if api.kind == "application" {
		extra = "It forwards traffic at the HTTP layer, and its rules route to target groups (see create_target_group) rather than to raw IP targets."
	}
	tools.RegisterTool(server, scope, tools.MethodPost, &mcp.Tool{
		Name: "create_" + api.toolName,
		Description: fmt.Sprintf("Create one %s. Two-phase: call first WITHOUT confirmation_token to get a preview and a one-time token, then call again WITH the token (and the same datacenter_id and name) to create it. %s ",
			api.label, extra) +
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
			if err := confirm.Consume(*input.ConfirmationToken, "create_"+api.toolName, target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("create_"+api.toolName, "datacenter_id and name", err)), nil, nil
			}
			created, err := api.create(ctx, client, dcID, name, input.ListenerLan, input.TargetLan, input.ManagedLoadBalancerFields)
			return tools.ToResult(created, err)
		}

		token, err := confirm.Mint("create_"+api.toolName, target)
		if err != nil {
			return nil, nil, err
		}
		return tools.TextResult(tools.Preview{
			Headline: fmt.Sprintf("About to CREATE one %s:", api.label),
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
			Tool:      "create_" + api.toolName,
			Replay:    tools.Fields("datacenter_id", dcID, "name", name),
			TokenNote: "It carries no traffic until a forwarding rule is added. The token authorizes creating only this load balancer in this data center",
		}.Render(token)), nil, nil
	})
}

func registerManagedLbUpdate(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, api managedLbAPI) {
	tools.RegisterTool(server, scope, tools.MethodPatch, &mcp.Tool{
		Name: "update_" + api.toolName,
		Description: fmt.Sprintf("Update a %s's name, listener LAN, target LAN, addresses or logging. Applies a partial update (only the fields you provide). ", api.label) +
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
		current, err := api.read(ctx, client, dcID, id)
		if err != nil {
			if tools.IsNotFound(err) {
				return tools.ErrorText(fmt.Sprintf("%s %s does not exist in data center %s; nothing to update", api.label, id, dcID)), nil, nil
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

		updated, err := api.update(ctx, client, dcID, id, name, listenerLan, targetLan, input.ManagedLoadBalancerFields)
		return tools.ToResult(updated, err)
	})
}

func registerManagedLbDelete(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore, api managedLbAPI) {
	tools.RegisterTool(server, scope, tools.MethodDelete, &mcp.Tool{
		Name: "delete_" + api.toolName,
		Description: fmt.Sprintf("Delete a %s and all of its forwarding rules. Two-phase: call first WITHOUT confirmation_token to get a blast-radius preview and a one-time token, then call again WITH the token to delete. ", api.label) +
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
			if err := confirm.Consume(*input.ConfirmationToken, "delete_"+api.toolName, target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("delete_"+api.toolName, "datacenter_id and loadbalancer_id", err)), nil, nil
			}
			if err := api.del(ctx, client, dcID, id); err != nil {
				return tools.ToResult(nil, err)
			}
			return tools.TextResult(tools.DeletedAsync(api.label, id)), nil, nil
		}

		current, err := api.read(ctx, client, dcID, id)
		if err != nil {
			if tools.IsNotFound(err) {
				return tools.ErrorText(fmt.Sprintf("%s %s does not exist in data center %s; nothing to delete", api.label, id, dcID)), nil, nil
			}
			return tools.ToResult(nil, err)
		}
		radius := &tools.BlastRadius{}
		radius.Add("forwarding rules deleted with it", current.RuleCount)
		token, mErr := confirm.Mint("delete_"+api.toolName, target)
		if mErr != nil {
			return nil, nil, mErr
		}
		return tools.TextResult(tools.Preview{
			Headline: fmt.Sprintf("About to DELETE a %s. This is IRREVERSIBLE.\n", api.label) +
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
			Tool:      "delete_" + api.toolName,
			Replay:    tools.Fields("datacenter_id", dcID, "loadbalancer_id", id),
			TokenNote: fmt.Sprintf("This token authorizes deleting ONLY this %s", api.label),
		}.Render(token)), nil, nil
	})
}

// validateListenerAndTargetLan catches the configuration that cannot work: a load
// balancer whose client side and backend side are the same LAN, or a non-positive
// LAN ID. The API rejects both, but not in terms that name the field.
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
