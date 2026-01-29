package main

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	ionoscloud "github.com/ionos-cloud/sdk-go/v6"
)

// TestServer wraps the MCP server for testing
type TestServer struct {
	*Server
}

func newTestServer(t *testing.T) *TestServer {
	// Skip if no credentials are set
	if os.Getenv("IONOS_USERNAME") == "" && os.Getenv("IONOS_TOKEN") == "" {
		t.Skip("Skipping E2E test: IONOS_USERNAME/IONOS_PASSWORD or IONOS_TOKEN not set")
	}

	return &TestServer{Server: NewServer()}
}

// waitForState polls a resource until it reaches the expected state or times out
func waitForState(t *testing.T, check func() (string, error), timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		state, err := check()
		if err == nil && state == "AVAILABLE" {
			return
		}
		time.Sleep(5 * time.Second)
	}
	t.Log("Warning: state check timed out")
}

// =============================================================================
// Datacenter Tests
// =============================================================================

func TestDatacenterCRUD(t *testing.T) {
	ts := newTestServer(t)

	var dcID string

	// Create datacenter
	t.Run("CreateDatacenter", func(t *testing.T) {
		result, err := ts.executeTool("create_datacenter", map[string]interface{}{
			"name":        "test-dc-mcp",
			"location":    "de/fra",
			"description": "Test datacenter for MCP E2E tests",
		})
		if err != nil {
			t.Fatalf("Failed to create datacenter: %v", err)
		}

		var dc ionoscloud.Datacenter
		if err := json.Unmarshal([]byte(result), &dc); err != nil {
			t.Fatalf("Failed to unmarshal datacenter: %v", err)
		}

		if dc.Id == nil || *dc.Id == "" {
			t.Fatal("Datacenter ID is empty")
		}

		dcID = *dc.Id
		t.Logf("Created datacenter: %s", dcID)
	})

	if dcID == "" {
		t.Fatal("No datacenter ID from create step")
	}

	// Wait for datacenter to be available
	waitForState(t, func() (string, error) {
		result, err := ts.executeTool("get_datacenter", map[string]interface{}{
			"datacenter_id": dcID,
		})
		if err != nil {
			return "", err
		}
		var dc ionoscloud.Datacenter
		json.Unmarshal([]byte(result), &dc)
		if dc.Metadata != nil && dc.Metadata.State != nil {
			return *dc.Metadata.State, nil
		}
		return "", nil
	}, 5*time.Minute)

	// Update datacenter
	t.Run("UpdateDatacenter", func(t *testing.T) {
		result, err := ts.executeTool("update_datacenter", map[string]interface{}{
			"datacenter_id": dcID,
			"name":          "test-dc-mcp-updated",
			"description":   "Updated description",
		})
		if err != nil {
			t.Fatalf("Failed to update datacenter: %v", err)
		}

		var dc ionoscloud.Datacenter
		if err := json.Unmarshal([]byte(result), &dc); err != nil {
			t.Fatalf("Failed to unmarshal datacenter: %v", err)
		}

		if dc.Properties == nil || dc.Properties.Name == nil || *dc.Properties.Name != "test-dc-mcp-updated" {
			t.Fatal("Datacenter name was not updated")
		}

		t.Log("Updated datacenter successfully")
	})

	// List datacenters
	t.Run("ListDatacenters", func(t *testing.T) {
		result, err := ts.executeTool("list_datacenters", map[string]interface{}{})
		if err != nil {
			t.Fatalf("Failed to list datacenters: %v", err)
		}

		var dcs ionoscloud.Datacenters
		if err := json.Unmarshal([]byte(result), &dcs); err != nil {
			t.Fatalf("Failed to unmarshal datacenters: %v", err)
		}

		if dcs.Items == nil || len(*dcs.Items) == 0 {
			t.Fatal("No datacenters found")
		}

		t.Logf("Found %d datacenters", len(*dcs.Items))
	})

	// Get datacenter
	t.Run("GetDatacenter", func(t *testing.T) {
		result, err := ts.executeTool("get_datacenter", map[string]interface{}{
			"datacenter_id": dcID,
		})
		if err != nil {
			t.Fatalf("Failed to get datacenter: %v", err)
		}

		var dc ionoscloud.Datacenter
		if err := json.Unmarshal([]byte(result), &dc); err != nil {
			t.Fatalf("Failed to unmarshal datacenter: %v", err)
		}

		if dc.Id == nil || *dc.Id != dcID {
			t.Fatal("Wrong datacenter returned")
		}

		t.Log("Got datacenter successfully")
	})

	// Delete datacenter
	t.Run("DeleteDatacenter", func(t *testing.T) {
		result, err := ts.executeTool("delete_datacenter", map[string]interface{}{
			"datacenter_id": dcID,
		})
		if err != nil {
			t.Fatalf("Failed to delete datacenter: %v", err)
		}

		var status map[string]string
		if err := json.Unmarshal([]byte(result), &status); err != nil {
			t.Fatalf("Failed to unmarshal result: %v", err)
		}

		if status["status"] != "deleted" {
			t.Fatal("Delete status is not 'deleted'")
		}

		t.Log("Deleted datacenter successfully")
	})
}

