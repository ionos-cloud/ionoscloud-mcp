package tools

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Class is the mutation class of a tool, derived from its HTTP method. It is the
// unit the scope gate reasons about: reads are always allowed; writes and
// destructive operations must be explicitly enabled via IONOS_MCP_TOOL_SCOPE.
type Class int

const (
	// ClassRead is a non-mutating tool (HTTP GET/HEAD; list_/get_/head_). Always allowed.
	ClassRead Class = iota
	// ClassWrite creates or updates a resource (HTTP POST/PUT/PATCH; create_/update_).
	ClassWrite
	// ClassDestructive deletes a resource (HTTP DELETE; delete_).
	ClassDestructive
)

func (c Class) String() string {
	switch c {
	case ClassWrite:
		return "write"
	case ClassDestructive:
		return "destructive"
	default:
		return "read"
	}
}

// Scope is the set of mutation classes an operator has enabled through
// IONOS_MCP_TOOL_SCOPE. Reads are always permitted; Write and Destructive are
// opt-in. The tokens are hierarchical: "destructive" implies "write".
type Scope struct {
	Write       bool
	Destructive bool
}

// Allows reports whether tools of the given class may be registered/invoked.
// Reads are unconditionally allowed — read access can never be turned off.
func (s Scope) Allows(c Class) bool {
	switch c {
	case ClassWrite:
		return s.Write
	case ClassDestructive:
		return s.Destructive
	default: // ClassRead
		return true
	}
}

func (s Scope) String() string {
	switch {
	case s.Destructive:
		return "read,write,destructive"
	case s.Write:
		return "read,write"
	default:
		return "read"
	}
}

// ParseScope reads the comma-separated IONOS_MCP_TOOL_SCOPE value into a Scope.
// Tokens are case-insensitive and hierarchical: "destructive" enables both
// destructive and write; "write" enables write; "read" (and the empty/default
// value) leave the server read-only. Unrecognised tokens are ignored with a
// warning so a typo can never silently widen access. It is a pure function so
// the precedence rules are unit-testable (mirrors resolveLoadMode).
func ParseScope(raw string) Scope {
	var s Scope
	for _, tok := range strings.Split(raw, ",") {
		switch strings.ToLower(strings.TrimSpace(tok)) {
		case "", "read":
			// read is always on; no-op.
		case "write":
			s.Write = true
		case "destructive":
			s.Destructive = true
			s.Write = true // hierarchy: destructive implies write
		default:
			log.Printf("unrecognised IONOS_MCP_TOOL_SCOPE token %q; valid tokens: read, write, destructive; ignoring", tok)
		}
	}
	return s
}

// Method is the HTTP method a tool maps to. Tools are classified by method at
// registration so the scope gate knows which to hide.
type Method string

const (
	MethodGet    Method = http.MethodGet
	MethodHead   Method = http.MethodHead
	MethodPost   Method = http.MethodPost
	MethodPut    Method = http.MethodPut
	MethodPatch  Method = http.MethodPatch
	MethodDelete Method = http.MethodDelete
)

// Class maps an HTTP method to its mutation class.
func (m Method) Class() Class {
	switch m {
	case MethodPost, MethodPut, MethodPatch:
		return ClassWrite
	case MethodDelete:
		return ClassDestructive
	default: // GET, HEAD
		return ClassRead
	}
}

// annotations returns the MCP tool annotations implied by the method. These are
// advisory hints for clients (so they can build their own confirmation UX);
// enforcement stays server-side in the scope gate and the confirmation store.
func (m Method) annotations() *mcp.ToolAnnotations {
	switch m {
	case MethodPost: // create: mutating, not destructive, not idempotent
		return &mcp.ToolAnnotations{ReadOnlyHint: false, DestructiveHint: boolPtr(false), IdempotentHint: false}
	case MethodPut, MethodPatch: // update: mutating, not destructive, idempotent
		return &mcp.ToolAnnotations{ReadOnlyHint: false, DestructiveHint: boolPtr(false), IdempotentHint: true}
	case MethodDelete: // delete: mutating, destructive, idempotent
		return &mcp.ToolAnnotations{ReadOnlyHint: false, DestructiveHint: boolPtr(true), IdempotentHint: true}
	default: // GET, HEAD: read-only
		return &mcp.ToolAnnotations{ReadOnlyHint: true}
	}
}

// ClassFromName derives a tool's class from its name prefix. The dynamic
// dispatcher only has tool names to work with, so it classifies by the strict
// naming convention (create_/update_ => write, delete_ => destructive, else
// read). RegisterTool guarantees write tools carry the matching prefix.
func ClassFromName(name string) Class {
	switch {
	case strings.HasPrefix(name, "delete_"):
		return ClassDestructive
	case strings.HasPrefix(name, "create_"), strings.HasPrefix(name, "update_"):
		return ClassWrite
	default: // list_/get_/head_ and anything else
		return ClassRead
	}
}

// nameMatchesMethod verifies a tool's name prefix agrees with its HTTP method,
// so a mis-prefixed tool can never be mis-classified. Called at registration.
func nameMatchesMethod(name string, m Method) bool {
	switch m {
	case MethodPost:
		return strings.HasPrefix(name, "create_")
	case MethodPut, MethodPatch:
		return strings.HasPrefix(name, "update_")
	case MethodDelete:
		return strings.HasPrefix(name, "delete_")
	default: // GET, HEAD
		return strings.HasPrefix(name, "list_") ||
			strings.HasPrefix(name, "get_") ||
			strings.HasPrefix(name, "head_")
	}
}

// RegisterTool is the single choke point for registering a tool behind the scope
// gate. It (1) asserts the name matches the method — panicking on mismatch, a
// boot-time coding error caught by tests before the server serves; (2) sets
// method-derived annotations; and (3) registers the tool only if the scope
// permits its class, otherwise skipping it entirely so it never appears in
// tools/list. Generic over In/Out exactly like mcp.AddTool, so the handler's
// typed input struct still drives JSON-schema inference.
func RegisterTool[In, Out any](s *mcp.Server, sc Scope, m Method, t *mcp.Tool, h mcp.ToolHandlerFor[In, Out]) {
	if !nameMatchesMethod(t.Name, m) {
		panic(fmt.Sprintf("RegisterTool: tool %q name prefix does not match HTTP method %s", t.Name, m))
	}
	if !sc.Allows(m.Class()) {
		return // gate: not permitted by the current scope — never registered
	}
	t.Annotations = m.annotations()
	mcp.AddTool(s, t, h)
}

func boolPtr(b bool) *bool { return &b }
