package main

import (
	"context"
	"encoding/json"
	"fmt"

	ionoscloud "github.com/ionos-cloud/sdk-go/v6"
)

func (s *Server) executeTool(name string, arguments map[string]interface{}) (string, error) {
	switch name {
	case "list_datacenters":
		return s.listDatacenters(s.client, s.ctx)
	case "get_datacenter":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		return s.getDatacenter(s.client, s.ctx, datacenterID)
	case "list_servers":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		return s.listServers(s.client, s.ctx, datacenterID)
	case "get_server":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		serverID, ok := arguments["server_id"].(string)
		if !ok {
			return "", fmt.Errorf("server_id is required")
		}
		return s.getServer(s.client, s.ctx, datacenterID, serverID)
	case "list_volumes":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		return s.listVolumes(s.client, s.ctx, datacenterID)
	case "get_volume":
		datacenterID, ok := arguments["datacenter_id"].(string)
		if !ok {
			return "", fmt.Errorf("datacenter_id is required")
		}
		volumeID, ok := arguments["volume_id"].(string)
		if !ok {
			return "", fmt.Errorf("volume_id is required")
		}
		return s.getVolume(s.client, s.ctx, datacenterID, volumeID)
	case "list_images":
		return s.listImages(s.client, s.ctx)
	case "list_locations":
		return s.listLocations(s.client, s.ctx)
	case "list_snapshots":
		return s.listSnapshots(s.client, s.ctx)
	case "get_snapshot":
		snapshotID, ok := arguments["snapshot_id"].(string)
		if !ok {
			return "", fmt.Errorf("snapshot_id is required")
		}
		return s.getSnapshot(s.client, s.ctx, snapshotID)
	default:
		return "", fmt.Errorf("unknown tool: %s", name)
	}
}

func (s *Server) listDatacenters(client *ionoscloud.APIClient, ctx context.Context) (string, error) {
	datacenters, _, err := client.DataCentersApi.DatacentersGet(ctx).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list datacenters: %w", err)
	}

	data, err := json.MarshalIndent(datacenters, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal datacenters: %w", err)
	}

	return string(data), nil
}

func (s *Server) getDatacenter(client *ionoscloud.APIClient, ctx context.Context, datacenterID string) (string, error) {
	datacenter, _, err := client.DataCentersApi.DatacentersFindById(ctx, datacenterID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get datacenter: %w", err)
	}

	data, err := json.MarshalIndent(datacenter, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal datacenter: %w", err)
	}

	return string(data), nil
}

func (s *Server) listServers(client *ionoscloud.APIClient, ctx context.Context, datacenterID string) (string, error) {
	servers, _, err := client.ServersApi.DatacentersServersGet(ctx, datacenterID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list servers: %w", err)
	}

	data, err := json.MarshalIndent(servers, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal servers: %w", err)
	}

	return string(data), nil
}

func (s *Server) getServer(client *ionoscloud.APIClient, ctx context.Context, datacenterID, serverID string) (string, error) {
	server, _, err := client.ServersApi.DatacentersServersFindById(ctx, datacenterID, serverID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get server: %w", err)
	}

	data, err := json.MarshalIndent(server, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal server: %w", err)
	}

	return string(data), nil
}

func (s *Server) listVolumes(client *ionoscloud.APIClient, ctx context.Context, datacenterID string) (string, error) {
	volumes, _, err := client.VolumesApi.DatacentersVolumesGet(ctx, datacenterID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list volumes: %w", err)
	}

	data, err := json.MarshalIndent(volumes, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal volumes: %w", err)
	}

	return string(data), nil
}

func (s *Server) getVolume(client *ionoscloud.APIClient, ctx context.Context, datacenterID, volumeID string) (string, error) {
	volume, _, err := client.VolumesApi.DatacentersVolumesFindById(ctx, datacenterID, volumeID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get volume: %w", err)
	}

	data, err := json.MarshalIndent(volume, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal volume: %w", err)
	}

	return string(data), nil
}

func (s *Server) listImages(client *ionoscloud.APIClient, ctx context.Context) (string, error) {
	images, _, err := client.ImagesApi.ImagesGet(ctx).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list images: %w", err)
	}

	data, err := json.MarshalIndent(images, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal images: %w", err)
	}

	return string(data), nil
}

func (s *Server) listLocations(client *ionoscloud.APIClient, ctx context.Context) (string, error) {
	locations, _, err := client.LocationsApi.LocationsGet(ctx).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list locations: %w", err)
	}

	data, err := json.MarshalIndent(locations, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal locations: %w", err)
	}

	return string(data), nil
}

func (s *Server) listSnapshots(client *ionoscloud.APIClient, ctx context.Context) (string, error) {
	snapshots, _, err := client.SnapshotsApi.SnapshotsGet(ctx).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to list snapshots: %w", err)
	}

	data, err := json.MarshalIndent(snapshots, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal snapshots: %w", err)
	}

	return string(data), nil
}

func (s *Server) getSnapshot(client *ionoscloud.APIClient, ctx context.Context, snapshotID string) (string, error) {
	snapshot, _, err := client.SnapshotsApi.SnapshotsFindById(ctx, snapshotID).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to get snapshot: %w", err)
	}

	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	return string(data), nil
}