// =============================================================================
// Server Tests
// =============================================================================

func TestServerCRUD(t *testing.T) {
	ts := newTestServer(t)

	// First, create a datacenter for the server
	result, err := ts.executeTool("create_datacenter", map[string]interface{}{
		"name":        "test-dc-server",
		"location":    "de/fra",
		"description": "Test datacenter for server tests",
	})
	if err != nil {
		t.Fatalf("Failed to create datacenter: %v", err)
	}

	var dc ionoscloud.Datacenter
	json.Unmarshal([]byte(result), &dc)
	dcID := *dc.Id

	// Cleanup at the end
	defer func() {
		// Wait a bit for resources to be released
		time.Sleep(10 * time.Second)
		ts.executeTool("delete_datacenter", map[string]interface{}{
			"datacenter_id": dcID,
		})
	}()

	// Wait for datacenter
	waitForState(t, func() (string, error) {
		result, _ := ts.executeTool("get_datacenter", map[string]interface{}{
			"datacenter_id": dcID,
		})
		var dc ionoscloud.Datacenter
		json.Unmarshal([]byte(result), &dc)
		if dc.Metadata != nil && dc.Metadata.State != nil {
			return *dc.Metadata.State, nil
		}
		return "", nil
	}, 5*time.Minute)

	var serverID string

	// Create server
	t.Run("CreateServer", func(t *testing.T) {
		result, err := ts.executeTool("create_server", map[string]interface{}{
			"datacenter_id": dcID,
			"name":          "test-server-mcp",
			"cores":         float64(1),
			"ram":           float64(1024),
		})
		if err != nil {
			t.Fatalf("Failed to create server: %v", err)
		}

		var server ionoscloud.Server
		if err := json.Unmarshal([]byte(result), &server); err != nil {
			t.Fatalf("Failed to unmarshal server: %v", err)
		}

		if server.Id == nil || *server.Id == "" {
			t.Fatal("Server ID is empty")
		}

		serverID = *server.Id
		t.Logf("Created server: %s", serverID)
	})

	if serverID == "" {
		t.Fatal("No server ID from create step")
	}

	// Wait for server
	waitForState(t, func() (string, error) {
		result, _ := ts.executeTool("get_server", map[string]interface{}{
			"datacenter_id": dcID,
			"server_id":     serverID,
		})
		var server ionoscloud.Server
		json.Unmarshal([]byte(result), &server)
		if server.Metadata != nil && server.Metadata.State != nil {
			return *server.Metadata.State, nil
		}
		return "", nil
	}, 5*time.Minute)

	// Update server
	t.Run("UpdateServer", func(t *testing.T) {
		result, err := ts.executeTool("update_server", map[string]interface{}{
			"datacenter_id": dcID,
			"server_id":     serverID,
			"name":          "test-server-mcp-updated",
			"cores":         float64(2),
		})
		if err != nil {
			t.Fatalf("Failed to update server: %v", err)
		}

		var server ionoscloud.Server
		if err := json.Unmarshal([]byte(result), &server); err != nil {
			t.Fatalf("Failed to unmarshal server: %v", err)
		}

		t.Log("Updated server successfully")
	})

	// List servers
	t.Run("ListServers", func(t *testing.T) {
		result, err := ts.executeTool("list_servers", map[string]interface{}{
			"datacenter_id": dcID,
		})
		if err != nil {
			t.Fatalf("Failed to list servers: %v", err)
		}

		var servers ionoscloud.Servers
		if err := json.Unmarshal([]byte(result), &servers); err != nil {
			t.Fatalf("Failed to unmarshal servers: %v", err)
		}

		if servers.Items == nil || len(*servers.Items) == 0 {
			t.Fatal("No servers found")
		}

		t.Logf("Found %d servers", len(*servers.Items))
	})

	// Get server
	t.Run("GetServer", func(t *testing.T) {
		result, err := ts.executeTool("get_server", map[string]interface{}{
			"datacenter_id": dcID,
			"server_id":     serverID,
		})
		if err != nil {
			t.Fatalf("Failed to get server: %v", err)
		}

		var server ionoscloud.Server
		if err := json.Unmarshal([]byte(result), &server); err != nil {
			t.Fatalf("Failed to unmarshal server: %v", err)
		}

		t.Log("Got server successfully")
	})

	// Stop server (it starts in INACTIVE state, so we may not need this)
	t.Run("StopServer", func(t *testing.T) {
		_, err := ts.executeTool("stop_server", map[string]interface{}{
			"datacenter_id": dcID,
			"server_id":     serverID,
		})
		// This may fail if server is already stopped, which is OK
		if err != nil {
			t.Logf("Stop server returned: %v (may be expected)", err)
		} else {
			t.Log("Stop server command sent")
		}
	})

	// Wait a bit
	time.Sleep(10 * time.Second)

	// Delete server
	t.Run("DeleteServer", func(t *testing.T) {
		result, err := ts.executeTool("delete_server", map[string]interface{}{
			"datacenter_id": dcID,
			"server_id":     serverID,
		})
		if err != nil {
			t.Fatalf("Failed to delete server: %v", err)
		}

		var status map[string]string
		if err := json.Unmarshal([]byte(result), &status); err != nil {
			t.Fatalf("Failed to unmarshal result: %v", err)
		}

		if status["status"] != "deleted" {
			t.Fatal("Delete status is not 'deleted'")
		}

		t.Log("Deleted server successfully")
	})
}

