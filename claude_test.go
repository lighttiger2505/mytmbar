package main

import "testing"

func TestClaudeStatusEmoji(t *testing.T) {
	tests := []struct {
		name   string
		status ClaudeStatus
		want   string
	}{
		{"running", ClaudeStatus{State: claudeStateRunning}, "⚙️"},
		{"idle", ClaudeStatus{State: claudeStateIdle}, "🤖"},
		{"waiting", ClaudeStatus{State: claudeStateWaiting}, "⏳"},
		{"plan mode idle", ClaudeStatus{Mode: claudeModePlan, State: claudeStateIdle}, "📋"},
		{"accept edits running", ClaudeStatus{Mode: claudeModeAccept, State: claudeStateRunning}, "✏️"},
		{"unknown", ClaudeStatus{State: claudeStateUnknown}, "🤖"},
		{"empty", ClaudeStatus{}, "🤖"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := claudeStatusEmoji(tt.status)
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
		wantState string
		wantMode  string
	}{
		{
			name: "Idle with prompt only",
			content: `Some output
───────────────────────────────────────
❯
───────────────────────────────────────`,
			wantState: claudeStateIdle,
		},
		{
			name: "Idle with completion suggestion",
			content: `Some output
───────────────────────────────────────
❯ Try "edit file.go to..."
───────────────────────────────────────`,
			wantState: claudeStateIdle,
		},
		{
			name: "Running with Clauding",
			content: `Some output
✢ Clauding… (esc to interrupt · 1m 45s · ↓ 1.2k tokens)`,
			wantState: claudeStateRunning,
		},
		{
			name: "Running with time first format",
			content: `Some output
✢ Reticulating… (1m 52s · ↓ 11.5k tokens · thought for 7s)`,
			wantState: claudeStateRunning,
		},
		{
			name: "Running with esc to interrupt at end of status line",
			content: `Some output
✶ Proofing… (thinking)
───────────────────────────────────────
❯
───────────────────────────────────────
  4 files +20 -0 · esc to interrupt`,
			wantState: claudeStateRunning,
		},
		{
			name: "Running fallback (ctrl+c to interrupt)",
			content: `Some output
✻ Thinking… (ctrl+c to interrupt)`,
			wantState: claudeStateRunning,
		},
		{
			name: "Running fallback (esc to interrupt)",
			content: `Some output
✻ Processing… (esc to interrupt)`,
			wantState: claudeStateRunning,
		},
		{
			name: "Not Running when text mentions esc to interrupt in quotes",
			content: `Some output about "esc to interrupt" in quotes
❯ `,
			wantState: claudeStateIdle,
		},
		{
			name: "Not Running when indented status line (quoted text)",
			content: `⏺ 現在の内容は：
  ✻ Galloping… (esc to interrupt · 1m 19s · ↓ 5.9k tokens · thinking)

✻ Cooked for 1m 29s

───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
❯
───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────`,
			wantState: claudeStateIdle,
		},
		{
			name: "Waiting with permission prompt",
			content: `Some output
Yes, allow once
Yes, allow always`,
			wantState: claudeStateWaiting,
		},
		{
			name: "Waiting with confirmation prompt",
			content: `Some output
Continue? (Y/n)`,
			wantState: claudeStateWaiting,
		},
		{
			name: "Idle after task completion",
			content: `Some output
✻ Cooked for 43s
───────────────────────────────────────
❯ `,
			wantState: claudeStateIdle,
		},
		{
			name: "Idle after task completion with file changes",
			content: `✻ Sautéed for 2m 55s

───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
❯
───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
  4 files +73 -3`,
			wantState: claudeStateIdle,
		},
		{
			name: "Idle with plan mode",
			content: `Some output
⏸ plan mode on
───────────────────────────────────────
❯
───────────────────────────────────────`,
			wantState: claudeStateIdle,
			wantMode:  claudeModePlan,
		},
		{
			name: "Idle with accept edits",
			content: `Some output
⏵⏵ accept edits on
───────────────────────────────────────
❯
───────────────────────────────────────`,
			wantState: claudeStateIdle,
			wantMode:  claudeModeAccept,
		},
		{
			name: "Waiting - Interview mode",
			content: `  3. ドキュメントのレビュー
  4. Issue の修正
  5. Type something.
  Chat about this
  Skip interview and plan immediately
Enter to select · ↑/↓ to navigate · Esc to cancel`,
			wantState: claudeStateWaiting,
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
			wantState: claudeStateWaiting,
		},
		{
			name: "Running with action text and accept edits mode",
			content: `✶ Adding handler types… (ctrl+c to interrupt · ctrl+t to hide todos · 3m 27s · ↑ 11.0k tokens)
  ⎿  ☐ Add handler types

───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
❯
───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
  ⏵⏵ accept edits on (shift+tab to cycle)`,
			wantState: claudeStateRunning,
			wantMode:  claudeModeAccept,
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
			wantState: claudeStateIdle,
		},
		{
			name: "Idle with file changes status line",
			content: `✻ Churned for 3m 5s

───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
❯
───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────────
  4 files +42 -0`,
			wantState: claudeStateIdle,
		},
		{
			name: "Unknown state",
			content: `Some random output
without any recognizable pattern`,
			wantState: claudeStateUnknown,
		},
		{
			name:      "Empty content",
			content:   "",
			wantState: claudeStateUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseClaudeStatus(tt.content)
			if got.State != tt.wantState {
				t.Errorf("parseClaudeStatus().State = %q, want %q", got.State, tt.wantState)
			}
			if got.Mode != tt.wantMode {
				t.Errorf("parseClaudeStatus().Mode = %q, want %q", got.Mode, tt.wantMode)
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
