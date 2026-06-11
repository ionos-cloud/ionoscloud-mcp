package compute

import (
	"context"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterNicTools(server *mcp.Server, client *ionos.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_nics",
		Annotations: tools.ReadOnly,
		Description: "List all NICs attached to a specific server, with IPs, LAN assignment, and firewall state. Use list_lan_nics instead when starting from a network rather than a server.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ServerIDInput) (*mcp.CallToolResult, any, error) {
		nics, _, err := client.NetworkInterfacesApi.DatacentersServersNicsGet(ctx, input.DatacenterID, input.ServerID).Execute()
		return tools.ToResult(nics, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_nic",
		Annotations: tools.ReadOnly,
		Description: "Get details of a single NIC: IPs, MAC address, LAN ID, DHCP/IPv6 settings, and whether its firewall is active. Use list_nics to find NIC IDs; use list_firewall_rules for its firewall rules.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.NicIDInput) (*mcp.CallToolResult, any, error) {
		nic, _, err := client.NetworkInterfacesApi.DatacentersServersNicsFindById(ctx, input.DatacenterID, input.ServerID, input.NicID).Execute()
		return tools.ToResult(nic, err)
	})
}
