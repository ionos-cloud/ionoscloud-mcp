package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
)

const (
	serverName    = "ionoscloud-mcp"
	serverVersion = "1.0.0"
)

func buildUserAgent() string {
	bundleVersion := "unknown"
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, dep := range info.Deps {
			if dep.Path == "github.com/ionos-cloud/sdk-go-bundle/shared" {
				bundleVersion = strings.TrimPrefix(dep.Version, "v")
				break
			}
		}
	}

	return fmt.Sprintf("ionoscloud-mcp/%s_ionos-cloud-sdk-go-bundle/%s_os/%s_arch/%s",
		serverVersion, bundleVersion, runtime.GOOS, runtime.GOARCH)
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