// =============================================================================
// Volume Tests
// =============================================================================

func TestVolumeCRUD(t *testing.T) {
	ts := newTestServer(t)

	// First, create a datacenter for the volume
	result, err := ts.executeTool("create_datacenter", map[string]interface{}{
		"name":        "test-dc-volume",
		"location":    "de/fra",
		"description": "Test datacenter for volume tests",
	})
	if err != nil {
		t.Fatalf("Failed to create datacenter: %v", err)
	}

	var dc ionoscloud.Datacenter
	json.Unmarshal([]byte(result), &dc)
	dcID := *dc.Id

	// Cleanup at the end
	defer func() {
		time.Sleep(10 * time.Second)
		ts.executeTool("delete_datacenter", map[string]interface{}{
			"datacenter_id": dcID,
		})
	}()

	// Wait for datacenter
	waitForState(t, func() (string, error) {
		result, _ := ts.executeTool("get_datacenter", map[string]interface{}{
			"datacenter_id": dcID,
		})
		var dc ionoscloud.Datacenter
		json.Unmarshal([]byte(result), &dc)
		if dc.Metadata != nil && dc.Metadata.State != nil {
			return *dc.Metadata.State, nil
		}
		return "", nil
	}, 5*time.Minute)

	var volumeID string

	// Create volume
	t.Run("CreateVolume", func(t *testing.T) {
		result, err := ts.executeTool("create_volume", map[string]interface{}{
			"datacenter_id": dcID,
			"name":          "test-volume-mcp",
			"size":          float64(10),
			"type":          "HDD",
			"licence_type":  "LINUX",
		})
		if err != nil {
			t.Fatalf("Failed to create volume: %v", err)
		}

		var volume ionoscloud.Volume
		if err := json.Unmarshal([]byte(result), &volume); err != nil {
			t.Fatalf("Failed to unmarshal volume: %v", err)
		}

		if volume.Id == nil || *volume.Id == "" {
			t.Fatal("Volume ID is empty")
		}

		volumeID = *volume.Id
		t.Logf("Created volume: %s", volumeID)
	})

	if volumeID == "" {
		t.Fatal("No volume ID from create step")
	}

	// Wait for volume
	waitForState(t, func() (string, error) {
		result, _ := ts.executeTool("get_volume", map[string]interface{}{
			"datacenter_id": dcID,
			"volume_id":     volumeID,
		})
		var volume ionoscloud.Volume
		json.Unmarshal([]byte(result), &volume)
		if volume.Metadata != nil && volume.Metadata.State != nil {
			return *volume.Metadata.State, nil
		}
		return "", nil
	}, 5*time.Minute)

	// Update volume
	t.Run("UpdateVolume", func(t *testing.T) {
		result, err := ts.executeTool("update_volume", map[string]interface{}{
			"datacenter_id": dcID,
			"volume_id":     volumeID,
			"name":          "test-volume-mcp-updated",
			"size":          float64(20), // Increase size
		})
		if err != nil {
			t.Fatalf("Failed to update volume: %v", err)
		}

		var volume ionoscloud.Volume
		if err := json.Unmarshal([]byte(result), &volume); err != nil {
			t.Fatalf("Failed to unmarshal volume: %v", err)
		}

		t.Log("Updated volume successfully")
	})

	// Wait for volume update
	waitForState(t, func() (string, error) {
		result, _ := ts.executeTool("get_volume", map[string]interface{}{
			"datacenter_id": dcID,
			"volume_id":     volumeID,
		})
		var volume ionoscloud.Volume
		json.Unmarshal([]byte(result), &volume)
		if volume.Metadata != nil && volume.Metadata.State != nil {
			return *volume.Metadata.State, nil
		}
		return "", nil
	}, 5*time.Minute)

	// List volumes
	t.Run("ListVolumes", func(t *testing.T) {
		result, err := ts.executeTool("list_volumes", map[string]interface{}{
			"datacenter_id": dcID,
		})
		if err != nil {
			t.Fatalf("Failed to list volumes: %v", err)
		}

		var volumes ionoscloud.Volumes
		if err := json.Unmarshal([]byte(result), &volumes); err != nil {
			t.Fatalf("Failed to unmarshal volumes: %v", err)
		}

		if volumes.Items == nil || len(*volumes.Items) == 0 {
			t.Fatal("No volumes found")
		}

		t.Logf("Found %d volumes", len(*volumes.Items))
	})

	// Get volume
	t.Run("GetVolume", func(t *testing.T) {
		result, err := ts.executeTool("get_volume", map[string]interface{}{
			"datacenter_id": dcID,
			"volume_id":     volumeID,
		})
		if err != nil {
			t.Fatalf("Failed to get volume: %v", err)
		}

		var volume ionoscloud.Volume
		if err := json.Unmarshal([]byte(result), &volume); err != nil {
			t.Fatalf("Failed to unmarshal volume: %v", err)
		}

		t.Log("Got volume successfully")
	})

	// Delete volume
	t.Run("DeleteVolume", func(t *testing.T) {
		result, err := ts.executeTool("delete_volume", map[string]interface{}{
			"datacenter_id": dcID,
			"volume_id":     volumeID,
		})
		if err != nil {
			t.Fatalf("Failed to delete volume: %v", err)
		}

		var status map[string]string
		if err := json.Unmarshal([]byte(result), &status); err != nil {
			t.Fatalf("Failed to unmarshal result: %v", err)
		}

		if status["status"] != "deleted" {
			t.Fatal("Delete status is not 'deleted'")
		}

		t.Log("Deleted volume successfully")
	})
}

