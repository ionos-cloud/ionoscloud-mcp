package loader

import (
	"context"
	"sync"

	"github.com/ionos-cloud/ionoscloud-mcp/tools/compute"
	"github.com/ionos-cloud/ionoscloud-mcp/tools/objectstorage"
	computeSDK "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	objstSDK "github.com/ionos-cloud/sdk-go-bundle/products/objectstorage/v2"
	objmgmtSDK "github.com/ionos-cloud/sdk-go-bundle/products/objectstoragemanagement/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterComputeLoader registers an MCP tool that lazily loads
// all Compute Engine tools on first call.
func RegisterComputeLoader(server *mcp.Server, client *computeSDK.APIClient) {
	var (
		mu     sync.Mutex
		loaded bool
	)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "ionos_load_compute_tools",
		Description: "Load IONOS Compute Engine tools into this session. Call this before managing virtual infrastructure: servers, datacenters, volumes, NICs, LANs, firewall rules, IP blocks, load balancers, NAT gateways, security groups, snapshots, and more.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		mu.Lock()
		defer mu.Unlock()
		if loaded {
			return textResult("Compute tools are already loaded."), nil, nil
		}
		compute.RegisterAll(server, client)
		loaded = true
		return textResult("Compute tools loaded. The tool list has been updated — you can now call list_servers, list_datacenters, and other compute tools."), nil, nil
	})
}

// RegisterObjectStorageLoader registers an MCP tool that lazily loads
// all Object Storage tools on first call.
func RegisterObjectStorageLoader(server *mcp.Server, client *objstSDK.APIClient, mgmtClient *objmgmtSDK.APIClient) {
	var (
		mu     sync.Mutex
		loaded bool
	)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "ionos_load_objectstorage_tools",
		Description: "Load IONOS Object Storage tools into this session. Call this before managing S3-compatible object storage: buckets, bucket configuration (CORS, encryption, lifecycle, policy, replication, versioning), objects, access keys, and regions.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		mu.Lock()
		defer mu.Unlock()
		if loaded {
			return textResult("Object Storage tools are already loaded."), nil, nil
		}
		objectstorage.RegisterAll(server, client, mgmtClient)
		loaded = true
		return textResult("Object Storage tools loaded. The tool list has been updated — you can now call list_object_storage_buckets and other object storage tools."), nil, nil
	})
}

func textResult(text string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}
}
