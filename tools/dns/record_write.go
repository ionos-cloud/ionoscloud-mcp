package dns

import (
	"context"
	"fmt"
	"strings"

	dnsSDK "github.com/ionos-cloud/sdk-go-bundle/products/dns/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

// RegisterRecordWriteTools registers the create/update/delete record tools.
func RegisterRecordWriteTools(server *mcp.Server, client *dnsSDK.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	registerCreateRecord(server, client, scope, confirm)
	registerUpdateRecord(server, client, scope)
	registerDeleteRecord(server, client, scope, confirm)
}

func registerCreateRecord(server *mcp.Server, client *dnsSDK.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterTool(server, scope, tools.MethodPost, &mcp.Tool{
		Name: "create_dns_record",
		Description: "Create one record in a primary DNS zone. Two-phase: call first WITHOUT confirmation_token to get a preview and a one-time token, then call again WITH the token (and the same name, type and content) to create it. Creates exactly one record per call. " +
			"Pass an empty name for a record on the zone apex. The API answers 409 when the record conflicts with an existing one — a CNAME cannot coexist with another record of the same name, and a TXT record cannot carry a second SPF entry." + asyncZoneNote,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.CreateDnsRecordInput) (*mcp.CallToolResult, any, error) {
		zoneID := strings.TrimSpace(input.ZoneID)
		if zoneID == "" {
			return tools.ErrorText("zone_id is required"), nil, nil
		}
		// An empty name is legal and means the zone apex, so it is trimmed but
		// never rejected.
		name := strings.TrimSpace(input.Name)
		content := strings.TrimSpace(input.Content)
		if content == "" {
			return tools.ErrorText("content is required to create a DNS record (e.g. 192.0.2.1 for an A record)"), nil, nil
		}
		recordType, msg := normalizeRecordType(input.Type)
		if msg != "" {
			return tools.ErrorText(msg), nil, nil
		}
		if msg := validateTTL(input.Ttl); msg != "" {
			return tools.ErrorText(msg), nil, nil
		}
		if msg := validatePriority(input.Priority, recordType, true); msg != "" {
			return tools.ErrorText(msg), nil, nil
		}
		priority := input.Priority
		priorityDropped := false
		if priority != nil && !priorityTypes[recordType] {
			priority, priorityDropped = nil, true
		}
		target := tools.Target(req, zoneID, name, string(recordType), content)

		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "create_dns_record", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("create_dns_record", "zone_id, name, type and content", err)), nil, nil
			}
			// A zero-valued literal rather than dnsSDK.NewRecord, which injects
			// ttl=3600 and enabled=true — see the "PATCH bodies" note in CLAUDE.md.
			props := dnsSDK.Record{Name: name, Type: recordType, Content: content}
			props.Ttl = input.Ttl
			props.Enabled = input.Enabled
			props.Priority = priority
			created, _, err := client.RecordsApi.ZonesRecordsPost(ctx, zoneID).RecordCreate(dnsSDK.RecordCreate{Properties: props}).Execute()
			return tools.ToResult(created, err)
		}

		token, err := confirm.Mint("create_dns_record", target)
		if err != nil {
			return nil, nil, err
		}
		displayName := name
		if displayName == "" {
			displayName = "(zone apex)"
		}
		headline := "About to CREATE one DNS record:"
		if priorityDropped {
			headline += fmt.Sprintf("\nNOTE: priority is ignored for a %s record, so it will not be sent.", recordType)
		}
		return tools.TextResult(tools.Preview{
			Headline: headline,
			Fields: tools.Fields(
				"zone_id", zoneID,
				"name", displayName,
				"type", string(recordType),
				"content", content,
				"ttl", tools.OptInt32(input.Ttl),
				"priority", tools.OptInt32(priority),
				"enabled", tools.OptBool(input.Enabled),
			),
			Tool:      "create_dns_record",
			Replay:    tools.Fields("zone_id", zoneID, "name", name, "type", string(recordType), "content", content),
			TokenNote: "This creates exactly one record. The token authorizes creating only this name+type+content in this zone",
		}.Render(token)), nil, nil
	})
}

