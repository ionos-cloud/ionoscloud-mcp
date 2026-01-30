// Package iam provides the IAM (Identity and Access Management) toolset for IONOS Cloud MCP.
package iam

import (
	"github.com/ionos-cloud/ionoscloud-mcp/pkg/api"
	"github.com/ionos-cloud/ionoscloud-mcp/pkg/toolsets"
)

// Toolset provides IAM-related tools.
type Toolset struct{}

func (t *Toolset) GetName() string        { return "iam" }
func (t *Toolset) GetDescription() string { return "Identity and Access Management (users, groups, S3 keys, contract)" }

func (t *Toolset) GetTools() []api.ServerTool {
	var tools []api.ServerTool
	tools = append(tools, initUsers()...)
	tools = append(tools, initGroups()...)
	tools = append(tools, initS3Keys()...)
	tools = append(tools, initContract()...)
	return tools
}

func init() {
	toolsets.Register(&Toolset{})
}
