package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// saveClaudeDebug appends a debug block containing the raw captured pane text
// and its parsed Claude status to the configured log file.
// All errors are silently ignored to avoid breaking the status-bar output.
func saveClaudeDebug(paneID, content string, cfg *Config) {
	status := parseClaudeStatus(content)
	if len(cfg.Debug.States) > 0 {
		stateStr := status.State.String()
		matched := false
		for _, s := range cfg.Debug.States {
			if strings.EqualFold(s, stateStr) {
				matched = true
				break
			}
		}
		if !matched {
			return
		}
	}

	path := debugLogPath(cfg)
	if path == "" {
		return
	}

	// Skip consecutive duplicate captures for this pane.
	hash := contentHash(content)
	if statePath := debugLastStatePath(path, paneID); statePath != "" {
		if prev, err := os.ReadFile(statePath); err == nil && strings.TrimSpace(string(prev)) == hash {
			return
		}
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()
	header := fmt.Sprintf("===== %s pane=%s state=%s mode=%s =====\n",
		time.Now().Format(time.RFC3339),
		paneID,
		status.State.String(),
		status.Mode.String(),
	)
	_, _ = fmt.Fprintf(f, "%s%s\n", header, content)

	// Persist the hash so the next invocation can detect duplicates.
	if statePath := debugLastStatePath(path, paneID); statePath != "" {
		if err := os.MkdirAll(filepath.Dir(statePath), 0o755); err == nil {
			_ = os.WriteFile(statePath, []byte(hash), 0o644)
		}
	}
}

// contentHash returns the hex-encoded SHA-256 of s.
func contentHash(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

// paneIDUnsafe matches characters that are not safe for use in a filename.
var paneIDUnsafe = regexp.MustCompile(`[^A-Za-z0-9]`)

// debugLastStatePath returns the path of the per-pane state file that stores
// the hash of the last written capture. Returns "" when the log path is empty.
func debugLastStatePath(logPath, paneID string) string {
	if logPath == "" {
		return ""
	}
	safe := paneIDUnsafe.ReplaceAllString(paneID, "_")
	return filepath.Join(filepath.Dir(logPath), "last", safe)
}