// =============================================================================
// Snapshot Tests
// =============================================================================

func TestSnapshotCRUD(t *testing.T) {
	ts := newTestServer(t)

	// First, create a datacenter and volume
	result, err := ts.executeTool("create_datacenter", map[string]interface{}{
		"name":        "test-dc-snapshot",
		"location":    "de/fra",
		"description": "Test datacenter for snapshot tests",
	})
	if err != nil {
		t.Fatalf("Failed to create datacenter: %v", err)
	}

	var dc ionoscloud.Datacenter
	json.Unmarshal([]byte(result), &dc)
	dcID := *dc.Id

	// Cleanup at the end
	defer func() {
		time.Sleep(10 * time.Second)
		ts.executeTool("delete_datacenter", map[string]interface{}{
			"datacenter_id": dcID,
		})
	}()

	// Wait for datacenter
	waitForState(t, func() (string, error) {
		result, _ := ts.executeTool("get_datacenter", map[string]interface{}{
			"datacenter_id": dcID,
		})
		var dc ionoscloud.Datacenter
		json.Unmarshal([]byte(result), &dc)
		if dc.Metadata != nil && dc.Metadata.State != nil {
			return *dc.Metadata.State, nil
		}
		return "", nil
	}, 5*time.Minute)

	// Create volume
	result, err = ts.executeTool("create_volume", map[string]interface{}{
		"datacenter_id": dcID,
		"name":          "test-volume-snapshot",
		"size":          float64(10),
		"type":          "HDD",
		"licence_type":  "LINUX",
	})
	if err != nil {
		t.Fatalf("Failed to create volume: %v", err)
	}

	var volume ionoscloud.Volume
	json.Unmarshal([]byte(result), &volume)
	volumeID := *volume.Id

	// Wait for volume
	waitForState(t, func() (string, error) {
		result, _ := ts.executeTool("get_volume", map[string]interface{}{
			"datacenter_id": dcID,
			"volume_id":     volumeID,
		})
		var volume ionoscloud.Volume
		json.Unmarshal([]byte(result), &volume)
		if volume.Metadata != nil && volume.Metadata.State != nil {
			return *volume.Metadata.State, nil
		}
		return "", nil
	}, 5*time.Minute)

	var snapshotID string

	// Create snapshot
	t.Run("CreateSnapshot", func(t *testing.T) {
		result, err := ts.executeTool("create_snapshot", map[string]interface{}{
			"datacenter_id": dcID,
			"volume_id":     volumeID,
			"name":          "test-snapshot-mcp",
			"description":   "Test snapshot",
		})
		if err != nil {
			t.Fatalf("Failed to create snapshot: %v", err)
		}

		var snapshot ionoscloud.Snapshot
		if err := json.Unmarshal([]byte(result), &snapshot); err != nil {
			t.Fatalf("Failed to unmarshal snapshot: %v", err)
		}

		if snapshot.Id == nil || *snapshot.Id == "" {
			t.Fatal("Snapshot ID is empty")
		}

		snapshotID = *snapshot.Id
		t.Logf("Created snapshot: %s", snapshotID)
	})

	if snapshotID == "" {
		t.Fatal("No snapshot ID from create step")
	}

	// Wait for snapshot
	waitForState(t, func() (string, error) {
		result, _ := ts.executeTool("get_snapshot", map[string]interface{}{
			"snapshot_id": snapshotID,
		})
		var snapshot ionoscloud.Snapshot
		json.Unmarshal([]byte(result), &snapshot)
		if snapshot.Metadata != nil && snapshot.Metadata.State != nil {
			return *snapshot.Metadata.State, nil
		}
		return "", nil
	}, 10*time.Minute)

	// Update snapshot
	t.Run("UpdateSnapshot", func(t *testing.T) {
		result, err := ts.executeTool("update_snapshot", map[string]interface{}{
			"snapshot_id": snapshotID,
			"name":        "test-snapshot-mcp-updated",
			"description": "Updated description",
		})
		if err != nil {
			t.Fatalf("Failed to update snapshot: %v", err)
		}

		var snapshot ionoscloud.Snapshot
		if err := json.Unmarshal([]byte(result), &snapshot); err != nil {
			t.Fatalf("Failed to unmarshal snapshot: %v", err)
		}

		t.Log("Updated snapshot successfully")
	})

	// List snapshots
	t.Run("ListSnapshots", func(t *testing.T) {
		result, err := ts.executeTool("list_snapshots", map[string]interface{}{})
		if err != nil {
			t.Fatalf("Failed to list snapshots: %v", err)
		}

		var snapshots ionoscloud.Snapshots
		if err := json.Unmarshal([]byte(result), &snapshots); err != nil {
			t.Fatalf("Failed to unmarshal snapshots: %v", err)
		}

		if snapshots.Items == nil || len(*snapshots.Items) == 0 {
			t.Fatal("No snapshots found")
		}

		t.Logf("Found %d snapshots", len(*snapshots.Items))
	})

	// Get snapshot
	t.Run("GetSnapshot", func(t *testing.T) {
		result, err := ts.executeTool("get_snapshot", map[string]interface{}{
			"snapshot_id": snapshotID,
		})
		if err != nil {
			t.Fatalf("Failed to get snapshot: %v", err)
		}

		var snapshot ionoscloud.Snapshot
		if err := json.Unmarshal([]byte(result), &snapshot); err != nil {
			t.Fatalf("Failed to unmarshal snapshot: %v", err)
		}

		t.Log("Got snapshot successfully")
	})

	// Delete snapshot
	t.Run("DeleteSnapshot", func(t *testing.T) {
		result, err := ts.executeTool("delete_snapshot", map[string]interface{}{
			"snapshot_id": snapshotID,
		})
		if err != nil {
			t.Fatalf("Failed to delete snapshot: %v", err)
		}

		var status map[string]string
		if err := json.Unmarshal([]byte(result), &status); err != nil {
			t.Fatalf("Failed to unmarshal result: %v", err)
		}

		if status["status"] != "deleted" {
			t.Fatal("Delete status is not 'deleted'")
		}

		t.Log("Deleted snapshot successfully")
	})

	// Delete volume
	time.Sleep(10 * time.Second)
	ts.executeTool("delete_volume", map[string]interface{}{
		"datacenter_id": dcID,
		"volume_id":     volumeID,
	})
}

