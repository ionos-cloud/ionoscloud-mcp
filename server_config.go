package main

import (
	"log"
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
	serverVersion = resolveVersion(info, ok)
}

// resolveVersion derives the server version from build info, following the
// priority order documented on serverVersion. It is a pure function (no globals,
// no I/O) so the resolution rules can be unit-tested; init() supplies the live
// build info. A non-empty ldflag value pre-empts this entirely (init returns
// early before calling it).
func resolveVersion(info *debug.BuildInfo, ok bool) string {
	if !ok || info == nil {
		return "dev"
	}
	if v := info.Main.Version; v != "" && v != "(devel)" {
		return v
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
		if len(revision) > 7 {
			revision = revision[:7]
		}
		if modified == "true" {
			revision += "-dirty"
		}
		return revision
	}
	return "dev"
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

// LoadMode selects how the server exposes tools to MCP clients.
type LoadMode string

const (
	// LoadModeEager registers all tools at startup. Default. Optimal for
	// Claude Code (defers schemas client-side via ToolSearch) and required
	// for clients that ignore notifications/tools/list_changed.
	LoadModeEager LoadMode = "eager"

	// LoadModeLazy defers Compute and Object Storage behind ionos_load_*_tools
	// sentinel tools. Requires MCP client support for
	// notifications/tools/list_changed.
	LoadModeLazy LoadMode = "lazy"

	// LoadModeRouter is reserved for a future search + invoke pattern targeting
	// clients with hard tool caps (Cursor 40, Windsurf 100) or no client-side
	// schema deferral. Currently falls back to eager with a stderr warning.
	LoadModeRouter LoadMode = "router"
)

// loadMode returns the tool registration strategy selected via
// IONOS_MCP_LOAD_MODE. Default: eager. Unknown values and the reserved
// 'router' value log a warning to stderr and fall back to eager.
func loadMode() LoadMode {
	v := strings.ToLower(strings.TrimSpace(os.Getenv("IONOS_MCP_LOAD_MODE")))
	switch v {
	case "", "eager":
		return LoadModeEager
	case "lazy":
		return LoadModeLazy
	case "router":
		log.Println("IONOS_MCP_LOAD_MODE=router not yet implemented; falling back to eager")
		return LoadModeEager
	default:
		log.Printf("IONOS_MCP_LOAD_MODE=%q unrecognised; falling back to eager", v)
		return LoadModeEager
	}
}

// loadModeLabel returns the User-Agent token reflecting current load mode.
func loadModeLabel() string {
	return string(loadMode())
}
