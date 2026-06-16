package test

import (
	"testing"
)

func TestCertToolEndpoints(t *testing.T) {
	h := setup(t)

	certificate := "c-1"
	autoCert := "ac-1"
	provider := "p-1"

	tests := []toolTest{
		// Certificates
		{name: "list_cert_certificates", args: map[string]any{}, wantMethods: []string{"GET"}, wantPaths: []string{"/certificates"}},
		{name: "get_cert_certificate", args: map[string]any{"certificate_id": certificate}, wantMethods: []string{"GET"}, wantPaths: []string{"/certificates/" + certificate}},

		// Auto-Certificates
		{name: "list_cert_auto_certificates", args: map[string]any{}, wantMethods: []string{"GET"}, wantPaths: []string{"/auto-certificates"}},
		{name: "get_cert_auto_certificate", args: map[string]any{"auto_certificate_id": autoCert}, wantMethods: []string{"GET"}, wantPaths: []string{"/auto-certificates/" + autoCert}},

		// Providers
		{name: "list_cert_providers", args: map[string]any{}, wantMethods: []string{"GET"}, wantPaths: []string{"/providers"}},
		{name: "get_cert_provider", args: map[string]any{"provider_id": provider}, wantMethods: []string{"GET"}, wantPaths: []string{"/providers/" + provider}},
	}

	h.run(t, tests)
}
