package test

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// All DNS tool names for registration test.
var dnsToolNames = []string{
	"list_dns_zones", "get_dns_zone", "get_dns_zone_file",
	"list_dns_records", "list_dns_zone_records", "get_dns_record",
	"list_dns_secondary_zone_records",
	"list_dns_reverse_records", "get_dns_reverse_record",
	"list_dns_secondary_zones", "get_dns_secondary_zone", "get_dns_secondary_zone_axfr",
	"list_dns_zone_dnssec_keys",
	"get_dns_quota",
}

func TestDnsToolRegistration(t *testing.T) {
	h := setup(t)

	ctx := context.Background()
	result, err := h.session.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		t.Fatalf("ListTools failed: %v", err)
	}

	registered := make(map[string]bool)
	for _, tool := range result.Tools {
		registered[tool.Name] = true
	}

	for _, name := range dnsToolNames {
		if !registered[name] {
			t.Errorf("DNS tool %q not registered", name)
		}
	}
}

func TestDnsToolEndpoints(t *testing.T) {
	h := setup(t)

	zone := "z-1"
	record := "r-1"
	secondaryZone := "sz-1"
	reverseRecord := "rr-1"

	tests := []toolTest{
		// Zones
		{"list_dns_zones", map[string]any{}, []string{"/zones"}},
		{"get_dns_zone", map[string]any{"zone_id": zone}, []string{"/zones/" + zone}},
		{"get_dns_zone_file", map[string]any{"zone_id": zone}, []string{"/zones/" + zone + "/zonefile"}},

		// Records
		{"list_dns_records", map[string]any{}, []string{"/records"}},
		{"list_dns_zone_records", map[string]any{"zone_id": zone}, []string{"/zones/" + zone + "/records"}},
		{"get_dns_record", map[string]any{"zone_id": zone, "record_id": record}, []string{"/zones/" + zone + "/records/" + record}},
		{"list_dns_secondary_zone_records", map[string]any{"secondary_zone_id": secondaryZone}, []string{"/secondaryzones/" + secondaryZone + "/records"}},

		// Reverse Records
		{"list_dns_reverse_records", map[string]any{}, []string{"/reverserecords"}},
		{"get_dns_reverse_record", map[string]any{"reverse_record_id": reverseRecord}, []string{"/reverserecords/" + reverseRecord}},

		// Secondary Zones
		{"list_dns_secondary_zones", map[string]any{}, []string{"/secondaryzones"}},
		{"get_dns_secondary_zone", map[string]any{"secondary_zone_id": secondaryZone}, []string{"/secondaryzones/" + secondaryZone}},
		{"get_dns_secondary_zone_axfr", map[string]any{"secondary_zone_id": secondaryZone}, []string{"/secondaryzones/" + secondaryZone + "/axfr"}},

		// DNSSEC
		{"list_dns_zone_dnssec_keys", map[string]any{"zone_id": zone}, []string{"/zones/" + zone + "/keys"}},

		// Quota
		{"get_dns_quota", map[string]any{}, []string{"/quota"}},
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
				if req.Path != tt.wantPaths[i] {
					t.Errorf("CallTool(%q) request[%d] path = %q, want %q", tt.name, i, req.Path, tt.wantPaths[i])
				}
			}
		})
	}
}
