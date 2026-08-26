package dns

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	dnsSDK "github.com/ionos-cloud/sdk-go-bundle/products/dns/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

// recordCountLimit is the API's maximum page size. There is no total-count field on
// a list response, so a full page means "at least this many" rather than an exact
// figure, and the previews below say so.
const recordCountLimit = 1000

// RegisterZoneWriteTools registers the create/update/delete zone tools and the
// zone-file import.
func RegisterZoneWriteTools(server *mcp.Server, client *dnsSDK.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	registerCreateZone(server, client, scope, confirm)
	registerUpdateZone(server, client, scope)
	registerDeleteZone(server, client, scope, confirm)
	registerImportZoneFile(server, client, scope, confirm)
}

func registerCreateZone(server *mcp.Server, client *dnsSDK.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterTool(server, scope, tools.MethodPost, &mcp.Tool{
		Name: "create_dns_zone",
		Description: "Create one primary DNS zone, with default NS and SOA records. Two-phase: call first WITHOUT confirmation_token to get a preview and a one-time token, then call again WITH the token (and the same zone_name) to create it. Creates exactly one zone per call. " +
			"The zone does not serve your domain until you point the registrar at the nameservers listed in the response's metadata.nameservers." + asyncZoneNote,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.CreateDnsZoneInput) (*mcp.CallToolResult, any, error) {
		name := strings.TrimSpace(input.ZoneName)
		if name == "" {
			return tools.ErrorText("zone_name is required to create a DNS zone (e.g. example.com)"), nil, nil
		}
		target := tools.Target(req, name)

		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "create_dns_zone", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("create_dns_zone", "zone_name", err)), nil, nil
			}
			// A zero-valued literal rather than dnsSDK.NewZone, which injects
			// enabled=true — see the "PATCH bodies" note in CLAUDE.md.
			props := dnsSDK.Zone{ZoneName: name}
			if input.Description != nil {
				props.Description = input.Description
			}
			if input.Enabled != nil {
				props.Enabled = input.Enabled
			}
			created, _, err := client.ZonesApi.ZonesPost(ctx).ZoneCreate(dnsSDK.ZoneCreate{Properties: props}).Execute()
			return tools.ToResult(created, err)
		}

		token, err := confirm.Mint("create_dns_zone", target)
		if err != nil {
			return nil, nil, err
		}
		return tools.TextResult(tools.Preview{
			Headline: "About to CREATE one primary DNS zone:",
			Fields: tools.Fields(
				"zone_name", name,
				"description", tools.OptStr(input.Description),
				"enabled", tools.OptBool(input.Enabled),
			),
			Tool:      "create_dns_zone",
			Replay:    tools.Fields("zone_name", name),
			TokenNote: "This creates exactly one zone. The token authorizes creating only this zone_name",
		}.Render(token)), nil, nil
	})
}

func registerUpdateZone(server *mcp.Server, client *dnsSDK.APIClient, scope tools.Scope) {
	tools.RegisterTool(server, scope, tools.MethodPut, &mcp.Tool{
		Name: "update_dns_zone",
		Description: "Update a primary DNS zone's description, or enable/disable it. The zone name is immutable and is read and sent back unchanged. " +
			"Disabling a zone stops it answering lookups for every record it contains, without deleting anything." + asyncZoneNote,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.UpdateDnsZoneInput) (*mcp.CallToolResult, any, error) {
		id := strings.TrimSpace(input.ZoneID)
		if id == "" {
			return tools.ErrorText("zone_id is required"), nil, nil
		}
		if input.Description == nil && input.Enabled == nil {
			return tools.ErrorText("nothing to update: provide at least one of description, enabled"), nil, nil
		}

		// The endpoint replaces the zone's properties and the SDK serializes
		// zoneName unconditionally, so read the zone and override only what the
		// caller supplied. Without this the PUT would send an empty zone name.
		current, _, err := client.ZonesApi.ZonesFindById(ctx, id).Execute()
		if err != nil {
			if tools.IsNotFound(err) {
				return tools.ErrorText(fmt.Sprintf("DNS zone %s does not exist; nothing to update", id)), nil, nil
			}
			return tools.ToResult(nil, err)
		}
		cp := current.GetProperties()

		props := dnsSDK.Zone{ZoneName: cp.ZoneName}
		switch {
		case input.Description != nil:
			props.Description = input.Description
		case cp.Description != nil:
			props.Description = cp.Description
		}
		switch {
		case input.Enabled != nil:
			props.Enabled = input.Enabled
		case cp.Enabled != nil:
			props.Enabled = cp.Enabled
		}

		updated, _, err := client.ZonesApi.ZonesPut(ctx, id).ZoneEnsure(dnsSDK.ZoneEnsure{Properties: props}).Execute()
		return tools.ToResult(updated, err)
	})
}

