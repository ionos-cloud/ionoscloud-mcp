package compute

import (
	"context"
	"fmt"
	"strings"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

// Security group assignment. Both endpoints are PUTs that REPLACE the whole
// assignment set rather than adding to it, which is the trap these tools exist
// to signpost: a caller that passes one group ID meaning "also add this" would
// unassign every other group. The verb is assign_ rather than update_ because
// the tool takes a complete set, and the class is write rather than destructive —
// no resource is destroyed, only the mapping changes, and it is reversible by
// assigning the previous set back.

// RegisterSecurityGroupAssignTools registers the server and NIC security-group
// assignment tools.
func RegisterSecurityGroupAssignTools(server *mcp.Server, client *ionos.APIClient, scope tools.Scope) {
	registerAssignServerSecurityGroups(server, client, scope)
	registerAssignNicSecurityGroups(server, client, scope)
}

func registerAssignServerSecurityGroups(server *mcp.Server, client *ionos.APIClient, scope tools.Scope) {
	tools.RegisterActionTool(server, scope,
		tools.Action{Verb: "assign_", Method: tools.MethodPut, Idempotent: true},
		&mcp.Tool{
			Name: "assign_server_security_groups",
			Description: "Set which security groups a server has. Requires IONOS_MCP_TOOL_SCOPE to include write. Single call. " +
				"This REPLACES the server's entire set of security groups: any group you omit is unassigned, and an empty list unassigns all of them, which removes the protection those groups provided. " +
				"To ADD a group, first read the current set with get_server at depth 2, then pass that set plus the new ID. Only the assignment changes — no security group is created or deleted.",
		},
		func(ctx context.Context, req *mcp.CallToolRequest, input tools.AssignServerSecurityGroupsInput) (*mcp.CallToolResult, any, error) {
			dcID := strings.TrimSpace(input.DatacenterID)
			serverID := strings.TrimSpace(input.ServerID)
			if dcID == "" {
				return tools.ErrorText("datacenter_id is required"), nil, nil
			}
			if serverID == "" {
				return tools.ErrorText("server_id is required"), nil, nil
			}
			ids, err := cleanIDs(input.SecurityGroupIDs)
			if err != nil {
				return tools.ErrorText(err.Error()), nil, nil
			}
			body := ionos.NewListOfIds(ids)
			assigned, _, apiErr := client.SecurityGroupsApi.DatacentersServersSecuritygroupsPut(ctx, dcID, serverID).Securitygroups(*body).Execute()
			return tools.ToResult(assigned, apiErr)
		})
}

func registerAssignNicSecurityGroups(server *mcp.Server, client *ionos.APIClient, scope tools.Scope) {
	tools.RegisterActionTool(server, scope,
		tools.Action{Verb: "assign_", Method: tools.MethodPut, Idempotent: true},
		&mcp.Tool{
			Name: "assign_nic_security_groups",
			Description: "Set which security groups a NIC has. Requires IONOS_MCP_TOOL_SCOPE to include write. Single call. " +
				"This REPLACES the NIC's entire set of security groups: any group you omit is unassigned, and an empty list unassigns all of them. " +
				"To ADD a group, first read the current set with get_nic at depth 2, then pass that set plus the new ID. Only the assignment changes — no security group is created or deleted.",
		},
		func(ctx context.Context, req *mcp.CallToolRequest, input tools.AssignNicSecurityGroupsInput) (*mcp.CallToolResult, any, error) {
			dcID := strings.TrimSpace(input.DatacenterID)
			serverID := strings.TrimSpace(input.ServerID)
			nicID := strings.TrimSpace(input.NicID)
			if dcID == "" {
				return tools.ErrorText("datacenter_id is required"), nil, nil
			}
			if serverID == "" {
				return tools.ErrorText("server_id is required"), nil, nil
			}
			if nicID == "" {
				return tools.ErrorText("nic_id is required"), nil, nil
			}
			ids, err := cleanIDs(input.SecurityGroupIDs)
			if err != nil {
				return tools.ErrorText(err.Error()), nil, nil
			}
			body := ionos.NewListOfIds(ids)
			assigned, _, apiErr := client.SecurityGroupsApi.DatacentersServersNicsSecuritygroupsPut(ctx, dcID, serverID, nicID).Securitygroups(*body).Execute()
			return tools.ToResult(assigned, apiErr)
		})
}

// cleanIDs trims the supplied IDs and rejects blank entries. A blank ID in a
// full-set replacement is almost certainly a mistake, and letting it through
// would either error opaquely at the API or unassign more than intended. A nil or
// empty list is allowed — that is the documented way to unassign everything —
// but is normalised to a non-nil slice so the body serialises as [] rather than
// null.
func cleanIDs(ids []string) ([]string, error) {
	out := make([]string, 0, len(ids))
	for i, id := range ids {
		trimmed := strings.TrimSpace(id)
		if trimmed == "" {
			return nil, fmt.Errorf("security_group_ids[%d] is empty; remove it, or pass an empty list to unassign all groups", i)
		}
		out = append(out, trimmed)
	}
	return out, nil
}
