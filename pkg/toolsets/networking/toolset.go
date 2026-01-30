// Package networking provides the networking toolset for IONOS Cloud MCP.
// This toolset includes tools for managing LANs, NICs, IP blocks, firewall rules, NAT gateways, and PCCs.
package networking

import (
	"github.com/ionos-cloud/ionoscloud-mcp/pkg/api"
	"github.com/ionos-cloud/ionoscloud-mcp/pkg/toolsets"
)

// Toolset provides networking-related tools.
type Toolset struct{}

// GetName returns the name of the toolset.
func (t *Toolset) GetName() string {
	return "networking"
}

// GetDescription returns a description of the toolset.
func (t *Toolset) GetDescription() string {
	return "Networking resources (LANs, NICs, IP blocks, firewall rules, NAT gateways, PCCs)"
}

// GetTools returns all tools provided by this toolset.
func (t *Toolset) GetTools() []api.ServerTool {
	var tools []api.ServerTool
	tools = append(tools, initLans()...)
	tools = append(tools, initNics()...)
	tools = append(tools, initIpblocks()...)
	tools = append(tools, initFirewall()...)
	tools = append(tools, initNatgateways()...)
	tools = append(tools, initPcc()...)
	return tools
}

func init() {
	toolsets.Register(&Toolset{})
}
