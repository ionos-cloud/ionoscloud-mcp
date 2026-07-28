package compute

import (
	"context"
	"fmt"
	"strings"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

// RegisterPccWriteTools registers the create/update/delete private cross connect
// tools. A cross connect is account-level, so these take no datacenter_id.
//
// Which LANs are peered through a cross connect is controlled from the LAN side —
// update_lan's pcc field — not from here, so these tools only manage the cross
// connect's own name and description.
func RegisterPccWriteTools(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	registerCreatePcc(server, client, scope, confirm)
	registerUpdatePcc(server, client, scope)
	registerDeletePcc(server, client, scope, confirm)
}

func registerCreatePcc(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterTool(server, scope, tools.MethodPost, &mcp.Tool{
		Name: "create_pcc",
		Description: "Create one private cross connect, which links private LANs together — including across different data centers — without traffic leaving the IONOS network. Two-phase: call first WITHOUT confirmation_token to get a preview and a one-time token, then call again WITH the token (and the same name) to create it. " +
			"The new cross connect connects nothing on its own: attach LANs to it afterwards with update_lan's pcc field. Every LAN on one cross connect must use non-overlapping IP ranges within the same subnet, so plan the addressing before attaching the second LAN. Creates exactly one cross connect per call.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.CreatePccInput) (*mcp.CallToolResult, any, error) {
		name := strings.TrimSpace(input.Name)
		if name == "" {
			return tools.ErrorText("name is required to create a private cross connect"), nil, nil
		}
		target := tools.Target(name)

		// Phase 2: token present -> validate and execute.
		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "create_pcc", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("create_pcc", "name", err)), nil, nil
			}
			props := &ionos.PrivateCrossConnectProperties{}
			props.SetName(name)
			if input.Description != nil {
				props.SetDescription(*input.Description)
			}
			body := ionos.NewPrivateCrossConnect(*props)
			created, _, err := client.PrivateCrossConnectsApi.PccsPost(ctx).Pcc(*body).Execute()
			return tools.ToResult(created, err)
		}

		// Phase 1: no token -> preview and mint a one-time token.
		token, err := confirm.Mint("create_pcc", target)
		if err != nil {
			return nil, nil, err
		}
		return tools.TextResult(tools.Preview{
			Headline: "About to CREATE one private cross connect:",
			Fields: tools.Fields(
				"name", name,
				"description", tools.OptStr(input.Description),
			),
			Tool:      "create_pcc",
			Replay:    tools.Fields("name", name),
			TokenNote: "It connects nothing until you attach LANs with update_lan's pcc field. The token authorizes creating only this cross connect",
		}.Render(token)), nil, nil
	})
}

func registerUpdatePcc(server *mcp.Server, client *ionos.APIClient, scope tools.Scope) {
	tools.RegisterTool(server, scope, tools.MethodPatch, &mcp.Tool{
		Name: "update_pcc",
		Description: "Update a private cross connect's name or description. Applies a partial update (only the fields you provide). " +
			"This does not change which LANs are peered through it — attach or detach a LAN with update_lan's pcc field instead.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.UpdatePccInput) (*mcp.CallToolResult, any, error) {
		id := strings.TrimSpace(input.PccID)
		if id == "" {
			return tools.ErrorText("pcc_id is required"), nil, nil
		}
		if input.Name == nil && input.Description == nil {
			return tools.ErrorText("nothing to update: provide at least one of name, description"), nil, nil
		}
		// PrivateCrossConnectProperties is all-pointer and its ToMap guards every
		// field, so a zero literal sends only what the caller supplied.
		props := &ionos.PrivateCrossConnectProperties{}
		if input.Name != nil {
			props.SetName(*input.Name)
		}
		if input.Description != nil {
			props.SetDescription(*input.Description)
		}
		updated, _, err := client.PrivateCrossConnectsApi.PccsPatch(ctx, id).Pcc(*props).Execute()
		return tools.ToResult(updated, err)
	})
}

func registerDeletePcc(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterTool(server, scope, tools.MethodDelete, &mcp.Tool{
		Name: "delete_pcc",
		Description: "Delete a private cross connect. Two-phase: call first WITHOUT confirmation_token to get a blast-radius preview listing the LANs peered through it, plus a one-time token, then call again WITH the token to delete. " +
			"The LANs themselves are not deleted, but they lose the private connection between them, which breaks any traffic that relied on it. This is irreversible.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.DeletePccInput) (*mcp.CallToolResult, any, error) {
		id := strings.TrimSpace(input.PccID)
		if id == "" {
			return tools.ErrorText("pcc_id is required"), nil, nil
		}
		target := tools.Target(id)

		// Phase 2: token present -> validate and execute.
		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "delete_pcc", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("delete_pcc", "pcc_id", err)), nil, nil
			}
			_, err := client.PrivateCrossConnectsApi.PccsDelete(ctx, id).Execute()
			if err != nil {
				return tools.ToResult(nil, err)
			}
			return tools.TextResult(tools.DeletedAsync("private cross connect", id)), nil, nil
		}

		// Phase 1: no token -> list the peered LANs, preview, mint a token.
		pcc, _, err := client.PrivateCrossConnectsApi.PccsFindById(ctx, id).Depth(1).Execute()
		if err != nil {
			if tools.IsNotFound(err) {
				return tools.ErrorText(fmt.Sprintf("private cross connect %s does not exist; nothing to delete", id)), nil, nil
			}
			return tools.ToResult(nil, err)
		}
		props := pcc.GetProperties()
		radius := &tools.BlastRadius{}
		// Peers are the LANs currently joined by this cross connect. They survive
		// the delete but stop being able to reach each other privately.
		radius.Add("LANs that lose their private connection", len(props.GetPeers()))
		token, mErr := confirm.Mint("delete_pcc", target)
		if mErr != nil {
			return nil, nil, mErr
		}
		return tools.TextResult(tools.Preview{
			Headline: "About to DELETE a private cross connect. This is IRREVERSIBLE.\n" +
				"The peered LANs are not deleted, but they lose the private connection between them and any traffic relying on it stops working.",
			Fields: tools.Fields(
				"pcc_id", id,
				"name", props.GetName(),
				"description", props.GetDescription(),
				"peered LANs", pccPeerSummary(props.GetPeers()),
			),
			Radius:    radius,
			EmptyNote: "No LANs are peered through this cross connect; deleting it affects nothing else.",
			Tool:      "delete_pcc",
			Replay:    tools.Fields("pcc_id", id),
			TokenNote: "This token authorizes deleting ONLY this cross connect",
		}.Render(token)), nil, nil
	})
}

// pccPeerSummary names the peered LANs and their data centers, so the preview says
// which connections break rather than only how many.
func pccPeerSummary(peers []ionos.Peer) string {
	if len(peers) == 0 {
		return ""
	}
	parts := make([]string, 0, len(peers))
	for _, p := range peers {
		desc := p.GetName()
		if desc == "" {
			desc = p.GetId()
		}
		if dc := p.GetDatacenterName(); dc != "" {
			desc += " (in " + dc + ")"
		}
		parts = append(parts, desc)
	}
	return strings.Join(parts, ", ")
}