func registerUpdateRecord(server *mcp.Server, client *dnsSDK.APIClient, scope tools.Scope) {
	tools.RegisterTool(server, scope, tools.MethodPut, &mcp.Tool{
		Name: "update_dns_record",
		Description: "Update a DNS record's content, TTL, priority, or whether it is published. The record's name and type are immutable here and are read and sent back unchanged — changing either is a delete plus a create. " +
			"This endpoint replaces the record's properties, so every field you omit is read and sent back explicitly rather than left out, which would let the API re-apply its own defaults." + asyncZoneNote,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.UpdateDnsRecordInput) (*mcp.CallToolResult, any, error) {
		zoneID := strings.TrimSpace(input.ZoneID)
		recordID := strings.TrimSpace(input.RecordID)
		if zoneID == "" {
			return tools.ErrorText("zone_id is required"), nil, nil
		}
		if recordID == "" {
			return tools.ErrorText("record_id is required"), nil, nil
		}
		if input.Content == nil && input.Ttl == nil && input.Priority == nil && input.Enabled == nil {
			return tools.ErrorText("nothing to update: provide at least one of content, ttl, priority, enabled"), nil, nil
		}
		if input.Content != nil && strings.TrimSpace(*input.Content) == "" {
			return tools.ErrorText("content must not be empty; omit it entirely to keep the current value"), nil, nil
		}
		if msg := validateTTL(input.Ttl); msg != "" {
			return tools.ErrorText(msg), nil, nil
		}

		current, _, err := client.RecordsApi.ZonesRecordsFindById(ctx, zoneID, recordID).Execute()
		if err != nil {
			if tools.IsNotFound(err) {
				return tools.ErrorText(fmt.Sprintf("DNS record %s does not exist in zone %s; nothing to update", recordID, zoneID)), nil, nil
			}
			return tools.ToResult(nil, err)
		}
		cp := current.GetProperties()
		if msg := validatePriority(input.Priority, cp.Type, false); msg != "" {
			return tools.ErrorText(msg), nil, nil
		}

		props := dnsSDK.Record{Name: cp.Name, Type: cp.Type, Content: cp.Content}
		if input.Content != nil {
			props.Content = strings.TrimSpace(*input.Content)
		}
		// ttl, priority and enabled are pointers the SDK omits when nil, and the
		// spec gives ttl and enabled defaults, so an omitted field must be sent
		// back explicitly rather than dropped.
		props.Ttl = firstNonNilInt32(input.Ttl, cp.Ttl)
		props.Enabled = firstNonNilBool(input.Enabled, cp.Enabled)
		props.Priority = firstNonNilInt32(input.Priority, cp.Priority)

		updated, _, err := client.RecordsApi.ZonesRecordsPut(ctx, zoneID, recordID).RecordEnsure(dnsSDK.RecordEnsure{Properties: props}).Execute()
		return tools.ToResult(updated, err)
	})
}

func registerDeleteRecord(server *mcp.Server, client *dnsSDK.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterTool(server, scope, tools.MethodDelete, &mcp.Tool{
		Name: "delete_dns_record",
		Description: "Delete one record from a primary DNS zone. Two-phase: call first WITHOUT confirmation_token to get a preview and a one-time token, then call again WITH the token to delete. This is irreversible. " +
			"Resolvers may keep serving the old answer until its TTL expires." + asyncZoneNote,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.DeleteDnsRecordInput) (*mcp.CallToolResult, any, error) {
		zoneID := strings.TrimSpace(input.ZoneID)
		recordID := strings.TrimSpace(input.RecordID)
		if zoneID == "" {
			return tools.ErrorText("zone_id is required"), nil, nil
		}
		if recordID == "" {
			return tools.ErrorText("record_id is required"), nil, nil
		}
		target := tools.Target(req, zoneID, recordID)

		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "delete_dns_record", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("delete_dns_record", "zone_id and record_id", err)), nil, nil
			}
			if _, _, err := client.RecordsApi.ZonesRecordsDelete(ctx, zoneID, recordID).Execute(); err != nil {
				return tools.ToResult(nil, err)
			}
			return tools.TextResult(tools.DeletedAsync("DNS record", recordID)), nil, nil
		}

		record, _, err := client.RecordsApi.ZonesRecordsFindById(ctx, zoneID, recordID).Execute()
		if err != nil {
			if tools.IsNotFound(err) {
				return tools.ErrorText(fmt.Sprintf("DNS record %s does not exist in zone %s; nothing to delete", recordID, zoneID)), nil, nil
			}
			return tools.ToResult(nil, err)
		}
		token, mErr := confirm.Mint("delete_dns_record", target)
		if mErr != nil {
			return nil, nil, mErr
		}
		cp := record.GetProperties()
		return tools.TextResult(tools.Preview{
			Headline: "About to DELETE a DNS record. This is IRREVERSIBLE.",
			Fields: tools.Fields(
				"record_id", recordID,
				"fqdn", record.Metadata.Fqdn,
				"type", string(cp.Type),
				"content", cp.Content,
				"ttl", tools.OptInt32(cp.Ttl),
				"enabled", tools.OptBool(cp.Enabled),
			),
			Tool:      "delete_dns_record",
			Replay:    tools.Fields("zone_id", zoneID, "record_id", recordID),
			TokenNote: "This token authorizes deleting ONLY this record",
		}.Render(token)), nil, nil
	})
}

func firstNonNilInt32(vals ...*int32) *int32 {
	for _, v := range vals {
		if v != nil {
			return v
		}
	}
	return nil
}

func firstNonNilBool(vals ...*bool) *bool {
	for _, v := range vals {
		if v != nil {
			return v
		}
	}
	return nil
}
