package main

import (
	"testing"
)

func TestVersionString(t *testing.T) {
	tests := []struct {
		name    string
		version string
		commit  string
		want    string
	}{
		{"タグ付きバージョン", "v1.2.0", "a1b2c3d", "mytmbar v1.2.0 (a1b2c3d)"},
		{"デフォルト値", "dev", "unknown", "mytmbar dev (unknown)"},
		{"git describe 形式", "v0.1.0-3-gabcdef0", "abcdef0", "mytmbar v0.1.0-3-gabcdef0 (abcdef0)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := versionString(tt.version, tt.commit)
			if got != tt.want {
				t.Errorf("versionString(%q, %q) = %q, want %q", tt.version, tt.commit, got, tt.want)
			}
		})
	}
}
