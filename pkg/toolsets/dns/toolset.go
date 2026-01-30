// Package dns provides the DNS toolset for IONOS Cloud MCP.
package dns

import (
	"github.com/ionos-cloud/ionoscloud-mcp/pkg/api"
	"github.com/ionos-cloud/ionoscloud-mcp/pkg/toolsets"
)

// Toolset provides DNS-related tools.
type Toolset struct{}

func (t *Toolset) GetName() string        { return "dns" }
func (t *Toolset) GetDescription() string { return "DNS management (zones, records)" }

func (t *Toolset) GetTools() []api.ServerTool {
	var tools []api.ServerTool
	tools = append(tools, initZones()...)
	tools = append(tools, initRecords()...)
	return tools
}

func init() {
	toolsets.Register(&Toolset{})
}
