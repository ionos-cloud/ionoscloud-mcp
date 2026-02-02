// Package testing provides test utilities for IONOS Cloud MCP tests.
package testing

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/ionos-cloud/ionoscloud-mcp/pkg/api"
	"github.com/ionos-cloud/ionoscloud-mcp/pkg/ionos"
	"github.com/ionos-cloud/ionoscloud-mcp/pkg/toolsets"

	// Import all toolsets for registration
	_ "github.com/ionos-cloud/ionoscloud-mcp/pkg/toolsets/compute"
	_ "github.com/ionos-cloud/ionoscloud-mcp/pkg/toolsets/dns"
	_ "github.com/ionos-cloud/ionoscloud-mcp/pkg/toolsets/iam"
	_ "github.com/ionos-cloud/ionoscloud-mcp/pkg/toolsets/kubernetes"
	_ "github.com/ionos-cloud/ionoscloud-mcp/pkg/toolsets/loadbalancing"
	_ "github.com/ionos-cloud/ionoscloud-mcp/pkg/toolsets/networking"
)

// Test constants
const (
	TestLocation         = "de/fra"
	TestServerCores      = 1
	TestServerRAM        = 1024 // MB - minimum for most operations
	TestVolumeSize       = 10   // GB - minimum practical size
	TestVolumeSizeUp     = 20   // GB - for update tests
	DefaultWaitTimeout   = 5 * time.Minute
	SnapshotWaitTimeout  = 10 * time.Minute
	CleanupWaitTime      = 10 * time.Second
	AttachmentWaitTime   = 30 * time.Second
	PollInterval         = 5 * time.Second
)

// TestHelper provides utilities for testing tools.
type TestHelper struct {
	Client *ionos.Client
	T      *testing.T
}

// NewTestHelper creates a new test helper, skipping if no credentials are set.
func NewTestHelper(t *testing.T) *TestHelper {
	if os.Getenv("IONOS_USERNAME") == "" && os.Getenv("IONOS_TOKEN") == "" {
		t.Skip("Skipping E2E test: IONOS_USERNAME/IONOS_PASSWORD or IONOS_TOKEN not set")
	}

	client, err := ionos.NewClient()
	if err != nil {
		t.Fatalf("Failed to create IONOS client: %v", err)
	}

	return &TestHelper{
		Client: client,
		T:      t,
	}
}

// ExecuteTool finds and executes a tool by name with the given arguments.
func (h *TestHelper) ExecuteTool(toolName string, args map[string]interface{}) (*api.ToolCallResult, error) {
	tool, found := toolsets.GetTool(toolName)
	if !found {
		h.T.Fatalf("Tool not found: %s", toolName)
	}

	params := api.ToolHandlerParams{
		Context:   context.Background(),
		Client:    h.Client.Compute,
		DNSClient: h.Client.DNS,
		Arguments: args,
	}

	return tool.Handler(context.Background(), params)
}

// MustExecuteTool executes a tool and fails the test if there's an error.
func (h *TestHelper) MustExecuteTool(toolName string, args map[string]interface{}) *api.ToolCallResult {
	result, err := h.ExecuteTool(toolName, args)
	if err != nil {
		h.T.Fatalf("Failed to execute tool %s: %v", toolName, err)
	}
	return result
}

// CleanupResource attempts to delete a resource and logs any errors.
func (h *TestHelper) CleanupResource(toolName string, args map[string]interface{}) {
	if _, err := h.ExecuteTool(toolName, args); err != nil {
		h.T.Logf("Warning: cleanup failed for %s: %v", toolName, err)
	}
}

// WaitForState polls a resource until it reaches the expected state or times out.
func (h *TestHelper) WaitForState(check func() (string, error), timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		state, err := check()
		if err == nil && state == "AVAILABLE" {
			return
		}
		time.Sleep(PollInterval)
	}
	h.T.Log("Warning: state check timed out")
}

// RequireCredentials skips the test if IONOS credentials are not set.
func RequireCredentials(t *testing.T) {
	if os.Getenv("IONOS_USERNAME") == "" && os.Getenv("IONOS_TOKEN") == "" {
		t.Skip("Skipping: IONOS_USERNAME/IONOS_PASSWORD or IONOS_TOKEN not set")
	}
}
