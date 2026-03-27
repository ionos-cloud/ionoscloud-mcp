package compute

import (
	"context"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func RegisterServerTools(server *mcp.Server, client *ionos.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_servers",
		Description: "List all servers in a data center",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.DatacenterIDInput) (*mcp.CallToolResult, any, error) {
		servers, _, err := client.ServersApi.DatacentersServersGet(ctx, input.DatacenterID).Execute()
		return tools.ToResult(servers, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_server",
		Description: "Get details of a specific server",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ServerIDInput) (*mcp.CallToolResult, any, error) {
		s, _, err := client.ServersApi.DatacentersServersFindById(ctx, input.DatacenterID, input.ServerID).Execute()
		return tools.ToResult(s, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_server_volumes",
		Description: "List all volumes attached to a specific server",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ServerIDInput) (*mcp.CallToolResult, any, error) {
		vols, _, err := client.ServersApi.DatacentersServersVolumesGet(ctx, input.DatacenterID, input.ServerID).Execute()
		return tools.ToResult(vols, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_server_cdroms",
		Description: "List all CD-ROMs attached to a specific server",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ServerIDInput) (*mcp.CallToolResult, any, error) {
		cdroms, _, err := client.ServersApi.DatacentersServersCdromsGet(ctx, input.DatacenterID, input.ServerID).Execute()
		return tools.ToResult(cdroms, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_server_gpus",
		Description: "List all GPUs attached to a specific server",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ServerIDInput) (*mcp.CallToolResult, any, error) {
		gpus, _, err := client.GraphicsProcessingUnitCardsApi.DatacentersServersGPUsGet(ctx, input.DatacenterID, input.ServerID).Execute()
		return tools.ToResult(gpus, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_server_gpu",
		Description: "Get details of a specific GPU attached to a server",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.GpuIDInput) (*mcp.CallToolResult, any, error) {
		gpu, _, err := client.GraphicsProcessingUnitCardsApi.DatacentersServersGPUsFindById(ctx, input.DatacenterID, input.ServerID, input.GpuID).Execute()
		return tools.ToResult(gpu, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_server_remote_console",
		Description: "Get the remote console URL for a specific server",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ServerIDInput) (*mcp.CallToolResult, any, error) {
		console, _, err := client.ServersApi.DatacentersServersRemoteConsoleGet(ctx, input.DatacenterID, input.ServerID).Execute()
		return tools.ToResult(console, err)
	})
}
