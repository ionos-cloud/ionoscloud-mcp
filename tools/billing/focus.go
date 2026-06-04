package billing

import (
	"context"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterFocusTools registers the get_billing_focus_spec tool.
// focusSpec is the embedded FOCUS v1.3 markdown content passed from the
// main package where the //go:embed directive lives.
//
// This is a callable alternative to the ionos://billing/focus-v1.3 MCP
// resource for AI clients that do not support resources/read
// (e.g. Claude Desktop, Cursor, VS Code).
func RegisterFocusTools(server *mcp.Server, focusSpec string) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_billing_focus_spec",
		Description: "Returns the FOCUS v1.3 column specification and IONOS tool → FOCUS field mappings. Call this before producing standards-compliant billing output when the ionos://billing/focus-v1.3 resource is not accessible.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		return tools.TextResult(focusSpec), nil, nil
	})
}
