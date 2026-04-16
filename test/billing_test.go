package test

import (
	"context"
	"testing"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestBillingToolEndpoints(t *testing.T) {
	h := setup(t)

	c := int32(1)

	tests := []toolTest{
		// Profile (no contract param — tool calls ProfilesGet directly)
		{"get_billing_profile", map[string]any{}, []string{"GET"}, []string{"/billing/profile"}},

		// EVN
		{"list_billing_evn", map[string]any{"contract": c}, []string{"GET"}, []string{"/billing/1/evn"}},
		{"list_billing_evn_by_period", map[string]any{"contract": c, "period": "2026-04"}, []string{"GET"}, []string{"/billing/1/evn/2026-04"}},

		// Invoices
		{"list_billing_invoices", map[string]any{"contract": c}, []string{"GET"}, []string{"/billing/1/invoices"}},
		{"list_billing_invoices_by_period", map[string]any{"period": "2026-04"}, []string{"GET"}, []string{"/billing/invoices/2026-04"}},
		{"get_billing_invoice", map[string]any{"contract": c, "invoice_id": "INV123"}, []string{"GET"}, []string{"/billing/1/invoices/INV123"}},

		// Traffic
		{"list_billing_traffic", map[string]any{"contract": c}, []string{"GET"}, []string{"/billing/1/traffic"}},
		{"list_billing_traffic_by_period", map[string]any{"contract": c, "period": "2026-04"}, []string{"GET"}, []string{"/billing/1/traffic/2026-04"}},

		// Usage
		{"list_billing_usage", map[string]any{"contract": c}, []string{"GET"}, []string{"/billing/1/usage"}},
		{"get_billing_usage_by_datacenter", map[string]any{"contract": c, "datacenter_id": "dc-uuid-1"}, []string{"GET"}, []string{"/billing/1/usage/dc-uuid-1"}},

		// Utilization
		{"list_billing_utilization", map[string]any{"contract": c}, []string{"GET"}, []string{"/billing/1/utilization"}},
		{"list_billing_utilization_by_period", map[string]any{"contract": c, "period": "2026-04"}, []string{"GET"}, []string{"/billing/1/utilization/2026-04"}},
		{"get_billing_utilization_daily", map[string]any{"contract": c, "date": "2026-04-15"}, []string{"GET"}, []string{"/billing/1/utilization/daily/2026-04-15"}},

		// Products
		{"list_billing_products", map[string]any{"contract": c, "filter": "RAM"}, []string{"GET"}, []string{"/billing/1/products"}},
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

func TestBillingPeriodValidation(t *testing.T) {
	h := setup(t)
	ctx := context.Background()

	c := int32(1)
	periodTools := []struct {
		name string
		args map[string]any
	}{
		{"list_billing_evn_by_period", map[string]any{"contract": c, "period": "not-a-period"}},
		{"list_billing_invoices_by_period", map[string]any{"period": "2026-13"}},
		{"list_billing_traffic_by_period", map[string]any{"contract": c, "period": "04-2026"}},
		{"list_billing_utilization_by_period", map[string]any{"contract": c, "period": ""}},
	}

	for _, tt := range periodTools {
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
			if len(reqs) != 0 {
				t.Errorf("CallTool(%q) with invalid period made %d HTTP requests, want 0", tt.name, len(reqs))
			}
		})
	}
}

func TestValidatePeriod(t *testing.T) {
	tests := []struct {
		period  string
		wantErr bool
	}{
		{"2026-04", false},
		{"2026-01", false},
		{"2026-12", false},
		{"2026-13", true},
		{"2026-00", true},
		{"04-2026", true},
		{"2026-4", true},
		{"not-a-period", true},
		{"", true},
		{"2026", true},
	}

	for _, tt := range tests {
		t.Run(tt.period, func(t *testing.T) {
			err := tools.ValidatePeriod(tt.period)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePeriod(%q) error = %v, wantErr %v", tt.period, err, tt.wantErr)
			}
		})
	}
}
