package main

import (
	"fmt"
	"os"
	"regexp"
	"runtime"
	"runtime/debug"
	"strings"
)

const (
	serverName    = "ionos-cloud-mcp"
	serverVersion = "1.0.0"
)

// nonAlphanumHyphen matches characters that are not lowercase letters, digits, or hyphens.
var nonAlphanumHyphen = regexp.MustCompile(`[^a-z0-9-]+`)

// sanitizeClientName converts a raw MCP clientInfo.name into a safe UA token:
// lowercase, whitespace stripped, non-[a-z0-9-] characters replaced with "-".
// Returns an empty string if the result is empty.
func sanitizeClientName(name string) string {
	s := strings.ToLower(strings.TrimSpace(name))
	s = nonAlphanumHyphen.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}

// buildUserAgent constructs the User-Agent string for outbound API calls.
// host is the sanitized MCP client name (e.g. "claude-desktop"); omitted when empty.
// transport is the active MCP transport (e.g. "stdio"); omitted when empty.
func buildUserAgent(host, transport string) string {
	bundleVersion := "unknown"
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, dep := range info.Deps {
			if dep.Path == "github.com/ionos-cloud/sdk-go-bundle/shared" {
				bundleVersion = strings.TrimPrefix(dep.Version, "v")
				break
			}
		}
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "ionos-cloud-mcp/%s", serverVersion)
	if host != "" {
		fmt.Fprintf(&sb, "_host/%s", host)
	}
	if transport != "" {
		fmt.Fprintf(&sb, "_transport/%s", transport)
	}
	fmt.Fprintf(&sb, "_ionos-cloud-sdk-go-bundle/%s_os/%s_arch/%s",
		bundleVersion, runtime.GOOS, runtime.GOARCH)
	return sb.String()
}

// eagerLoad reports whether the server should register every tool at startup
// instead of gating Compute and Object Storage behind ionos_load_*_tools.
// Set IONOS_MCP_EAGER_LOAD=true for MCP clients that do not honour
// notifications/tools/list_changed (some Claude Desktop tool-search caches,
// claude.ai connectors, Claude in Chrome, custom agents, etc.). Default off
// preserves the lazy behaviour introduced in PR #17 for clients that do.
func eagerLoad() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("IONOS_MCP_EAGER_LOAD"))) {
	case "1", "true", "yes", "on":
		return true
	}
	return false
}
