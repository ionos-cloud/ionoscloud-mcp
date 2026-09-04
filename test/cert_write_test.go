package test

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/ionos-cloud/ionoscloud-mcp/tools"
)

// Certificate Manager write-tool tests. Two themes carry most of the weight: the
// write-only fields (a certificate's private key, a provider's external account
// binding secret) must never appear in a preview or a tool result, and all three
// PATCH endpoints accept only the resource name, so an update body that carried
// anything else would be the API's problem to reject rather than ours to send.

const (
	certID     = "c-1"
	autoCertID = "ac-1"
	providerID = "p-1"

	certsPath     = "/certificates"
	certPath      = certsPath + "/" + certID
	autoCertsPath = "/auto-certificates"
	autoCertPath  = autoCertsPath + "/" + autoCertID
	providersPath = "/providers"
	providerPath  = providersPath + "/" + providerID
	acmeDirectory = "https://acme-v02.api.letsencrypt.org/directory"
	eabKeySecret  = "s3cr3t-eab-material"

	// Fixtures the API would return, so a preview has something to show. A
	// certificate read carries its private key blank, as the API leaves it.
	certFixture     = `{"id":"c-1","properties":{"name":"prod cert","certificate":"-----BEGIN CERTIFICATE-----\nAAAA\n-----END CERTIFICATE-----","certificateChain":"","privateKey":""},"metadata":{"state":"AVAILABLE","commonName":"www.example.com","subjectAlternativeNames":["app.example.com"],"expired":false,"notAfter":"2027-01-31T23:59:59Z","serialNumber":"33:00","autoCertificate":"ac-1"}}`
	autoCertFixture = `{"id":"ac-1","properties":{"provider":"p-1","commonName":"www.example.com","keyAlgorithm":"rsa4096","name":"renewing cert","subjectAlternativeNames":["app.example.com"]},"metadata":{"state":"AVAILABLE","lastIssuedCertificate":"c-1"}}`
	providerFixture = `{"id":"p-1","properties":{"name":"Let's Encrypt","email":"ops@example.com","server":"https://acme-v02.api.letsencrypt.org/directory","externalAccountBinding":{"keyId":"key-1","keySecret":"leaked-if-returned"}},"metadata":{"state":"AVAILABLE"}}`
)

// certPatchProperties returns the decoded properties object of the one PATCH in the
// log, plus the body's top-level keys.
func certPatchProperties(t *testing.T, h *testSetup) (props map[string]any, topLevel []string) {
	t.Helper()
	req := singleRequest(t, h, http.MethodPatch)
	var body map[string]json.RawMessage
	if err := json.Unmarshal([]byte(req.Body), &body); err != nil {
		t.Fatalf("decoding PATCH body %q: %v", req.Body, err)
	}
	for k := range body {
		topLevel = append(topLevel, k)
	}
	if err := json.Unmarshal(body["properties"], &props); err != nil {
		t.Fatalf("decoding PATCH properties %q: %v", req.Body, err)
	}
	return props, topLevel
}

// TestCreateCertCertificateIsNotRegistered pins the deliberate omission. The API
// requires the private key in the create body, so the tool would have to accept it as
// an argument — putting it in the model's context, the on-disk transcript and every
// later request to the model provider, none of which redaction reaches. If someone
// adds the tool back, this fails and they have to read why first.
func TestCreateCertCertificateIsNotRegistered(t *testing.T) {
	for _, scope := range []tools.Scope{{}, {Write: true}, {Write: true, Destructive: true}} {
		h := setupWithScope(t, scope)
		if toolNames(t, context.Background(), h)["create_cert_certificate"] {
			t.Errorf("scope %s: create_cert_certificate must NOT be registered — it would take a private key as a tool argument", scope)
		}
	}
}

// ---------- update_cert_* ----------

