package compute

import (
	"context"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterServerTools(server *mcp.Server, client *ionos.APIClient, scope tools.Scope) {
	tools.RegisterTool(server, scope, tools.MethodGet, &mcp.Tool{
		Name:        "list_servers",
		Description: "List all servers in a data center",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ListInDatacenterInput) (*mcp.CallToolResult, any, error) {
		depth := int32(1)
		if input.Depth != nil {
			depth = *input.Depth
		}
		r := client.ServersApi.DatacentersServersGet(ctx, input.DatacenterID).Depth(depth)
		for k, v := range input.Filters {
			r = r.Filter(k, v)
		}
		servers, _, err := r.Execute()
		return tools.ToResult(servers, err)
	})

	tools.RegisterTool(server, scope, tools.MethodGet, &mcp.Tool{
		Name:        "get_server",
		Description: "Get details of a specific server",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ServerIDInput) (*mcp.CallToolResult, any, error) {
		apiReq := client.ServersApi.DatacentersServersFindById(ctx, input.DatacenterID, input.ServerID)
		if input.Depth != nil {
			apiReq = apiReq.Depth(*input.Depth)
		}
		s, _, err := apiReq.Execute()
		return tools.ToResult(s, err)
	})

	tools.RegisterTool(server, scope, tools.MethodGet, &mcp.Tool{
		Name:        "list_server_volumes",
		Description: "List all volumes attached to a specific server",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ListInServerInput) (*mcp.CallToolResult, any, error) {
		depth := int32(1)
		if input.Depth != nil {
			depth = *input.Depth
		}
		r := client.ServersApi.DatacentersServersVolumesGet(ctx, input.DatacenterID, input.ServerID).Depth(depth)
		for k, v := range input.Filters {
			r = r.Filter(k, v)
		}
		vols, _, err := r.Execute()
		return tools.ToResult(vols, err)
	})

	tools.RegisterTool(server, scope, tools.MethodGet, &mcp.Tool{
		Name:        "list_server_cdroms",
		Description: "List all CD-ROMs attached to a specific server",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ListInServerInput) (*mcp.CallToolResult, any, error) {
		depth := int32(1)
		if input.Depth != nil {
			depth = *input.Depth
		}
		r := client.ServersApi.DatacentersServersCdromsGet(ctx, input.DatacenterID, input.ServerID).Depth(depth)
		for k, v := range input.Filters {
			r = r.Filter(k, v)
		}
		cdroms, _, err := r.Execute()
		return tools.ToResult(cdroms, err)
	})

	tools.RegisterTool(server, scope, tools.MethodGet, &mcp.Tool{
		Name:        "list_server_gpus",
		Description: "List all GPUs attached to a specific server",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ListInServerInput) (*mcp.CallToolResult, any, error) {
		depth := int32(1)
		if input.Depth != nil {
			depth = *input.Depth
		}
		r := client.GraphicsProcessingUnitCardsApi.DatacentersServersGPUsGet(ctx, input.DatacenterID, input.ServerID).Depth(depth)
		for k, v := range input.Filters {
			r = r.Filter(k, v)
		}
		gpus, _, err := r.Execute()
		return tools.ToResult(gpus, err)
	})

	tools.RegisterTool(server, scope, tools.MethodGet, &mcp.Tool{
		Name:        "get_server_gpu",
		Description: "Get details of a specific GPU attached to a server",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.GpuIDInput) (*mcp.CallToolResult, any, error) {
		apiReq := client.GraphicsProcessingUnitCardsApi.DatacentersServersGPUsFindById(ctx, input.DatacenterID, input.ServerID, input.GpuID)
		if input.Depth != nil {
			apiReq = apiReq.Depth(*input.Depth)
		}
		gpu, _, err := apiReq.Execute()
		return tools.ToResult(gpu, err)
	})

	tools.RegisterTool(server, scope, tools.MethodGet, &mcp.Tool{
		Name:        "get_server_remote_console",
		Description: "Get the remote console URL for a specific server",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input tools.ServerIDInput) (*mcp.CallToolResult, any, error) {
		apiReq := client.ServersApi.DatacentersServersRemoteConsoleGet(ctx, input.DatacenterID, input.ServerID)
		if input.Depth != nil {
			apiReq = apiReq.Depth(*input.Depth)
		}
		console, _, err := apiReq.Execute()
		return tools.ToResult(console, err)
	})
}
