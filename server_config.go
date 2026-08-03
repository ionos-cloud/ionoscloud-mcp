package main

import (
	"log"
	"runtime/debug"
	"strings"
	"time"
)

const (
	// httpSessionTimeout closes sessions that stop sending requests
	httpSessionTimeout = 30 * time.Minute
	// httpReadHeaderTimeout bounds how long a client may take to send its headers
	httpReadHeaderTimeout = 10 * time.Second
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

	// LoadModeDynamic exposes only a tiny set of meta-tools (search + describe +
	// call) that let the model discover and invoke the full tool catalog at
	// runtime, without the catalog ever entering the client's tool list. The
	// real tool list never changes, so this needs no client cooperation
	// (no notifications/tools/list_changed). Targets clients with hard tool caps
	// and no client-side tool search of their own (e.g. Cursor 40, Windsurf 100).
	// Claude Code should stay eager — it defers schemas client-side via ToolSearch.
	LoadModeDynamic LoadMode = "dynamic"
)

// loadModeSource describes where an effective load mode came from, for startup
// diagnostics. It does not imply the provided value was valid — parseLoadMode
// logs a warning when it falls back to eager.
type loadModeSource string

const (
	sourceFlag    loadModeSource = "--load-mode flag"
	sourceEnv     loadModeSource = "IONOS_MCP_LOAD_MODE env"
	sourceDefault loadModeSource = "default"
)

// Transport selects how the server communicates with MCP clients.
type Transport string

const (
	// TransportStdio serves the MCP protocol over stdin/stdout. Default —
	// required by clients that spawn the server as a subprocess (Claude
	// Desktop, Claude Code, Cursor, Windsurf, etc.).
	TransportStdio Transport = "stdio"

	// TransportHTTP serves the MCP protocol over the Streamable HTTP
	// transport (a single endpoint accepting POSTed JSON-RPC and, for
	// server->client streaming, text/event-stream). For remote/networked
	// deployments where the client connects over HTTP rather than spawning
	// a subprocess.
	TransportHTTP Transport = "http"
)

// transportSource describes where an effective transport came from, for
// startup diagnostics. It does not imply the provided value was valid —
// parseTransport logs a warning when it falls back to stdio.
type transportSource string

const (
	transportSourceFlag    transportSource = "--transport flag"
	transportSourceEnv     transportSource = "IONOS_MCP_TRANSPORT env"
	transportSourceDefault transportSource = "default"
)

// resolveTransport picks the wire transport from, in priority order, the
// --transport flag value, the IONOS_MCP_TRANSPORT env value, then the
// default (stdio). Each input may be empty (meaning "not provided"). It is a
// pure function so the precedence rules can be unit-tested; callers pass the
// flag value and os.Getenv("IONOS_MCP_TRANSPORT"). The returned source
// reflects which input supplied the value (even if that value was invalid
// and fell back to stdio — parseTransport logs that case).
func resolveTransport(flagVal, envVal string) (Transport, transportSource) {
	if strings.TrimSpace(flagVal) != "" {
		return parseTransport(flagVal), transportSourceFlag
	}
	if strings.TrimSpace(envVal) != "" {
		return parseTransport(envVal), transportSourceEnv
	}
	return TransportStdio, transportSourceDefault
}

// parseTransport normalizes (lowercase + trim) and validates a transport
// string. Any unrecognised value logs an actionable warning and falls back
// to stdio.
func parseTransport(raw string) Transport {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "stdio":
		return TransportStdio
	case "http":
		return TransportHTTP
	default:
		log.Printf("unrecognised transport %q; valid values: stdio, http; falling back to stdio", raw)
		return TransportStdio
	}
}

// resolveLoadMode picks the tool registration strategy from, in priority order,
// the --load-mode flag value, the IONOS_MCP_LOAD_MODE env value, then the
// default (eager). Each input may be empty (meaning "not provided"). It is a
// pure function so the precedence rules can be unit-tested; callers pass the
// flag value and os.Getenv("IONOS_MCP_LOAD_MODE"). The returned source reflects
// which input supplied the value (even if that value was invalid and fell back
// to eager — parseLoadMode logs that case).
func resolveLoadMode(flagVal, envVal string) (LoadMode, loadModeSource) {
	if strings.TrimSpace(flagVal) != "" {
		return parseLoadMode(flagVal), sourceFlag
	}
	if strings.TrimSpace(envVal) != "" {
		return parseLoadMode(envVal), sourceEnv
	}
	return LoadModeEager, sourceDefault
}

// parseLoadMode normalizes (lowercase + trim) and validates a load mode string.
// "search" is accepted as an alias for "dynamic". Any unrecognised value logs an
// actionable warning and falls back to eager.
func parseLoadMode(raw string) LoadMode {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "eager":
		return LoadModeEager
	case "lazy":
		return LoadModeLazy
	case "dynamic", "search":
		return LoadModeDynamic
	default:
		log.Printf("unrecognised load mode %q; valid values: eager, lazy, dynamic (alias: search); falling back to eager", raw)
		return LoadModeEager
	}
}
