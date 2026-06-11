package compute

import (
	"context"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterServerTools(server *mcp.Server, client *ionos.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_servers",
		Annotations: tools.ReadOnly,
		Description: "List all servers in a data center, with their state, cores, RAM, and IDs. Use list_datacenters first to find the datacenter ID; use get_server for one server's full configuration.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.DatacenterIDInput) (*mcp.CallToolResult, any, error) {
		servers, _, err := client.ServersApi.DatacentersServersGet(ctx, input.DatacenterID).Execute()
		return tools.ToResult(servers, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_server",
		Annotations: tools.ReadOnly,
		Description: "Get the full configuration of a single server: state, CPU family, cores, RAM, boot device, and attached volume/NIC references. Use list_servers to find server IDs; use list_server_volumes or list_nics for the attachment details themselves.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ServerIDInput) (*mcp.CallToolResult, any, error) {
		s, _, err := client.ServersApi.DatacentersServersFindById(ctx, input.DatacenterID, input.ServerID).Execute()
		return tools.ToResult(s, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_server_volumes",
		Annotations: tools.ReadOnly,
		Description: "List the storage volumes attached to a specific server, with size, type, and boot flags. Use list_volumes instead to see every volume in the datacenter, including unattached ones.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ServerIDInput) (*mcp.CallToolResult, any, error) {
		vols, _, err := client.ServersApi.DatacentersServersVolumesGet(ctx, input.DatacenterID, input.ServerID).Execute()
		return tools.ToResult(vols, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_server_cdroms",
		Annotations: tools.ReadOnly,
		Description: "List the CD-ROM images currently attached to a specific server. Use list_images to browse all attachable CD-ROM/ISO images.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ServerIDInput) (*mcp.CallToolResult, any, error) {
		cdroms, _, err := client.ServersApi.DatacentersServersCdromsGet(ctx, input.DatacenterID, input.ServerID).Execute()
		return tools.ToResult(cdroms, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_server_gpus",
		Annotations: tools.ReadOnly,
		Description: "List the GPUs attached to a specific server, with model and configuration. Use get_server_gpu for one GPU's full details.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ServerIDInput) (*mcp.CallToolResult, any, error) {
		gpus, _, err := client.GraphicsProcessingUnitCardsApi.DatacentersServersGPUsGet(ctx, input.DatacenterID, input.ServerID).Execute()
		return tools.ToResult(gpus, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_server_gpu",
		Annotations: tools.ReadOnly,
		Description: "Get details of a single GPU attached to a server. Use list_server_gpus first to find the GPU ID.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.GpuIDInput) (*mcp.CallToolResult, any, error) {
		gpu, _, err := client.GraphicsProcessingUnitCardsApi.DatacentersServersGPUsFindById(ctx, input.DatacenterID, input.ServerID, input.GpuID).Execute()
		return tools.ToResult(gpu, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_server_remote_console",
		Annotations: tools.ReadOnly,
		Description: "Get a browser URL for the interactive remote (VNC) console of a specific server. Use it when the user needs console access, e.g. to a server with no network connectivity; the URL is meant to be opened by the user.",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ServerIDInput) (*mcp.CallToolResult, any, error) {
		console, _, err := client.ServersApi.DatacentersServersRemoteConsoleGet(ctx, input.DatacenterID, input.ServerID).Execute()
		return tools.ToResult(console, err)
	})
}
