package main

import "testing"

func TestTruncate(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		maxLen int
		want   string
	}{
		{"15文字はそのまま", "123456789012345", 16, "123456789012345"},
		{"ちょうど16文字は省略", "1234567890123456", 16, "123456789012345…"},
		{"17文字は省略", "12345678901234567", 16, "123456789012345…"},
		{"空文字列はそのまま", "", 16, ""},
		{"マルチバイト15文字はそのまま", "あいうえおかきくけこさしすせそ", 16, "あいうえおかきくけこさしすせそ"},
		{"マルチバイト16文字は省略", "あいうえおかきくけこさしすせそた", 16, "あいうえおかきくけこさしすせそ…"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncate(tt.s, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncate(%q, %d) = %q, want %q", tt.s, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestIsShellCommand(t *testing.T) {
	tests := []struct {
		name string
		cmd  string
		want bool
	}{
		{"zsh is shell", "zsh", true},
		{"bash is shell", "bash", true},
		{"sh is shell", "sh", true},
		{"fish is shell", "fish", true},
		{"tcsh is shell", "tcsh", true},
		{"csh is shell", "csh", true},
		{"ksh is shell", "ksh", true},
		{"dash is shell", "dash", true},
		{"nu is shell", "nu", true},
		{"vim is not shell", "vim", false},
		{"nvim is not shell", "nvim", false},
		{"tig is not shell", "tig", false},
		{"go is not shell", "go", false},
		{"python is not shell", "python", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isShellCommand(tt.cmd)
			if got != tt.want {
				t.Errorf("isShellCommand(%q) = %v, want %v", tt.cmd, got, tt.want)
			}
		})
	}
}
