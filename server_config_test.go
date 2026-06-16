package main

import (
	"runtime/debug"
	"testing"
)

func TestLoadMode(t *testing.T) {
	tests := []struct {
		env  string
		set  bool
		want LoadMode
	}{
		{set: false, want: LoadModeEager}, // unset
		{env: "", set: true, want: LoadModeEager},
		{env: "eager", set: true, want: LoadModeEager},
		{env: "EAGER", set: true, want: LoadModeEager},
		{env: " eager ", set: true, want: LoadModeEager},
		{env: "lazy", set: true, want: LoadModeLazy},
		{env: "Lazy", set: true, want: LoadModeLazy},
		{env: " lazy ", set: true, want: LoadModeLazy},
		{env: "router", set: true, want: LoadModeEager}, // reserved → eager
		{env: "bogus", set: true, want: LoadModeEager},  // unknown → eager
	}

	for _, tt := range tests {
		name := tt.env
		if !tt.set {
			name = "<unset>"
		}
		t.Run(name, func(t *testing.T) {
			if tt.set {
				t.Setenv("IONOS_MCP_LOAD_MODE", tt.env)
			} else {
				t.Setenv("IONOS_MCP_LOAD_MODE", "")
			}
			if got := loadMode(); got != tt.want {
				t.Errorf("loadMode() with %q = %q, want %q", tt.env, got, tt.want)
			}
		})
	}
}

func TestLoadModeLabel(t *testing.T) {
	t.Setenv("IONOS_MCP_LOAD_MODE", "lazy")
	if got := loadModeLabel(); got != "lazy" {
		t.Errorf("loadModeLabel() = %q, want %q", got, "lazy")
	}
}

func TestResolveVersion(t *testing.T) {
	tests := []struct {
		name string
		info *debug.BuildInfo
		ok   bool
		want string
	}{
		{name: "no build info", info: nil, ok: false, want: "dev"},
		{
			name: "module version wins",
			info: &debug.BuildInfo{Main: debug.Module{Version: "v1.2.3"}},
			ok:   true,
			want: "v1.2.3",
		},
		{
			name: "devel module falls through to vcs",
			info: &debug.BuildInfo{
				Main:     debug.Module{Version: "(devel)"},
				Settings: []debug.BuildSetting{{Key: "vcs.revision", Value: "abcdef1234567890"}},
			},
			ok:   true,
			want: "abcdef1",
		},
		{
			name: "dirty revision gets suffix",
			info: &debug.BuildInfo{
				Settings: []debug.BuildSetting{
					{Key: "vcs.revision", Value: "abcdef1234567890"},
					{Key: "vcs.modified", Value: "true"},
				},
			},
			ok:   true,
			want: "abcdef1-dirty",
		},
		{
			name: "short revision not truncated",
			info: &debug.BuildInfo{
				Settings: []debug.BuildSetting{{Key: "vcs.revision", Value: "abc"}},
			},
			ok:   true,
			want: "abc",
		},
		{
			name: "no version no vcs",
			info: &debug.BuildInfo{},
			ok:   true,
			want: "dev",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveVersion(tt.info, tt.ok); got != tt.want {
				t.Errorf("resolveVersion() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSdkBundleVersion(t *testing.T) {
	// Under `go test` build info is available and the bundle is a dependency,
	// so this should resolve to a concrete version, never panic.
	if got := sdkBundleVersion(); got == "" {
		t.Error("sdkBundleVersion() returned empty string")
	}
}
