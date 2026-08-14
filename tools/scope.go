package tools

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Class is a tool's mutation class, which decides the scope it needs.
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

// Scope is the set of mutation classes enabled via IONOS_MCP_TOOL_SCOPE.
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

// ParseScope reads IONOS_MCP_TOOL_SCOPE into a Scope. Hierarchical: destructive
// implies write. Unrecognised values leave the server read-only.
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

// annotations returns the MCP annotations implied by the HTTP method.
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

// actionVerbs maps a non-CRUD tool's name prefix to its mutation class. Power
// control, snapshot restore and attach/detach read better as domain verbs than as
// create_/delete_.
//
// Single source of truth for action classification, read by both RegisterActionTool
// and ClassFromName. No verb may be a prefix of another, so at most one can match.
var actionVerbs = map[string]Class{
	// Mutating but recoverable: they add or resume, never discard.
	"start_":    ClassWrite,
	"resume_":   ClassWrite,
	"attach_":   ClassWrite,
	"assign_":   ClassWrite,
	"stop_":     ClassDestructive,
	"reboot_":   ClassDestructive,
	"suspend_":  ClassDestructive,
	"upgrade_":  ClassDestructive,
	"restore_":  ClassDestructive,
	"detach_":   ClassDestructive,
	"recreate_": ClassDestructive,
}

// Action describes a non-CRUD operation named with a domain verb. Method is used
// only to issue the request; the mutation class comes from the verb, because the two
// disagree in both directions.
type Action struct {
	Verb       string // name prefix; must be a key of actionVerbs
	Method     Method // HTTP method used for the request
	Idempotent bool   // safe to repeat with the same effect (start/stop yes, reboot/upgrade no)
}

// Class returns the mutation class the verb implies.
func (a Action) Class() Class { return actionVerbs[a.Verb] }

// annotations returns the MCP annotations implied by the action's verb.
func (a Action) annotations() *mcp.ToolAnnotations {
	return &mcp.ToolAnnotations{
		ReadOnlyHint:    false,
		DestructiveHint: boolPtr(a.Class() == ClassDestructive),
		IdempotentHint:  a.Idempotent,
	}
}

// ClassFromName derives a class from a tool name alone, for the dynamic dispatcher
// which sees names rather than registrations.
func ClassFromName(name string) Class {
	// Action verbs first: a CRUD-prefix heuristic would misread them as reads.
	for verb, class := range actionVerbs {
		if strings.HasPrefix(name, verb) {
			return class
		}
	}
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

// RegisterTool is the single choke point for scope gating. Skips registration
// entirely when the scope disallows the class, so a gated tool never appears in
// tools/list. Panics if the name prefix and HTTP method disagree.
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

// RegisterActionTool is RegisterTool for domain-verb operations, classifying by
// verb instead of by method. Panics if the verb is absent from actionVerbs.
func RegisterActionTool[In, Out any](s *mcp.Server, sc Scope, a Action, t *mcp.Tool, h mcp.ToolHandlerFor[In, Out]) {
	if _, ok := actionVerbs[a.Verb]; !ok {
		panic(fmt.Sprintf("RegisterActionTool: tool %q declares unknown action verb %q; add it to actionVerbs with its class", t.Name, a.Verb))
	}
	if !strings.HasPrefix(t.Name, a.Verb) {
		panic(fmt.Sprintf("RegisterActionTool: tool %q name does not start with its declared action verb %q", t.Name, a.Verb))
	}
	if !sc.Allows(a.Class()) {
		return // gate: not permitted by the current scope — never registered
	}
	t.Annotations = a.annotations()
	mcp.AddTool(s, t, h)
}

func boolPtr(b bool) *bool { return &b }