// TestUpdateCertToolsSendOnlyTheName pins every PATCH body: the endpoints accept
// the spec's PatchName and nothing else, and metadata must stay out of the body.
func TestUpdateCertToolsSendOnlyTheName(t *testing.T) {
	tests := []struct {
		tool     string
		args     map[string]any
		wantPath string
	}{
		{"update_cert_certificate", map[string]any{"certificate_id": certID, "name": "renamed"}, certPath},
		{"update_cert_auto_certificate", map[string]any{"auto_certificate_id": autoCertID, "name": "renamed"}, autoCertPath},
		{"update_cert_provider", map[string]any{"provider_id": providerID, "name": "renamed"}, providerPath},
	}
	for _, tt := range tests {
		t.Run(tt.tool, func(t *testing.T) {
			h := destructiveSetup(t)
			res := callTool(t, h, tt.tool, tt.args)
			if res.IsError {
				t.Fatalf("%s failed: %s", tt.tool, resultText(res))
			}
			props, topLevel := certPatchProperties(t, h)
			if len(topLevel) != 1 || topLevel[0] != "properties" {
				t.Errorf("PATCH body top-level keys = %v, want only [properties]", topLevel)
			}
			assertDnsKeys(t, props, "name")
			if props["name"] != "renamed" {
				t.Errorf("properties.name = %v, want \"renamed\"", props["name"])
			}
			if got := h.log.allRequests()[0].Path; got != tt.wantPath {
				t.Errorf("PATCH path = %s, want %s", got, tt.wantPath)
			}
		})
	}
}

func TestUpdateCertToolsRejectABlankName(t *testing.T) {
	h := destructiveSetup(t)
	tests := []struct {
		tool string
		args map[string]any
	}{
		{"update_cert_certificate", map[string]any{"certificate_id": certID, "name": "  "}},
		{"update_cert_auto_certificate", map[string]any{"auto_certificate_id": autoCertID, "name": ""}},
		{"update_cert_provider", map[string]any{"provider_id": "   ", "name": "renamed"}},
	}
	for _, tt := range tests {
		t.Run(tt.tool, func(t *testing.T) {
			h.log.clear()
			res := callTool(t, h, tt.tool, tt.args)
			if !res.IsError {
				t.Fatalf("expected a rejection, got: %s", resultText(res))
			}
			assertNoMutation(t, h, tt.tool)
		})
	}
}

// TestUpdateCertProviderNeverReturnsTheEabSecret: the provider read models the
// secret, so a rename's response has to be redacted like a read's.
func TestUpdateCertProviderNeverReturnsTheEabSecret(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(providerPath, providerFixture)

	res := callTool(t, h, "update_cert_provider", map[string]any{"provider_id": providerID, "name": "renamed"})
	got := resultText(res)
	if strings.Contains(got, "leaked-if-returned") {
		t.Errorf("the result must not carry the EAB secret back:\n%s", got)
	}
	if !strings.Contains(got, `"keyId":"key-1"`) {
		t.Errorf("the non-secret half of the binding should survive:\n%s", got)
	}
}

// ---------- delete_cert_certificate ----------

func TestDeleteCertCertificateTwoPhase(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(certPath, certFixture)

	preview, res := previewThenExecute(t, h, "delete_cert_certificate", map[string]any{"certificate_id": certID})
	for _, want := range []string{"IRREVERSIBLE", "prod cert", "www.example.com", "2027-01-31T23:59:59Z", "ac-1"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing %q:\n%s", want, preview)
		}
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	if req := singleRequest(t, h, http.MethodDelete); req.Path != certPath {
		t.Errorf("DELETE path = %s, want %s", req.Path, certPath)
	}
}

// TestDeleteCertCertificateWarnsAboutTheAutoCertificate: deleting a certificate an
// auto-certificate issued does not stop renewal, and the preview has to say so.
func TestDeleteCertCertificateWarnsAboutTheAutoCertificate(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(certPath, certFixture)

	preview := resultText(callTool(t, h, "delete_cert_certificate", map[string]any{"certificate_id": certID}))
	if !strings.Contains(preview, "auto-certificate, which still exists") {
		t.Errorf("preview should warn that renewal continues:\n%s", preview)
	}
}

func TestDeleteCertCertificateNotFound(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serveStatus(certPath, http.StatusNotFound, `{"httpStatus":404}`)

	res := callTool(t, h, "delete_cert_certificate", map[string]any{"certificate_id": certID})
	if !res.IsError {
		t.Fatalf("expected an error for a missing certificate: %s", resultText(res))
	}
	if got := resultText(res); !strings.Contains(got, "does not exist") {
		t.Errorf("error = %q, want it to say the certificate does not exist", got)
	}
}

// ---------- create_cert_auto_certificate ----------

