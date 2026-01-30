// Package loadbalancing provides the load balancing toolset for IONOS Cloud MCP.
package loadbalancing

import (
	"github.com/ionos-cloud/ionoscloud-mcp/pkg/api"
	"github.com/ionos-cloud/ionoscloud-mcp/pkg/toolsets"
)

// Toolset provides load balancing-related tools.
type Toolset struct{}

func (t *Toolset) GetName() string        { return "loadbalancing" }
func (t *Toolset) GetDescription() string { return "Load balancing resources (ALB, NLB, target groups)" }

func (t *Toolset) GetTools() []api.ServerTool {
	var tools []api.ServerTool
	tools = append(tools, initAlb()...)
	tools = append(tools, initNlb()...)
	tools = append(tools, initTargetgroups()...)
	return tools
}

func init() {
	toolsets.Register(&Toolset{})
}
