package main

import (
	"fmt"
	"regexp"
	"slices"
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

func (s ClaudeState) ShortString() string {
	switch s {
	case ClaudeStateIdle:
		return "I"
	case ClaudeStateRunning:
		return "R"
	case ClaudeStateWaiting:
		return "W"
	default:
		return "U"
	}
}

type ClaudeMode int

const (
	ClaudeModeNone ClaudeMode = iota
	ClaudeModePlan
	ClaudeModeAccept
	ClaudeModeAuto
)

func (m ClaudeMode) String() string {
	switch m {
	case ClaudeModePlan:
		return "plan mode"
	case ClaudeModeAccept:
		return "accept edits"
	case ClaudeModeAuto:
		return "auto mode"
	default:
		return ""
	}
}

type ClaudeStatus struct {
	State ClaudeState
	Mode  ClaudeMode
}

var (
	claudeVersionPattern            = regexp.MustCompile(`^\d+\.\d+`)
	claudeRunningPattern            = regexp.MustCompile(`(?m)^[·*∗✢✳✶✻✽]\s+.+?…?\s*\([^)]*·\s*((?:\d+[smh]\s*)+)`)
	claudeRunningPatternTimeFirst   = regexp.MustCompile(`(?m)^[·*∗✢✳✶✻✽]\s+.+?…?\s*\(((?:\d+[smh]\s*)+)\s*·`)
	claudeRunningPatternTimeOnly    = regexp.MustCompile(`(?m)^[·*∗✢✳✶✻✽]\s+.+?\s*\(\s*(?:\d+[smh]\s*)+\)`)
	claudeRunningFallbackPattern    = regexp.MustCompile(`(?m)^[·*∗✢✳✶✻✽]\s+.+?…?\s*\((esc|ctrl\+c) to interrupt`)
	claudeEscToInterruptEndPattern  = regexp.MustCompile(`(?m)·\s*esc to interrupt(\s|·|$)`)
	claudeRunningPatternSpinnerOnly = regexp.MustCompile(`(?m)^[·*∗✢✳✶✻✽]\s+.+?…`)
	claudeAutoModePattern          = regexp.MustCompile(`⏵⏵\s+auto\s+mode\s+on`)
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

func isClaude(title, cmd string, cfg *Config) bool {
	return mayBeClaudeProcess(cmd, cfg) && mayBeClaudeTitle(title)
}

func mayBeClaudeProcess(cmd string, cfg *Config) bool {
	return cmd == "claude" || claudeVersionPattern.MatchString(cmd) || slices.Contains(cfg.Claude.Commands, cmd)
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
	return slices.Contains(claudeRunningSpinners, runes[0])
}

func lastNonEmptyLines(lines []string, n int) []string {
	var result []string
	for _, v := range slices.Backward(lines) {
		if len(result) >= n {
			break
		}
		line := strings.TrimSpace(v)
		if line == "" {
			continue
		}
		if isSeparatorLine(line) {
			continue
		}
		result = append([]string{v}, result...)
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
	for _, v := range slices.Backward(lines) {
		line := strings.TrimSpace(v)
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
	} else if claudeAutoModePattern.MatchString(combined) {
		s.Mode = ClaudeModeAuto
	} else if claudeAcceptEditsPattern.MatchString(combined) {
		s.Mode = ClaudeModeAccept
	}

	isRunning := claudeRunningPattern.MatchString(combined) ||
		claudeRunningPatternTimeFirst.MatchString(combined) ||
		claudeRunningPatternTimeOnly.MatchString(combined) ||
		claudeRunningFallbackPattern.MatchString(combined) ||
		claudeEscToInterruptEndPattern.MatchString(combined) ||
		claudeRunningPatternSpinnerOnly.MatchString(combined)

	isWaiting := false
	for _, kw := range claudeWaitingPatterns {
		if strings.Contains(combined, kw) {
			isWaiting = true
			break
		}
	}
	if !isWaiting && (claudeInterviewPattern.MatchString(combined) || claudeSelectionMenuPattern.MatchString(combined)) {
		isWaiting = true
	}

	// Idle is confirmed when the trailing line is a ❯ prompt, unless a Running spinner is present.
	// (The ❯ prompt is always visible in the CLI layout even while Running, so Running takes priority.)
	if isClaudePromptLine(lines) && !isRunning {
		s.State = ClaudeStateIdle
		return s
	}

	switch {
	case isWaiting:
		s.State = ClaudeStateWaiting
	case isRunning:
		s.State = ClaudeStateRunning
	case claudeIdlePattern.MatchString(combined):
		s.State = ClaudeStateIdle
	default:
		s.State = ClaudeStateUnknown
	}
	return s
}

func claudeStatusEmoji(s ClaudeStatus, cfg *Config) string {
	icons := cfg.Icons
	var mode string
	switch s.Mode {
	case ClaudeModePlan:
		mode = icons.ModePlan
	case ClaudeModeAccept:
		mode = icons.ModeAccept
	case ClaudeModeAuto:
		mode = icons.ModeAuto
	}
	var state string
	switch s.State {
	case ClaudeStateIdle:
		state = icons.StateIdle
	case ClaudeStateRunning:
		state = icons.StateRunning
	case ClaudeStateWaiting:
		state = icons.StateWaiting
	default:
		state = icons.StateUnknown
	}
	return mode + state
}

func claudeWindowStatus(content string, cfg *Config) string {
	claudeState := parseClaudeStatus(content)
	label := "Claude"
	stateStr := claudeState.State.String()
	if cfg.Display.Short {
		label = "CC"
		stateStr = claudeState.State.ShortString()
	}
	return fmt.Sprintf("%s%s[%s%s]", cfg.Icons.Claude, label, claudeStatusEmoji(claudeState, cfg), stateStr)
}