func TestCreateCertAutoCertificateTwoPhase(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(providerPath, providerFixture)

	preview, res := previewThenExecute(t, h, "create_cert_auto_certificate", map[string]any{
		"provider_id": providerID, "common_name": "www.example.com",
		"key_algorithm": "RSA4096", "name": "renewing cert",
		"subject_alternative_names": []any{"app.example.com", "  "},
	})
	for _, want := range []string{"CREATE one auto-certificate", "www.example.com", "rsa4096", "Let's Encrypt", acmeDirectory} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing %q:\n%s", want, preview)
		}
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodPost)
	if req.Path != autoCertsPath {
		t.Errorf("POST path = %s, want %s", req.Path, autoCertsPath)
	}
	for _, want := range []string{`"provider":"p-1"`, `"commonName":"www.example.com"`, `"keyAlgorithm":"rsa4096"`, `"subjectAlternativeNames":["app.example.com"]`} {
		if !strings.Contains(req.Body, want) {
			t.Errorf("POST body missing %s:\n%s", want, req.Body)
		}
	}
}

// TestCreateCertAutoCertificateResolvesTheProvider: an unknown provider gets a named
// field error from the preview instead of a 422 from the execute call.
func TestCreateCertAutoCertificateResolvesTheProvider(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serveStatus(providerPath, http.StatusNotFound, `{"httpStatus":404}`)

	res := callTool(t, h, "create_cert_auto_certificate", map[string]any{
		"provider_id": providerID, "common_name": "www.example.com",
		"key_algorithm": "rsa2048", "name": "renewing cert",
	})
	if !res.IsError {
		t.Fatalf("expected a rejection for an unknown provider: %s", resultText(res))
	}
	if got := resultText(res); !strings.Contains(got, "list_cert_providers") {
		t.Errorf("error = %q, want it to point at list_cert_providers", got)
	}
	assertNoMutation(t, h, "create_cert_auto_certificate with an unknown provider")
}

func TestCreateCertAutoCertificateValidation(t *testing.T) {
	h := destructiveSetup(t)
	valid := map[string]any{
		"provider_id": providerID, "common_name": "www.example.com",
		"key_algorithm": "rsa4096", "name": "renewing cert",
	}
	tests := []struct {
		name    string
		args    map[string]any
		wantMsg string
	}{
		{"blank provider", withArgs(valid, "provider_id", " "), "provider_id is required"},
		{"blank common name", withArgs(valid, "common_name", ""), "common_name is required"},
		{"blank name", withArgs(valid, "name", "  "), "name is required"},
		{"unknown algorithm", withArgs(valid, "key_algorithm", "ecdsa256"), "rsa2048, rsa3072, rsa4096"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h.log.clear()
			res := callTool(t, h, "create_cert_auto_certificate", tt.args)
			if !res.IsError {
				t.Fatalf("expected a rejection, got: %s", resultText(res))
			}
			if got := resultText(res); !strings.Contains(got, tt.wantMsg) {
				t.Errorf("error = %q, want it to contain %q", got, tt.wantMsg)
			}
			assertNoMutation(t, h, tt.name)
		})
	}
}

// ---------- delete_cert_auto_certificate ----------

func TestDeleteCertAutoCertificateCountsIssuedCertificates(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(autoCertPath, autoCertFixture)
	h.resp.serve(certsPath, `{"items":[{"id":"c-1"},{"id":"c-2"}]}`)

	preview, res := previewThenExecute(t, h, "delete_cert_auto_certificate", map[string]any{"auto_certificate_id": autoCertID})
	for _, want := range []string{"IRREVERSIBLE", "Not deleted, but affected", "2 certificates it issued", "renewing cert"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing %q:\n%s", want, preview)
		}
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	if req := singleRequest(t, h, http.MethodDelete); req.Path != autoCertPath {
		t.Errorf("DELETE path = %s, want %s", req.Path, autoCertPath)
	}
}

// TestDeleteCertAutoCertificateFiltersByItsOwnId: the count has to be scoped to this
// auto-certificate, or the preview reports the whole account's certificates.
func TestDeleteCertAutoCertificateFiltersByItsOwnId(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(autoCertPath, autoCertFixture)
	h.resp.serve(certsPath, `{"items":[{"id":"c-1"}]}`)

	callTool(t, h, "delete_cert_auto_certificate", map[string]any{"auto_certificate_id": autoCertID})
	filter, listed := "", false
	for _, r := range h.log.allRequests() {
		if r.Path == certsPath {
			filter, listed = r.Query.Get("filter.autoCertificate"), true
		}
	}
	if !listed {
		t.Fatalf("the preview should list certificates: %+v", h.log.allRequests())
	}
	if filter != autoCertID {
		t.Errorf("filter.autoCertificate = %q, want %q", filter, autoCertID)
	}
}

