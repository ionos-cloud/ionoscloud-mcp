package compute

import (
	"context"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterLanTools(server *mcp.Server, client *ionos.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_lans",
		Annotations: tools.ReadOnly,
		Description: "List all LANs (virtual networks) in a data center, including whether each is public or private. Use list_datacenters first to find the datacenter ID.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.DatacenterIDInput) (*mcp.CallToolResult, any, error) {
		lans, _, err := client.LANsApi.DatacentersLansGet(ctx, input.DatacenterID).Execute()
		return tools.ToResult(lans, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_lan",
		Annotations: tools.ReadOnly,
		Description: "Get details of a single LAN: public/private flag, IPv6 settings, and IP failover configuration. Use list_lans to find LAN IDs; use list_lan_nics to see which NICs are attached to it.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.LanIDInput) (*mcp.CallToolResult, any, error) {
		lan, _, err := client.LANsApi.DatacentersLansFindById(ctx, input.DatacenterID, input.LanID).Execute()
		return tools.ToResult(lan, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_lan_nics",
		Annotations: tools.ReadOnly,
		Description: "List all NICs attached to a specific LAN — useful for finding which servers participate in a network. Use list_nics instead when starting from a server rather than a network.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.LanIDInput) (*mcp.CallToolResult, any, error) {
		nics, _, err := client.LANsApi.DatacentersLansNicsGet(ctx, input.DatacenterID, input.LanID).Execute()
		return tools.ToResult(nics, err)
	})
}
