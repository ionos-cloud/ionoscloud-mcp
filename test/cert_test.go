package test

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestCertToolEndpoints(t *testing.T) {
	h := setup(t)

	certificate := "c-1"
	autoCert := "ac-1"
	provider := "p-1"

	tests := []toolTest{
		// Certificates
		{"list_certificates", map[string]any{}, []string{"GET"}, []string{"/certificates"}},
		{"get_certificate", map[string]any{"certificate_id": certificate}, []string{"GET"}, []string{"/certificates/" + certificate}},

		// Auto-Certificates
		{"list_auto_certificates", map[string]any{}, []string{"GET"}, []string{"/auto-certificates"}},
		{"get_auto_certificate", map[string]any{"auto_certificate_id": autoCert}, []string{"GET"}, []string{"/auto-certificates/" + autoCert}},

		// Providers
		{"list_providers", map[string]any{}, []string{"GET"}, []string{"/providers"}},
		{"get_provider", map[string]any{"provider_id": provider}, []string{"GET"}, []string{"/providers/" + provider}},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h.log.clear()

			_, err := h.session.CallTool(ctx, &mcp.CallToolParams{
				Name:      tt.name,
				Arguments: tt.args,
			})
			if err != nil {
				t.Fatalf("CallTool(%q) returned protocol error: %v", tt.name, err)
			}

			reqs := h.log.allRequests()
			if len(tt.wantMethods) != len(tt.wantPaths) {
				t.Fatalf("test %q: wantMethods has %d entries, wantPaths has %d", tt.name, len(tt.wantMethods), len(tt.wantPaths))
			}
			if len(reqs) != len(tt.wantPaths) {
				t.Fatalf("CallTool(%q) made %d requests, want %d", tt.name, len(reqs), len(tt.wantPaths))
			}
			for i, req := range reqs {
				if req.Method != tt.wantMethods[i] {
					t.Errorf("CallTool(%q) request[%d] method = %q, want %q", tt.name, i, req.Method, tt.wantMethods[i])
				}
				if req.Path != tt.wantPaths[i] {
					t.Errorf("CallTool(%q) request[%d] path = %q, want %q", tt.name, i, req.Path, tt.wantPaths[i])
				}
			}
		})
	}
}
