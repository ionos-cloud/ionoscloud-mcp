package main

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
)

func TestSanitizeClientName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Claude Desktop", "claude-desktop"},
		{"claude-ai", "claude-ai"},
		{"cursor", "cursor"},
		{"VS Code", "vs-code"},
		{"Windsurf", "windsurf"},
		{"Claude Code", "claude-code"},
		{"", ""},
		{"   ", ""},
		{"My@App!2", "my-app-2"},
		{"  Leading and trailing  ", "leading-and-trailing"},
		{"UPPER_CASE", "upper-case"},
		{"multiple---hyphens", "multiple---hyphens"},
		{"---leading-hyphens---", "leading-hyphens"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%q", tt.input), func(t *testing.T) {
			got := sanitizeClientName(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeClientName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestBuildUserAgent(t *testing.T) {
	tests := []struct {
		host      string
		transport string
		wantParts []string
		notParts  []string
	}{
		{
			host:      "",
			transport: "stdio",
			wantParts: []string{
				"ionos-cloud-mcp/" + serverVersion,
				"_transport/stdio_",
				"_ionos-cloud-sdk-go-bundle/",
				"_os/" + runtime.GOOS,
				"_arch/" + runtime.GOARCH,
			},
			notParts: []string{"_host/"},
		},
		{
			host:      "claude-desktop",
			transport: "stdio",
			wantParts: []string{
				"ionos-cloud-mcp/" + serverVersion,
				"_host/claude-desktop_",
				"_transport/stdio_",
				"_ionos-cloud-sdk-go-bundle/",
				"_os/" + runtime.GOOS,
				"_arch/" + runtime.GOARCH,
			},
		},
		{
			host:      "",
			transport: "",
			wantParts: []string{
				"ionos-cloud-mcp/" + serverVersion,
				"_ionos-cloud-sdk-go-bundle/",
			},
			notParts: []string{"_host/", "_transport/"},
		},
		{
			host:      "cursor",
			transport: "streamable-http",
			wantParts: []string{
				"ionos-cloud-mcp/" + serverVersion,
				"_host/cursor_",
				"_transport/streamable-http_",
			},
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("host=%q transport=%q", tt.host, tt.transport), func(t *testing.T) {
			ua := buildUserAgent(tt.host, tt.transport)
			for _, part := range tt.wantParts {
				if !strings.Contains(ua, part) {
					t.Errorf("buildUserAgent(%q, %q) = %q, missing expected part %q", tt.host, tt.transport, ua, part)
				}
			}
			for _, part := range tt.notParts {
				if strings.Contains(ua, part) {
					t.Errorf("buildUserAgent(%q, %q) = %q, should not contain %q", tt.host, tt.transport, ua, part)
				}
			}
			// Must start with the correct product token and never contain the old name.
			if !strings.HasPrefix(ua, "ionos-cloud-mcp/") {
				t.Errorf("UA %q does not start with ionos-cloud-mcp/", ua)
			}
			if strings.Contains(ua, "ionoscloud-mcp") {
				t.Errorf("UA %q still contains deprecated ionoscloud-mcp token", ua)
			}
		})
	}
}
