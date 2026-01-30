// Package kubernetes provides the Kubernetes toolset for IONOS Cloud MCP.
package kubernetes

import (
	"github.com/ionos-cloud/ionoscloud-mcp/pkg/api"
	"github.com/ionos-cloud/ionoscloud-mcp/pkg/toolsets"
)

// Toolset provides Kubernetes-related tools.
type Toolset struct{}

func (t *Toolset) GetName() string        { return "kubernetes" }
func (t *Toolset) GetDescription() string { return "Kubernetes resources (clusters, node pools, nodes)" }

func (t *Toolset) GetTools() []api.ServerTool {
	var tools []api.ServerTool
	tools = append(tools, initClusters()...)
	tools = append(tools, initNodepools()...)
	return tools
}

func init() {
	toolsets.Register(&Toolset{})
}