// =============================================================================
// Volume Attachment Tests
// =============================================================================

func TestVolumeAttachment(t *testing.T) {
	ts := newTestServer(t)

	// Create datacenter
	result, err := ts.executeTool("create_datacenter", map[string]interface{}{
		"name":        "test-dc-attach",
		"location":    "de/fra",
		"description": "Test datacenter for attachment tests",
	})
	if err != nil {
		t.Fatalf("Failed to create datacenter: %v", err)
	}

	var dc ionoscloud.Datacenter
	json.Unmarshal([]byte(result), &dc)
	dcID := *dc.Id

	// Cleanup
	defer func() {
		time.Sleep(10 * time.Second)
		ts.executeTool("delete_datacenter", map[string]interface{}{
			"datacenter_id": dcID,
		})
	}()

	// Wait for datacenter
	waitForState(t, func() (string, error) {
		result, _ := ts.executeTool("get_datacenter", map[string]interface{}{
			"datacenter_id": dcID,
		})
		var dc ionoscloud.Datacenter
		json.Unmarshal([]byte(result), &dc)
		if dc.Metadata != nil && dc.Metadata.State != nil {
			return *dc.Metadata.State, nil
		}
		return "", nil
	}, 5*time.Minute)

	// Create server
	result, err = ts.executeTool("create_server", map[string]interface{}{
		"datacenter_id": dcID,
		"name":          "test-server-attach",
		"cores":         float64(1),
		"ram":           float64(1024),
	})
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	var server ionoscloud.Server
	json.Unmarshal([]byte(result), &server)
	serverID := *server.Id

	// Wait for server
	waitForState(t, func() (string, error) {
		result, _ := ts.executeTool("get_server", map[string]interface{}{
			"datacenter_id": dcID,
			"server_id":     serverID,
		})
		var server ionoscloud.Server
		json.Unmarshal([]byte(result), &server)
		if server.Metadata != nil && server.Metadata.State != nil {
			return *server.Metadata.State, nil
		}
		return "", nil
	}, 5*time.Minute)

	// Create volume
	result, err = ts.executeTool("create_volume", map[string]interface{}{
		"datacenter_id": dcID,
		"name":          "test-volume-attach",
		"size":          float64(10),
		"type":          "HDD",
		"licence_type":  "LINUX",
	})
	if err != nil {
		t.Fatalf("Failed to create volume: %v", err)
	}

	var volume ionoscloud.Volume
	json.Unmarshal([]byte(result), &volume)
	volumeID := *volume.Id

	// Wait for volume
	waitForState(t, func() (string, error) {
		result, _ := ts.executeTool("get_volume", map[string]interface{}{
			"datacenter_id": dcID,
			"volume_id":     volumeID,
		})
		var volume ionoscloud.Volume
		json.Unmarshal([]byte(result), &volume)
		if volume.Metadata != nil && volume.Metadata.State != nil {
			return *volume.Metadata.State, nil
		}
		return "", nil
	}, 5*time.Minute)

	// Attach volume
	t.Run("AttachVolume", func(t *testing.T) {
		result, err := ts.executeTool("attach_volume", map[string]interface{}{
			"datacenter_id": dcID,
			"server_id":     serverID,
			"volume_id":     volumeID,
		})
		if err != nil {
			t.Fatalf("Failed to attach volume: %v", err)
		}

		var vol ionoscloud.Volume
		if err := json.Unmarshal([]byte(result), &vol); err != nil {
			t.Fatalf("Failed to unmarshal volume: %v", err)
		}

		t.Log("Attached volume successfully")
	})

	// Wait for attachment
	time.Sleep(30 * time.Second)

	// Detach volume
	t.Run("DetachVolume", func(t *testing.T) {
		result, err := ts.executeTool("detach_volume", map[string]interface{}{
			"datacenter_id": dcID,
			"server_id":     serverID,
			"volume_id":     volumeID,
		})
		if err != nil {
			t.Fatalf("Failed to detach volume: %v", err)
		}

		var status map[string]string
		if err := json.Unmarshal([]byte(result), &status); err != nil {
			t.Fatalf("Failed to unmarshal result: %v", err)
		}

		if status["status"] != "detached" {
			t.Fatal("Detach status is not 'detached'")
		}

		t.Log("Detached volume successfully")
	})

	// Cleanup
	time.Sleep(10 * time.Second)
	ts.executeTool("delete_volume", map[string]interface{}{
		"datacenter_id": dcID,
		"volume_id":     volumeID,
	})
	time.Sleep(10 * time.Second)
	ts.executeTool("delete_server", map[string]interface{}{
		"datacenter_id": dcID,
		"server_id":     serverID,
	})
}

