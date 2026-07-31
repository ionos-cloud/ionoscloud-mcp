package compute

import (
	"context"
	"fmt"
	"strings"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

// RegisterLanWriteTools registers the create/update/delete LAN tools. create and
// delete are two-phase confirmed; update is a single partial PATCH.
func RegisterLanWriteTools(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	registerCreateLan(server, client, scope, confirm)
	registerUpdateLan(server, client, scope)
	registerDeleteLan(server, client, scope, confirm)
}

func registerCreateLan(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterTool(server, scope, tools.MethodPost, &mcp.Tool{
		Name: "create_lan",
		Description: "Create one LAN in a data center. Two-phase: call first WITHOUT confirmation_token to get a preview and a one-time token, then call again WITH the token and the same datacenter_id to create it. name is optional, but if you gave one the second call must repeat it. " +
			"A public LAN is how servers reach the internet; a private LAN carries internal traffic only. The API assigns the LAN a small numeric ID, which is the value you pass as lan when creating a NIC. Creates exactly one LAN per call.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.CreateLanInput) (*mcp.CallToolResult, any, error) {
		dcID := strings.TrimSpace(input.DatacenterID)
		if dcID == "" {
			return tools.ErrorText("datacenter_id is required to create a LAN"), nil, nil
		}
		// The API allows an unnamed LAN, but one is hard to identify later.
		name := ""
		if input.Name != nil {
			name = strings.TrimSpace(*input.Name)
		}
		targetName := name
		if targetName == "" {
			targetName = "(unnamed)"
		}
		target := tools.Target(dcID, targetName)

		// Phase 2: token present -> validate and execute.
		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "create_lan", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("create_lan", "datacenter_id and the same name, or no name if you gave none", err)), nil, nil
			}
			props := ionos.NewLanPropertiesWithDefaults()
			if name != "" {
				props.SetName(name)
			}
			if input.Public != nil {
				props.SetPublic(*input.Public)
			}
			if input.Pcc != nil {
				props.SetPcc(*input.Pcc)
			}
			if input.Ipv6CidrBlock != nil {
				props.SetIpv6CidrBlock(*input.Ipv6CidrBlock)
			}
			body := ionos.NewLan(*props)
			created, _, err := client.LANsApi.DatacentersLansPost(ctx, dcID).Lan(*body).Execute()
			return tools.ToResult(created, err)
		}

		// Phase 1: no token -> preview and mint a one-time token.
		token, err := confirm.Mint("create_lan", target)
		if err != nil {
			return nil, nil, err
		}
		return tools.TextResult(tools.Preview{
			Headline: "About to CREATE one LAN:",
			Fields: tools.Fields(
				"datacenter_id", dcID,
				"name", name,
				"public", tools.OptBool(input.Public),
				"pcc", tools.OptStr(input.Pcc),
				"ipv6_cidr_block", tools.OptStr(input.Ipv6CidrBlock),
			),
			Tool:      "create_lan",
			Replay:    tools.Fields("datacenter_id", dcID, "name", name),
			TokenNote: "This creates exactly one LAN. The token authorizes creating only this LAN in this data center",
		}.Render(token)), nil, nil
	})
}

func registerUpdateLan(server *mcp.Server, client *ionos.APIClient, scope tools.Scope) {
	tools.RegisterTool(server, scope, tools.MethodPatch, &mcp.Tool{
		Name: "update_lan",
		Description: "Update a LAN's name, public/private setting, cross connect or IPv6 block. Applies a partial update (only the fields you provide). " +
			"Turning a public LAN private removes internet access for every server connected to it, and changing ipv6_cidr_block reassigns the IPv6 blocks and addresses of every connected NIC. ipv4_cidr_block is read-only and cannot be changed.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.UpdateLanInput) (*mcp.CallToolResult, any, error) {
		dcID := strings.TrimSpace(input.DatacenterID)
		id := strings.TrimSpace(input.LanID)
		if dcID == "" {
			return tools.ErrorText("datacenter_id is required"), nil, nil
		}
		if id == "" {
			return tools.ErrorText("lan_id is required"), nil, nil
		}
		if input.Name == nil && input.Public == nil && input.Pcc == nil && input.Ipv6CidrBlock == nil {
			return tools.ErrorText("nothing to update: provide at least one of name, public, pcc, ipv6_cidr_block"), nil, nil
		}
		// A zero-valued literal rather than NewLanPropertiesWithDefaults(): see the
		// "PATCH bodies" note in CLAUDE.md.
		props := &ionos.LanProperties{}
		if input.Name != nil {
			props.SetName(*input.Name)
		}
		if input.Public != nil {
			props.SetPublic(*input.Public)
		}
		if input.Pcc != nil {
			props.SetPcc(*input.Pcc)
		}
		if input.Ipv6CidrBlock != nil {
			props.SetIpv6CidrBlock(*input.Ipv6CidrBlock)
		}
		updated, _, err := client.LANsApi.DatacentersLansPatch(ctx, dcID, id).Lan(*props).Execute()
		return tools.ToResult(updated, err)
	})
}