// TestDeleteCertAutoCertificatePreviewNeverClaimsEmptyOnError: an unreadable list is
// not an empty one, and the difference matters when authorizing a delete.
func TestDeleteCertAutoCertificatePreviewNeverClaimsEmptyOnError(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(autoCertPath, autoCertFixture)
	h.resp.serveStatus(certsPath, http.StatusInternalServerError, `{"httpStatus":500}`)

	preview := resultText(callTool(t, h, "delete_cert_auto_certificate", map[string]any{"auto_certificate_id": autoCertID}))
	if strings.Contains(preview, "issued no certificates") {
		t.Errorf("an unreadable list must not be reported as an empty one:\n%s", preview)
	}
	if !strings.Contains(preview, "INCOMPLETE") {
		t.Errorf("preview should warn the radius is incomplete:\n%s", preview)
	}
}

// ---------- create_cert_provider ----------

func TestCreateCertProviderTwoPhase(t *testing.T) {
	h := destructiveSetup(t)
	preview, res := previewThenExecute(t, h, "create_cert_provider", map[string]any{
		"name": "Let's Encrypt", "email": "ops@example.com", "server": acmeDirectory,
	})
	for _, want := range []string{"CREATE one certificate provider", "Let's Encrypt", "ops@example.com", acmeDirectory} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing %q:\n%s", want, preview)
		}
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodPost)
	if req.Path != providersPath {
		t.Errorf("POST path = %s, want %s", req.Path, providersPath)
	}
	if strings.Contains(req.Body, "externalAccountBinding") {
		t.Errorf("an omitted binding must not be sent:\n%s", req.Body)
	}
}

func TestCreateCertProviderSendsTheExternalAccountBinding(t *testing.T) {
	h := destructiveSetup(t)
	preview, res := previewThenExecute(t, h, "create_cert_provider", map[string]any{
		"name": "ZeroSSL", "email": "ops@example.com", "server": acmeDirectory,
		"external_account_binding": map[string]any{"key_id": "key-1", "key_secret": eabKeySecret},
	})
	if strings.Contains(preview, eabKeySecret) {
		t.Errorf("the preview must not echo the binding secret:\n%s", preview)
	}
	if !strings.Contains(preview, "key_id") || !strings.Contains(preview, "key-1") {
		t.Errorf("the preview should show the non-secret half:\n%s", preview)
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	req := singleRequest(t, h, http.MethodPost)
	for _, want := range []string{`"keyId":"key-1"`, `"keySecret":"` + eabKeySecret + `"`} {
		if !strings.Contains(req.Body, want) {
			t.Errorf("POST body missing %s:\n%s", want, req.Body)
		}
	}
}

func TestCreateCertProviderValidation(t *testing.T) {
	h := destructiveSetup(t)
	valid := map[string]any{"name": "Let's Encrypt", "email": "ops@example.com", "server": acmeDirectory}

	tests := []struct {
		name    string
		args    map[string]any
		wantMsg string
	}{
		{"blank name", withArgs(valid, "name", " "), "name is required"},
		{"not an email", withArgs(valid, "email", "ops.example.com"), "not a plain email address"},
		{"email with a display name", withArgs(valid, "email", "Ops <ops@example.com>"), "not a plain email address"},
		{"server without a scheme", withArgs(valid, "server", "acme-v02.api.letsencrypt.org/directory"), "not an https ACME directory URL"},
		{"server over plain http", withArgs(valid, "server", "http://acme.example.com/directory"), "not an https ACME directory URL"},
		{"binding missing its secret", withArgs(valid, "external_account_binding", map[string]any{"key_id": "key-1", "key_secret": " "}), "needs both key_id and key_secret"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h.log.clear()
			res := callTool(t, h, "create_cert_provider", tt.args)
			if !res.IsError {
				t.Fatalf("expected a rejection, got: %s", resultText(res))
			}
			if got := resultText(res); !strings.Contains(got, tt.wantMsg) {
				t.Errorf("error = %q, want it to contain %q", got, tt.wantMsg)
			}
			assertNoMutation(t, h, tt.name)
		})
	}
}

// ---------- delete_cert_provider ----------

