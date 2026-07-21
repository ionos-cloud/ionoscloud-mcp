package compute

import (
	"context"
	"errors"
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
		if input.ConfirmationToken != nil && *input.ConfirmationToken != "" {
			if err := confirm.Consume(*input.ConfirmationToken, "create_datacenter", target); err != nil {
				return tools.ErrorText(createConfirmError(err)), nil, nil
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
		props := ionos.NewDatacenterPropertiesPutWithDefaults()
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
		if input.ConfirmationToken != nil && *input.ConfirmationToken != "" {
			if err := confirm.Consume(*input.ConfirmationToken, "delete_datacenter", id); err != nil {
				return tools.ErrorText(deleteConfirmError(err)), nil, nil
			}
			_, err := client.DataCentersApi.DatacentersDelete(ctx, id).Execute()
			if err != nil {
				return tools.ToResult(nil, err)
			}
			return tools.TextResult(fmt.Sprintf("Deleted data center %s. Deletion is asynchronous; the API has accepted the request.", id)), nil, nil
		}

		// Phase 1: no token -> compute blast radius, preview, and mint a token.
		dc, _, err := client.DataCentersApi.DatacentersFindById(ctx, id).Depth(2).Execute()
		if err != nil {
			if tools.IsNotFound(err) {
				return tools.ErrorText(fmt.Sprintf("data center %s does not exist; nothing to delete", id)), nil, nil
			}
			return tools.ToResult(nil, err)
		}
		counts, total := datacenterBlastRadius(dc)
		token, mErr := confirm.Mint("delete_datacenter", id)
		if mErr != nil {
			return nil, nil, mErr
		}
		return tools.TextResult(formatDeletePreview(dc, id, counts, total, token)), nil, nil
	})
}

// labeledCount is one line of a blast-radius preview.
type labeledCount struct {
	label string
	count int
}

// datacenterBlastRadius counts the resources a delete would destroy, from a
// datacenter fetched at depth 2 (which populates the direct child collections).
// It returns per-category counts (non-zero only) and the total.
func datacenterBlastRadius(dc ionos.Datacenter) ([]labeledCount, int) {
	var counts []labeledCount
	total := 0
	e := dc.Entities
	if e == nil {
		return counts, 0
	}
	add := func(label string, n int) {
		if n > 0 {
			counts = append(counts, labeledCount{label: label, count: n})
			total += n
		}
	}
	if e.Servers != nil {
		add("servers", len(e.Servers.Items))
	}
	if e.Volumes != nil {
		add("volumes", len(e.Volumes.Items))
	}
	if e.Loadbalancers != nil {
		add("load balancers", len(e.Loadbalancers.Items))
	}
	if e.Lans != nil {
		add("LANs", len(e.Lans.Items))
	}
	if e.Networkloadbalancers != nil {
		add("network load balancers", len(e.Networkloadbalancers.Items))
	}
	if e.Natgateways != nil {
		add("NAT gateways", len(e.Natgateways.Items))
	}
	if e.Securitygroups != nil {
		add("security groups", len(e.Securitygroups.Items))
	}
	return counts, total
}

func formatCreatePreview(in tools.CreateDatacenterInput, name, location, token string) string {
	var b strings.Builder
	b.WriteString("About to CREATE one data center:\n")
	fmt.Fprintf(&b, "  name:     %s\n", name)
	fmt.Fprintf(&b, "  location: %s\n", location)
	if in.Description != nil && *in.Description != "" {
		fmt.Fprintf(&b, "  description: %s\n", *in.Description)
	}
	if in.SecAuthProtection != nil {
		fmt.Fprintf(&b, "  sec_auth_protection: %t\n", *in.SecAuthProtection)
	}
	b.WriteString("\nThis creates exactly one data center. To proceed, call create_datacenter again with the same name and location plus:\n")
	fmt.Fprintf(&b, "  confirmation_token: %s\n", token)
	fmt.Fprintf(&b, "The token authorizes creating only this name+location and expires in %s.", tools.ConfirmationTTL)
	return b.String()
}

func formatDeletePreview(dc ionos.Datacenter, id string, counts []labeledCount, total int, token string) string {
	props := dc.GetProperties()
	name := props.GetName()
	location := props.GetLocation()

	var b strings.Builder
	b.WriteString("About to DELETE a data center and everything inside it. This is IRREVERSIBLE.\n")
	fmt.Fprintf(&b, "  id:       %s\n", id)
	if name != "" {
		fmt.Fprintf(&b, "  name:     %s\n", name)
	}
	if location != "" {
		fmt.Fprintf(&b, "  location: %s\n", location)
	}
	if total == 0 {
		b.WriteString("\nThis data center is empty; deleting removes only the (empty) data center itself.\n")
	} else {
		b.WriteString("\nContained resources that will be destroyed:\n")
		for _, c := range counts {
			fmt.Fprintf(&b, "  - %d %s\n", c.count, c.label)
		}
		fmt.Fprintf(&b, "Total resources that will be destroyed: %d\n", total)
	}
	b.WriteString("\nTo proceed, call delete_datacenter again with:\n")
	fmt.Fprintf(&b, "  datacenter_id: %s\n", id)
	fmt.Fprintf(&b, "  confirmation_token: %s\n", token)
	fmt.Fprintf(&b, "This token authorizes deleting ONLY this data center and expires in %s.", tools.ConfirmationTTL)
	return b.String()
}

func createConfirmError(err error) string {
	switch {
	case errors.Is(err, tools.ErrTokenMismatch):
		return "confirmation_token was issued for a different name/location; re-run create_datacenter with only name and location to preview and get a fresh token"
	case errors.Is(err, tools.ErrTokenExpired):
		return "confirmation_token expired; re-run create_datacenter with only name and location for a fresh preview and token"
	default: // ErrTokenUnknown
		return "confirmation_token not recognized (already used or never issued); re-run create_datacenter with only name and location for a preview and token"
	}
}

func deleteConfirmError(err error) string {
	switch {
	case errors.Is(err, tools.ErrTokenMismatch):
		return "confirmation_token was issued for a different data center; re-run delete_datacenter with only datacenter_id to preview THIS one and get a fresh token"
	case errors.Is(err, tools.ErrTokenExpired):
		return "confirmation_token expired; re-run delete_datacenter with only datacenter_id for a fresh preview and token"
	default: // ErrTokenUnknown
		return "confirmation_token not recognized (already used or never issued); re-run delete_datacenter with only datacenter_id for a preview and token"
	}
}
