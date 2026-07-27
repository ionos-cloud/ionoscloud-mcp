package compute

import (
	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

// RegisterAll registers all compute engine tools with the MCP server. Every
// resource threads scope so all tools register through tools.RegisterTool (or
// tools.RegisterActionTool): reads are always allowed but still get read-only
// annotations, while write, destructive and action tools register only if
// IONOS_MCP_TOOL_SCOPE opts their class in. confirm is the shared two-phase
// confirmation store used by create/delete and destructive actions.
func RegisterAll(server *mcp.Server, client *ionos.APIClient, scope tools.Scope, confirm *tools.ConfirmationStore) {
	RegisterDatacenterTools(server, client, scope)
	RegisterDatacenterWriteTools(server, client, scope, confirm)
	RegisterServerTools(server, client, scope)
	RegisterServerWriteTools(server, client, scope, confirm)
	RegisterServerActionTools(server, client, scope, confirm)
	RegisterSecurityGroupAssignTools(server, client, scope)
	RegisterVolumeTools(server, client, scope)
	RegisterVolumeWriteTools(server, client, scope, confirm)
	RegisterVolumeActionTools(server, client, scope, confirm)
	RegisterNicTools(server, client, scope)
	RegisterNicWriteTools(server, client, scope, confirm)
	RegisterLanTools(server, client, scope)
	RegisterLanWriteTools(server, client, scope, confirm)
	RegisterFirewallRuleTools(server, client, scope)
	RegisterIpBlockTools(server, client, scope)
	RegisterLoadBalancerTools(server, client, scope)
	RegisterNetworkLoadBalancerTools(server, client, scope)
	RegisterApplicationLoadBalancerTools(server, client, scope)
	RegisterTargetGroupTools(server, client, scope)
	RegisterNatGatewayTools(server, client, scope)
	RegisterPrivateCrossConnectTools(server, client, scope)
	RegisterSecurityGroupTools(server, client, scope)
	RegisterContractTools(server, client, scope)
	RegisterRequestTools(server, client, scope)
	RegisterTemplateTools(server, client, scope)
	RegisterImageTools(server, client, scope)
	RegisterLocationTools(server, client, scope)
	RegisterSnapshotTools(server, client, scope)
}