func TestDeleteCertProviderCountsAutoCertificatesUsingIt(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(providerPath, providerFixture)
	// One matching auto-certificate and one on another provider, so the client-side
	// match is doing real work.
	h.resp.serve(autoCertsPath, `{"items":[{"id":"ac-1","properties":{"provider":"p-1"}},{"id":"ac-2","properties":{"provider":"p-2"}}]}`)

	preview, res := previewThenExecute(t, h, "delete_cert_provider", map[string]any{"provider_id": providerID})
	for _, want := range []string{"IRREVERSIBLE", "1 auto-certificates issuing through it", "ops@example.com"} {
		if !strings.Contains(preview, want) {
			t.Errorf("preview missing %q:\n%s", want, preview)
		}
	}
	if strings.Contains(preview, "leaked-if-returned") {
		t.Errorf("the preview must not echo the binding secret:\n%s", preview)
	}
	if res.IsError {
		t.Fatalf("execute failed: %s", resultText(res))
	}
	if req := singleRequest(t, h, http.MethodDelete); req.Path != providerPath {
		t.Errorf("DELETE path = %s, want %s", req.Path, providerPath)
	}
}

func TestDeleteCertProviderPreviewNeverClaimsEmptyOnError(t *testing.T) {
	h := destructiveSetup(t)
	h.resp.serve(providerPath, providerFixture)
	h.resp.serveStatus(autoCertsPath, http.StatusTooManyRequests, `{"httpStatus":429}`)

	preview := resultText(callTool(t, h, "delete_cert_provider", map[string]any{"provider_id": providerID}))
	if strings.Contains(preview, "No auto-certificate issues through") {
		t.Errorf("an unreadable list must not be reported as an empty one:\n%s", preview)
	}
	if !strings.Contains(preview, "INCOMPLETE") {
		t.Errorf("preview should warn the radius is incomplete:\n%s", preview)
	}
}

// ---------- registration, annotations, scope ----------

func TestCertReadToolsAreAnnotatedReadOnly(t *testing.T) {
	h := setup(t) // read-only scope: only the read tools register
	ctx := context.Background()

	want := map[string]bool{
		"list_cert_certificates": true, "get_cert_certificate": true,
		"list_cert_auto_certificates": true, "get_cert_auto_certificate": true,
		"list_cert_providers": true, "get_cert_provider": true,
	}
	seen := map[string]bool{}
	for tool, err := range h.session.Tools(ctx, nil) {
		if err != nil {
			t.Fatalf("listing tools: %v", err)
		}
		if !want[tool.Name] {
			continue
		}
		seen[tool.Name] = true
		if tool.Annotations == nil {
			t.Errorf("%s has no annotations; it must carry ReadOnlyHint", tool.Name)
			continue
		}
		if !tool.Annotations.ReadOnlyHint {
			t.Errorf("%s ReadOnlyHint = false, want true", tool.Name)
		}
	}
	for name := range want {
		if !seen[name] {
			t.Errorf("read tool %s was not registered", name)
		}
	}
}

func TestCertWriteToolAnnotations(t *testing.T) {
	h := destructiveSetup(t)
	ctx := context.Background()

	want := map[string]struct{ destructive, idempotent bool }{
		"update_cert_certificate":      {false, true},
		"delete_cert_certificate":      {true, true},
		"create_cert_auto_certificate": {false, false},
		"update_cert_auto_certificate": {false, true},
		"delete_cert_auto_certificate": {true, true},
		"create_cert_provider":         {false, false},
		"update_cert_provider":         {false, true},
		"delete_cert_provider":         {true, true},
	}
	seen := map[string]bool{}
	for tool, err := range h.session.Tools(ctx, nil) {
		if err != nil {
			t.Fatalf("listing tools: %v", err)
		}
		w, ok := want[tool.Name]
		if !ok {
			continue
		}
		seen[tool.Name] = true
		if tool.Annotations == nil {
			t.Errorf("%s has no annotations", tool.Name)
			continue
		}
		if tool.Annotations.ReadOnlyHint {
			t.Errorf("%s ReadOnlyHint = true, want false", tool.Name)
		}
		if tool.Annotations.DestructiveHint == nil || *tool.Annotations.DestructiveHint != w.destructive {
			t.Errorf("%s DestructiveHint = %v, want %v", tool.Name, tool.Annotations.DestructiveHint, w.destructive)
		}
		if tool.Annotations.IdempotentHint != w.idempotent {
			t.Errorf("%s IdempotentHint = %v, want %v", tool.Name, tool.Annotations.IdempotentHint, w.idempotent)
		}
	}
	for name := range want {
		if !seen[name] {
			t.Errorf("write tool %s was not registered", name)
		}
	}
}

