package main

import "testing"

func TestClaudeStatusEmoji(t *testing.T) {
	tests := []struct {
		name   string
		status ClaudeStatus
		want   string
	}{
		{"running", ClaudeStatus{State: ClaudeStateRunning}, "🏃"},
		{"idle", ClaudeStatus{State: ClaudeStateIdle}, "⌛"},
		{"waiting", ClaudeStatus{State: ClaudeStateWaiting}, "🚧"},
		{"plan mode idle", ClaudeStatus{Mode: ClaudeModePlan, State: ClaudeStateIdle}, "📋⌛"},
		{"accept edits running", ClaudeStatus{Mode: ClaudeModeAccept, State: ClaudeStateRunning}, "✏️🏃"},
		{"unknown", ClaudeStatus{State: ClaudeStateUnknown}, "❓"},
		{"empty", ClaudeStatus{}, "❓"},
	}

	cfg := defaultConfig()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := claudeStatusEmoji(tt.status, cfg)
			if got != tt.want {
				t.Errorf("claudeStatusEmoji(%+v) = %q, want %q", tt.status, got, tt.want)
			}
		})
	}
}

func TestParseClaudeStatus(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantState ClaudeState
		wantMode  ClaudeMode
	}{
		{
			name: "Idle with prompt only",
			content: `Some output
───────────────────────────────────────
❯
───────────────────────────────────────`,
			wantState: ClaudeStateIdle,
		},
		{
			name: "Idle with completion suggestion",
			content: `Some output
───────────────────────────────────────
❯ Try "edit file.go to..."
───────────────────────────────────────`,
			wantState: ClaudeStateIdle,
		},
		{
			name: "Running with Clauding",
			content: `Some output
✢ Clauding… (esc to interrupt · 1m 45s · ↓ 1.2k tokens)`,
			wantState: ClaudeStateRunning,
		},
		{
			name: "Running with time first format",
			content: `Some output
✢ Reticulating… (1m 52s · ↓ 11.5k tokens · thought for 7s)`,
			wantState: ClaudeStateRunning,
		},
		{
			name: "Running with esc to interrupt at end of status line",
			content: `Some output
✶ Proofing… (thinking)
───────────────────────────────────────
❯
───────────────────────────────────────
  4 files +20 -0 · esc to interrupt`,
			wantState: ClaudeStateRunning,
		},
		{
			name: "Running fallback (ctrl+c to interrupt)",
			content: `Some output
✻ Thinking… (ctrl+c to interrupt)`,
			wantState: ClaudeStateRunning,
		},
		{
			name: "Running fallback (esc to interrupt)",
			content: `Some output
✻ Processing… (esc to interrupt)`,
			wantState: ClaudeStateRunning,
		},
		{
			name: "Not Running when text mentions esc to interrupt in quotes",
			content: `Some output about "esc to interrupt" in quotes
❯ `,
			wantState: ClaudeStateIdle,
		},
		{
			name: "Not Running when indented status line (quoted text)",
			content: `⏺ 現在の内容は：
  ✻ Galloping… (esc to interrupt · 1m 19s · ↓ 5.9k tokens · thinking)

✻ Cooked for 1m 29s

───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
❯
───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────`,
			wantState: ClaudeStateIdle,
		},
		{
			name: "Waiting with permission prompt",
			content: `Some output
Yes, allow once
Yes, allow always`,
			wantState: ClaudeStateWaiting,
		},
		{
			name: "Waiting with confirmation prompt",
			content: `Some output
Continue? (Y/n)`,
			wantState: ClaudeStateWaiting,
		},
		{
			name: "Idle after task completion",
			content: `Some output
✻ Cooked for 43s
───────────────────────────────────────
❯ `,
			wantState: ClaudeStateIdle,
		},
		{
			name: "Idle after task completion with file changes",
			content: `✻ Sautéed for 2m 55s

───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
❯
───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
  4 files +73 -3`,
			wantState: ClaudeStateIdle,
		},
		{
			name: "Idle with plan mode",
			content: `Some output
⏸ plan mode on
───────────────────────────────────────
❯
───────────────────────────────────────`,
			wantState: ClaudeStateIdle,
			wantMode:  ClaudeModePlan,
		},
		{
			name: "Idle with accept edits",
			content: `Some output
⏵⏵ accept edits on
───────────────────────────────────────
❯
───────────────────────────────────────`,
			wantState: ClaudeStateIdle,
			wantMode:  ClaudeModeAccept,
		},
		{
			name: "Waiting - Interview mode",
			content: `  3. ドキュメントのレビュー
  4. Issue の修正
  5. Type something.
  Chat about this
  Skip interview and plan immediately
Enter to select · ↑/↓ to navigate · Esc to cancel`,
			wantState: ClaudeStateWaiting,
		},
		{
			name: "Waiting - Bash command confirmation dialog",
			content: `⏺ Bash(grep --help 2>/dev/null | head -10)
  ⎿  Running…

 Do you want to proceed?
 ❯ 1. Yes
   2. Yes, and don't ask again
   3. No

 Esc to cancel · Tab to amend · ctrl+e to explain`,
			wantState: ClaudeStateWaiting,
		},
		{
			name: "Running with action text and accept edits mode",
			content: `✶ Adding handler types… (ctrl+c to interrupt · ctrl+t to hide todos · 3m 27s · ↑ 11.0k tokens)
  ⎿  ☐ Add handler types

───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
❯
───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
  ⏵⏵ accept edits on (shift+tab to cycle)`,
			wantState: ClaudeStateRunning,
			wantMode:  ClaudeModeAccept,
		},
		{
			name: "Idle with trust dialog overlay (prompt is last meaningful line)",
			content: ` /home/user/projects/myapp

 ❯ 1. Yes, proceed
   2. No, exit

 Enter to confirm · Esc to cancel

───────────────────────────────────────
❯ Try "fix typecheck errors"
───────────────────────────────────────
  ? for shortcuts`,
			wantState: ClaudeStateIdle,
		},
		{
			name: "Idle with file changes status line",
			content: `✻ Churned for 3m 5s

───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
❯
───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
  4 files +42 -0`,
			wantState: ClaudeStateIdle,
		},
		{
			name: "Unknown state",
			content: `Some random output
without any recognizable pattern`,
			wantState: ClaudeStateUnknown,
		},
		{
			name:      "Empty content",
			content:   "",
			wantState: ClaudeStateUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseClaudeStatus(tt.content)
			if got.State != tt.wantState {
				t.Errorf("parseClaudeStatus().State = %v, want %v", got.State, tt.wantState)
			}
			if got.Mode != tt.wantMode {
				t.Errorf("parseClaudeStatus().Mode = %v, want %v", got.Mode, tt.wantMode)
			}
		})
	}
}

