package main

import "testing"

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
