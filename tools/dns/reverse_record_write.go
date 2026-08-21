package dns

import (
	"context"
	"fmt"
	"strings"

	dnsSDK "github.com/ionos-cloud/sdk-go-bundle/products/dns/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

// Reverse records are the one DNS resource whose writes are not all asynchronous:
// create answers 201 and update 200 with the finished record, while delete answers
// 202. None of them expose a provisioning state, so none can be polled.

// RegisterReverseRecordWriteTools registers the create/update/delete reverse record
// tools.
func RegisterReverseRecordWriteTools(server *mcp.Server, client *dnsSDK.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	registerCreateReverseRecord(server, client, scope, confirm)
	registerUpdateReverseRecord(server, client, scope)
	registerDeleteReverseRecord(server, client, scope, confirm)
}

func registerCreateReverseRecord(server *mcp.Server, client *dnsSDK.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterTool(server, scope, tools.MethodPost, &mcp.Tool{
		Name: "create_dns_reverse_record",
		Description: "Create one reverse DNS (PTR) record, making an IP resolve back to a hostname. Two-phase: call first WITHOUT confirmation_token to get a preview and a one-time token, then call again WITH the token (and the same ip) to create it. Creates exactly one record per call. " +
			"The IP must be one your contract owns: an IPv4 from one of your IP blocks (see list_ip_blocks) or an IPv6 from a VDC. The API answers 409 for an IP that is not eligible, and each IP can carry only one reverse record." + syncNote,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.CreateDnsReverseRecordInput) (*mcp.CallToolResult, any, error) {
		name := strings.TrimSpace(input.Name)
		if name == "" {
			return tools.ErrorText("name is required to create a reverse DNS record (the hostname the IP should resolve back to, e.g. mail.example.com)"), nil, nil
		}
		ip, msg := validateIP("ip", input.Ip)
		if msg != "" {
			return tools.ErrorText(msg), nil, nil
		}
		target := tools.Target(req, ip)

		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "create_dns_reverse_record", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("create_dns_reverse_record", "name and ip", err)), nil, nil
			}
			props := dnsSDK.ReverseRecord{Name: name, Ip: ip}
			if input.Description != nil {
				props.Description = input.Description
			}
			created, _, err := client.ReverseRecordsApi.ReverserecordsPost(ctx).ReverseRecordCreate(dnsSDK.ReverseRecordCreate{Properties: props}).Execute()
			return tools.ToResult(created, err)
		}

		token, err := confirm.Mint("create_dns_reverse_record", target)
		if err != nil {
			return nil, nil, err
		}
		return tools.TextResult(tools.Preview{
			Headline: "About to CREATE one reverse DNS record:",
			Fields: tools.Fields(
				"ip", ip,
				"name", name,
				"description", tools.OptStr(input.Description),
			),
			Tool:      "create_dns_reverse_record",
			Replay:    tools.Fields("name", name, "ip", ip),
			TokenNote: "This creates exactly one record. The token authorizes creating only the record for this ip",
		}.Render(token)), nil, nil
	})
}

func registerUpdateReverseRecord(server *mcp.Server, client *dnsSDK.APIClient, scope tools.Scope) {
	tools.RegisterTool(server, scope, tools.MethodPut, &mcp.Tool{
		Name:        "update_dns_reverse_record",
		Description: "Update the hostname or description of a reverse DNS record. The IP is immutable here and is read and sent back unchanged — pointing a different IP at a hostname is a create, not an update." + syncNote,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.UpdateDnsReverseRecordInput) (*mcp.CallToolResult, any, error) {
		id := strings.TrimSpace(input.ReverseRecordID)
		if id == "" {
			return tools.ErrorText("reverse_record_id is required"), nil, nil
		}
		if input.Name == nil && input.Description == nil {
			return tools.ErrorText("nothing to update: provide at least one of name, description"), nil, nil
		}
		if input.Name != nil && strings.TrimSpace(*input.Name) == "" {
			return tools.ErrorText("name must not be empty; omit it entirely to keep the current hostname"), nil, nil
		}

		// name and ip are both serialized unconditionally, so read the record and
		// override only what the caller supplied.
		current, _, err := client.ReverseRecordsApi.ReverserecordsFindById(ctx, id).Execute()
		if err != nil {
			if tools.IsNotFound(err) {
				return tools.ErrorText(fmt.Sprintf("reverse DNS record %s does not exist; nothing to update", id)), nil, nil
			}
			return tools.ToResult(nil, err)
		}
		cp := current.GetProperties()

		props := dnsSDK.ReverseRecord{Name: cp.Name, Ip: cp.Ip}
		if input.Name != nil {
			props.Name = strings.TrimSpace(*input.Name)
		}
		switch {
		case input.Description != nil:
			props.Description = input.Description
		case cp.Description != nil:
			props.Description = cp.Description
		}

		updated, _, err := client.ReverseRecordsApi.ReverserecordsPut(ctx, id).ReverseRecordEnsure(dnsSDK.ReverseRecordEnsure{Properties: props}).Execute()
		return tools.ToResult(updated, err)
	})
}

func registerDeleteReverseRecord(server *mcp.Server, client *dnsSDK.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterTool(server, scope, tools.MethodDelete, &mcp.Tool{
		Name: "delete_dns_reverse_record",
		Description: "Delete a reverse DNS record, so its IP stops resolving back to a hostname. Two-phase: call first WITHOUT confirmation_token to get a preview and a one-time token, then call again WITH the token to delete. This is irreversible. " +
			"Mail servers commonly reject senders with no reverse DNS, so check what depends on the record first." + acceptedNoPollNote,
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.DeleteDnsReverseRecordInput) (*mcp.CallToolResult, any, error) {
		id := strings.TrimSpace(input.ReverseRecordID)
		if id == "" {
			return tools.ErrorText("reverse_record_id is required"), nil, nil
		}
		target := tools.Target(req, id)

		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "delete_dns_reverse_record", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("delete_dns_reverse_record", "reverse_record_id", err)), nil, nil
			}
			if _, _, err := client.ReverseRecordsApi.ReverserecordsDelete(ctx, id).Execute(); err != nil {
				return tools.ToResult(nil, err)
			}
			return tools.TextResult(tools.DeletedAsync("reverse DNS record", id)), nil, nil
		}

		record, _, err := client.ReverseRecordsApi.ReverserecordsFindById(ctx, id).Execute()
		if err != nil {
			if tools.IsNotFound(err) {
				return tools.ErrorText(fmt.Sprintf("reverse DNS record %s does not exist; nothing to delete", id)), nil, nil
			}
			return tools.ToResult(nil, err)
		}
		token, mErr := confirm.Mint("delete_dns_reverse_record", target)
		if mErr != nil {
			return nil, nil, mErr
		}
		cp := record.GetProperties()
		return tools.TextResult(tools.Preview{
			Headline: "About to DELETE a reverse DNS record. This is IRREVERSIBLE.",
			Fields: tools.Fields(
				"reverse_record_id", id,
				"ip", cp.Ip,
				"name", cp.Name,
				"description", tools.OptStr(cp.Description),
			),
			Tool:      "delete_dns_reverse_record",
			Replay:    tools.Fields("reverse_record_id", id),
			TokenNote: "This token authorizes deleting ONLY this record",
		}.Render(token)), nil, nil
	})
}
