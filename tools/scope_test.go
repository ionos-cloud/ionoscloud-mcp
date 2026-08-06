package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestParseScope(t *testing.T) {
	tests := []struct {
		name        string
		raw         string
		wantWrite   bool
		wantDestroy bool
	}{
		{"empty defaults to read", "", false, false},
		{"read only", "read", false, false},
		{"write", "write", true, false},
		{"read,write", "read,write", true, false},
		{"destructive implies write", "destructive", true, true},
		{"read,write,destructive", "read,write,destructive", true, true},
		{"case insensitive", "WRITE", true, false},
		{"whitespace tolerated", "  write , destructive ", true, true},
		{"garbage stays read-only", "banana", false, false},
		{"garbage mixed with write keeps write", "write,banana", true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := ParseScope(tt.raw)
			if s.Write != tt.wantWrite || s.Destructive != tt.wantDestroy {
				t.Errorf("ParseScope(%q) = {Write:%v Destructive:%v}, want {Write:%v Destructive:%v}",
					tt.raw, s.Write, s.Destructive, tt.wantWrite, tt.wantDestroy)
			}
		})
	}
}

func TestScopeAllows(t *testing.T) {
	tests := []struct {
		name                 string
		scope                Scope
		read, write, destroy bool
	}{
		{"read-only", Scope{}, true, false, false},
		{"write", Scope{Write: true}, true, true, false},
		{"destructive", Scope{Write: true, Destructive: true}, true, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.scope.Allows(ClassRead); got != tt.read {
				t.Errorf("Allows(ClassRead) = %v, want %v", got, tt.read)
			}
			if got := tt.scope.Allows(ClassWrite); got != tt.write {
				t.Errorf("Allows(ClassWrite) = %v, want %v", got, tt.write)
			}
			if got := tt.scope.Allows(ClassDestructive); got != tt.destroy {
				t.Errorf("Allows(ClassDestructive) = %v, want %v", got, tt.destroy)
			}
		})
	}
}

func TestReadCannotBeDisabled(t *testing.T) {
	for _, s := range []Scope{{}, {Write: true}, {Write: true, Destructive: true}} {
		if !s.Allows(ClassRead) {
			t.Errorf("scope %v disallowed reads; reads must always be allowed", s)
		}
	}
}

func TestClassFromName(t *testing.T) {
	tests := map[string]Class{
		"list_datacenters":  ClassRead,
		"get_datacenter":    ClassRead,
		"head_object":       ClassRead,
		"create_datacenter": ClassWrite,
		"update_datacenter": ClassWrite,
		"delete_datacenter": ClassDestructive,
		"weird_name":        ClassRead,
	}
	for name, want := range tests {
		if got := ClassFromName(name); got != want {
			t.Errorf("ClassFromName(%q) = %v, want %v", name, got, want)
		}
	}
}

// TestActionVerbsAreNotPrefixesOfEachOther guards the invariant documented on
// actionVerbs: ClassFromName iterates the map in unspecified order, so if one
// verb were a prefix of another the classification would be non-deterministic.
func TestActionVerbsAreNotPrefixesOfEachOther(t *testing.T) {
	for a := range actionVerbs {
		for b := range actionVerbs {
			if a != b && strings.HasPrefix(a, b) {
				t.Errorf("action verb %q is a prefix of %q; classification would depend on map iteration order", b, a)
			}
		}
	}
}

// TestClassFromNameActionVerbs is the test that keeps the two readers of
// actionVerbs honest. Every verb in the table must classify through
// ClassFromName to the same class the registration gate would use — otherwise
// the dynamic dispatcher's defence-in-depth check silently no-ops for that verb.
func TestClassFromNameActionVerbs(t *testing.T) {
	if len(actionVerbs) == 0 {
		t.Fatal("actionVerbs is empty; the action gate would classify every action tool as read")
	}
	for verb, want := range actionVerbs {
		name := verb + "server"
		if got := ClassFromName(name); got != want {
			t.Errorf("ClassFromName(%q) = %v, want %v (from actionVerbs[%q])", name, got, want, verb)
		}
		if got := (Action{Verb: verb}).Class(); got != want {
			t.Errorf("Action{Verb:%q}.Class() = %v, want %v", verb, got, want)
		}
	}
}

