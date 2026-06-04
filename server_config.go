package main

import (
	"os"
	"runtime/debug"
	"strings"
)

const serverName = "ionos-cloud-mcp"

// serverVersion is overridden at release time via
// `-ldflags "-X main.serverVersion=<tag>"` by GoReleaser. The default
// applies to local builds (`go build`, `go run`) where no tag is set.
var serverVersion = "dev"

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
