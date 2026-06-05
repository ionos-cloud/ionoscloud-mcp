package tools

import (
	"errors"
	"strings"
	"testing"

	"github.com/ionos-cloud/sdk-go-bundle/shared"
)

func TestEnrichSDKError_401(t *testing.T) {
	sdkErr := shared.NewGenericOpenAPIError("401 Unauthorized", []byte(`{"errCode":401,"message":"Invalid token"}`), nil, 401)

	got := enrichSDKError(sdkErr)

	for _, want := range []string{
		"IONOS API 401 Unauthorized",
		"IONOS_TOKEN",
		".mcp.json",
		"restart the MCP client",
		"Invalid token",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("enrichSDKError output missing %q\ngot: %s", want, got)
		}
	}
}

func TestEnrichSDKError_401_TruncatesLargeBody(t *testing.T) {
	body := make([]byte, maxErrorBodyBytes*2)
	for i := range body {
		body[i] = 'x'
	}
	sdkErr := shared.NewGenericOpenAPIError("401", body, nil, 401)

	got := enrichSDKError(sdkErr)

	if !strings.Contains(got, "...") {
		t.Errorf("expected truncation marker '...' in enriched 401 with oversized body")
	}
	// Pull out just the body portion after the marker prefix and verify the
	// embedded body is capped (ellipsis excluded).
	_, after, ok := strings.Cut(got, "Original response: ")
	if !ok {
		t.Fatalf("enriched output missing body marker: %s", got)
	}
	embedded := strings.TrimSuffix(after, "...")
	if len(embedded) > maxErrorBodyBytes {
		t.Errorf("embedded body not truncated: got %d bytes, want <= %d", len(embedded), maxErrorBodyBytes)
	}
}

func TestEnrichSDKError_NonAuthStatusPassthrough(t *testing.T) {
	sdkErr := shared.NewGenericOpenAPIError("500 Internal Server Error", []byte(`{"errCode":500}`), nil, 500)

	got := enrichSDKError(sdkErr)

	if strings.Contains(got, "IONOS_TOKEN") {
		t.Errorf("non-401 status should pass through unchanged, got enriched message: %s", got)
	}
	if got != sdkErr.Error() {
		t.Errorf("expected pass-through of original error\ngot:  %s\nwant: %s", got, sdkErr.Error())
	}
}

func TestEnrichSDKError_403Passthrough(t *testing.T) {
	// 403 is deliberately not enriched yet — IONOS uses 403 for several
	// distinct causes (wrong contract, missing role, resource ACL) and a
	// generic "IONOS_TOKEN" hint would mislead the LLM.
	sdkErr := shared.NewGenericOpenAPIError("403 Forbidden", []byte(`{"errCode":403}`), nil, 403)

	got := enrichSDKError(sdkErr)

	if strings.Contains(got, "IONOS_TOKEN") {
		t.Errorf("403 should not be enriched with token guidance, got: %s", got)
	}
}

func TestEnrichSDKError_NonSDKError(t *testing.T) {
	plain := errors.New("network blew up")

	got := enrichSDKError(plain)

	if got != plain.Error() {
		t.Errorf("non-SDK error should pass through, got: %s", got)
	}
}

func TestEnrichSDKError_WrappedSDKError(t *testing.T) {
	sdkErr := shared.NewGenericOpenAPIError("401", []byte("body"), nil, 401)
	wrapped := errors.Join(errors.New("outer context"), sdkErr)

	got := enrichSDKError(wrapped)

	if !strings.Contains(got, "IONOS_TOKEN") {
		t.Errorf("wrapped SDK 401 should still be enriched, got: %s", got)
	}
}