// TestActionVerbsCoverDestructivePowerOps pins the classifications that are easy
// to get wrong: a POST that is destructive, and a DELETE that is not a deletion.
func TestActionVerbsCoverDestructivePowerOps(t *testing.T) {
	destructive := []string{"stop_server", "reboot_server", "suspend_server", "upgrade_server", "restore_volume_snapshot", "detach_server_volume"}
	for _, name := range destructive {
		if got := ClassFromName(name); got != ClassDestructive {
			t.Errorf("ClassFromName(%q) = %v, want destructive", name, got)
		}
	}
	write := []string{"start_server", "resume_server", "attach_server_volume", "assign_server_security_groups"}
	for _, name := range write {
		if got := ClassFromName(name); got != ClassWrite {
			t.Errorf("ClassFromName(%q) = %v, want write", name, got)
		}
	}
}

func TestActionAnnotations(t *testing.T) {
	stop := Action{Verb: "stop_", Method: MethodPost, Idempotent: true}
	a := stop.annotations()
	if a.ReadOnlyHint {
		t.Error("stop_ annotations should not be read-only")
	}
	if a.DestructiveHint == nil || !*a.DestructiveHint {
		t.Error("stop_ is a destructive POST; DestructiveHint should be true even though the method is POST")
	}
	if !a.IdempotentHint {
		t.Error("stop_ declared Idempotent; IdempotentHint should be true")
	}

	reboot := Action{Verb: "reboot_", Method: MethodPost}
	if reboot.annotations().IdempotentHint {
		t.Error("reboot_ did not declare Idempotent; IdempotentHint should be false")
	}

	attach := Action{Verb: "attach_", Method: MethodPost, Idempotent: true}
	if d := attach.annotations().DestructiveHint; d == nil || *d {
		t.Error("attach_ is a write-class action; DestructiveHint should be false")
	}
}

// dummyInput is a minimal handler input for registration tests.
type dummyInput struct {
	ID string `json:"id"`
}

func dummyHandler(context.Context, *mcp.CallToolRequest, dummyInput) (*mcp.CallToolResult, any, error) {
	return nil, nil, nil
}

func newTestServer() *mcp.Server {
	return mcp.NewServer(&mcp.Implementation{Name: "scope-test", Version: "0"}, nil)
}

// mustPanic runs fn and fails unless it panicked with a message containing want.
func mustPanic(t *testing.T, want string, fn func()) {
	t.Helper()
	defer func() {
		r := recover()
		if r == nil {
			t.Fatalf("expected a panic containing %q, got none", want)
		}
		if msg, ok := r.(string); !ok || !strings.Contains(msg, want) {
			t.Fatalf("panic = %v, want a message containing %q", r, want)
		}
	}()
	fn()
}

func TestRegisterActionToolPanics(t *testing.T) {
	// A verb absent from actionVerbs would classify as read in ClassFromName,
	// so registration must refuse it rather than register an unclassified tool.
	mustPanic(t, "unknown action verb", func() {
		RegisterActionTool(newTestServer(), Scope{Write: true, Destructive: true},
			Action{Verb: "frobnicate_", Method: MethodPost},
			&mcp.Tool{Name: "frobnicate_server"}, dummyHandler)
	})

	// A name that does not carry its declared verb would be classified by
	// whatever its actual prefix implies — here, read.
	mustPanic(t, "does not start with its declared action verb", func() {
		RegisterActionTool(newTestServer(), Scope{Write: true, Destructive: true},
			Action{Verb: "stop_", Method: MethodPost},
			&mcp.Tool{Name: "halt_server"}, dummyHandler)
	})
}

func TestRegisterActionToolAnnotatesOnRegistration(t *testing.T) {
	tool := &mcp.Tool{Name: "stop_server"}
	RegisterActionTool(newTestServer(), Scope{Write: true, Destructive: true},
		Action{Verb: "stop_", Method: MethodPost, Idempotent: true}, tool, dummyHandler)
	if tool.Annotations == nil {
		t.Fatal("RegisterActionTool did not set annotations")
	}
	if d := tool.Annotations.DestructiveHint; d == nil || !*d {
		t.Error("registered stop_server should carry DestructiveHint=true")
	}
}