func TestClaudeStateShortString(t *testing.T) {
	tests := []struct {
		state ClaudeState
		want  string
	}{
		{ClaudeStateIdle, "I"},
		{ClaudeStateRunning, "R"},
		{ClaudeStateWaiting, "W"},
		{ClaudeStateUnknown, "U"},
	}
	for _, tt := range tests {
		t.Run(tt.state.String(), func(t *testing.T) {
			got := tt.state.ShortString()
			if got != tt.want {
				t.Errorf("ClaudeState(%d).ShortString() = %q, want %q", tt.state, got, tt.want)
			}
		})
	}
}

func TestClaudeWindowStatusShort(t *testing.T) {
	// Idle pane content with ❯ prompt
	idleContent := "───────────────────────────────────────\n❯\n───────────────────────────────────────"
	// Running pane content with spinner
	runningContent := "✢ Clauding… (esc to interrupt · 1m 45s · ↓ 1.2k tokens)"

	tests := []struct {
		name    string
		content string
		short   bool
		want    string
	}{
		{"normal idle", idleContent, false, "🤖Claude[⌛Idle]"},
		{"short idle", idleContent, true, "🤖CC[⌛I]"},
		{"normal running", runningContent, false, "🤖Claude[🏃Running]"},
		{"short running", runningContent, true, "🤖CC[🏃R]"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := defaultConfig()
			cfg.Display.Short = tt.short
			got := claudeWindowStatus(tt.content, cfg)
			if got != tt.want {
				t.Errorf("claudeWindowStatus() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsClaude(t *testing.T) {
	tests := []struct {
		name  string
		title string
		cmd   string
		want  bool
	}{
		{"claude cmd with idle title", "✳ claude ~/foo", "claude", true},
		{"node cmd with spinner title", "✢ Clauding…", "node", false},
		{"version cmd with spinner title", "✢ Clauding…", "2.1.34", true},
		{"zsh with no claude title", "zsh", "zsh", false},
		{"claude cmd but no claude title", "just a title", "claude", false},
		{"spinner title but zsh cmd", "✢ Clauding…", "zsh", false},
		{"lbox cmd with idle title", "✳ claude ~/foo", "lbox", true},
		{"lbox cmd but no claude title", "just a title", "lbox", false},
	}

	cfg := defaultConfig()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isClaude(tt.title, tt.cmd, cfg)
			if got != tt.want {
				t.Errorf("isClaude(%q, %q) = %v, want %v", tt.title, tt.cmd, got, tt.want)
			}
		})
	}
}

func TestIsClaudeCustomCommand(t *testing.T) {
	cfg := defaultConfig()
	cfg.Claude.Commands = []string{"lbox", "mybot"}

	tests := []struct {
		name  string
		title string
		cmd   string
		want  bool
	}{
		{"custom cmd with idle title", "✳ claude ~/foo", "mybot", true},
		{"custom cmd but no claude title", "just a title", "mybot", false},
		{"lbox still works", "✳ claude ~/foo", "lbox", true},
		{"unlisted cmd", "✳ claude ~/foo", "node", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isClaude(tt.title, tt.cmd, cfg)
			if got != tt.want {
				t.Errorf("isClaude(%q, %q) = %v, want %v", tt.title, tt.cmd, got, tt.want)
			}
		})
	}
}
