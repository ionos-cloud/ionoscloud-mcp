package main

import (
	"os"
	"runtime/debug"
	"strings"
)

const serverName = "ionos-cloud-mcp"

// serverVersion is resolved from three sources in priority order:
//  1. -ldflags "-X main.serverVersion=<tag>"  — set by GoReleaser at release
//  2. info.Main.Version                       — set by `go install <url>@<ver>`
//  3. info.Settings vcs.revision              — local-checkout fallback
//
// The init() below fills in (2) and (3); a release-time ldflag value pre-empts
// init() because the linker writes the constant before init() runs.
var serverVersion string

func init() {
	if serverVersion != "" {
		return
	}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		serverVersion = "dev"
		return
	}
	if v := info.Main.Version; v != "" && v != "(devel)" {
		serverVersion = v
		return
	}
	var revision, modified string
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			revision = s.Value
		case "vcs.modified":
			modified = s.Value
		}
	}
	if revision != "" {
		if len(revision) > 12 {
			revision = revision[:12]
		}
		if modified == "true" {
			revision += "-dirty"
		}
		serverVersion = revision
		return
	}
	serverVersion = "dev"
}

// sdkBundleVersion returns the resolved version of the IONOS SDK bundle's
// shared package, read from the embedded build info. Returns "unknown"
// when the binary was built without module information (e.g. `go run`).
func sdkBundleVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}
	for _, dep := range info.Deps {
		if dep.Path == "github.com/ionos-cloud/sdk-go-bundle/shared" {
			return strings.TrimPrefix(dep.Version, "v")
		}
	}
	return "unknown"
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

// loadModeLabel returns the User-Agent token reflecting current load mode.
func loadModeLabel() string {
	if eagerLoad() {
		return "eager"
	}
	return "lazy"
}
