package main

import (
	"context"
	"encoding/json"
	"fmt"

	compute "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Input types for tool parameters.

type DatacenterIDInput struct {
	DatacenterID string `json:"datacenter_id" jsonschema:"the ID of the data center"`
}

type ServerIDInput struct {
	DatacenterID string `json:"datacenter_id" jsonschema:"the ID of the data center"`
	ServerID     string `json:"server_id" jsonschema:"the ID of the server"`
}

type VolumeIDInput struct {
	DatacenterID string `json:"datacenter_id" jsonschema:"the ID of the data center"`
	VolumeID     string `json:"volume_id" jsonschema:"the ID of the volume"`
}

type SnapshotIDInput struct {
	SnapshotID string `json:"snapshot_id" jsonschema:"the ID of the snapshot"`
}

// toResult marshals an API response into an MCP text result.
func toResult(data any, apiErr error) (*mcp.CallToolResult, any, error) {
	if apiErr != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: apiErr.Error()},
			},
			IsError: true,
		}, nil, nil
	}
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal response: %w", err)
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(bytes)},
		},
	}, nil, nil
}

func registerTools(server *mcp.Server, client *compute.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_datacenters",
		Description: "List all virtual data centers in your IONOS Cloud account",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		datacenters, _, err := client.DataCentersApi.DatacentersGet(ctx).Execute()
		return toResult(datacenters, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_datacenter",
		Description: "Get details of a specific virtual data center",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input DatacenterIDInput) (*mcp.CallToolResult, any, error) {
		datacenter, _, err := client.DataCentersApi.DatacentersFindById(ctx, input.DatacenterID).Execute()
		return toResult(datacenter, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_servers",
		Description: "List all servers in a data center",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input DatacenterIDInput) (*mcp.CallToolResult, any, error) {
		servers, _, err := client.ServersApi.DatacentersServersGet(ctx, input.DatacenterID).Execute()
		return toResult(servers, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_server",
		Description: "Get details of a specific server",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input ServerIDInput) (*mcp.CallToolResult, any, error) {
		s, _, err := client.ServersApi.DatacentersServersFindById(ctx, input.DatacenterID, input.ServerID).Execute()
		return toResult(s, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_volumes",
		Description: "List all volumes in a data center",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input DatacenterIDInput) (*mcp.CallToolResult, any, error) {
		volumes, _, err := client.VolumesApi.DatacentersVolumesGet(ctx, input.DatacenterID).Execute()
		return toResult(volumes, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_volume",
		Description: "Get details of a specific volume",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input VolumeIDInput) (*mcp.CallToolResult, any, error) {
		volume, _, err := client.VolumesApi.DatacentersVolumesFindById(ctx, input.DatacenterID, input.VolumeID).Execute()
		return toResult(volume, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_images",
		Description: "List all available images (OS templates) in IONOS Cloud",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		images, _, err := client.ImagesApi.ImagesGet(ctx).Execute()
		return toResult(images, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_locations",
		Description: "List all available locations (regions) in IONOS Cloud",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		locations, _, err := client.LocationsApi.LocationsGet(ctx).Execute()
		return toResult(locations, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_snapshots",
		Description: "List all snapshots in your IONOS Cloud account",
	}, func(ctx context.Context, req *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		snapshots, _, err := client.SnapshotsApi.SnapshotsGet(ctx).Execute()
		return toResult(snapshots, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_snapshot",
		Description: "Get details of a specific snapshot",
	}, func(ctx context.Context, req *mcp.CallToolRequest, input SnapshotIDInput) (*mcp.CallToolResult, any, error) {
		snapshot, _, err := client.SnapshotsApi.SnapshotsFindById(ctx, input.SnapshotID).Execute()
		return toResult(snapshot, err)
	})
}
