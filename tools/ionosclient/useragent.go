// Package ionosclient builds and maintains the User-Agent string the MCP
// server attaches to outbound IONOS CLOUD API calls.
//
// The User-Agent carries diagnostic metadata: server version, SDK bundle
// version, OS/arch, transport, eager-load mode, and — once the MCP
// initialize handshake completes — the calling host's name, version, and
// negotiated protocol version.
//
// Composition is split into a static prefix built once at startup and a
// dynamic full string updated after handshake. UserAgent stores the
// composed string in an atomic.Pointer so updates are race-free.
//
// UA injection happens at the HTTP boundary via the Transport method,
// which wraps the cfg.HTTPClient.Transport with a RoundTripper that sets
// the User-Agent header on every outbound request. Because the SDK shares
// the underlying *http.Client across every shallow cfg copy (compute,
// DNS, billing, cert, object storage base + regional clients), a single
// RoundTripper covers all clients without chasing cfg snapshots.
package ionosclient

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync/atomic"
)

const (
	maxSegmentLen = 32
)

var (
	// nonNameChar matches anything outside [a-z0-9-] for sanitised name tokens.
	nonNameChar = regexp.MustCompile(`[^a-z0-9-]+`)
	// nonVersionChar matches anything outside [a-z0-9.+-] for sanitised
	// version/protocol tokens. Dots and plus signs are preserved so semver
	// values like "1.2.3+build" stay legible.
	nonVersionChar = regexp.MustCompile(`[^a-z0-9.+-]+`)
)

// Options configures the static portion of the User-Agent that is fixed at
// startup. All fields are best-effort: empty values cause the matching
// segment to be omitted.
type Options struct {
	Product          string // e.g. "ionoscloud-mcp"
	Version          string // server version
	SDKBundleVersion string // detected from build info
	Transport        string // "stdio", "streamable-http"
	Mode             string // "eager", "lazy"
	GOOS             string
	GOARCH           string
}

// UserAgent owns the composed User-Agent string and updates it atomically.
type UserAgent struct {
	staticPrefix string // product/version_transport/_mode/_sdk/_os/_arch
	staticSuffix string // sdk/os/arch portion that always trails dynamic segments
	full         atomic.Pointer[string]
}

// New constructs a UserAgent with the provided static metadata. Before
// SetClient is called (i.e. before the MCP initialize handshake), String
// returns the static portion only.
func New(opts Options) *UserAgent {
	ua := &UserAgent{
		staticPrefix: buildStaticPrefix(opts),
		staticSuffix: buildStaticSuffix(opts),
	}
	initial := ua.staticPrefix + ua.staticSuffix
	ua.full.Store(&initial)
	return ua
}

// SetClient updates the dynamic segments with information from the MCP
// initialize handshake. Inputs are sanitised; empty fields are omitted.
// Safe to call concurrently with String and outbound requests.
func (u *UserAgent) SetClient(name, version, protocol string) {
	host := sanitizeName(name)
	hostVer := sanitizeVersion(version)
	proto := sanitizeVersion(protocol)

	var dyn strings.Builder
	if host != "" {
		fmt.Fprintf(&dyn, "_host/%s", host)
	}
	if hostVer != "" {
		fmt.Fprintf(&dyn, "_host-version/%s", hostVer)
	}
	if proto != "" {
		fmt.Fprintf(&dyn, "_protocol/%s", proto)
	}

	composed := u.staticPrefix + dyn.String() + u.staticSuffix
	u.full.Store(&composed)
}

// String returns the current User-Agent.
func (u *UserAgent) String() string {
	if s := u.full.Load(); s != nil {
		return *s
	}
	return ""
}

// Transport wraps base so that every outbound request carries the current
// User-Agent. If base is nil, http.DefaultTransport is used.
func (u *UserAgent) Transport(base http.RoundTripper) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	return &uaTransport{base: base, ua: u}
}

type uaTransport struct {
	base http.RoundTripper
	ua   *UserAgent
}

func (t *uaTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Header.Set replaces any value the SDK installed via its default UA;
	// this is the single source of truth for outbound User-Agent.
	req.Header.Set("User-Agent", t.ua.String())
	return t.base.RoundTrip(req)
}

func buildStaticPrefix(opts Options) string {
	product := strings.TrimSpace(opts.Product)
	if product == "" {
		product = "ionoscloud-mcp"
	}
	version := strings.TrimSpace(opts.Version)
	if version == "" {
		version = "unknown"
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "%s/%s", product, version)
	if t := sanitizeName(opts.Transport); t != "" {
		fmt.Fprintf(&sb, "_transport/%s", t)
	}
	if m := sanitizeName(opts.Mode); m != "" {
		fmt.Fprintf(&sb, "_mode/%s", m)
	}
	return sb.String()
}

func buildStaticSuffix(opts Options) string {
	sdkVer := strings.TrimSpace(opts.SDKBundleVersion)
	if sdkVer == "" {
		sdkVer = "unknown"
	}
	goos := strings.TrimSpace(opts.GOOS)
	if goos == "" {
		goos = "unknown"
	}
	goarch := strings.TrimSpace(opts.GOARCH)
	if goarch == "" {
		goarch = "unknown"
	}
	return fmt.Sprintf("_ionos-cloud-sdk-go-bundle/%s_os/%s_arch/%s", sdkVer, goos, goarch)
}

// sanitizeName turns an arbitrary string into a User-Agent name token:
// lowercase, ASCII letters/digits/hyphens only, hyphen-trimmed, length-capped.
// Returns an empty string when the input has no usable characters.
func sanitizeName(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = nonNameChar.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > maxSegmentLen {
		s = strings.TrimRight(s[:maxSegmentLen], "-")
	}
	return s
}

// sanitizeVersion is sanitizeName but preserves dots and plus signs so
// semantic version strings (and the MCP protocol version "2024-11-05")
// remain legible.
func sanitizeVersion(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = nonVersionChar.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > maxSegmentLen {
		s = strings.TrimRight(s[:maxSegmentLen], "-")
	}
	return s
}