func registerDeleteZone(server *mcp.Server, client *dnsSDK.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterTool(server, scope, tools.MethodDelete, &mcp.Tool{
		Name: "delete_dns_zone",
		Description: "Delete a primary DNS zone and every record it contains. Two-phase: call first WITHOUT confirmation_token to get a blast-radius preview and a one-time token, then call again WITH the token to delete. This is irreversible. " +
			"Any domain still pointing at this zone's nameservers stops resolving." + asyncZoneNote,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.DeleteDnsZoneInput) (*mcp.CallToolResult, any, error) {
		id := strings.TrimSpace(input.ZoneID)
		if id == "" {
			return tools.ErrorText("zone_id is required"), nil, nil
		}
		target := tools.Target(req, id)

		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "delete_dns_zone", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("delete_dns_zone", "zone_id", err)), nil, nil
			}
			if _, _, err := client.ZonesApi.ZonesDelete(ctx, id).Execute(); err != nil {
				return tools.ToResult(nil, err)
			}
			return tools.TextResult(tools.DeletedAsync("DNS zone", id) +
				" Poll get_dns_zone to follow it: metadata.state goes DESTROYING and then the zone stops resolving."), nil, nil
		}

		zone, _, err := client.ZonesApi.ZonesFindById(ctx, id).Execute()
		if err != nil {
			if tools.IsNotFound(err) {
				return tools.ErrorText(fmt.Sprintf("DNS zone %s does not exist; nothing to delete", id)), nil, nil
			}
			return tools.ToResult(nil, err)
		}
		count, capped, countErr := zoneRecordCount(ctx, client, id)
		keyCount, keyErr := zoneDnssecKeyCount(ctx, client, id)
		radius := tools.DestroyedRadius()
		radius.Add("records", count)
		radius.Add("DNSSEC signing keys", keyCount)

		token, mErr := confirm.Mint("delete_dns_zone", target)
		if mErr != nil {
			return nil, nil, mErr
		}
		headline := "About to DELETE a DNS zone and every record in it. This is IRREVERSIBLE."
		if capped {
			headline += fmt.Sprintf("\nNOTE: the record count below is a floor — the API returns at most %d per page and reports no total.", recordCountLimit)
		}
		emptyNote := "This zone has no records; deleting removes only the (empty) zone itself."
		if unreadable := incompleteRadiusNote(errLabel(countErr, "records"), errLabel(keyErr, "DNSSEC keys")); unreadable != "" {
			headline += unreadable
			emptyNote = "" // an unreadable collection must not read as an empty one
		}
		cp := zone.GetProperties()
		return tools.TextResult(tools.Preview{
			Headline: headline,
			Fields: tools.Fields(
				"zone_id", id,
				"zone_name", cp.ZoneName,
				"state", string(zone.Metadata.State),
				"enabled", tools.OptBool(cp.Enabled),
			),
			Radius:    radius,
			EmptyNote: emptyNote,
			Tool:      "delete_dns_zone",
			Replay:    tools.Fields("zone_id", id),
			TokenNote: "This token authorizes deleting ONLY this zone",
		}.Render(token)), nil, nil
	})
}

