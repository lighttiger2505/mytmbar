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

func TestIsSpecialCommand(t *testing.T) {
	defaultCmds := []string{"zsh", "bash", "vim", "nvim", "tig"}

	tests := []struct {
		name        string
		cmd         string
		specialCmds []string
		want        bool
	}{
		{"zsh is special", "zsh", defaultCmds, true},
		{"bash is special", "bash", defaultCmds, true},
		{"vim is special", "vim", defaultCmds, true},
		{"nvim is special", "nvim", defaultCmds, true},
		{"tig is special", "tig", defaultCmds, true},
		{"go is not special", "go", defaultCmds, false},
		{"python is not special", "python", defaultCmds, false},
		{"custom cmd in custom list", "htop", []string{"htop", "top"}, true},
		{"cmd not in custom list", "vim", []string{"htop", "top"}, false},
		{"empty list", "zsh", []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isSpecialCommand(tt.cmd, tt.specialCmds)
			if got != tt.want {
				t.Errorf("isSpecialCommand(%q, %v) = %v, want %v", tt.cmd, tt.specialCmds, got, tt.want)
			}
		})
	}
}
