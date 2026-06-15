package main

import (
	"runtime/debug"
	"testing"
)

func TestResolveLoadMode(t *testing.T) {
	tests := []struct {
		name       string
		flag       string
		env        string
		wantMode   LoadMode
		wantSource loadModeSource
	}{
		// Precedence.
		{name: "both empty -> default eager", flag: "", env: "", wantMode: LoadModeEager, wantSource: sourceDefault},
		{name: "env only", flag: "", env: "lazy", wantMode: LoadModeLazy, wantSource: sourceEnv},
		{name: "flag only", flag: "dynamic", env: "", wantMode: LoadModeDynamic, wantSource: sourceFlag},
		{name: "flag beats env", flag: "dynamic", env: "lazy", wantMode: LoadModeDynamic, wantSource: sourceFlag},
		{name: "blank flag falls through to env", flag: "  ", env: "lazy", wantMode: LoadModeLazy, wantSource: sourceEnv},

		// Values and aliases.
		{name: "eager", flag: "eager", wantMode: LoadModeEager, wantSource: sourceFlag},
		{name: "lazy", flag: "lazy", wantMode: LoadModeLazy, wantSource: sourceFlag},
		{name: "dynamic", flag: "dynamic", wantMode: LoadModeDynamic, wantSource: sourceFlag},
		{name: "search alias -> dynamic", flag: "search", wantMode: LoadModeDynamic, wantSource: sourceFlag},

		// Normalization.
		{name: "uppercase", flag: "DYNAMIC", wantMode: LoadModeDynamic, wantSource: sourceFlag},
		{name: "whitespace", env: " Lazy ", wantMode: LoadModeLazy, wantSource: sourceEnv},

		// Fallbacks (still report the input source; parseLoadMode warns).
		{name: "router retired -> eager", flag: "router", wantMode: LoadModeEager, wantSource: sourceFlag},
		{name: "unknown -> eager", env: "bogus", wantMode: LoadModeEager, wantSource: sourceEnv},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMode, gotSource := resolveLoadMode(tt.flag, tt.env)
			if gotMode != tt.wantMode {
				t.Errorf("resolveLoadMode(%q, %q) mode = %q, want %q", tt.flag, tt.env, gotMode, tt.wantMode)
			}
			if gotSource != tt.wantSource {
				t.Errorf("resolveLoadMode(%q, %q) source = %q, want %q", tt.flag, tt.env, gotSource, tt.wantSource)
			}
		})
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
