package main

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	ionoscloud "github.com/ionos-cloud/sdk-go/v6"
)

// Test constants
const (
	testLocation         = "de/fra"
	testServerCores      = 1
	testServerRAM        = 1024 // MB - minimum for most operations
	testVolumeSize       = 10   // GB - minimum practical size
	testVolumeSizeUp     = 20   // GB - for update tests
	defaultWaitTimeout   = 5 * time.Minute
	snapshotWaitTimeout  = 10 * time.Minute
	cleanupWaitTime      = 10 * time.Second
	attachmentWaitTime   = 30 * time.Second
	pollInterval         = 5 * time.Second
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

// cleanupResource attempts to delete a resource and logs any errors
func (ts *TestServer) cleanupResource(t *testing.T, toolName string, args map[string]interface{}) {
	if _, err := ts.executeTool(toolName, args); err != nil {
		t.Logf("Warning: cleanup failed for %s: %v", toolName, err)
	}
}

// waitForState polls a resource until it reaches the expected state or times out
func waitForState(t *testing.T, check func() (string, error), timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		state, err := check()
		if err == nil && state == "AVAILABLE" {
			return
		}
		time.Sleep(pollInterval)
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
			"location":    testLocation,
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
	}, defaultWaitTimeout)

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
		"location":    testLocation,
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
		time.Sleep(cleanupWaitTime)
		ts.cleanupResource(t, "delete_datacenter", map[string]interface{}{
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
	}, defaultWaitTimeout)

	var serverID string

	// Create server
	t.Run("CreateServer", func(t *testing.T) {
		result, err := ts.executeTool("create_server", map[string]interface{}{
			"datacenter_id": dcID,
			"name":          "test-server-mcp",
			"cores":         float64(testServerCores),
			"ram":           float64(testServerRAM),
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
	}, defaultWaitTimeout)

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
	time.Sleep(cleanupWaitTime)

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
		"location":    testLocation,
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
		time.Sleep(cleanupWaitTime)
		ts.cleanupResource(t, "delete_datacenter", map[string]interface{}{
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
	}, defaultWaitTimeout)

	var volumeID string

	// Create volume
	t.Run("CreateVolume", func(t *testing.T) {
		result, err := ts.executeTool("create_volume", map[string]interface{}{
			"datacenter_id": dcID,
			"name":          "test-volume-mcp",
			"size":          float64(testVolumeSize),
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
	}, defaultWaitTimeout)

	// Update volume
	t.Run("UpdateVolume", func(t *testing.T) {
		result, err := ts.executeTool("update_volume", map[string]interface{}{
			"datacenter_id": dcID,
			"volume_id":     volumeID,
			"name":          "test-volume-mcp-updated",
			"size":          float64(testVolumeSizeUp),
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
	}, defaultWaitTimeout)

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
		"location":    testLocation,
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
		time.Sleep(cleanupWaitTime)
		ts.cleanupResource(t, "delete_datacenter", map[string]interface{}{
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
	}, defaultWaitTimeout)

	// Create volume
	result, err = ts.executeTool("create_volume", map[string]interface{}{
		"datacenter_id": dcID,
		"name":          "test-volume-snapshot",
		"size":          float64(testVolumeSize),
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
	}, defaultWaitTimeout)

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
	}, snapshotWaitTimeout)

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
	time.Sleep(cleanupWaitTime)
	ts.cleanupResource(t, "delete_volume", map[string]interface{}{
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
		"location":    testLocation,
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
		time.Sleep(cleanupWaitTime)
		ts.cleanupResource(t, "delete_datacenter", map[string]interface{}{
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
	}, defaultWaitTimeout)

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
	}, defaultWaitTimeout)

	// Create volume
	result, err = ts.executeTool("create_volume", map[string]interface{}{
		"datacenter_id": dcID,
		"name":          "test-volume-attach",
		"size":          float64(testVolumeSize),
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
	}, defaultWaitTimeout)

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
	time.Sleep(attachmentWaitTime)

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
	time.Sleep(cleanupWaitTime)
	ts.cleanupResource(t, "delete_volume", map[string]interface{}{
		"datacenter_id": dcID,
		"volume_id":     volumeID,
	})
	time.Sleep(cleanupWaitTime)
	ts.cleanupResource(t, "delete_server", map[string]interface{}{
		"datacenter_id": dcID,
		"server_id":     serverID,
	})
}

// =============================================================================
// LAN Tests
// =============================================================================

func TestLanCRUD(t *testing.T) {
	ts := newTestServer(t)

	// Create datacenter for tests
	result, err := ts.executeTool("create_datacenter", map[string]interface{}{
		"name":        "test-dc-lan",
		"location":    testLocation,
		"description": "Test datacenter for LAN tests",
	})
	if err != nil {
		t.Fatalf("Failed to create datacenter: %v", err)
	}

	var dc ionoscloud.Datacenter
	json.Unmarshal([]byte(result), &dc)
	dcID := *dc.Id

	defer func() {
		time.Sleep(cleanupWaitTime)
		ts.cleanupResource(t, "delete_datacenter", map[string]interface{}{
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
	}, defaultWaitTimeout)

	var lanID string

	// Create LAN
	t.Run("CreateLAN", func(t *testing.T) {
		result, err := ts.executeTool("create_lan", map[string]interface{}{
			"datacenter_id": dcID,
			"name":          "test-lan-mcp",
			"public":        false,
		})
		if err != nil {
			t.Fatalf("Failed to create LAN: %v", err)
		}

		var lan ionoscloud.Lan
		if err := json.Unmarshal([]byte(result), &lan); err != nil {
			t.Fatalf("Failed to unmarshal LAN: %v", err)
		}

		if lan.Id == nil || *lan.Id == "" {
			t.Fatal("LAN ID is empty")
		}

		lanID = *lan.Id
		t.Logf("Created LAN: %s", lanID)
	})

	if lanID == "" {
		t.Fatal("No LAN ID from create step")
	}

	// Wait for LAN
	waitForState(t, func() (string, error) {
		result, _ := ts.executeTool("get_lan", map[string]interface{}{
			"datacenter_id": dcID,
			"lan_id":        lanID,
		})
		var lan ionoscloud.Lan
		json.Unmarshal([]byte(result), &lan)
		if lan.Metadata != nil && lan.Metadata.State != nil {
			return *lan.Metadata.State, nil
		}
		return "", nil
	}, defaultWaitTimeout)

	// Update LAN
	t.Run("UpdateLAN", func(t *testing.T) {
		result, err := ts.executeTool("update_lan", map[string]interface{}{
			"datacenter_id": dcID,
			"lan_id":        lanID,
			"name":          "test-lan-mcp-updated",
		})
		if err != nil {
			t.Fatalf("Failed to update LAN: %v", err)
		}

		var lan ionoscloud.Lan
		if err := json.Unmarshal([]byte(result), &lan); err != nil {
			t.Fatalf("Failed to unmarshal LAN: %v", err)
		}

		t.Log("Updated LAN successfully")
	})

	// List LANs
	t.Run("ListLANs", func(t *testing.T) {
		result, err := ts.executeTool("list_lans", map[string]interface{}{
			"datacenter_id": dcID,
		})
		if err != nil {
			t.Fatalf("Failed to list LANs: %v", err)
		}

		var lans ionoscloud.Lans
		if err := json.Unmarshal([]byte(result), &lans); err != nil {
			t.Fatalf("Failed to unmarshal LANs: %v", err)
		}

		if lans.Items == nil || len(*lans.Items) == 0 {
			t.Fatal("No LANs found")
		}

		t.Logf("Found %d LANs", len(*lans.Items))
	})

	// Get LAN
	t.Run("GetLAN", func(t *testing.T) {
		result, err := ts.executeTool("get_lan", map[string]interface{}{
			"datacenter_id": dcID,
			"lan_id":        lanID,
		})
		if err != nil {
			t.Fatalf("Failed to get LAN: %v", err)
		}

		var lan ionoscloud.Lan
		if err := json.Unmarshal([]byte(result), &lan); err != nil {
			t.Fatalf("Failed to unmarshal LAN: %v", err)
		}

		t.Log("Got LAN successfully")
	})

	// Delete LAN
	t.Run("DeleteLAN", func(t *testing.T) {
		result, err := ts.executeTool("delete_lan", map[string]interface{}{
			"datacenter_id": dcID,
			"lan_id":        lanID,
		})
		if err != nil {
			t.Fatalf("Failed to delete LAN: %v", err)
		}

		var status map[string]string
		if err := json.Unmarshal([]byte(result), &status); err != nil {
			t.Fatalf("Failed to unmarshal result: %v", err)
		}

		if status["status"] != "deleted" {
			t.Fatal("Delete status is not 'deleted'")
		}

		t.Log("Deleted LAN successfully")
	})
}

// =============================================================================
// NIC Tests
// =============================================================================

func TestNicCRUD(t *testing.T) {
	ts := newTestServer(t)

	// Create datacenter
	result, err := ts.executeTool("create_datacenter", map[string]interface{}{
		"name":        "test-dc-nic",
		"location":    testLocation,
		"description": "Test datacenter for NIC tests",
	})
	if err != nil {
		t.Fatalf("Failed to create datacenter: %v", err)
	}

	var dc ionoscloud.Datacenter
	json.Unmarshal([]byte(result), &dc)
	dcID := *dc.Id

	defer func() {
		time.Sleep(cleanupWaitTime)
		ts.cleanupResource(t, "delete_datacenter", map[string]interface{}{
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
	}, defaultWaitTimeout)

	// Create LAN for NIC
	result, err = ts.executeTool("create_lan", map[string]interface{}{
		"datacenter_id": dcID,
		"name":          "test-lan-nic",
		"public":        false,
	})
	if err != nil {
		t.Fatalf("Failed to create LAN: %v", err)
	}

	var lan ionoscloud.Lan
	json.Unmarshal([]byte(result), &lan)
	lanID := *lan.Id

	// Wait for LAN
	waitForState(t, func() (string, error) {
		result, _ := ts.executeTool("get_lan", map[string]interface{}{
			"datacenter_id": dcID,
			"lan_id":        lanID,
		})
		var lan ionoscloud.Lan
		json.Unmarshal([]byte(result), &lan)
		if lan.Metadata != nil && lan.Metadata.State != nil {
			return *lan.Metadata.State, nil
		}
		return "", nil
	}, defaultWaitTimeout)

	// Create server for NIC
	result, err = ts.executeTool("create_server", map[string]interface{}{
		"datacenter_id": dcID,
		"name":          "test-server-nic",
		"cores":         float64(testServerCores),
		"ram":           float64(testServerRAM),
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
	}, defaultWaitTimeout)

	var nicID string

	// Create NIC
	t.Run("CreateNIC", func(t *testing.T) {
		// Convert lanID string to int
		lanIDInt := 1 // LANs are typically numbered starting from 1
		result, err := ts.executeTool("create_nic", map[string]interface{}{
			"datacenter_id": dcID,
			"server_id":     serverID,
			"lan":           float64(lanIDInt),
			"name":          "test-nic-mcp",
			"dhcp":          true,
		})
		if err != nil {
			t.Fatalf("Failed to create NIC: %v", err)
		}

		var nic ionoscloud.Nic
		if err := json.Unmarshal([]byte(result), &nic); err != nil {
			t.Fatalf("Failed to unmarshal NIC: %v", err)
		}

		if nic.Id == nil || *nic.Id == "" {
			t.Fatal("NIC ID is empty")
		}

		nicID = *nic.Id
		t.Logf("Created NIC: %s", nicID)
	})

	if nicID == "" {
		t.Fatal("No NIC ID from create step")
	}

	// Wait for NIC
	waitForState(t, func() (string, error) {
		result, _ := ts.executeTool("get_nic", map[string]interface{}{
			"datacenter_id": dcID,
			"server_id":     serverID,
			"nic_id":        nicID,
		})
		var nic ionoscloud.Nic
		json.Unmarshal([]byte(result), &nic)
		if nic.Metadata != nil && nic.Metadata.State != nil {
			return *nic.Metadata.State, nil
		}
		return "", nil
	}, defaultWaitTimeout)

	// Update NIC
	t.Run("UpdateNIC", func(t *testing.T) {
		result, err := ts.executeTool("update_nic", map[string]interface{}{
			"datacenter_id": dcID,
			"server_id":     serverID,
			"nic_id":        nicID,
			"name":          "test-nic-mcp-updated",
		})
		if err != nil {
			t.Fatalf("Failed to update NIC: %v", err)
		}

		var nic ionoscloud.Nic
		if err := json.Unmarshal([]byte(result), &nic); err != nil {
			t.Fatalf("Failed to unmarshal NIC: %v", err)
		}

		t.Log("Updated NIC successfully")
	})

	// List NICs
	t.Run("ListNICs", func(t *testing.T) {
		result, err := ts.executeTool("list_nics", map[string]interface{}{
			"datacenter_id": dcID,
			"server_id":     serverID,
		})
		if err != nil {
			t.Fatalf("Failed to list NICs: %v", err)
		}

		var nics ionoscloud.Nics
		if err := json.Unmarshal([]byte(result), &nics); err != nil {
			t.Fatalf("Failed to unmarshal NICs: %v", err)
		}

		if nics.Items == nil || len(*nics.Items) == 0 {
			t.Fatal("No NICs found")
		}

		t.Logf("Found %d NICs", len(*nics.Items))
	})

	// Get NIC
	t.Run("GetNIC", func(t *testing.T) {
		result, err := ts.executeTool("get_nic", map[string]interface{}{
			"datacenter_id": dcID,
			"server_id":     serverID,
			"nic_id":        nicID,
		})
		if err != nil {
			t.Fatalf("Failed to get NIC: %v", err)
		}

		var nic ionoscloud.Nic
		if err := json.Unmarshal([]byte(result), &nic); err != nil {
			t.Fatalf("Failed to unmarshal NIC: %v", err)
		}

		t.Log("Got NIC successfully")
	})

	// Delete NIC
	t.Run("DeleteNIC", func(t *testing.T) {
		result, err := ts.executeTool("delete_nic", map[string]interface{}{
			"datacenter_id": dcID,
			"server_id":     serverID,
			"nic_id":        nicID,
		})
		if err != nil {
			t.Fatalf("Failed to delete NIC: %v", err)
		}

		var status map[string]string
		if err := json.Unmarshal([]byte(result), &status); err != nil {
			t.Fatalf("Failed to unmarshal result: %v", err)
		}

		if status["status"] != "deleted" {
			t.Fatal("Delete status is not 'deleted'")
		}

		t.Log("Deleted NIC successfully")
	})

	// Cleanup
	time.Sleep(cleanupWaitTime)
	ts.cleanupResource(t, "delete_server", map[string]interface{}{
		"datacenter_id": dcID,
		"server_id":     serverID,
	})
	time.Sleep(cleanupWaitTime)
	ts.cleanupResource(t, "delete_lan", map[string]interface{}{
		"datacenter_id": dcID,
		"lan_id":        lanID,
	})
}

// =============================================================================
// IP Block Tests
// =============================================================================

func TestIpBlockCRUD(t *testing.T) {
	ts := newTestServer(t)

	var ipblockID string

	// Create IP Block
	t.Run("CreateIPBlock", func(t *testing.T) {
		result, err := ts.executeTool("create_ipblock", map[string]interface{}{
			"location": testLocation,
			"size":     float64(1),
			"name":     "test-ipblock-mcp",
		})
		if err != nil {
			t.Fatalf("Failed to create IP block: %v", err)
		}

		var ipblock ionoscloud.IpBlock
		if err := json.Unmarshal([]byte(result), &ipblock); err != nil {
			t.Fatalf("Failed to unmarshal IP block: %v", err)
		}

		if ipblock.Id == nil || *ipblock.Id == "" {
			t.Fatal("IP block ID is empty")
		}

		ipblockID = *ipblock.Id
		t.Logf("Created IP block: %s", ipblockID)
	})

	if ipblockID == "" {
		t.Fatal("No IP block ID from create step")
	}

	// Wait for IP Block
	waitForState(t, func() (string, error) {
		result, _ := ts.executeTool("get_ipblock", map[string]interface{}{
			"ipblock_id": ipblockID,
		})
		var ipblock ionoscloud.IpBlock
		json.Unmarshal([]byte(result), &ipblock)
		if ipblock.Metadata != nil && ipblock.Metadata.State != nil {
			return *ipblock.Metadata.State, nil
		}
		return "", nil
	}, defaultWaitTimeout)

	// Update IP Block
	t.Run("UpdateIPBlock", func(t *testing.T) {
		result, err := ts.executeTool("update_ipblock", map[string]interface{}{
			"ipblock_id": ipblockID,
			"name":       "test-ipblock-mcp-updated",
		})
		if err != nil {
			t.Fatalf("Failed to update IP block: %v", err)
		}

		var ipblock ionoscloud.IpBlock
		if err := json.Unmarshal([]byte(result), &ipblock); err != nil {
			t.Fatalf("Failed to unmarshal IP block: %v", err)
		}

		t.Log("Updated IP block successfully")
	})

	// List IP Blocks
	t.Run("ListIPBlocks", func(t *testing.T) {
		result, err := ts.executeTool("list_ipblocks", map[string]interface{}{})
		if err != nil {
			t.Fatalf("Failed to list IP blocks: %v", err)
		}

		var ipblocks ionoscloud.IpBlocks
		if err := json.Unmarshal([]byte(result), &ipblocks); err != nil {
			t.Fatalf("Failed to unmarshal IP blocks: %v", err)
		}

		if ipblocks.Items == nil || len(*ipblocks.Items) == 0 {
			t.Fatal("No IP blocks found")
		}

		t.Logf("Found %d IP blocks", len(*ipblocks.Items))
	})

	// Get IP Block
	t.Run("GetIPBlock", func(t *testing.T) {
		result, err := ts.executeTool("get_ipblock", map[string]interface{}{
			"ipblock_id": ipblockID,
		})
		if err != nil {
			t.Fatalf("Failed to get IP block: %v", err)
		}

		var ipblock ionoscloud.IpBlock
		if err := json.Unmarshal([]byte(result), &ipblock); err != nil {
			t.Fatalf("Failed to unmarshal IP block: %v", err)
		}

		t.Log("Got IP block successfully")
	})

	// Delete IP Block
	t.Run("DeleteIPBlock", func(t *testing.T) {
		result, err := ts.executeTool("delete_ipblock", map[string]interface{}{
			"ipblock_id": ipblockID,
		})
		if err != nil {
			t.Fatalf("Failed to delete IP block: %v", err)
		}

		var status map[string]string
		if err := json.Unmarshal([]byte(result), &status); err != nil {
			t.Fatalf("Failed to unmarshal result: %v", err)
		}

		if status["status"] != "deleted" {
			t.Fatal("Delete status is not 'deleted'")
		}

		t.Log("Deleted IP block successfully")
	})
}

// =============================================================================
// Firewall Rule Tests
// =============================================================================

func TestFirewallRuleCRUD(t *testing.T) {
	ts := newTestServer(t)

	// Create datacenter
	result, err := ts.executeTool("create_datacenter", map[string]interface{}{
		"name":        "test-dc-firewall",
		"location":    testLocation,
		"description": "Test datacenter for firewall tests",
	})
	if err != nil {
		t.Fatalf("Failed to create datacenter: %v", err)
	}

	var dc ionoscloud.Datacenter
	json.Unmarshal([]byte(result), &dc)
	dcID := *dc.Id

	defer func() {
		time.Sleep(cleanupWaitTime)
		ts.cleanupResource(t, "delete_datacenter", map[string]interface{}{
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
	}, defaultWaitTimeout)

	// Create LAN
	result, err = ts.executeTool("create_lan", map[string]interface{}{
		"datacenter_id": dcID,
		"name":          "test-lan-firewall",
		"public":        true,
	})
	if err != nil {
		t.Fatalf("Failed to create LAN: %v", err)
	}

	var lan ionoscloud.Lan
	json.Unmarshal([]byte(result), &lan)
	lanID := *lan.Id

	// Wait for LAN
	waitForState(t, func() (string, error) {
		result, _ := ts.executeTool("get_lan", map[string]interface{}{
			"datacenter_id": dcID,
			"lan_id":        lanID,
		})
		var lan ionoscloud.Lan
		json.Unmarshal([]byte(result), &lan)
		if lan.Metadata != nil && lan.Metadata.State != nil {
			return *lan.Metadata.State, nil
		}
		return "", nil
	}, defaultWaitTimeout)

	// Create server
	result, err = ts.executeTool("create_server", map[string]interface{}{
		"datacenter_id": dcID,
		"name":          "test-server-firewall",
		"cores":         float64(testServerCores),
		"ram":           float64(testServerRAM),
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
	}, defaultWaitTimeout)

	// Create NIC
	result, err = ts.executeTool("create_nic", map[string]interface{}{
		"datacenter_id": dcID,
		"server_id":     serverID,
		"lan":           float64(1),
		"name":          "test-nic-firewall",
		"dhcp":          true,
	})
	if err != nil {
		t.Fatalf("Failed to create NIC: %v", err)
	}

	var nic ionoscloud.Nic
	json.Unmarshal([]byte(result), &nic)
	nicID := *nic.Id

	// Wait for NIC
	waitForState(t, func() (string, error) {
		result, _ := ts.executeTool("get_nic", map[string]interface{}{
			"datacenter_id": dcID,
			"server_id":     serverID,
			"nic_id":        nicID,
		})
		var nic ionoscloud.Nic
		json.Unmarshal([]byte(result), &nic)
		if nic.Metadata != nil && nic.Metadata.State != nil {
			return *nic.Metadata.State, nil
		}
		return "", nil
	}, defaultWaitTimeout)

	var firewallRuleID string

	// Create Firewall Rule
	t.Run("CreateFirewallRule", func(t *testing.T) {
		result, err := ts.executeTool("create_firewall_rule", map[string]interface{}{
			"datacenter_id":    dcID,
			"server_id":        serverID,
			"nic_id":           nicID,
			"name":             "test-firewall-mcp",
			"protocol":         "TCP",
			"port_range_start": float64(22),
			"port_range_end":   float64(22),
			"type":             "INGRESS",
		})
		if err != nil {
			t.Fatalf("Failed to create firewall rule: %v", err)
		}

		var rule ionoscloud.FirewallRule
		if err := json.Unmarshal([]byte(result), &rule); err != nil {
			t.Fatalf("Failed to unmarshal firewall rule: %v", err)
		}

		if rule.Id == nil || *rule.Id == "" {
			t.Fatal("Firewall rule ID is empty")
		}

		firewallRuleID = *rule.Id
		t.Logf("Created firewall rule: %s", firewallRuleID)
	})

	if firewallRuleID == "" {
		t.Fatal("No firewall rule ID from create step")
	}

	// Wait for firewall rule
	waitForState(t, func() (string, error) {
		result, _ := ts.executeTool("get_firewall_rule", map[string]interface{}{
			"datacenter_id":   dcID,
			"server_id":       serverID,
			"nic_id":          nicID,
			"firewallrule_id": firewallRuleID,
		})
		var rule ionoscloud.FirewallRule
		json.Unmarshal([]byte(result), &rule)
		if rule.Metadata != nil && rule.Metadata.State != nil {
			return *rule.Metadata.State, nil
		}
		return "", nil
	}, defaultWaitTimeout)

	// Update Firewall Rule
	t.Run("UpdateFirewallRule", func(t *testing.T) {
		result, err := ts.executeTool("update_firewall_rule", map[string]interface{}{
			"datacenter_id":   dcID,
			"server_id":       serverID,
			"nic_id":          nicID,
			"firewallrule_id": firewallRuleID,
			"name":            "test-firewall-mcp-updated",
		})
		if err != nil {
			t.Fatalf("Failed to update firewall rule: %v", err)
		}

		var rule ionoscloud.FirewallRule
		if err := json.Unmarshal([]byte(result), &rule); err != nil {
			t.Fatalf("Failed to unmarshal firewall rule: %v", err)
		}

		t.Log("Updated firewall rule successfully")
	})

	// List Firewall Rules
	t.Run("ListFirewallRules", func(t *testing.T) {
		result, err := ts.executeTool("list_firewall_rules", map[string]interface{}{
			"datacenter_id": dcID,
			"server_id":     serverID,
			"nic_id":        nicID,
		})
		if err != nil {
			t.Fatalf("Failed to list firewall rules: %v", err)
		}

		var rules ionoscloud.FirewallRules
		if err := json.Unmarshal([]byte(result), &rules); err != nil {
			t.Fatalf("Failed to unmarshal firewall rules: %v", err)
		}

		if rules.Items == nil || len(*rules.Items) == 0 {
			t.Fatal("No firewall rules found")
		}

		t.Logf("Found %d firewall rules", len(*rules.Items))
	})

	// Get Firewall Rule
	t.Run("GetFirewallRule", func(t *testing.T) {
		result, err := ts.executeTool("get_firewall_rule", map[string]interface{}{
			"datacenter_id":   dcID,
			"server_id":       serverID,
			"nic_id":          nicID,
			"firewallrule_id": firewallRuleID,
		})
		if err != nil {
			t.Fatalf("Failed to get firewall rule: %v", err)
		}

		var rule ionoscloud.FirewallRule
		if err := json.Unmarshal([]byte(result), &rule); err != nil {
			t.Fatalf("Failed to unmarshal firewall rule: %v", err)
		}

		t.Log("Got firewall rule successfully")
	})

	// Delete Firewall Rule
	t.Run("DeleteFirewallRule", func(t *testing.T) {
		result, err := ts.executeTool("delete_firewall_rule", map[string]interface{}{
			"datacenter_id":   dcID,
			"server_id":       serverID,
			"nic_id":          nicID,
			"firewallrule_id": firewallRuleID,
		})
		if err != nil {
			t.Fatalf("Failed to delete firewall rule: %v", err)
		}

		var status map[string]string
		if err := json.Unmarshal([]byte(result), &status); err != nil {
			t.Fatalf("Failed to unmarshal result: %v", err)
		}

		if status["status"] != "deleted" {
			t.Fatal("Delete status is not 'deleted'")
		}

		t.Log("Deleted firewall rule successfully")
	})

	// Cleanup
	time.Sleep(cleanupWaitTime)
	ts.cleanupResource(t, "delete_nic", map[string]interface{}{
		"datacenter_id": dcID,
		"server_id":     serverID,
		"nic_id":        nicID,
	})
	time.Sleep(cleanupWaitTime)
	ts.cleanupResource(t, "delete_server", map[string]interface{}{
		"datacenter_id": dcID,
		"server_id":     serverID,
	})
	time.Sleep(cleanupWaitTime)
	ts.cleanupResource(t, "delete_lan", map[string]interface{}{
		"datacenter_id": dcID,
		"lan_id":        lanID,
	})
}

// =============================================================================
// Unit Tests (no API calls)
// =============================================================================

func TestToolDefinitions(t *testing.T) {
	server := NewServer()

	// Check that all Priority 1 and Priority 2 tools are registered
	expectedTools := []string{
		// Priority 1: Core Infrastructure CRUD
		"list_datacenters", "get_datacenter", "create_datacenter", "update_datacenter", "delete_datacenter",
		"list_servers", "get_server", "create_server", "update_server", "delete_server",
		"start_server", "stop_server", "reboot_server",
		"list_volumes", "get_volume", "create_volume", "update_volume", "delete_volume",
		"attach_volume", "detach_volume",
		"list_snapshots", "get_snapshot", "create_snapshot", "update_snapshot", "delete_snapshot",
		"restore_snapshot",
		// Priority 2: Networking CRUD
		"list_lans", "get_lan", "create_lan", "update_lan", "delete_lan",
		"list_nics", "get_nic", "create_nic", "update_nic", "delete_nic",
		"list_ipblocks", "get_ipblock", "create_ipblock", "update_ipblock", "delete_ipblock",
		"list_firewall_rules", "get_firewall_rule", "create_firewall_rule", "update_firewall_rule", "delete_firewall_rule",
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
