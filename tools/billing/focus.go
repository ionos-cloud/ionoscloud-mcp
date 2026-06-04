package billing

import (
	"context"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// RegisterFocusTools registers get_billing_focus_spec for clients that don't support resources/read.
func RegisterFocusTools(server *mcp.Server, focusSpec string) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_billing_focus_spec",
		Description: "Returns the FOCUS v1.3 column specification and IONOS tool → FOCUS field mappings. Call before mapping IONOS invoice/usage/traffic data to FOCUS format, or when user asks for FOCUS-compliant cost output.",
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, any, error) {
		return tools.TextResult(focusSpec), nil, nil
	})
}
