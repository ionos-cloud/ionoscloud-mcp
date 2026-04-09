package test

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestDnsToolEndpoints(t *testing.T) {
	h := setup(t)

	zone := "z-1"
	record := "r-1"
	secondaryZone := "sz-1"
	reverseRecord := "rr-1"

	tests := []toolTest{
		// Zones
		{"list_dns_zones", map[string]any{}, []string{"GET"}, []string{"/zones"}},
		{"get_dns_zone", map[string]any{"zone_id": zone}, []string{"GET"}, []string{"/zones/" + zone}},
		{"get_dns_zone_file", map[string]any{"zone_id": zone}, []string{"GET"}, []string{"/zones/" + zone + "/zonefile"}},

		// Records
		{"list_dns_records", map[string]any{}, []string{"GET"}, []string{"/records"}},
		{"list_dns_zone_records", map[string]any{"zone_id": zone}, []string{"GET"}, []string{"/zones/" + zone + "/records"}},
		{"get_dns_record", map[string]any{"zone_id": zone, "record_id": record}, []string{"GET"}, []string{"/zones/" + zone + "/records/" + record}},
		{"list_dns_secondary_zone_records", map[string]any{"secondary_zone_id": secondaryZone}, []string{"GET"}, []string{"/secondaryzones/" + secondaryZone + "/records"}},

		// Reverse Records
		{"list_dns_reverse_records", map[string]any{}, []string{"GET"}, []string{"/reverserecords"}},
		{"get_dns_reverse_record", map[string]any{"reverse_record_id": reverseRecord}, []string{"GET"}, []string{"/reverserecords/" + reverseRecord}},

		// Secondary Zones
		{"list_dns_secondary_zones", map[string]any{}, []string{"GET"}, []string{"/secondaryzones"}},
		{"get_dns_secondary_zone", map[string]any{"secondary_zone_id": secondaryZone}, []string{"GET"}, []string{"/secondaryzones/" + secondaryZone}},
		{"get_dns_secondary_zone_axfr", map[string]any{"secondary_zone_id": secondaryZone}, []string{"GET"}, []string{"/secondaryzones/" + secondaryZone + "/axfr"}},

		// DNSSEC
		{"list_dns_zone_dnssec_keys", map[string]any{"zone_id": zone}, []string{"GET"}, []string{"/zones/" + zone + "/keys"}},

		// Quota
		{"get_dns_quota", map[string]any{}, []string{"GET"}, []string{"/quota"}},
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
