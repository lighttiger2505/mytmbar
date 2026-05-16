package main

import (
	"regexp"
	"strings"
	"unicode"
)

const (
	claudePrefixIdle  = "✳"
	stateRunning      = "Running"
	stateIdle         = "Idle"
	stateWaiting      = "Waiting"
	statePlan         = "Plan"
	stateAcceptEdits  = "AcceptEdits"
	stateUnknown      = "Unknown"
)

var (
	claudeVersionPattern    = regexp.MustCompile(`^\d+\.\d+`)
	claudeRunningPattern    = regexp.MustCompile(`(?m)^[✢✽✶✻·]\s+.+?…?\s*\(`)
	claudePlanModePattern   = regexp.MustCompile(`⏸\s+plan\s+mode\s+on`)
	claudeAcceptEditsPattern = regexp.MustCompile(`⏵⏵\s+accept\s+edits\s+on`)

	claudeWaitingKeywords = []string{
		"Yes, allow once", "Yes, allow always", "Allow once", "Allow always",
		"❯ Yes", "❯ No", "Do you trust", "Run this command?",
		"Allow this MCP server", "Continue?", "Proceed?",
		"(Y/n)", "(y/N)", "[Y/n]", "[y/N]",
	}
)

func isClaude(title, cmd string) bool {
	return mayBeClaudeProcess(cmd) && mayBeClaudeTitle(title)
}

func mayBeClaudeProcess(cmd string) bool {
	return cmd == "claude" || cmd == "node" || claudeVersionPattern.MatchString(cmd)
}

var claudeRunningSpinners = []rune{'✢', '✽', '✶', '✻', '·'}

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

func parseClaudeState(title string) string {
	if claudePlanModePattern.MatchString(title) {
		return statePlan
	}
	if claudeAcceptEditsPattern.MatchString(title) {
		return stateAcceptEdits
	}
	if claudeRunningPattern.MatchString(title) {
		return stateRunning
	}
	for _, kw := range claudeWaitingKeywords {
		if strings.Contains(title, kw) {
			return stateWaiting
		}
	}
	if strings.HasPrefix(title, claudePrefixIdle) {
		return stateIdle
	}
	return stateUnknown
}

func claudeStatusEmoji(state string) string {
	switch state {
	case stateRunning:
		return "⚙️"
	case stateIdle:
		return "🤖"
	case stateWaiting:
		return "⏳"
	case statePlan:
		return "📋"
	case stateAcceptEdits:
		return "✏️"
	default:
		return "🤖"
	}
}

func claudeWindowStatus(title string) string {
	state := parseClaudeState(title)
	return claudeStatusEmoji(state) + " Claude"
}
