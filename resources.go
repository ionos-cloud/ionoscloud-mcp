package main

import (
	"context"
	_ "embed"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed docs/billing/focus-v1.3.md
var focusSpec string

func registerResources(server *mcp.Server) {
	server.AddResource(&mcp.Resource{
		URI:         "ionos://billing/focus-v1.3", // not a real URI, just an identifier for the resource
		Name:        "focus-v1.3",
		Title:       "FOCUS v1.3 Billing Spec",
		Description: "FOCUS v1.3 column names, allowed values, and IONOS tool → FOCUS field mappings. Read this when producing standards-compliant billing output.",
		MIMEType:    "text/markdown",
	}, func(_ context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{{
				URI:      req.Params.URI,
				Text:     focusSpec,
				MIMEType: "text/markdown",
			}},
		}, nil
	})
}
