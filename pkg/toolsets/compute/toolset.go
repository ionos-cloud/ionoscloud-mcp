// Package compute provides the compute toolset for IONOS Cloud MCP.
// This toolset includes tools for managing datacenters, servers, volumes, snapshots, and images.
package compute

import (
	"github.com/ionos-cloud/ionoscloud-mcp/pkg/api"
	"github.com/ionos-cloud/ionoscloud-mcp/pkg/toolsets"
)

// Toolset provides compute-related tools.
type Toolset struct{}

// GetName returns the name of the toolset.
func (t *Toolset) GetName() string {
	return "compute"
}

// GetDescription returns a description of the toolset.
func (t *Toolset) GetDescription() string {
	return "Compute resources (datacenters, servers, volumes, snapshots, images)"
}

// GetTools returns all tools provided by this toolset.
func (t *Toolset) GetTools() []api.ServerTool {
	var tools []api.ServerTool
	tools = append(tools, initDatacenters()...)
	tools = append(tools, initServers()...)
	tools = append(tools, initVolumes()...)
	tools = append(tools, initSnapshots()...)
	tools = append(tools, initImages()...)
	return tools
}

func init() {
	toolsets.Register(&Toolset{})
}
