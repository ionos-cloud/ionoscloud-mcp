package dns

import (
	"context"
	"fmt"
	"strings"

	dnsSDK "github.com/ionos-cloud/sdk-go-bundle/products/dns/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

// notifyNote names the addresses IONOS sends DNS notify messages from. A secondary
// zone silently fails to transfer if the primary does not accept them.
const notifyNote = " Whitelist IONOS's notify sources on your primary nameservers, or the transfer will not work: IPv4 212.227.123.25 and IPv6 2001:8d8:fe:53::5cd:25."

// RegisterSecondaryZoneWriteTools registers the create/update/delete secondary zone
// tools and the zone-transfer trigger.
func RegisterSecondaryZoneWriteTools(server *mcp.Server, client *dnsSDK.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	registerCreateSecondaryZone(server, client, scope, confirm)
	registerUpdateSecondaryZone(server, client, scope)
	registerDeleteSecondaryZone(server, client, scope, confirm)
	registerStartZoneTransfer(server, client, scope)
}

func registerCreateSecondaryZone(server *mcp.Server, client *dnsSDK.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterTool(server, scope, tools.MethodPost, &mcp.Tool{
		Name: "create_dns_secondary_zone",
		Description: "Create one secondary DNS zone, which mirrors a zone hosted on nameservers you run elsewhere. Two-phase: call first WITHOUT confirmation_token to get a preview and a one-time token, then call again WITH the token (and the same zone_name) to create it. Creates exactly one zone per call. " +
			"Its records are transferred from the primary IPs, not authored here — use start_dns_zone_transfer to trigger a transfer and get_dns_secondary_zone_axfr to check it." + notifyNote + asyncZoneNote,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.CreateDnsSecondaryZoneInput) (*mcp.CallToolResult, any, error) {
		name := strings.TrimSpace(input.ZoneName)
		if name == "" {
			return tools.ErrorText("zone_name is required to create a secondary DNS zone (e.g. example.com)"), nil, nil
		}
		ips, msg := validatePrimaryIps(input.PrimaryIps)
		if msg != "" {
			return tools.ErrorText(msg), nil, nil
		}
		target := tools.Target(req, name)

		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "create_dns_secondary_zone", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("create_dns_secondary_zone", "zone_name and primary_ips", err)), nil, nil
			}
			props := dnsSDK.SecondaryZone{ZoneName: name, PrimaryIps: ips}
			if input.Description != nil {
				props.Description = input.Description
			}
			created, _, err := client.SecondaryZonesApi.SecondaryzonesPost(ctx).SecondaryZoneCreate(dnsSDK.SecondaryZoneCreate{Properties: props}).Execute()
			return tools.ToResult(created, err)
		}

		token, err := confirm.Mint("create_dns_secondary_zone", target)
		if err != nil {
			return nil, nil, err
		}
		return tools.TextResult(tools.Preview{
			Headline: "About to CREATE one secondary DNS zone:",
			Fields: tools.Fields(
				"zone_name", name,
				"primary_ips", strings.Join(ips, ", "),
				"description", tools.OptStr(input.Description),
			),
			Tool:      "create_dns_secondary_zone",
			Replay:    tools.Fields("zone_name", name, "primary_ips", strings.Join(ips, ", ")),
			TokenNote: "This creates exactly one zone. The token authorizes creating only this zone_name",
		}.Render(token)), nil, nil
	})
}

func registerUpdateSecondaryZone(server *mcp.Server, client *dnsSDK.APIClient, scope tools.Scope) {
	tools.RegisterTool(server, scope, tools.MethodPut, &mcp.Tool{
		Name: "update_dns_secondary_zone",
		Description: "Update a secondary DNS zone's primary nameserver IPs or its description. The zone name is immutable and is read and sent back unchanged. " +
			"primary_ips REPLACES the current list when supplied, so include every IP that should remain." + notifyNote + asyncZoneNote,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.UpdateDnsSecondaryZoneInput) (*mcp.CallToolResult, any, error) {
		id := strings.TrimSpace(input.SecondaryZoneID)
		if id == "" {
			return tools.ErrorText("secondary_zone_id is required"), nil, nil
		}
		if input.PrimaryIps == nil && input.Description == nil {
			return tools.ErrorText("nothing to update: provide at least one of primary_ips, description"), nil, nil
		}
		var ips []string
		if input.PrimaryIps != nil {
			var msg string
			if ips, msg = validatePrimaryIps(input.PrimaryIps); msg != "" {
				return tools.ErrorText(msg), nil, nil
			}
		}

		// zoneName and primaryIps are both serialized unconditionally, and a nil
		// primaryIps would go out as null, which the API rejects. Read the zone and
		// override only what the caller supplied.
		current, _, err := client.SecondaryZonesApi.SecondaryzonesFindById(ctx, id).Execute()
		if err != nil {
			if tools.IsNotFound(err) {
				return tools.ErrorText(fmt.Sprintf("secondary DNS zone %s does not exist; nothing to update", id)), nil, nil
			}
			return tools.ToResult(nil, err)
		}
		cp := current.GetProperties()

		props := dnsSDK.SecondaryZone{ZoneName: cp.ZoneName, PrimaryIps: cp.PrimaryIps}
		if ips != nil {
			props.PrimaryIps = ips
		}
		switch {
		case input.Description != nil:
			props.Description = input.Description
		case cp.Description != nil:
			props.Description = cp.Description
		}

		updated, _, err := client.SecondaryZonesApi.SecondaryzonesPut(ctx, id).SecondaryZoneEnsure(dnsSDK.SecondaryZoneEnsure{Properties: props}).Execute()
		return tools.ToResult(updated, err)
	})
}

