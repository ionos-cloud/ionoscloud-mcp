package dns

import (
	"context"
	"fmt"
	"strings"

	dnsSDK "github.com/ionos-cloud/sdk-go-bundle/products/dns/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

// RegisterDNSSECWriteTools registers the DNSSEC enable/disable tools.
func RegisterDNSSECWriteTools(server *mcp.Server, client *dnsSDK.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	registerCreateDnssecKey(server, client, scope, confirm)
	registerDeleteDnssecKey(server, client, scope, confirm)
}

func registerCreateDnssecKey(server *mcp.Server, client *dnsSDK.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterTool(server, scope, tools.MethodPost, &mcp.Tool{
		Name: "create_dns_zone_dnssec_key",
		Description: "Enable DNSSEC on a primary DNS zone, generating its signing keys and DNSKEY records. Two-phase: call first WITHOUT confirmation_token to get a preview and a one-time token, then call again WITH the token to enable it. " +
			"A zone can hold only one DNSSEC configuration, so the API answers 409 if it is already signed. After enabling, read the DS digest from list_dns_zone_dnssec_keys and add it at your domain's registrar — until you do, the zone is signed but the chain of trust is not established." + asyncZoneNote,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.CreateDnsDnssecKeyInput) (*mcp.CallToolResult, any, error) {
		id := strings.TrimSpace(input.ZoneID)
		if id == "" {
			return tools.ErrorText("zone_id is required"), nil, nil
		}
		params, msg := buildDnssecParams(input)
		if msg != "" {
			return tools.ErrorText(msg), nil, nil
		}
		target := tools.Target(req, id)

		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "create_dns_zone_dnssec_key", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("create_dns_zone_dnssec_key", "zone_id and validity", err)), nil, nil
			}
			created, _, err := client.DNSSECApi.ZonesKeysPost(ctx, id).DnssecKeyCreate(dnsSDK.DnssecKeyCreate{Properties: params}).Execute()
			return tools.ToResult(created, err)
		}

		token, err := confirm.Mint("create_dns_zone_dnssec_key", target)
		if err != nil {
			return nil, nil, err
		}
		return tools.TextResult(tools.Preview{
			Headline: "About to ENABLE DNSSEC on a DNS zone:",
			Fields: tools.Fields(
				"zone_id", id,
				"algorithm", string(params.KeyParameters.Algorithm),
				"ksk_bits", fmt.Sprintf("%d", params.KeyParameters.KskBits),
				"zsk_bits", fmt.Sprintf("%d", params.KeyParameters.ZskBits),
				"nsec_mode", string(params.NsecParameters.NsecMode),
				"nsec3_iterations", fmt.Sprintf("%d", params.NsecParameters.Nsec3Iterations),
				"nsec3_salt_bits", fmt.Sprintf("%d", params.NsecParameters.Nsec3SaltBits),
				"validity_days", fmt.Sprintf("%d", params.Validity),
			),
			Tool:      "create_dns_zone_dnssec_key",
			Replay:    tools.Fields("zone_id", id, "validity", fmt.Sprintf("%d", params.Validity)),
			TokenNote: "This token authorizes enabling DNSSEC on ONLY this zone",
		}.Render(token)), nil, nil
	})
}

func registerDeleteDnssecKey(server *mcp.Server, client *dnsSDK.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterTool(server, scope, tools.MethodDelete, &mcp.Tool{
		Name: "delete_dns_zone_dnssec_key",
		Description: "Disable DNSSEC on a primary DNS zone, removing its signing keys and DNSKEY records. Two-phase: call first WITHOUT confirmation_token to get a preview and a one-time token, then call again WITH the token to disable it. This is irreversible: re-enabling generates new keys. " +
			"REMOVE THE DS RECORD AT YOUR REGISTRAR FIRST. If the parent zone still publishes a DS record for keys that no longer exist, validating resolvers answer SERVFAIL and the whole zone stops resolving." + asyncZoneNote,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.DeleteDnsDnssecKeyInput) (*mcp.CallToolResult, any, error) {
		id := strings.TrimSpace(input.ZoneID)
		if id == "" {
			return tools.ErrorText("zone_id is required"), nil, nil
		}
		target := tools.Target(req, id)

		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "delete_dns_zone_dnssec_key", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("delete_dns_zone_dnssec_key", "zone_id", err)), nil, nil
			}
			if _, _, err := client.DNSSECApi.ZonesKeysDelete(ctx, id).Execute(); err != nil {
				return tools.ToResult(nil, err)
			}
			return tools.TextResult(fmt.Sprintf("Disabled DNSSEC on zone %s. Deletion is asynchronous; the API has accepted the request. If you have not already removed the DS record at your registrar, do it now — a stale DS record makes validating resolvers answer SERVFAIL for the entire zone.", id)), nil, nil
		}

		keys, _, err := client.DNSSECApi.ZonesKeysGet(ctx, id).Execute()
		if err != nil {
			if tools.IsNotFound(err) {
				return tools.ErrorText(fmt.Sprintf("zone %s has no DNSSEC configuration; nothing to delete", id)), nil, nil
			}
			return tools.ToResult(nil, err)
		}
		token, mErr := confirm.Mint("delete_dns_zone_dnssec_key", target)
		if mErr != nil {
			return nil, nil, mErr
		}
		radius := tools.DestroyedRadius()
		if keys.Metadata != nil {
			radius.Add("signing keys (and their DNSKEY records)", len(keys.Metadata.Items))
		}
		return tools.TextResult(tools.Preview{
			Headline: "About to DISABLE DNSSEC on a DNS zone. This is IRREVERSIBLE — re-enabling generates new keys.\n" +
				"REMOVE THE DS RECORD AT YOUR REGISTRAR FIRST: while the parent zone publishes a DS record for keys that no longer exist, validating resolvers answer SERVFAIL and the entire zone stops resolving.",
			Fields: tools.Fields(
				"zone_id", id,
				"algorithm", dnssecAlgorithm(keys),
				"nsec_mode", dnssecNsecMode(keys),
			),
			Radius:    radius,
			EmptyNote: "The API reports no signing keys for this zone.",
			Tool:      "delete_dns_zone_dnssec_key",
			Replay:    tools.Fields("zone_id", id),
			TokenNote: "This token authorizes disabling DNSSEC on ONLY this zone",
		}.Render(token)), nil, nil
	})
}

func dnssecAlgorithm(keys dnsSDK.DnssecKeyReadList) string {
	if keys.Properties == nil || keys.Properties.KeyParameters.Algorithm == nil {
		return ""
	}
	return string(*keys.Properties.KeyParameters.Algorithm)
}

func dnssecNsecMode(keys dnsSDK.DnssecKeyReadList) string {
	if keys.Properties == nil || keys.Properties.NsecParameters.NsecMode == nil {
		return ""
	}
	return string(*keys.Properties.NsecParameters.NsecMode)
}
