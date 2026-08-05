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

// RegisterNicWriteTools registers the create/update/delete NIC tools.
func RegisterNicWriteTools(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	registerCreateNic(server, client, scope, confirm)
	registerUpdateNic(server, client, scope)
	registerDeleteNic(server, client, scope, confirm)
}

func registerCreateNic(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterTool(server, scope, tools.MethodPost, &mcp.Tool{
		Name: "create_nic",
		Description: "Create one network interface (NIC) on a server and connect it to a LAN. Two-phase: call first WITHOUT confirmation_token to get a preview and a one-time token, then call again WITH the token (and the same datacenter_id, server_id and lan) to create it. " +
			"lan is the LAN's small numeric ID from list_lans, not a UUID. If no LAN with that ID exists the API creates one implicitly, so a wrong number silently puts the server on a new isolated LAN instead of the intended one — check list_lans first. Creates exactly one NIC per call.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.CreateNicInput) (*mcp.CallToolResult, any, error) {
		dcID := strings.TrimSpace(input.DatacenterID)
		serverID := strings.TrimSpace(input.ServerID)
		if dcID == "" {
			return tools.ErrorText("datacenter_id is required to create a NIC"), nil, nil
		}
		if serverID == "" {
			return tools.ErrorText("server_id is required to create a NIC"), nil, nil
		}
		if input.Lan <= 0 {
			return tools.ErrorText("lan is required and must be a positive LAN ID; list the data center's LANs with list_lans to find it"), nil, nil
		}
		target := tools.Target(req, dcID, serverID, strconv.FormatInt(int64(input.Lan), 10))

		// Phase 2: token present -> validate and execute.
		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "create_nic", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("create_nic", "datacenter_id, server_id and lan", err)), nil, nil
			}
			props := ionos.NewNicProperties(input.Lan)
			if input.Name != nil {
				props.SetName(*input.Name)
			}
			if len(input.Ips) > 0 {
				props.SetIps(input.Ips)
			}
			if input.Dhcp != nil {
				props.SetDhcp(*input.Dhcp)
			}
			if input.FirewallActive != nil {
				props.SetFirewallActive(*input.FirewallActive)
			}
			if input.FirewallType != nil {
				props.SetFirewallType(*input.FirewallType)
			}
			body := ionos.NewNic(*props)
			created, _, err := client.NetworkInterfacesApi.DatacentersServersNicsPost(ctx, dcID, serverID).Nic(*body).Execute()
			return tools.ToResult(created, err)
		}

		// Phase 1: no token -> preview and mint a one-time token.
		token, err := confirm.Mint("create_nic", target)
		if err != nil {
			return nil, nil, err
		}
		return tools.TextResult(tools.Preview{
			Headline: "About to CREATE one network interface (NIC):",
			Fields: tools.Fields(
				"datacenter_id", dcID,
				"server_id", serverID,
				"lan", strconv.FormatInt(int64(input.Lan), 10),
				"name", tools.OptStr(input.Name),
				"ips", ipSummary(input.Ips),
				"dhcp", tools.OptBool(input.Dhcp),
				"firewall_active", tools.OptBool(input.FirewallActive),
				"firewall_type", tools.OptStr(input.FirewallType),
			),
			Tool:      "create_nic",
			Replay:    tools.Fields("datacenter_id", dcID, "server_id", serverID, "lan", strconv.FormatInt(int64(input.Lan), 10)),
			TokenNote: firewallWarning(input.FirewallActive) + "This creates exactly one NIC. The token authorizes creating only this NIC on this server and LAN",
		}.Render(token)), nil, nil
	})
}

func registerUpdateNic(server *mcp.Server, client *ionos.APIClient, scope tools.Scope) {
	tools.RegisterTool(server, scope, tools.MethodPatch, &mcp.Tool{
		Name: "update_nic",
		Description: "Update a NIC's name, LAN, IP addresses or firewall settings. Applies a partial update (only the fields you provide). " +
			"Omit lan to leave the NIC on its current LAN — the current value is read and sent back unchanged, so omitting it never moves the NIC. " +
			"Setting ips REPLACES the whole address list, so include every address the NIC should keep.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.UpdateNicInput) (*mcp.CallToolResult, any, error) {
		dcID := strings.TrimSpace(input.DatacenterID)
		serverID := strings.TrimSpace(input.ServerID)
		id := strings.TrimSpace(input.NicID)
		if dcID == "" {
			return tools.ErrorText("datacenter_id is required"), nil, nil
		}
		if serverID == "" {
			return tools.ErrorText("server_id is required"), nil, nil
		}
		if id == "" {
			return tools.ErrorText("nic_id is required"), nil, nil
		}
		if input.Name == nil && input.Lan == nil && input.Ips == nil &&
			input.Dhcp == nil && input.FirewallActive == nil && input.FirewallType == nil {
			return tools.ErrorText("nothing to update: provide at least one of name, lan, ips, dhcp, firewall_active, firewall_type"), nil, nil
		}

		// The SDK always serializes lan, so an update that omits it would send 0 and
		// move the NIC off its LAN. Read the current value and carry it forward.
		lan := int32(0)
		if input.Lan != nil {
			lan = *input.Lan
		} else {
			current, _, err := client.NetworkInterfacesApi.DatacentersServersNicsFindById(ctx, dcID, serverID, id).Depth(1).Execute()
			if err != nil {
				if tools.IsNotFound(err) {
					return tools.ErrorText(fmt.Sprintf("NIC %s does not exist on server %s; nothing to update", id, serverID)), nil, nil
				}
				return tools.ToResult(nil, err)
			}
			currentProps := current.GetProperties()
			lan = currentProps.GetLan()
		}

		// A literal, not a generated constructor: NewNicProperties injects dhcp=true.
		props := &ionos.NicProperties{Lan: lan}
		if input.Name != nil {
			props.SetName(*input.Name)
		}
		if input.Ips != nil {
			props.SetIps(input.Ips)
		}
		if input.Dhcp != nil {
			props.SetDhcp(*input.Dhcp)
		}
		if input.FirewallActive != nil {
			props.SetFirewallActive(*input.FirewallActive)
		}
		if input.FirewallType != nil {
			props.SetFirewallType(*input.FirewallType)
		}
		updated, _, err := client.NetworkInterfacesApi.DatacentersServersNicsPatch(ctx, dcID, serverID, id).Nic(*props).Execute()
		return tools.ToResult(updated, err)
	})
}

