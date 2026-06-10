package k8s

import (
	"context"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterVersionTools(server *mcp.Server, client *ionos.APIClient) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_k8s_versions",
		Description: "List all available Kubernetes versions in IONOS CLOUD",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		versions, _, err := client.KubernetesApi.K8sVersionsGet(ctx).Execute()
		return tools.ToResult(versions, err)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_k8s_default_version",
		Description: "Get the current default Kubernetes version used by new clusters and node pools in IONOS CLOUD",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		version, _, err := client.KubernetesApi.K8sVersionsDefaultGet(ctx).Execute()
		if err != nil {
			return tools.ToResult(nil, err)
		}
		return tools.TextResult(version), nil, nil
	})
}
