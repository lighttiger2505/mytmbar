package main

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

const claudePrefixIdle = "✳"

type ClaudeState int

const (
	ClaudeStateUnknown ClaudeState = iota
	ClaudeStateIdle
	ClaudeStateRunning
	ClaudeStateWaiting
)

func (s ClaudeState) String() string {
	switch s {
	case ClaudeStateIdle:
		return "Idle"
	case ClaudeStateRunning:
		return "Running"
	case ClaudeStateWaiting:
		return "Waiting"
	default:
		return "Unknown"
	}
}

type ClaudeMode int

const (
	ClaudeModeNone ClaudeMode = iota
	ClaudeModePlan
	ClaudeModeAccept
)

func (m ClaudeMode) String() string {
	switch m {
	case ClaudeModePlan:
		return "plan mode"
	case ClaudeModeAccept:
		return "accept edits"
	default:
		return ""
	}
}

type ClaudeStatus struct {
	State ClaudeState
	Mode  ClaudeMode
}

var (
	claudeVersionPattern           = regexp.MustCompile(`^\d+\.\d+`)
	claudeRunningPattern           = regexp.MustCompile(`(?m)^[✢✽✶✻·]\s+.+?…?\s*\([^)]*·\s*((?:\d+[smh]\s*)+)`)
	claudeRunningPatternTimeFirst  = regexp.MustCompile(`(?m)^[✢✽✶✻·]\s+.+?…?\s*\(((?:\d+[smh]\s*)+)\s*·`)
	claudeRunningFallbackPattern   = regexp.MustCompile(`(?m)^[✢✽✶✻·]\s+.+?…?\s*\((esc|ctrl\+c) to interrupt`)
	claudeEscToInterruptEndPattern = regexp.MustCompile(`(?m)·\s*esc to interrupt(\s|·|$)`)
	claudePlanModePattern          = regexp.MustCompile(`⏸\s+plan\s+mode\s+on`)
	claudeAcceptEditsPattern       = regexp.MustCompile(`⏵⏵\s+accept\s+edits\s+on`)
	claudeIdlePattern              = regexp.MustCompile(`(?m)^\s*❯`)
	claudeSelectionMenuPattern     = regexp.MustCompile(`❯\s+\d+\.`)
	claudeInterviewPattern         = regexp.MustCompile(`Enter to select.*↑/↓ to navigate.*Esc to cancel`)
	claudeFileChangesPattern       = regexp.MustCompile(`^\s*\d+\s+files?\s+[+-]`)
)

var claudeWaitingPatterns = []string{
	"Yes, allow once", "Yes, allow always",
	"Allow once", "Allow always",
	"❯ Yes", "❯ No",
	"Do you trust", "Run this command?",
	"Allow this MCP server", "Continue?", "Proceed?",
	"Do you want to proceed?",
	"(Y/n)", "(y/N)", "[Y/n]", "[y/N]",
}

var claudeRunningSpinners = []rune{'✢', '✽', '✶', '✻', '·'}

func isClaude(title, cmd string) bool {
	return mayBeClaudeProcess(cmd) && mayBeClaudeTitle(title)
}

func mayBeClaudeProcess(cmd string) bool {
	return cmd == "claude" || cmd == "node" || claudeVersionPattern.MatchString(cmd)
}

func mayBeClaudeTitle(title string) bool {
	if strings.HasPrefix(title, claudePrefixIdle) {
		return true
	}
	runes := []rune(title)
	if len(runes) == 0 {
		return false
	}
	if unicode.In(runes[0], unicode.Braille) {
		return true
	}
	for _, s := range claudeRunningSpinners {
		if runes[0] == s {
			return true
		}
	}
	return false
}

func lastNonEmptyLines(lines []string, n int) []string {
	var result []string
	for i := len(lines) - 1; i >= 0 && len(result) < n; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		if isSeparatorLine(line) {
			continue
		}
		result = append([]string{lines[i]}, result...)
	}
	return result
}

func isSeparatorLine(line string) bool {
	for _, r := range line {
		if r < 0x2500 || r > 0x257F {
			return false
		}
	}
	return true
}

func isClaudePromptLine(lines []string) bool {
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		if isSeparatorLine(line) {
			continue
		}
		if strings.Contains(line, "? for shortcuts") ||
			strings.Contains(line, "ctrl+") ||
			strings.Contains(line, "shift+") ||
			claudeFileChangesPattern.MatchString(line) {
			continue
		}
		if strings.HasPrefix(line, "❯") {
			return !claudeSelectionMenuPattern.MatchString(line)
		}
		return false
	}
	return false
}

func parseClaudeStatus(content string) ClaudeStatus {
	lines := strings.Split(content, "\n")
	lastLines := lastNonEmptyLines(lines, 30)
	combined := strings.Join(lastLines, "\n")

	var s ClaudeStatus

	if claudePlanModePattern.MatchString(combined) {
		s.Mode = ClaudeModePlan
	} else if claudeAcceptEditsPattern.MatchString(combined) {
		s.Mode = ClaudeModeAccept
	}

	if claudeRunningPattern.MatchString(combined) ||
		claudeRunningPatternTimeFirst.MatchString(combined) ||
		claudeRunningFallbackPattern.MatchString(combined) ||
		claudeEscToInterruptEndPattern.MatchString(combined) {
		s.State = ClaudeStateRunning
		return s
	}

	if isClaudePromptLine(lines) {
		s.State = ClaudeStateIdle
		return s
	}

	for _, kw := range claudeWaitingPatterns {
		if strings.Contains(combined, kw) {
			s.State = ClaudeStateWaiting
			return s
		}
	}
	if claudeInterviewPattern.MatchString(combined) || claudeSelectionMenuPattern.MatchString(combined) {
		s.State = ClaudeStateWaiting
		return s
	}

	if claudeIdlePattern.MatchString(combined) {
		s.State = ClaudeStateIdle
		return s
	}

	s.State = ClaudeStateUnknown
	return s
}

func claudeStatusEmoji(s ClaudeStatus) string {
	switch s.Mode {
	case ClaudeModePlan:
		return "📋"
	case ClaudeModeAccept:
		return "✏️"
	}
	switch s.State {
	case ClaudeStateIdle:
		return "⌛ "
	case ClaudeStateRunning:
		return "🏃"
	case ClaudeStateWaiting:
		return "🚧"
	default:
		return "❓"
	}
}

func claudeWindowStatus(content string) string {
	claudeState := parseClaudeStatus(content)
	return fmt.Sprintf("🤖 Claude[%s%s]", claudeStatusEmoji(claudeState), claudeState.State)
}
