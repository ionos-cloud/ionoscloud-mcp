package compute

import (
	"context"
	"fmt"
	"strings"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

// RegisterDatacenterWriteTools registers the create/update/delete data center
// tools. Each is gated by scope inside tools.RegisterTool (create/update need
// write, delete needs destructive), so they never appear in tools/list unless
// IONOS_MCP_TOOL_SCOPE opts in. create and delete share one confirmation store
// so their two-phase preview->execute flow is consistent across load modes.
func RegisterDatacenterWriteTools(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	registerCreateDatacenter(server, client, scope, confirm)
	registerUpdateDatacenter(server, client, scope)
	registerDeleteDatacenter(server, client, scope, confirm)
}

func registerCreateDatacenter(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterTool(server, scope, tools.MethodPost, &mcp.Tool{
		Name:        "create_datacenter",
		Description: "Create one virtual data center. Two-phase: call first WITHOUT confirmation_token to get a preview and a one-time token, then call again WITH the token (and the same name and location) to create it. Creates exactly one data center per call.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.CreateDatacenterInput) (*mcp.CallToolResult, any, error) {
		name := strings.TrimSpace(input.Name)
		location := strings.TrimSpace(input.Location)
		if name == "" {
			return tools.ErrorText("name is required to create a data center"), nil, nil
		}
		if location == "" {
			return tools.ErrorText("location is required to create a data center (e.g. de/fra)"), nil, nil
		}
		target := name + "|" + location

		// Phase 2: token present -> validate and execute.
		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "create_datacenter", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("create_datacenter", "name and location", err)), nil, nil
			}
			props := ionos.NewDatacenterPropertiesPost(location)
			props.SetName(name)
			if input.Description != nil {
				props.SetDescription(*input.Description)
			}
			if input.SecAuthProtection != nil {
				props.SetSecAuthProtection(*input.SecAuthProtection)
			}
			body := ionos.NewDatacenterPost(*props)
			created, _, err := client.DataCentersApi.DatacentersPost(ctx).Datacenter(*body).Execute()
			return tools.ToResult(created, err)
		}

		// Phase 1: no token -> preview and mint a one-time token.
		token, err := confirm.Mint("create_datacenter", target)
		if err != nil {
			return nil, nil, err
		}
		return tools.TextResult(formatCreatePreview(input, name, location, token)), nil, nil
	})
}

func registerUpdateDatacenter(server *mcp.Server, client *ionos.APIClient, scope tools.Scope) {
	tools.RegisterTool(server, scope, tools.MethodPatch, &mcp.Tool{
		Name:        "update_datacenter",
		Description: "Update a virtual data center's name, description, or protection flag. The location is immutable and cannot be changed. Applies a partial update (only the fields you provide).",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.UpdateDatacenterInput) (*mcp.CallToolResult, any, error) {
		id := strings.TrimSpace(input.DatacenterID)
		if id == "" {
			return tools.ErrorText("datacenter_id is required"), nil, nil
		}
		if input.Name == nil && input.Description == nil && input.SecAuthProtection == nil {
			return tools.ErrorText("nothing to update: provide at least one of name, description, sec_auth_protection"), nil, nil
		}
		// A zero-valued literal rather than NewDatacenterPropertiesPutWithDefaults():
		// see the "PATCH bodies" note in CLAUDE.md.
		props := &ionos.DatacenterPropertiesPut{}
		if input.Name != nil {
			props.SetName(*input.Name)
		}
		if input.Description != nil {
			props.SetDescription(*input.Description)
		}
		if input.SecAuthProtection != nil {
			props.SetSecAuthProtection(*input.SecAuthProtection)
		}
		updated, _, err := client.DataCentersApi.DatacentersPatch(ctx, id).Datacenter(*props).Execute()
		return tools.ToResult(updated, err)
	})
}

