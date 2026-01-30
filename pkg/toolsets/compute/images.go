package compute

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ionos-cloud/ionoscloud-mcp/pkg/api"
)

func initImages() []api.ServerTool {
	return []api.ServerTool{
		{
			Tool: api.Tool{
				Name:        "list_images",
				Description: "List all available images (OS templates) in IONOS Cloud",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {},
					"required": []
				}`),
				Annotations: api.ReadOnly("List Images"),
			},
			Handler: listImages,
		},
		{
			Tool: api.Tool{
				Name:        "list_locations",
				Description: "List all available locations (regions) in IONOS Cloud",
				InputSchema: json.RawMessage(`{
					"type": "object",
					"properties": {},
					"required": []
				}`),
				Annotations: api.ReadOnly("List Locations"),
			},
			Handler: listLocations,
		},
	}
}

func listImages(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	images, _, err := params.Client.ImagesApi.ImagesGet(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}
	return api.MarshalResult(images, "images")
}

func listLocations(ctx context.Context, params api.ToolHandlerParams) (*api.ToolCallResult, error) {
	locations, _, err := params.Client.LocationsApi.LocationsGet(ctx).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to list locations: %w", err)
	}
	return api.MarshalResult(locations, "locations")
}
