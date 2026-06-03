package ionosclient

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"", ""},
		{"   ", ""},
		{"Claude Desktop", "claude-desktop"},
		{"claude-ai", "claude-ai"},
		{"cursor", "cursor"},
		{"VS Code", "vs-code"},
		{"Windsurf", "windsurf"},
		{"Claude Code", "claude-code"},
		{"My@App!2", "my-app-2"},
		{"  pad  ", "pad"},
		{"UPPER_CASE", "upper-case"},
		{"---leading-and-trailing---", "leading-and-trailing"},
		{"multiple   spaces", "multiple-spaces"},
		{"unicode-é-é", "unicode"},
		{strings.Repeat("a", maxSegmentLen+10), strings.Repeat("a", maxSegmentLen)},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := sanitizeName(tt.in); got != tt.want {
				t.Errorf("sanitizeName(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestSanitizeVersion(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"", ""},
		{"1.2.3", "1.2.3"},
		{"v1.2.3", "v1.2.3"},
		{"1.2.3+build.7", "1.2.3+build.7"},
		{"2024-11-05", "2024-11-05"},
		{"  1.0  ", "1.0"},
		{"weird/version*", "weird-version"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := sanitizeVersion(tt.in); got != tt.want {
				t.Errorf("sanitizeVersion(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestNewStaticUserAgent(t *testing.T) {
	ua := New(Options{
		Product:          "ionos-cloud-mcp",
		Version:          "1.0.0",
		SDKBundleVersion: "0.1.6",
		Transport:        "stdio",
		Mode:             "lazy",
		GOOS:             "linux",
		GOARCH:           "amd64",
	})
	got := ua.String()
	for _, want := range []string{
		"ionos-cloud-mcp/1.0.0",
		"_transport/stdio",
		"_mode/lazy",
		"_ionos-cloud-sdk-go-bundle/0.1.6",
		"_os/linux",
		"_arch/amd64",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("UA %q missing segment %q", got, want)
		}
	}
	for _, banned := range []string{"_host/", "_host-version/", "_protocol/"} {
		if strings.Contains(got, banned) {
			t.Errorf("UA %q must not carry dynamic segment %q pre-handshake", got, banned)
		}
	}
}

func TestSetClientAddsDynamicSegments(t *testing.T) {
	ua := New(Options{
		Product:          "ionos-cloud-mcp",
		Version:          "1.0.0",
		SDKBundleVersion: "0.1.6",
		Transport:        "stdio",
		Mode:             "eager",
		GOOS:             "darwin",
		GOARCH:           "arm64",
	})
	ua.SetClient("Claude Code", "1.0.42", "2024-11-05")

	got := ua.String()
	for _, want := range []string{
		"_host/claude-code",
		"_host-version/1.0.42",
		"_protocol/2024-11-05",
		"_transport/stdio",
		"_mode/eager",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("UA %q missing segment %q", got, want)
		}
	}

	// Order: dynamic segments belong between prefix and suffix.
	if !strings.HasPrefix(got, "ionos-cloud-mcp/1.0.0") {
		t.Errorf("UA %q must start with product/version", got)
	}
	if !strings.HasSuffix(got, "_arch/arm64") {
		t.Errorf("UA %q must end with arch segment", got)
	}
}

func TestSetClientOmitsEmptyFields(t *testing.T) {
	ua := New(Options{Product: "p", Version: "v", Transport: "stdio"})
	ua.SetClient("", "", "")

	got := ua.String()
	for _, banned := range []string{"_host/", "_host-version/", "_protocol/"} {
		if strings.Contains(got, banned) {
			t.Errorf("UA %q must omit %q when input is empty", got, banned)
		}
	}
}

func TestSetClientReplacesPreviousValue(t *testing.T) {
	ua := New(Options{Product: "p", Version: "v"})
	ua.SetClient("cursor", "1.0", "2024-11-05")
	first := ua.String()
	ua.SetClient("claude-code", "2.0", "2025-03-01")
	second := ua.String()

	if first == second {
		t.Fatalf("expected UA to change after second SetClient; got %q twice", first)
	}
	if strings.Contains(second, "_host/cursor") {
		t.Errorf("UA %q still carries previous host", second)
	}
	if !strings.Contains(second, "_host/claude-code") {
		t.Errorf("UA %q missing new host", second)
	}
}

type captureRT struct {
	headers []string
}

func (c *captureRT) RoundTrip(req *http.Request) (*http.Response, error) {
	c.headers = append(c.headers, req.Header.Get("User-Agent"))
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader("")),
		Header:     make(http.Header),
		Request:    req,
	}, nil
}

func TestTransportInjectsCurrentUA(t *testing.T) {
	ua := New(Options{Product: "ionos-cloud-mcp", Version: "1.0.0", Transport: "stdio"})
	capture := &captureRT{}
	wrapped := ua.Transport(capture)

	req, err := http.NewRequest(http.MethodGet, "https://api.ionos.com/cloudapi/v6/datacenters", nil)
	if err != nil {
		t.Fatal(err)
	}
	// Simulate the SDK seeding its own UA; the RoundTripper must overwrite it.
	req.Header.Set("User-Agent", "sdk-go-bundle/products/compute/v2.0.5")
	if _, err := wrapped.RoundTrip(req); err != nil {
		t.Fatal(err)
	}

	if len(capture.headers) != 1 {
		t.Fatalf("expected 1 captured request, got %d", len(capture.headers))
	}
	if !strings.HasPrefix(capture.headers[0], "ionos-cloud-mcp/1.0.0") {
		t.Errorf("expected wrapped UA, got %q", capture.headers[0])
	}
}

func TestTransportReflectsLiveSetClient(t *testing.T) {
	ua := New(Options{Product: "ionos-cloud-mcp", Version: "1.0.0"})
	capture := &captureRT{}
	wrapped := ua.Transport(capture)

	send := func() {
		req, _ := http.NewRequest(http.MethodGet, "https://api.ionos.com/", nil)
		if _, err := wrapped.RoundTrip(req); err != nil {
			t.Fatal(err)
		}
	}

	send()
	ua.SetClient("claude-code", "1.0.0", "2024-11-05")
	send()

	if len(capture.headers) != 2 {
		t.Fatalf("expected 2 captured requests, got %d", len(capture.headers))
	}
	if strings.Contains(capture.headers[0], "_host/") {
		t.Errorf("first request must not carry host segment: %q", capture.headers[0])
	}
	if !strings.Contains(capture.headers[1], "_host/claude-code") {
		t.Errorf("second request must reflect SetClient: %q", capture.headers[1])
	}
}

func TestTransportNilBaseFallsBackToDefault(t *testing.T) {
	ua := New(Options{Product: "p", Version: "v"})
	rt := ua.Transport(nil)
	if rt == nil {
		t.Fatal("Transport(nil) returned nil")
	}
}