func registerDeleteDatacenter(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterTool(server, scope, tools.MethodDelete, &mcp.Tool{
		Name:        "delete_datacenter",
		Description: "Delete a virtual data center and EVERYTHING inside it (servers, volumes, LANs, load balancers, and more). Two-phase: call first WITHOUT confirmation_token to get a blast-radius preview and a one-time token, then call again WITH the token to delete. This is irreversible.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.DeleteDatacenterInput) (*mcp.CallToolResult, any, error) {
		id := strings.TrimSpace(input.DatacenterID)
		if id == "" {
			return tools.ErrorText("datacenter_id is required"), nil, nil
		}

		// Phase 2: token present -> validate and execute.
		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "delete_datacenter", id); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("delete_datacenter", "datacenter_id", err)), nil, nil
			}
			_, err := client.DataCentersApi.DatacentersDelete(ctx, id).Execute()
			if err != nil {
				return tools.ToResult(nil, err)
			}
			return tools.TextResult(tools.DeletedAsync("data center", id)), nil, nil
		}

		// Phase 1: no token -> compute blast radius, preview, and mint a token.
		dc, _, err := client.DataCentersApi.DatacentersFindById(ctx, id).Depth(2).Execute()
		if err != nil {
			if tools.IsNotFound(err) {
				return tools.ErrorText(fmt.Sprintf("data center %s does not exist; nothing to delete", id)), nil, nil
			}
			return tools.ToResult(nil, err)
		}
		radius := datacenterBlastRadius(dc)
		token, mErr := confirm.Mint("delete_datacenter", id)
		if mErr != nil {
			return nil, nil, mErr
		}
		return tools.TextResult(formatDeletePreview(dc, id, radius, token)), nil, nil
	})
}

// datacenterBlastRadius counts the resources a delete would destroy, from a
// datacenter fetched at depth 2 (which populates the direct child collections).
func datacenterBlastRadius(dc ionos.Datacenter) *tools.BlastRadius {
	r := &tools.BlastRadius{}
	e := dc.Entities
	if e == nil {
		return r
	}
	if e.Servers != nil {
		r.Add("servers", len(e.Servers.Items))
	}
	if e.Volumes != nil {
		r.Add("volumes", len(e.Volumes.Items))
	}
	if e.Loadbalancers != nil {
		r.Add("load balancers", len(e.Loadbalancers.Items))
	}
	if e.Lans != nil {
		r.Add("LANs", len(e.Lans.Items))
	}
	if e.Networkloadbalancers != nil {
		r.Add("network load balancers", len(e.Networkloadbalancers.Items))
	}
	if e.Natgateways != nil {
		r.Add("NAT gateways", len(e.Natgateways.Items))
	}
	if e.Securitygroups != nil {
		r.Add("security groups", len(e.Securitygroups.Items))
	}
	return r
}

func formatCreatePreview(in tools.CreateDatacenterInput, name, location, token string) string {
	return tools.Preview{
		Headline: "About to CREATE one data center:",
		Fields: tools.Fields(
			"name", name,
			"location", location,
			"description", tools.OptStr(in.Description),
			"sec_auth_protection", tools.OptBool(in.SecAuthProtection),
		),
		Tool:      "create_datacenter",
		Replay:    tools.Fields("name", name, "location", location),
		TokenNote: "This creates exactly one data center. The token authorizes creating only this name+location",
	}.Render(token)
}

func formatDeletePreview(dc ionos.Datacenter, id string, radius *tools.BlastRadius, token string) string {
	props := dc.GetProperties()
	return tools.Preview{
		Headline: "About to DELETE a data center and everything inside it. This is IRREVERSIBLE.",
		Fields: tools.Fields(
			"id", id,
			"name", props.GetName(),
			"location", props.GetLocation(),
		),
		Radius:    radius,
		EmptyNote: "This data center is empty; deleting removes only the (empty) data center itself.",
		Tool:      "delete_datacenter",
		Replay:    tools.Fields("datacenter_id", id),
		TokenNote: "This token authorizes deleting ONLY this data center",
	}.Render(token)
}
