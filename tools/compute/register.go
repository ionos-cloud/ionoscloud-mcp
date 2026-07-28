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
	RegisterFirewallRuleWriteTools(server, client, scope, confirm)
	RegisterIpBlockTools(server, client, scope)
	RegisterIpBlockWriteTools(server, client, scope, confirm)
	RegisterLoadBalancerTools(server, client, scope)
	RegisterLoadBalancerWriteTools(server, client, scope, confirm)
	RegisterNetworkLoadBalancerTools(server, client, scope)
	RegisterManagedLoadBalancerWriteTools(server, client, scope, confirm)
	RegisterForwardingRuleWriteTools(server, client, scope, confirm)
	RegisterApplicationLoadBalancerTools(server, client, scope)
	RegisterTargetGroupTools(server, client, scope)
	RegisterTargetGroupWriteTools(server, client, scope, confirm)
	RegisterNatGatewayTools(server, client, scope)
	RegisterNatGatewayWriteTools(server, client, scope, confirm)
	RegisterPrivateCrossConnectTools(server, client, scope)
	RegisterPccWriteTools(server, client, scope, confirm)
	RegisterSecurityGroupTools(server, client, scope)
	RegisterSecurityGroupWriteTools(server, client, scope, confirm)
	RegisterContractTools(server, client, scope)
	RegisterRequestTools(server, client, scope)
	RegisterTemplateTools(server, client, scope)
	RegisterImageTools(server, client, scope)
	RegisterLocationTools(server, client, scope)
	RegisterSnapshotTools(server, client, scope)
	RegisterSnapshotImageWriteTools(server, client, scope, confirm)
}