func registerDeleteNic(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterTool(server, scope, tools.MethodDelete, &mcp.Tool{
		Name: "delete_nic",
		Description: "Delete a network interface (NIC) from a server, along with its firewall rules and flow logs. Two-phase: call first WITHOUT confirmation_token to get a blast-radius preview and a one-time token, then call again WITH the token to delete. " +
			"The server loses its connectivity on that LAN, and any public IP the NIC held is released back to its IP block. This is irreversible.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.DeleteNicInput) (*mcp.CallToolResult, any, error) {
		dcID := strings.TrimSpace(input.DatacenterID)
		serverID := strings.TrimSpace(input.ServerID)
		id := strings.TrimSpace(input.NicID)
		if dcID == "" {
			return tools.ErrorText("datacenter_id is required"), nil, nil
		}
		if serverID == "" {
			return tools.ErrorText("server_id is required"), nil, nil
		}
		if id == "" {
			return tools.ErrorText("nic_id is required"), nil, nil
		}
		target := tools.Target(req, dcID, serverID, id)

		// Phase 2: token present -> validate and execute.
		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "delete_nic", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("delete_nic", "datacenter_id, server_id and nic_id", err)), nil, nil
			}
			_, err := client.NetworkInterfacesApi.DatacentersServersNicsDelete(ctx, dcID, serverID, id).Execute()
			if err != nil {
				return tools.ToResult(nil, err)
			}
			return tools.TextResult(tools.DeletedAsync("NIC", id)), nil, nil
		}

		// Phase 1: no token -> compute blast radius, preview, mint a token.
		nic, _, err := client.NetworkInterfacesApi.DatacentersServersNicsFindById(ctx, dcID, serverID, id).Depth(2).Execute()
		if err != nil {
			if tools.IsNotFound(err) {
				return tools.ErrorText(fmt.Sprintf("NIC %s does not exist on server %s; nothing to delete", id, serverID)), nil, nil
			}
			return tools.ToResult(nil, err)
		}
		props := nic.GetProperties()
		token, mErr := confirm.Mint("delete_nic", target)
		if mErr != nil {
			return nil, nil, mErr
		}
		return tools.TextResult(tools.Preview{
			Headline: "About to DELETE a network interface (NIC). This is IRREVERSIBLE.",
			Fields: tools.Fields(
				"datacenter_id", dcID,
				"server_id", serverID,
				"nic_id", id,
				"name", props.GetName(),
				"lan", strconv.FormatInt(int64(props.GetLan()), 10),
				"ips", ipSummary(props.GetIps()),
				"mac", props.GetMac(),
			),
			Radius:    nicBlastRadius(nic),
			EmptyNote: "This NIC has no firewall rules or flow logs; deleting removes only the NIC itself.",
			Tool:      "delete_nic",
			Replay:    tools.Fields("datacenter_id", dcID, "server_id", serverID, "nic_id", id),
			TokenNote: "The server loses connectivity on this LAN. This token authorizes deleting ONLY this NIC",
		}.Render(token)), nil, nil
	})
}

// nicBlastRadius counts what a NIC delete takes with it, from a NIC at depth 2.
func nicBlastRadius(nic ionos.Nic) *tools.BlastRadius {
	r := tools.DestroyedRadius()
	e := nic.Entities
	if e == nil {
		return r
	}
	if e.Firewallrules != nil {
		r.Add("firewall rules", len(e.Firewallrules.Items))
	}
	if e.Flowlogs != nil {
		r.Add("flow logs", len(e.Flowlogs.Items))
	}
	return r
}

// ipSummary renders an IP list for a preview, listing the addresses because they
// are short and the exact values matter when reviewing a network change.
func ipSummary(ips []string) string {
	if len(ips) == 0 {
		return ""
	}
	return strings.Join(ips, ", ")
}

// firewallWarning flags the trap of enabling a firewall with no rules yet, which
// silently blocks all inbound traffic to the server.
func firewallWarning(active *bool) string {
	if active != nil && *active {
		return "NOTE: firewall_active is TRUE and a newly created NIC has no firewall rules, so ALL incoming traffic will be blocked until you add rules with create_firewall_rule. "
	}
	return ""
}
