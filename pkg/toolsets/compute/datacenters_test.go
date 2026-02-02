package compute

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/ionos-cloud/ionoscloud-mcp/pkg/api"
	"github.com/ionos-cloud/ionoscloud-mcp/pkg/ionos"
	ionoscloud "github.com/ionos-cloud/sdk-go/v6"
)

const (
	testLocation       = "de/fra"
	defaultWaitTimeout = 5 * time.Minute
	pollInterval       = 5 * time.Second
)

func skipIfNoCredentials(t *testing.T) {
	if os.Getenv("IONOS_USERNAME") == "" && os.Getenv("IONOS_TOKEN") == "" {
		t.Skip("Skipping E2E test: IONOS_USERNAME/IONOS_PASSWORD or IONOS_TOKEN not set")
	}
}

func newTestClient(t *testing.T) *ionos.Client {
	client, err := ionos.NewClient()
	if err != nil {
		t.Fatalf("Failed to create IONOS client: %v", err)
	}
	return client
}

func newTestParams(t *testing.T, args map[string]interface{}) api.ToolHandlerParams {
	client := newTestClient(t)
	return api.ToolHandlerParams{
		Context:   context.Background(),
		Client:    client.Compute,
		DNSClient: client.DNS,
		Arguments: args,
	}
}

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

func TestDatacenterCRUD(t *testing.T) {
	skipIfNoCredentials(t)
	client := newTestClient(t)
	ctx := context.Background()

	var dcID string

	// Create datacenter
	t.Run("CreateDatacenter", func(t *testing.T) {
		params := api.ToolHandlerParams{
			Context: ctx,
			Client:  client.Compute,
			Arguments: map[string]interface{}{
				"name":        "test-dc-mcp",
				"location":    testLocation,
				"description": "Test datacenter for MCP E2E tests",
			},
		}

		result, err := createDatacenter(ctx, params)
		if err != nil {
			t.Fatalf("Failed to create datacenter: %v", err)
		}

		var dc ionoscloud.Datacenter
		if err := json.Unmarshal([]byte(result.Content), &dc); err != nil {
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

	// Cleanup at the end
	defer func() {
		params := api.ToolHandlerParams{
			Context:   ctx,
			Client:    client.Compute,
			Arguments: map[string]interface{}{"datacenter_id": dcID},
		}
		if _, err := deleteDatacenter(ctx, params); err != nil {
			t.Logf("Warning: cleanup failed: %v", err)
		}
	}()

	// Wait for datacenter to be available
	waitForState(t, func() (string, error) {
		params := api.ToolHandlerParams{
			Context:   ctx,
			Client:    client.Compute,
			Arguments: map[string]interface{}{"datacenter_id": dcID},
		}
		result, err := getDatacenter(ctx, params)
		if err != nil {
			return "", err
		}
		var dc ionoscloud.Datacenter
		json.Unmarshal([]byte(result.Content), &dc)
		if dc.Metadata != nil && dc.Metadata.State != nil {
			return *dc.Metadata.State, nil
		}
		return "", nil
	}, defaultWaitTimeout)

	// Update datacenter
	t.Run("UpdateDatacenter", func(t *testing.T) {
		params := api.ToolHandlerParams{
			Context: ctx,
			Client:  client.Compute,
			Arguments: map[string]interface{}{
				"datacenter_id": dcID,
				"name":          "test-dc-mcp-updated",
				"description":   "Updated description",
			},
		}

		result, err := updateDatacenter(ctx, params)
		if err != nil {
			t.Fatalf("Failed to update datacenter: %v", err)
		}

		var dc ionoscloud.Datacenter
		if err := json.Unmarshal([]byte(result.Content), &dc); err != nil {
			t.Fatalf("Failed to unmarshal datacenter: %v", err)
		}

		if dc.Properties == nil || dc.Properties.Name == nil || *dc.Properties.Name != "test-dc-mcp-updated" {
			t.Fatal("Datacenter name was not updated")
		}

		t.Log("Updated datacenter successfully")
	})

	// List datacenters
	t.Run("ListDatacenters", func(t *testing.T) {
		params := api.ToolHandlerParams{
			Context:   ctx,
			Client:    client.Compute,
			Arguments: map[string]interface{}{},
		}

		result, err := listDatacenters(ctx, params)
		if err != nil {
			t.Fatalf("Failed to list datacenters: %v", err)
		}

		var dcs ionoscloud.Datacenters
		if err := json.Unmarshal([]byte(result.Content), &dcs); err != nil {
			t.Fatalf("Failed to unmarshal datacenters: %v", err)
		}

		if dcs.Items == nil || len(*dcs.Items) == 0 {
			t.Fatal("No datacenters found")
		}

		t.Logf("Found %d datacenters", len(*dcs.Items))
	})

	// Get datacenter
	t.Run("GetDatacenter", func(t *testing.T) {
		params := api.ToolHandlerParams{
			Context:   ctx,
			Client:    client.Compute,
			Arguments: map[string]interface{}{"datacenter_id": dcID},
		}

		result, err := getDatacenter(ctx, params)
		if err != nil {
			t.Fatalf("Failed to get datacenter: %v", err)
		}

		var dc ionoscloud.Datacenter
		if err := json.Unmarshal([]byte(result.Content), &dc); err != nil {
			t.Fatalf("Failed to unmarshal datacenter: %v", err)
		}

		if dc.Id == nil || *dc.Id != dcID {
			t.Fatal("Wrong datacenter returned")
		}

		t.Log("Got datacenter successfully")
	})
}
