package k8s

import (
	"context"

	ionos "github.com/ionos-cloud/sdk-go-bundle/products/compute/v2"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func RegisterVersionTools(server *mcp.Server, client *ionos.APIClient, scope tools.Scope) {
	tools.RegisterTool(server, scope, tools.MethodGet, &mcp.Tool{
		Name:        "list_k8s_versions",
		Description: "List all available Kubernetes versions in IONOS CLOUD. These are the versions a new cluster or node pool may be created with; to upgrade an existing one, use its availableUpgradeVersions instead (get_k8s_cluster / get_k8s_nodepool).",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		versions, _, err := client.KubernetesApi.K8sVersionsGet(ctx).Execute()
		return tools.ToResult(versions, err)
	})

	tools.RegisterTool(server, scope, tools.MethodGet, &mcp.Tool{
		Name:        "get_k8s_default_version",
		Description: "Get the current default Kubernetes version used by new clusters and node pools in IONOS CLOUD. This is the version create_k8s_cluster picks when k8s_version is omitted.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		version, _, err := client.KubernetesApi.K8sVersionsDefaultGet(ctx).Execute()
		if err != nil {
			return tools.ToResult(nil, err)
		}
		return tools.TextResult(version), nil, nil
	})
}