// TestRegisterActionToolScopeGate asserts the gate hides destructive actions
// under a write-only scope: a write-class action registers, a destructive one
// does not, and neither is annotated when skipped.
func TestRegisterActionToolScopeGate(t *testing.T) {
	tests := []struct {
		name    string
		scope   Scope
		action  Action
		tool    string
		wantReg bool
	}{
		{"read-only hides write action", Scope{}, Action{Verb: "start_", Method: MethodPost}, "start_server", false},
		{"read-only hides destructive action", Scope{}, Action{Verb: "stop_", Method: MethodPost}, "stop_server", false},
		{"write shows write action", Scope{Write: true}, Action{Verb: "start_", Method: MethodPost}, "start_server", true},
		{"write hides destructive action", Scope{Write: true}, Action{Verb: "stop_", Method: MethodPost}, "stop_server", false},
		{"destructive shows destructive action", Scope{Write: true, Destructive: true}, Action{Verb: "stop_", Method: MethodPost}, "stop_server", true},
		{"destructive shows detach", Scope{Write: true, Destructive: true}, Action{Verb: "detach_", Method: MethodDelete}, "detach_server_volume", true},
		{"write hides detach", Scope{Write: true}, Action{Verb: "detach_", Method: MethodDelete}, "detach_server_volume", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool := &mcp.Tool{Name: tt.tool}
			RegisterActionTool(newTestServer(), tt.scope, tt.action, tool, dummyHandler)
			// Annotations are set only on the registration path, so their
			// presence is a proxy for "was registered" that needs no session.
			gotReg := tool.Annotations != nil
			if gotReg != tt.wantReg {
				t.Errorf("registered = %v, want %v (scope %s, class %s)", gotReg, tt.wantReg, tt.scope, tt.action.Class())
			}
		})
	}
}

func TestMethodClass(t *testing.T) {
	tests := map[Method]Class{
		MethodGet:    ClassRead,
		MethodHead:   ClassRead,
		MethodPost:   ClassWrite,
		MethodPut:    ClassWrite,
		MethodPatch:  ClassWrite,
		MethodDelete: ClassDestructive,
	}
	for m, want := range tests {
		if got := m.Class(); got != want {
			t.Errorf("Method(%q).Class() = %v, want %v", m, got, want)
		}
	}
}

func TestNameMatchesMethod(t *testing.T) {
	ok := []struct {
		name string
		m    Method
	}{
		{"list_x", MethodGet}, {"get_x", MethodGet}, {"head_x", MethodHead},
		{"create_x", MethodPost}, {"update_x", MethodPut}, {"update_x", MethodPatch},
		{"delete_x", MethodDelete},
	}
	for _, c := range ok {
		if !nameMatchesMethod(c.name, c.m) {
			t.Errorf("nameMatchesMethod(%q, %q) = false, want true", c.name, c.m)
		}
	}
	bad := []struct {
		name string
		m    Method
	}{
		{"create_x", MethodGet}, {"list_x", MethodPost}, {"delete_x", MethodPost}, {"get_x", MethodDelete},
	}
	for _, c := range bad {
		if nameMatchesMethod(c.name, c.m) {
			t.Errorf("nameMatchesMethod(%q, %q) = true, want false", c.name, c.m)
		}
	}
}

func TestMethodAnnotations(t *testing.T) {
	if a := MethodGet.annotations(); !a.ReadOnlyHint {
		t.Error("GET annotations should set ReadOnlyHint")
	}
	if a := MethodPost.annotations(); a.ReadOnlyHint || a.IdempotentHint {
		t.Error("POST annotations should be non-read-only and non-idempotent")
	}
	if a := MethodPatch.annotations(); a.ReadOnlyHint || !a.IdempotentHint {
		t.Error("PATCH annotations should be non-read-only and idempotent")
	}
	if a := MethodDelete.annotations(); a.DestructiveHint == nil || !*a.DestructiveHint {
		t.Error("DELETE annotations should set DestructiveHint=true")
	}
}