func TestCertWriteToolsAreScopeGated(t *testing.T) {
	reads := []string{
		"list_cert_certificates", "get_cert_certificate",
		"list_cert_auto_certificates", "get_cert_auto_certificate",
		"list_cert_providers", "get_cert_provider",
	}
	writes := []string{
		"update_cert_certificate",
		"create_cert_auto_certificate", "update_cert_auto_certificate",
		"create_cert_provider", "update_cert_provider",
	}
	destructives := []string{
		"delete_cert_certificate", "delete_cert_auto_certificate", "delete_cert_provider",
	}

	tests := []struct {
		name    string
		scope   tools.Scope
		present []string
		absent  []string
	}{
		{"read only", tools.Scope{}, reads, append(append([]string{}, writes...), destructives...)},
		{"write", tools.Scope{Write: true}, append(append([]string{}, reads...), writes...), destructives},
		{
			"destructive",
			tools.Scope{Write: true, Destructive: true},
			append(append(append([]string{}, reads...), writes...), destructives...),
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := setupWithScope(t, tt.scope)
			names := toolNames(t, context.Background(), h)
			for _, n := range tt.present {
				if !names[n] {
					t.Errorf("scope %s: %s should be registered", tt.scope, n)
				}
			}
			for _, n := range tt.absent {
				if names[n] {
					t.Errorf("scope %s: %s must NOT be registered", tt.scope, n)
				}
			}
		})
	}
}

// TestCertWriteToolsDeclareCompletionSemantics: the POSTs answer 201 and the PATCHes
// 200 with the stored resource, while only the DELETEs are asynchronous. A tool that
// claimed otherwise would send the model polling for a change already applied, or
// chaining off one that has not happened.
func TestCertWriteToolsDeclareCompletionSemantics(t *testing.T) {
	h := destructiveSetup(t)
	ctx := context.Background()

	async := map[string]bool{
		"delete_cert_certificate": true, "delete_cert_auto_certificate": true,
		"delete_cert_provider": true,
	}
	sync := map[string]bool{
		"update_cert_certificate":      true,
		"create_cert_auto_certificate": true, "update_cert_auto_certificate": true,
		"create_cert_provider": true, "update_cert_provider": true,
	}
	seen := map[string]bool{}
	for tool, err := range h.session.Tools(ctx, nil) {
		if err != nil {
			t.Fatalf("listing tools: %v", err)
		}
		switch {
		case async[tool.Name]:
			seen[tool.Name] = true
			if !strings.Contains(tool.Description, "Asynchronous (202)") {
				t.Errorf("%s must declare that it is asynchronous: %s", tool.Name, tool.Description)
			}
		case sync[tool.Name]:
			seen[tool.Name] = true
			if !strings.Contains(tool.Description, "Synchronous") {
				t.Errorf("%s must say it is synchronous, so no polling step is invented: %s", tool.Name, tool.Description)
			}
			if strings.Contains(tool.Description, "Asynchronous (202)") {
				t.Errorf("%s must not claim to be asynchronous: %s", tool.Name, tool.Description)
			}
		}
	}
	for name := range async {
		if !seen[name] {
			t.Errorf("%s was not registered", name)
		}
	}
	for name := range sync {
		if !seen[name] {
			t.Errorf("%s was not registered", name)
		}
	}
}

// TestUpdateCertToolsSayOnlyTheNameChanges: the PATCH endpoints take the spec's
// PatchName, so a description that implied more would invite a silent no-op.
func TestUpdateCertToolsSayOnlyTheNameChanges(t *testing.T) {
	h := destructiveSetup(t)
	want := map[string]bool{
		"update_cert_certificate": true, "update_cert_auto_certificate": true,
		"update_cert_provider": true,
	}
	seen := map[string]bool{}
	for tool, err := range h.session.Tools(context.Background(), nil) {
		if err != nil {
			t.Fatalf("listing tools: %v", err)
		}
		if !want[tool.Name] {
			continue
		}
		seen[tool.Name] = true
		if !strings.Contains(tool.Description, "Only the name can be changed") {
			t.Errorf("%s must say only the name can be changed: %s", tool.Name, tool.Description)
		}
	}
	for name := range want {
		if !seen[name] {
			t.Errorf("%s was not registered", name)
		}
	}
}

// withArgs copies args with one key replaced, so a validation table can start from
// one valid call.
func withArgs(args map[string]any, key string, value any) map[string]any {
	out := make(map[string]any, len(args))
	for k, v := range args {
		out[k] = v
	}
	out[key] = value
	return out
}