func registerImportZoneFile(server *mcp.Server, client *dnsSDK.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterActionTool(server, scope,
		tools.Action{Verb: "import_", Method: tools.MethodPut, Idempotent: true},
		&mcp.Tool{
			Name: "import_dns_zone_file",
			Description: "Replace ALL records in a primary DNS zone with the contents of a zone file in BIND format (RFC 1035). Every record currently in the zone is deleted, so this needs IONOS_MCP_TOOL_SCOPE to include destructive even though the method is a PUT. " +
				"Two-phase: call first WITHOUT confirmation_token to see how many existing records would be replaced and get a one-time token, then call again WITH the token to import. " +
				"SOA and NS records in the file are accepted but ignored, so a file exported from another provider can be imported as-is. Returns the resulting record list." + syncNote,
		}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ImportDnsZoneFileInput) (*mcp.CallToolResult, any, error) {
			id := strings.TrimSpace(input.ZoneID)
			if id == "" {
				return tools.ErrorText("zone_id is required"), nil, nil
			}
			if strings.TrimSpace(input.ZoneFile) == "" {
				return tools.ErrorText("zone_file is required; it must be a zone file in BIND format (RFC 1035)"), nil, nil
			}

			sum := sha256.Sum256([]byte(input.ZoneFile))
			target := tools.Target(req, id, hex.EncodeToString(sum[:]))

			if tools.HasToken(input.ConfirmationToken) {
				if err := confirm.Consume(*input.ConfirmationToken, "import_dns_zone_file", target); err != nil {
					return tools.ErrorText(tools.ConfirmErrorText("import_dns_zone_file", "zone_id and zone_file", err)), nil, nil
				}
				records, _, err := client.ZoneFilesApi.ZonesZonefilePut(ctx, id).Body(input.ZoneFile).Execute()
				return tools.ToResult(records, err)
			}

			zone, _, err := client.ZonesApi.ZonesFindById(ctx, id).Execute()
			if err != nil {
				if tools.IsNotFound(err) {
					return tools.ErrorText(fmt.Sprintf("DNS zone %s does not exist; nothing to import into", id)), nil, nil
				}
				return tools.ToResult(nil, err)
			}
			count, capped, countErr := zoneRecordCount(ctx, client, id)
			radius := tools.DestroyedRadius()
			radius.Add("existing records (replaced by the file's contents)", count)

			token, mErr := confirm.Mint("import_dns_zone_file", target)
			if mErr != nil {
				return nil, nil, mErr
			}
			headline := "About to REPLACE every record in a DNS zone from a zone file. This is IRREVERSIBLE."
			if capped {
				headline += fmt.Sprintf("\nNOTE: the record count below is a floor — the API returns at most %d per page and reports no total.", recordCountLimit)
			}
			emptyNote := "This zone has no records yet, so the import only adds the file's contents."
			if unreadable := incompleteRadiusNote(errLabel(countErr, "records")); unreadable != "" {
				headline += unreadable
				emptyNote = ""
			}
			return tools.TextResult(tools.Preview{
				Headline: headline,
				Fields: tools.Fields(
					"zone_id", id,
					"zone_name", zone.GetProperties().ZoneName,
					"zone_file_lines", fmt.Sprintf("%d", strings.Count(strings.TrimRight(input.ZoneFile, "\n"), "\n")+1),
				),
				Radius:    radius,
				EmptyNote: emptyNote,
				Tool:      "import_dns_zone_file",
				Replay:    tools.Fields("zone_id", id, "zone_file", "(the same file, byte for byte)"),
				TokenNote: "This token authorizes importing ONLY this exact file into this zone; any change to the file needs a fresh preview",
			}.Render(token)), nil, nil
		})
}

// zoneRecordCount counts a primary zone's records for a blast-radius preview, and
// reports whether the page came back full. ZonesRecordsGet cannot be used: the SDK
// exposes no limit on it, so its 100-item default would understate the count.
//
// A failure is returned rather than folded into a zero, because "none" and "could not
// tell" are different claims to put in front of someone authorizing a delete.
func zoneRecordCount(ctx context.Context, client *dnsSDK.APIClient, zoneID string) (n int, capped bool, err error) {
	records, _, err := client.RecordsApi.RecordsGet(ctx).FilterZoneId(zoneID).Limit(recordCountLimit).Execute()
	if err != nil {
		return 0, false, err
	}
	return len(records.Items), len(records.Items) >= recordCountLimit, nil
}

// zoneDnssecKeyCount reports how many DNSSEC keys the zone has. Only a 404 means
// "unsigned";
func zoneDnssecKeyCount(ctx context.Context, client *dnsSDK.APIClient, zoneID string) (int, error) {
	keys, _, err := client.DNSSECApi.ZonesKeysGet(ctx, zoneID).Execute()
	switch {
	case err != nil && tools.IsNotFound(err):
		return 0, nil
	case err != nil:
		return 0, err
	case keys.Metadata == nil:
		return 0, nil
	}
	return len(keys.Metadata.Items), nil
}

// incompleteRadiusNote warns that a blast radius could not be fully determined. An
// unreadable collection must never render as an empty one.
func incompleteRadiusNote(what ...string) string {
	var named []string
	for _, w := range what {
		if w != "" {
			named = append(named, w)
		}
	}
	if len(named) == 0 {
		return ""
	}
	return fmt.Sprintf("\nWARNING: could not read this zone's %s, so the list below is INCOMPLETE — this may destroy more than it shows.", strings.Join(named, " or "))
}

// errLabel names a collection when reading it failed, and "" when it succeeded, so
// callers can build one warning covering however many lookups went wrong.
func errLabel(err error, label string) string {
	if err == nil {
		return ""
	}
	return label
}
