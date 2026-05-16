package main

import "testing"

func TestClaudeStatusEmoji(t *testing.T) {
	tests := []struct {
		state string
		want  string
	}{
		{"Running", "⚙️"},
		{"Idle", "🤖"},
		{"Waiting", "⏳"},
		{"Plan", "📋"},
		{"AcceptEdits", "✏️"},
		{"Unknown", "🤖"},
		{"", "🤖"},
	}

	for _, tt := range tests {
		t.Run(tt.state, func(t *testing.T) {
			got := claudeStatusEmoji(tt.state)
			if got != tt.want {
				t.Errorf("claudeStatusEmoji(%q) = %q, want %q", tt.state, got, tt.want)
			}
		})
	}
}

func TestParseClaudeState(t *testing.T) {
	tests := []struct {
		name  string
		title string
		want  string
	}{
		{
			"running with spinner and time",
			"✢ Clauding… (esc to interrupt · 1m 45s · ↓ 1.2k tokens)",
			"Running",
		},
		{
			"running with dot spinner",
			"· Thinking… (esc to interrupt · 30s)",
			"Running",
		},
		{
			"idle with prompt",
			"✳ claude  ~/projects/foo",
			"Idle",
		},
		{
			"waiting with yes/no",
			"Run this command? (Y/n)",
			"Waiting",
		},
		{
			"plan mode",
			"⏸ plan mode on\n❯ Try something",
			"Plan",
		},
		{
			"accept edits mode",
			"⏵⏵ accept edits on\n❯ Try something",
			"AcceptEdits",
		},
		{
			"unknown title",
			"some random title",
			"Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseClaudeState(tt.title)
			if got != tt.want {
				t.Errorf("parseClaudeState(%q) = %q, want %q", tt.title, got, tt.want)
			}
		})
	}
}

func TestIsClaude(t *testing.T) {
	tests := []struct {
		name    string
		title   string
		cmd     string
		want    bool
	}{
		{"claude cmd with idle title", "✳ claude ~/foo", "claude", true},
		{"node cmd with spinner title", "✢ Clauding…", "node", true},
		{"version cmd with spinner title", "✢ Clauding…", "2.1.34", true},
		{"zsh with no claude title", "zsh", "zsh", false},
		{"claude cmd but no claude title", "just a title", "claude", false},
		{"spinner title but zsh cmd", "✢ Clauding…", "zsh", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isClaude(tt.title, tt.cmd)
			if got != tt.want {
				t.Errorf("isClaude(%q, %q) = %v, want %v", tt.title, tt.cmd, got, tt.want)
			}
		})
	}
}
