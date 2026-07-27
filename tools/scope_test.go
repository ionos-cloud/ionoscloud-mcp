package tools

import "testing"

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
