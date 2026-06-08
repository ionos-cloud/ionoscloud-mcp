package compute

import (
	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterAll registers all compute engine tools with the MCP server.
func RegisterAll(server *mcp.Server, client *ionos.APIClient) {
	RegisterDatacenterTools(server, client)
	RegisterServerTools(server, client)
	RegisterVolumeTools(server, client)
	RegisterNicTools(server, client)
	RegisterLanTools(server, client)
	RegisterFirewallRuleTools(server, client)
	RegisterIpBlockTools(server, client)
	RegisterLoadBalancerTools(server, client)
	RegisterNetworkLoadBalancerTools(server, client)
	RegisterApplicationLoadBalancerTools(server, client)
	RegisterTargetGroupTools(server, client)
	RegisterNatGatewayTools(server, client)
	RegisterPrivateCrossConnectTools(server, client)
	RegisterSecurityGroupTools(server, client)
	RegisterContractTools(server, client)
	RegisterRequestTools(server, client)
	RegisterTemplateTools(server, client)
	RegisterImageTools(server, client)
	RegisterLocationTools(server, client)
	RegisterSnapshotTools(server, client)
	RegisterK8sTools(server, client)
}
