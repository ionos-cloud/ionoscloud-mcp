package compute

import (
	"context"
	"fmt"
	"strings"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

// RegisterSecurityGroupWriteTools registers the create/update/delete security group
// tools. Rules and assignment are handled elsewhere.
func RegisterSecurityGroupWriteTools(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	registerCreateSecurityGroup(server, client, scope, confirm)
	registerUpdateSecurityGroup(server, client, scope)
	registerDeleteSecurityGroup(server, client, scope, confirm)
}

func registerCreateSecurityGroup(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterTool(server, scope, tools.MethodPost, &mcp.Tool{
		Name: "create_security_group",
		Description: "Create one security group in a data center. Two-phase: call first WITHOUT confirmation_token to get a preview and a one-time token, then call again WITH the token (and the same datacenter_id and name) to create it. " +
			"A security group is a reusable set of firewall rules that can be assigned to several servers and NICs at once. The new group starts with NO rules, so it permits nothing until you add some with create_security_group_rule, and assigning an empty group to a server has no effect. " +
			"Assign it with assign_server_security_groups or assign_nic_security_groups. Creates exactly one group per call.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.CreateSecurityGroupInput) (*mcp.CallToolResult, any, error) {
		dcID := strings.TrimSpace(input.DatacenterID)
		name := strings.TrimSpace(input.Name)
		if dcID == "" {
			return tools.ErrorText("datacenter_id is required to create a security group"), nil, nil
		}
		if name == "" {
			return tools.ErrorText("name is required to create a security group"), nil, nil
		}
		target := tools.Target(dcID, name)

		// Phase 2: token present -> validate and execute.
		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "create_security_group", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("create_security_group", "datacenter_id and name", err)), nil, nil
			}
			props := ionos.NewSecurityGroupProperties(name)
			if input.Description != nil {
				props.SetDescription(*input.Description)
			}
			// The POST body is SecurityGroupRequest, which carries only properties.
			body := ionos.NewSecurityGroupRequest(*props)
			created, _, err := client.SecurityGroupsApi.DatacentersSecuritygroupsPost(ctx, dcID).SecurityGroup(*body).Execute()
			return tools.ToResult(created, err)
		}

		// Phase 1: no token -> preview and mint a one-time token.
		token, err := confirm.Mint("create_security_group", target)
		if err != nil {
			return nil, nil, err
		}
		return tools.TextResult(tools.Preview{
			Headline: "About to CREATE one security group:",
			Fields: tools.Fields(
				"datacenter_id", dcID,
				"name", name,
				"description", tools.OptStr(input.Description),
			),
			Tool:      "create_security_group",
			Replay:    tools.Fields("datacenter_id", dcID, "name", name),
			TokenNote: "The group starts with no rules, so it permits nothing until you add them with create_security_group_rule. The token authorizes creating only this group in this data center",
		}.Render(token)), nil, nil
	})
}

func registerUpdateSecurityGroup(server *mcp.Server, client *ionos.APIClient, scope tools.Scope) {
	tools.RegisterTool(server, scope, tools.MethodPatch, &mcp.Tool{
		Name: "update_security_group",
		Description: "Update a security group's name or description. Applies a partial update (only the fields you provide). " +
			"Omit name to keep the current one — it is read and sent back unchanged, because the API always receives this field and an empty value would clear the group's name. " +
			"This changes only the group's own properties; its rules are managed with create_security_group_rule, update_security_group_rule and delete_security_group_rule.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.UpdateSecurityGroupInput) (*mcp.CallToolResult, any, error) {
		dcID := strings.TrimSpace(input.DatacenterID)
		id := strings.TrimSpace(input.SecurityGroupID)
		if dcID == "" {
			return tools.ErrorText("datacenter_id is required"), nil, nil
		}
		if id == "" {
			return tools.ErrorText("security_group_id is required"), nil, nil
		}
		if input.Name == nil && input.Description == nil {
			return tools.ErrorText("nothing to update: provide at least one of name, description"), nil, nil
		}

		// The SDK always serializes name, so read the current value forward — otherwise
		// changing only the description would wipe it.
		name := ""
		if input.Name != nil {
			name = strings.TrimSpace(*input.Name)
			if name == "" {
				return tools.ErrorText("name must not be empty; omit it entirely to keep the group's current name"), nil, nil
			}
		} else {
			current, _, err := client.SecurityGroupsApi.DatacentersSecuritygroupsFindById(ctx, dcID, id).Depth(1).Execute()
			if err != nil {
				if tools.IsNotFound(err) {
					return tools.ErrorText(fmt.Sprintf("security group %s does not exist in data center %s; nothing to update", id, dcID)), nil, nil
				}
				return tools.ToResult(nil, err)
			}
			currentProps := current.GetProperties()
			name = currentProps.GetName()
		}

		props := ionos.NewSecurityGroupProperties(name)
		if input.Description != nil {
			props.SetDescription(*input.Description)
		}
		updated, _, err := client.SecurityGroupsApi.DatacentersSecuritygroupsPatch(ctx, dcID, id).SecurityGroup(*props).Execute()
		return tools.ToResult(updated, err)
	})
}

