package test

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

func TestBillingToolEndpoints(t *testing.T) {
	h := setup(t)

	c := int32(1)

	tests := []toolTest{
		// Profile (no contract param — tool calls ProfilesGet directly)
		{name: "get_billing_profile", args: map[string]any{}, wantMethods: []string{"GET"}, wantPaths: []string{"/billing/profile"}},

		// EVN
		{name: "list_billing_evn", args: map[string]any{"contract": c}, wantMethods: []string{"GET"}, wantPaths: []string{"/billing/1/evn"}},
		{name: "list_billing_evn_by_period", args: map[string]any{"contract": c, "period": "2026-04"}, wantMethods: []string{"GET"}, wantPaths: []string{"/billing/1/evn/2026-04"}},

		// Invoices
		{name: "list_billing_invoices", args: map[string]any{"contract": c}, wantMethods: []string{"GET"}, wantPaths: []string{"/billing/1/invoices"}},
		{name: "list_billing_invoices_by_period", args: map[string]any{"period": "2026-04"}, wantMethods: []string{"GET"}, wantPaths: []string{"/billing/invoices/2026-04"}},
		{name: "get_billing_invoice", args: map[string]any{"contract": c, "invoice_id": "INV123"}, wantMethods: []string{"GET"}, wantPaths: []string{"/billing/1/invoices/INV123"}},

		// Traffic
		{name: "list_billing_traffic", args: map[string]any{"contract": c}, wantMethods: []string{"GET"}, wantPaths: []string{"/billing/1/traffic"}},
		{name: "list_billing_traffic_by_period", args: map[string]any{"contract": c, "period": "2026-04"}, wantMethods: []string{"GET"}, wantPaths: []string{"/billing/1/traffic/2026-04"}},

		// Usage
		{name: "list_billing_usage", args: map[string]any{"contract": c}, wantMethods: []string{"GET"}, wantPaths: []string{"/billing/1/usage"}},
		{name: "get_billing_usage_by_datacenter", args: map[string]any{"contract": c, "datacenter_id": "dc-uuid-1"}, wantMethods: []string{"GET"}, wantPaths: []string{"/billing/1/usage/dc-uuid-1"}},

		// Utilization
		{name: "list_billing_utilization", args: map[string]any{"contract": c}, wantMethods: []string{"GET"}, wantPaths: []string{"/billing/1/utilization"}},
		{name: "list_billing_utilization_by_period", args: map[string]any{"contract": c, "period": "2026-04"}, wantMethods: []string{"GET"}, wantPaths: []string{"/billing/1/utilization/2026-04"}},
		{name: "get_billing_utilization_daily", args: map[string]any{"contract": c, "date": "2026-04-15"}, wantMethods: []string{"GET"}, wantPaths: []string{"/billing/1/utilization/daily/2026-04-15"}},

		// Products
		{name: "list_billing_products", args: map[string]any{"contract": c, "filter": "RAM"}, wantMethods: []string{"GET"}, wantPaths: []string{"/billing/1/products"}},
	}

	h.run(t, tests)
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

func TestBillingUtilizationInputValidation(t *testing.T) {
	h := setup(t)
	ctx := context.Background()
	c := int32(1)

	// Invalid group_by must produce 0 HTTP requests.
	cases := []struct {
		name string
		tool string
		args map[string]any
	}{
		{"util_get bad group_by", "list_billing_utilization", map[string]any{"contract": c, "group_by": "bogus"}},
		{"util_period bad group_by", "list_billing_utilization_by_period", map[string]any{"contract": c, "period": "2026-04", "group_by": "bogus"}},
		{"util_daily bad group_by", "get_billing_utilization_daily", map[string]any{"contract": c, "date": "2026-04-15", "group_by": "bogus"}},
		{"util_daily bad date", "get_billing_utilization_daily", map[string]any{"contract": c, "date": "not-a-date"}},
		{"util_get top_n zero", "list_billing_utilization", map[string]any{"contract": c, "top_n": int32(0)}},
		{"util_get top_n negative", "list_billing_utilization", map[string]any{"contract": c, "top_n": int32(-5)}},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			h.log.clear()
			_, err := h.session.CallTool(ctx, &mcp.CallToolParams{Name: tt.tool, Arguments: tt.args})
			if err != nil {
				t.Fatalf("CallTool returned protocol error: %v", err)
			}
			if reqs := h.log.allRequests(); len(reqs) != 0 {
				t.Errorf("invalid input made %d HTTP requests, want 0", len(reqs))
			}
		})
	}
}

func TestBillingCompactionFlagsRouting(t *testing.T) {
	h := setup(t)
	ctx := context.Background()
	c := int32(1)

	// Flags do not change endpoint routing — same path regardless of flag combination.
	cases := []struct {
		name string
		tool string
		args map[string]any
		path string
	}{
		{"util include_zero", "list_billing_utilization", map[string]any{"contract": c, "include_zero": true}, "/billing/1/utilization"},
		{"util group_by meter", "list_billing_utilization", map[string]any{"contract": c, "group_by": "meter"}, "/billing/1/utilization"},
		{"util group_by datacenter", "list_billing_utilization", map[string]any{"contract": c, "group_by": "datacenter"}, "/billing/1/utilization"},
		{"util scope dc", "list_billing_utilization", map[string]any{"contract": c, "datacenter_id": "dc-x"}, "/billing/1/utilization"},
		{"util scope types", "list_billing_utilization", map[string]any{"contract": c, "meter_types": []string{"DBAAS"}}, "/billing/1/utilization"},
		{"util scope regions", "list_billing_utilization", map[string]any{"contract": c, "regions": []string{"de/fra"}}, "/billing/1/utilization"},
		{"util top_n", "list_billing_utilization", map[string]any{"contract": c, "top_n": int32(10)}, "/billing/1/utilization"},
		{"usage include_zero", "list_billing_usage", map[string]any{"contract": c, "include_zero": true}, "/billing/1/usage"},
		{"usage scope dc", "list_billing_usage", map[string]any{"contract": c, "datacenter_id": "dc-y"}, "/billing/1/usage"},
		{"usage_by_dc include_zero", "get_billing_usage_by_datacenter", map[string]any{"contract": c, "datacenter_id": "dc-z", "include_zero": true}, "/billing/1/usage/dc-z"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			h.log.clear()
			_, err := h.session.CallTool(ctx, &mcp.CallToolParams{Name: tt.tool, Arguments: tt.args})
			if err != nil {
				t.Fatalf("CallTool returned protocol error: %v", err)
			}
			reqs := h.log.allRequests()
			if len(reqs) != 1 {
				t.Fatalf("want 1 request, got %d", len(reqs))
			}
			if reqs[0].Path != tt.path {
				t.Errorf("path = %q, want %q", reqs[0].Path, tt.path)
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
