package loader

import (
	"context"
	"sync"

	computeSDK "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	objstSDK "github.com/ionos-cloud/sdk-go-bundle/products/objectstorage/v2"
	objmgmtSDK "github.com/ionos-cloud/sdk-go-bundle/products/objectstoragemanagement/v2"
	"github.com/ionos-cloud/sdk-go-bundle/shared"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
	"github.com/ionos-cloud/ionoscloud-mcp/tools/compute"
	"github.com/ionos-cloud/ionoscloud-mcp/tools/objectstorage"
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
		Description: "Load IONOS Compute Engine tools into this session. Call this before managing virtual infrastructure: servers, datacenters, volumes, NICs, LANs, firewall rules, IP blocks, load balancers, NAT gateways, security groups, snapshots, Kubernetes and more.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		mu.Lock()
		defer mu.Unlock()
		if loaded {
			return tools.TextResult("Compute tools are already loaded."), nil, nil
		}
		compute.RegisterAll(server, client)
		loaded = true
		return tools.TextResult("Compute tools loaded. The tool list has been updated — you can now call list_servers, list_datacenters, and other compute tools."), nil, nil
	})
}

// RegisterObjectStorageLoader registers an MCP tool that lazily loads
// all Object Storage tools on first call.
func RegisterObjectStorageLoader(server *mcp.Server, client *objstSDK.APIClient, mgmtClient *objmgmtSDK.APIClient, cfg *shared.Configuration) {
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
			return tools.TextResult("Object Storage tools are already loaded."), nil, nil
		}
		objectstorage.RegisterAll(server, client, mgmtClient, cfg)
		loaded = true
		return tools.TextResult("Object Storage tools loaded. The tool list has been updated — you can now call list_object_storage_buckets and other object storage tools."), nil, nil
	})
}