func registerDeleteSecondaryZone(server *mcp.Server, client *dnsSDK.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterTool(server, scope, tools.MethodDelete, &mcp.Tool{
		Name: "delete_dns_secondary_zone",
		Description: "Delete a secondary DNS zone. Two-phase: call first WITHOUT confirmation_token to get a preview and a one-time token, then call again WITH the token to delete. This is irreversible. " +
			"The zone's records live on your primary nameservers, so only the IONOS-side mirror is removed." + asyncZoneNote,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.DeleteDnsSecondaryZoneInput) (*mcp.CallToolResult, any, error) {
		id := strings.TrimSpace(input.SecondaryZoneID)
		if id == "" {
			return tools.ErrorText("secondary_zone_id is required"), nil, nil
		}
		target := tools.Target(req, id)

		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "delete_dns_secondary_zone", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("delete_dns_secondary_zone", "secondary_zone_id", err)), nil, nil
			}
			if _, _, err := client.SecondaryZonesApi.SecondaryzonesDelete(ctx, id).Execute(); err != nil {
				return tools.ToResult(nil, err)
			}
			return tools.TextResult(tools.DeletedAsync("secondary DNS zone", id)), nil, nil
		}

		zone, _, err := client.SecondaryZonesApi.SecondaryzonesFindById(ctx, id).Execute()
		if err != nil {
			if tools.IsNotFound(err) {
				return tools.ErrorText(fmt.Sprintf("secondary DNS zone %s does not exist; nothing to delete", id)), nil, nil
			}
			return tools.ToResult(nil, err)
		}
		token, mErr := confirm.Mint("delete_dns_secondary_zone", target)
		if mErr != nil {
			return nil, nil, mErr
		}
		cp := zone.GetProperties()
		return tools.TextResult(tools.Preview{
			Headline: "About to DELETE a secondary DNS zone. This is IRREVERSIBLE.",
			Fields: tools.Fields(
				"secondary_zone_id", id,
				"zone_name", cp.ZoneName,
				"primary_ips", strings.Join(cp.PrimaryIps, ", "),
				"state", string(zone.Metadata.State),
			),
			Tool:      "delete_dns_secondary_zone",
			Replay:    tools.Fields("secondary_zone_id", id),
			TokenNote: "This token authorizes deleting ONLY this zone",
		}.Render(token)), nil, nil
	})
}

func registerStartZoneTransfer(server *mcp.Server, client *dnsSDK.APIClient, scope tools.Scope) {
	tools.RegisterActionTool(server, scope,
		tools.Action{Verb: "start_", Method: tools.MethodPut, Idempotent: true},
		&mcp.Tool{
			Name: "start_dns_zone_transfer",
			Description: "Trigger a zone transfer (AXFR) for a secondary DNS zone, pulling its records from the primary nameservers. Requires IONOS_MCP_TOOL_SCOPE to include write. Single call — a transfer only refreshes the IONOS-side copy, so there is no confirmation step." +
				notifyNote + asyncTransferNote,
		}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.StartDnsZoneTransferInput) (*mcp.CallToolResult, any, error) {
			id := strings.TrimSpace(input.SecondaryZoneID)
			if id == "" {
				return tools.ErrorText("secondary_zone_id is required"), nil, nil
			}
			if _, _, err := client.SecondaryZonesApi.SecondaryzonesAxfrPut(ctx, id).Execute(); err != nil {
				return tools.ToResult(nil, err)
			}
			// The endpoint returns an empty body, so report acceptance in text.
			return tools.TextResult(fmt.Sprintf("Requested a zone transfer for secondary zone %s. Check get_dns_secondary_zone_axfr for the per-primary-IP status.", id)), nil, nil
		})
}