func registerDeleteSecurityGroup(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	tools.RegisterTool(server, scope, tools.MethodDelete, &mcp.Tool{
		Name: "delete_security_group",
		Description: "Delete a security group and all the rules in it. Two-phase: call first WITHOUT confirmation_token to get a blast-radius preview (its rules, and the servers and NICs it is assigned to) and a one-time token, then call again WITH the token to delete. " +
			"Every server and NIC the group was assigned to loses the protection its rules provided, which can either expose or cut off traffic depending on what the rules allowed. This is irreversible.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.DeleteSecurityGroupInput) (*mcp.CallToolResult, any, error) {
		dcID := strings.TrimSpace(input.DatacenterID)
		id := strings.TrimSpace(input.SecurityGroupID)
		if dcID == "" {
			return tools.ErrorText("datacenter_id is required"), nil, nil
		}
		if id == "" {
			return tools.ErrorText("security_group_id is required"), nil, nil
		}
		target := tools.Target(dcID, id)

		// Phase 2: token present -> validate and execute.
		if tools.HasToken(input.ConfirmationToken) {
			if err := confirm.Consume(*input.ConfirmationToken, "delete_security_group", target); err != nil {
				return tools.ErrorText(tools.ConfirmErrorText("delete_security_group", "datacenter_id and security_group_id", err)), nil, nil
			}
			_, err := client.SecurityGroupsApi.DatacentersSecuritygroupsDelete(ctx, dcID, id).Execute()
			if err != nil {
				return tools.ToResult(nil, err)
			}
			return tools.TextResult(tools.DeletedAsync("security group", id)), nil, nil
		}

		// Phase 1: no token -> blast radius, preview, mint a token.
		group, _, err := client.SecurityGroupsApi.DatacentersSecuritygroupsFindById(ctx, dcID, id).Depth(2).Execute()
		if err != nil {
			if tools.IsNotFound(err) {
				return tools.ErrorText(fmt.Sprintf("security group %s does not exist in data center %s; nothing to delete", id, dcID)), nil, nil
			}
			return tools.ToResult(nil, err)
		}
		props := group.GetProperties()
		radius := tools.AffectedRadius()
		if e := group.Entities; e != nil {
			if e.Rules != nil {
				radius.Add("rules deleted with the group", len(e.Rules.Items))
			}
			// Servers and NICs are not deleted; they lose the group's protection.
			if e.Servers != nil {
				radius.Add("servers that lose these rules", len(e.Servers.Items))
			}
			if e.Nics != nil {
				radius.Add("NICs that lose these rules", len(e.Nics.Items))
			}
		}
		token, mErr := confirm.Mint("delete_security_group", target)
		if mErr != nil {
			return nil, nil, mErr
		}
		return tools.TextResult(tools.Preview{
			Headline: "About to DELETE a security group and every rule in it. This is IRREVERSIBLE.\n" +
				"The servers and NICs listed below are not deleted, but they stop being governed by these rules — which may expose traffic the rules were restricting, or cut off traffic they were allowing.",
			Fields: tools.Fields(
				"datacenter_id", dcID,
				"security_group_id", id,
				"name", props.GetName(),
				"description", props.GetDescription(),
			),
			Radius:    radius,
			EmptyNote: "This group has no rules and is not assigned to anything; deleting it affects nothing else.",
			Tool:      "delete_security_group",
			Replay:    tools.Fields("datacenter_id", dcID, "security_group_id", id),
			TokenNote: "This token authorizes deleting ONLY this security group",
		}.Render(token)), nil, nil
	})
}
