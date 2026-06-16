package test

import (
	"testing"
)

func TestDnsToolEndpoints(t *testing.T) {
	h := setup(t)

	zone := "z-1"
	record := "r-1"
	secondaryZone := "sz-1"
	reverseRecord := "rr-1"

	tests := []toolTest{
		// Zones
		{name: "list_dns_zones", args: map[string]any{}, wantMethods: []string{"GET"}, wantPaths: []string{"/zones"}},
		{name: "get_dns_zone", args: map[string]any{"zone_id": zone}, wantMethods: []string{"GET"}, wantPaths: []string{"/zones/" + zone}},
		{name: "get_dns_zone_file", args: map[string]any{"zone_id": zone}, wantMethods: []string{"GET"}, wantPaths: []string{"/zones/" + zone + "/zonefile"}},

		// Records
		{name: "list_dns_records", args: map[string]any{}, wantMethods: []string{"GET"}, wantPaths: []string{"/records"}},
		{name: "list_dns_zone_records", args: map[string]any{"zone_id": zone}, wantMethods: []string{"GET"}, wantPaths: []string{"/zones/" + zone + "/records"}},
		{name: "get_dns_record", args: map[string]any{"zone_id": zone, "record_id": record}, wantMethods: []string{"GET"}, wantPaths: []string{"/zones/" + zone + "/records/" + record}},
		{name: "list_dns_secondary_zone_records", args: map[string]any{"secondary_zone_id": secondaryZone}, wantMethods: []string{"GET"}, wantPaths: []string{"/secondaryzones/" + secondaryZone + "/records"}},

		// Reverse Records
		{name: "list_dns_reverse_records", args: map[string]any{}, wantMethods: []string{"GET"}, wantPaths: []string{"/reverserecords"}},
		{name: "get_dns_reverse_record", args: map[string]any{"reverse_record_id": reverseRecord}, wantMethods: []string{"GET"}, wantPaths: []string{"/reverserecords/" + reverseRecord}},

		// Secondary Zones
		{name: "list_dns_secondary_zones", args: map[string]any{}, wantMethods: []string{"GET"}, wantPaths: []string{"/secondaryzones"}},
		{name: "get_dns_secondary_zone", args: map[string]any{"secondary_zone_id": secondaryZone}, wantMethods: []string{"GET"}, wantPaths: []string{"/secondaryzones/" + secondaryZone}},
		{name: "get_dns_secondary_zone_axfr", args: map[string]any{"secondary_zone_id": secondaryZone}, wantMethods: []string{"GET"}, wantPaths: []string{"/secondaryzones/" + secondaryZone + "/axfr"}},

		// DNSSEC
		{name: "list_dns_zone_dnssec_keys", args: map[string]any{"zone_id": zone}, wantMethods: []string{"GET"}, wantPaths: []string{"/zones/" + zone + "/keys"}},

		// Quota
		{name: "get_dns_quota", args: map[string]any{}, wantMethods: []string{"GET"}, wantPaths: []string{"/quota"}},
	}

	h.run(t, tests)
}

// TestDnsOutput asserts list output is returned to the caller verbatim.
func TestDnsOutput(t *testing.T) {
	h := setup(t)

	tests := []toolTest{
		{
			name:        "list_dns_zones",
			args:        map[string]any{},
			wantMethods: []string{"GET"},
			wantPaths:   []string{"/zones"},
			fixture:     `{"items":[{"id":"z-1","properties":{"zoneName":"example.com"}}]}`,
			wantContain: []string{"z-1", "example.com"},
		},
	}

	h.run(t, tests)
}