// =============================================================================
// Unit Tests (no API calls)
// =============================================================================

func TestToolDefinitions(t *testing.T) {
	server := NewServer()

	// Check that all Priority 1 tools are registered
	expectedTools := []string{
		"list_datacenters", "get_datacenter", "create_datacenter", "update_datacenter", "delete_datacenter",
		"list_servers", "get_server", "create_server", "update_server", "delete_server",
		"start_server", "stop_server", "reboot_server",
		"list_volumes", "get_volume", "create_volume", "update_volume", "delete_volume",
		"attach_volume", "detach_volume",
		"list_snapshots", "get_snapshot", "create_snapshot", "update_snapshot", "delete_snapshot",
		"restore_snapshot",
	}

	toolMap := make(map[string]bool)
	for _, tool := range server.tools {
		toolMap[tool.Name] = true
	}

	for _, expected := range expectedTools {
		if !toolMap[expected] {
			t.Errorf("Tool %s is not registered", expected)
		}
	}
}

func TestExecuteToolUnknownTool(t *testing.T) {
	server := NewServer()
	server.ctx = context.Background()

	_, err := server.executeTool("unknown_tool", map[string]interface{}{})
	if err == nil {
		t.Error("Expected error for unknown tool")
	}
}

func TestExecuteToolMissingParameters(t *testing.T) {
	server := NewServer()
	server.ctx = context.Background()

	// Test missing datacenter_id for get_datacenter
	_, err := server.executeTool("get_datacenter", map[string]interface{}{})
	if err == nil {
		t.Error("Expected error for missing datacenter_id")
	}

	// Test missing parameters for create_server
	_, err = server.executeTool("create_server", map[string]interface{}{
		"datacenter_id": "test",
	})
	if err == nil {
		t.Error("Expected error for missing server parameters")
	}
}