func registerDeleteLan(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterTool(server, scope, tools.MethodDelete, &mcp.Tool{
		Name: "delete_lan",
		Description: "Delete a LAN. Two-phase: call first WITHOUT confirmation_token to get a preview and a one-time token, then call again WITH the token to delete. " +
			"The LAN must be EMPTY first: the API refuses to delete a LAN that still has NICs on it. Move each NIC to another LAN with update_nic, or remove it with delete_nic (or delete its server), then retry. This is irreversible.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.DeleteLanInput) (*mcp.CallToolResult, any, error) {
		dcID := strings.TrimSpace(input.DatacenterID)
		id := strings.TrimSpace(input.LanID)
		if dcID == "" {
			return tools.ErrorText("datacenter_id is required"), nil, nil
		}
		if id == "" {
			return tools.ErrorText("lan_id is required"), nil, nil
		}
		target := tools.Target(dcID, id)

		// Phase 2: token present -> validate and execute.
		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "delete_lan", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("delete_lan", "datacenter_id and lan_id", err)), nil, nil
			}
			_, err := client.LANsApi.DatacentersLansDelete(ctx, dcID, id).Execute()
			if err != nil {
				return tools.ToResult(nil, err)
			}
			return tools.TextResult(tools.DeletedAsync("LAN", id)), nil, nil
		}

		// Phase 1: no token -> count attached NICs, preview, mint a token.
		lan, _, err := client.LANsApi.DatacentersLansFindById(ctx, dcID, id).Depth(2).Execute()
		if err != nil {
			if tools.IsNotFound(err) {
				return tools.ErrorText(fmt.Sprintf("LAN %s does not exist in data center %s; nothing to delete", id, dcID)), nil, nil
			}
			return tools.ToResult(nil, err)
		}
		props := lan.GetProperties()

		if e := lan.Entities; e != nil && e.Nics != nil && len(e.Nics.Items) > 0 {
			return tools.ErrorText(fmt.Sprintf(
				"LAN %s still has %d NIC(s) on it, and the API refuses to delete a LAN that contains NICs, so this call would be rejected:\n%s\n\n"+
					"Empty the LAN first. Move each NIC to a different LAN with update_nic (set an explicit lan), or remove it with delete_nic, or delete the server it belongs to. Then retry this delete.",
				id, len(e.Nics.Items), lanNicSummary(e.Nics.Items))), nil, nil
		}

		token, mErr := confirm.Mint("delete_lan", target)
		if mErr != nil {
			return nil, nil, mErr
		}
		return tools.TextResult(tools.Preview{
			Headline: "About to DELETE a LAN. This is IRREVERSIBLE.\n" +
				"No NICs are on it, so nothing else is affected.",
			Fields: tools.Fields(
				"datacenter_id", dcID,
				"lan_id", id,
				"name", props.GetName(),
				"public", fmt.Sprintf("%t", props.GetPublic()),
				"ipv4_cidr_block", props.GetIpv4CidrBlock(),
				"pcc", props.GetPcc(),
			),
			Tool:      "delete_lan",
			Replay:    tools.Fields("datacenter_id", dcID, "lan_id", id),
			TokenNote: "This token authorizes deleting ONLY this LAN",
		}.Render(token)), nil, nil
	})
}

// lanNicSummary names the NICs blocking a LAN delete, and the servers they belong to,
// so the caller knows which ones to move or remove.
func lanNicSummary(nics []ionos.Nic) string {
	parts := make([]string, 0, len(nics))
	for _, n := range nics {
		desc := n.GetId()
		props := n.GetProperties()
		if name := props.GetName(); name != "" {
			desc = fmt.Sprintf("%s (%s)", name, desc)
		}
		parts = append(parts, "  - "+desc)
	}
	return strings.Join(parts, "\n")
}
